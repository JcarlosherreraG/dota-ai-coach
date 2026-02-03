// Package main handles the core application logic and background workers.
package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/BrightGir/game-ai-helper/assets"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"log"
	"sync"
	"time"
)

// fetchAdviceWithRetry performs an AI request with retry logic and linear backoff.
// It handles API errors and retries the request if the error is deemed retryable.
func (a *App) fetchAdviceWithRetry(ctx context.Context, prompt string, baseDelay time.Duration, maxRetries int) (string, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Check if the overall context has been cancelled
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		// Execute the AI request using the configured provider
		advice, err := a.aiProvider.Ask(context.Background(), string(assets.BaseCoachSystemPrompt), prompt)
		if err == nil {
			return advice, nil // Success
		}

		lastErr = err
		// Check if the error is retryable
		if !ai.ShouldRetry(err) {
			return "", err // Fatal error, don't retry
		}

		// Wait before the next attempt with linear backoff
		if attempt < maxRetries-1 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(baseDelay * time.Duration(attempt+1)):
				continue
			}
		}
	}
	return "", lastErr // All attempts failed
}

// runAIWorker is a background worker that processes AI requests from the prompt channel.
// It ensures that LLM calls don't block the main application flow.
func (a *App) runAIWorker(ctx context.Context, wg *sync.WaitGroup, baseDelay time.Duration, maxRetries int) {
	defer wg.Done()
	defer close(a.adviceChan)

	// Panic recovery to keep the application running if a worker fails unexpectedly
	defer func() {
		if r := recover(); r != nil {
			log.Printf("[AI Worker] PANIC RECOVERED: %v", r)
		}
	}()

	// Continuously process prompts from the channel until it's closed
	for prompt := range a.promptChan {
		advice, err := a.fetchAdviceWithRetry(ctx, prompt, baseDelay, maxRetries)

		if err != nil {
			log.Printf("[AI Worker] Error: %v", err)
			// Stop the worker if the context is done
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			continue // Log error and wait for the next prompt
		}

		// Send the received advice to the consumer channel
		select {
		case a.adviceChan <- advice:
		case <-ctx.Done():
			return
		}
	}
	log.Println("[AI Worker] Stopped")
}

// runAdviceConsumer receives AI responses from the advice channel and updates the UI.
func (a *App) runAdviceConsumer(wg *sync.WaitGroup) {
	defer wg.Done()
	// Continuously receive advice and update the overlay
	for advice := range a.adviceChan {
		loggedAdvice := fmt.Sprintf("\n--- ADVICE (%s) ---\n%s\n------------------\n",
			time.Now().Format("15:04:05"), advice)

		// Update the overlay text for the user to see
		a.overlay.SetAiAdvice(advice)
		log.Println("[ADVICE]" + loggedAdvice)
	}
	log.Println("[Advice Consumer] Stopped")
}
