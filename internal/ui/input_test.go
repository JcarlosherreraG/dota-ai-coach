package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"testing"
	"time"
)

func TestOverlay_InputLogic(t *testing.T) {
	promptChan := make(chan string, 1)
	overlay := NewOverlay(0, 0, promptChan)

	overlay.textBoxRect = rl.Rectangle{X: 10, Y: 10, Width: 100, Height: 50}
	overlay.contextRect = rl.Rectangle{X: 200, Y: 10, Width: 100, Height: 50}

	t.Run("Mouse Click Focus", func(t *testing.T) {

		overlay.processMouseClick(rl.Vector2{X: 15, Y: 15})
		if overlay.activeField != FieldPrompt {
			t.Error("Expected FieldPrompt active")
		}

		overlay.processMouseClick(rl.Vector2{X: 500, Y: 500})
		if overlay.activeField != FieldNone {
			t.Error("Expected FieldNone active")
		}

		overlay.processMouseClick(rl.Vector2{X: 210, Y: 15})
		if overlay.activeField != FieldContext {
			t.Error("Expected FieldContext active")
		}
	})

	t.Run("Typing", func(t *testing.T) {
		overlay.activeField = FieldPrompt
		overlay.promptText = ""

		overlay.processCharInput('A')
		overlay.processCharInput('B')

		if overlay.promptText != "AB" {
			t.Errorf("Expected 'AB', got '%s'", overlay.promptText)
		}
	})

	t.Run("Backspace Logic", func(t *testing.T) {
		overlay.activeField = FieldPrompt
		overlay.promptText = "Test"

		overlay.processBackspace(true, false, 0.016)
		if overlay.promptText != "Tes" {
			t.Errorf("Expected 'Tes', got '%s'", overlay.promptText)
		}

		overlay.processBackspace(false, true, 0.1) // dt = 0.1s
		if overlay.promptText != "Tes" {
			t.Error("Should not delete yet (timer not ready)")
		}

		overlay.processBackspace(false, true, 0.5)
		if overlay.promptText != "Te" {
			t.Errorf("Should delete by timer. Got '%s'", overlay.promptText)
		}
	})

	t.Run("Submit Prompt", func(t *testing.T) {
		overlay.activeField = FieldPrompt
		overlay.promptText = "Hello AI"

		overlay.processSubmit()

		if overlay.promptText != "" {
			t.Error("Prompt text should be cleared after submit")
		}

		select {
		case msg := <-promptChan:
			if msg != "Hello AI" {
				t.Errorf("Wrong message in channel: %s", msg)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("Message not sent to channel")
		}
	})

	t.Run("Submit Context", func(t *testing.T) {
		overlay.activeField = FieldContext
		overlay.gameContext = "My Notes"

		overlay.processSubmit()

		if overlay.activeField != FieldNone {
			t.Error("Should drop focus after context submit")
		}
		if overlay.gameContext != "My Notes" {
			t.Error("Should save context text")
		}
	})
}

func TestOverlay_sendPrompt_DropOldest(t *testing.T) {
	promptChan := make(chan string, 1)
	overlay := NewOverlay(0, 0, promptChan)

	promptChan <- "old message"

	overlay.sendPrompt("new message")

	select {
	case msg := <-promptChan:
		if msg != "new message" {
			t.Errorf("Expected 'new message', got '%s'", msg)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Channel is empty, sendPrompt failed")
	}

	select {
	case <-promptChan:
		t.Error("Channel should be empty")
	default:

	}
}

func TestOverlay_deleteLastChar(t *testing.T) {
	overlay := &Overlay{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Normal delete", "Hello", "Hell"},
		{"Single char", "A", ""},
		{"Empty string", "", ""},
		{"Cyrillic", "Привет", "Приве"},
		{"Emoji/Symbols", "A😊", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.input
			overlay.deleteLastChar(&str)
			if str != tt.expected {
				t.Errorf("deleteLastChar(%q) = %q, want %q", tt.input, str, tt.expected)
			}
		})
	}

	overlay.deleteLastChar(nil)
}
