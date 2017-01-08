package clipboard

import (
	"runtime"
)

func NewClipboard() Clipboarder {
	switch runtime.GOOS {
	case "linux":
		return NewXClip()
	case "darwin":
		return NewPBClip()
	default:
		return NewXClip()
	}
}
