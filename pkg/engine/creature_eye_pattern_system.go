// Package engine provides the CreatureEyePatternSystem which renders creature-type-specific
// eye patterns onto nonhumanoid entity sprites. Each creature form (arachnid, serpent, insect,
// blob, flying, multi-limbed, mechanical) gets a distinct eye arrangement that makes it
// instantly recognizable from above:
//   - Arachnids: 8 eyes in characteristic spider arrangement
//   - Serpents: 2 slit-pupil eyes
//   - Insects: 2 compound faceted eyes
//   - Blobs: Internal nucleus "eye" visible through translucent body
//   - Flying: 2 forward-facing raptor eyes
//   - Multi-limbed: Asymmetric random eye cluster (3-6 eyes)
//   - Mechanical: 1-3 sensor lights arranged geometrically
//   - Undead: Empty socket eyes or ghostly wisps
//
// The system reads CreatureVisualComponent.Form to determine eye pattern, applies
// genre-specific coloring, and marks sprites dirty for re-rendering.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// CreatureEyePatternComponent stores creature-type-specific eye layout parameters.
// This is a pure data component that feeds into the sprite rendering pipeline.
type CreatureEyePatternComponent struct {
	// EyePattern identifies which eye layout to use ("arachnid_8", "serpent_slit", etc.)
	EyePattern string

	// EyeCount is the number of eyes to render (1-8 depending on creature type)
	EyeCount int

	// EyePositions stores relative eye positions as [x,y] pairs (0.0-1.0 within head area)
	// Each pair represents one eye. For 8 eyes, there are 16 values.
	EyePositions []float64

	// EyeSizes stores relative size for each eye (0.5-2.0, where 1.0 is standard)
	EyeSizes []float64

	// PrimaryEyeColor (RGB 0.0-1.0) - main eye/iris color
	EyeR, EyeG, EyeB float64

	// PupilStyle controls pupil shape ("round", "slit_vertical", "slit_horizontal", "faceted", "none")
	PupilStyle string

	// GlowIntensity for mechanical/magical eyes (0.0-1.0)
	GlowIntensity float64

	// Asymmetric indicates random/chaotic eye placement (for multi-limbed horrors)
	Asymmetric bool

	// Dirty flag for lazy re-rendering
	Dirty bool
}

// Type implements the Component interface.
func (c *CreatureEyePatternComponent) Type() string { return "creature_eye_pattern" }

// NewCreatureEyePatternComponent creates a default component with humanoid settings.
func NewCreatureEyePatternComponent() *CreatureEyePatternComponent {
	return &CreatureEyePatternComponent{
		EyePattern:    "humanoid_2",
		EyeCount:      2,
		EyePositions:  []float64{0.35, 0.55, 0.65, 0.55}, // Left and right eye
		EyeSizes:      []float64{1.0, 1.0},
		EyeR:          0.3,
		EyeG:          0.2,
		EyeB:          0.1,
		PupilStyle:    "round",
		GlowIntensity: 0.0,
		Asymmetric:    false,
		Dirty:         true,
	}
}

// creatureEyeConfig holds per-creature-form eye pattern configuration.
type creatureEyeConfig struct {
	Pattern      string
	EyeCount     int
	Positions    []float64 // Relative positions within head bounding box
	Sizes        []float64 // Relative sizes
	PupilStyle   string
	BaseGlow     float64
	Asymmetric   bool
	ColorVariant string // "warm", "cold", "neutral", "spectral"
}

// CreatureEyePatternSystem assigns creature-type-specific eye patterns to
// nonhumanoid entities based on their CreatureVisualComponent.Form.
type CreatureEyePatternSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID     string
	lastGenreID string
	initialized bool

	updateInterval float64
	timeSinceCheck float64

	// Predefined eye patterns per creature form
	patterns map[CreatureForm]creatureEyeConfig

	// Genre-specific eye color palettes
	genreColors map[string][3]float64
}

// NewCreatureEyePatternSystem creates a new creature eye pattern system.
func NewCreatureEyePatternSystem(world *World, seed int64) *CreatureEyePatternSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "creature_eye_pattern")
	}

	sys := &CreatureEyePatternSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		lastGenreID:    "",
		updateInterval: 0.5,
		patterns:       buildCreatureEyePatterns(),
		genreColors:    buildGenreEyeColors(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("creature eye pattern system created")
	}
	return sys
}

// SetGenre configures the active genre for eye color computation.
func (s *CreatureEyePatternSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for creature eye patterns")
	}
}

