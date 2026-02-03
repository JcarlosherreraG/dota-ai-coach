// Package main is the entry point for the GameHelper application.
// It initializes all components and starts the main application loop.
package main

import (
	"github.com/BrightGir/game-ai-helper/assets"
	"github.com/BrightGir/game-ai-helper/internal/ai/factory"
	"github.com/BrightGir/game-ai-helper/internal/config"
	"github.com/BrightGir/game-ai-helper/internal/dota"
	"github.com/BrightGir/game-ai-helper/internal/prompt"
	"github.com/BrightGir/game-ai-helper/internal/retriever"
	"github.com/BrightGir/game-ai-helper/internal/state"
	"github.com/BrightGir/game-ai-helper/internal/transport"
	"github.com/BrightGir/game-ai-helper/internal/ui"
	"github.com/joho/godotenv"
	"log"
	"os"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("WARN: .env not found, will use system variables")
	}

	// Ensure API_KEY is set for AI provider access
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		log.Fatal("ERROR: variable API_KEY not set. Shutdown...")
	}

	// Load application configuration from config.json
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		log.Fatal("Error config initialization:", err)
	}

	// Initialize AI client (Gemini or OpenRouter) based on configuration
	aiClient, err := factory.CreateClient(cfg.Provider, cfg.Model, 15)
	if err != nil {
		log.Fatal("Error AI initialization:", err)
	}

	// Initialize BERT embedding function for RAG (Retrieval-Augmented Generation)
	bert, err := retriever.CreateEmbedFunction()
	if err != nil {
		log.Fatal("Error initializing BERT:", err)
	}

	// Initialize local RAG retriever with the pre-defined knowledge base
	retriever, err := retriever.CreateLocal(bert, cfg.MinSimilarity, assets.KnowledgeJSON)
	if err != nil {
		log.Fatal("Error initializing RAG retriever:", err)
	}

	// Setup communication channels and core components
	userPromptChan := make(chan string, 1) // Channel for user questions
	overlay := ui.NewOverlay(cfg.HotKeyTurnOverlay, cfg.HotKeyFocusOverlay, userPromptChan)
	store := state.NewStore()  // Thread-safe game state storage
	parser := dota.NewParser() // Dota 2 GSI data parser

	// Setup prompt generation pipeline with RAG support
	pipeline := prompt.NewPipeline(aiClient, retriever)
	promptBuilder := prompt.NewBuilder(store, overlay, pipeline)

	// Setup HTTP handler for Dota 2 GSI (Game State Integration)
	gsiHandler := transport.NewGSIHandler(parser, store)

	// Create and run the main application instance
	app := NewApp(cfg, store, aiClient, promptBuilder, gsiHandler, overlay, userPromptChan)
	app.run()
}
