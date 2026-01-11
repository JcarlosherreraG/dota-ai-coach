package dota

import (
	"testing"
)

func TestParser_Parse(t *testing.T) {
	parser := NewParser()
	validJSON := []byte(`{
		"map": {
			"clock_time": 100,
			"game_state": "DOTA_GAMERULES_STATE_GAME_IN_PROGRESS"
		},
		"hero": {
			"name": "npc_dota_hero_axe"
		}
	}`)

	state, err := parser.Parse(validJSON)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if state.Map.ClockTime != 100 {
		t.Errorf("Expected ClockTime 100, got %d", state.Map.ClockTime)
	}
	if state.Hero.Name != "npc_dota_hero_axe" {
		t.Errorf("Expected Hero Name 'npc_dota_hero_axe', got '%s'", state.Hero.Name)
	}

	invalidJSON := []byte(`not a json`)
	_, err = parser.Parse(invalidJSON)
	if err == nil {
		t.Error("Expected error on invalid JSON, got nil")
	}
}
