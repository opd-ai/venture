package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// genreShadowPreset holds genre-specific shadow color and opacity scaling.
type genreShadowPreset struct {
	R, G, B      float64
	OpacityScale float64
}

// EntityDropShadowSystem computes drop shadow parameters for entities that have
// both a position and a collider (or sprite). It scales shadow size from the
// entity's collision box, tints the shadow color by genre, and lazily attaches
// a DropShadowComponent to qualifying entities.
type EntityDropShadowSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string
	preset  genreShadowPreset

	// Throttle: only process new entities periodically
	updateInterval float64
	timeSinceCheck float64
}

// NewEntityDropShadowSystem creates a new drop shadow system.
func NewEntityDropShadowSystem(world *World, seed int64) *EntityDropShadowSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "entity_drop_shadow")
	}

	sys := &EntityDropShadowSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		updateInterval: 1.0,
	}
	sys.preset = sys.getPreset(sys.genreID)

	if logEntry != nil {
		logEntry.Debug("EntityDropShadowSystem created")
	}
	return sys
}

// SetGenre configures genre-aware shadow color and opacity.
func (s *EntityDropShadowSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getPreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for entity drop shadow")
	}
}

// getPreset returns a shadow color/opacity preset based on genre.
func (s *EntityDropShadowSystem) getPreset(genreID string) genreShadowPreset {
	switch genreID {
	case "horror":
		return genreShadowPreset{R: 0.05, G: 0.0, B: 0.0, OpacityScale: 1.4}
	case "cyberpunk":
		return genreShadowPreset{R: 0.0, G: 0.02, B: 0.06, OpacityScale: 1.1}
	case "sci-fi", "scifi":
		return genreShadowPreset{R: 0.02, G: 0.02, B: 0.04, OpacityScale: 0.9}
	case "post-apocalyptic", "postapoc":
		return genreShadowPreset{R: 0.04, G: 0.03, B: 0.01, OpacityScale: 1.2}
	case "fantasy":
		return genreShadowPreset{R: 0.02, G: 0.01, B: 0.0, OpacityScale: 1.0}
	default:
		return genreShadowPreset{R: 0.0, G: 0.0, B: 0.0, OpacityScale: 1.0}
	}
}

// Update iterates over entities and ensures qualifying ones have a correctly
// sized DropShadowComponent. Full scan runs every updateInterval seconds;
// otherwise only entities already having the component are updated.
func (s *EntityDropShadowSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	fullScan := false
	if s.timeSinceCheck >= s.updateInterval {
		s.timeSinceCheck = 0
		fullScan = true
	}

	for _, entity := range entities {
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		comp, _ := entity.GetComponent("drop_shadow")
		shadow, hasShadow := comp.(*DropShadowComponent)

		if !hasShadow {
			if !fullScan {
				continue
			}
			// Attach shadow to entities with a collider or sprite
			if entity.GetCollider() == nil && entity.GetSprite() == nil {
				continue
			}
			shadow = NewDropShadowComponent()
			entity.AddComponent(shadow)
		}

		s.computeShadow(entity, shadow)
	}
}

// computeShadow derives shadow dimensions and color from entity size and genre.
func (s *EntityDropShadowSystem) computeShadow(entity *Entity, shadow *DropShadowComponent) {
	// Determine entity width from collider, falling back to sprite
	width := 32.0
	height := 32.0
	if col := entity.GetCollider(); col != nil {
		width = col.Width
		height = col.Height
	} else if spr := entity.GetSprite(); spr != nil {
		width = spr.Width
		height = spr.Height
	}

	// Shadow is an ellipse: 70% of entity width, 25% of entity height
	shadow.ShadowWidth = width * 0.70
	shadow.ShadowHeight = height * 0.25
	if shadow.ShadowWidth < 6 {
		shadow.ShadowWidth = 6
	}
	if shadow.ShadowHeight < 3 {
		shadow.ShadowHeight = 3
	}

	// Base opacity with genre scaling
	shadow.Opacity = clampFloat(0.35*s.preset.OpacityScale, 0.1, 0.7)

	// Genre-tinted shadow color
	shadow.ColorR = s.preset.R
	shadow.ColorG = s.preset.G
	shadow.ColorB = s.preset.B

	// Shadow offset: slightly below entity center
	shadow.OffsetX = 0
	shadow.OffsetY = height * 0.4

	shadow.Enabled = true
}
