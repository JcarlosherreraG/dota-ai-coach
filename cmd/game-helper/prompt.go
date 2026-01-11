package main

import (
	"context"
	"log"
	"sync"
	"time"
)

// runPromptHandler управляет генерацией промптов: автоматические по таймеру + ручные от пользователя.
func (a *App) runPromptHandler(ctx context.Context, wg *sync.WaitGroup, silenceDuration time.Duration, requestInterval time.Duration) {
	defer wg.Done()
	defer close(a.promptChan)

	ticker := time.NewTicker(requestInterval)
	defer ticker.Stop()

	lastUserAction := time.Now().Add(-1 * time.Hour)

	log.Println("[Prompt Handler] Started")

	for {
		select {
		case <-ctx.Done():
			log.Println("[Prompt Handler] Stopped")
			return

		case userQuestion, ok := <-a.userPromptChan:
			if !ok {
				return
			}

			log.Printf("[Prompt Handler] User Question: %s", userQuestion)
			lastUserAction = time.Now()
			ticker.Reset(requestInterval)
			select {
			case <-ticker.C:
			default:
			}
			fullPrompt := a.advisor.BuildPromptWithQuestion(userQuestion)
			log.Printf("[Prompt Handler] Prompt with user question: %s", fullPrompt)
			a.promptChan <- fullPrompt
		case <-ticker.C:
			if time.Since(lastUserAction) < silenceDuration {
				continue
			}
			prompt := a.advisor.BuildPrompt()
			log.Printf("[Prompt Handler] Default prompt: %s", prompt)
			if prompt != "" {
				select {
				case a.promptChan <- prompt:
					log.Println("[Prompt Handler] Auto-advice queued")
				default:

				}
			}
		}
	}
}
