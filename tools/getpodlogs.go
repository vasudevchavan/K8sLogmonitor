package tools

import (
	"context"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetPodLogsSince(client *kubernetes.Clientset, namespace, podName, containerName string, tailLines int64, sinceTime *metav1.Time) (string, error) {
	podLogOpts := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &tailLines,
		SinceTime: sinceTime,
	}

	req := client.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	podLogs, err := req.Stream(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to stream logs for pod %s/%s container %s: %w", namespace, podName, containerName, err)
	}
	defer podLogs.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", fmt.Errorf("failed to read log stream: %w", err)
	}

	return buf.String(), nil
}
