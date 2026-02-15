// Package engine provides the SpriteDepthEnhanceSystem which applies
// form-aware volumetric shading to entity sprites at runtime. It runs
// after DirectionalSpriteSystem and before SpriteFinalizerSystem,
// adding 3D-form-appropriate highlights, diffuse lighting, contact
// shadows, and subsurface scattering to each body part zone. The
// enhancement is applied once per sprite (tracked via DepthProcessed
// flag) and is baked into the image for zero per-frame cost.
package engine

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// SpriteDepthEnhanceSystem applies volumetric depth shading to entity sprites.
// It reads EbitenSprite images and applies form-aware 3D lighting, making
// head regions look spherical, torsos look cylindrical, and limbs look tubular
// when viewed from the top-down camera.
type SpriteDepthEnhanceSystem struct {
	world  *World
	logger *logrus.Entry
	seed   int64
	genre  string
}

// NewSpriteDepthEnhanceSystem creates a new sprite depth enhancement system.
func NewSpriteDepthEnhanceSystem(world *World, seed int64) *SpriteDepthEnhanceSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "sprite_depth_enhance",
	})
	logger.Debug("sprite depth enhance system created")
	return &SpriteDepthEnhanceSystem{
		world:  world,
		logger: logger,
		seed:   seed,
	}
}

// SetGenre configures genre-specific shading parameters.
func (s *SpriteDepthEnhanceSystem) SetGenre(genreID string) {
	s.genre = genreID
}

// Update scans entities for sprites that need depth enhancement.
func (s *SpriteDepthEnhanceSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		spriteComp := entity.GetSprite()
		if spriteComp == nil || spriteComp.Image == nil || !spriteComp.Visible {
			continue
		}
		if spriteComp.DepthProcessed {
			continue
		}
		// Must run before finalization — skip already-finalized sprites
		if spriteComp.Finalized {
			continue
		}

		s.enhanceSprite(entity, spriteComp)
		spriteComp.DepthProcessed = true
	}
}

// enhanceSprite applies volumetric depth shading to a single entity's sprite.
func (s *SpriteDepthEnhanceSystem) enhanceSprite(entity *Entity, spriteComp *EbitenSprite) {
	entitySeed := s.seed ^ int64(entity.ID)
	cfg := s.buildConfig(entitySeed)

	// Determine creature form for nonhumanoid entities
	creatureForm := ""
	if comp, ok := entity.GetComponent("creature_visual"); ok {
		if cv, ok := comp.(*CreatureVisualComponent); ok {
			creatureForm = string(cv.Form)
		}
	}

	// Enhance main sprite
	spriteComp.Image = s.processImage(spriteComp.Image, cfg, creatureForm, entitySeed)

	// Enhance directional sprites
	for dir, dirImg := range spriteComp.DirectionalImages {
		if dirImg != nil {
			dirCfg := s.buildConfig(entitySeed + int64(dir)*13)
			spriteComp.DirectionalImages[dir] = s.processImage(dirImg, dirCfg, creatureForm, entitySeed+int64(dir)*13)
		}
	}
}

// processImage reads pixels, applies depth enhancement, and returns new image.
func (s *SpriteDepthEnhanceSystem) processImage(img *ebiten.Image, cfg sprites.DepthEnhanceConfig, creatureForm string, seed int64) *ebiten.Image {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return img
	}

	rgba, err := safeReadPixelsForDepth(img, w, h)
	if err != nil {
		return img // graceful fallback
	}

	if creatureForm != "" && creatureForm != "humanoid" {
		sprites.ApplyDepthEnhancementForCreature(rgba, creatureForm, cfg)
	} else {
		sprites.ApplyDepthEnhancement(rgba, cfg)
	}

	return ebiten.NewImageFromImage(rgba)
}

// safeReadPixelsForDepth reads pixel data from an ebiten.Image, recovering from
// panics that occur when called before the game loop starts.
func safeReadPixelsForDepth(img *ebiten.Image, w, h int) (srcRGBA *image.RGBA, err error) {
	defer func() {
		if r := recover(); r != nil {
			srcRGBA = nil
			err = fmt.Errorf("pixel read unavailable: %v", r)
		}
	}()
	srcRGBA = image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			srcRGBA.Set(x, y, img.At(x, y))
		}
	}
	return srcRGBA, nil
}

// buildConfig returns a depth enhancement config tuned to the current genre.
func (s *SpriteDepthEnhanceSystem) buildConfig(seed int64) sprites.DepthEnhanceConfig {
	cfg := sprites.DefaultDepthEnhanceConfig(seed)

	switch s.genre {
	case "horror":
		cfg.DiffuseStrength = 0.20
		cfg.SpecularIntensity = 0.20
		cfg.ContactShadowStrength = 0.35
		cfg.SubsurfaceStrength = 0.04
	case "cyberpunk":
		cfg.SpecularIntensity = 0.50
		cfg.SpecularPower = 24.0
		cfg.DiffuseStrength = 0.35
		cfg.SubsurfaceStrength = 0.03
	case "scifi", "sci-fi":
		cfg.SpecularIntensity = 0.45
		cfg.SpecularPower = 20.0
		cfg.DiffuseStrength = 0.32
	case "postapoc", "post-apocalyptic":
		cfg.DiffuseStrength = 0.25
		cfg.SpecularIntensity = 0.20
		cfg.ContactShadowStrength = 0.30
	case "fantasy":
		cfg.SubsurfaceStrength = 0.12
		cfg.SpecularIntensity = 0.30
	}

	return cfg
}

// GetProcessedCount returns the number of depth-enhanced sprites for diagnostics.
func (s *SpriteDepthEnhanceSystem) GetProcessedCount() int {
	count := 0
	if s.world != nil {
		for _, entity := range s.world.GetEntities() {
			if sp := entity.GetSprite(); sp != nil && sp.DepthProcessed {
				count++
			}
		}
	}
	return count
}
