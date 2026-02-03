package main

import (
	"context"
	"github.com/BrightGir/game-ai-helper/internal/dota"
	"github.com/BrightGir/game-ai-helper/internal/state"
	"sync"
	"testing"
	"time"
)

type mockAdvisor struct{}

func (m *mockAdvisor) Build(ctx context.Context, question string) (string, error) {
	if question == "" {
		return "Auto prompt from mock", nil
	}
	return "Answer to: " + question, nil
}

func TestPromptHandler_SilenceLogic(t *testing.T) {
	requestInterval := 100 * time.Millisecond
	silenceDuration := 30 * time.Millisecond

	mockAdvisor := &mockAdvisor{}
	store := state.NewStore()
	
	testState := &dota.GameState{
		Hero: dota.Hero{Name: "pudge"},
	}
	store.Update(testState)

	app := &App{
		promptBuilder:  mockAdvisor,
		promptChan:     make(chan string, 10),
		userPromptChan: make(chan string, 10),
		store:          store,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go app.runPromptHandler(ctx, wg, silenceDuration, requestInterval)

	time.Sleep(20 * time.Millisecond)

	t.Log("Step 1: Sending user prompt")
	app.userPromptChan <- "Why?"

	select {
	case p := <-app.promptChan:
		if p == "Auto prompt from mock" {
			t.Fatal("Got auto-prompt instead of user answer")
		}
		t.Logf("Received user response: %s", p)
	case <-time.After(50 * time.Millisecond):
		t.Fatal("User prompt stuck")
	}

	t.Log("Step 2: Checking silence interval")

	time.Sleep(30 * time.Millisecond)

	select {
	case p := <-app.promptChan:
		t.Fatalf("Got prompt too early! Ticker fired before interval? Msg: %s", p)
	default:
		t.Log("Good: Channel is empty (Silence works)")
	}

	t.Log("Step 3: Waiting for ticker to fire...")

	time.Sleep(60 * time.Millisecond)

	select {
	case p := <-app.promptChan:
		t.Logf("Success! Received auto-prompt: %s", p)
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Ticker failed to send prompt")
	}
}
