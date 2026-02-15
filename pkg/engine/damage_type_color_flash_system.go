// Package engine provides the DamageTypeColorFlashSystem which applies genre-aware
// colored sprite flashes based on the elemental damage type received. It hooks into
// the CombatSystem via AddDamageCallback and reads the attacker's AttackComponent
// DamageType to determine flash color, then applies the tint through the entity's
// VisualFeedbackComponent for per-frame decay.
package engine

import (
	"image/color"
	"math"

	"github.com/opd-ai/venture/pkg/combat"
	"github.com/sirupsen/logrus"
)

// DamageTypeFlashComponent stores the active colored flash state for an entity.
type DamageTypeFlashComponent struct {
	LastDamageType combat.DamageType
	FlashTimer     float64 // Seconds remaining
	FlashDuration  float64 // Total duration for lerp
	FlashIntensity float64 // 0.0–1.0, scales with damage proportion
	FlashColor     color.RGBA
}

// Type returns the component type identifier.
func (d *DamageTypeFlashComponent) Type() string {
	return "damage_type_flash"
}

// NewDamageTypeFlashComponent creates a new damage type flash component with defaults.
func NewDamageTypeFlashComponent() *DamageTypeFlashComponent {
	return &DamageTypeFlashComponent{
		FlashDuration: 0.3,
	}
}

// damageTypeColorPreset holds genre-specific flash colors per damage type.
type damageTypeColorPreset struct {
	Physical  color.RGBA
	Magical   color.RGBA
	Fire      color.RGBA
	Ice       color.RGBA
	Lightning color.RGBA
	Poison    color.RGBA
}

// DamageTypeColorFlashSystem applies elemental-colored sprite flashes when entities
// take damage. Each damage type produces a distinctive color and genre-specific tint.
type DamageTypeColorFlashSystem struct {
	world   *World
	logger  *logrus.Entry
	genreID string
	preset  damageTypeColorPreset

	flashDuration float64 // Base flash duration in seconds
	maxIntensity  float64 // Cap for flash intensity
	damageScale   float64 // Damage-to-intensity divisor
}

// NewDamageTypeColorFlashSystem creates a new damage type color flash system.
func NewDamageTypeColorFlashSystem(world *World, seed int64) *DamageTypeColorFlashSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "damage_type_color_flash",
	})

	sys := &DamageTypeColorFlashSystem{
		world:         world,
		logger:        logger,
		genreID:       "fantasy",
		flashDuration: 0.3,
		maxIntensity:  1.0,
		damageScale:   50.0,
	}

	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("damage type color flash system created")
	return sys
}

// SetGenre configures genre-specific damage flash colors.
func (s *DamageTypeColorFlashSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for damage flash colors")
	}
}

// getGenrePreset returns the color palette for a given genre.
func (s *DamageTypeColorFlashSystem) getGenrePreset(genreID string) damageTypeColorPreset {
	switch genreID {
	case "horror":
		return damageTypeColorPreset{
			Physical:  color.RGBA{R: 180, G: 20, B: 20, A: 255},
			Magical:   color.RGBA{R: 100, G: 0, B: 120, A: 255},
			Fire:      color.RGBA{R: 200, G: 60, B: 0, A: 255},
			Ice:       color.RGBA{R: 80, G: 100, B: 160, A: 255},
			Lightning: color.RGBA{R: 200, G: 180, B: 30, A: 255},
			Poison:    color.RGBA{R: 40, G: 140, B: 20, A: 255},
		}
	case "scifi":
		return damageTypeColorPreset{
			Physical:  color.RGBA{R: 200, G: 220, B: 255, A: 255},
			Magical:   color.RGBA{R: 160, G: 80, B: 255, A: 255},
			Fire:      color.RGBA{R: 255, G: 120, B: 40, A: 255},
			Ice:       color.RGBA{R: 100, G: 200, B: 255, A: 255},
			Lightning: color.RGBA{R: 255, G: 255, B: 100, A: 255},
			Poison:    color.RGBA{R: 80, G: 255, B: 120, A: 255},
		}
	case "cyberpunk":
		return damageTypeColorPreset{
			Physical:  color.RGBA{R: 220, G: 220, B: 240, A: 255},
			Magical:   color.RGBA{R: 200, G: 50, B: 255, A: 255},
			Fire:      color.RGBA{R: 255, G: 80, B: 20, A: 255},
			Ice:       color.RGBA{R: 60, G: 180, B: 255, A: 255},
			Lightning: color.RGBA{R: 255, G: 240, B: 60, A: 255},
			Poison:    color.RGBA{R: 50, G: 255, B: 80, A: 255},
		}
	case "postapoc":
		return damageTypeColorPreset{
			Physical:  color.RGBA{R: 180, G: 170, B: 150, A: 255},
			Magical:   color.RGBA{R: 130, G: 80, B: 140, A: 255},
			Fire:      color.RGBA{R: 200, G: 100, B: 40, A: 255},
			Ice:       color.RGBA{R: 120, G: 150, B: 180, A: 255},
			Lightning: color.RGBA{R: 200, G: 190, B: 80, A: 255},
			Poison:    color.RGBA{R: 100, G: 140, B: 60, A: 255},
		}
	default: // fantasy
		return damageTypeColorPreset{
			Physical:  color.RGBA{R: 240, G: 240, B: 240, A: 255},
			Magical:   color.RGBA{R: 160, G: 60, B: 220, A: 255},
			Fire:      color.RGBA{R: 255, G: 100, B: 30, A: 255},
			Ice:       color.RGBA{R: 80, G: 180, B: 255, A: 255},
			Lightning: color.RGBA{R: 255, G: 230, B: 50, A: 255},
			Poison:    color.RGBA{R: 60, G: 200, B: 40, A: 255},
		}
	}
}

