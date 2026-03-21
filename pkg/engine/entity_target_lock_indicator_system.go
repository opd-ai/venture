// Package engine provides the EntityTargetLockIndicatorSystem which renders
// genre-aware targeting reticle visuals around the nearest hostile entity to
// each player. The system automatically acquires the closest hostile within
// acquisition range, draws a rotating reticle ring whose style varies by genre
// (fantasy = golden arcane circle, horror = pulsing blood ring, sci-fi =
// holographic diamond, cyberpunk = neon crosshair, post-apocalyptic = rusty
// brackets). Reticle opacity and rotation speed scale with distance to target.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// TargetLockIndicatorComponent stores the visual parameters for a targeting
// reticle drawn around an entity that is being targeted by a player.
type TargetLockIndicatorComponent struct {
	// Reticle color (RGB, 0.0-1.0) from genre palette
	ReticleR float64
	ReticleG float64
	ReticleB float64

	// Base opacity before distance modulation (0.0-1.0)
	BaseOpacity float64

	// Current opacity after modulation (0.0-1.0)
	CurrentOpacity float64

	// Rotation angle in radians (reticle spins slowly)
	RotationAngle float64

	// Rotation speed in radians per second
	RotationSpeed float64

	// Reticle radius in pixels
	ReticleRadius float64

	// Pulse speed for lock-on emphasis (Hz)
	PulseSpeed float64

	// Accumulated pulse phase in radians
	PulsePhase float64

	// ID of the player entity targeting this entity
	LockedByPlayerID uint64

	// Whether this indicator is active
	Enabled bool
}

// Type returns the component type identifier.
func (c *TargetLockIndicatorComponent) Type() string {
	return "target_lock_indicator"
}

// NewTargetLockIndicatorComponent creates a disabled default component.
func NewTargetLockIndicatorComponent() *TargetLockIndicatorComponent {
	return &TargetLockIndicatorComponent{
		BaseOpacity:   0.0,
		ReticleRadius: 12.0,
		RotationSpeed: 1.0,
		Enabled:       false,
	}
}

// genreReticlePalette holds genre-specific reticle visual parameters.
type genreReticlePalette struct {
	R, G, B       float64 // Primary reticle color
	Opacity       float64 // Base opacity
	RotationSpeed float64 // Radians per second
	PulseSpeed    float64 // Hz
}

// EntityTargetLockIndicatorSystem finds the closest hostile to each player
// and assigns a rotating reticle visual to that entity. Target acquisition
// runs at 4 Hz; animation (rotation + pulse) runs every frame.
type EntityTargetLockIndicatorSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID     string
	lastGenreID string
	initialized bool

	acquisitionInterval float64
	timeSinceAcquire    float64

	// Maximum distance for target acquisition in pixels
	acquireRange float64

	palettes map[string]genreReticlePalette
}

// NewEntityTargetLockIndicatorSystem creates a new target lock indicator system.
func NewEntityTargetLockIndicatorSystem(world *World, seed int64) *EntityTargetLockIndicatorSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "entity_target_lock_indicator")
	}

	sys := &EntityTargetLockIndicatorSystem{
		world:               world,
		logger:              logEntry,
		rng:                 rand.New(rand.NewSource(seed)),
		genreID:             "fantasy",
		acquisitionInterval: 0.25,
		acquireRange:        160.0,
		palettes:            buildReticlePalettes(),
	}

	if logEntry != nil && logEntry.Logger.GetLevel() >= logrus.DebugLevel {
		logEntry.Debug("entity target lock indicator system created")
	}
	return sys
}

