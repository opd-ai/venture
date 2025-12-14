package engine

import (
	"github.com/sirupsen/logrus"
)

// MusicLayer represents a music layer that can be mixed dynamically.
type MusicLayer int

const (
	LayerAmbient MusicLayer = iota
	LayerMelody
	LayerBass
	LayerPercussion
	LayerHarmony
	LayerDrone
)

// MusicIntensity represents the current intensity level of gameplay.
type MusicIntensity int

const (
	IntensityCalm MusicIntensity = iota
	IntensityLow
	IntensityMedium
	IntensityHigh
	IntensityCombat
)

// AdaptiveSoundtrackComponent controls dynamic music based on game state.
// Phase 29: Adaptive Soundtrack
type AdaptiveSoundtrackComponent struct {
	// CurrentIntensity tracks the current music intensity level
	CurrentIntensity MusicIntensity

	// TargetIntensity is the desired intensity based on game state
	TargetIntensity MusicIntensity

	// ActiveLayers are the currently playing music layers
	ActiveLayers map[MusicLayer]bool

	// LayerVolumes control individual layer volumes (0.0-1.0)
	LayerVolumes map[MusicLayer]float64

	// TransitionSpeed controls how fast music adapts (0.0-1.0)
	TransitionSpeed float64

	// CombatThreshold is the number of nearby enemies to trigger combat music
	CombatThreshold int

	// ExplorationBonus increases intensity during new area discovery
	ExplorationBonus float64

	// GenreTheme identifies the music genre (fantasy, sci-fi, etc.)
	GenreTheme string
}

// Type returns the component type identifier.
func (a *AdaptiveSoundtrackComponent) Type() string {
	return "adaptive_soundtrack"
}

// NewAdaptiveSoundtrackComponent creates an adaptive soundtrack component.
func NewAdaptiveSoundtrackComponent(genreTheme string) *AdaptiveSoundtrackComponent {
	return &AdaptiveSoundtrackComponent{
		CurrentIntensity: IntensityCalm,
		TargetIntensity:  IntensityCalm,
		ActiveLayers:     make(map[MusicLayer]bool),
		LayerVolumes:     make(map[MusicLayer]float64),
		TransitionSpeed:  0.1,
		CombatThreshold:  3,
		GenreTheme:       genreTheme,
	}
}

// AdaptiveSoundtrackSystem manages dynamic music adaptation.
type AdaptiveSoundtrackSystem struct {
	world  *World
	logger *logrus.Entry
}

// NewAdaptiveSoundtrackSystem creates a new adaptive soundtrack system.
func NewAdaptiveSoundtrackSystem(world *World) *AdaptiveSoundtrackSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system": "adaptive_soundtrack",
	})
	return &AdaptiveSoundtrackSystem{
		world:  world,
		logger: logger,
	}
}

// Update analyzes game state and adapts music accordingly.
func (s *AdaptiveSoundtrackSystem) Update(deltaTime float64) {
	// Get player entity - prefer entities with adaptive_soundtrack component
	var player *Entity
	var soundtrackComp *AdaptiveSoundtrackComponent

	// First, try to find entities that already have adaptive_soundtrack
	soundtrackEntities := s.world.GetEntitiesWith("adaptive_soundtrack")
	if len(soundtrackEntities) > 0 {
		player = soundtrackEntities[0]
		comp, ok := player.GetComponent("adaptive_soundtrack")
		if ok && comp != nil {
			soundtrackComp = comp.(*AdaptiveSoundtrackComponent)
		}
	}

	if soundtrackComp == nil {
		// Fall back to finding player by position+health
		players := s.world.GetEntitiesWith("position", "health")
		if len(players) == 0 {
			return
		}
		player = players[0]

		// Create default soundtrack component
		soundtrackComp = NewAdaptiveSoundtrackComponent("fantasy")
		player.AddComponent(soundtrackComp)
	}

	// Analyze game state to determine target intensity
	s.analyzeGameState(player, soundtrackComp)

	// Transition toward target intensity
	s.transitionIntensity(soundtrackComp, deltaTime)

	// Update active layers based on intensity
	s.updateLayers(soundtrackComp)

	// Apply layer volumes
	s.applyLayerVolumes(soundtrackComp, deltaTime)
}

// analyzeGameState determines the target music intensity.
func (s *AdaptiveSoundtrackSystem) analyzeGameState(player *Entity, soundtrack *AdaptiveSoundtrackComponent) {
	comp, hasPos := player.GetComponent("position")
	if !hasPos {
		return
	}
	playerPos := comp.(*PositionComponent)

	// Count nearby enemies
	enemies := s.countNearbyEnemies(playerPos, 400.0) // Within 400 pixel radius

	// Check player health
	healthComp, hasHealth := player.GetComponent("health")
	healthPercent := 1.0
	if hasHealth {
		health := healthComp.(*HealthComponent)
		healthPercent = health.Current / health.Max
	}

	// Determine intensity based on multiple factors
	if enemies >= soundtrack.CombatThreshold {
		// Combat situation
		soundtrack.TargetIntensity = IntensityCombat
	} else if enemies > 0 || healthPercent < 0.5 {
		// Danger or low health
		soundtrack.TargetIntensity = IntensityHigh
	} else if healthPercent < 0.75 {
		// Moderate tension
		soundtrack.TargetIntensity = IntensityMedium
	} else if soundtrack.ExplorationBonus > 0 {
		// Exploring new areas
		soundtrack.TargetIntensity = IntensityLow
		soundtrack.ExplorationBonus -= 0.01 // Decay exploration bonus
	} else {
		// Safe, calm
		soundtrack.TargetIntensity = IntensityCalm
	}
}

