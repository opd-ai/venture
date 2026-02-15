// Package engine provides the FloatingDamageNumberSystem which spawns genre-aware
// floating damage numbers above entities when they take damage. It monitors
// HealthComponent changes and writes entries to FloatingDamageNumberComponent,
// connecting combat damage to visible numeric feedback.
package engine

import (
	"image/color"
	"math"

	"github.com/sirupsen/logrus"
)

// FloatingNumber represents a single floating damage number entry.
type FloatingNumber struct {
	Amount   float64    // Damage or healing amount
	OffsetX  float64    // Horizontal offset from entity center
	OffsetY  float64    // Vertical offset (negative = upward)
	Opacity  float64    // Current opacity 0.0–1.0
	Age      float64    // Seconds since spawn
	Color    color.RGBA // Display color
	Scale    float64    // Size scale (1.0 = normal, larger for crits)
	IsHeal   bool       // True if this is a healing number
	IsCrit   bool       // True if this is a critical hit
	Duration float64    // Total display duration
}

// FloatingDamageNumberComponent stores active floating damage numbers for an entity.
type FloatingDamageNumberComponent struct {
	Numbers  []FloatingNumber // Active floating numbers (max 8)
	MaxCount int              // Maximum simultaneous numbers
}

// Type returns the component type identifier.
func (f *FloatingDamageNumberComponent) Type() string {
	return "floating_damage_number"
}

// NewFloatingDamageNumberComponent creates a new component with defaults.
func NewFloatingDamageNumberComponent() *FloatingDamageNumberComponent {
	return &FloatingDamageNumberComponent{
		Numbers:  make([]FloatingNumber, 0, 8),
		MaxCount: 8,
	}
}

// floatingNumberPreset holds genre-specific visual configuration.
type floatingNumberPreset struct {
	DamageColor color.RGBA // Normal damage color
	HealColor   color.RGBA // Healing color
	CritColor   color.RGBA // Critical hit color
	RiseSpeed   float64    // Pixels per second upward
	Duration    float64    // Display duration in seconds
	CritScale   float64    // Scale multiplier for crits
	SpreadX     float64    // Horizontal spread range
}

// FloatingDamageNumberSystem monitors entity health changes and spawns floating
// damage numbers. Numbers rise upward and fade out over time with genre-aware coloring.
type FloatingDamageNumberSystem struct {
	world  *World
	logger *logrus.Entry
	seed   int64

	prevHealth map[uint64]float64
	genreID    string
	preset     floatingNumberPreset

	cleanupTimer    float64
	cleanupInterval float64
	spreadCounter   uint64 // deterministic spread offset
}

// NewFloatingDamageNumberSystem creates a new floating damage number system.
func NewFloatingDamageNumberSystem(world *World, seed int64) *FloatingDamageNumberSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "floating_damage_number",
	})

	sys := &FloatingDamageNumberSystem{
		world:           world,
		logger:          logger,
		seed:            seed,
		prevHealth:      make(map[uint64]float64, 128),
		genreID:         "fantasy",
		cleanupInterval: 5.0,
	}

	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("floating damage number system created")
	return sys
}

// SetGenre configures genre-specific number colors and behavior.
func (s *FloatingDamageNumberSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
}

// getGenrePreset returns visual configuration for the given genre.
func (s *FloatingDamageNumberSystem) getGenrePreset(genreID string) floatingNumberPreset {
	switch genreID {
	case "horror":
		return floatingNumberPreset{
			DamageColor: color.RGBA{R: 200, G: 30, B: 30, A: 255},
			HealColor:   color.RGBA{R: 80, G: 180, B: 60, A: 255},
			CritColor:   color.RGBA{R: 255, G: 10, B: 10, A: 255},
			RiseSpeed:   28.0,
			Duration:    1.0,
			CritScale:   1.6,
			SpreadX:     6.0,
		}
	case "cyberpunk":
		return floatingNumberPreset{
			DamageColor: color.RGBA{R: 0, G: 230, B: 255, A: 255},
			HealColor:   color.RGBA{R: 0, G: 255, B: 120, A: 255},
			CritColor:   color.RGBA{R: 255, G: 0, B: 200, A: 255},
			RiseSpeed:   35.0,
			Duration:    0.8,
			CritScale:   1.5,
			SpreadX:     8.0,
		}
	case "scifi":
		return floatingNumberPreset{
			DamageColor: color.RGBA{R: 200, G: 220, B: 255, A: 255},
			HealColor:   color.RGBA{R: 100, G: 255, B: 180, A: 255},
			CritColor:   color.RGBA{R: 255, G: 255, B: 100, A: 255},
			RiseSpeed:   32.0,
			Duration:    0.9,
			CritScale:   1.5,
			SpreadX:     7.0,
		}
	case "postapoc":
		return floatingNumberPreset{
			DamageColor: color.RGBA{R: 255, G: 180, B: 60, A: 255},
			HealColor:   color.RGBA{R: 120, G: 200, B: 80, A: 255},
			CritColor:   color.RGBA{R: 255, G: 100, B: 30, A: 255},
			RiseSpeed:   26.0,
			Duration:    1.1,
			CritScale:   1.7,
			SpreadX:     5.0,
		}
	default: // fantasy
		return floatingNumberPreset{
			DamageColor: color.RGBA{R: 255, G: 220, B: 100, A: 255},
			HealColor:   color.RGBA{R: 80, G: 255, B: 80, A: 255},
			CritColor:   color.RGBA{R: 255, G: 80, B: 40, A: 255},
			RiseSpeed:   30.0,
			Duration:    1.0,
			CritScale:   1.5,
			SpreadX:     6.0,
		}
	}
}

