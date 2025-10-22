// Package ui provides UI type definitions.
// This file defines UI element data structures, layout parameters,
// and styling options used by the UI generator.
package ui

import (
	"fmt"
)

// ElementType represents different types of UI elements.
type ElementType int

const (
	// ElementButton represents a clickable button
	ElementButton ElementType = iota
	// ElementPanel represents a container panel
	ElementPanel
	// ElementHealthBar represents a health/progress bar
	ElementHealthBar
	// ElementLabel represents a text label
	ElementLabel
	// ElementIcon represents a small icon
	ElementIcon
	// ElementFrame represents a decorative frame
	ElementFrame
)

// String returns the string representation of an element type.
func (e ElementType) String() string {
	switch e {
	case ElementButton:
		return "button"
	case ElementPanel:
		return "panel"
	case ElementHealthBar:
		return "healthbar"
	case ElementLabel:
		return "label"
	case ElementIcon:
		return "icon"
	case ElementFrame:
		return "frame"
	default:
		return "unknown"
	}
}

// Config contains parameters for UI element generation.
type Config struct {
	// Type of UI element
	Type ElementType

	// Width and height in pixels
	Width  int
	Height int

	// GenreID for visual styling
	GenreID string

	// Seed for deterministic generation
	Seed int64

	// Text content (for buttons, labels)
	Text string

	// Value for progress bars (0.0 - 1.0)
	Value float64

	// State of the element (normal, hover, pressed, disabled)
	State ElementState

	// Custom parameters for specific element types
	Custom map[string]interface{}
}

// ElementState represents the current state of a UI element.
type ElementState int

const (
	// StateNormal is the default state
	StateNormal ElementState = iota
	// StateHover when cursor is over the element
	StateHover
	// StatePressed when element is being clicked
	StatePressed
	// StateDisabled when element is not interactive
	StateDisabled
)

// String returns the string representation of an element state.
func (s ElementState) String() string {
	switch s {
	case StateNormal:
		return "normal"
	case StateHover:
		return "hover"
	case StatePressed:
		return "pressed"
	case StateDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// DefaultConfig returns a default UI element configuration.
func DefaultConfig() Config {
	return Config{
		Type:    ElementButton,
		Width:   100,
		Height:  30,
		GenreID: "fantasy",
		Seed:    0,
		Text:    "",
		Value:   1.0,
		State:   StateNormal,
		Custom:  make(map[string]interface{}),
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.Width <= 0 {
		return fmt.Errorf("width must be positive, got %d", c.Width)
	}
	if c.Height <= 0 {
		return fmt.Errorf("height must be positive, got %d", c.Height)
	}
	if c.GenreID == "" {
		return fmt.Errorf("genreID cannot be empty")
	}
	if c.Value < 0.0 || c.Value > 1.0 {
		return fmt.Errorf("value must be between 0.0 and 1.0, got %f", c.Value)
	}
	return nil
}

// BorderStyle represents different border rendering styles.
type BorderStyle int

const (
	// BorderSolid is a simple solid border
	BorderSolid BorderStyle = iota
	// BorderDouble is a double-line border
	BorderDouble
	// BorderOrnate is a decorative border with corners
	BorderOrnate
	// BorderGlow is a glowing border effect
	BorderGlow
)

// String returns the string representation of a border style.
func (b BorderStyle) String() string {
	switch b {
	case BorderSolid:
		return "solid"
	case BorderDouble:
		return "double"
	case BorderOrnate:
		return "ornate"
	case BorderGlow:
		return "glow"
	default:
		return "unknown"
	}
}
