// Package engine provides the DodgeAfterimageSystem which spawns translucent
// ghost copies of entity sprites at recent positions when entities move above a
// speed threshold (dodging, dashing, sprinting). Ghosts fade over time with
// genre-aware tint colors. This creates a motion trail effect distinct from the
// particle-based SprintTrailParticleSystem.
package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// afterimageGenrePreset holds genre-specific tint and decay configuration.
type afterimageGenrePreset struct {
	R, G, B      float64
	DecayRate    float64
	MaxGhosts    int
	SpawnInterval float64
}

// DodgeAfterimageSystem manages afterimage ghost creation and decay for
// fast-moving entities. It lazily attaches AfterimageComponent to entities
// that have both position and velocity, then updates ghost lists each frame.
type DodgeAfterimageSystem struct {
	world   *World
	logger  *logrus.Entry
	rng     *rand.Rand
	genreID string
	preset  afterimageGenrePreset

	// Throttle attachment scan for new entities
	scanInterval   float64
	timeSinceScan  float64
}

// NewDodgeAfterimageSystem creates a dodge afterimage system.
func NewDodgeAfterimageSystem(world *World, seed int64) *DodgeAfterimageSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "dodge_afterimage",
	})

	sys := &DodgeAfterimageSystem{
		world:        world,
		logger:       logger,
		rng:          rand.New(rand.NewSource(seed)),
		genreID:      "fantasy",
		scanInterval: 1.0,
	}
	sys.preset = sys.getPreset("fantasy")
	logger.Debug("dodge afterimage system created")
	return sys
}

// SetGenre configures genre-aware afterimage visuals.
func (s *DodgeAfterimageSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getPreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for afterimage")
	}
}

// getPreset returns genre-specific afterimage tint and behavior.
func (s *DodgeAfterimageSystem) getPreset(genreID string) afterimageGenrePreset {
	switch genreID {
	case "horror":
		// Dark reddish shadow ghosts, fast decay
		return afterimageGenrePreset{R: 0.4, G: 0.1, B: 0.1, DecayRate: 4.0, MaxGhosts: 4, SpawnInterval: 0.07}
	case "cyberpunk":
		// Neon cyan ghosts, slow decay for longer trails
		return afterimageGenrePreset{R: 0.2, G: 1.0, B: 0.9, DecayRate: 2.0, MaxGhosts: 6, SpawnInterval: 0.05}
	case "sci-fi", "scifi":
		// Holographic blue ghosts
		return afterimageGenrePreset{R: 0.3, G: 0.6, B: 1.0, DecayRate: 2.5, MaxGhosts: 5, SpawnInterval: 0.06}
	case "post-apocalyptic", "postapoc":
		// Dusty amber ghosts
		return afterimageGenrePreset{R: 0.8, G: 0.6, B: 0.3, DecayRate: 3.5, MaxGhosts: 4, SpawnInterval: 0.07}
	default: // fantasy
		// Golden ethereal ghosts
		return afterimageGenrePreset{R: 1.0, G: 0.85, B: 0.4, DecayRate: 3.0, MaxGhosts: 5, SpawnInterval: 0.06}
	}
}

// Update processes entities with position+velocity, spawning afterimage ghosts
// for fast-moving entities and decaying existing ghosts.
func (s *DodgeAfterimageSystem) Update(entities []*Entity, deltaTime float64) {
	if deltaTime <= 0 || deltaTime > 0.5 {
		return
	}

	s.timeSinceScan += deltaTime

	for _, entity := range entities {
		pos := entity.GetPosition()
		if pos == nil {
			continue
		}
		vel := entity.GetVelocity()
		if vel == nil {
			continue
		}

		speedSq := vel.VX*vel.VX + vel.VY*vel.VY

		comp, hasComp := entity.GetComponent("afterimage")
		if !hasComp {
			// Only attach during scan intervals and when entity is moving fast
			if s.timeSinceScan < s.scanInterval {
				continue
			}
			if speedSq < 14400.0 { // 120 px/s minimum to even bother attaching
				continue
			}
			ai := NewAfterimageComponent()
			ai.TintR = s.preset.R
			ai.TintG = s.preset.G
			ai.TintB = s.preset.B
			ai.Decay = s.preset.DecayRate
			ai.MaxGhost = s.preset.MaxGhosts
			ai.Interval = s.preset.SpawnInterval
			entity.AddComponent(ai)
			continue
		}

		ai, ok := comp.(*AfterimageComponent)
		if !ok {
			continue
		}

		// Decay existing ghosts
		alive := ai.Ghosts[:0]
		for i := range ai.Ghosts {
			ai.Ghosts[i].Opacity -= ai.Decay * deltaTime
			if ai.Ghosts[i].Opacity > 0.01 {
				alive = append(alive, ai.Ghosts[i])
			}
		}
		ai.Ghosts = alive

		// Spawn new ghost if moving fast enough and cooldown elapsed
		ai.TimeSinceSpawn += deltaTime
		if speedSq >= ai.SpeedThresholdSq && ai.TimeSinceSpawn >= ai.Interval {
			ai.TimeSinceSpawn = 0.0
			ghost := AfterimageGhost{
				X:       pos.X,
				Y:       pos.Y,
				Opacity: 0.6,
			}
			if len(ai.Ghosts) >= ai.MaxGhost {
				// Evict oldest (index 0)
				copy(ai.Ghosts, ai.Ghosts[1:])
				ai.Ghosts[len(ai.Ghosts)-1] = ghost
			} else {
				ai.Ghosts = append(ai.Ghosts, ghost)
			}
		}
	}

	if s.timeSinceScan >= s.scanInterval {
		s.timeSinceScan = 0.0
	}
}