// SetGenre configures the active genre for reticle visuals.
func (s *EntityTargetLockIndicatorSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update drives reticle animation every frame and re-acquires targets periodically.
func (s *EntityTargetLockIndicatorSystem) Update(entities []*Entity, deltaTime float64) {
	if len(entities) == 0 {
		return
	}

	// Re-acquire targets periodically
	s.timeSinceAcquire += deltaTime
	reacquire := s.timeSinceAcquire >= s.acquisitionInterval
	if reacquire {
		s.timeSinceAcquire = 0
	}

	// Detect genre change
	genreChanged := s.genreID != s.lastGenreID
	if genreChanged {
		s.lastGenreID = s.genreID
	}

	// Collect players and hostiles
	if reacquire || !s.initialized {
		s.acquireTargets(entities, genreChanged)
		s.initialized = true
	}

	// Animate all active reticles
	s.animateReticles(entities, deltaTime)
}

// acquireTargets finds the closest hostile to each player and assigns reticle components.
func (s *EntityTargetLockIndicatorSystem) acquireTargets(entities []*Entity, genreChanged bool) {
	players, hostiles := s.partitionEntities(entities)
	s.clearStaleIndicators(entities, players)
	s.assignTargetsToPlayers(players, hostiles, genreChanged)
}

// partitionEntities separates entities into players and hostiles.
func (s *EntityTargetLockIndicatorSystem) partitionEntities(entities []*Entity) ([]*Entity, []*Entity) {
	var players, hostiles []*Entity
	for _, e := range entities {
		if e == nil {
			continue
		}
		if _, ok := e.GetComponent("input"); ok {
			players = append(players, e)
		} else if _, ok := e.GetComponent("ai"); ok {
			hostiles = append(hostiles, e)
		}
	}
	return players, hostiles
}

// clearStaleIndicators disables indicators for entities no longer in range.
func (s *EntityTargetLockIndicatorSystem) clearStaleIndicators(entities, players []*Entity) {
	for _, e := range entities {
		if e == nil {
			continue
		}
		tlc := s.getIndicatorComponent(e)
		if tlc == nil || !tlc.Enabled {
			continue
		}
		if !s.isTargetStillValid(tlc, players, e) {
			tlc.Enabled = false
			tlc.CurrentOpacity = 0
		}
	}
}

// getIndicatorComponent retrieves the target lock indicator component if present.
func (s *EntityTargetLockIndicatorSystem) getIndicatorComponent(e *Entity) *TargetLockIndicatorComponent {
	comp, ok := e.GetComponent("target_lock_indicator")
	if !ok {
		return nil
	}
	tlc, ok := comp.(*TargetLockIndicatorComponent)
	if !ok {
		return nil
	}
	return tlc
}

// isTargetStillValid checks if the locking player still exists and is in range.
func (s *EntityTargetLockIndicatorSystem) isTargetStillValid(tlc *TargetLockIndicatorComponent, players []*Entity, target *Entity) bool {
	for _, p := range players {
		if p.ID == tlc.LockedByPlayerID {
			return s.entityDistance(p, target) <= s.acquireRange*1.2
		}
	}
	return false
}

// assignTargetsToPlayers finds and assigns the closest hostile for each player.
func (s *EntityTargetLockIndicatorSystem) assignTargetsToPlayers(players, hostiles []*Entity, genreChanged bool) {
	palette := s.getCurrentPalette()
	for _, player := range players {
		px, py := s.getPosition(player)
		if px == 0 && py == 0 {
			continue
		}
		bestTarget := s.findClosestHostile(px, py, hostiles)
		if bestTarget == nil {
			continue
		}
		s.configureTargetIndicator(bestTarget, player.ID, palette, genreChanged)
	}
}

// findClosestHostile returns the nearest living hostile within acquisition range.
func (s *EntityTargetLockIndicatorSystem) findClosestHostile(px, py float64, hostiles []*Entity) *Entity {
	var bestTarget *Entity
	bestDist := s.acquireRange + 1
	for _, hostile := range hostiles {
		hx, hy := s.getPosition(hostile)
		if hx == 0 && hy == 0 {
			continue
		}
		if !s.isHostileAlive(hostile) {
			continue
		}
		dist := math.Sqrt((px-hx)*(px-hx) + (py-hy)*(py-hy))
		if dist < bestDist {
			bestDist = dist
			bestTarget = hostile
		}
	}
	return bestTarget
}

// isHostileAlive checks if the hostile entity has health > 0.
func (s *EntityTargetLockIndicatorSystem) isHostileAlive(hostile *Entity) bool {
	comp, ok := hostile.GetComponent("health")
	if !ok {
		return true
	}
	hc, ok := comp.(*HealthComponent)
	return !ok || hc.Current > 0
}

// configureTargetIndicator sets up the indicator component with palette and positioning.
func (s *EntityTargetLockIndicatorSystem) configureTargetIndicator(target *Entity, playerID uint64, palette genreReticlePalette, genreChanged bool) {
	tlc := s.getOrCreateComponent(target)
	tlc.Enabled = true
	tlc.LockedByPlayerID = playerID
	tlc.ReticleR = palette.R
	tlc.ReticleG = palette.G
	tlc.ReticleB = palette.B
	tlc.BaseOpacity = palette.Opacity
	tlc.RotationSpeed = palette.RotationSpeed
	tlc.PulseSpeed = palette.PulseSpeed
	tlc.ReticleRadius = s.computeReticleRadius(target)
	if genreChanged {
		tlc.PulsePhase = 0
		tlc.RotationAngle = 0
	}
}

// animateReticles updates rotation and pulse for all active reticle components.
func (s *EntityTargetLockIndicatorSystem) animateReticles(entities []*Entity, deltaTime float64) {
	for _, e := range entities {
		if e == nil {
			continue
		}
		comp, ok := e.GetComponent("target_lock_indicator")
		if !ok {
			continue
		}
		tlc, ok := comp.(*TargetLockIndicatorComponent)
		if !ok || !tlc.Enabled {
			continue
		}

		// Rotate the reticle
		tlc.RotationAngle += tlc.RotationSpeed * deltaTime
		if tlc.RotationAngle > 2*math.Pi {
			tlc.RotationAngle -= 2 * math.Pi
		}

		// Pulse opacity
		tlc.PulsePhase += tlc.PulseSpeed * 2 * math.Pi * deltaTime
		if tlc.PulsePhase > 2*math.Pi {
			tlc.PulsePhase -= 2 * math.Pi
		}
		pulseFactor := 0.5 + 0.5*math.Sin(tlc.PulsePhase)
		tlc.CurrentOpacity = tlc.BaseOpacity * (0.7 + 0.3*pulseFactor)
	}
}

// getPosition extracts X,Y from a PositionComponent.
func (s *EntityTargetLockIndicatorSystem) getPosition(entity *Entity) (float64, float64) {
	comp, ok := entity.GetComponent("position")
	if !ok {
		return 0, 0
	}
	pos, ok := comp.(*PositionComponent)
	if !ok {
		return 0, 0
	}
	return pos.X, pos.Y
}

// entityDistance computes Euclidean distance between two entities.
func (s *EntityTargetLockIndicatorSystem) entityDistance(a, b *Entity) float64 {
	ax, ay := s.getPosition(a)
	bx, by := s.getPosition(b)
	dx := ax - bx
	dy := ay - by
	return math.Sqrt(dx*dx + dy*dy)
}

// computeReticleRadius returns a radius scaled by entity size.
func (s *EntityTargetLockIndicatorSystem) computeReticleRadius(entity *Entity) float64 {
	baseRadius := 12.0
	if comp, ok := entity.GetComponent("collider"); ok {
		if col, ok := comp.(*ColliderComponent); ok {
			size := math.Max(col.Width, col.Height)
			if size > 0 {
				baseRadius = size*0.6 + 4.0
			}
		}
	}
	if baseRadius < 8.0 {
		baseRadius = 8.0
	}
	if baseRadius > 32.0 {
		baseRadius = 32.0
	}
	return baseRadius
}

// getOrCreateComponent retrieves or lazily creates the target lock component.
func (s *EntityTargetLockIndicatorSystem) getOrCreateComponent(entity *Entity) *TargetLockIndicatorComponent {
	comp, ok := entity.GetComponent("target_lock_indicator")
	if ok {
		if tlc, ok := comp.(*TargetLockIndicatorComponent); ok {
			return tlc
		}
	}
	tlc := NewTargetLockIndicatorComponent()
	entity.AddComponent(tlc)
	return tlc
}

// getCurrentPalette returns the genre-specific reticle palette.
func (s *EntityTargetLockIndicatorSystem) getCurrentPalette() genreReticlePalette {
	if p, ok := s.palettes[s.genreID]; ok {
		return p
	}
	return s.palettes["fantasy"]
}

// buildReticlePalettes returns genre-specific reticle visual presets.
func buildReticlePalettes() map[string]genreReticlePalette {
	return map[string]genreReticlePalette{
		"fantasy":   {R: 0.95, G: 0.80, B: 0.25, Opacity: 0.75, RotationSpeed: 0.8, PulseSpeed: 0.6},
		"horror":    {R: 0.90, G: 0.15, B: 0.10, Opacity: 0.70, RotationSpeed: 0.4, PulseSpeed: 0.3},
		"scifi":     {R: 0.25, G: 0.90, B: 0.95, Opacity: 0.80, RotationSpeed: 1.2, PulseSpeed: 0.8},
		"cyberpunk": {R: 0.95, G: 0.20, B: 0.90, Opacity: 0.85, RotationSpeed: 1.5, PulseSpeed: 1.0},
		"postapoc":  {R: 0.80, G: 0.55, B: 0.20, Opacity: 0.65, RotationSpeed: 0.6, PulseSpeed: 0.4},
	}
}
