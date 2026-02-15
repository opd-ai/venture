// Package engine provides the MeleeSwingArcSystem which creates genre-aware
// visual arc effects during melee attacks. When an entity transitions to the
// "attack" animation state, this system activates a MeleeSwingArcComponent
// whose phase, color, and opacity are driven by the weapon material, entity
// facing direction, and genre preset. The render pipeline reads this component
// to draw a fading semicircular slash overlay at the entity's position.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// swingArcGenrePreset holds genre-specific color modifiers for swing arcs.
type swingArcGenrePreset struct {
	R, G, B        float64
	IntensityScale float64 // 0.0–1.0 multiplier on arc opacity
	ThicknessScale float64 // multiplier on arc thickness
}

// swingArcMaterialColor defines base arc color per weapon material.
type swingArcMaterialColor struct {
	R, G, B float64
}

var swingArcMaterialColors = map[sprites.MaterialType]swingArcMaterialColor{
	sprites.MaterialMetal:   {R: 1.0, G: 0.92, B: 0.78},
	sprites.MaterialCrystal: {R: 0.6, G: 0.95, B: 1.0},
	sprites.MaterialEnergy:  {R: 0.85, G: 0.5, B: 1.0},
	sprites.MaterialWood:    {R: 0.8, G: 0.65, B: 0.4},
	sprites.MaterialLeather: {R: 0.75, G: 0.6, B: 0.45},
	sprites.MaterialCloth:   {R: 0.9, G: 0.9, B: 0.95},
}

// MeleeSwingArcComponent holds visual state for a melee swing arc overlay.
type MeleeSwingArcComponent struct {
	Active        bool
	Phase         float64 // 0.0 (start) to 1.0 (fully faded)
	ArcAngleStart float64 // radians
	ArcAngleEnd   float64 // radians
	ArcRadius     float64 // pixels
	Thickness     float64 // pixels
	R, G, B, A    float64 // current color with opacity
	FadeRate      float64 // phase advancement per second
	PrevAttacking bool    // tracks previous frame's attack state
}

// Type returns the component type identifier.
func (c *MeleeSwingArcComponent) Type() string { return "melee_swing_arc" }

// MeleeSwingArcSystem monitors attack animation transitions and manages
// per-entity swing arc visuals. It lazily attaches MeleeSwingArcComponent
// to entities with both animation and attack components.
type MeleeSwingArcSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string
	preset  swingArcGenrePreset

	updateInterval float64
	timeSinceCheck float64
}

// NewMeleeSwingArcSystem creates a new melee swing arc visual system.
func NewMeleeSwingArcSystem(world *World, seed int64) *MeleeSwingArcSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "melee_swing_arc")
		logEntry.Debug("melee swing arc system created")
	}

	sys := &MeleeSwingArcSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		updateInterval: 1.0,
	}
	sys.preset = sys.getPreset(sys.genreID)
	return sys
}

// SetGenre configures genre-aware arc color and intensity.
func (s *MeleeSwingArcSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getPreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for swing arc")
	}
}

// getPreset returns genre-specific swing arc color modifiers.
func (s *MeleeSwingArcSystem) getPreset(genreID string) swingArcGenrePreset {
	switch genreID {
	case "horror":
		return swingArcGenrePreset{R: 0.8, G: 0.2, B: 0.15, IntensityScale: 0.7, ThicknessScale: 1.3}
	case "cyberpunk":
		return swingArcGenrePreset{R: 0.3, G: 1.0, B: 1.0, IntensityScale: 1.2, ThicknessScale: 0.8}
	case "sci-fi", "scifi":
		return swingArcGenrePreset{R: 0.7, G: 0.85, B: 1.0, IntensityScale: 1.1, ThicknessScale: 0.9}
	case "post-apocalyptic", "postapoc":
		return swingArcGenrePreset{R: 1.0, G: 0.7, B: 0.3, IntensityScale: 0.6, ThicknessScale: 1.1}
	case "fantasy":
		return swingArcGenrePreset{R: 1.0, G: 0.95, B: 0.8, IntensityScale: 1.0, ThicknessScale: 1.0}
	default:
		return swingArcGenrePreset{R: 1.0, G: 1.0, B: 1.0, IntensityScale: 1.0, ThicknessScale: 1.0}
	}
}

