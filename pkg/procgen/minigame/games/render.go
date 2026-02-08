// Package games contains implementations of mini-game types for Venture.
// This file defines render output types and shared rendering utilities
// for Phase 27.3: Mini-Game Rendering.
package games

import (
	"fmt"

	"github.com/opd-ai/venture/pkg/engine"
)

// RenderOutput represents the computed visual state of a minigame.
// The ECS render system reads this data to draw actual pixels to the screen.
// Each PrepareRender() call populates this struct with the current visual state.
//
// Implements engine.MiniGameRenderOutput interface.
//
// Phase 27.3: Mini-Game Rendering
type RenderOutput struct {
	// Title is the display name of the minigame
	Title string
	// Status describes the current game state ("Playing", "Won", "Lost")
	Status string
	// Width is the available screen width in pixels
	Width int
	// Height is the available screen height in pixels
	Height int
	// Elements contains all visual elements to be drawn
	Elements []RenderElement
}

// GetTitle implements engine.MiniGameRenderOutput interface.
func (r *RenderOutput) GetTitle() string {
	return r.Title
}

// GetStatus implements engine.MiniGameRenderOutput interface.
func (r *RenderOutput) GetStatus() string {
	return r.Status
}

// GetDimensions implements engine.MiniGameRenderOutput interface.
func (r *RenderOutput) GetDimensions() (width, height int) {
	return r.Width, r.Height
}

// GetElements implements engine.MiniGameRenderOutput interface.
func (r *RenderOutput) GetElements() interface{} {
	return r.Elements
}

// RenderElement represents a single visual element in the minigame display.
// Elements are drawn in order, allowing layered composition.
//
// Phase 27.3: Mini-Game Rendering
type RenderElement struct {
	// Type identifies the element kind: "text", "rect", "progress", "card", "die", "tile", "pin", "symbol", "terminal"
	Type string
	// X is the horizontal position in pixels from top-left
	X int
	// Y is the vertical position in pixels from top-left
	Y int
	// W is the element width in pixels
	W int
	// H is the element height in pixels
	H int
	// Label is the display text for this element
	Label string
	// Value is a numeric value (0.0-1.0 for progress, card value, etc.)
	Value float64
	// Highlighted indicates this element should be visually emphasized
	Highlighted bool
}

// validateScreen checks that the screen parameter is valid for rendering.
// Returns an error if the screen is nil or has zero dimensions.
func validateScreen(screen engine.ImageProvider) (int, int, error) {
	if screen == nil {
		return 0, 0, fmt.Errorf("screen is nil")
	}
	w, h := screen.GetSize()
	if w <= 0 || h <= 0 {
		return 0, 0, fmt.Errorf("invalid screen dimensions: %dx%d", w, h)
	}
	return w, h, nil
}

// validateScreenDimensions checks that screen dimensions are valid for rendering.
// Returns an error if dimensions are zero or negative.
func validateScreenDimensions(width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("invalid screen dimensions: %dx%d", width, height)
	}
	return nil
}
