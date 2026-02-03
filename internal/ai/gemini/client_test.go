package gemini

import (
	"context"
	"errors"
	aierrors "github.com/BrightGir/game-ai-helper/internal/ai/errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_Ask(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		apiKey := "fake-key-123"
		model := "gemini-pro"
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST request, got %s", r.Method)
			}

			queryKey := r.URL.Query().Get("key")
			if queryKey != apiKey {
				t.Errorf("Wrong API Key in URL. Want %s, got %s", apiKey, queryKey)
			}

			responseJSON := `{
			  "candidates": [
				{
				  "content": {
					"parts": [
					  { "text": "I am Gemini" }
					],
					"role": "model"
				  },
				  "finishReason": "STOP",
				  "index": 0,
				  "safetyRatings": []
				}
			  ]
			}`

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(responseJSON))
		}))
		defer server.Close()

		client := NewClient(apiKey, model, 30)
		client.baseUrl = server.URL + "/"

		got, err := client.Ask(context.Background(), "system", "hello")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got != "I am Gemini" {
			t.Errorf("Want 'I am Gemini', got '%s'", got)
		}
	})

	t.Run("API Error 429", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests) // 429
			w.Write([]byte(`{"error": {"message": "Rate limit exceeded"}}`))
		}))
		defer server.Close()

		client := NewClient("k", "m", 30)
		client.baseUrl = server.URL + "/"

		_, err := client.Ask(context.Background(), "system", "hi")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		var apiErr *aierrors.APIError
		ok := errors.As(err, &apiErr)
		if !ok {
			t.Fatalf("Expected *APIError, got %T: %v", err, err)
		}

		if apiErr.StatusCode != 429 {
			t.Errorf("Expected status 429, got %d", apiErr.StatusCode)
		}
	})

	t.Run("Empty Candidates (Filtered)", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{ "candidates": [], "promptFeedback": {} }`))
		}))
		defer server.Close()

		client := NewClient("k", "m", 30)
		client.baseUrl = server.URL + "/"

		_, err := client.Ask(context.Background(), "system", "prompt")
		if err == nil {
			t.Fatal("Expected error on empty candidates, got nil")
		}

		expectedPart := "gemini API response is empty"
		if err.Error() != expectedPart && err.Error() != "gemini API response is empty" {
			if err.Error() != expectedPart {
				t.Errorf("Wrong error msg. Want '%s', got '%v'", expectedPart, err)
			}
		}
	})
}
