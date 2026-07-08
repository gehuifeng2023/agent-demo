package main

import (
	"agent-demo/internal/llm"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"agent-demo/internal/agent"
	"agent-demo/internal/handler"
)

func main() {
	llmClient, mode, err := newLLMClientFromEnv()
	if err != nil {
		log.Fatalf("create llm client: %v", err)
	}
	log.Printf("LLM mode: %s", mode)

	agentCore, err := agent.NewAgent(llmClient)
	if err != nil {
		panic(err)
	}

	chatHandler := handler.NewChatHandler(agentCore)
	fileHandler := handler.NewFileHandler("uploads", 20<<20, agentCore)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/chat", chatHandler)
	mux.HandleFunc("/api/v1/files/upload", fileHandler.Upload)

	addr := ":8080"

	log.Printf("agent-demo server started, addr=%s", addr)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func newLLMClientFromEnv() (llm.Client, string, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("LLM_MODE")))

	switch mode {
	case "", "openai":
		client, err := llm.NewOpenAIClient()
		if err != nil {
			return nil, "openai", err
		}
		return client, "openai", nil
	case "mock":
		return llm.NewMockClient(), "mock", nil
	case "gemini":
		client, err := llm.NewGeminiClient()
		if err != nil {
			return nil, "gemini", err
		}
		return client, "gemini", nil
	default:
		return nil, "", fmt.Errorf("unsupported LLM_MODE: %s", mode)
	}
}
