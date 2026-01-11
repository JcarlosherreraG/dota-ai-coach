package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"sync"
)

// Run запускает главный цикл отрисовки оверлея.
func (o *Overlay) Run() {
	var wg sync.WaitGroup
	stopHook := make(chan struct{})
	wg.Add(1)
	go o.runKeyBoardHook(&wg, stopHook)
	defer rl.CloseWindow()

	rl.SetConfigFlags(rl.FlagWindowUndecorated | rl.FlagWindowTransparent | rl.FlagWindowTopmost | rl.FlagMsaa4xHint)

	scrW := int32(rl.GetScreenWidth())
	scrH := int32(rl.GetScreenHeight())

	rl.InitWindow(scrW, scrH, "Dota AI Overlay")
	rl.SetWindowPosition(0, 0)
	rl.SetTargetFPS(60)

	runes := o.generateCyrillicRunes()
	o.font = rl.LoadFontEx("assets/Arial.ttf", 32, runes, int32(len(runes)))
	defer rl.UnloadFont(o.font)

	o.SetClickThrough(true)

	lastFocused := false
	lastVisible := false

	for !rl.WindowShouldClose() {

		exit := o.GetShouldClose()
		if exit {
			break
		}
		currentFocused := o.GetFocused()
		currentVisible := o.GetVisible()

		if currentFocused != lastFocused || currentVisible != lastVisible {
			shouldPassThrough := !currentFocused
			o.SetClickThrough(shouldPassThrough)
			lastFocused = currentFocused
			lastVisible = currentVisible
		}

		o.calculateLayout()
		o.handleInput()

		rl.BeginDrawing()
		rl.ClearBackground(rl.Blank)

		if o.GetVisible() {
			o.renderAdviceLayer()
			o.renderInputLayer()
			o.renderContextLayer()
		}

		rl.EndDrawing()
	}

	close(stopHook)
	wg.Wait()
}

// generateCyrillicRunes генерирует набор символов ASCII + кириллицу.
func (o *Overlay) generateCyrillicRunes() []rune {
	runes := make([]rune, 0, 512)
	for i := rune(32); i <= 126; i++ {
		runes = append(runes, i)
	}
	for i := rune(1024); i <= 1105; i++ {
		runes = append(runes, i)
	} // rus
	runes = append(runes, 'ё', 'Ё')
	return runes
}
