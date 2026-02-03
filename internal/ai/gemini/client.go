// Package gemini implements the client for Google Gemini API.
package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"io"
	"log"
	"net/http"
	"time"
)

const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/"

// Client represents a client for interacting with the Gemini API.
type Client struct {
	apiKey     string       // API Key for Google services
	model      string       // Model name (e.g., gemini-flash-2.5)
	httpClient *http.Client // HTTP client for requests
	baseUrl    string       // API base URL
}

// NewClient creates a new instance of the Gemini client.
func NewClient(apiKey string, model string, httpTimeOutSeconds int) *Client {
	return &Client{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: time.Duration(httpTimeOutSeconds) * time.Second,
		},
		baseUrl: defaultBaseURL,
	}
}

// Ask sends a request to the Gemini model with system and user prompts.
func (c *Client) Ask(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	// Format URL with API key
	url := fmt.Sprintf("%s%s:generateContent?key=%s", c.baseUrl, c.model, c.apiKey)

	// Prepare request payload (message history)
	requestPayload := Request{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{
					{
						Text: systemPrompt,
					},
				},
			},
			{
				Role: "model",
				Parts: []Part{
					{
						Text: "Got it. I'm ready to analyze.",
					},
				},
			},
			{
				Role: "user",
				Parts: []Part{
					{
						Text: userPrompt,
					},
				},
			},
		},
	}

	// Encode request to JSON
	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return "", fmt.Errorf("error while prompt json encoding: %w", err)
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error while creating HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

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
		return "", ai.NewApiError(resp.StatusCode, errMsg)
	}

	// Decode response
	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decoding JSON response error: %w", err)
	}

	// Check if response contains candidates
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini API response is empty")
	}

	// Return text from the first part of the first candidate
	return response.Candidates[0].Content.Parts[0].Text, nil
}