// Update assigns eye patterns to creatures that need them.
func (s *CreatureEyePatternSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	if s.timeSinceCheck < s.updateInterval {
		return
	}
	s.timeSinceCheck = 0

	genreChanged := s.genreID != s.lastGenreID || !s.initialized
	s.lastGenreID = s.genreID
	s.initialized = true

	count := 0
	for _, entity := range entities {
		// Skip entities without creature visual classification
		comp, ok := entity.GetComponent("creature_visual")
		if !ok {
			continue
		}
		cv, ok := comp.(*CreatureVisualComponent)
		if !ok {
			continue
		}

		// Skip humanoids - they use the NpcFacialDetailComponent system
		if cv.Form == FormHumanoid {
			continue
		}

		// Get or create eye pattern component
		eyeComp := s.getOrCreateComponent(entity)

		// Recompute if genre changed or component is new/dirty
		if genreChanged || eyeComp.Dirty {
			s.assignEyePattern(entity, cv, eyeComp)
			count++
		}
	}

	if s.logger != nil && count > 0 && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"genre":            s.genreID,
			"patterns_applied": count,
		}).Debug("creature eye patterns assigned")
	}
}

// assignEyePattern configures the eye pattern based on creature form.
func (s *CreatureEyePatternSystem) assignEyePattern(entity *Entity, cv *CreatureVisualComponent, eyeComp *CreatureEyePatternComponent) {
	config, ok := s.patterns[cv.Form]
	if !ok {
		config = s.patterns[FormQuadruped] // Default fallback
	}

	// Copy base pattern
	eyeComp.EyePattern = config.Pattern
	eyeComp.EyeCount = config.EyeCount
	eyeComp.PupilStyle = config.PupilStyle
	eyeComp.GlowIntensity = config.BaseGlow
	eyeComp.Asymmetric = config.Asymmetric

	// Copy positions with slight randomization for variety
	eyeComp.EyePositions = make([]float64, len(config.Positions))
	copy(eyeComp.EyePositions, config.Positions)

	// Add per-entity position jitter for asymmetric creatures
	if config.Asymmetric {
		for i := range eyeComp.EyePositions {
			eyeComp.EyePositions[i] += (s.rng.Float64() - 0.5) * 0.15
			eyeComp.EyePositions[i] = clampEyePattern(eyeComp.EyePositions[i], 0.1, 0.9)
		}
	} else {
		// Minor jitter for natural variety
		for i := range eyeComp.EyePositions {
			eyeComp.EyePositions[i] += (s.rng.Float64() - 0.5) * 0.04
			eyeComp.EyePositions[i] = clampEyePattern(eyeComp.EyePositions[i], 0.15, 0.85)
		}
	}

	// Copy sizes with variation
	eyeComp.EyeSizes = make([]float64, len(config.Sizes))
	for i, size := range config.Sizes {
		variation := 1.0 + (s.rng.Float64()-0.5)*0.3
		eyeComp.EyeSizes[i] = clampEyePattern(size*variation, 0.3, 2.5)
	}

	// Apply genre-specific eye color
	s.applyGenreColor(eyeComp, cv.Form, config.ColorVariant)

	// For multi-limbed horrors, randomize eye count within range
	if cv.Form == FormMultiLimbed {
		eyeComp.EyeCount = 3 + s.rng.Intn(4) // 3-6 eyes
		s.generateAsymmetricEyes(eyeComp)
	}

	eyeComp.Dirty = false

	// Mark sprite as needing re-render
	if animComp, ok := entity.GetComponent("animation"); ok {
		if anim, ok := animComp.(*AnimationComponent); ok {
			anim.Dirty = true
		}
	}
}

