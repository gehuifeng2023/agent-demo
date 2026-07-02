package main

import (
	"fmt"
	"log"
	"net/http"

	"agent-demo/internal/config"
	"agent-demo/internal/handler"
	"agent-demo/internal/llm"
	"agent-demo/internal/service"
)

const configPath = "configs/config.yaml"

func main() {
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	cfg.LLM = config.ApplyModeOverride(cfg.LLM, config.EnvLLMMode())

	llmClient, err := newLLMClient(cfg)
	if err != nil {
		log.Fatalf("create llm client: %v", err)
	}

	chatService := service.NewChatService(llmClient)
	chatHandler := handler.NewChatHandler(chatService)

	mux := http.NewServeMux()
	mux.Handle("/api/v1/chat", chatHandler)

	log.Printf("agent-demo listening on %s", cfg.Service.Address)
	if err := http.ListenAndServe(cfg.Service.Address, mux); err != nil {
		log.Fatalf("server stopped: %v", err)
	}
}

func newLLMClient(cfg config.Config) (llm.Client, error) {
	switch cfg.LLM.Mode {
	case "mock":
		return llm.NewMockClient(), nil
	case "model":
		return llm.NewBailianClient(cfg.LLM.APIKey, cfg.LLM.Model, cfg.LLM.BaseURL, cfg.LLM.Timeout())
	case "deepseek":
		return llm.NewDeepSeekClient(cfg.LLM.APIKey, cfg.LLM.Model, cfg.LLM.BaseURL, cfg.LLM.Timeout())
	case "gemini":
		return llm.NewGeminiClient(cfg.LLM.APIKey, cfg.LLM.Model, cfg.LLM.BaseURL, cfg.LLM.Timeout())
	default:
		return nil, fmt.Errorf("unknown LLM_MODE %q, want mock, model, deepseek, or gemini", cfg.LLM.Mode)
	}
}
