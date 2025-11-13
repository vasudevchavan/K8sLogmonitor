package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type LLMTool struct {
	apiKey string
	client *http.Client
}

type OpenAIRequest struct {
	Model     string    `json:"model"`
	Messages  []Message `json:"messages"`
	MaxTokens int       `json:"max_tokens"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func NewLLMTool(apiKey string) *LLMTool {
	return &LLMTool{
		apiKey: apiKey,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (t *LLMTool) GenerateRecommendation(context string) (string, error) {
	if context == "" {
		return "", errors.New("context cannot be empty")
	}
	if t.apiKey == "" {
		return "No API key provided. Set LLM_API_KEY environment variable.", nil
	}

	prompt := fmt.Sprintf("Analyze these Kubernetes pod failures and provide specific troubleshooting recommendations:\n\n%s\n\nProvide actionable steps to resolve these issues.", context)

	reqBody := OpenAIRequest{
		Model: "gpt-3.5-turbo",
		Messages: []Message{{
			Role:    "user",
			Content: prompt,
		}},
		MaxTokens: 300,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	resp, err := t.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 429 {
		return t.getFallbackRecommendation(context), nil
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", errors.New("no response from OpenAI")
	}

	return openAIResp.Choices[0].Message.Content, nil
}

// ADK Tool interface methods
func (t *LLMTool) Name() string {
	return "llm_recommendation"
}

func (t *LLMTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	contextStr, ok := input["context"].(string)
	if !ok {
		return nil, errors.New("context must be a string")
	}
	return t.GenerateRecommendation(contextStr)
}

func (t *LLMTool) getFallbackRecommendation(context string) string {
	if strings.Contains(context, "pull image") {
		return "Image Pull Error - Check: 1) Image name/tag correctness 2) Registry accessibility 3) Image pull secrets 4) Network connectivity"
	}
	if strings.Contains(context, "oomkilled") {
		return "OOM Error - Increase memory limits, check resource usage patterns, optimize application memory usage"
	}
	if strings.Contains(context, "crashloopbackoff") {
		return "CrashLoop Error - Check application logs, verify startup commands, review health checks, fix configuration issues"
	}
	if strings.Contains(context, "probe failed") {
		return "Health Check Failed - Verify probe endpoints, adjust timeouts, check application startup time"
	}
	return "General troubleshooting: 1) Check pod events 2) Review logs 3) Verify resources 4) Check dependencies"
}
