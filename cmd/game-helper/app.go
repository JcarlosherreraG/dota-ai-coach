package main

import (
	"context"
	"errors"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"github.com/BrightGir/game-ai-helper/internal/config"
	"github.com/BrightGir/game-ai-helper/internal/ui"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"time"
)

// App основное приложение, координирует работу всех компонентов.
type App struct {
	cfg            *config.Config
	aiProvider     ai.Client
	advisor        AdvisorProvider
	gsiHandler     http.Handler
	server         *http.Server
	overlay        *ui.Overlay
	promptChan     chan string
	adviceChan     chan string
	userPromptChan chan string
}

// AdvisorProvider интерфейс для генерации промптов.
type AdvisorProvider interface {
	BuildPrompt() string
	BuildPromptWithQuestion(question string) string
}

func NewApp(cfg *config.Config, client ai.Client, advisor AdvisorProvider, handler http.Handler, overlay *ui.Overlay, userPromptChan chan string) *App {
	return &App{
		cfg:        cfg,
		aiProvider: client,
		advisor:    advisor,
		gsiHandler: handler,
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

func (a *App) runServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.server.Shutdown(shutdownCtx)
	}()

	log.Printf("[Server] GSI listener started on %s", a.server.Addr)
	if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("[Server] Fatal error: %v", err)
	}
	log.Println("[Server] Stopped")
}

func (a *App) run() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)

	var wg sync.WaitGroup
	defer stop()
	wg.Add(1)
	go a.runPromptHandler(ctx, &wg, time.Duration(a.cfg.SilenceDurationSeconds)*time.Second, time.Duration(a.cfg.RequestIntervalSeconds)*time.Second)
	wg.Add(1)
	go a.runAIWorker(ctx, &wg, 500*time.Millisecond, 3)
	wg.Add(1)
	go a.runAdviceConsumer(&wg)
	wg.Add(1)
	go a.runServer(ctx, &wg)

	go func() {
		<-ctx.Done()
		wg.Wait()
		if a.overlay != nil {
			a.overlay.Quit()
		}
	}()
	a.overlay.Run()
}
