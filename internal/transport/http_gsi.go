// Package transport provides an HTTP server to receive data from Dota 2 GSI.
package transport

import (
	"github.com/BrightGir/game-ai-helper/internal/dota"
	"github.com/BrightGir/game-ai-helper/internal/state"
	"io"
	"log"
	"net/http"
)

// gsiHandler processes incoming GSI requests from Dota 2.
type gsiHandler struct {
	parser     *dota.Parser
	stateStore *state.Store
}

// NewGSIHandler creates a new GSI HTTP handler.
func NewGSIHandler(p *dota.Parser, sm *state.Store) http.Handler {
	return &gsiHandler{
		parser:     p,
		stateStore: sm,
	}
}

func (h *gsiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Read request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ERROR: failed to read request body: %v", err)
		http.Error(w, "Could not read request body", http.StatusInternalServerError)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("WARN: failed to close request body: %v", err)
		}
	}(r.Body)

	// Handle empty body
	if len(body) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse game state
	gameState, err := h.parser.Parse(body)
	if err != nil {
		log.Printf("ERROR: failed to parse game state: %v", err)
		http.Error(w, "Could not parse game state", http.StatusBadRequest)
		return
	}

	// Update state store
	h.stateStore.Update(gameState)

	w.WriteHeader(http.StatusOK)
}
