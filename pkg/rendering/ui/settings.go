// Package ui provides procedural UI generation systems for Venture.
// This file implements the unified settings menu system for Phase 60.1.
package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// SettingsCategory represents major settings groupings
type SettingsCategory int

const (
	CategoryGraphics SettingsCategory = iota
	CategoryAudio
	CategoryControls
	CategoryGameplay
	CategoryNetwork
	CategoryAccessibility
	CategoryInterface
	CategoryPerformance
	CategorySocial
	CategoryAdvanced
)

func (c SettingsCategory) String() string {
	switch c {
	case CategoryGraphics:
		return "Graphics"
	case CategoryAudio:
		return "Audio"
	case CategoryControls:
		return "Controls"
	case CategoryGameplay:
		return "Gameplay"
	case CategoryNetwork:
		return "Network"
	case CategoryAccessibility:
		return "Accessibility"
	case CategoryInterface:
		return "Interface"
	case CategoryPerformance:
		return "Performance"
	case CategorySocial:
		return "Social"
	case CategoryAdvanced:
		return "Advanced"
	default:
		return "Unknown"
	}
}

// SettingType defines the data type of a setting
type SettingType int

const (
	TypeBool SettingType = iota
	TypeInt
	TypeFloat
	TypeString
	TypeEnum
)

func (t SettingType) String() string {
	switch t {
	case TypeBool:
		return "Boolean"
	case TypeInt:
		return "Integer"
	case TypeFloat:
		return "Float"
	case TypeString:
		return "String"
	case TypeEnum:
		return "Enum"
	default:
		return "Unknown"
	}
}

// Setting represents a single configurable option
type Setting struct {
	ID              string
	Name            string
	Description     string
	Category        SettingsCategory
	Type            SettingType
	DefaultValue    interface{}
	CurrentValue    interface{}
	MinValue        interface{} // for int/float
	MaxValue        interface{} // for int/float
	EnumOptions     []string    // for enum type
	RequiresRestart bool
}

// SettingsManager manages all game settings
type SettingsManager struct {
	mu       sync.RWMutex
	settings map[string]*Setting
	modified bool
}

// NewSettingsManager creates a new settings manager with defaults
func NewSettingsManager() *SettingsManager {
	sm := &SettingsManager{
		settings: make(map[string]*Setting),
	}
	sm.registerDefaultSettings()
	return sm
}

