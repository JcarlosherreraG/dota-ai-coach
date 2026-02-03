package factory

import (
	"fmt"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"github.com/BrightGir/game-ai-helper/internal/ai/gemini"
	"github.com/BrightGir/game-ai-helper/internal/ai/openrouter"
	"os"
)

// CreateClient creates a client for the specified AI provider.
// Supported providers: "gemini", "openrouter".
func CreateClient(providerName, model string, httpTimeOutSeconds int) (ai.Client, error) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("API_KEY is not set in .env")
	}

	switch providerName {
	case "gemini":
		return gemini.NewClient(apiKey, model, httpTimeOutSeconds), nil
	case "openrouter":
		return openrouter.NewClient(apiKey, model, httpTimeOutSeconds), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", providerName)
	}
}
