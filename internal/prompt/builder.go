// Package prompt is responsible for forming advice based on the game state.
package prompt

import (
	"context"
	"fmt"
	"github.com/BrightGir/game-ai-helper/internal/dota"
	"github.com/BrightGir/game-ai-helper/internal/state"
	"strings"
)

// Builder constructs prompts for AI based on the current game state.
type Builder struct {
	store       *state.Store
	pipeline    *Pipeline
	ctxProvider ContextProvider
}

// ContextProvider provides additional context for prompts (e.g., player notes).
type ContextProvider interface {
	GetContextText() string
}

// NewBuilder creates a new Builder instance.
func NewBuilder(st *state.Store, cp ContextProvider, p *Pipeline) *Builder {
	return &Builder{
		store:       st,
		pipeline:    p,
		ctxProvider: cp,
	}
}

// Build forms a prompt to answer a player's question.
func (b *Builder) Build(ctx context.Context, question string) (string, error) {
	gameContext, ok := b.buildGameContext()
	if !ok {
		gameContext = "Game context unavailable"
	}
	finalPrompt, err := b.pipeline.Execute(ctx, gameContext, question)
	return finalPrompt, err
}

// buildGameContext collects and formats game information into a single string.
func (b *Builder) buildGameContext() (string, bool) {
	// Ensure state is present
	if !b.store.HasState() {
		return "", false
	}

	s := b.store.Get()

	notes := "None"
	if b.ctxProvider != nil {
		txt := b.ctxProvider.GetContextText()
		if txt != "" {
			notes = txt
		}
	}

	// Format all game state components
	ctx := fmt.Sprintf(
		"Match Time: %s | State: %s\n"+
			"Hero: %s (Lvl %d) | HP: %d/%d | Mana: %d/%d | Alive: %v\n"+
			"Stats: K/D/A %d/%d/%d | Gold: %d | LastHits: %d\n"+
			"Inventory: [%s]\n"+
			"Abilities: [%s]\n"+
			"User Situation Notes: %s",
		formatTime(s.Game.Map.ClockTime), s.Game.Map.GameState,
		strings.TrimPrefix(s.Game.Hero.Name, "npc_dota_hero_"), s.Game.Hero.Level, s.Game.Hero.Health, s.Game.Hero.MaxHealth, s.Game.Hero.Mana, s.Game.Hero.MaxMana, s.Game.Hero.Alive,
		s.Game.Player.Kills, s.Game.Player.Deaths, s.Game.Player.Assists, s.Game.Player.Gold, s.Game.Player.LastHits,
		formatItems(s.Game.Items),
		formatAbilities(s.Game.Abilities),
		notes,
	)

	return ctx, true
}

// formatItems converts the items map into a readable string.
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

// formatAbilities converts the abilities map into a readable string.
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

// formatTime converts seconds into MM:SS format.
func formatTime(seconds int) string {
	minutes := seconds / 60
	secs := seconds % 60
	return fmt.Sprintf("%02d:%02d", minutes, secs)
}
