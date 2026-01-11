package state

import (
	"sync"
	"testing"

	"github.com/BrightGir/game-ai-helper/internal/dota"
)

func TestStore_UpdateAndGet(t *testing.T) {
	store := NewStore()

	newState := &dota.GameState{
		Hero: dota.Hero{Name: "Pudge"},
	}

	store.Update(newState)

	snapshot := store.Get()
	if snapshot.Game.Hero.Name != "Pudge" {
		t.Errorf("Expected Pudge, got %s", snapshot.Game.Hero.Name)
	}
}

func TestStore_Concurrency(t *testing.T) {
	store := NewStore()
	wg := &sync.WaitGroup{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			store.Update(&dota.GameState{Hero: dota.Hero{Name: "Some hero"}})
		}()
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = store.Get()
		}()
	}

	wg.Wait()
}
