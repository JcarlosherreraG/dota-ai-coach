// Package openrouter implements the client for OpenRouter API.
package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/BrightGir/game-ai-helper/internal/ai/errors"
	"io"
	"log"
	"net/http"
	"time"
)

const defaultBaseURL = "https://openrouter.ai/api/v1/chat/completions"

// Client represents a client for interacting with the OpenRouter API.
type Client struct {
	apiKey     string       // OpenRouter API Key
	model      string       // Model identifier (e.g., openai/gpt-4o)
	httpClient *http.Client // HTTP client for requests
	baseURL    string       // API base URL
}

// NewClient creates a new instance of the OpenRouter client.
func NewClient(apiKey string, model string, httpTimeOutSeconds int) *Client {
	return &Client{
		apiKey:     apiKey,
		model:      model,
		httpClient: &http.Client{Timeout: time.Duration(httpTimeOutSeconds) * time.Second},
		baseURL:    defaultBaseURL,
	}
}

// Ask sends a request to OpenRouter with system and user prompts.
func (c *Client) Ask(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	// Prepare request payload in Chat Completion format
	requestPayload := Request{
		Model: c.model,
		Messages: []Message{
			{
				Role:    "system",
				Content: systemPrompt,
			},
			{
				Role:    "assistant",
				Content: "Got it. I'm ready to analyze.",
			},
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
		MaxTokens: 13000,
	}

	// Encode request to JSON
	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return "", fmt.Errorf("error while prompt json encoding: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error while creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error while doing HTTP request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("WARN: failed to close response body: %v", err)
		}
	}(resp.Body)

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		errMsg := string(bodyBytes)
		if err != nil {
			errMsg = "failed to read error body"
		}
		return "", errors.NewAPIError(resp.StatusCode, errMsg)
	}

	// Decode response
	var rsp Response
	if err := json.NewDecoder(resp.Body).Decode(&rsp); err != nil {
		return "", fmt.Errorf("decoding JSON response error: %w", err)
	}

	// Check if response contains choices
	if len(rsp.Choices) == 0 {
		return "", fmt.Errorf("openrouter API response is empty")
	}

	// Return content from the first message of the first choice
	return rsp.Choices[0].Message.Content, nil
}
