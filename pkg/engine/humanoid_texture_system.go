// Package engine provides the HumanoidTextureSystem which assigns humanoid-specific
// procedural surface textures to player characters and NPCs. It reads the
// CreatureVisualComponent to identify humanoid entities and populates a
// HumanoidTextureComponent with skin textures (smooth, freckled, scarred, weathered,
// tattooed), clothing fabric textures (linen, leather, silk, wool, chainmail, plate),
// and hair textures (straight, wavy, curly, braided). Texture parameters are derived
// from the entity seed for deterministic generation. Genre changes trigger re-population.
package engine

import (
	"image/color"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// HumanoidTextureSystem scans humanoid entities and attaches/updates
// HumanoidTextureComponent with form-appropriate surface texture parameters.
type HumanoidTextureSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	genreID string

	scanInterval  float64
	timeSinceScan float64
}

// NewHumanoidTextureSystem creates a new humanoid texture system.
func NewHumanoidTextureSystem(world *World, seed int64) *HumanoidTextureSystem {
	var logEntry *logrus.Entry
	if world != nil && world.GetLogger() != nil {
		logEntry = world.GetLogger().WithField("system_name", "humanoid_texture")
		logEntry.Debug("humanoid texture system created")
	}

	return &HumanoidTextureSystem{
		world:         world,
		logger:        logEntry,
		rng:           rand.New(rand.NewSource(seed)),
		genreID:       "fantasy",
		scanInterval:  2.0, // Scan every 2 seconds
		timeSinceScan: 0,
	}
}

// SetGenre configures genre-aware texture preferences.
func (s *HumanoidTextureSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for humanoid textures")
	}
}

// Update scans entities and attaches/configures humanoid texture components.
func (s *HumanoidTextureSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceScan += deltaTime
	if s.timeSinceScan < s.scanInterval {
		return
	}
	s.timeSinceScan = 0

	for _, entity := range entities {
		if entity == nil {
			continue
		}

		// Check if entity is a humanoid
		if !s.isHumanoid(entity) {
			continue
		}

		comp, has := entity.GetComponent("humanoid_texture")
		if !has {
			ht := NewHumanoidTextureComponent()
			s.populateFromSeed(ht, entity)
			entity.AddComponent(ht)
			continue
		}

		existing, ok := comp.(*HumanoidTextureComponent)
		if !ok {
			continue
		}

		// Re-populate if genre changed
		if existing.GenreID != s.genreID {
			s.populateFromSeed(existing, entity)
			existing.Dirty = true
		}
	}
}

// isHumanoid determines if an entity should receive humanoid textures.
func (s *HumanoidTextureSystem) isHumanoid(entity *Entity) bool {
	// Check if it's a player entity
	if _, hasPlayer := entity.GetComponent("player"); hasPlayer {
		return true
	}

	// Check creature visual component for humanoid form
	cvComp, hasCv := entity.GetComponent("creature_visual")
	if hasCv {
		if cv, ok := cvComp.(*CreatureVisualComponent); ok {
			return cv.Form == FormHumanoid
		}
	}

	// Check for NPC component (NPCs are typically humanoid)
	if _, hasNPC := entity.GetComponent("npc"); hasNPC {
		return true
	}

	// Check for merchant component
	if _, hasMerchant := entity.GetComponent("merchant"); hasMerchant {
		return true
	}

	return false
}

// populateFromSeed fills a HumanoidTextureComponent using entity seed and genre.
func (s *HumanoidTextureSystem) populateFromSeed(ht *HumanoidTextureComponent, entity *Entity) {
	seed := int64(entity.ID)
	texSet := sprites.GenerateHumanoidTextureSet(seed, s.genreID)

	// Skin texture
	ht.SkinTextureType = int(texSet.SkinTexture.Type)
	ht.SkinIntensity = texSet.SkinTexture.Intensity
	ht.SkinScale = texSet.SkinTexture.Scale
	ht.SkinPrimaryColor = texSet.SkinTexture.PrimaryColor
	ht.SkinSecondaryColor = texSet.SkinTexture.SecondaryColor

	// Upper body clothing
	ht.ClothingTopTextureType = int(texSet.ClothingTop.Type)
	ht.ClothingTopIntensity = texSet.ClothingTop.Intensity
	ht.ClothingTopScale = texSet.ClothingTop.Scale
	ht.ClothingTopPrimaryColor = texSet.ClothingTop.PrimaryColor
	ht.ClothingTopSecondaryColor = texSet.ClothingTop.SecondaryColor

	// Lower body clothing
	ht.ClothingBottomTextureType = int(texSet.ClothingBottom.Type)
	ht.ClothingBottomIntensity = texSet.ClothingBottom.Intensity
	ht.ClothingBottomScale = texSet.ClothingBottom.Scale
	ht.ClothingBottomPrimaryColor = texSet.ClothingBottom.PrimaryColor
	ht.ClothingBottomSecondaryColor = texSet.ClothingBottom.SecondaryColor

	// Hair
	ht.HairTextureType = int(texSet.HairTexture.Type)
	ht.HairIntensity = texSet.HairTexture.Intensity
	ht.HairScale = texSet.HairTexture.Scale
	ht.HairPrimaryColor = texSet.HairTexture.PrimaryColor
	ht.HairSecondaryColor = texSet.HairTexture.SecondaryColor
	ht.HairDirection = texSet.HairTexture.Direction

	ht.GenreID = s.genreID
	ht.Enabled = true

	// Mark animation dirty so sprite regenerates with texture
	if animComp := entity.GetAnimation(); animComp != nil {
		animComp.Dirty = true
	}

	if s.logger != nil && s.logger.Logger.GetLevel() >= logrus.DebugLevel {
		s.logger.WithFields(logrus.Fields{
			"entity_id":     entity.ID,
			"skin_type":     sprites.HumanoidTextureType(ht.SkinTextureType).String(),
			"clothing_type": sprites.HumanoidTextureType(ht.ClothingTopTextureType).String(),
			"hair_type":     sprites.HumanoidTextureType(ht.HairTextureType).String(),
		}).Debug("humanoid texture assigned")
	}
}

