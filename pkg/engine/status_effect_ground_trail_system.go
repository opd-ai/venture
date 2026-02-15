// Package engine provides the StatusEffectGroundTrailSystem for ground-level
// DoT trail effects. When entities move while afflicted by burning, poison,
// bleeding, or frozen status effects, this system drops genre-aware ground
// particles at their previous positions — ember smears, poison puddles,
// blood drips, or frost patches — giving players a visual breadcrumb trail
// of afflicted entity movement.
package engine

import (
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/particles"
	"github.com/sirupsen/logrus"
)

// StatusEffectGroundTrailComponent tracks per-entity position for movement
// delta detection and drop cooldown timers per effect type.
type StatusEffectGroundTrailComponent struct {
	PrevX       float64            // Last recorded X position
	PrevY       float64            // Last recorded Y position
	Initialized bool               // Whether PrevX/PrevY have been set
	DropTimers  map[string]float64 // Per-effect-type cooldown until next drop
}

// Type returns the component type identifier.
func (c *StatusEffectGroundTrailComponent) Type() string { return "status_effect_ground_trail" }

// groundTrailEffectPreset holds genre-specific visual parameters per DoT type.
type groundTrailEffectPreset struct {
	particleType particles.ParticleType
	minSize      float64
	maxSize      float64
	duration     float64
	gravity      float64
	baseCount    int
}

// groundTrailGenrePresets maps effect type → genre → preset.
type groundTrailGenrePresets map[string]map[string]groundTrailEffectPreset

func defaultGroundTrailPresets() groundTrailGenrePresets {
	return groundTrailGenrePresets{
		"burning": {
			"fantasy":   {particles.ParticleEmber, 2.0, 4.0, 1.2, 2.0, 3},
			"scifi":     {particles.ParticleSpark, 1.5, 3.5, 0.9, 1.5, 4},
			"horror":    {particles.ParticleFlame, 2.0, 5.0, 1.5, 1.0, 3},
			"cyberpunk": {particles.ParticleSpark, 1.5, 3.0, 0.8, 2.5, 5},
			"postapoc":  {particles.ParticleEmber, 2.5, 5.0, 1.4, 1.5, 3},
		},
		"poisoned": {
			"fantasy":   {particles.ParticleMagic, 2.0, 4.0, 1.8, 0.5, 2},
			"scifi":     {particles.ParticleMagic, 1.5, 3.5, 1.4, 0.3, 3},
			"horror":    {particles.ParticleSmoke, 2.5, 5.0, 2.0, 0.0, 2},
			"cyberpunk": {particles.ParticleMagic, 1.5, 3.0, 1.2, 0.5, 3},
			"postapoc":  {particles.ParticleDust, 2.5, 5.0, 2.2, 0.0, 2},
		},
		"bleeding": {
			"fantasy":   {particles.ParticleBlood, 2.0, 4.0, 2.0, 3.0, 2},
			"scifi":     {particles.ParticleBlood, 1.5, 3.0, 1.5, 4.0, 3},
			"horror":    {particles.ParticleBlood, 3.0, 6.0, 2.5, 2.0, 3},
			"cyberpunk": {particles.ParticleBlood, 1.5, 3.0, 1.5, 3.5, 2},
			"postapoc":  {particles.ParticleBlood, 2.5, 5.0, 2.2, 2.5, 2},
		},
		"frozen": {
			"fantasy":   {particles.ParticleSparkle, 2.0, 4.0, 2.5, 0.0, 3},
			"scifi":     {particles.ParticleSpark, 1.5, 3.0, 2.0, 0.0, 4},
			"horror":    {particles.ParticleSparkle, 2.0, 5.0, 3.0, 0.0, 2},
			"cyberpunk": {particles.ParticleSpark, 1.5, 3.0, 1.8, 0.0, 4},
			"postapoc":  {particles.ParticleDust, 2.5, 5.0, 2.5, 0.0, 2},
		},
	}
}

// StatusEffectGroundTrailSystem drops ground-level particles behind moving
// entities that are afflicted by damage-over-time status effects.
type StatusEffectGroundTrailSystem struct {
	world          *World
	particleSystem *ParticleSystem
	genreID        string
	seed           int64
	rng            *rand.Rand
	logger         *logrus.Entry
	presets        groundTrailGenrePresets

	// Tuning
	dropInterval     float64  // Minimum seconds between drops per effect type
	minMoveDist      float64  // Minimum pixel distance to trigger a drop
	spreadRadius     float64  // Particle spread around drop point
	trailEffectTypes []string // Which effect types produce ground trails
}

// NewStatusEffectGroundTrailSystem creates a new ground trail system.
func NewStatusEffectGroundTrailSystem(world *World, seed int64) *StatusEffectGroundTrailSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "status_effect_ground_trail")
		logEntry.Debug("status effect ground trail system created")
	}

	return &StatusEffectGroundTrailSystem{
		world:            world,
		seed:             seed,
		rng:              rand.New(rand.NewSource(seed)),
		logger:           logEntry,
		presets:          defaultGroundTrailPresets(),
		dropInterval:     0.35,
		minMoveDist:      6.0,
		spreadRadius:     10.0,
		trailEffectTypes: []string{"burning", "poisoned", "bleeding", "frozen"},
	}
}