// colorForDamageType returns the flash color for the given damage type.
func (s *DamageTypeColorFlashSystem) colorForDamageType(dt combat.DamageType) color.RGBA {
	switch dt {
	case combat.DamageFire:
		return s.preset.Fire
	case combat.DamageIce:
		return s.preset.Ice
	case combat.DamageLightning:
		return s.preset.Lightning
	case combat.DamagePoison:
		return s.preset.Poison
	case combat.DamageMagical:
		return s.preset.Magical
	default:
		return s.preset.Physical
	}
}

// OnDamageDealt is the callback registered with CombatSystem.AddDamageCallback.
// It reads the attacker's damage type and triggers a colored flash on the target.
func (s *DamageTypeColorFlashSystem) OnDamageDealt(attacker, target *Entity, damage float64) {
	if target == nil {
		return
	}

	comp, ok := target.GetComponent("damage_type_flash")
	if !ok {
		return
	}
	flash, ok := comp.(*DamageTypeFlashComponent)
	if !ok {
		return
	}

	// Determine damage type from attacker's attack component
	var damageType combat.DamageType
	if attacker != nil {
		if atk := attacker.GetAttack(); atk != nil {
			damageType = atk.DamageType
		}
	}

	intensity := math.Min(damage/s.damageScale, s.maxIntensity)
	if intensity < 0.2 {
		intensity = 0.2
	}

	flash.LastDamageType = damageType
	flash.FlashTimer = s.flashDuration
	flash.FlashDuration = s.flashDuration
	flash.FlashIntensity = intensity
	flash.FlashColor = s.colorForDamageType(damageType)
}

// Update decays active damage flashes and applies the tint to VisualFeedbackComponent.
func (s *DamageTypeColorFlashSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		comp, ok := entity.GetComponent("damage_type_flash")
		if !ok {
			continue
		}
		flash, ok := comp.(*DamageTypeFlashComponent)
		if !ok || flash.FlashTimer <= 0 {
			continue
		}

		flash.FlashTimer -= deltaTime
		if flash.FlashTimer < 0 {
			flash.FlashTimer = 0
		}

		// Apply color tint to visual feedback based on remaining timer ratio
		feedback := entity.GetVisualFeedback()
		if feedback == nil {
			continue
		}

		ratio := flash.FlashTimer / flash.FlashDuration
		currentIntensity := flash.FlashIntensity * ratio

		// Blend tint multipliers toward flash color; 1.0 = no tint, flash shifts toward color
		tr := 1.0 + currentIntensity*(float64(flash.FlashColor.R)/255.0-1.0)
		tg := 1.0 + currentIntensity*(float64(flash.FlashColor.G)/255.0-1.0)
		tb := 1.0 + currentIntensity*(float64(flash.FlashColor.B)/255.0-1.0)

		feedback.SetTint(tr, tg, tb, 1.0)
	}
}
