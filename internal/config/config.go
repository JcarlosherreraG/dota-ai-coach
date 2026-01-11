// Package config обрабатывает конфигурацию приложения из config.json.
package config

import (
	"encoding/json"
	"os"
)

// Config конфигурация приложения.
type Config struct {
	Provider               string `json:"provider"`                 // AI провайдер: "gemini" или "openrouter"
	Model                  string `json:"model"`                    // Модель AI (например, "gemini-2.5-pro")
	SystemPrompt           string `json:"system_prompt"`            // Системный промпт для AI
	RequestIntervalSeconds int    `json:"request_interval_seconds"` // Интервал автоматических запросов
	SilenceDurationSeconds int    `json:"silence_duration_seconds"` // Пауза после действий пользователя, после которой будут формироваться автоматические промпты
	GsiPort                int    `json:"local_gsi_port"`           // Порт для GSI сервера
	HotKeyTurnOverlay      int    `json:"hotkey_turn_overlay"`      // VK код для вкл/выкл оверлея
	HotKeyFocusOverlay     int    `json:"hotkey_focus_overlay"`     // VK код для фокуса оверлея
}

// LoadConfig загружает конфигурацию из JSON файла.
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
