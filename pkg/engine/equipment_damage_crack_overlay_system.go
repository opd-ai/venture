// Package engine provides the EquipmentDamageCrackOverlaySystem which reads
// per-entity EquipmentWearTintComponent crack density and edge roughness values,
// then generates seeded procedural crack line segments in a genre-aware style.
// This bridges aggregate damage state data with renderable crack overlay geometry.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// crackGenreStyle holds genre-specific crack generation parameters.
type crackGenreStyle struct {
	// Branching angle range (radians) for crack propagation
	MinAngle, MaxAngle float64
	// How much cracks curve per step (0=straight, 1=very curved)
	Curvature float64
	// Branch probability per segment
	BranchChance float64
	// Crack color (RGB 0.0–1.0)
	R, G, B float64
	// Segment length range (normalized)
	MinLen, MaxLen float64
}

// EquipmentDamageCrackOverlaySystem generates procedural crack overlays from
// equipment wear data. Reads EquipmentWearTintComponent for crack density and
// edge roughness, writes EquipmentCrackOverlayComponent with crack segments.
type EquipmentDamageCrackOverlaySystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string
	style   crackGenreStyle

	// Throttle regeneration to 1 Hz
	updateInterval float64
	timeSinceCheck float64
}

// NewEquipmentDamageCrackOverlaySystem creates a new crack overlay system.
func NewEquipmentDamageCrackOverlaySystem(world *World, seed int64) *EquipmentDamageCrackOverlaySystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "equipment_damage_crack_overlay")
		logEntry.Debug("equipment damage crack overlay system created")
	}

	sys := &EquipmentDamageCrackOverlaySystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		updateInterval: 1.0,
	}
	sys.style = sys.getStyle(sys.genreID)
	return sys
}

// SetGenre configures genre-aware crack visual style.
func (s *EquipmentDamageCrackOverlaySystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.style = s.getStyle(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for crack overlay")
	}
}

// getStyle returns genre-specific crack generation parameters.
func (s *EquipmentDamageCrackOverlaySystem) getStyle(genreID string) crackGenreStyle {
	switch genreID {
	case "horror":
		// Organic decay: curved, spreading, dark reddish
		return crackGenreStyle{
			MinAngle: -0.8, MaxAngle: 0.8, Curvature: 0.6,
			BranchChance: 0.35, R: 0.35, G: 0.15, B: 0.1,
			MinLen: 0.04, MaxLen: 0.12,
		}
	case "cyberpunk":
		// Glitch lines: mostly horizontal, sharp angles, cyan-tinted
		return crackGenreStyle{
			MinAngle: -0.3, MaxAngle: 0.3, Curvature: 0.1,
			BranchChance: 0.2, R: 0.15, G: 0.25, B: 0.3,
			MinLen: 0.06, MaxLen: 0.15,
		}
	case "sci-fi", "scifi":
		// Circuit fractures: geometric, angular, blue-white
		return crackGenreStyle{
			MinAngle: -math.Pi / 4, MaxAngle: math.Pi / 4, Curvature: 0.05,
			BranchChance: 0.4, R: 0.3, G: 0.35, B: 0.45,
			MinLen: 0.03, MaxLen: 0.10,
		}
	case "post-apocalyptic", "postapoc":
		// Deep rust erosion: rough, wide, brown
		return crackGenreStyle{
			MinAngle: -0.6, MaxAngle: 0.6, Curvature: 0.4,
			BranchChance: 0.25, R: 0.4, G: 0.25, B: 0.1,
			MinLen: 0.05, MaxLen: 0.14,
		}
	default: // fantasy
		// Natural stone fractures: irregular branching, dark grey
		return crackGenreStyle{
			MinAngle: -0.7, MaxAngle: 0.7, Curvature: 0.3,
			BranchChance: 0.3, R: 0.2, G: 0.2, B: 0.2,
			MinLen: 0.03, MaxLen: 0.11,
		}
	}
}

// Update iterates entities with wear tint data and generates crack overlays.
func (s *EquipmentDamageCrackOverlaySystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime
	fullScan := s.timeSinceCheck >= s.updateInterval
	if fullScan {
		s.timeSinceCheck = 0.0
	}

	for _, entity := range entities {
		comp, _ := entity.GetComponent("equipment_crack_overlay")
		overlay, hasOverlay := comp.(*EquipmentCrackOverlayComponent)

		if !hasOverlay {
			if !fullScan {
				continue
			}
			if !entity.HasComponent("equipment_wear_tint") {
				continue
			}
			overlay = NewEquipmentCrackOverlayComponent()
			entity.AddComponent(overlay)
		}

		if fullScan {
			s.generateCracks(entity, overlay)
		}
	}
}

