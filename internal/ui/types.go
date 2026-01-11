// Package ui предоставляет оверлей для отображения AI советов и ввода команд.
package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"sync"
)

// FieldType тип активного поля ввода.
type FieldType int

const (
	FieldNone    FieldType = iota // Нет активного поля
	FieldPrompt                   // Поле для ввода вопроса
	FieldContext                  // Поле для ввода контекста
)

// Overlay прозрачный оверлей поверх игры на базе Raylib.
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

// SetAiAdvice обновляет текст AI совета.
func (o *Overlay) SetAiAdvice(a string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.aiAdvice = a
}

// ToggleVisible переключает видимость оверлея.
func (o *Overlay) ToggleVisible() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.visible = !o.visible
}

// ToggleFocus переключает режим фокуса (для ввода текста).
func (o *Overlay) ToggleFocus() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.focused = !o.focused
}

// SetFocused устанавливает режим фокуса.
func (o *Overlay) SetFocused(b bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.focused = b
}

// GetContextText возвращает текст игрового контекста введённый пользователем.
func (o *Overlay) GetContextText() string {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.gameContext
}

// GetShouldClose возвращает флаг необходимости закрытия оверлея.
func (o *Overlay) GetShouldClose() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.shouldClose
}

// GetFocused возвращает текущий режим фокуса.
func (o *Overlay) GetFocused() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.focused
}

// GetVisible возвращает текущее состояние видимости.
func (o *Overlay) GetVisible() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.visible
}

// GetActiveField возвращает активное поле ввода.
func (o *Overlay) GetActiveField() FieldType {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.activeField
}

// Quit сигнализирует о необходимости закрытия оверлея.
func (o *Overlay) Quit() {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.shouldClose = true
}
