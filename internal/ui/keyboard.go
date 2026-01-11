package ui

import (
	"github.com/moutend/go-hook/pkg/keyboard"
	"github.com/moutend/go-hook/pkg/types"
	"log"
	"sync"
)

func (o *Overlay) runKeyBoardHook(wg *sync.WaitGroup, stopChan <-chan struct{}) {
	defer wg.Done()
	keyboardChan := make(chan types.KeyboardEvent, 100)
	err := keyboard.Install(nil, keyboardChan)
	if err != nil {
		log.Println("[ERROR] error keyboard events listening setup")
		return
	}
	defer keyboard.Uninstall()

	for {
		select {
		case <-stopChan:
			return
		case event := <-keyboardChan:
			if event.VKCode == types.VKCode(o.hotkeyTurnOverlay) && event.Message == MessageKeyDown {
				o.handleVisibleHotkey()
			}
			if event.VKCode == types.VKCode(o.hotkeyFocusOverlay) && event.Message == MessageKeyDown {
				o.handleFocusHotkey()
			}
		}
	}
}

func (o *Overlay) handleVisibleHotkey() {
	o.ToggleVisible()
	if !o.GetVisible() {
		o.SetFocused(false)
	}
}

func (o *Overlay) handleFocusHotkey() {
	if o.GetVisible() {
		o.ToggleFocus()
	} else {
		o.SetFocused(false)
	}
}
