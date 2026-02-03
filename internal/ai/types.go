package ai

import (
	"context"
	"github.com/BrightGir/game-ai-helper/internal/ai/errors"
)

type Client interface {
	Ask(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type ApiError = errors.APIError

var NewApiError = errors.NewAPIError

var ShouldRetry = errors.ShouldRetry
