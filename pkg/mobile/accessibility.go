package mobile

// Accessibility provides screen reader and assistive technology support for mobile platforms.
// This file implements accessibility hints for iOS VoiceOver and Android TalkBack.

// AccessibilityHint defines an accessibility hint for UI elements.
// Used by screen readers to provide context to visually impaired users.
type AccessibilityHint struct {
	Label       string // Short label for the element (e.g., "Health Bar")
	Hint        string // Detailed hint (e.g., "Shows current player health percentage")
	Traits      []AccessibilityTrait
	Value       string // Current value (e.g., "75 percent")
	IsEnabled   bool   // Whether the element is interactive
	IsContainer bool   // Whether element contains other elements
}

// AccessibilityTrait defines the type of UI element for screen readers.
type AccessibilityTrait string

const (
	// TraitButton indicates an interactive button element
	TraitButton AccessibilityTrait = "button"
	// TraitImage indicates a static image element
	TraitImage AccessibilityTrait = "image"
	// TraitStaticText indicates non-interactive text
	TraitStaticText AccessibilityTrait = "staticText"
	// TraitHeader indicates a heading element
	TraitHeader AccessibilityTrait = "header"
	// TraitLink indicates a clickable link
	TraitLink AccessibilityTrait = "link"
	// TraitAdjustable indicates a value that can be incremented/decremented (slider, stepper)
	TraitAdjustable AccessibilityTrait = "adjustable"
	// TraitSelected indicates a currently selected element
	TraitSelected AccessibilityTrait = "selected"
	// TraitPlaysSound indicates an element that produces audio
	TraitPlaysSound AccessibilityTrait = "playsSound"
	// TraitUpdatesFrequently indicates dynamic content (health bar, timer)
	TraitUpdatesFrequently AccessibilityTrait = "updatesFrequently"
)

// NewAccessibilityHint creates a new accessibility hint for a UI element.
func NewAccessibilityHint(label, hint string, traits []AccessibilityTrait) *AccessibilityHint {
	return &AccessibilityHint{
		Label:     label,
		Hint:      hint,
		Traits:    traits,
		IsEnabled: true,
	}
}

// SetValue updates the current value for dynamic elements (health bars, progress indicators).
func (h *AccessibilityHint) SetValue(value string) {
	h.Value = value
}

// SetEnabled sets whether the element is interactive.
func (h *AccessibilityHint) SetEnabled(enabled bool) {
	h.IsEnabled = enabled
}

// GetARIAAttributes returns ARIA attributes for WASM builds.
// Use these in HTML elements to provide screen reader support in browsers.
func (h *AccessibilityHint) GetARIAAttributes() map[string]string {
	attrs := make(map[string]string)

	// Basic attributes
	attrs["aria-label"] = h.Label
	if h.Hint != "" {
		attrs["aria-description"] = h.Hint
	}
	if h.Value != "" {
		attrs["aria-valuenow"] = h.Value
	}

	// Role based on traits
	for _, trait := range h.Traits {
		switch trait {
		case TraitButton:
			attrs["role"] = "button"
		case TraitImage:
			attrs["role"] = "img"
		case TraitHeader:
			attrs["role"] = "heading"
		case TraitLink:
			attrs["role"] = "link"
		case TraitAdjustable:
			attrs["role"] = "slider"
		case TraitStaticText:
			attrs["role"] = "text"
		case TraitUpdatesFrequently:
			attrs["aria-live"] = "polite"
		case TraitPlaysSound:
			// No specific ARIA role, but can use aria-label to indicate
		}
	}

	// State
	if !h.IsEnabled {
		attrs["aria-disabled"] = "true"
	}
	if h.containsTrait(TraitSelected) {
		attrs["aria-selected"] = "true"
	}

	return attrs
}

// containsTrait checks if the hint contains a specific trait.
func (h *AccessibilityHint) containsTrait(trait AccessibilityTrait) bool {
	for _, t := range h.Traits {
		if t == trait {
			return true
		}
	}
	return false
}

// StandardHints provides pre-configured accessibility hints for common game UI elements.
var StandardHints = struct {
	HealthBar        *AccessibilityHint
	ManaBar          *AccessibilityHint
	ExperienceBar    *AccessibilityHint
	InventoryButton  *AccessibilityHint
	MapButton        *AccessibilityHint
	MenuButton       *AccessibilityHint
	AttackButton     *AccessibilityHint
	InteractButton   *AccessibilityHint
	Minimap          *AccessibilityHint
	NotificationArea *AccessibilityHint
}{
	HealthBar: NewAccessibilityHint(
		"Health",
		"Current player health percentage",
		[]AccessibilityTrait{TraitAdjustable, TraitUpdatesFrequently},
	),
	ManaBar: NewAccessibilityHint(
		"Mana",
		"Current player mana or magic energy percentage",
		[]AccessibilityTrait{TraitAdjustable, TraitUpdatesFrequently},
	),
	ExperienceBar: NewAccessibilityHint(
		"Experience",
		"Progress toward next character level",
		[]AccessibilityTrait{TraitAdjustable, TraitUpdatesFrequently},
	),
	InventoryButton: NewAccessibilityHint(
		"Inventory",
		"Open your inventory to view and manage items",
		[]AccessibilityTrait{TraitButton},
	),
	MapButton: NewAccessibilityHint(
		"Map",
		"Open the world map to view explored areas",
		[]AccessibilityTrait{TraitButton},
	),
	MenuButton: NewAccessibilityHint(
		"Menu",
		"Open game menu for settings, save, and quit options",
		[]AccessibilityTrait{TraitButton},
	),
	AttackButton: NewAccessibilityHint(
		"Attack",
		"Perform a melee attack with your equipped weapon",
		[]AccessibilityTrait{TraitButton},
	),
	InteractButton: NewAccessibilityHint(
		"Interact",
		"Interact with nearby objects, NPCs, or doors",
		[]AccessibilityTrait{TraitButton},
	),
	Minimap: NewAccessibilityHint(
		"Minimap",
		"Shows nearby terrain and your current position",
		[]AccessibilityTrait{TraitImage, TraitUpdatesFrequently},
	),
	NotificationArea: NewAccessibilityHint(
		"Notifications",
		"Game notifications and status messages",
		[]AccessibilityTrait{TraitUpdatesFrequently},
	),
}