// SetParticleSystem sets the particle system used for spawning effects.
func (s *StatusEffectGroundTrailSystem) SetParticleSystem(ps *ParticleSystem) {
	s.particleSystem = ps
}

// SetGenre sets the genre ID for genre-aware particle selection.
func (s *StatusEffectGroundTrailSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update iterates entities, detects movement while afflicted, and drops trails.
func (s *StatusEffectGroundTrailSystem) Update(entities []*Entity, deltaTime float64) {
	if s.particleSystem == nil || s.world == nil {
		return
	}

	genre := s.genreID
	if genre == "" {
		genre = "fantasy"
	}

	for _, entity := range entities {
		statusComp, hasStatus := entity.GetComponent("status_effect")
		if !hasStatus {
			continue
		}
		status := statusComp.(*StatusEffectComponent)

		// Only process trail-eligible effect types
		if !s.isTrailEffect(status.EffectType) || status.IsExpired() {
			continue
		}

		pos := entity.GetPosition()
		if pos == nil {
			continue
		}

		trail := s.ensureComponent(entity)

		// Tick per-effect drop timer
		if timer, ok := trail.DropTimers[status.EffectType]; ok && timer > 0 {
			trail.DropTimers[status.EffectType] = timer - deltaTime
		}

		if !trail.Initialized {
			trail.PrevX = pos.X
			trail.PrevY = pos.Y
			trail.Initialized = true
			continue
		}

		// Calculate movement distance
		dx := pos.X - trail.PrevX
		dy := pos.Y - trail.PrevY
		distSq := dx*dx + dy*dy

		if distSq >= s.minMoveDist*s.minMoveDist {
			timer := trail.DropTimers[status.EffectType]
			if timer <= 0 {
				s.spawnGroundTrail(trail.PrevX, trail.PrevY, entity.ID, status.EffectType, status.Magnitude, genre)
				trail.DropTimers[status.EffectType] = s.dropInterval
			}
			trail.PrevX = pos.X
			trail.PrevY = pos.Y
		}
	}
}

// isTrailEffect returns true if the effect type produces ground trails.
func (s *StatusEffectGroundTrailSystem) isTrailEffect(effectType string) bool {
	for _, t := range s.trailEffectTypes {
		if t == effectType {
			return true
		}
	}
	return false
}

// ensureComponent lazily attaches a StatusEffectGroundTrailComponent to the entity.
func (s *StatusEffectGroundTrailSystem) ensureComponent(entity *Entity) *StatusEffectGroundTrailComponent {
	if comp, ok := entity.GetComponent("status_effect_ground_trail"); ok {
		return comp.(*StatusEffectGroundTrailComponent)
	}
	comp := &StatusEffectGroundTrailComponent{
		DropTimers: make(map[string]float64),
	}
	entity.AddComponent(comp)
	return comp
}

// spawnGroundTrail creates genre-aware ground-level particles at the given position.
func (s *StatusEffectGroundTrailSystem) spawnGroundTrail(x, y float64, entityID uint64, effectType string, magnitude float64, genre string) {
	preset := s.getPreset(effectType, genre)

	// Scale count by magnitude (1x at mag≤1, up to 2x at mag≥5)
	intensityMul := 1.0
	if magnitude > 1.0 {
		intensityMul = 1.0 + (magnitude-1.0)*0.25
	}
	if intensityMul > 2.0 {
		intensityMul = 2.0
	}
	count := int(float64(preset.baseCount) * intensityMul)
	if count < 1 {
		count = 1
	}
	if count > 12 {
		count = 12
	}

	effectSeed := s.seed + int64(entityID)*37 + int64(x*11) + int64(y*17)

	config := particles.Config{
		Type:     preset.particleType,
		Count:    count,
		GenreID:  genre,
		Seed:     effectSeed,
		Duration: preset.duration,
		SpreadX:  s.spreadRadius,
		SpreadY:  s.spreadRadius * 0.5,
		Gravity:  preset.gravity,
		MinSize:  preset.minSize,
		MaxSize:  preset.maxSize,
		ZLayer:   particles.ZLayerGround,
		Custom:   map[string]interface{}{"ground_trail": effectType},
	}

	s.particleSystem.SpawnParticles(s.world, config, x, y)
}

// getPreset returns the genre-specific preset for an effect type, with fantasy fallback.
func (s *StatusEffectGroundTrailSystem) getPreset(effectType, genre string) groundTrailEffectPreset {
	if effects, ok := s.presets[effectType]; ok {
		if preset, ok := effects[genre]; ok {
			return preset
		}
		if preset, ok := effects["fantasy"]; ok {
			return preset
		}
	}
	// Ultimate fallback
	return groundTrailEffectPreset{
		particleType: particles.ParticleDust,
		minSize:      2.0, maxSize: 4.0,
		duration: 1.0, gravity: 1.0, baseCount: 2,
	}
}