// GetActiveTextureCount returns the number of entities with active humanoid textures.
func (s *HumanoidTextureSystem) GetActiveTextureCount() int {
	if s.world == nil {
		return 0
	}
	count := 0
	for _, entity := range s.world.GetEntities() {
		if comp, has := entity.GetComponent("humanoid_texture"); has {
			if ht, ok := comp.(*HumanoidTextureComponent); ok && ht.Enabled {
				count++
			}
		}
	}
	return count
}

// GetTextureBreakdown returns counts by texture type for debugging.
func (s *HumanoidTextureSystem) GetTextureBreakdown() map[string]int {
	breakdown := make(map[string]int)
	if s.world == nil {
		return breakdown
	}

	for _, entity := range s.world.GetEntities() {
		comp, has := entity.GetComponent("humanoid_texture")
		if !has {
			continue
		}
		ht, ok := comp.(*HumanoidTextureComponent)
		if !ok || !ht.Enabled {
			continue
		}

		skinName := sprites.HumanoidTextureType(ht.SkinTextureType).String()
		breakdown["skin:"+skinName]++

		clothingName := sprites.HumanoidTextureType(ht.ClothingTopTextureType).String()
		breakdown["clothing:"+clothingName]++

		hairName := sprites.HumanoidTextureType(ht.HairTextureType).String()
		breakdown["hair:"+hairName]++
	}

	return breakdown
}

// ApplyTextureToSprite applies humanoid textures to a rendered sprite buffer.
// This is called during sprite generation/regeneration.
func ApplyHumanoidTexturesToSprite(ht *HumanoidTextureComponent, buf sprites.TextureBuffer, seed int64) {
	if ht == nil || !ht.Enabled || buf == nil {
		return
	}

	bounds := buf.Bounds()
	h := bounds.Dy()

	// Approximate body part regions for a 32x32 top-down sprite
	// Head: top 35%, Torso: middle 50%, Legs: bottom 15%
	headRegion := bounds
	headRegion.Max.Y = bounds.Min.Y + h*35/100

	torsoRegion := bounds
	torsoRegion.Min.Y = bounds.Min.Y + h*35/100
	torsoRegion.Max.Y = bounds.Min.Y + h*85/100

	legRegion := bounds
	legRegion.Min.Y = bounds.Min.Y + h*85/100

	// Apply hair texture to head region
	sprites.ApplyHumanoidTexture(
		buf,
		headRegion,
		sprites.HumanoidTextureParams{
			Type:           sprites.HumanoidTextureType(ht.HairTextureType),
			Intensity:      ht.HairIntensity,
			Scale:          ht.HairScale,
			PrimaryColor:   ht.HairPrimaryColor,
			SecondaryColor: ht.HairSecondaryColor,
			Direction:      ht.HairDirection,
		},
		seed^0x48414952, // "HAIR"
	)

	// Apply skin texture to exposed skin (face portion of head)
	faceRegion := headRegion
	faceRegion.Min.Y = headRegion.Min.Y + (headRegion.Dy() / 2)
	sprites.ApplyHumanoidTexture(
		buf,
		faceRegion,
		sprites.HumanoidTextureParams{
			Type:           sprites.HumanoidTextureType(ht.SkinTextureType),
			Intensity:      ht.SkinIntensity,
			Scale:          ht.SkinScale,
			PrimaryColor:   ht.SkinPrimaryColor,
			SecondaryColor: ht.SkinSecondaryColor,
		},
		seed^0x534B494E, // "SKIN"
	)

	// Apply clothing texture to torso
	sprites.ApplyHumanoidTexture(
		buf,
		torsoRegion,
		sprites.HumanoidTextureParams{
			Type:           sprites.HumanoidTextureType(ht.ClothingTopTextureType),
			Intensity:      ht.ClothingTopIntensity,
			Scale:          ht.ClothingTopScale,
			PrimaryColor:   ht.ClothingTopPrimaryColor,
			SecondaryColor: ht.ClothingTopSecondaryColor,
		},
		seed^0x54525342, // "TRSB"
	)

	// Apply clothing texture to legs
	sprites.ApplyHumanoidTexture(
		buf,
		legRegion,
		sprites.HumanoidTextureParams{
			Type:           sprites.HumanoidTextureType(ht.ClothingBottomTextureType),
			Intensity:      ht.ClothingBottomIntensity,
			Scale:          ht.ClothingBottomScale,
			PrimaryColor:   ht.ClothingBottomPrimaryColor,
			SecondaryColor: ht.ClothingBottomSecondaryColor,
		},
		seed^0x4C454753, // "LEGS"
	)
}

// ResolveTextureColor returns the effective texture color blending
// the component's stored colors with an optional base color.
func ResolveTextureColor(primary, secondary, base color.RGBA, intensity float64) color.RGBA {
	if intensity <= 0 {
		return base
	}
	if intensity > 1 {
		intensity = 1
	}

	inv := 1.0 - intensity
	r := uint8(float64(base.R)*inv + float64(primary.R)*intensity)
	g := uint8(float64(base.G)*inv + float64(primary.G)*intensity)
	b := uint8(float64(base.B)*inv + float64(primary.B)*intensity)

	return color.RGBA{R: r, G: g, B: b, A: base.A}
}