// countNearbyEnemies counts hostile entities within radius.
func (s *AdaptiveSoundtrackSystem) countNearbyEnemies(playerPos *PositionComponent, radius float64) int {
	// Get all entities with health (potential enemies)
	entities := s.world.GetEntitiesWith("position", "health")

	enemyCount := 0
	radiusSq := radius * radius

	for _, entity := range entities {
		// Skip entities without AI component (not enemies)
		aiComp, hasAI := entity.GetComponent("ai")
		if !hasAI {
			continue
		}
		_ = aiComp // Unused, just checking presence

		posComp, _ := entity.GetComponent("position")
		entityPos := posComp.(*PositionComponent)
		dx := entityPos.X - playerPos.X
		dy := entityPos.Y - playerPos.Y
		distSq := dx*dx + dy*dy

		if distSq <= radiusSq {
			enemyCount++
		}
	}

	return enemyCount
}

// transitionIntensity smoothly changes current intensity toward target.
func (s *AdaptiveSoundtrackSystem) transitionIntensity(soundtrack *AdaptiveSoundtrackComponent, deltaTime float64) {
	if soundtrack.CurrentIntensity == soundtrack.TargetIntensity {
		return
	}

	if soundtrack.CurrentIntensity < soundtrack.TargetIntensity {
		// Increase intensity immediately
		soundtrack.CurrentIntensity++
		if soundtrack.CurrentIntensity > soundtrack.TargetIntensity {
			soundtrack.CurrentIntensity = soundtrack.TargetIntensity
		}
	} else {
		// Decrease intensity (can also be immediate)
		soundtrack.CurrentIntensity--
		if soundtrack.CurrentIntensity < soundtrack.TargetIntensity {
			soundtrack.CurrentIntensity = soundtrack.TargetIntensity
		}
	}
}

// updateLayers activates/deactivates music layers based on intensity.
func (s *AdaptiveSoundtrackSystem) updateLayers(soundtrack *AdaptiveSoundtrackComponent) {
	// Reset all layers
	for layer := LayerAmbient; layer <= LayerDrone; layer++ {
		soundtrack.ActiveLayers[layer] = false
	}

	// Activate layers based on intensity
	switch soundtrack.CurrentIntensity {
	case IntensityCalm:
		// Minimal music - ambient only
		soundtrack.ActiveLayers[LayerAmbient] = true

	case IntensityLow:
		// Ambient + light melody
		soundtrack.ActiveLayers[LayerAmbient] = true
		soundtrack.ActiveLayers[LayerMelody] = true

	case IntensityMedium:
		// Add harmony
		soundtrack.ActiveLayers[LayerAmbient] = true
		soundtrack.ActiveLayers[LayerMelody] = true
		soundtrack.ActiveLayers[LayerHarmony] = true

	case IntensityHigh:
		// Add bass for tension
		soundtrack.ActiveLayers[LayerAmbient] = true
		soundtrack.ActiveLayers[LayerMelody] = true
		soundtrack.ActiveLayers[LayerHarmony] = true
		soundtrack.ActiveLayers[LayerBass] = true

	case IntensityCombat:
		// Full intensity - all layers
		soundtrack.ActiveLayers[LayerAmbient] = true
		soundtrack.ActiveLayers[LayerMelody] = true
		soundtrack.ActiveLayers[LayerHarmony] = true
		soundtrack.ActiveLayers[LayerBass] = true
		soundtrack.ActiveLayers[LayerPercussion] = true
		soundtrack.ActiveLayers[LayerDrone] = true
	}
}

// applyLayerVolumes fades layers in/out smoothly.
func (s *AdaptiveSoundtrackSystem) applyLayerVolumes(soundtrack *AdaptiveSoundtrackComponent, deltaTime float64) {
	fadeSpeed := 2.0 * deltaTime // Fade over ~0.5 seconds

	for layer := LayerAmbient; layer <= LayerDrone; layer++ {
		targetVolume := 0.0
		if soundtrack.ActiveLayers[layer] {
			targetVolume = 1.0
		}

		currentVolume := soundtrack.LayerVolumes[layer]

		// Fade toward target volume
		if currentVolume < targetVolume {
			currentVolume += fadeSpeed
			if currentVolume > targetVolume {
				currentVolume = targetVolume
			}
		} else if currentVolume > targetVolume {
			currentVolume -= fadeSpeed
			if currentVolume < targetVolume {
				currentVolume = targetVolume
			}
		}

		soundtrack.LayerVolumes[layer] = currentVolume
	}
}

// SetExplorationBonus increases music intensity when discovering new areas.
func (s *AdaptiveSoundtrackSystem) SetExplorationBonus(player *Entity, bonus float64) {
	comp, hasComp := player.GetComponent("adaptive_soundtrack")
	if hasComp {
		if soundtrack, ok := comp.(*AdaptiveSoundtrackComponent); ok {
			soundtrack.ExplorationBonus = bonus
		}
	}
}

// GetCurrentIntensity returns the current music intensity.
func (s *AdaptiveSoundtrackSystem) GetCurrentIntensity(player *Entity) MusicIntensity {
	comp, hasComp := player.GetComponent("adaptive_soundtrack")
	if hasComp {
		if soundtrack, ok := comp.(*AdaptiveSoundtrackComponent); ok {
			return soundtrack.CurrentIntensity
		}
	}
	return IntensityCalm
}
