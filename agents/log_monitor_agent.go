package agents

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/vasudevchavan/K8sLogmonitor/adk"
)

type LogMonitorAgent struct {
	*adk.BaseAgent
	registry adk.ToolRegistry
}

func NewLogMonitorAgent(registry adk.ToolRegistry) *LogMonitorAgent {
	agent := &LogMonitorAgent{
		BaseAgent: adk.NewBaseAgent("log_monitor"),
		registry:  registry,
	}
	return agent
}

func (a *LogMonitorAgent) Execute(ctx context.Context, input string) (string, error) {
	parts := strings.Split(input, "|")
	if len(parts) != 3 {
		return "", fmt.Errorf("input format: namespace|pod_name|container_name")
	}
	
	namespace, podName, containerName := parts[0], parts[1], parts[2]
	
	// Get K8s logs tool
	k8sTool, exists := a.registry.GetTool("k8s_logs")
	if !exists {
		return "", fmt.Errorf("k8s_logs tool not found")
	}
	
	// Fetch logs
	logResult, err := k8sTool.Execute(ctx, map[string]interface{}{
		"namespace":      namespace,
		"pod_name":       podName,
		"container_name": containerName,
		"tail_lines":     int64(100),
	})
	
	var logs string
	if err != nil {
		// If error contains startup/runtime issues, treat as failure context
		errorStr := strings.ToLower(err.Error())
		if strings.Contains(errorStr, "waiting to start") || strings.Contains(errorStr, "pull image") ||
			strings.Contains(errorStr, "crashloopbackoff") || strings.Contains(errorStr, "oomkilled") ||
			strings.Contains(errorStr, "evicted") || strings.Contains(errorStr, "pending") ||
			strings.Contains(errorStr, "probe failed") || strings.Contains(errorStr, "mount") ||
			strings.Contains(errorStr, "volume") || strings.Contains(errorStr, "secret") ||
			strings.Contains(errorStr, "configmap") {
			logs = fmt.Sprintf("Container error: %v", err)
		} else {
			return "", fmt.Errorf("failed to fetch logs: %w", err)
		}
	} else {
		var ok bool
		logs, ok = logResult.(string)
		if !ok {
			return "", fmt.Errorf("unexpected log format")
		}
		if logs == "" {
			return "No failures detected", nil
		}
	}
	
	// Detect failures
	failureTool, exists := a.registry.GetTool("failure_detection")
	if !exists {
		return "", fmt.Errorf("failure_detection tool not found")
	}
	
	failureResult, err := failureTool.Execute(ctx, map[string]interface{}{
		"logs": logs,
	})
	if err != nil {
		return "", fmt.Errorf("failed to detect failures: %w", err)
	}
	
	failures, ok := failureResult.([]string)
	if !ok {
		return "", fmt.Errorf("unexpected failure format")
	}
	
	log.Printf("DEBUG: Detected %d failures for %s/%s: %v", len(failures), podName, containerName, failures)
	
	if len(failures) == 0 {
		return "No failures detected", nil
	}
	
	// Get comprehensive K8s context
	contextTool, exists := a.registry.GetTool("k8s_context")
	if !exists {
		return "", fmt.Errorf("k8s_context tool not found")
	}
	
	k8sContext, err := contextTool.Execute(ctx, map[string]interface{}{
		"namespace": namespace,
		"pod_name":  podName,
	})
	if err != nil {
		log.Printf("Failed to get K8s context: %v", err)
	}
	
	// Search GitHub issues using GitHub agent
	githubAgent := NewGitHubAgent(a.registry)
	githubIssues := "No related issues found."
	
	// Search based on pod name and failure type
	query := podName
	if strings.Contains(strings.Join(failures, " "), "oom") {
		query += " oom memory"
	} else if strings.Contains(strings.Join(failures, " "), "crash") {
		query += " crash"
	} else {
		query += " image"
	}
	log.Printf("DEBUG: Searching GitHub for: %s", query)
	
	// Use GitHub agent to search issues
	githubInput := fmt.Sprintf("%s|vasudevchavan/K8sLogmonitor", query)
	githubResult, err := githubAgent.Execute(ctx, githubInput)
	if err != nil {
		log.Printf("DEBUG: GitHub agent error: %v", err)
	} else {
		log.Printf("DEBUG: GitHub agent result: %s", githubResult)
		// Extract formatted issues from agent result
		if strings.Contains(githubResult, "Related GitHub Issues:") {
			githubIssues = githubResult
		}
	}
	
	// Generate recommendations with full context
	llmTool, exists := a.registry.GetTool("llm_recommendation")
	if !exists {
		return "", fmt.Errorf("llm_recommendation tool not found")
	}
	
	contextStr := fmt.Sprintf(`Pod: %s
Namespace: %s
Failures: %s
Logs: %s
K8s Context: %v
%s`,
		podName, namespace, strings.Join(failures, ", "), logs, k8sContext, githubIssues)
	
	log.Printf("DEBUG: Calling LLM with enhanced context including GitHub issues")
	
	recommendation, err := llmTool.Execute(ctx, map[string]interface{}{
		"context": contextStr,
	})
	if err != nil {
		log.Printf("Failed to generate recommendation: %v", err)
		return fmt.Sprintf("Failures detected: %v", failures), nil
	}
	
	log.Printf("DEBUG: LLM recommendation: %s", recommendation)
	
	return fmt.Sprintf("Failures: %v\nRecommendation: %s", failures, recommendation), nil
}