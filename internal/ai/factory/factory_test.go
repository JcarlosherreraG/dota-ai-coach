package factory

import (
	"testing"

	"github.com/BrightGir/game-ai-helper/internal/ai/gemini"
	"github.com/BrightGir/game-ai-helper/internal/ai/openrouter"
)

func TestCreateClient(t *testing.T) {
	t.Run("Missing API Key", func(t *testing.T) {
		t.Setenv("API_KEY", "")

		_, err := CreateClient("gemini", "mod", 30)
		if err == nil {
			t.Error("Expected error when API_KEY is missing, got nil")
		}
		if err.Error() != "API_KEY is not set in .env" {
			t.Errorf("Wrong error message: %s", err.Error())
		}
	})

	t.Run("Unknown Provider", func(t *testing.T) {
		t.Setenv("API_KEY", "dummy-key")

		_, err := CreateClient("deepseek-fake", "mod", 30)
		if err == nil {
			t.Error("Expected error for unknown provider")
		}
	})

	t.Run("Create Gemini", func(t *testing.T) {
		t.Setenv("API_KEY", "gemini-secret-key")

		client, err := CreateClient("gemini", "gemini-pro", 30)
		if err != nil {
			t.Fatalf("Failed to create gemini client: %v", err)
		}

		_, ok := client.(*gemini.Client)
		if !ok {
			t.Errorf("Expected *gemini.Client, got %T", client)
		}
	})

	t.Run("Create OpenRouter", func(t *testing.T) {
		t.Setenv("API_KEY", "or-secret-key")

		client, err := CreateClient("openrouter", "gpt-4", 30)
		if err != nil {
			t.Fatalf("Failed to create openrouter client: %v", err)
		}

		_, ok := client.(*openrouter.Client)
		if !ok {
			t.Errorf("Expected *openrouter.Client, got %T", client)
		}
	})
}
