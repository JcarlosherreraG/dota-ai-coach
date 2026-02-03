// Package state provides a thread-safe storage for the game state.
package state

import (
	"github.com/BrightGir/game-ai-helper/internal/dota"
	"sync"
)

// Store is a thread-safe storage for the Dota 2 game state.
type Store struct {
	mu        sync.RWMutex
	dotaState *dota.GameState
}

// SnapShot is a snapshot of the game state at a specific point in time.
type SnapShot struct {
	Game dota.GameState
}

// NewStore creates a new state storage instance.
func NewStore() *Store {
	return &Store{
		dotaState: &dota.GameState{},
	}
}

// Get returns a copy of the current game state.
func (m *Store) Get() SnapShot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return SnapShot{
		Game: *m.dotaState,
	}
}

// HasState checks if the game state is initialized (hero name is set).
func (m *Store) HasState() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.dotaState.Hero.Name != ""
}

// Update updates the game state with new data from GSI.
func (m *Store) Update(newState *dota.GameState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dotaState = newState
}