// generateCracks builds procedural crack segments from wear tint data.
func (s *EquipmentDamageCrackOverlaySystem) generateCracks(entity *Entity, overlay *EquipmentCrackOverlayComponent) {
	tintComp, _ := entity.GetComponent("equipment_wear_tint")
	tint, ok := tintComp.(*EquipmentWearTintComponent)
	if !ok || tint == nil || !tint.Enabled {
		overlay.Enabled = false
		overlay.Segments = overlay.Segments[:0]
		overlay.Intensity = 0.0
		overlay.TreeCount = 0
		return
	}

	density := tint.CrackDensity
	roughness := tint.EdgeRoughness

	// No cracks if density too low (Pristine/Worn states)
	if density < 0.05 {
		overlay.Enabled = false
		overlay.Segments = overlay.Segments[:0]
		overlay.Intensity = 0.0
		overlay.TreeCount = 0
		return
	}

	// Compute cache key from density+roughness quantized to avoid jitter
	quantDensity := uint64(density * 20)
	quantRough := uint64(roughness * 20)
	cacheKey := (quantDensity << 8) | quantRough
	if overlay.CacheKey == cacheKey {
		return // Pattern unchanged
	}

	// Number of crack trees scales with density (1–5)
	treeCount := int(1 + density*4)
	if treeCount > 5 {
		treeCount = 5
	}

	// Max segments per tree scales with density (2–6)
	maxPerTree := int(2 + density*4)
	if maxPerTree > 6 {
		maxPerTree = 6
	}

	// Reuse slice capacity
	overlay.Segments = overlay.Segments[:0]

	for tree := 0; tree < treeCount; tree++ {
		// Seed each tree deterministically from entity position
		originX := 0.1 + s.rng.Float64()*0.8
		originY := 0.1 + s.rng.Float64()*0.8
		angle := s.rng.Float64() * 2 * math.Pi

		curX, curY := originX, originY

		for seg := 0; seg < maxPerTree; seg++ {
			// Apply curvature
			angleShift := s.style.MinAngle + s.rng.Float64()*(s.style.MaxAngle-s.style.MinAngle)
			angle += angleShift * s.style.Curvature

			segLen := s.style.MinLen + s.rng.Float64()*(s.style.MaxLen-s.style.MinLen)
			nextX := curX + math.Cos(angle)*segLen
			nextY := curY + math.Sin(angle)*segLen

			// Clamp to [0,1]
			nextX = clampFloat(nextX, 0.0, 1.0)
			nextY = clampFloat(nextY, 0.0, 1.0)

			segWidth := 0.5 + density*1.5
			segDepth := 0.3 + density*0.5
			if segDepth > 1.0 {
				segDepth = 1.0
			}

			overlay.Segments = append(overlay.Segments, CrackSegment{
				X1: curX, Y1: curY,
				X2: nextX, Y2: nextY,
				Width:    segWidth,
				Depth:    segDepth,
				BranchID: tree,
			})

			curX, curY = nextX, nextY

			// Branch chance
			if s.rng.Float64() < s.style.BranchChance && seg < maxPerTree-1 {
				branchAngle := angle + (s.rng.Float64()-0.5)*math.Pi*0.8
				brLen := segLen * 0.6
				bx := curX + math.Cos(branchAngle)*brLen
				by := curY + math.Sin(branchAngle)*brLen
				bx = clampFloat(bx, 0.0, 1.0)
				by = clampFloat(by, 0.0, 1.0)

				overlay.Segments = append(overlay.Segments, CrackSegment{
					X1: curX, Y1: curY,
					X2: bx, Y2: by,
					Width:    segWidth * 0.7,
					Depth:    segDepth * 0.8,
					BranchID: tree,
				})
			}
		}
	}

	overlay.Intensity = clampFloat(density, 0.0, 1.0)
	overlay.EdgeRoughness = roughness
	overlay.ColorR = s.style.R
	overlay.ColorG = s.style.G
	overlay.ColorB = s.style.B
	overlay.Enabled = true
	overlay.CacheKey = cacheKey
	overlay.TreeCount = treeCount
}