// Update checks all entities for health changes and manages floating number lifecycle.
func (s *FloatingDamageNumberSystem) Update(entities []*Entity, deltaTime float64) {
	s.cleanupTimer += deltaTime

	for _, entity := range entities {
		health := entity.GetHealth()
		if health == nil {
			continue
		}

		currentHP := health.Current
		entityID := entity.ID

		// Check for health change
		if prevHP, exists := s.prevHealth[entityID]; exists {
			diff := currentHP - prevHP
			if diff != 0 {
				s.spawnNumber(entity, diff)
			}
		}
		s.prevHealth[entityID] = currentHP

		// Update existing floating numbers
		s.updateNumbers(entity, deltaTime)
	}

	// Periodic cleanup of stale health tracking entries
	if s.cleanupTimer >= s.cleanupInterval {
		s.cleanupTimer = 0
		s.cleanupStaleEntries(entities)
	}
}

// spawnNumber creates a new floating number for the given health change.
func (s *FloatingDamageNumberSystem) spawnNumber(entity *Entity, healthDiff float64) {
	comp := s.getOrCreateComponent(entity)
	if comp == nil {
		return
	}

	// Evict oldest if at capacity
	if len(comp.Numbers) >= comp.MaxCount {
		comp.Numbers = comp.Numbers[1:]
	}

	isHeal := healthDiff > 0
	amount := math.Abs(healthDiff)

	// Crits are damage ≥ 20% of max health
	health := entity.GetHealth()
	isCrit := false
	if !isHeal && health != nil && health.Max > 0 {
		isCrit = amount >= health.Max*0.2
	}

	// Deterministic horizontal spread
	s.spreadCounter++
	spreadFrac := float64(s.spreadCounter%7)/6.0 - 0.5 // range -0.5 to 0.5
	offsetX := spreadFrac * s.preset.SpreadX * 2.0

	numColor := s.preset.DamageColor
	scale := 1.0
	if isHeal {
		numColor = s.preset.HealColor
	} else if isCrit {
		numColor = s.preset.CritColor
		scale = s.preset.CritScale
	}

	fn := FloatingNumber{
		Amount:   amount,
		OffsetX:  offsetX,
		OffsetY:  0,
		Opacity:  1.0,
		Age:      0,
		Color:    numColor,
		Scale:    scale,
		IsHeal:   isHeal,
		IsCrit:   isCrit,
		Duration: s.preset.Duration,
	}

	comp.Numbers = append(comp.Numbers, fn)
}

// updateNumbers advances the floating number lifecycle for an entity.
func (s *FloatingDamageNumberSystem) updateNumbers(entity *Entity, deltaTime float64) {
	fdnComp, exists := entity.GetComponent("floating_damage_number")
	if !exists {
		return
	}
	comp, ok := fdnComp.(*FloatingDamageNumberComponent)
	if !ok {
		return
	}

	if len(comp.Numbers) == 0 {
		return
	}

	// Update in-place, remove expired
	alive := 0
	for i := range comp.Numbers {
		fn := &comp.Numbers[i]
		fn.Age += deltaTime
		if fn.Age >= fn.Duration {
			continue
		}

		progress := fn.Age / fn.Duration
		fn.OffsetY = -s.preset.RiseSpeed * fn.Age
		// Smooth fade: full opacity for first 40%, then linear fade
		if progress < 0.4 {
			fn.Opacity = 1.0
		} else {
			fn.Opacity = 1.0 - (progress-0.4)/0.6
		}

		comp.Numbers[alive] = comp.Numbers[i]
		alive++
	}
	comp.Numbers = comp.Numbers[:alive]
}

// getOrCreateComponent retrieves or creates the FloatingDamageNumberComponent.
func (s *FloatingDamageNumberSystem) getOrCreateComponent(entity *Entity) *FloatingDamageNumberComponent {
	existing, exists := entity.GetComponent("floating_damage_number")
	if exists {
		if comp, ok := existing.(*FloatingDamageNumberComponent); ok {
			return comp
		}
	}
	comp := NewFloatingDamageNumberComponent()
	entity.AddComponent(comp)
	return comp
}

// cleanupStaleEntries removes health tracking for entities no longer in the update set.
func (s *FloatingDamageNumberSystem) cleanupStaleEntries(entities []*Entity) {
	active := make(map[uint64]struct{}, len(entities))
	for _, e := range entities {
		active[e.ID] = struct{}{}
	}
	for id := range s.prevHealth {
		if _, exists := active[id]; !exists {
			delete(s.prevHealth, id)
		}
	}
}
