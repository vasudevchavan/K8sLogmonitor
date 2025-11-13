package tools

import (
	"context"
	"errors"
	"regexp"

	"k8s.io/client-go/kubernetes"
)

type K8sTool struct {
	client *kubernetes.Clientset
}

func NewK8sTool(client *kubernetes.Clientset) *K8sTool {
	return &K8sTool{client: client}
}

// ADK Tool interface methods
func (t *K8sTool) Name() string {
	return "k8s_logs"
}

func (t *K8sTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	namespace, ok := input["namespace"].(string)
	if !ok {
		return nil, errors.New("namespace must be a string")
	}
	
	podName, ok := input["pod_name"].(string)
	if !ok {
		return nil, errors.New("pod_name must be a string")
	}
	
	containerName, ok := input["container_name"].(string)
	if !ok {
		return nil, errors.New("container_name must be a string")
	}
	
	tailLines, ok := input["tail_lines"].(int64)
	if !ok {
		tailLines = 100
	}
	
	return GetPodLogs(t.client, namespace, podName, containerName, tailLines)
}

type FailureDetectionTool struct {
	patterns []*regexp.Regexp
}

func NewFailureDetectionTool() *FailureDetectionTool {
	patternStrings := []string{
		"panic:", "error:", "failed to .*", "connection refused",
		"pull image", "startup error", "waiting to start", "imagepullbackoff",
		"crashloopbackoff", "oomkilled", "out of memory", "memory limit",
		"cpu throttling", "disk pressure", "evicted", "pending",
		"readiness probe failed", "liveness probe failed", "startup probe failed",
		"mount.*failed", "volume.*error", "secret.*not found", "configmap.*not found",
		"service unavailable", "timeout", "deadline exceeded", "context canceled",
		"permission denied", "forbidden", "unauthorized", "tls.*error",
		"dns.*error", "network.*unreachable", "no route to host",
	}
	patterns := make([]*regexp.Regexp, len(patternStrings))
	for i, p := range patternStrings {
		patterns[i] = regexp.MustCompile("(?i)" + p)
	}
	return &FailureDetectionTool{patterns: patterns}
}

func (t *FailureDetectionTool) Name() string {
	return "failure_detection"
}

func (t *FailureDetectionTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	logs, ok := input["logs"].(string)
	if !ok {
		return nil, errors.New("logs must be a string")
	}
	
	var failures []string
	for _, re := range t.patterns {
		matches := re.FindAllString(logs, -1)
		failures = append(failures, matches...)
	}
	return failures, nil
}