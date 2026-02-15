// Package engine provides the CombatHitStaggerSystem which applies genre-aware
// positional offsets to entity sprites when they take damage. It monitors
// HealthComponent for decreases and writes decaying X/Y offsets into
// HitStaggerComponent, connecting combat damage to spatial visual feedback.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// HitStaggerComponent stores per-entity stagger offset state.
// Pure data — all logic lives in CombatHitStaggerSystem.
type HitStaggerComponent struct {
	// OffsetX is the horizontal pixel offset applied during rendering.
	OffsetX float64
	// OffsetY is the vertical pixel offset applied during rendering.
	OffsetY float64
	// Active indicates a stagger is in progress.
	Active bool
	// Timer tracks remaining stagger time in seconds.
	Timer float64
	// Duration is the total stagger duration for the current hit.
	Duration float64
	// InitialOffsetX stores the peak offset for decay interpolation.
	InitialOffsetX float64
	// InitialOffsetY stores the peak offset for decay interpolation.
	InitialOffsetY float64
}

// Type returns the component type identifier.
func (h *HitStaggerComponent) Type() string { return "hit_stagger" }

// hitStaggerPreset holds genre-specific stagger configuration.
type hitStaggerPreset struct {
	MaxOffset    float64 // Max displacement in pixels (2–6)
	BaseDuration float64 // Base stagger duration in seconds
	DecayPower   float64 // Decay exponent (1=linear, 2=quadratic, higher=snappier)
	MinDamagePct float64 // Minimum damage proportion to trigger stagger
}

// CombatHitStaggerSystem monitors entity health changes and triggers visual
// positional offsets on HitStaggerComponent when damage is taken. Offset
// magnitude scales with damage proportion and decays over time.
type CombatHitStaggerSystem struct {
	world  *World
	seed   int64
	rng    *rand.Rand
	logger *logrus.Entry

	// Previous health values keyed by entity ID
	prevHealth map[uint64]float64

	genreID string
	preset  hitStaggerPreset

	// Throttle cleanup of stale entries
	cleanupTimer    float64
	cleanupInterval float64
}

// NewCombatHitStaggerSystem creates a new combat hit stagger system.
func NewCombatHitStaggerSystem(world *World, seed int64) *CombatHitStaggerSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "combat_hit_stagger",
	})

	sys := &CombatHitStaggerSystem{
		world:           world,
		seed:            seed,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logger,
		prevHealth:      make(map[uint64]float64, 128),
		genreID:         "fantasy",
		cleanupInterval: 5.0,
	}

	sys.preset = sys.getGenrePreset("fantasy")
	logger.Debug("combat hit stagger system created")
	return sys
}

// SetGenre configures genre-specific stagger parameters.
func (s *CombatHitStaggerSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getGenrePreset(genreID)
}

// getGenrePreset returns stagger configuration for the given genre.
func (s *CombatHitStaggerSystem) getGenrePreset(genreID string) hitStaggerPreset {
	switch genreID {
	case "horror":
		// Violent jerk, slow recovery for visceral feel
		return hitStaggerPreset{MaxOffset: 5.0, BaseDuration: 0.25, DecayPower: 1.5, MinDamagePct: 0.02}
	case "cyberpunk":
		// Glitchy sharp displacement, fast snap-back
		return hitStaggerPreset{MaxOffset: 4.0, BaseDuration: 0.12, DecayPower: 3.0, MinDamagePct: 0.03}
	case "scifi":
		// Energy knockback, quick elastic recovery
		return hitStaggerPreset{MaxOffset: 3.5, BaseDuration: 0.15, DecayPower: 2.5, MinDamagePct: 0.03}
	case "postapoc":
		// Heavy, gritty stagger with slow decay
		return hitStaggerPreset{MaxOffset: 5.5, BaseDuration: 0.22, DecayPower: 1.2, MinDamagePct: 0.02}
	default: // fantasy
		// Moderate dramatic knockback, smooth recovery
		return hitStaggerPreset{MaxOffset: 4.0, BaseDuration: 0.18, DecayPower: 2.0, MinDamagePct: 0.03}
	}
}

// Update checks all entities for health decreases and triggers/updates stagger offsets.
func (s *CombatHitStaggerSystem) Update(entities []*Entity, deltaTime float64) {
	s.cleanupTimer += deltaTime

	for _, entity := range entities {
		health := entity.GetHealth()
		if health == nil {
			continue
		}

		stagger := s.getOrCreateStagger(entity)
		current := health.Current
		prev, tracked := s.prevHealth[entity.ID]

		if tracked && current < prev && health.Max > 0 {
			damage := prev - current
			proportion := damage / health.Max
			if proportion >= s.preset.MinDamagePct {
				s.triggerStagger(stagger, proportion, entity.ID)
			}
		}

		s.prevHealth[entity.ID] = current

		// Decay active stagger
		if stagger.Active {
			s.updateDecay(stagger, deltaTime)
		}
	}

	if s.cleanupTimer >= s.cleanupInterval {
		s.cleanupTimer = 0
		s.cleanupStaleEntries(entities)
	}
}

// getOrCreateStagger retrieves or lazily creates a HitStaggerComponent on the entity.
func (s *CombatHitStaggerSystem) getOrCreateStagger(entity *Entity) *HitStaggerComponent {
	if comp, ok := entity.GetComponent("hit_stagger"); ok {
		return comp.(*HitStaggerComponent)
	}
	stagger := &HitStaggerComponent{}
	entity.AddComponent(stagger)
	return stagger
}

// triggerStagger initiates a stagger with direction seeded from entity ID.
func (s *CombatHitStaggerSystem) triggerStagger(stagger *HitStaggerComponent, proportion float64, entityID uint64) {
	if proportion > 1.0 {
		proportion = 1.0
	}

	// Scale offset magnitude: lerp between 30% and 100% of max based on damage
	magnitude := s.preset.MaxOffset * (0.3 + 0.7*proportion)

	// Deterministic direction derived from entity ID and a counter in the RNG
	angle := s.rng.Float64() * 2.0 * math.Pi

	stagger.InitialOffsetX = magnitude * math.Cos(angle)
	stagger.InitialOffsetY = magnitude * math.Sin(angle)
	stagger.OffsetX = stagger.InitialOffsetX
	stagger.OffsetY = stagger.InitialOffsetY
	stagger.Duration = s.preset.BaseDuration * (1.0 + 0.5*proportion)
	stagger.Timer = stagger.Duration
	stagger.Active = true
}

// updateDecay smoothly decays the stagger offset toward zero.
func (s *CombatHitStaggerSystem) updateDecay(stagger *HitStaggerComponent, deltaTime float64) {
	stagger.Timer -= deltaTime
	if stagger.Timer <= 0 {
		stagger.OffsetX = 0
		stagger.OffsetY = 0
		stagger.Active = false
		stagger.Timer = 0
		return
	}

	// Normalized remaining time [0,1] where 1 = just started, 0 = finished
	t := stagger.Timer / stagger.Duration
	// Apply power curve for genre-specific decay feel
	decay := math.Pow(t, s.preset.DecayPower)

	stagger.OffsetX = stagger.InitialOffsetX * decay
	stagger.OffsetY = stagger.InitialOffsetY * decay
}

// cleanupStaleEntries removes health tracking for entities no longer in the world.
func (s *CombatHitStaggerSystem) cleanupStaleEntries(entities []*Entity) {
	active := make(map[uint64]struct{}, len(entities))
	for _, e := range entities {
		active[e.ID] = struct{}{}
	}
	for id := range s.prevHealth {
		if _, ok := active[id]; !ok {
			delete(s.prevHealth, id)
		}
	}
}
