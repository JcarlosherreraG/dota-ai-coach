package ui

import (
	"fmt"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	keyRepeatDelay = 0.5  // Delay before key repeat starts
	keyRepeatRate  = 0.05 // Interval between key repeats
	MessageKeyDown = 256  // Key down event code
)

// handleInput processes user input from keyboard and mouse.
func (o *Overlay) handleInput() {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Handle mouse clicks
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		o.processMouseClick(rl.GetMousePosition())
	}

	// If no field is active, skip keyboard input
	if o.activeField == FieldNone {
		return
	}

	// Handle character input
	char := rl.GetCharPressed()
	for char > 0 {
		if char >= 32 {
			o.processCharInput(char)
		}
		char = rl.GetCharPressed()
	}

	// Handle backspace with repeat logic
	isPressed := rl.IsKeyPressed(rl.KeyBackspace)
	isDown := rl.IsKeyDown(rl.KeyBackspace)
	dt := rl.GetFrameTime()
	o.processBackspace(isPressed, isDown, dt)

	// Handle submit (Enter)
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter) {
		o.processSubmit()
	}
}

// processMouseClick determines which field was clicked.
func (o *Overlay) processMouseClick(pos rl.Vector2) {
	if rl.CheckCollisionPointRec(pos, o.textBoxRect) {
		o.activeField = FieldPrompt
	} else if rl.CheckCollisionPointRec(pos, o.contextRect) {
		o.activeField = FieldContext
	} else {
		o.activeField = FieldNone
	}
}

// processCharInput appends a character to the active field.
func (o *Overlay) processCharInput(char int32) {
	ptr := o.getActiveFieldPtr()
	if ptr != nil {
		*ptr += string(rune(char))
	}
}

// processBackspace handles backspace logic including repeat.
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

// processSubmit handles Enter key press for active fields.
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

// sendPrompt sends user input to the prompt channel.
func (o *Overlay) sendPrompt(userInput string) {
	for {
		select {
		case o.userPromptChan <- userInput:
			return
		default:
			// If channel is full, drop the oldest message
			select {
			case <-o.userPromptChan:
			default:

			}
		}
	}
}

// deleteLastChar removes the last character from a string (handles multi-byte runes).
func (o *Overlay) deleteLastChar(text *string) {
	if text == nil || *text == "" {
		return
	}
	runes := []rune(*text)
	if len(runes) > 0 {
		*text = string(runes[:len(runes)-1])
	}
}

// getActiveFieldPtr returns a pointer to the string of the currently active field.
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
