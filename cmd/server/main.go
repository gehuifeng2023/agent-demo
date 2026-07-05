package main

import (
	"agent-demo/internal/llm"
	"log"
	"net/http"
	"os"

	"agent-demo/internal/agent"
	"agent-demo/internal/handler"
)

func main() {
	var llmClient llm.Client
	if os.Getenv("LLM_MODE") == "mock" {
		llmClient = llm.NewMockClient()
		log.Println("LLM mode: mock")
	} else {
		client, err := llm.NewOpenAIClient()
		if err != nil {
			log.Fatalf("create llm client: %v", err)
		}
		llmClient = client
		log.Println("LLM mode: openai")
	}

	agentCore, err := agent.NewAgent(llmClient)
	if err != nil {
		panic(err)
	}

	chatHandler := handler.NewChatHandler(agentCore)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/chat", chatHandler)

	addr := ":8080"

	log.Printf("agent-demo server started, addr=%s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}
