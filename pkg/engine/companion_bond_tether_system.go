// Package engine provides the CompanionBondTetherSystem for genre-aware visual
// tethers between companions and their owners. When a companion entity is within
// range of its owner, the system writes tether geometry and color data into
// CompanionBondTetherComponent for the render system to draw a pulsing line
// connecting the two. Genre presets control color, pulse speed, and opacity:
// fantasy uses a golden thread, horror a crimson flicker, sci-fi a cyan beam,
// cyberpunk a neon magenta wire, and post-apocalyptic an amber rope.
package engine

import (
	"image/color"
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CompanionBondTetherComponent stores per-companion tether visual state.
// Pure data — all logic lives in CompanionBondTetherSystem.
type CompanionBondTetherComponent struct {
	// Active indicates the tether is currently visible.
	Active bool
	// OwnerX, OwnerY is the world position of the owner endpoint.
	OwnerX, OwnerY float64
	// CompanionX, CompanionY is the world position of the companion endpoint.
	CompanionX, CompanionY float64
	// TetherColor is the current RGBA color of the tether line.
	TetherColor color.RGBA
	// Opacity ranges 0.0–1.0 controlling tether alpha.
	Opacity float64
	// PulsePhase tracks the current pulse animation phase (radians).
	PulsePhase float64
	// Thickness is the tether line width in pixels.
	Thickness float64
	// LoyaltyFactor is a 0.0–1.0 factor derived from companion loyalty.
	LoyaltyFactor float64
}

// Type returns the component type identifier.
func (c *CompanionBondTetherComponent) Type() string { return "companion_bond_tether" }

// companionBondTetherPreset holds genre-specific tether configuration.
type companionBondTetherPreset struct {
	BaseColor  color.RGBA // Base tether color
	PulseSpeed float64    // Pulse cycles per second
	PulseDepth float64    // Amplitude of opacity pulse (0.0–0.5)
	MinOpacity float64    // Minimum opacity when visible
	MaxOpacity float64    // Maximum opacity at full loyalty
	Thickness  float64    // Base line thickness in pixels
	Jitter     float64    // Position jitter for horror/cyberpunk
}

// CompanionBondTetherSystem writes tether geometry and color into
// CompanionBondTetherComponent for companions near their owners.
// It reads CompanionComponent.OwnerID and PositionComponent to
// calculate endpoints, and uses loyalty to modulate visual intensity.
type CompanionBondTetherSystem struct {
	world   *World
	seed    int64
	rng     *rand.Rand
	logger  *logrus.Entry
	genreID string
	preset  companionBondTetherPreset

	// Maximum distance in pixels before tether fades out.
	maxDistance float64
	// Distance beyond which tether starts fading (fade zone = maxDistance - fadeStart).
	fadeStart float64
}

// NewCompanionBondTetherSystem creates a new companion bond tether system.
func NewCompanionBondTetherSystem(world *World, seed int64) *CompanionBondTetherSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "companion_bond_tether",
	})

	sys := &CompanionBondTetherSystem{
		world:       world,
		seed:        seed,
		rng:         rand.New(rand.NewSource(seed)),
		logger:      logger,
		genreID:     "fantasy",
		maxDistance:  250.0,
		fadeStart:   150.0,
	}
	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("companion bond tether system created")
	return sys
}

// SetGenre configures genre-specific tether appearance.
func (s *CompanionBondTetherSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre preset applied")
	}
}

// getGenrePreset returns tether configuration for the given genre.
func (s *CompanionBondTetherSystem) getGenrePreset(genreID string) companionBondTetherPreset {
	switch genreID {
	case "horror":
		return companionBondTetherPreset{
			BaseColor:  color.RGBA{R: 180, G: 30, B: 30, A: 255},
			PulseSpeed: 2.5,
			PulseDepth: 0.35,
			MinOpacity: 0.15,
			MaxOpacity: 0.7,
			Thickness:  1.5,
			Jitter:     1.2,
		}
	case "scifi":
		return companionBondTetherPreset{
			BaseColor:  color.RGBA{R: 60, G: 200, B: 255, A: 255},
			PulseSpeed: 1.2,
			PulseDepth: 0.15,
			MinOpacity: 0.3,
			MaxOpacity: 0.85,
			Thickness:  1.0,
			Jitter:     0.0,
		}
	case "cyberpunk":
		return companionBondTetherPreset{
			BaseColor:  color.RGBA{R: 255, G: 50, B: 200, A: 255},
			PulseSpeed: 3.0,
			PulseDepth: 0.3,
			MinOpacity: 0.25,
			MaxOpacity: 0.9,
			Thickness:  1.0,
			Jitter:     0.8,
		}
	case "postapoc":
		return companionBondTetherPreset{
			BaseColor:  color.RGBA{R: 200, G: 150, B: 50, A: 255},
			PulseSpeed: 0.6,
			PulseDepth: 0.2,
			MinOpacity: 0.2,
			MaxOpacity: 0.65,
			Thickness:  2.0,
			Jitter:     0.3,
		}
	default: // fantasy
		return companionBondTetherPreset{
			BaseColor:  color.RGBA{R: 255, G: 215, B: 80, A: 255},
			PulseSpeed: 0.8,
			PulseDepth: 0.2,
			MinOpacity: 0.2,
			MaxOpacity: 0.8,
			Thickness:  1.5,
			Jitter:     0.0,
		}
	}
}