// Update processes all entities, activating swing arcs on attack start and
// advancing/fading active arcs each frame.
func (s *MeleeSwingArcSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	fullScan := false
	if s.timeSinceCheck >= s.updateInterval {
		s.timeSinceCheck = 0
		fullScan = true
	}

	for _, entity := range entities {
		comp, _ := entity.GetComponent("melee_swing_arc")
		arc, hasArc := comp.(*MeleeSwingArcComponent)

		if !hasArc {
			if !fullScan {
				continue
			}
			if !entity.HasComponent("animation") || !entity.HasComponent("attack") {
				continue
			}
			arc = &MeleeSwingArcComponent{}
			entity.AddComponent(arc)
		}

		animComp, _ := entity.GetComponent("animation")
		anim, ok := animComp.(*AnimationComponent)
		if !ok || anim == nil {
			arc.PrevAttacking = false
			continue
		}

		attacking := anim.CurrentState == AnimationStateAttack
		justStarted := attacking && !arc.PrevAttacking
		arc.PrevAttacking = attacking

		if justStarted {
			s.activateArc(entity, arc, anim)
		}

		if arc.Active {
			arc.Phase += deltaTime * arc.FadeRate
			arc.A = (1.0 - arc.Phase) * s.preset.IntensityScale
			if arc.A < 0 {
				arc.A = 0
			}
			if arc.Phase >= 1.0 {
				arc.Active = false
				arc.Phase = 0
			}
		}
	}
}

// activateArc initializes the swing arc component for a new attack.
func (s *MeleeSwingArcSystem) activateArc(entity *Entity, arc *MeleeSwingArcComponent, anim *AnimationComponent) {
	arc.Active = true
	arc.Phase = 0

	// Determine arc angles from facing direction
	start, end := s.facingToArcAngles(anim.Facing)
	arc.ArcAngleStart = start
	arc.ArcAngleEnd = end

	// Base arc radius from attack range, defaulting to 20px
	arc.ArcRadius = 20.0
	if attackComp, has := entity.GetComponent("attack"); has {
		if atk, ok := attackComp.(*AttackComponent); ok && atk.Range > 0 {
			arc.ArcRadius = math.Min(atk.Range*0.8, 48.0)
		}
	}

	// Fade rate: arc lasts ~0.3s
	arc.FadeRate = 3.3

	// Resolve weapon material for color
	matColor := s.resolveWeaponColor(entity)

	// Blend material color with genre preset
	arc.R = matColor.R*0.6 + s.preset.R*0.4
	arc.G = matColor.G*0.6 + s.preset.G*0.4
	arc.B = matColor.B*0.6 + s.preset.B*0.4
	arc.A = s.preset.IntensityScale

	arc.Thickness = 3.0 * s.preset.ThicknessScale
}

// facingToArcAngles converts a direction to swing arc angle range (radians).
func (s *MeleeSwingArcSystem) facingToArcAngles(facing Direction) (float64, float64) {
	switch facing {
	case DirRight:
		return -math.Pi / 3, math.Pi / 3
	case DirLeft:
		return 2 * math.Pi / 3, 4 * math.Pi / 3
	case DirUp:
		return -5 * math.Pi / 6, -math.Pi / 6
	case DirDown:
		return math.Pi / 6, 5 * math.Pi / 6
	default:
		return -math.Pi / 3, math.Pi / 3
	}
}

// resolveWeaponColor determines arc color from the entity's equipped weapon material.
func (s *MeleeSwingArcSystem) resolveWeaponColor(entity *Entity) swingArcMaterialColor {
	comp, has := entity.GetComponent("equipment")
	if !has || comp == nil {
		return swingArcMaterialColor{R: 0.9, G: 0.9, B: 0.9}
	}
	equip, ok := comp.(*EquipmentComponent)
	if !ok {
		return swingArcMaterialColor{R: 0.9, G: 0.9, B: 0.9}
	}
	weapon := equip.GetEquipped(SlotMainHand)
	if weapon == nil {
		return swingArcMaterialColor{R: 0.9, G: 0.9, B: 0.9}
	}

	mat := sprites.GetMaterialTypeFromTags(weapon.Tags, s.genreID)
	if c, ok := swingArcMaterialColors[mat]; ok {
		return s.applyRarityBoost(c, weapon.Rarity)
	}
	return swingArcMaterialColor{R: 0.9, G: 0.9, B: 0.9}
}

// applyRarityBoost increases color intensity for higher rarity weapons.
func (s *MeleeSwingArcSystem) applyRarityBoost(base swingArcMaterialColor, rarity item.Rarity) swingArcMaterialColor {
	var boost float64
	switch rarity {
	case item.RarityUncommon:
		boost = 0.05
	case item.RarityRare:
		boost = 0.10
	case item.RarityEpic:
		boost = 0.15
	case item.RarityLegendary:
		boost = 0.20
	default:
		return base
	}
	return swingArcMaterialColor{
		R: math.Min(base.R+boost, 1.0),
		G: math.Min(base.G+boost, 1.0),
		B: math.Min(base.B+boost, 1.0),
	}
}
