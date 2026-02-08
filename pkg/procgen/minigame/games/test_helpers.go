// Package games contains test helpers for minigame testing.
package games

import (
	"image/color"

	"github.com/opd-ai/venture/pkg/engine"
)

// renderableGame is a helper interface for backward compatibility testing.
// It includes the deprecated Render method that concrete game types still support.
type renderableGame interface {
	engine.MiniGame
	// Render is the deprecated rendering method (for backward compatibility tests)
	Render(screen engine.ImageProvider) error
}

// stubScreen implements engine.ImageProvider for testing.
type stubScreen struct {
	width, height int
}

func (s *stubScreen) GetSize() (int, int)           { return s.width, s.height }
func (s *stubScreen) GetPixel(x, y int) color.Color { return color.Transparent }
