// Package engine provides the event decoration component for seasonal events.
// EventDecorationComponent enables entities to display seasonal decorations
// that are applied and removed based on active events in the world.
package engine

import (
	"encoding/json"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// DecorationTheme defines the visual theme for event decorations.
type DecorationTheme string

const (
	// DecorationThemeNone indicates no active decorations.
	DecorationThemeNone DecorationTheme = ""
	// DecorationThemeSpring uses floral decorations with bright colors.
	DecorationThemeSpring DecorationTheme = "spring"
	// DecorationThemeSummer uses sun and festival decorations.
	DecorationThemeSummer DecorationTheme = "summer"
	// DecorationThemeAutumn uses harvest and leaf decorations.
	DecorationThemeAutumn DecorationTheme = "autumn"
	// DecorationThemeWinter uses snow and light decorations.
	DecorationThemeWinter DecorationTheme = "winter"
)

// DecorationElement represents a single decorative item.
type DecorationElement struct {
	// Type is the decoration category (banner, garland, lantern, etc.)
	Type string `json:"type"`
	// OffsetX is the horizontal offset from entity center
	OffsetX float64 `json:"offset_x"`
	// OffsetY is the vertical offset from entity center
	OffsetY float64 `json:"offset_y"`
	// Scale is the size multiplier (1.0 = normal)
	Scale float64 `json:"scale"`
	// ColorHue is the HSV hue for coloring (0-360)
	ColorHue int `json:"color_hue"`
}

// ParticleEffectConfig configures a celebration particle effect.
type ParticleEffectConfig struct {
	// EffectType is the particle effect name
	EffectType string `json:"effect_type"`
	// Rate is particles per second (0 = disabled)
	Rate float64 `json:"rate"`
	// OffsetX is the horizontal spawn offset
	OffsetX float64 `json:"offset_x"`
	// OffsetY is the vertical spawn offset
	OffsetY float64 `json:"offset_y"`
}

// EventDecorationComponent stores decoration state for an entity.
// Decorations are automatically applied and removed by EventDecorationSystem.
type EventDecorationComponent struct {
	// ActiveTheme is the current decoration theme
	ActiveTheme DecorationTheme `json:"active_theme"`
	// EventID is the ID of the event causing decorations
	EventID string `json:"event_id"`
	// DecorationLevel controls decoration density (0.0-1.0)
	DecorationLevel float64 `json:"decoration_level"`
	// Elements contains individual decoration items
	Elements []DecorationElement `json:"elements"`
	// ParticleEffects contains celebration particle configurations
	ParticleEffects []ParticleEffectConfig `json:"particle_effects"`
	// CostumeVariant is the NPC costume override (0 = none)
	CostumeVariant int `json:"costume_variant"`
	// TransitionProgress tracks apply/remove animation (0.0-1.0)
	TransitionProgress float64 `json:"transition_progress"`
	// IsTransitioning indicates if decorations are being applied/removed
	IsTransitioning bool `json:"is_transitioning"`
	// TransitionDirection is 1 for applying, -1 for removing
	TransitionDirection int `json:"transition_direction"`
	// Seed for deterministic decoration generation
	Seed int64 `json:"seed"`
}

// NewEventDecorationComponent creates a new event decoration component.
func NewEventDecorationComponent(seed int64) *EventDecorationComponent {
	logrus.WithFields(logrus.Fields{
		"component_type": "event_decoration",
		"seed":           seed,
	}).Debug("Creating event decoration component")

	return &EventDecorationComponent{
		ActiveTheme:         DecorationThemeNone,
		EventID:             "",
		DecorationLevel:     0.0,
		Elements:            make([]DecorationElement, 0),
		ParticleEffects:     make([]ParticleEffectConfig, 0),
		CostumeVariant:      0,
		TransitionProgress:  0.0,
		IsTransitioning:     false,
		TransitionDirection: 0,
		Seed:                seed,
	}
}

// Type returns the component type identifier.
func (e *EventDecorationComponent) Type() string {
	return "event_decoration"
}

// HasDecorations returns true if the entity has active decorations.
func (e *EventDecorationComponent) HasDecorations() bool {
	return e.ActiveTheme != DecorationThemeNone && e.DecorationLevel > 0
}

// GetVisibleElements returns decoration elements based on current progress.
// During transitions, only returns elements up to the current progress.
func (e *EventDecorationComponent) GetVisibleElements() []DecorationElement {
	if len(e.Elements) == 0 || e.DecorationLevel <= 0 {
		return nil
	}

	visibleCount := len(e.Elements)
	if e.IsTransitioning {
		visibleCount = int(float64(len(e.Elements)) * e.TransitionProgress)
	}

	if visibleCount <= 0 {
		return nil
	}
	if visibleCount > len(e.Elements) {
		visibleCount = len(e.Elements)
	}

	return e.Elements[:visibleCount]
}

// GetActiveParticleEffects returns particle effects that should be active.
func (e *EventDecorationComponent) GetActiveParticleEffects() []ParticleEffectConfig {
	if !e.HasDecorations() || e.TransitionProgress < 0.5 {
		return nil
	}
	return e.ParticleEffects
}

// IsFullyDecorated returns true if decorations are fully applied.
func (e *EventDecorationComponent) IsFullyDecorated() bool {
	return e.HasDecorations() && !e.IsTransitioning && e.TransitionProgress >= 1.0
}

// ClearDecorations removes all decorations from the entity.
func (e *EventDecorationComponent) ClearDecorations() {
	logrus.WithFields(logrus.Fields{
		"component_type": "event_decoration",
		"event_id":       e.EventID,
		"theme":          e.ActiveTheme,
	}).Debug("Clearing event decorations")

	e.ActiveTheme = DecorationThemeNone
	e.EventID = ""
	e.DecorationLevel = 0.0
	e.Elements = make([]DecorationElement, 0)
	e.ParticleEffects = make([]ParticleEffectConfig, 0)
	e.CostumeVariant = 0
	e.TransitionProgress = 0.0
	e.IsTransitioning = false
	e.TransitionDirection = 0
}

// Serialize encodes the component to bytes for persistence.
func (e *EventDecorationComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type": "event_decoration",
		"theme":          e.ActiveTheme,
		"elements":       len(e.Elements),
	}).Debug("Serializing event decoration component")

	data, err := json.Marshal(e)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "event_decoration",
			"error":          err.Error(),
		}).Error("Failed to serialize event decoration component")
		return nil, err
	}
	return data, nil
}

