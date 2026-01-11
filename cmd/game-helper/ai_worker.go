package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"log"
	"sync"
	"time"
)

// fetchAdviceWithRetry выполняет запрос к AI с retry логикой и линейным backoff.
func (a *App) fetchAdviceWithRetry(ctx context.Context, prompt string, baseDelay time.Duration, maxRetries int) (string, error) {
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		advice, err := a.aiProvider.Ask(prompt)
		if err == nil {
			return advice, nil
		}

		lastErr = err
		if !ai.ShouldRetry(err) {
			return "", err
		}

		if attempt < maxRetries-1 {
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(baseDelay * time.Duration(attempt+1)):
				continue
			}
		}
	}
	return "", lastErr
}

// runAIWorker воркер для обработки AI запросов из promptChan.
func (a *App) runAIWorker(ctx context.Context, wg *sync.WaitGroup, baseDelay time.Duration, maxRetries int) {
	defer wg.Done()
	defer close(a.adviceChan)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("[AI Worker] PANIC RECOVERED: %v", r)
		}
	}()

	for prompt := range a.promptChan {
		advice, err := a.fetchAdviceWithRetry(ctx, prompt, baseDelay, maxRetries)

		if err != nil {
			log.Printf("[AI Worker] Error: %v", err)
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return
			}
			continue
		}

		select {
		case a.adviceChan <- advice:
		case <-ctx.Done():
			return
		}
	}
	log.Println("[AI Worker] Stopped")
}

// runAdviceConsumer получает советы из adviceChan и отправляет их в оверлей.
func (a *App) runAdviceConsumer(wg *sync.WaitGroup) {
	defer wg.Done()
	for advice := range a.adviceChan {
		loggedAdvice := fmt.Sprintf("\n--- ADVICE (%s) ---\n%s\n------------------\n",
			time.Now().Format("15:04:05"), advice)
		a.overlay.SetAiAdvice(advice)
		log.Println("[ADVICE]" + loggedAdvice)
	}
	log.Println("[Advice Consumer] Stopped")
}
