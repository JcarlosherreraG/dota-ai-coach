package transport

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/BrightGir/game-ai-helper/internal/dota"
	"github.com/BrightGir/game-ai-helper/internal/state"
)

func TestGSIHandler_ServeHTTP(t *testing.T) {

	parser := dota.NewParser()
	store := state.NewStore()
	handler := NewGSIHandler(parser, store)

	t.Run("Valid GSI Payload", func(t *testing.T) {
		jsonBody := []byte(`{
			"provider": {"name": "dota"},
			"map": {"clock_time": 50},
			"hero": {"name": "npc_dota_hero_invoker"}
		}`)

		req := httptest.NewRequest("POST", "/", bytes.NewBuffer(jsonBody))

		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, http.StatusOK)
		}

		snapshot := store.Get()
		if snapshot.Game.Hero.Name != "npc_dota_hero_invoker" {
			t.Errorf("Store was not updated! Got name: %s", snapshot.Game.Hero.Name)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", bytes.NewBufferString("not json"))
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("Expected 400 Bad Request, got %d", rr.Code)
		}
	})

	t.Run("Empty Body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Expected 200 OK on empty body, got %d", rr.Code)
		}
	})
}
