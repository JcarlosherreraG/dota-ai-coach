package ui

import (
	"testing"
)

func TestOverlay_handleVisibleHotkey(t *testing.T) {
	t.Run("Hide Overlay", func(t *testing.T) {
		o := NewOverlay(0, 0, nil)
		o.visible = true
		o.focused = true

		o.handleVisibleHotkey()

		if o.GetVisible() != false {
			t.Error("Overlay should be invisible now")
		}
		if o.GetFocused() != false {
			t.Error("Focus should be removed when hiding overlay")
		}
	})

	t.Run("Show Overlay", func(t *testing.T) {
		o := NewOverlay(0, 0, nil)
		o.visible = false

		o.handleVisibleHotkey()

		if o.GetVisible() != true {
			t.Error("Overlay should be visible now")
		}
	})
}

func TestOverlay_handleFocusHotkey(t *testing.T) {
	t.Run("Toggle Focus when Visible", func(t *testing.T) {
		o := NewOverlay(0, 0, nil)
		o.visible = true
		o.focused = false

		o.handleFocusHotkey()
		if o.GetFocused() != true {
			t.Error("Focus should be ON")
		}

		o.handleFocusHotkey()
		if o.GetFocused() != false {
			t.Error("Focus should be OFF")
		}
	})

	t.Run("No Focus when Hidden", func(t *testing.T) {
		o := NewOverlay(0, 0, nil)
		o.visible = false
		o.focused = false

		o.handleFocusHotkey()

		if o.GetFocused() != false {
			t.Error("Should NOT be able to focus hidden overlay")
		}
	})
}
