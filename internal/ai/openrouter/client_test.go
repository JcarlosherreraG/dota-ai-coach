package openrouter

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
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer test-key" {
				t.Errorf("Wrong Auth header: %s", r.Header.Get("Authorization"))
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("Wrong Content-Type: %s", r.Header.Get("Content-Type"))
			}

			responseJSON := `{
				"id": "gen-123",
				"choices": [
					{
						"message": {
							"role": "assistant",
							"content": "Hello, Human!"
						}
					}
				]
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(responseJSON))
		}))
		defer server.Close()

		client := NewClient("test-key", "m", 30)
		client.baseURL = server.URL

		got, err := client.Ask(context.Background(), "system", "hi")

		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if got != "Hello, Human!" {
			t.Errorf("Want 'Hello, Human!', got '%s'", got)
		}
	})

	t.Run("API Error 500", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		client := NewClient("k", "m", 30)
		client.baseURL = server.URL

		_, err := client.Ask(context.Background(), "system", "hi")

		if err == nil {
			t.Fatal("Expected error, got nil")
		}

		var apiErr *aierrors.APIError
		ok := errors.As(err, &apiErr)
		if !ok {
			t.Fatalf("Expected *errors.APIError, got %T: %v", err, err)
		}

		if apiErr.StatusCode != 500 {
			t.Errorf("Expected status 500, got %d", apiErr.StatusCode)
		}
	})

	t.Run("Malformed JSON Response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{ "invalid": json, `))
		}))
		defer server.Close()

		client := NewClient("k", "m", 30)
		client.baseURL = server.URL

		_, err := client.Ask(context.Background(), "system", "hi")
		if err == nil {
			t.Fatal("Expected JSON decode error, got nil")
		}
	})
}
