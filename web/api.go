package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PodFailure struct {
	Namespace     string `json:"namespace"`
	PodName       string `json:"pod_name"`
	ContainerName string `json:"container_name"`
	Failures      string `json:"failures"`
	Recommendation string `json:"recommendation"`
}

func (s *Server) monitorAllHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	k8sClient := s.k8sClient
	
	// Get all namespaces
	namespaces, err := k8sClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		http.Error(w, "Failed to list namespaces", http.StatusInternalServerError)
		return
	}

	var allFailures []PodFailure

	for _, ns := range namespaces.Items {
		// Get all pods in namespace
		pods, err := k8sClient.CoreV1().Pods(ns.Name).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			continue
		}

		for _, pod := range pods.Items {
			for _, container := range pod.Spec.Containers {
				input := strings.Join([]string{ns.Name, pod.Name, container.Name}, "|")
				result, err := s.agent.Execute(context.Background(), input)
				
				if err == nil && result != "No failures detected" && strings.Contains(result, "Failures:") {
					// Parse failures and recommendation from result
					parts := strings.Split(result, "\nRecommendation: ")
					failuresPart := parts[0]
					recommendation := "No recommendation available"
					if len(parts) > 1 {
						recommendation = parts[1]
					}
					
					allFailures = append(allFailures, PodFailure{
						Namespace:      ns.Name,
						PodName:        pod.Name,
						ContainerName:  container.Name,
						Failures:       failuresPart,
						Recommendation: recommendation,
					})
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allFailures)
}