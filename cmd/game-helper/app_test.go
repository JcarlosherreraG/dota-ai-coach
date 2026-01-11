package main

import (
	"context"
	"github.com/BrightGir/game-ai-helper/internal/config"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestApp_runServer_GracefulShutdown(t *testing.T) {
	cfg := &config.Config{GsiPort: 45678}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	app := &App{
		cfg:        cfg,
		gsiHandler: handler,
		server: &http.Server{
			Addr:    ":" + strconv.Itoa(cfg.GsiPort),
			Handler: handler,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	wg := &sync.WaitGroup{}
	wg.Add(1)

	go app.runServer(ctx, wg)

	time.Sleep(100 * time.Millisecond)

	go func() {
		resp, err := http.Get("http://localhost:45678")
		if err != nil {
			t.Logf("Warning: Server didn't respond: %v", err)
			return
		}
		resp.Body.Close()
	}()

	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:

	case <-time.After(6 * time.Second):
		t.Fatal("Server failed to shutdown gracefully within timeout")
	}
}
