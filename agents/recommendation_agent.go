package agents

import (
	"fmt"
	"strings"

	"github.com/vasudevchavan/K8sLogmonitor/tools"
)

type RecommendationAgent struct {
	llmTool *tools.LLMTool
}

func NewRecommendationAgent(tool *tools.LLMTool) *RecommendationAgent {
	if tool == nil {
		return &RecommendationAgent{llmTool: tools.NewLLMTool("")}
	}
	return &RecommendationAgent{llmTool: tool}
}

func (a *RecommendationAgent) GenerateRecommendation(failures []string, podName, namespace string) (string, error) {
	if len(failures) == 0 {
		return "", fmt.Errorf("no failures provided")
	}
	context := fmt.Sprintf("Pod: %s\nNamespace: %s\nFailures:\n%s",
		podName, namespace, strings.Join(failures, "\n"))
	return a.llmTool.GenerateRecommendation(context)
}
