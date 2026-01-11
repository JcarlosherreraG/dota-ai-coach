// Package state предоставляет потокобезопасное хранилище игрового состояния.
package state

import (
	"github.com/BrightGir/game-ai-helper/internal/dota"
	"sync"
)

// Store потокобезопасное хранилище состояния игры Dota 2.
type Store struct {
	mu        sync.RWMutex
	dotaState *dota.GameState
}

// SnapShot снимок состояния игры в определённый момент времени.
type SnapShot struct {
	Game dota.GameState
}

// NewStore создаёт новое хранилище состояния.
func NewStore() *Store {
	return &Store{
		dotaState: &dota.GameState{},
	}
}

// Get возвращает копию текущего состояния игры.
// Примечание: map поля (Items, Abilities) копируются по ссылке.
func (m *Store) Get() SnapShot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return SnapShot{
		Game: *m.dotaState,
	}
}

// Update обновляет состояние игры новыми данными от GSI.
func (m *Store) Update(newState *dota.GameState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.dotaState = newState
}