// registerDefaultSettings populates default settings for all categories
func (sm *SettingsManager) registerDefaultSettings() {
	// Graphics settings
	sm.registerSetting(&Setting{
		ID:           "graphics.resolution",
		Name:         "Resolution",
		Description:  "Screen resolution",
		Category:     CategoryGraphics,
		Type:         TypeEnum,
		DefaultValue: "1920x1080",
		CurrentValue: "1920x1080",
		EnumOptions:  []string{"1280x720", "1920x1080", "2560x1440", "3840x2160"},
	})
	sm.registerSetting(&Setting{
		ID:              "graphics.fullscreen",
		Name:            "Fullscreen",
		Description:     "Enable fullscreen mode",
		Category:        CategoryGraphics,
		Type:            TypeBool,
		DefaultValue:    false,
		CurrentValue:    false,
		RequiresRestart: true,
	})
	sm.registerSetting(&Setting{
		ID:           "graphics.vsync",
		Name:         "VSync",
		Description:  "Enable vertical synchronization",
		Category:     CategoryGraphics,
		Type:         TypeBool,
		DefaultValue: true,
		CurrentValue: true,
	})
	sm.registerSetting(&Setting{
		ID:           "graphics.antialiasing",
		Name:         "Anti-Aliasing",
		Description:  "Enable anti-aliasing for smooth edges",
		Category:     CategoryGraphics,
		Type:         TypeBool,
		DefaultValue: true,
		CurrentValue: true,
	})
	sm.registerSetting(&Setting{
		ID:           "graphics.particles",
		Name:         "Particle Quality",
		Description:  "Particle effect quality level",
		Category:     CategoryGraphics,
		Type:         TypeEnum,
		DefaultValue: "High",
		CurrentValue: "High",
		EnumOptions:  []string{"Low", "Medium", "High", "Ultra"},
	})

	// Audio settings
	sm.registerSetting(&Setting{
		ID:           "audio.master_volume",
		Name:         "Master Volume",
		Description:  "Overall audio volume (0-100)",
		Category:     CategoryAudio,
		Type:         TypeInt,
		DefaultValue: 100,
		CurrentValue: 100,
		MinValue:     0,
		MaxValue:     100,
	})
	sm.registerSetting(&Setting{
		ID:           "audio.music_volume",
		Name:         "Music Volume",
		Description:  "Background music volume (0-100)",
		Category:     CategoryAudio,
		Type:         TypeInt,
		DefaultValue: 80,
		CurrentValue: 80,
		MinValue:     0,
		MaxValue:     100,
	})
	sm.registerSetting(&Setting{
		ID:           "audio.sfx_volume",
		Name:         "SFX Volume",
		Description:  "Sound effects volume (0-100)",
		Category:     CategoryAudio,
		Type:         TypeInt,
		DefaultValue: 90,
		CurrentValue: 90,
		MinValue:     0,
		MaxValue:     100,
	})
	sm.registerSetting(&Setting{
		ID:           "audio.ambient_volume",
		Name:         "Ambient Volume",
		Description:  "Environmental sounds volume (0-100)",
		Category:     CategoryAudio,
		Type:         TypeInt,
		DefaultValue: 70,
		CurrentValue: 70,
		MinValue:     0,
		MaxValue:     100,
	})

	// Gameplay settings
	sm.registerSetting(&Setting{
		ID:           "gameplay.difficulty",
		Name:         "Difficulty",
		Description:  "Global difficulty multiplier",
		Category:     CategoryGameplay,
		Type:         TypeFloat,
		DefaultValue: 0.5,
		CurrentValue: 0.5,
		MinValue:     0.0,
		MaxValue:     1.0,
	})
	sm.registerSetting(&Setting{
		ID:           "gameplay.autosave",
		Name:         "Auto-Save",
		Description:  "Enable automatic saving",
		Category:     CategoryGameplay,
		Type:         TypeBool,
		DefaultValue: true,
		CurrentValue: true,
	})
	sm.registerSetting(&Setting{
		ID:           "gameplay.autosave_interval",
		Name:         "Auto-Save Interval",
		Description:  "Minutes between auto-saves",
		Category:     CategoryGameplay,
		Type:         TypeInt,
		DefaultValue: 5,
		CurrentValue: 5,
		MinValue:     1,
		MaxValue:     30,
	})
	sm.registerSetting(&Setting{
		ID:           "gameplay.show_tutorial",
		Name:         "Show Tutorial",
		Description:  "Display tutorial popups",
		Category:     CategoryGameplay,
		Type:         TypeBool,
		DefaultValue: true,
		CurrentValue: true,
	})

	// Network settings
	sm.registerSetting(&Setting{
		ID:           "network.server_address",
		Name:         "Server Address",
		Description:  "Multiplayer server address",
		Category:     CategoryNetwork,
		Type:         TypeString,
		DefaultValue: "localhost:8080",
		CurrentValue: "localhost:8080",
	})
	sm.registerSetting(&Setting{
		ID:           "network.auto_connect",
		Name:         "Auto-Connect",
		Description:  "Auto-connect to last server",
		Category:     CategoryNetwork,
		Type:         TypeBool,
		DefaultValue: false,
		CurrentValue: false,
	})
	sm.registerSetting(&Setting{
		ID:           "network.lag_compensation",
		Name:         "Lag Compensation",
		Description:  "Enable client-side prediction",
		Category:     CategoryNetwork,
		Type:         TypeBool,
		DefaultValue: true,
		CurrentValue: true,
	})

	// Accessibility settings
	sm.registerSetting(&Setting{
		ID:           "accessibility.colorblind_mode",
		Name:         "Colorblind Mode",
		Description:  "Colorblind-friendly palette",
		Category:     CategoryAccessibility,
		Type:         TypeEnum,
		DefaultValue: "None",
		CurrentValue: "None",
		EnumOptions:  []string{"None", "Protanopia", "Deuteranopia", "Tritanopia"},
	})
	sm.registerSetting(&Setting{
		ID:           "accessibility.font_scale",
		Name:         "Font Scale",
		Description:  "Text size multiplier (0.5-2.0)",
		Category:     CategoryAccessibility,
		Type:         TypeFloat,
		DefaultValue: 1.0,
		CurrentValue: 1.0,
		MinValue:     0.5,
		MaxValue:     2.0,
	})
	sm.registerSetting(&Setting{
		ID:           "accessibility.high_contrast",
		Name:         "High Contrast",
		Description:  "Increase UI contrast",
		Category:     CategoryAccessibility,
		Type:         TypeBool,
		DefaultValue: false,
		CurrentValue: false,
	})
	sm.registerSetting(&Setting{
		ID:           "accessibility.screen_reader",
		Name:         "Screen Reader Support",
		Description:  "Enable screen reader compatibility",
		Category:     CategoryAccessibility,
		Type:         TypeBool,
		DefaultValue: false,
		CurrentValue: false,
	})

	// Interface settings
	sm.registerSetting(&Setting{
		ID:           "interface.ui_scale",
		Name:         "UI Scale",
		Description:  "Overall UI size multiplier (0.5-2.0)",
		Category:     CategoryInterface,
		Type:         TypeFloat,
		DefaultValue: 1.0,
		CurrentValue: 1.0,
		MinValue:     0.5,
		MaxValue:     2.0,
	})
	sm.registerSetting(&Setting{
		ID:           "interface.show_fps",
		Name:         "Show FPS",
		Description:  "Display frame rate counter",
		Category:     CategoryInterface,
		Type:         TypeBool,
		DefaultValue: false,
		CurrentValue: false,
	})
	sm.registerSetting(&Setting{
		ID:           "interface.minimap_size",
		Name:         "Minimap Size",
		Description:  "Minimap display size",
		Category:     CategoryInterface,
		Type:         TypeEnum,
		DefaultValue: "Medium",
		CurrentValue: "Medium",
		EnumOptions:  []string{"Small", "Medium", "Large"},
	})

	// Performance settings
	sm.registerSetting(&Setting{
		ID:           "performance.sprite_cache_size",
		Name:         "Sprite Cache Size",
		Description:  "Sprite cache size in MB",
		Category:     CategoryPerformance,
		Type:         TypeInt,
		DefaultValue: 300,
		CurrentValue: 300,
		MinValue:     100,
		MaxValue:     1000,
	})
	sm.registerSetting(&Setting{
		ID:           "performance.entity_culling",
		Name:         "Entity Culling",
		Description:  "Cull off-screen entities",
		Category:     CategoryPerformance,
		Type:         TypeBool,
		DefaultValue: true,
		CurrentValue: true,
	})

	// Social settings
	sm.registerSetting(&Setting{
		ID:           "social.show_online_status",
		Name:         "Show Online Status",
		Description:  "Display online status to others",
		Category:     CategorySocial,
		Type:         TypeBool,
		DefaultValue: true,
		CurrentValue: true,
	})
	sm.registerSetting(&Setting{
		ID:           "social.auto_decline_invites",
		Name:         "Auto-Decline Invites",
		Description:  "Auto-decline all invitations",
		Category:     CategorySocial,
		Type:         TypeBool,
		DefaultValue: false,
		CurrentValue: false,
	})
}

