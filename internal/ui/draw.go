package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// renderAdviceLayer отрисовывает верхнюю плашку с советом AI.
func (o *Overlay) renderAdviceLayer() {
	rl.DrawRectangleRec(o.adviceRect, rl.NewColor(20, 20, 30, 220))
	rl.DrawRectangleLinesEx(o.adviceRect, 2, rl.NewColor(0, 255, 255, 255))

	padding := float32(10)

	titleSize := float32(16)
	rl.DrawTextEx(o.font, "AI STRATEGIST", rl.Vector2{X: o.adviceRect.X + padding, Y: o.adviceRect.Y + padding}, titleSize, 1, rl.NewColor(0, 255, 255, 150))

	textSize := int32(16)
	textY := int32(o.adviceRect.Y + padding + titleSize + 10)

	o.drawWrappedText(o.aiAdvice, int32(o.adviceRect.X+padding), textY, int32(o.adviceRect.Width-padding*2), textSize, rl.Yellow)
}

// renderInputLayer отрисовывает нижнюю плашку для ввода вопроса.
func (o *Overlay) renderInputLayer() {
	borderColor := rl.NewColor(100, 100, 120, 255)
	if o.GetActiveField() == FieldPrompt {
		borderColor = rl.NewColor(0, 255, 0, 255)
	}

	rl.DrawRectangleRec(o.textBoxRect, rl.NewColor(20, 20, 30, 220))
	rl.DrawRectangleLinesEx(o.textBoxRect, 2, borderColor)

	padding := float32(10)

	headerSize := float32(12)
	rl.DrawTextEx(o.font, "YOUR PROMPT:", rl.Vector2{X: o.textBoxRect.X + padding, Y: o.textBoxRect.Y + padding}, headerSize, 1, rl.Gray)

	displayText := o.promptText
	if o.GetActiveField() == FieldPrompt {
		o.cursorCounter++
		if (o.cursorCounter/30)%2 == 0 {
			displayText += "|"
		}
	}

	textY := int32(o.textBoxRect.Y + padding + headerSize + 5)
	textSize := int32(16)

	color := rl.White
	if o.promptText == "" && o.GetActiveField() != FieldPrompt {
		displayText = "Click to type..."
		color = rl.DarkGray
	}

	o.drawWrappedText(displayText, int32(o.textBoxRect.X+padding), textY, int32(o.textBoxRect.Width-padding*2), textSize, color)
}

// renderContextLayer отрисовывает правую плашку для ввода игрового контекста.
func (o *Overlay) renderContextLayer() {
	borderColor := rl.NewColor(80, 80, 100, 255)
	if o.GetActiveField() == FieldContext {
		borderColor = rl.Green
	}

	rl.DrawRectangleRec(o.contextRect, rl.NewColor(30, 30, 45, 220))
	rl.DrawRectangleLinesEx(o.contextRect, 2, borderColor)

	padding := float32(10)

	headerSize := float32(14)
	rl.DrawTextEx(o.font, "CONTEXT:", rl.Vector2{X: o.contextRect.X + padding, Y: o.contextRect.Y + padding}, headerSize, 1, rl.SkyBlue)

	text := o.gameContext
	if text == "" && o.GetActiveField() != FieldContext {
		text = "Enemy items, roles, etc..."
	}

	o.drawWrappedText(text, int32(o.contextRect.X+padding), int32(o.contextRect.Y+30), int32(o.contextRect.Width-padding*2), 14, rl.LightGray)
}

// drawWrappedText отрисовывает текст с переносом строк.
func (o *Overlay) drawWrappedText(text string, x, y, maxWidth, fontSize int32, color rl.Color) {
	if text == "" {
		return
	}

	runes := []rune(text)
	var currentLine string
	currentY := float32(y)
	lineHeight := float32(fontSize + 4)
	spacing := float32(1)

	for i := 0; i < len(runes); i++ {
		char := string(runes[i])

		if char == "\n" {
			rl.DrawTextEx(o.font, currentLine, rl.Vector2{X: float32(x), Y: currentY}, float32(fontSize), spacing, color)
			currentLine = ""
			currentY += lineHeight
			continue
		}

		testLine := currentLine + char
		textSize := rl.MeasureTextEx(o.font, testLine, float32(fontSize), spacing)

		if int32(textSize.X) > maxWidth {
			rl.DrawTextEx(o.font, currentLine, rl.Vector2{X: float32(x), Y: currentY}, float32(fontSize), spacing, color)
			currentY += lineHeight
			currentLine = char
		} else {
			currentLine = testLine
		}
	}
	rl.DrawTextEx(o.font, currentLine, rl.Vector2{X: float32(x), Y: currentY}, float32(fontSize), spacing, color)
}
