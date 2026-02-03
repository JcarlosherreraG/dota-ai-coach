package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// calculateLayout calculates the positions and sizes of overlay elements on the screen.
func (o *Overlay) calculateLayout() {
	screenWidth := float32(rl.GetScreenWidth())
	screenHeight := float32(rl.GetScreenHeight())

	// Set total width as 18% of screen width
	o.totalWidth = screenWidth * 0.18

	// Set total height as 45% of screen height
	o.totalHeight = screenHeight * 0.45

	// Enforce minimum dimensions
	if o.totalWidth < 250 {
		o.totalWidth = 250
	}
	if o.totalHeight < 300 {
		o.totalHeight = 300
	}

	leftX := float32(20)

	// Center vertically
	leftStartY := (screenHeight - o.totalHeight) / 2

	gap := float32(10)

	// Advice panel takes 70% of total height
	adviceH := (o.totalHeight * 0.70) - (gap / 2)

	// Prompt panel takes the remaining height
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

	// Context panel height
	contextH := o.totalHeight * 0.4

	rightX := screenWidth - o.totalWidth - 20

	// Center vertically
	contextStartY := (screenHeight - contextH) / 2

	o.contextRect = rl.Rectangle{
		X:      rightX,
		Y:      contextStartY,
		Width:  o.totalWidth,
		Height: contextH,
	}
}
