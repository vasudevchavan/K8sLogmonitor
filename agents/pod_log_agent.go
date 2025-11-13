package agents

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/vasudevchavan/K8sLogmonitor/tools"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodLogAgent struct {
	client         *kubernetes.Clientset
	tailLines      int64
	lastTimestamps map[string]time.Time // key: namespace/pod/container
}

func NewPodLogAgent(client *kubernetes.Clientset, tailLines int64) *PodLogAgent {
	return &PodLogAgent{
		client:         client,
		tailLines:      tailLines,
		lastTimestamps: make(map[string]time.Time),
	}
}

func (p *PodLogAgent) FetchLogs(namespace string) (map[string]string, error) {
	if namespace == "" {
		return nil, fmt.Errorf("namespace cannot be empty")
	}

	pods, err := p.client.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods in namespace %s: %w", namespace, err)
	}

	podLogs := make(map[string]string)

	for _, pod := range pods.Items {
		for _, container := range pod.Spec.Containers {
			key := fmt.Sprintf("%s/%s/%s", namespace, pod.Name, container.Name)
			sinceTime := metav1.NewTime(p.lastTimestamps[key])
			logContent, err := tools.GetPodLogsSince(p.client, namespace, pod.Name, container.Name, p.tailLines, &sinceTime)
			if err != nil {
				// Check if error indicates container startup issues
				if strings.Contains(err.Error(), "waiting to start") || strings.Contains(err.Error(), "pull image") {
					podLogs[key] = fmt.Sprintf("Container startup error: %v", err)
					p.lastTimestamps[key] = time.Now()
				} else {
					log.Printf("Error fetching logs for %s: %v", key, err)
				}
				continue
			}
			if logContent != "" {
				podLogs[key] = logContent
				p.lastTimestamps[key] = time.Now()
			}
		}
	}
	return podLogs, nil
}
