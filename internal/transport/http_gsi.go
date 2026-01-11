// Package transport предоставляет HTTP сервер для приёма данных от Dota 2 GSI.
package transport

import (
	"github.com/BrightGir/game-ai-helper/internal/dota"
	"github.com/BrightGir/game-ai-helper/internal/state"
	"io"
	"log"
	"net/http"
)

// gsiHandler обрабатывает входящие GSI запросы от Dota 2.
type gsiHandler struct {
	parser     *dota.Parser
	stateStore *state.Store
}

func NewGSIHandler(p *dota.Parser, sm *state.Store) http.Handler {
	return &gsiHandler{
		parser:     p,
		stateStore: sm,
	}
}

func (h *gsiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	if len(body) == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}
	gameState, err := h.parser.Parse(body)
	if err != nil {
		log.Printf("ERROR: failed to parse game state: %v", err)
		http.Error(w, "Could not parse game state", http.StatusBadRequest)
		return
	}
	h.stateStore.Update(gameState)

	w.WriteHeader(http.StatusOK)
}
