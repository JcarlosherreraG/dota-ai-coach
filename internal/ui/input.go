package ui

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	keyRepeatDelay = 0.5  // Задержка перед повтором клавиши
	keyRepeatRate  = 0.05 // Интервал повтора клавиши
	MessageKeyDown = 256  // Код события нажатия клавиши
)

// handleInput обрабатывает пользовательский ввод с клавиатуры и мыши.
func (o *Overlay) handleInput() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		o.processMouseClick(rl.GetMousePosition())
	}

	if o.activeField == FieldNone {
		return
	}

	char := rl.GetCharPressed()
	for char > 0 {
		if char >= 32 {
			o.processCharInput(char)
		}
		char = rl.GetCharPressed()
	}

	isPressed := rl.IsKeyPressed(rl.KeyBackspace)
	isDown := rl.IsKeyDown(rl.KeyBackspace)
	dt := rl.GetFrameTime()
	o.processBackspace(isPressed, isDown, dt)

	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter) {
		o.processSubmit()
	}
}

func (o *Overlay) processMouseClick(pos rl.Vector2) {
	if rl.CheckCollisionPointRec(pos, o.textBoxRect) {
		o.activeField = FieldPrompt
	} else if rl.CheckCollisionPointRec(pos, o.contextRect) {
		o.activeField = FieldContext
	} else {
		o.activeField = FieldNone
	}
}

func (o *Overlay) processCharInput(char int32) {
	ptr := o.getActiveFieldPtr()
	if ptr != nil {
		*ptr += string(rune(char))
	}
}

func (o *Overlay) processBackspace(isPressed, isDown bool, dt float32) {
	ptr := o.getActiveFieldPtr()
	if ptr != nil {
		if isPressed {
			o.deleteLastChar(ptr)
			o.keyRepeatTimer = 0
		} else if isDown {
			o.keyRepeatTimer += dt
			if o.keyRepeatTimer >= keyRepeatDelay {
				o.deleteLastChar(ptr)
				o.keyRepeatTimer = keyRepeatDelay - keyRepeatRate
			}
		} else {
			o.keyRepeatTimer = 0
		}
	}
}

func (o *Overlay) processSubmit() {
	if o.activeField == FieldPrompt && o.promptText != "" {
		fmt.Println("Send AI prompt:", o.promptText)
		o.sendPrompt(o.promptText)
		o.promptText = ""
	}
	if o.activeField == FieldContext {
		fmt.Println("Saved game context:", o.gameContext)
		o.activeField = FieldNone
	}
}

func (o *Overlay) sendPrompt(userInput string) {
	for {
		select {
		case o.userPromptChan <- userInput:
			return
		default:
			select {
			case <-o.userPromptChan:
			default:

			}
		}
	}
}

func (o *Overlay) deleteLastChar(text *string) {
	if text == nil || *text == "" {
		return
	}
	runes := []rune(*text)
	if len(runes) > 0 {
		*text = string(runes[:len(runes)-1])
	}
}

func (o *Overlay) getActiveFieldPtr() *string {
	switch o.activeField {
	case FieldPrompt:
		return &o.promptText
	case FieldContext:
		return &o.gameContext
	default:
		return nil
	}
}
