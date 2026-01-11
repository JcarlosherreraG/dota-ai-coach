package ui

import (
	"sync"
	"testing"
)

func TestOverlay_StateConcurrency(t *testing.T) {
	overlay := NewOverlay(0, 0, nil)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			overlay.ToggleVisible()
			overlay.SetAiAdvice("New advice")
			_ = overlay.GetVisible()
			_ = overlay.GetContextText()
		}()
	}

	wg.Wait()
}
