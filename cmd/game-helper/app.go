// Package main contains the core application logic and entry point.
package main

import (
	"context"
	"errors"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"github.com/BrightGir/game-ai-helper/internal/config"
	"github.com/BrightGir/game-ai-helper/internal/state"
	"github.com/BrightGir/game-ai-helper/internal/ui"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

// App is the main application, coordinates the work of all components.
type App struct {
	cfg            *config.Config // Application configuration
	aiProvider     ai.Client      // AI client for LLM interaction
	promptBuilder  PromptBuilder  // Interface for building prompts
	store          *state.Store   // Thread-safe game state storage
	gsiHandler     http.Handler   // HTTP handler for Dota 2 GSI
	server         *http.Server   // HTTP server for GSI data
	overlay        *ui.Overlay    // Graphical overlay (Raylib)
	promptChan     chan string    // Channel for prompts waiting for AI processing
	adviceChan     chan string    // Channel for AI advice waiting for UI display
	userPromptChan chan string    // Channel for user questions
}

// PromptBuilder interface for generating prompts.
type PromptBuilder interface {
	// Build constructs a prompt based on the current context and question.
	Build(ctx context.Context, question string) (string, error)
}

// NewApp creates and initializes a new App instance.
func NewApp(cfg *config.Config, store *state.Store, client ai.Client, pb PromptBuilder, handler http.Handler, overlay *ui.Overlay, userPromptChan chan string) *App {
	return &App{
		cfg:           cfg,
		aiProvider:    client,
		promptBuilder: pb,
		gsiHandler:    handler,
		store:         store,
		server: &http.Server{
			Addr:    ":" + strconv.Itoa(cfg.GsiPort),
			Handler: handler,
		},
		overlay:        overlay,
		userPromptChan: userPromptChan,
		promptChan:     make(chan string, 1),
		adviceChan:     make(chan string, 1),
	}
}

// runServer starts the HTTP server to receive GSI data from Dota 2.
func (a *App) runServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// Launch a goroutine for graceful server shutdown when context is cancelled
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.server.Shutdown(shutdownCtx)
	}()

	log.Printf("[Server] GSI listener started on %s", a.server.Addr)
	// Start listening for incoming GSI requests
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("[Server] Fatal error: %v", err)
	}
	log.Println("[Server] Stopped")
}

// run initializes and starts all application components.
func (a *App) run() {
	// Setup interrupt signal handling (Ctrl+C) for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	var wg sync.WaitGroup
	defer stop()

	// 1. Start prompt handler (manages timers and user question queue)
	wg.Add(1)
	go a.runPromptHandler(ctx, &wg, time.Duration(a.cfg.SilenceDurationSeconds)*time.Second, time.Duration(a.cfg.RequestIntervalSeconds)*time.Second)

	// 2. Start AI worker (processes prompts from queue and calls LLM)
	wg.Add(1)
	go a.runAIWorker(ctx, &wg, 500*time.Millisecond, 3)

	// 3. Start advice consumer (receives AI responses and updates overlay)
	wg.Add(1)
	go a.runAdviceConsumer(&wg)

	// 4. Start GSI server (receives real-time data from Dota 2)
	wg.Add(1)
	go a.runServer(ctx, &wg)

	// Goroutine to signal overlay to quit when the application terminates
	go func() {
		<-ctx.Done()
		wg.Wait()
		if a.overlay != nil {
			a.overlay.Quit()
		}
	}()

	// Start the main overlay rendering cycle (blocking call)
	a.overlay.Run()
}
