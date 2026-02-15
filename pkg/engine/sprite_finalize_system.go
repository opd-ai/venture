// Package engine provides the SpriteFinalizerSystem which applies post-processing
// effects (adaptive outline, rim lighting, edge shadow) to entity sprites at
// runtime. It monitors sprites for changes and re-applies finalization, ensuring
// all entities have crisp outlines and depth cues in the top-down view.
package engine

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// SpriteFinalizerSystem applies outline, rim lighting, and edge shadow
// post-processing to entity sprites. It tracks which sprites have been
// finalized via EbitenSprite.Finalized and re-processes when needed.
type SpriteFinalizerSystem struct {
	world  *World
	logger *logrus.Entry
	seed   int64
}

// NewSpriteFinalizerSystem creates a new sprite finalizer system.
func NewSpriteFinalizerSystem(world *World, seed int64) *SpriteFinalizerSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "sprite_finalizer",
	})
	logger.Debug("sprite finalizer system created")
	return &SpriteFinalizerSystem{
		world:  world,
		logger: logger,
		seed:   seed,
	}
}

// Update scans entities for sprites that need finalization.
func (s *SpriteFinalizerSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		spriteComp := entity.GetSprite()
		if spriteComp == nil || spriteComp.Image == nil || !spriteComp.Visible {
			continue
		}
		if spriteComp.Finalized {
			continue
		}

		s.finalizeSprite(entity, spriteComp)
		spriteComp.Finalized = true
	}
}

// finalizeSprite applies post-processing to a single entity's sprite.
func (s *SpriteFinalizerSystem) finalizeSprite(entity *Entity, spriteComp *EbitenSprite) {
	entitySeed := s.seed ^ int64(entity.ID)
	cfg := sprites.DefaultFinalizerConfig(entitySeed)

	// Finalize the main sprite
	spriteComp.Image = s.processImage(spriteComp.Image, cfg)

	// Finalize directional sprites if present
	for dir, dirImg := range spriteComp.DirectionalImages {
		if dirImg != nil {
			dirCfg := sprites.DefaultFinalizerConfig(entitySeed + int64(dir)*7)
			spriteComp.DirectionalImages[dir] = s.processImage(dirImg, dirCfg)
		}
	}
}

// processImage converts an ebiten image to RGBA, finalizes it, and converts back.
// Returns the original image unchanged if pixel reading is unavailable (e.g., during tests).
func (s *SpriteFinalizerSystem) processImage(img *ebiten.Image, cfg sprites.FinalizerConfig) *ebiten.Image {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return img
	}

	// Read pixels from ebiten.Image — may fail before game starts
	srcRGBA, err := safeReadPixels(img, w, h)
	if err != nil {
		return img // graceful fallback
	}

	result := sprites.FinalizeEntitySprite(srcRGBA, cfg)
	return ebiten.NewImageFromImage(result)
}

// safeReadPixels reads pixel data from an ebiten.Image, recovering from panics
// that occur when called before the game loop starts.
func safeReadPixels(img *ebiten.Image, w, h int) (srcRGBA *image.RGBA, err error) {
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

// GetProcessedCount returns the number of finalized sprites for diagnostics.
func (s *SpriteFinalizerSystem) GetProcessedCount() int {
	count := 0
	if s.world != nil {
		for _, entity := range s.world.GetEntities() {
			if sp := entity.GetSprite(); sp != nil && sp.Finalized {
				count++
			}
		}
	}
	return count
}
