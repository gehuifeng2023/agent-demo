package main

import (
	"log"
	"net/http"

	"agent-demo/internal/agent"
	"agent-demo/internal/handler"
)

func main() {
	agentCore := agent.NewAgent()
	chatHandler := handler.NewChatHandler(agentCore)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/chat", chatHandler)

	addr := ":8080"

	log.Printf("agent-demo server started, addr=%s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
