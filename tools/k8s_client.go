package tools

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func NewK8sClient() (*kubernetes.Clientset, error) {
	var kubeconfig string
	if kubeConfigPath := os.Getenv("KUBECONFIG"); kubeConfigPath != "" {
		kubeconfig = kubeConfigPath
		klog.Infof("KUBECONFIG is defined:%s", kubeconfig)
	} else {
		kubeconfig = filepath.Join(
			homeDir(), ".kube", "config",
		)
		klog.Info("kubeconfig file was loaded from ~/home/.kube/config")
	}
	var config *rest.Config
	var err error
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig %s: %w", kubeconfig, err)
		}
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to build in-cluster config: %w", err)
		}
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}
	return clientset, nil
}

// GetPodLogs fetches the last 'tailLines' of logs from a pod container
func Int64Ptr(i int64) *int64 { return &i }

func GetPodLogs(client *kubernetes.Clientset, namespace, podName, containerName string, tailLines int64) (string, error) {
	podLogOpts := &corev1.PodLogOptions{
		Container: containerName,
		TailLines: Int64Ptr(tailLines),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req := client.CoreV1().Pods(namespace).GetLogs(podName, podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", fmt.Errorf("error streaming logs: %w", err)
	}
	defer podLogs.Close()

	buf := new(strings.Builder)
	if _, err := io.Copy(buf, podLogs); err != nil {
		return "", fmt.Errorf("error reading logs: %w", err)
	}

	return buf.String(), nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}