// applyGenreColor sets eye color based on genre and creature color variant.
func (s *CreatureEyePatternSystem) applyGenreColor(eyeComp *CreatureEyePatternComponent, form CreatureForm, variant string) {
	baseColor, ok := s.genreColors[s.genreID]
	if !ok {
		baseColor = s.genreColors["fantasy"]
	}

	// Modify base color based on variant and creature form
	r, g, b := baseColor[0], baseColor[1], baseColor[2]

	switch variant {
	case "warm":
		r = clampEyePattern(r+0.15, 0, 1)
		g = clampEyePattern(g-0.05, 0, 1)
		b = clampEyePattern(b-0.10, 0, 1)
	case "cold":
		r = clampEyePattern(r-0.10, 0, 1)
		g = clampEyePattern(g+0.05, 0, 1)
		b = clampEyePattern(b+0.15, 0, 1)
	case "spectral":
		// Ghostly/undead - desaturate and brighten
		avg := (r + g + b) / 3.0
		r = clampEyePattern(avg+0.2, 0, 1)
		g = clampEyePattern(avg+0.3, 0, 1)
		b = clampEyePattern(avg+0.35, 0, 1)
	}

	// Form-specific color adjustments
	switch form {
	case FormArachnid:
		// Spiders often have dark, beady eyes
		r = clampEyePattern(r*0.4, 0.1, 0.5)
		g = clampEyePattern(g*0.4, 0.1, 0.5)
		b = clampEyePattern(b*0.4, 0.1, 0.5)
	case FormSerpentine:
		// Reptilian yellow/gold eyes
		r = clampEyePattern(0.8+s.rng.Float64()*0.2, 0.7, 1.0)
		g = clampEyePattern(0.6+s.rng.Float64()*0.2, 0.5, 0.8)
		b = clampEyePattern(0.1+s.rng.Float64()*0.1, 0.0, 0.3)
	case FormInsect:
		// Compound eyes often have iridescent quality
		r = clampEyePattern(r*1.2, 0.2, 1.0)
		g = clampEyePattern(g*0.8, 0.1, 0.8)
		b = clampEyePattern(b*1.1, 0.2, 1.0)
	case FormMechanical:
		// Bright sensor glow
		eyeComp.GlowIntensity = clampEyePattern(eyeComp.GlowIntensity+0.3, 0.2, 1.0)
	case FormMultiLimbed:
		// Eldritch creatures - slightly unnatural colors
		r = clampEyePattern(r+0.1, 0, 1)
		g = clampEyePattern(g*0.7, 0, 1)
		b = clampEyePattern(b+0.2, 0, 1)
	}

	// Add per-entity variation
	r += (s.rng.Float64() - 0.5) * 0.1
	g += (s.rng.Float64() - 0.5) * 0.1
	b += (s.rng.Float64() - 0.5) * 0.1

	eyeComp.EyeR = clampEyePattern(r, 0, 1)
	eyeComp.EyeG = clampEyePattern(g, 0, 1)
	eyeComp.EyeB = clampEyePattern(b, 0, 1)
}

// generateAsymmetricEyes creates random eye positions for multi-limbed horrors.
func (s *CreatureEyePatternSystem) generateAsymmetricEyes(eyeComp *CreatureEyePatternComponent) {
	count := eyeComp.EyeCount
	positions := make([]float64, count*2)
	sizes := make([]float64, count)

	// Generate clustered but chaotic positions
	centerX := 0.4 + s.rng.Float64()*0.2
	centerY := 0.3 + s.rng.Float64()*0.2

	for i := 0; i < count; i++ {
		// Random angle and distance from center
		angle := s.rng.Float64() * 2 * math.Pi
		dist := 0.08 + s.rng.Float64()*0.25

		positions[i*2] = clampEyePattern(centerX+math.Cos(angle)*dist, 0.1, 0.9)
		positions[i*2+1] = clampEyePattern(centerY+math.Sin(angle)*dist, 0.1, 0.9)

		// Varied sizes - some eyes larger than others
		sizes[i] = 0.6 + s.rng.Float64()*1.2
	}

	eyeComp.EyePositions = positions
	eyeComp.EyeSizes = sizes
}

// getOrCreateComponent retrieves or lazily creates the eye pattern component.
func (s *CreatureEyePatternSystem) getOrCreateComponent(entity *Entity) *CreatureEyePatternComponent {
	comp, ok := entity.GetComponent("creature_eye_pattern")
	if ok {
		if ec, ok := comp.(*CreatureEyePatternComponent); ok {
			return ec
		}
	}
	ec := NewCreatureEyePatternComponent()
	entity.AddComponent(ec)
	return ec
}

