// Package engine provides the EquipmentMaterialSheenSystem which bridges the
// MaterialVisualProperties defined in pkg/rendering/sprites/equipment.go (Sheen,
// Roughness, Reflectivity) with per-entity visual state. It reads equipped items,
// aggregates their material properties, and writes a MaterialSheenComponent that
// the render pipeline can use for specular highlight overlays.
package engine

import (
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// genreSheenPreset holds genre-specific highlight color tinting and intensity scaling.
type genreSheenPreset struct {
	R, G, B        float64
	IntensityScale float64
	PulseSpeed     float64
}

// EquipmentMaterialSheenSystem computes animated specular highlight parameters
// from the material types of equipped items. It lazily attaches a
// MaterialSheenComponent to entities that have equipment.
type EquipmentMaterialSheenSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string
	preset  genreSheenPreset

	// Throttle full scans to avoid per-frame overhead for new entities
	updateInterval float64
	timeSinceCheck float64
}

// NewEquipmentMaterialSheenSystem creates a new equipment material sheen system.
func NewEquipmentMaterialSheenSystem(world *World, seed int64) *EquipmentMaterialSheenSystem {
	var logEntry *logrus.Entry
	if world != nil && world.logger != nil {
		logEntry = world.logger.Logger.WithField("system_name", "equipment_material_sheen")
		logEntry.Debug("equipment material sheen system created")
	}

	sys := &EquipmentMaterialSheenSystem{
		world:          world,
		logger:         logEntry,
		rng:            rand.New(rand.NewSource(seed)),
		genreID:        "fantasy",
		updateInterval: 1.0,
	}
	sys.preset = sys.getPreset(sys.genreID)
	return sys
}

// SetGenre configures genre-aware highlight tinting and intensity.
func (s *EquipmentMaterialSheenSystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.preset = s.getPreset(genreID)
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for material sheen")
	}
}

// getPreset returns a genre-specific highlight color and intensity.
func (s *EquipmentMaterialSheenSystem) getPreset(genreID string) genreSheenPreset {
	switch genreID {
	case "horror":
		// Dim, reddish highlights
		return genreSheenPreset{R: 0.9, G: 0.7, B: 0.6, IntensityScale: 0.5, PulseSpeed: 0.8}
	case "cyberpunk":
		// Bright neon cyan highlights
		return genreSheenPreset{R: 0.4, G: 1.0, B: 1.0, IntensityScale: 1.3, PulseSpeed: 2.0}
	case "sci-fi", "scifi":
		// Cool blue-white highlights
		return genreSheenPreset{R: 0.8, G: 0.9, B: 1.0, IntensityScale: 1.1, PulseSpeed: 1.5}
	case "post-apocalyptic", "postapoc":
		// Warm dusty highlights
		return genreSheenPreset{R: 1.0, G: 0.85, B: 0.6, IntensityScale: 0.7, PulseSpeed: 1.0}
	case "fantasy":
		// Neutral warm white highlights
		return genreSheenPreset{R: 1.0, G: 0.95, B: 0.9, IntensityScale: 1.0, PulseSpeed: 1.5}
	default:
		return genreSheenPreset{R: 1.0, G: 1.0, B: 1.0, IntensityScale: 1.0, PulseSpeed: 1.5}
	}
}

// Update iterates entities with equipment and computes material sheen parameters.
func (s *EquipmentMaterialSheenSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceCheck += deltaTime

	fullScan := false
	if s.timeSinceCheck >= s.updateInterval {
		s.timeSinceCheck = 0
		fullScan = true
	}

	for _, entity := range entities {
		comp, _ := entity.GetComponent("material_sheen")
		sheen, hasSheen := comp.(*MaterialSheenComponent)

		if !hasSheen {
			if !fullScan {
				continue
			}
			if !entity.HasComponent("equipment") {
				continue
			}
			sheen = NewMaterialSheenComponent()
			entity.AddComponent(sheen)
		}

		// Advance animation phase every frame for smooth highlight pulse
		sheen.Phase += deltaTime * sheen.PulseSpeed
		if sheen.Phase > 2*math.Pi {
			sheen.Phase -= 2 * math.Pi
		}

		// Only recompute material aggregation on full scans (throttled)
		if fullScan {
			s.computeSheen(entity, sheen)
		}
	}
}

// computeSheen aggregates material visual properties from all equipped items.
func (s *EquipmentMaterialSheenSystem) computeSheen(entity *Entity, sheen *MaterialSheenComponent) {
	equipComp := s.getEquipmentComponent(entity)
	if equipComp == nil {
		sheen.Enabled = false
		return
	}

	slots := []EquipmentSlot{
		SlotMainHand, SlotOffHand, SlotHead, SlotChest,
		SlotLegs, SlotBoots, SlotGloves,
	}

	var totalSheen, totalReflect, totalRough float64
	var count int
	materialCounts := make(map[sprites.MaterialType]int)

	for _, slot := range slots {
		itm := equipComp.GetEquipped(slot)
		if itm == nil {
			continue
		}

		material := sprites.GetMaterialTypeFromTags(itm.Tags, s.genreID)
		props := sprites.GetMaterialVisualProperties(material)

		totalSheen += props.Sheen
		totalReflect += props.Reflectivity
		totalRough += props.Roughness
		materialCounts[material]++
		count++
	}

	if count == 0 {
		sheen.Enabled = false
		return
	}

	// Average across equipped items
	avgSheen := totalSheen / float64(count)
	avgReflect := totalReflect / float64(count)
	avgRough := totalRough / float64(count)

	// Apply genre intensity scaling
	sheen.SheenIntensity = clampFloat(avgSheen*s.preset.IntensityScale, 0.0, 1.0)
	sheen.Reflectivity = clampFloat(avgReflect*s.preset.IntensityScale, 0.0, 1.0)
	sheen.Roughness = avgRough

	// Genre-tinted highlight color
	sheen.ColorR = s.preset.R
	sheen.ColorG = s.preset.G
	sheen.ColorB = s.preset.B

	// Pulse speed from genre preset, dampened by roughness
	sheen.PulseSpeed = s.preset.PulseSpeed * (1.0 - avgRough*0.5)

	// Highlight size proportional to sheen intensity
	sheen.HighlightSize = 2.0 + avgSheen*6.0

	// Find dominant material
	sheen.DominantMaterial = s.getDominantMaterial(materialCounts)

	// Enable only if sheen is meaningful
	sheen.Enabled = sheen.SheenIntensity > 0.05
}

// getDominantMaterial returns the most frequently occurring material type.
func (s *EquipmentMaterialSheenSystem) getDominantMaterial(counts map[sprites.MaterialType]int) string {
	var dominant sprites.MaterialType
	maxCount := 0
	for mat, c := range counts {
		if c > maxCount {
			maxCount = c
			dominant = mat
		}
	}
	return dominant.String()
}

// getEquipmentComponent retrieves the typed equipment component from an entity.
func (s *EquipmentMaterialSheenSystem) getEquipmentComponent(entity *Entity) *EquipmentComponent {
	comp, ok := entity.GetComponent("equipment")
	if !ok || comp == nil {
		return nil
	}
	equipComp, ok := comp.(*EquipmentComponent)
	if !ok {
		return nil
	}
	return equipComp
}
