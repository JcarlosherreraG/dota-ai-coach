// Package gemini реализует клиент для Google Gemini API.
package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"io"
	"log"
	"net/http"
	"time"
)

const defaultBaseURL = "https://generativelanguage.googleapis.com/v1beta/models/"

type Client struct {
	apiKey       string
	systemPrompt string
	model        string
	httpClient   *http.Client
	baseUrl      string
}

func NewClient(apiKey string, model string, systemPrompt string) *Client {
	return &Client{
		apiKey:       apiKey,
		model:        model,
		systemPrompt: systemPrompt,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
		baseUrl: defaultBaseURL,
	}
}

func (c *Client) Ask(prompt string) (string, error) {
	url := fmt.Sprintf("%s%s:generateContent?key=%s", c.baseUrl, c.model, c.apiKey)
	requestPayload := Request{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{
					{
						Text: c.systemPrompt,
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
						Text: prompt,
					},
				},
			},
		},
	}
	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return "", fmt.Errorf("error while prompt json encoding: %w", err)
	}
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error while doing HTTP request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("WARN: failed to close request body: %v", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, err := io.ReadAll(resp.Body)
		errMsg := string(bodyBytes)
		if err != nil {
			errMsg = "failed to read error body"
		}
		return "", ai.NewApiError(resp.StatusCode, errMsg)
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decoding JSON response error: %w", err)
	}
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini API response is empty")
	}
	return response.Candidates[0].Content.Parts[0].Text, nil
}
