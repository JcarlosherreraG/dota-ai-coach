package main

import (
	"github.com/BrightGir/game-ai-helper/internal/ai/factory"
	"github.com/BrightGir/game-ai-helper/internal/coach"
	"github.com/BrightGir/game-ai-helper/internal/config"
	"github.com/BrightGir/game-ai-helper/internal/dota"
	"github.com/BrightGir/game-ai-helper/internal/state"
	"github.com/BrightGir/game-ai-helper/internal/transport"
	"github.com/BrightGir/game-ai-helper/internal/ui"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("WARN: .env not found, will use system variables")
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("ERROR: variable API_KEY not set. Shutdown...")
	}

	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal("Error config initialization:", err)
	}

	aiClient, err := factory.CreateClient(cfg.Provider, cfg.Model, cfg.SystemPrompt)
	if err != nil {
		log.Fatal("Error AI initialization:", err)
	}

	userPromptChan := make(chan string, 1)
	overlay := ui.NewOverlay(cfg.HotKeyTurnOverlay, cfg.HotKeyFocusOverlay, userPromptChan)
	stateMgr := state.NewStore()
	parser := dota.NewParser()
	advisor := coach.NewAdvisor(stateMgr, aiClient, overlay)
	gsiHandler := transport.NewGSIHandler(parser, stateMgr)

	app := NewApp(cfg, aiClient, advisor, gsiHandler, overlay, userPromptChan)
	app.run()
}
