// Package config handles the application configuration from config.json.
package config

import (
	"encoding/json"
	"os"
)

// Config represents the application configuration.
type Config struct {
	Provider               string  `json:"provider"`                 // AI Provider: "gemini" or "openrouter"
	Model                  string  `json:"model"`                    // AI Model (e.g., "gemini-2.5-pro")
	SystemPrompt           string  `json:"system_prompt"`            // System prompt for AI
	RequestIntervalSeconds int     `json:"request_interval_seconds"` // Interval for automatic requests
	SilenceDurationSeconds int     `json:"silence_duration_seconds"` // Pause after user actions before auto prompts
	GsiPort                int     `json:"local_gsi_port"`           // Port for GSI server
	HotKeyTurnOverlay      int     `json:"hotkey_turn_overlay"`      // VK code for overlay toggle
	HotKeyFocusOverlay     int     `json:"hotkey_focus_overlay"`     // VK code for overlay focus
	MinSimilarity          float32 `json:"minSimilarity"`            // Minimum similarity for RAG search
}

// LoadConfig loads the configuration from a JSON file.
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