// Deserialize decodes the component from bytes.
func (e *EventDecorationComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "event_decoration",
		"bytes":          len(data),
	}).Debug("Deserializing event decoration component")

	if err := json.Unmarshal(data, e); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "event_decoration",
			"error":          err.Error(),
		}).Error("Failed to deserialize event decoration component")
		return err
	}
	return nil
}

// ThemeFromEventTheme converts an EventTheme to a DecorationTheme.
func ThemeFromEventTheme(eventTheme EventTheme) DecorationTheme {
	switch eventTheme {
	case EventThemeSpring:
		return DecorationThemeSpring
	case EventThemeSummer:
		return DecorationThemeSummer
	case EventThemeAutumn:
		return DecorationThemeAutumn
	case EventThemeWinter:
		return DecorationThemeWinter
	default:
		return DecorationThemeNone
	}
}

// GenerateDecorations creates decorations for an entity based on theme.
// Uses deterministic generation based on the component's seed.
func (e *EventDecorationComponent) GenerateDecorations(theme DecorationTheme, level float64) {
	rng := rand.New(rand.NewSource(e.Seed))

	e.ActiveTheme = theme
	e.DecorationLevel = clampFloat(level, 0.0, 1.0)

	// Generate decoration elements based on theme
	e.Elements = generateDecorationElements(rng, theme, level)

	// Generate particle effects based on theme
	e.ParticleEffects = generateParticleEffects(rng, theme, level)

	// Generate costume variant for NPCs (1-4 variants per theme)
	e.CostumeVariant = rng.Intn(4) + 1

	logrus.WithFields(logrus.Fields{
		"component_type": "event_decoration",
		"theme":          theme,
		"level":          level,
		"elements":       len(e.Elements),
		"effects":        len(e.ParticleEffects),
		"costume":        e.CostumeVariant,
	}).Debug("Generated event decorations")
}

