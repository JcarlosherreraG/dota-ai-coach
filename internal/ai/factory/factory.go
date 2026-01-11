// Package factory создаёт AI клиентов на основе конфигурации.
package factory

import (
	"fmt"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"github.com/BrightGir/game-ai-helper/internal/ai/gemini"
	"github.com/BrightGir/game-ai-helper/internal/ai/openrouter"
	"os"
)

// CreateClient создаёт клиент для указанного AI провайдера.
// Поддерживаемые провайдеры: "gemini", "openrouter".
func CreateClient(providerName, model, systemPrompt string) (ai.Client, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY is not set in .env")
	}

	switch providerName {
	case "gemini":
		return gemini.NewClient(apiKey, model, systemPrompt), nil
	case "openrouter":
		return openrouter.NewClient(apiKey, model, systemPrompt), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}
}
