package main

import (
	"context"
	"sync"
	"testing"
	"time"
)

type mockAdvisor struct{}

func (m *mockAdvisor) BuildPrompt() string {
	return "Auto prompt from mock"
}

func (m *mockAdvisor) BuildPromptWithQuestion(question string) string {
	return "Answer to: " + question
}

func TestPromptHandler_SilenceLogic(t *testing.T) {
	requestInterval := 100 * time.Millisecond
	silenceDuration := 30 * time.Millisecond

	mockAdvisor := &mockAdvisor{}

	app := &App{
		advisor:        mockAdvisor,
		promptChan:     make(chan string, 10),
		userPromptChan: make(chan string, 10),
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
