package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"syscall"
)

var (
	user32            = syscall.NewLazyDLL("user32.dll")
	procGetWindowLong = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLong = user32.NewProc("SetWindowLongPtrW")
)

const (
	GWL_EXSTYLE       = ^uintptr(20 - 1) // -20
	WS_EX_TRANSPARENT = 0x00000020
	WS_EX_LAYERED     = 0x00080000
)

// SetClickThrough enables or disables the click-through mode for the window (Windows-specific).
func (o *Overlay) SetClickThrough(enable bool) {
	hwnd := uintptr(rl.GetWindowHandle())

	ret, _, _ := procGetWindowLong.Call(hwnd, GWL_EXSTYLE)
	style := uintptr(ret)

	if enable {
		// Add transparent and layered styles to enable click-through
		style |= WS_EX_TRANSPARENT | WS_EX_LAYERED
	} else {
		// Remove transparent style to disable click-through
		style &^= WS_EX_TRANSPARENT
	}

	procSetWindowLong.Call(hwnd, GWL_EXSTYLE, style)
}
