// Package coach отвечает за формирование советов на основе игрового состояния.
package coach

import (
	"fmt"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"github.com/BrightGir/game-ai-helper/internal/dota"
	"github.com/BrightGir/game-ai-helper/internal/state"
	"strings"
)

// Advisor формирует промпты для AI на основе текущего состояния игры.
type Advisor struct {
	store       *state.Store
	ai          ai.Client
	ctxProvider ContextProvider
}

// ContextProvider предоставляет дополнительный контекст для промптов (например, заметки игрока).
type ContextProvider interface {
	GetContextText() string
}

func NewAdvisor(st *state.Store, ai ai.Client, cp ContextProvider) *Advisor {
	return &Advisor{
		store:       st,
		ai:          ai,
		ctxProvider: cp,
	}
}

func (a *Advisor) buildGameContext() (string, bool) {
	s := a.store.Get()

	if s.Game.Hero.Name == "" {
		return "", false
	}

	notes := "None"
	if a.ctxProvider != nil {
		txt := a.ctxProvider.GetContextText()
		if txt != "" {
			notes = txt
		}
	}

	context := fmt.Sprintf(
		"Match Time: %s | State: %s\n"+
			"Hero: %s (Lvl %d) | HP: %d/%d | Mana: %d/%d | Alive: %v\n"+
			"Stats: K/D/A %d/%d/%d | Gold: %d | LastHits: %d\n"+
			"Inventory: [%s]\n"+
			"Abilities: [%s]\n"+
			"User Situation Notes: %s",
		formatTime(s.Game.Map.ClockTime), s.Game.Map.GameState,
		s.Game.Hero.Name, s.Game.Hero.Level, s.Game.Hero.Health, s.Game.Hero.MaxHealth, s.Game.Hero.Mana, s.Game.Hero.MaxMana, s.Game.Hero.Alive,
		s.Game.Player.Kills, s.Game.Player.Deaths, s.Game.Player.Assists, s.Game.Player.Gold, s.Game.Player.LastHits,
		formatItems(s.Game.Items),
		formatAbilities(s.Game.Abilities),
		notes,
	)

	return context, true
}

// BuildPrompt формирует промпт для автоматического совета.
func (a *Advisor) BuildPrompt() string {
	ctx, ok := a.buildGameContext()
	if !ok {
		return ""
	}

	return fmt.Sprintf(
		"[GAME STATE]\n%s\n\n"+
			"[TASK]\n"+
			"Analyze the game state. Give 1 most important tactical advice for the next 30 seconds. Be concise.",
		ctx,
	)
}

// BuildPromptWithQuestion формирует промпт для ответа на вопрос игрока.
func (a *Advisor) BuildPromptWithQuestion(question string) string {
	ctx, ok := a.buildGameContext()
	if !ok {
		return fmt.Sprintf("Game data unavailable. User Question: %s", question)
	}

	return fmt.Sprintf(
		"[GAME STATE]\n%s\n\n"+
			"[USER QUESTION]\n\"%s\"\n\n"+
			"[TASK]\n"+
			"Answer the user's question based on the game state (items, cooldowns, health). Keep it under 2 sentences.",
		ctx, question,
	)
}
func formatTime(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}

func formatItems(items map[string]dota.Item) string {
	var activeItems []string
	for _, item := range items {
		if item.Name == "item_empty" || item.Name == "" {
			continue
		}

		info := item.Name
		if !item.CanCast {
			info += " (CD)" // Cooldown
		}
		if item.Charges > 0 {
			info += fmt.Sprintf(" x%d", item.Charges)
		}
		activeItems = append(activeItems, info)
	}
	if len(activeItems) == 0 {
		return "Empty Inventory"
	}
	return strings.Join(activeItems, ", ")
}

func formatAbilities(abilities map[string]dota.Ability) string {
	var skillList []string
	for _, ab := range abilities {
		if ab.Name == "" || ab.Level == 0 {
			continue
		}

		status := "Rdy"
		if ab.Cooldown > 0 {
			status = fmt.Sprintf("CD:%ds", ab.Cooldown)
		} else if !ab.CanCast {
			status = "NoMana/Silenced"
		} else if ab.IsUltimate {
			status = "ULTIMATE-READY"
		}

		skillList = append(skillList, fmt.Sprintf("%s(Lvl%d)[%s]", ab.Name, ab.Level, status))
	}
	return strings.Join(skillList, " | ")
}
