package tools

import (
	"context"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type K8sContextTool struct {
	client *kubernetes.Clientset
}

type PodContext struct {
	PodStatus    string                 `json:"pod_status"`
	Events       []string               `json:"events"`
	Resources    map[string]interface{} `json:"resources"`
	NodeInfo     string                 `json:"node_info"`
	Dependencies []string               `json:"dependencies"`
}

func NewK8sContextTool(client *kubernetes.Clientset) *K8sContextTool {
	return &K8sContextTool{client: client}
}

func (t *K8sContextTool) Name() string {
	return "k8s_context"
}

func (t *K8sContextTool) Execute(ctx context.Context, input map[string]interface{}) (interface{}, error) {
	namespace, _ := input["namespace"].(string)
	podName, _ := input["pod_name"].(string)
	
	if namespace == "" || podName == "" {
		return nil, errors.New("namespace and pod_name required")
	}

	// Get pod details
	pod, err := t.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	// Get events
	events, _ := t.client.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", podName),
	})

	var eventMsgs []string
	for _, event := range events.Items {
		eventMsgs = append(eventMsgs, fmt.Sprintf("%s: %s", event.Reason, event.Message))
	}

	// Get node info if pod is scheduled
	nodeInfo := "Not scheduled"
	if pod.Spec.NodeName != "" {
		node, err := t.client.CoreV1().Nodes().Get(ctx, pod.Spec.NodeName, metav1.GetOptions{})
		if err == nil {
			nodeInfo = fmt.Sprintf("Node: %s, Ready: %v", node.Name, isNodeReady(node))
		}
	}

	// Get resource info
	resources := make(map[string]interface{})
	for _, container := range pod.Spec.Containers {
		if container.Resources.Requests != nil || container.Resources.Limits != nil {
			resources[container.Name] = map[string]interface{}{
				"requests": container.Resources.Requests,
				"limits":   container.Resources.Limits,
			}
		}
	}

	podContext := PodContext{
		PodStatus:    string(pod.Status.Phase),
		Events:       eventMsgs,
		Resources:    resources,
		NodeInfo:     nodeInfo,
		Dependencies: getDependencies(pod),
	}

	return podContext, nil
}

func isNodeReady(node *corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func getDependencies(pod *corev1.Pod) []string {
	var deps []string
	if pod.Spec.ServiceAccountName != "" {
		deps = append(deps, "ServiceAccount: "+pod.Spec.ServiceAccountName)
	}
	for _, volume := range pod.Spec.Volumes {
		if volume.Secret != nil {
			deps = append(deps, "Secret: "+volume.Secret.SecretName)
		}
		if volume.ConfigMap != nil {
			deps = append(deps, "ConfigMap: "+volume.ConfigMap.Name)
		}
	}
	return deps
}