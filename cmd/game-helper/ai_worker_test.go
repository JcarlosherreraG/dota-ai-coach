package main

import (
	"context"
	"errors"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"sync"
	"testing"
	"time"
)

type mockAIClient struct {
	AskFunc func(prompt string) (string, error)
}

func (m *mockAIClient) Ask(prompt string) (string, error) {
	if m.AskFunc != nil {
		return m.AskFunc(prompt)
	}
	return "default answer", nil
}

func TestApp_fetchAdviceWithRetry(t *testing.T) {
	tests := []struct {
		name         string
		prompt       string
		mockBehavior func() func(string) (string, error)
		wantAdvice   string
		wantErr      bool
	}{
		{
			name:   "Success on first try",
			prompt: "hello",
			mockBehavior: func() func(string) (string, error) {
				return func(p string) (string, error) {
					return "Easy Answer", nil
				}
			},
			wantAdvice: "Easy Answer",
			wantErr:    false,
		},
		{
			name:   "Success after retry (Network Glitch)",
			prompt: "retry me",
			mockBehavior: func() func(string) (string, error) {
				attempts := 0
				return func(p string) (string, error) {
					attempts++
					if attempts < 3 {
						return "", ai.NewApiError(500, "some error")
					}
					return "Hard Answer", nil
				}
			},
			wantAdvice: "Hard Answer",
			wantErr:    false,
		},
		{
			name:   "Fail max retries",
			prompt: "die",
			mockBehavior: func() func(string) (string, error) {
				return func(p string) (string, error) {
					return "", errors.New("network glitch")
				}
			},
			wantAdvice: "",
			wantErr:    true,
		},
		{
			name:   "Fatal error (No retry)",
			prompt: "fatal",
			mockBehavior: func() func(string) (string, error) {
				return func(p string) (string, error) {
					return "", errors.New("fatal error: apiKey invalid")
				}
			},
			wantAdvice: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockAIClient{
				AskFunc: tt.mockBehavior(),
			}

			app := &App{
				aiProvider: mock,
			}

			got, err := app.fetchAdviceWithRetry(context.Background(), tt.prompt, time.Nanosecond, 3)

			if (err != nil) != tt.wantErr {
				t.Errorf("fetchAdviceWithRetry() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantAdvice {
				t.Errorf("fetchAdviceWithRetry() = %v, want %v", got, tt.wantAdvice)
			}
		})
	}
}

func TestWorker_Panic(t *testing.T) {
	mock := &mockAIClient{
		AskFunc: func(prompt string) (string, error) {
			panic("boom!")
		},
	}

	app := &App{
		promptChan: make(chan string, 1),
		adviceChan: make(chan string),
		aiProvider: mock,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go app.runAIWorker(context.Background(), wg, time.Nanosecond, 3)

	app.promptChan <- "help me"
	close(app.promptChan)

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:

	case <-time.After(1 * time.Second):
		t.Fatal("Worker died without recover (wg.Done not called)")
	}
}

func TestWorker_Cancel(t *testing.T) {
	mock := &mockAIClient{
		AskFunc: func(prompt string) (string, error) {
			time.Sleep(2 * time.Second)
			return "too late", nil
		},
	}

	app := &App{
		promptChan: make(chan string, 1),
		adviceChan: make(chan string),
		aiProvider: mock,
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go app.runAIWorker(ctx, wg, time.Millisecond, 3)

	app.promptChan <- "wait"

	cancel()

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()

	select {
	case <-done:

	case <-time.After(100 * time.Millisecond):
		t.Fatal("Worker stuck and ignored context cancellation")
	}
}

func TestWorker_Retries(t *testing.T) {
	var callCount int

	mock := &mockAIClient{
		AskFunc: func(prompt string) (string, error) {
			callCount++
			if callCount < 3 {
				return "", ai.NewApiError(500, "some error")
			}
			return "success", nil
		},
	}

	app := &App{
		promptChan: make(chan string, 1),
		adviceChan: make(chan string, 1),
		aiProvider: mock,
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	go app.runAIWorker(context.Background(), wg, time.Nanosecond, 3)

	app.promptChan <- "try"
	close(app.promptChan)

	wg.Wait()

	select {
	case res := <-app.adviceChan:
		if res != "success" {
			t.Errorf("Want success, got %s", res)
		}
	default:
		t.Error("No result in channel")
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
}
