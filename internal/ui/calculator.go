package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// calculateLayout вычисляет позиции и размеры элементов оверлея на экране.
func (o *Overlay) calculateLayout() {
	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())

	o.totalWidth = screenWidth * 0.18

	o.totalHeight = screenHeight * 0.45

	if o.totalWidth < 250 {
		o.totalWidth = 250
	}
	if o.totalHeight < 300 {
		o.totalHeight = 300
	}

	leftX := float32(20)

	leftStartY := (screenHeight - o.totalHeight) / 2

	gap := float32(10)

	adviceH := (o.totalHeight * 0.70) - (gap / 2)

	promptH := o.totalHeight - adviceH - gap

	o.adviceRect = rl.Rectangle{
		X:      leftX,
		Y:      leftStartY,
		Width:  o.totalWidth,
		Height: adviceH,
	}

	o.textBoxRect = rl.Rectangle{
		X:      leftX,
		Y:      leftStartY + adviceH + gap,
		Width:  o.totalWidth,
		Height: promptH,
	}

	contextH := o.totalHeight * 0.4

	rightX := screenWidth - o.totalWidth - 20

	contextStartY := (screenHeight - contextH) / 2

	o.contextRect = rl.Rectangle{
		X:      rightX,
		Y:      contextStartY,
		Width:  o.totalWidth,
		Height: contextH,
	}
}
