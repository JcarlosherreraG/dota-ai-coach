package ui

import (
	"context"
	_ "embed"
	rl "github.com/gen2brain/raylib-go/raylib"
	"sync"
)

//go:embed resources/Arial.ttf
var fontData []byte

// Run starts the main overlay rendering cycle.
func (o *Overlay) Run() {
	var wg sync.WaitGroup
	stopHookCtx, stopHookCancel := context.WithCancel(context.Background())
	wg.Add(1)
	// Start keyboard hook in a separate goroutine
	go o.runKeyBoardHook(&wg, stopHookCtx)
	defer rl.CloseWindow()

	// Configure window flags: undecorated, transparent, topmost
	rl.SetConfigFlags(rl.FlagWindowUndecorated | rl.FlagWindowTransparent | rl.FlagWindowTopmost | rl.FlagMsaa4xHint)

	scrW := int32(rl.GetScreenWidth())
	scrH := int32(rl.GetScreenHeight())

	// Initialize Raylib window
	rl.InitWindow(scrW, scrH, "Dota AI Overlay")
	rl.SetWindowPosition(0, 0)
	rl.SetTargetFPS(60)

	// Load font with Cyrillic support
	runes := o.generateCyrillicRunes()
	o.font = rl.LoadFontFromMemory(".ttf", fontData, 32, runes)
	defer rl.UnloadFont(o.font)

	// By default, clicks pass through the window
	o.SetClickThrough(true)

	lastFocused := false
	lastVisible := false

	// Main rendering loop
	for !rl.WindowShouldClose() {
		// Check close flag from other components
		exit := o.GetShouldClose()
		if exit {
			break
		}

		currentFocused := o.GetFocused()
		currentVisible := o.GetVisible()

		// Update ClickThrough mode only when focus or visibility state changes
		if currentFocused != lastFocused || currentVisible != lastVisible {
			shouldPassThrough := !currentFocused
			o.SetClickThrough(shouldPassThrough)
			lastFocused = currentFocused
			lastVisible = currentVisible
		}

		// Calculate layout and handle input
		o.calculateLayout()
		o.handleInput()

		// Frame rendering
		rl.BeginDrawing()
		rl.ClearBackground(rl.Blank) // Transparent background

		if o.GetVisible() {
			o.renderAdviceLayer()
			o.renderInputLayer()
			o.renderContextLayer()
		}

		rl.EndDrawing()
	}

	// Terminate keyboard hook
	stopHookCancel()
	wg.Wait()
}

// generateCyrillicRunes generates a set of ASCII + Cyrillic characters for font loading.
func (o *Overlay) generateCyrillicRunes() []rune {
	runes := make([]rune, 0, 512)
	// ASCII
	for i := rune(32); i <= 126; i++ {
		runes = append(runes, i)
	}
	// Cyrillic (main block)
	for i := rune(1024); i <= 1105; i++ {
		runes = append(runes, i)
	}
	// Letters ё and Ё
	runes = append(runes, 'ё', 'Ё')
	return runes
}
