// Package ai предоставляет интерфейс для работы с AI провайдерами.
// Поддерживаются Gemini и OpenRouter через фабрику клиентов.
package ai

import "github.com/BrightGir/game-ai-helper/internal/ai/errors"

type Client interface {
	Ask(prompt string) (string, error)
}

type ApiError = errors.APIError

var NewApiError = errors.NewAPIError

var ShouldRetry = errors.ShouldRetry
