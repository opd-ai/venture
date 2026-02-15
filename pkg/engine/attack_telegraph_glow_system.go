package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// genreTelegraphPreset holds genre-specific telegraph glow color and intensity scaling.
type genreTelegraphPreset struct {
	R, G, B        float64
	IntensityScale float64
}

// AttackTelegraphGlowSystem watches AI entities approaching attack readiness and
// writes a ramping glow intensity into AttackTelegraphComponent. This gives
// players a visual warning that a hostile entity is about to strike.
//
// The glow activates during the last 40% of the attack cooldown when the AI is
// in Chase or Attack state, and intensity follows a quadratic ease-in curve for
// a natural wind-up feel. Genre presets control glow color.
type AttackTelegraphGlowSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string
	preset  genreTelegraphPreset

	// Throttle full-scan for new entities
	updateInterval float64
	timeSinceCheck float64

	// Cooldown fraction at which telegraph begins (0.0-1.0 of remaining/total)
	telegraphThreshold float64
}

// NewAttackTelegraphGlowSystem creates a new attack telegraph glow system.
func NewAttackTelegraphGlowSystem(world *World, seed int64) *AttackTelegraphGlowSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "attack_telegraph_glow")
	}

	sys := &AttackTelegraphGlowSystem{
		world:              world,
		logger:             logEntry,
		rng:                rand.New(rand.NewSource(seed)),
		genreID:            "fantasy",
		updateInterval:     0.5,
		telegraphThreshold: 0.40,
	}
	sys.preset = sys.getPreset(sys.genreID)

	if logEntry != nil {
		logEntry.Debug("AttackTelegraphGlowSystem created")
	}
	return sys
}

// SetGenre configures genre-aware telegraph glow color and intensity.
func (s *AttackTelegraphGlowSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getPreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for attack telegraph glow")
	}
}

// getPreset returns genre-specific telegraph glow color and intensity scaling.
func (s *AttackTelegraphGlowSystem) getPreset(genreID string) genreTelegraphPreset {
	switch genreID {
	case "horror":
		// Deep crimson, high intensity — horror enemies telegraph menacingly
		return genreTelegraphPreset{R: 0.85, G: 0.05, B: 0.05, IntensityScale: 1.3}
	case "cyberpunk":
		// Neon magenta, moderate intensity
		return genreTelegraphPreset{R: 0.9, G: 0.1, B: 0.7, IntensityScale: 1.1}
	case "sci-fi", "scifi":
		// Electric cyan, lower intensity — clean sci-fi aesthetic
		return genreTelegraphPreset{R: 0.1, G: 0.7, B: 0.9, IntensityScale: 0.9}
	case "post-apocalyptic", "postapoc":
		// Amber/rust warning — gritty, industrial danger
		return genreTelegraphPreset{R: 0.9, G: 0.5, B: 0.1, IntensityScale: 1.2}
	case "fantasy":
		// Warm orange-red — classic fantasy danger glow
		return genreTelegraphPreset{R: 0.9, G: 0.3, B: 0.1, IntensityScale: 1.0}
	default:
		return genreTelegraphPreset{R: 0.9, G: 0.2, B: 0.1, IntensityScale: 1.0}
	}
}

// Update processes entities with AI and attack components, ramping telegraph
// glow as cooldown approaches ready.
func (s *AttackTelegraphGlowSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	fullScan := false
	if s.timeSinceCheck >= s.updateInterval {
		s.timeSinceCheck = 0
		fullScan = true
	}

	for _, entity := range entities {
		// Only process AI-controlled entities (not player)
		if entity.HasComponent("input") {
			continue
		}

		aiComp, _ := entity.GetComponent("ai")
		ai, hasAI := aiComp.(*AIComponent)
		if !hasAI {
			continue
		}

		attackComp, _ := entity.GetComponent("attack")
		attack, hasAttack := attackComp.(*AttackComponent)
		if !hasAttack {
			continue
		}

		comp, _ := entity.GetComponent("attack_telegraph")
		telegraph, hasTelegraph := comp.(*AttackTelegraphComponent)

		if !hasTelegraph {
			if !fullScan {
				continue
			}
			telegraph = NewAttackTelegraphComponent()
			entity.AddComponent(telegraph)
		}

		s.updateTelegraph(entity, ai, attack, telegraph)
	}
}

// updateTelegraph computes telegraph glow intensity from AI state and cooldown progress.
func (s *AttackTelegraphGlowSystem) updateTelegraph(
	entity *Entity,
	ai *AIComponent,
	attack *AttackComponent,
	telegraph *AttackTelegraphComponent,
) {
	// Telegraph only active when AI is in aggressive states
	if ai.State != AIStateAttack && ai.State != AIStateChase {
		telegraph.Active = false
		telegraph.Intensity = 0
		return
	}

	// Need a valid cooldown to telegraph
	if attack.Cooldown <= 0 {
		telegraph.Active = false
		telegraph.Intensity = 0
		return
	}

	// Fraction of cooldown remaining (1.0 = just attacked, 0.0 = ready)
	remainingFraction := attack.CooldownTimer / attack.Cooldown
	if remainingFraction < 0 {
		remainingFraction = 0
	}
	if remainingFraction > 1 {
		remainingFraction = 1
	}

	// Telegraph activates in the last telegraphThreshold portion of cooldown
	if remainingFraction > s.telegraphThreshold {
		telegraph.Active = false
		telegraph.Intensity = 0
		return
	}

	// Map remaining fraction within threshold to 0.0-1.0 intensity
	// When remainingFraction=threshold → raw=0, when remainingFraction=0 → raw=1
	raw := 1.0 - (remainingFraction / s.telegraphThreshold)

	// Quadratic ease-in for natural wind-up feel
	intensity := raw * raw * s.preset.IntensityScale
	if intensity > 1.0 {
		intensity = 1.0
	}

	telegraph.Active = true
	telegraph.Intensity = intensity
	telegraph.ColorR = s.preset.R
	telegraph.ColorG = s.preset.G
	telegraph.ColorB = s.preset.B

	// Scale radius from entity size
	radius := 16.0
	if col := entity.GetCollider(); col != nil {
		radius = math.Max(col.Width, col.Height) * 0.6
	} else if spr := entity.GetSprite(); spr != nil {
		radius = math.Max(spr.Width, spr.Height) * 0.6
	}
	if radius < 10 {
		radius = 10
	}
	telegraph.Radius = radius
}
