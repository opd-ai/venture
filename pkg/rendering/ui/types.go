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

	// HierarchyLevel for visual importance (affects sizing, emphasis)
	HierarchyLevel HierarchyLevel

	// Transition for animation effects
	Transition *TransitionConfig

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
		Type:           ElementButton,
		Width:          100,
		Height:         30,
		GenreID:        "fantasy",
		Seed:           0,
		Text:           "",
		Value:          1.0,
		State:          StateNormal,
		HierarchyLevel: HierarchySecondary,
		Transition:     nil, // No transition by default
		Custom:         make(map[string]interface{}),
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
	// BorderDashed is a dashed border pattern
	BorderDashed
	// BorderDotted is a dotted border pattern
	BorderDotted
	// BorderEmbossed is a 3D embossed border effect
	BorderEmbossed
	// BorderEngraved is a 3D engraved border effect
	BorderEngraved
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
	case BorderDashed:
		return "dashed"
	case BorderDotted:
		return "dotted"
	case BorderEmbossed:
		return "embossed"
	case BorderEngraved:
		return "engraved"
	default:
		return "unknown"
	}
}

// TransitionType represents different animation transition types.
type TransitionType int

const (
	// TransitionNone is no animation
	TransitionNone TransitionType = iota
	// TransitionFade is a fade in/out animation
	TransitionFade
	// TransitionSlideLeft slides element from right to left
	TransitionSlideLeft
	// TransitionSlideRight slides element from left to right
	TransitionSlideRight
	// TransitionSlideUp slides element from bottom to top
	TransitionSlideUp
	// TransitionSlideDown slides element from top to bottom
	TransitionSlideDown
	// TransitionZoom zooms element in/out
	TransitionZoom
)

// String returns the string representation of a transition type.
func (t TransitionType) String() string {
	switch t {
	case TransitionNone:
		return "none"
	case TransitionFade:
		return "fade"
	case TransitionSlideLeft:
		return "slide-left"
	case TransitionSlideRight:
		return "slide-right"
	case TransitionSlideUp:
		return "slide-up"
	case TransitionSlideDown:
		return "slide-down"
	case TransitionZoom:
		return "zoom"
	default:
		return "unknown"
	}
}

// EasingFunction represents different easing functions for animations.
type EasingFunction int

const (
	// EaseLinear is a linear easing (constant speed)
	EaseLinear EasingFunction = iota
	// EaseInQuad is a quadratic ease-in (slow start)
	EaseInQuad
	// EaseOutQuad is a quadratic ease-out (slow end)
	EaseOutQuad
	// EaseInOutQuad is a quadratic ease-in-out (slow start and end)
	EaseInOutQuad
	// EaseInCubic is a cubic ease-in (very slow start)
	EaseInCubic
	// EaseOutCubic is a cubic ease-out (very slow end)
	EaseOutCubic
	// EaseInOutCubic is a cubic ease-in-out (very slow start and end)
	EaseInOutCubic
)

// String returns the string representation of an easing function.
func (e EasingFunction) String() string {
	switch e {
	case EaseLinear:
		return "linear"
	case EaseInQuad:
		return "ease-in-quad"
	case EaseOutQuad:
		return "ease-out-quad"
	case EaseInOutQuad:
		return "ease-in-out-quad"
	case EaseInCubic:
		return "ease-in-cubic"
	case EaseOutCubic:
		return "ease-out-cubic"
	case EaseInOutCubic:
		return "ease-in-out-cubic"
	default:
		return "unknown"
	}
}

// HierarchyLevel represents the visual importance level of UI elements.
type HierarchyLevel int

const (
	// HierarchyPrimary is the most important level (titles, main actions)
	HierarchyPrimary HierarchyLevel = iota
	// HierarchySecondary is important content (section headers, key info)
	HierarchySecondary
	// HierarchyTertiary is supporting content (details, descriptions)
	HierarchyTertiary
	// HierarchyQuaternary is minimal emphasis (footnotes, hints)
	HierarchyQuaternary
)

// String returns the string representation of a hierarchy level.
func (h HierarchyLevel) String() string {
	switch h {
	case HierarchyPrimary:
		return "primary"
	case HierarchySecondary:
		return "secondary"
	case HierarchyTertiary:
		return "tertiary"
	case HierarchyQuaternary:
		return "quaternary"
	default:
		return "unknown"
	}
}

// TransitionConfig contains parameters for UI transition animations.
type TransitionConfig struct {
	// Type of transition
	Type TransitionType
	// Duration in milliseconds
	Duration float64
	// Easing function to use
	Easing EasingFunction
	// Progress from 0.0 (start) to 1.0 (end)
	Progress float64
}

// DefaultTransitionConfig returns a default transition configuration.
func DefaultTransitionConfig() TransitionConfig {
	return TransitionConfig{
		Type:     TransitionFade,
		Duration: 300.0, // 300ms
		Easing:   EaseInOutQuad,
		Progress: 0.0,
	}
}

// Validate checks if the transition configuration is valid.
func (t TransitionConfig) Validate() error {
	if t.Duration < 0 {
		return fmt.Errorf("duration must be non-negative, got %f", t.Duration)
	}
	if t.Progress < 0.0 || t.Progress > 1.0 {
		return fmt.Errorf("progress must be between 0.0 and 1.0, got %f", t.Progress)
	}
	return nil
}
