package agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/vasudevchavan/K8sLogmonitor/adk"
	"github.com/vasudevchavan/K8sLogmonitor/tools"
)

type GitHubAgent struct {
	*adk.BaseAgent
	registry adk.ToolRegistry
}

func NewGitHubAgent(registry adk.ToolRegistry) *GitHubAgent {
	agent := &GitHubAgent{
		BaseAgent: adk.NewBaseAgent("github_agent"),
		registry:  registry,
	}
	return agent
}

func (a *GitHubAgent) Execute(ctx context.Context, input string) (string, error) {
	parts := strings.Split(input, "|")
	if len(parts) < 2 {
		return "", fmt.Errorf("input format: query|repo")
	}
	
	query, repo := parts[0], parts[1]
	
	// Get GitHub issues tool
	githubTool, exists := a.registry.GetTool("github_issues")
	if !exists {
		return "", fmt.Errorf("github_issues tool not found")
	}
	
	// Search for issues
	issues, err := githubTool.Execute(ctx, map[string]interface{}{
		"query": query,
		"repo":  repo,
	})
	if err != nil {
		return "", fmt.Errorf("failed to search GitHub issues: %w", err)
	}
	
	// Format issues for LLM if GitHub tool is available
	if gt, ok := githubTool.(*tools.GitHubTool); ok {
		if issueList, ok := issues.([]tools.GitHubIssue); ok {
			return gt.FormatIssuesForLLM(issueList), nil
		}
	}
	
	return "No related GitHub issues found.", nil
}