// registerSetting adds a setting to the manager
func (sm *SettingsManager) registerSetting(s *Setting) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.settings[s.ID] = s
}

// GetSetting retrieves a setting by ID
func (sm *SettingsManager) GetSetting(id string) (*Setting, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	s, exists := sm.settings[id]
	if !exists {
		return nil, fmt.Errorf("setting not found: %s", id)
	}
	return s, nil
}

// SetValue updates a setting's value
func (sm *SettingsManager) SetValue(id string, value interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	s, exists := sm.settings[id]
	if !exists {
		return fmt.Errorf("setting not found: %s", id)
	}

	// Type validation
	if err := sm.validateValue(s, value); err != nil {
		return err
	}

	s.CurrentValue = value
	sm.modified = true
	return nil
}

// validateValue checks if a value is valid for a setting
func (sm *SettingsManager) validateValue(s *Setting, value interface{}) error {
	switch s.Type {
	case TypeBool:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected bool, got %T", value)
		}
	case TypeInt:
		intVal, ok := value.(int)
		if !ok {
			return fmt.Errorf("expected int, got %T", value)
		}
		if s.MinValue != nil {
			if intVal < s.MinValue.(int) {
				return fmt.Errorf("value %d below minimum %d", intVal, s.MinValue.(int))
			}
		}
		if s.MaxValue != nil {
			if intVal > s.MaxValue.(int) {
				return fmt.Errorf("value %d above maximum %d", intVal, s.MaxValue.(int))
			}
		}
	case TypeFloat:
		floatVal, ok := value.(float64)
		if !ok {
			return fmt.Errorf("expected float64, got %T", value)
		}
		if s.MinValue != nil {
			if floatVal < s.MinValue.(float64) {
				return fmt.Errorf("value %.2f below minimum %.2f", floatVal, s.MinValue.(float64))
			}
		}
		if s.MaxValue != nil {
			if floatVal > s.MaxValue.(float64) {
				return fmt.Errorf("value %.2f above maximum %.2f", floatVal, s.MaxValue.(float64))
			}
		}
	case TypeString:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case TypeEnum:
		strVal, ok := value.(string)
		if !ok {
			return fmt.Errorf("expected string for enum, got %T", value)
		}
		valid := false
		for _, opt := range s.EnumOptions {
			if strVal == opt {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("invalid enum value: %s", strVal)
		}
	}
	return nil
}