// generateDecorationElements creates decoration items for a theme.
func generateDecorationElements(rng *rand.Rand, theme DecorationTheme, level float64) []DecorationElement {
	// Base element count scales with decoration level
	baseCount := int(level * 5)
	if baseCount < 1 {
		baseCount = 1
	}

	// Theme-specific decoration types
	types := getDecorationTypesForTheme(theme)
	if len(types) == 0 {
		return nil
	}

	// Theme-specific color hues
	hues := getColorHuesForTheme(theme)

	elements := make([]DecorationElement, 0, baseCount)
	for i := 0; i < baseCount; i++ {
		elem := DecorationElement{
			Type:     types[rng.Intn(len(types))],
			OffsetX:  (rng.Float64() - 0.5) * 32.0, // -16 to +16 pixels
			OffsetY:  (rng.Float64() - 0.5) * 32.0,
			Scale:    0.8 + rng.Float64()*0.4, // 0.8 to 1.2 scale
			ColorHue: hues[rng.Intn(len(hues))],
		}
		elements = append(elements, elem)
	}

	return elements
}

// getDecorationTypesForTheme returns decoration types for a theme.
func getDecorationTypesForTheme(theme DecorationTheme) []string {
	switch theme {
	case DecorationThemeSpring:
		return []string{"flower", "garland", "ribbon", "butterfly", "blossom"}
	case DecorationThemeSummer:
		return []string{"sunburst", "streamer", "banner", "torch", "wreath"}
	case DecorationThemeAutumn:
		return []string{"pumpkin", "leaf", "cornucopia", "hay", "lantern"}
	case DecorationThemeWinter:
		return []string{"snowflake", "icicle", "lights", "holly", "star"}
	default:
		return nil
	}
}

// getColorHuesForTheme returns HSV hues for a theme.
func getColorHuesForTheme(theme DecorationTheme) []int {
	switch theme {
	case DecorationThemeSpring:
		return []int{90, 120, 300, 330, 60} // Green, pink, purple, yellow
	case DecorationThemeSummer:
		return []int{40, 20, 200, 60, 0} // Orange, gold, blue, yellow, red
	case DecorationThemeAutumn:
		return []int{30, 15, 45, 0, 50} // Orange, brown, red, gold
	case DecorationThemeWinter:
		return []int{200, 220, 240, 0, 120} // Blue, cyan, white (blue tint), red, green
	default:
		return []int{0}
	}
}

// generateParticleEffects creates particle configurations for a theme.
func generateParticleEffects(rng *rand.Rand, theme DecorationTheme, level float64) []ParticleEffectConfig {
	// Only generate particles at higher decoration levels
	if level < 0.3 {
		return nil
	}

	effects := make([]ParticleEffectConfig, 0)

	switch theme {
	case DecorationThemeSpring:
		// Floating flower petals
		effects = append(effects, ParticleEffectConfig{
			EffectType: "petals",
			Rate:       level * 2.0,
			OffsetX:    0,
			OffsetY:    -16,
		})
	case DecorationThemeSummer:
		// Fireflies at night, confetti during day
		effects = append(effects, ParticleEffectConfig{
			EffectType: "confetti",
			Rate:       level * 3.0,
			OffsetX:    0,
			OffsetY:    -24,
		})
	case DecorationThemeAutumn:
		// Falling leaves
		effects = append(effects, ParticleEffectConfig{
			EffectType: "leaves",
			Rate:       level * 2.5,
			OffsetX:    0,
			OffsetY:    -20,
		})
	case DecorationThemeWinter:
		// Snowflakes and sparkles
		effects = append(effects, ParticleEffectConfig{
			EffectType: "snow",
			Rate:       level * 3.0,
			OffsetX:    0,
			OffsetY:    -32,
		})
		if level >= 0.7 {
			effects = append(effects, ParticleEffectConfig{
				EffectType: "sparkle",
				Rate:       level * 1.5,
				OffsetX:    0,
				OffsetY:    -8,
			})
		}
	}

	return effects
}

// GetCostumeOffset returns the sprite sheet offset for the costume variant.
// Returns (0, 0) if no costume variant is active.
func (e *EventDecorationComponent) GetCostumeOffset() (int, int) {
	if e.CostumeVariant == 0 || !e.HasDecorations() {
		return 0, 0
	}

	// Each costume variant is a row offset in the sprite sheet
	// Convention: row 0-3 = normal, row 4-7 = spring, 8-11 = summer, etc.
	themeOffset := 0
	switch e.ActiveTheme {
	case DecorationThemeSpring:
		themeOffset = 4
	case DecorationThemeSummer:
		themeOffset = 8
	case DecorationThemeAutumn:
		themeOffset = 12
	case DecorationThemeWinter:
		themeOffset = 16
	}

	return 0, themeOffset + (e.CostumeVariant - 1)
}
