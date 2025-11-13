package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/vasudevchavan/K8sLogmonitor/adk"
	"github.com/vasudevchavan/K8sLogmonitor/agents"
	"github.com/vasudevchavan/K8sLogmonitor/tools"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// Initialize Kubernetes client
	k8sClient, err := tools.NewK8sClient()
	if err != nil {
		log.Fatalf("failed to initialize k8s client: %v", err)
	}

	// Initialize ADK registry and tools
	registry := adk.NewToolRegistry()
	
	// Register tools
	registry.RegisterTool("k8s_logs", tools.NewK8sTool(k8sClient))
	registry.RegisterTool("k8s_context", tools.NewK8sContextTool(k8sClient))
	registry.RegisterTool("failure_detection", tools.NewFailureDetectionTool())
	registry.RegisterTool("github_issues", tools.NewGitHubTool(os.Getenv("GITHUB_TOKEN")))
	registry.RegisterTool("llm_recommendation", tools.NewLLMTool(os.Getenv("LLM_API_KEY")))

	// Initialize log monitor agent
	logMonitorAgent := agents.NewLogMonitorAgent(registry)

	const namespace = "default"
	const monitorInterval = 1 * time.Minute

	ticker := time.NewTicker(monitorInterval)
	for range ticker.C {
		// Get all pods in namespace
		pods, err := k8sClient.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Printf("Error listing pods: %v", err)
			continue
		}

		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				input := strings.Join([]string{namespace, pod.Name, container.Name}, "|")
				result, err := logMonitorAgent.Execute(context.Background(), input)
				if err != nil {
					log.Printf("Agent execution failed for %s/%s: %v", pod.Name, container.Name, err)
					continue
				}
				if result != "No failures detected" {
					log.Printf("Pod: %s/%s\n%s\n", pod.Name, container.Name, result)
				}
			}
		}
	}
}