// Update iterates all entities, finds those with CompanionComponent, looks up
// their owner, and writes tether visual state into CompanionBondTetherComponent.
func (s *CompanionBondTetherSystem) Update(entities []*Entity, deltaTime float64) {
	if s.world == nil {
		return
	}

	for _, entity := range entities {
		compRaw, hasComp := entity.GetComponent("companion")
		if !hasComp || compRaw == nil {
			continue
		}
		comp, ok := compRaw.(*CompanionComponent)
		if !ok || comp.OwnerID == 0 {
			continue
		}

		ownerEntity, ownerExists := s.world.GetEntity(comp.OwnerID)
		if !ownerExists {
			s.deactivateTether(entity)
			continue
		}

		ownerPos := ownerEntity.GetPosition()
		companionPos := entity.GetPosition()
		if ownerPos == nil || companionPos == nil {
			s.deactivateTether(entity)
			continue
		}

		dx := ownerPos.X - companionPos.X
		dy := ownerPos.Y - companionPos.Y
		dist := math.Sqrt(dx*dx + dy*dy)

		tether := s.getOrCreateTether(entity)

		if dist > s.maxDistance {
			tether.Active = false
			continue
		}

		tether.Active = true
		tether.OwnerX = ownerPos.X
		tether.OwnerY = ownerPos.Y
		tether.CompanionX = companionPos.X
		tether.CompanionY = companionPos.Y

		// Loyalty factor: 0.0 at loyalty 0, 1.0 at loyalty 100
		loyaltyNorm := math.Max(0.0, math.Min(comp.Loyalty/100.0, 1.0))
		tether.LoyaltyFactor = loyaltyNorm

		// Advance pulse phase
		tether.PulsePhase += deltaTime * s.preset.PulseSpeed * 2.0 * math.Pi
		if tether.PulsePhase > 2.0*math.Pi {
			tether.PulsePhase -= 2.0 * math.Pi
		}

		// Compute opacity: base from loyalty, modulated by pulse, faded by distance
		baseOpacity := s.preset.MinOpacity + loyaltyNorm*(s.preset.MaxOpacity-s.preset.MinOpacity)
		pulse := s.preset.PulseDepth * math.Sin(tether.PulsePhase)
		opacity := math.Max(0.0, math.Min(baseOpacity+pulse, 1.0))

		// Distance fade
		if dist > s.fadeStart {
			fadeFactor := 1.0 - (dist-s.fadeStart)/(s.maxDistance-s.fadeStart)
			opacity *= math.Max(0.0, fadeFactor)
		}
		tether.Opacity = opacity

		// Color with loyalty-scaled alpha
		alphaScaled := uint8(math.Min(255.0, opacity*255.0))
		tether.TetherColor = color.RGBA{
			R: s.preset.BaseColor.R,
			G: s.preset.BaseColor.G,
			B: s.preset.BaseColor.B,
			A: alphaScaled,
		}

		// Thickness modulated by loyalty
		tether.Thickness = s.preset.Thickness * (0.5 + 0.5*loyaltyNorm)

		// Apply jitter for horror/cyberpunk genres
		if s.preset.Jitter > 0 {
			tether.OwnerX += (s.rng.Float64() - 0.5) * s.preset.Jitter * 2.0
			tether.OwnerY += (s.rng.Float64() - 0.5) * s.preset.Jitter * 2.0
			tether.CompanionX += (s.rng.Float64() - 0.5) * s.preset.Jitter * 2.0
			tether.CompanionY += (s.rng.Float64() - 0.5) * s.preset.Jitter * 2.0
		}
	}
}

// getOrCreateTether retrieves or lazily adds the tether component.
func (s *CompanionBondTetherSystem) getOrCreateTether(entity *Entity) *CompanionBondTetherComponent {
	raw, exists := entity.GetComponent("companion_bond_tether")
	if exists && raw != nil {
		if t, ok := raw.(*CompanionBondTetherComponent); ok {
			return t
		}
	}
	t := &CompanionBondTetherComponent{}
	entity.AddComponent(t)
	return t
}

// deactivateTether marks the tether inactive if present.
func (s *CompanionBondTetherSystem) deactivateTether(entity *Entity) {
	raw, exists := entity.GetComponent("companion_bond_tether")
	if !exists || raw == nil {
		return
	}
	if t, ok := raw.(*CompanionBondTetherComponent); ok {
		t.Active = false
	}
}
