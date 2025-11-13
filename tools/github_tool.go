package tools

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type GitHubTool struct {
	client *http.Client
	token  string
}

type GitHubIssue struct {
	Title   string `json:"title"`
	Body    string `json:"body"`
	HTMLURL string `json:"html_url"`
	State   string `json:"state"`
}

type GitHubSearchResponse struct {
	Items []GitHubIssue `json:"items"`
}

func NewGitHubTool(token string) *GitHubTool {
	return &GitHubTool{
		client: &http.Client{Timeout: 10 * time.Second},
		token:  token,
	}
}

func (t *GitHubTool) Name() string {
	return "github_issues"
}

func (t *GitHubTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	query, ok := input["query"].(string)
	if !ok {
		return nil, errors.New("query must be a string")
	}

	repo, _ := input["repo"].(string)
	if repo == "" {
		repo = "vasudevchavan/K8sLogmonitor"
	}

	return t.SearchIssues(query, repo)
}

func (t *GitHubTool) SearchIssues(query, repo string) ([]GitHubIssue, error) {
	// Search in both title and body content
	searchQuery := fmt.Sprintf("repo:%s %s in:title,body", repo, query)
	encodedQuery := url.QueryEscape(searchQuery)

	apiURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s&sort=updated&per_page=5", encodedQuery)
	
	issues, err := t.doSearch(apiURL)
	if err != nil {
		return nil, err
	}
	
	// If no results, try broader search (all issues in repo)
	if len(issues) == 0 {
		broadQuery := fmt.Sprintf("repo:%s", repo)
		encodedBroadQuery := url.QueryEscape(broadQuery)
		broadURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s&sort=updated&per_page=5", encodedBroadQuery)
		return t.doSearch(broadURL)
	}
	
	return issues, nil
}

func (t *GitHubTool) doSearch(url string) ([]GitHubIssue, error) {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if t.token != "" {
		req.Header.Set("Authorization", "token "+t.token)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API request failed with status: %d", resp.StatusCode)
	}

	var searchResp GitHubSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return searchResp.Items, nil
}

func (t *GitHubTool) FormatIssuesForLLM(issues []GitHubIssue) string {
	if len(issues) == 0 {
		return "No related GitHub issues found."
	}

	var result strings.Builder
	result.WriteString("Related GitHub Issues:\n")

	for i, issue := range issues {
		if i >= 3 { // Limit to top 3 issues
			break
		}
		// Extract issue number from URL
		issueNum := t.extractIssueNumber(issue.HTMLURL)
		result.WriteString(fmt.Sprintf("%d. Issue #%s: %s (%s)\n",
			i+1, issueNum, issue.Title, issue.State))
		
		// Include issue body if available
		if issue.Body != "" {
			body := strings.TrimSpace(issue.Body)
			if len(body) > 100 {
				body = body[:100] + "..."
			}
			result.WriteString(fmt.Sprintf("   Description: %s\n", body))
		}
		result.WriteString(fmt.Sprintf("   URL: %s\n", issue.HTMLURL))
	}

	return result.String()
}

func (t *GitHubTool) extractIssueNumber(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return "unknown"
}
