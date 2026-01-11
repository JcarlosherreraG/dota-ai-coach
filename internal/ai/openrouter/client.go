// Package openrouter реализует клиент для OpenRouter API.
package openrouter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/BrightGir/game-ai-helper/internal/ai/errors"
	"io"
	"log"
	"net/http"
	"time"
)

const defaultBaseURL = "https://openrouter.ai/api/v1/chat/completions"

type Client struct {
	apiKey       string
	model        string
	systemPrompt string
	httpClient   *http.Client
	baseURL      string
}

func NewClient(apiKey, model, systemPrompt string) *Client {
	return &Client{
		apiKey:       apiKey,
		model:        model,
		systemPrompt: systemPrompt,
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		baseURL:      defaultBaseURL,
	}
}

func (c *Client) Ask(prompt string) (string, error) {
	requestPayload := Request{
		Model: c.model,
		Messages: []Message{
			{
				Role:    "system",
				Content: c.systemPrompt,
			},
			{
				Role:    "assistant",
				Content: "Got it. I'm ready to analyze.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(requestPayload)
	if err != nil {
		return "", fmt.Errorf("error while prompt json encoding: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error while creating HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
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
		return "", errors.NewAPIError(resp.StatusCode, errMsg)
	}

	var rsp Response
	if err := json.NewDecoder(resp.Body).Decode(&rsp); err != nil {
		return "", fmt.Errorf("decoding JSON reponse error: %w", err)
	}

	if len(rsp.Choices) == 0 {
		return "", fmt.Errorf("openrouter API response is empty")
	}

	return rsp.Choices[0].Message.Content, nil
}
