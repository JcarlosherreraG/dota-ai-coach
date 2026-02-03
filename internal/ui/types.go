// Package ui provides an overlay for displaying AI hints and entering commands.
package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"sync"
)

// FieldType type of active field
type FieldType int

const (
	FieldNone    FieldType = iota // No active field
	FieldPrompt                   // Input field for question
	FieldContext                  // Input field for context
)

// Overlay transparent overlay on top of the Raylib-based game.
type Overlay struct {
	hotkeyTurnOverlay  int
	hotkeyFocusOverlay int
	aiAdvice           string
	promptText         string
	visible            bool
	gameContext        string
	activeField        FieldType
	width              int
	height             int
	cursorCounter      int
	focused            bool
	shouldClose        bool
	font               rl.Font
	adviceRect         rl.Rectangle
	textBoxRect        rl.Rectangle
	contextRect        rl.Rectangle
	totalWidth         float32
	totalHeight        float32
	userPromptChan     chan string
	keyRepeatTimer     float32
	mu                 sync.RWMutex
}

func NewOverlay(turnHotkey int, focusHotkey int, userPromptChan chan string) *Overlay {
	return &Overlay{
		visible:            true,
		hotkeyTurnOverlay:  turnHotkey,
		hotkeyFocusOverlay: focusHotkey,
		userPromptChan:     userPromptChan,
		aiAdvice:           "AI Ready.",
	}
}

// SetAiAdvice updates the AI board text.
func (o *Overlay) SetAiAdvice(a string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.aiAdvice = a
}

// ToggleVisible switches overlay visibility.
func (o *Overlay) ToggleVisible() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.visible = !o.visible
}

// ToggleFocus switches focus mode (for text input).
func (o *Overlay) ToggleFocus() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.focused = !o.focused
}

// SetFocused sets the focus mode.
func (o *Overlay) SetFocused(b bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.focused = b
}

// GetContextText returns the text of the game context entered by the user.
func (o *Overlay) GetContextText() string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.gameContext
}

// GetShouldClose returns the flag to close the overlay.
func (o *Overlay) GetShouldClose() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.shouldClose
}

// GetFocused returns the current focus mode.
func (o *Overlay) GetFocused() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.focused
}

// GetVisible returns the current visibility state.
func (o *Overlay) GetVisible() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.visible
}

// GetActiveField returns the active input field.
func (o *Overlay) GetActiveField() FieldType {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.activeField
}

// Quit signals the need to close the overlay.
func (o *Overlay) Quit() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.shouldClose = true
}
