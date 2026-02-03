package ui

import (
	"context"
	"github.com/moutend/go-hook/pkg/keyboard"
	"github.com/moutend/go-hook/pkg/types"
	"log"
	"sync"
)

// runKeyBoardHook installs a global keyboard hook to listen for hotkeys.
func (o *Overlay) runKeyBoardHook(wg *sync.WaitGroup, stopCtx context.Context) {
	defer wg.Done()
	keyboardChan := make(chan types.KeyboardEvent, 100)

	// Install global hook
	err := keyboard.Install(nil, keyboardChan)
	if err != nil {
		log.Println("[ERROR] error keyboard events listening setup")
		return
	}
	defer keyboard.Uninstall()

	for {
		select {
		case <-stopCtx.Done():
			return
		case event := <-keyboardChan:
			// Check if the hotkey for toggling overlay visibility was pressed
			if event.VKCode == types.VKCode(o.hotkeyTurnOverlay) && event.Message == MessageKeyDown {
				o.handleVisibleHotkey()
			}
			// Check if the hotkey for focusing overlay input was pressed
			if event.VKCode == types.VKCode(o.hotkeyFocusOverlay) && event.Message == MessageKeyDown {
				o.handleFocusHotkey()
			}
		}
	}
}

// handleVisibleHotkey toggles overlay visibility and resets focus if hidden.
func (o *Overlay) handleVisibleHotkey() {
	o.ToggleVisible()
	if !o.GetVisible() {
		o.SetFocused(false)
	}
}

// handleFocusHotkey toggles input focus only if the overlay is visible.
func (o *Overlay) handleFocusHotkey() {
	if o.GetVisible() {
		o.ToggleFocus()
	} else {
		o.SetFocused(false)
	}
}