// buildCreatureEyePatterns returns predefined eye patterns for each creature form.
func buildCreatureEyePatterns() map[CreatureForm]creatureEyeConfig {
	return map[CreatureForm]creatureEyeConfig{
		// Quadruped: 2 forward-facing eyes near top of head (from above)
		FormQuadruped: {
			Pattern:  "quadruped_2",
			EyeCount: 2,
			Positions: []float64{
				0.30, 0.35, // Left eye
				0.70, 0.35, // Right eye
			},
			Sizes:        []float64{1.0, 1.0},
			PupilStyle:   "round",
			BaseGlow:     0.0,
			Asymmetric:   false,
			ColorVariant: "warm",
		},

		// Arachnid: 8 eyes in characteristic spider arrangement
		// From above: 4 larger eyes at front, 4 smaller at sides
		FormArachnid: {
			Pattern:  "arachnid_8",
			EyeCount: 8,
			Positions: []float64{
				0.35, 0.25, // Front-left primary
				0.65, 0.25, // Front-right primary
				0.25, 0.30, // Side-left primary
				0.75, 0.30, // Side-right primary
				0.30, 0.40, // Front-left secondary
				0.70, 0.40, // Front-right secondary
				0.20, 0.45, // Side-left secondary
				0.80, 0.45, // Side-right secondary
			},
			Sizes:        []float64{1.2, 1.2, 1.0, 1.0, 0.7, 0.7, 0.5, 0.5},
			PupilStyle:   "none", // Spiders have simple dark eyes
			BaseGlow:     0.0,
			Asymmetric:   false,
			ColorVariant: "neutral",
		},

		// Serpent: 2 slit-pupil reptilian eyes
		FormSerpentine: {
			Pattern:  "serpent_slit",
			EyeCount: 2,
			Positions: []float64{
				0.30, 0.40, // Left eye
				0.70, 0.40, // Right eye
			},
			Sizes:        []float64{1.3, 1.3},
			PupilStyle:   "slit_vertical",
			BaseGlow:     0.0,
			Asymmetric:   false,
			ColorVariant: "warm",
		},

		// Insect: 2 large compound eyes on sides of head
		FormInsect: {
			Pattern:  "insect_compound",
			EyeCount: 2,
			Positions: []float64{
				0.22, 0.40, // Left compound eye
				0.78, 0.40, // Right compound eye
			},
			Sizes:        []float64{1.8, 1.8}, // Large compound eyes
			PupilStyle:   "faceted",
			BaseGlow:     0.0,
			Asymmetric:   false,
			ColorVariant: "cold",
		},

		// Flying: 2 forward-facing predator eyes
		FormFlying: {
			Pattern:  "flying_raptor",
			EyeCount: 2,
			Positions: []float64{
				0.35, 0.30, // Left eye
				0.65, 0.30, // Right eye
			},
			Sizes:        []float64{1.4, 1.4},
			PupilStyle:   "round",
			BaseGlow:     0.0,
			Asymmetric:   false,
			ColorVariant: "warm",
		},

		// Blob: Single internal "nucleus" eye visible through body
		FormBlob: {
			Pattern:  "blob_nucleus",
			EyeCount: 1,
			Positions: []float64{
				0.45, 0.42, // Central nucleus
			},
			Sizes:        []float64{2.0}, // Large central feature
			PupilStyle:   "none",
			BaseGlow:     0.15, // Slight internal glow
			Asymmetric:   false,
			ColorVariant: "cold",
		},

		// Mechanical: 1-3 sensor lights arranged geometrically
		FormMechanical: {
			Pattern:  "mechanical_sensors",
			EyeCount: 3,
			Positions: []float64{
				0.50, 0.30, // Primary central sensor
				0.30, 0.50, // Left secondary
				0.70, 0.50, // Right secondary
			},
			Sizes:        []float64{1.5, 0.8, 0.8},
			PupilStyle:   "none",
			BaseGlow:     0.6, // Strong sensor glow
			Asymmetric:   false,
			ColorVariant: "cold",
		},

		// Undead: Empty socket eyes or spectral wisps
		FormUndead: {
			Pattern:  "undead_sockets",
			EyeCount: 2,
			Positions: []float64{
				0.35, 0.40, // Left socket
				0.65, 0.40, // Right socket
			},
			Sizes:        []float64{1.1, 1.1},
			PupilStyle:   "none", // Empty sockets
			BaseGlow:     0.2,    // Faint spectral glow
			Asymmetric:   false,
			ColorVariant: "spectral",
		},

		// Multi-limbed: Chaotic eye cluster (positions generated dynamically)
		FormMultiLimbed: {
			Pattern:  "multi_limbed_cluster",
			EyeCount: 5, // Base count, will be randomized 3-6
			Positions: []float64{
				0.45, 0.35, 0.55, 0.40, 0.40, 0.50,
				0.60, 0.45, 0.50, 0.55,
			},
			Sizes:        []float64{1.0, 1.2, 0.8, 1.1, 0.9},
			PupilStyle:   "slit_horizontal", // Alien/eldritch look
			BaseGlow:     0.15,
			Asymmetric:   true,
			ColorVariant: "cold",
		},
	}
}

// buildGenreEyeColors returns genre-specific base eye colors.
func buildGenreEyeColors() map[string][3]float64 {
	return map[string][3]float64{
		"fantasy":   {0.4, 0.3, 0.2},   // Earthy brown tones
		"horror":    {0.7, 0.15, 0.1},  // Blood red undertones
		"scifi":     {0.2, 0.6, 0.8},   // Cool blue-cyan
		"cyberpunk": {0.8, 0.3, 0.7},   // Neon magenta-purple
		"postapoc":  {0.5, 0.55, 0.25}, // Sickly yellow-green
	}
}

// clampEyePattern ensures a value stays within bounds.
func clampEyePattern(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