// GetValue retrieves current value of a setting
func (sm *SettingsManager) GetValue(id string) (interface{}, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	s, exists := sm.settings[id]
	if !exists {
		return nil, fmt.Errorf("setting not found: %s", id)
	}
	return s.CurrentValue, nil
}

// GetBool is a typed getter for bool settings
func (sm *SettingsManager) GetBool(id string) bool {
	val, err := sm.GetValue(id)
	if err != nil {
		return false
	}
	boolVal, ok := val.(bool)
	if !ok {
		return false
	}
	return boolVal
}

// GetInt is a typed getter for int settings
func (sm *SettingsManager) GetInt(id string) int {
	val, err := sm.GetValue(id)
	if err != nil {
		return 0
	}
	intVal, ok := val.(int)
	if !ok {
		return 0
	}
	return intVal
}

// GetFloat is a typed getter for float settings
func (sm *SettingsManager) GetFloat(id string) float64 {
	val, err := sm.GetValue(id)
	if err != nil {
		return 0.0
	}
	floatVal, ok := val.(float64)
	if !ok {
		return 0.0
	}
	return floatVal
}

// GetString is a typed getter for string settings
func (sm *SettingsManager) GetString(id string) string {
	val, err := sm.GetValue(id)
	if err != nil {
		return ""
	}
	strVal, ok := val.(string)
	if !ok {
		return ""
	}
	return strVal
}

// ListByCategory returns all settings in a category
func (sm *SettingsManager) ListByCategory(category SettingsCategory) []*Setting {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]*Setting, 0)
	for _, s := range sm.settings {
		if s.Category == category {
			result = append(result, s)
		}
	}
	return result
}

// ResetToDefaults resets all settings to default values
func (sm *SettingsManager) ResetToDefaults() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for _, s := range sm.settings {
		s.CurrentValue = s.DefaultValue
	}
	sm.modified = true
}

// IsModified returns whether settings have been changed
func (sm *SettingsManager) IsModified() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.modified
}

// Save persists settings to a file
func (sm *SettingsManager) Save(filename string) error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// Create settings data map
	data := make(map[string]interface{})
	for id, s := range sm.settings {
		data[id] = s.CurrentValue
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal settings: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filename, jsonData, 0o644); err != nil {
		return fmt.Errorf("failed to write settings file: %w", err)
	}

	sm.modified = false
	return nil
}

// Load reads settings from a file
func (sm *SettingsManager) Load(filename string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Read file
	jsonData, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read settings file: %w", err)
	}

	// Unmarshal JSON
	var data map[string]interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal settings: %w", err)
	}

	// Apply values
	for id, value := range data {
		if s, exists := sm.settings[id]; exists {
			// JSON unmarshals numbers as float64, convert if needed
			if s.Type == TypeInt {
				if floatVal, ok := value.(float64); ok {
					value = int(floatVal)
				}
			}
			// Validate and set
			if err := sm.validateValue(s, value); err == nil {
				s.CurrentValue = value
			}
		}
	}

	sm.modified = false
	return nil
}
