// Package engine provides the SpriteColorTemperatureSystem which applies
// warm/cool color temperature grading and sharp specular highlights to entity
// sprites. It runs after SpriteDepthEnhanceSystem and before
// SpriteFinalizerSystem, adding genre-aware lighting color that makes sprites
// look lit by a real light source rather than flat ambient. The enhancement is
// applied once per sprite (tracked via ColorTempProcessed flag) and baked into
// the image for zero per-frame cost.
package engine

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// SpriteColorTemperatureSystem applies color temperature grading and specular
// highlights to entity sprites. Warm tones are added on the light-facing side,
// cool tones on the shadow side, and a sharp specular spot marks the point of
// maximum reflection. Genre presets control the color balance.
type SpriteColorTemperatureSystem struct {
	world  *World
	logger *logrus.Entry
	seed   int64
	genre  string
}

// NewSpriteColorTemperatureSystem creates a new color temperature system.
func NewSpriteColorTemperatureSystem(world *World, seed int64) *SpriteColorTemperatureSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "sprite_color_temperature",
	})
	logger.Debug("sprite color temperature system created")
	return &SpriteColorTemperatureSystem{
		world:  world,
		logger: logger,
		seed:   seed,
	}
}

// SetGenre configures genre-specific color temperature parameters.
func (s *SpriteColorTemperatureSystem) SetGenre(genreID string) {
	s.genre = genreID
}

// Update scans entities for sprites that need color temperature grading.
func (s *SpriteColorTemperatureSystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		spriteComp := entity.GetSprite()
		if spriteComp == nil || spriteComp.Image == nil || !spriteComp.Visible {
			continue
		}
		if spriteComp.ColorTempProcessed {
			continue
		}
		// Must run after depth enhancement but before finalization
		if !spriteComp.DepthProcessed {
			continue
		}
		if spriteComp.Finalized {
			continue
		}

		s.applyColorTemperature(entity, spriteComp)
		spriteComp.ColorTempProcessed = true
	}
}

// applyColorTemperature applies genre-aware color temperature grading to a
// single entity's sprite and all its directional variants.
func (s *SpriteColorTemperatureSystem) applyColorTemperature(entity *Entity, spriteComp *EbitenSprite) {
	entitySeed := s.seed ^ int64(entity.ID)
	cfg := sprites.GenreColorTemperatureConfig(s.genre, entitySeed)

	// Grade the main sprite
	spriteComp.Image = s.processImage(spriteComp.Image, cfg, entitySeed)

	// Grade directional sprites
	for dir, dirImg := range spriteComp.DirectionalImages {
		if dirImg != nil {
			dirCfg := sprites.GenreColorTemperatureConfig(s.genre, entitySeed+int64(dir)*17)
			spriteComp.DirectionalImages[dir] = s.processImage(dirImg, dirCfg, entitySeed+int64(dir)*17)
		}
	}
}

// processImage reads pixels from an ebiten.Image, applies color temperature
// grading, and returns a new ebiten.Image with the result.
func (s *SpriteColorTemperatureSystem) processImage(img *ebiten.Image, cfg sprites.ColorTemperatureConfig, seed int64) *ebiten.Image {
	if img == nil {
		return nil
	}

	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w == 0 || h == 0 {
		return img
	}

	rgba, err := safeReadPixelsForColorTemp(img, w, h)
	if err != nil {
		return img
	}

	sprites.ApplyColorTemperature(rgba, cfg)

	return ebiten.NewImageFromImage(rgba)
}

// safeReadPixelsForColorTemp reads pixel data from an ebiten.Image, recovering
// from panics that can occur when called before the game loop starts.
func safeReadPixelsForColorTemp(img *ebiten.Image, w, h int) (result *image.RGBA, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("ReadPixels panic: %v", r)
		}
	}()

	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	img.ReadPixels(rgba.Pix)
	return rgba, nil
}
