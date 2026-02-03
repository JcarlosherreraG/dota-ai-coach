package main

import (
	"context"
	"github.com/BrightGir/game-ai-helper/assets"
	"log"
	"sync"
	"time"
)

// runPromptHandler manages the prompt generation lifecycle:
// It handles both periodic automatic advice and immediate user questions.
func (a *App) runPromptHandler(ctx context.Context, wg *sync.WaitGroup, silenceDuration time.Duration, requestInterval time.Duration) {
	defer wg.Done()
	defer close(a.promptChan)

	// Ticker for periodic automatic advice
	var ticker *time.Ticker
	var tickerChan <-chan time.Time = nil

	if requestInterval > 0 {
		ticker = time.NewTicker(requestInterval)
		tickerChan = ticker.C
		defer ticker.Stop()
		log.Printf("[Prompt Handler] Auto-advice enabled, interval: %v", requestInterval)
	} else {
		log.Println("[Prompt Handler] Auto-advice disabled (interval <= 0)")
	}

	// Track the last time the user asked a question to prevent overlapping auto-advice
	lastUserAction := time.Now().Add(-1 * time.Hour)

	log.Println("[Prompt Handler] Started")

	for {
		select {
		case <-ctx.Done():
			log.Println("[Prompt Handler] Stopped")
			return

		case userQuestion, ok := <-a.userPromptChan:
			// Handle question from the user via the overlay
			if !ok {
				return
			}

			log.Printf("[Prompt Handler] User Question: %s", userQuestion)
			lastUserAction = time.Now()

			// Reset the automatic ticker when a question is asked
			if ticker != nil {
				ticker.Reset(requestInterval)
				select {
				case <-ticker.C:
				default:
				}
			}

			// Build the full prompt and send to the worker
			fullPrompt, err := a.buildPrompt(ctx, userQuestion)
			if err != nil {
				log.Printf("[Prompt Handler] failed to build prompt: %v", err)
				continue
			}
			log.Printf("[Prompt Handler] Prompt with user question: %s", fullPrompt)
			a.promptChan <- fullPrompt

		case <-tickerChan:
			// Handle periodic automatic advice generation

			// Only generate auto-advice if the user has been silent for the configured duration
			if time.Since(lastUserAction) < silenceDuration {
				continue
			}

			// Ensure we have valid game state before requesting advice
			if !a.store.HasState() {
				continue
			}

			// Build the default tactical advice prompt
			fullPrompt, err := a.buildPrompt(ctx, assets.AutoUserQuestion)
			if err != nil {
				log.Printf("[Prompt Handler] failed to build prompt: %v", err)
				continue
			}

			log.Printf("[Prompt Handler] Default prompt: %s", fullPrompt)
			if fullPrompt != "" {
				// Queue the prompt for the AI worker
				select {
				case a.promptChan <- fullPrompt:
					log.Println("[Prompt Handler] Auto-advice queued")
				default:
					// If the channel is full, we skip this cycle to avoid blocking
				}
			}
		}
	}
}

// buildPrompt uses the prompt builder to construct a detailed prompt with a timeout.
func (a *App) buildPrompt(rootCtx context.Context, question string) (string, error) {
	// Use a dedicated context with timeout for the prompt building process (including RAG)
	promptCtx, cancel := context.WithTimeout(rootCtx, 5*time.Second)
	defer cancel()

	// Delegate the actual building logic to the promptBuilder component
	fullPrompt, err := a.promptBuilder.Build(promptCtx, question)
	return fullPrompt, err
}
