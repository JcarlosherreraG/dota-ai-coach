package coach

import (
	"fmt"
	"strings"
	"testing"

	"github.com/BrightGir/game-ai-helper/internal/dota"
	"github.com/BrightGir/game-ai-helper/internal/state"
)

type MockContextProvider struct {
	Text string
}

func (m *MockContextProvider) GetContextText() string {
	return m.Text
}

func createTestState() *dota.GameState {
	return &dota.GameState{
		Map: dota.Map{
			ClockTime: 125, // 02:05
			GameState: "DOTA_GAMERULES_STATE_GAME_IN_PROGRESS",
		},
		Player: dota.Player{
			Kills:    5,
			Deaths:   0,
			Assists:  10,
			Gold:     2500,
			LastHits: 42,
		},
		Hero: dota.Hero{
			Name:      "npc_dota_hero_pudge",
			Level:     7,
			Health:    1000,
			MaxHealth: 2000,
			Mana:      150,
			MaxMana:   300,
			Alive:     true,
		},
		Abilities: map[string]dota.Ability{
			"skill1": {Name: "hook", Level: 4, CanCast: true},
			"skill2": {Name: "rot", Level: 1, CanCast: false, Cooldown: 5},
			"ult":    {Name: "dismember", Level: 1, CanCast: true, IsUltimate: true},
		},
		Items: map[string]dota.Item{
			"slot0": {Name: "item_blink", CanCast: true},
			"slot1": {Name: "item_tango", CanCast: true, Charges: 3},
			"slot2": {Name: "item_black_king_bar", CanCast: false},
			"slot3": {Name: "item_empty"},
		},
	}
}

func TestAdvisor_BuildPrompt(t *testing.T) {
	store := state.NewStore()

	advisorEmpty := NewAdvisor(store, nil, nil)
	if p := advisorEmpty.BuildPrompt(); p != "" {
		t.Errorf("Expected empty prompt for empty state, got: %s", p)
	}

	testState := createTestState()
	store.Update(testState)

	mockCP := &MockContextProvider{Text: "Laning phase"}
	advisor := NewAdvisor(store, nil, mockCP)

	prompt := advisor.BuildPrompt()

	checks := []struct {
		name     string
		contains string
	}{
		{"Hero Name", "Hero: npc_dota_hero_pudge (Lvl 7)"},
		{"HP", "HP: 1000/2000"},
		{"Clock", "Match Time: 02:05"},
		{"KDA", "K/D/A 5/0/10"},
		{"Notes", "User Situation Notes: Laning phase"},
		{"Item Blink", "item_blink"},
		{"Item Tango", "item_tango x3"},
		{"Item BKB (CD)", "item_black_king_bar (CD)"},
		// Скиллы
		{"Skill Hook", "hook(Lvl4)[Rdy]"},
		{"Skill Rot (CD)", "rot(Lvl1)[CD:5s]"},
		{"Ult Ready", "dismember(Lvl1)[ULTIMATE-READY]"},
	}

	for _, check := range checks {
		if !strings.Contains(prompt, check.contains) {
			t.Errorf("Prompt missing %s info. Expected substring: '%s'.\nFull Prompt:\n%s",
				check.name, check.contains, prompt)
		}
	}
}

func TestAdvisor_BuildPromptWithQuestion(t *testing.T) {
	store := state.NewStore()
	store.Update(createTestState())

	advisor := NewAdvisor(store, nil, nil)

	q := "Should I buy BKB?"
	prompt := advisor.BuildPromptWithQuestion(q)

	if !strings.Contains(prompt, "[USER QUESTION]") {
		t.Error("Missing [USER QUESTION] section")
	}
	if !strings.Contains(prompt, fmt.Sprintf("\"%s\"", q)) {
		t.Error("User question text not found in prompt")
	}

	storeEmpty := state.NewStore()
	advisorEmpty := NewAdvisor(storeEmpty, nil, nil)
	fallbackPrompt := advisorEmpty.BuildPromptWithQuestion("Help?")

	if !strings.Contains(fallbackPrompt, "Game data unavailable") {
		t.Error("Should return fallback message when state is empty")
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		seconds int
		want    string
	}{
		{0, "00:00"},
		{59, "00:59"},
		{65, "01:05"},
		{3600, "60:00"},
	}
	for _, tt := range tests {
		if got := formatTime(tt.seconds); got != tt.want {
			t.Errorf("formatTime(%d) = %s, want %s", tt.seconds, got, tt.want)
		}
	}
}
