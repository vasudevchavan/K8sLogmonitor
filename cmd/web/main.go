package main

import (
	"log"

	"github.com/vasudevchavan/K8sLogmonitor/web"
)

func main() {
	server, err := web.NewServer()
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	log.Println("Starting K8s Log Monitor Web UI on http://localhost:8080")
	if err := server.Start("8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}