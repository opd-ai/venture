package modding

import (
	"fmt"
	"time"
)

// ModType represents the type of modification a mod provides.
type ModType string

const (
	// ModTypeRule modifies gameplay rules and parameters
	ModTypeRule ModType = "rule"
	// ModTypeGenerator customizes procedural generation parameters
	ModTypeGenerator ModType = "generator"
	// ModTypeEvent adds custom server events and triggers
	ModTypeEvent ModType = "event"
)

// String returns the string representation of the ModType.
func (m ModType) String() string {
	return string(m)
}

// Mod represents a single server modification.
type Mod struct {
	// ID is a unique identifier for the mod
	ID string `json:"id"`

	// Name is the human-readable name of the mod
	Name string `json:"name"`

	// Version follows semantic versioning (e.g., "1.0.0")
	Version string `json:"version"`

	// Author is the creator of the mod
	Author string `json:"author"`

	// Description explains what the mod does
	Description string `json:"description"`

	// Type specifies the mod type
	Type ModType `json:"type"`

	// Dependencies lists other mods required by this mod
	Dependencies []string `json:"dependencies,omitempty"`

	// Rules contains gameplay parameter modifications
	Rules map[string]interface{} `json:"rules,omitempty"`

	// GeneratorParams contains procedural generation parameter modifications
	GeneratorParams map[string]interface{} `json:"generator_params,omitempty"`

	// EventHandlers defines custom event triggers
	EventHandlers map[string]EventHandler `json:"-"`

	// Enabled indicates whether the mod is currently active
	Enabled bool `json:"enabled"`

	// LoadedAt is the timestamp when the mod was loaded
	LoadedAt time.Time `json:"loaded_at"`
}

// Validate checks if the mod configuration is valid.
func (m *Mod) Validate() error {
	if err := m.validateBasicFields(); err != nil {
		return err
	}

	if err := m.validateModType(); err != nil {
		return err
	}

	if err := m.validateRules(); err != nil {
		return err
	}

	return nil
}

// validateBasicFields validates required mod fields
func (m *Mod) validateBasicFields() error {
	if m.ID == "" {
		return fmt.Errorf("mod ID cannot be empty")
	}
	if m.Name == "" {
		return fmt.Errorf("mod name cannot be empty")
	}
	if m.Version == "" {
		return fmt.Errorf("mod version cannot be empty")
	}
	if m.Type == "" {
		m.Type = ModTypeRule
	}
	return nil
}

// validateModType validates the mod type is valid
func (m *Mod) validateModType() error {
	validTypes := []ModType{ModTypeRule, ModTypeGenerator, ModTypeEvent}
	for _, t := range validTypes {
		if m.Type == t {
			return nil
		}
	}
	return fmt.Errorf("invalid mod type: %s", m.Type)
}

// validateRules validates mod rules if present
func (m *Mod) validateRules() error {
	if m.Type != ModTypeRule || len(m.Rules) == 0 {
		return nil
	}

	for key, value := range m.Rules {
		if key == "" {
			return fmt.Errorf("rule key cannot be empty")
		}
		if value == nil {
			return fmt.Errorf("rule value for %s cannot be nil", key)
		}
	}
	return nil
}

// EventHandler represents a custom event handler function.
type EventHandler func(event Event) error

// Event represents a game event that can trigger mod handlers.
type Event struct {
	// Type is the event type (e.g., "player_join", "entity_spawn")
	Type string

	// Data contains event-specific data
	Data map[string]interface{}

	// Timestamp is when the event occurred
	Timestamp time.Time
}

// ModConfig represents the configuration for the mod system.
type ModConfig struct {
	// ModsDirectory is the path to the mods directory
	ModsDirectory string

	// EnableSandbox enables execution sandboxing
	EnableSandbox bool

	// MaxMods is the maximum number of loaded mods
	MaxMods int

	// RuleChangeRateLimit limits how often rules can be changed (per second)
	RuleChangeRateLimit float64
}

// DefaultConfig returns the default mod configuration.
func DefaultConfig() ModConfig {
	return ModConfig{
		ModsDirectory:       "mods",
		EnableSandbox:       true,
		MaxMods:             50,
		RuleChangeRateLimit: 10.0, // 10 changes per second max
	}
}

// RuleContext contains the context for applying a rule modification.
type RuleContext struct {
	// ModID is the ID of the mod applying the rule
	ModID string

	// RuleName is the name of the rule being applied
	RuleName string

	// OldValue is the previous value (nil if not set)
	OldValue interface{}

	// NewValue is the new value to apply
	NewValue interface{}

	// AppliedAt is when the rule was applied
	AppliedAt time.Time
}

// LoadError represents an error that occurred during mod loading.
type LoadError struct {
	ModID string
	Err   error
}

func (e *LoadError) Error() string {
	return fmt.Sprintf("failed to load mod %s: %v", e.ModID, e.Err)
}

// ValidationError represents a mod validation error.
type ValidationError struct {
	ModID  string
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for mod %s (field %s): %s", e.ModID, e.Field, e.Reason)
}
