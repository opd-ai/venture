// Package engine provides the shadow casting system.
// This file implements ShadowSystem which processes shadow-casting entities
// and generates shadows based on light sources. The system supports multiple
// shadow types (hard, soft, contact) and integrates with the lighting pipeline.
//
// Design Philosophy:
// - Performance-conscious: ray-casting optimization, viewport culling
// - Genre-aware: shadow intensity varies by genre atmosphere
// - Extensible: supports multiple shadow types and rendering modes
// - Integration: works with LightingSystem for unified lighting/shadow pass
package engine

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// ShadowSystem processes shadow-casting entities and renders shadows.
// This system runs in conjunction with LightingSystem to create
// realistic lighting and shadow effects.
type ShadowSystem struct {
	world  *World
	logger *logrus.Entry

	// Viewport tracking for culling
	cameraX     float64
	cameraY     float64
	viewportW   int
	viewportH   int
	viewportSet bool

	// Shadow buffer (reused each frame)
	shadowBuffer *ebiten.Image

	// Configuration
	enabled       bool
	maxShadows    int
	renderQuality float64 // 0.5 = half-res, 1.0 = full-res, 2.0 = super-sample
}

// shadowCaster combines entity info for shadow casting.
type shadowCaster struct {
	shadow   *ShadowComponent
	position *PositionComponent
	x        float64
	y        float64
}

// NewShadowSystem creates a new shadow system.
func NewShadowSystem(world *World) *ShadowSystem {
	return NewShadowSystemWithLogger(world, nil)
}

// NewShadowSystemWithLogger creates a new shadow system with a logger.
func NewShadowSystemWithLogger(world *World, logger *logrus.Logger) *ShadowSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system", "shadow")
	}

	return &ShadowSystem{
		world:         world,
		logger:        logEntry,
		enabled:       true,
		maxShadows:    100,
		renderQuality: 1.0,
	}
}

// SetEnabled toggles shadow rendering on/off.
func (s *ShadowSystem) SetEnabled(enabled bool) {
	s.enabled = enabled
	if s.logger != nil {
		s.logger.WithField("enabled", enabled).Debug("Shadow system toggled")
	}
}

// SetMaxShadows sets the maximum number of shadows to render per frame.
func (s *ShadowSystem) SetMaxShadows(max int) {
	s.maxShadows = max
}

// SetRenderQuality sets the shadow rendering quality (0.5-2.0).
func (s *ShadowSystem) SetRenderQuality(quality float64) {
	if quality < 0.25 {
		quality = 0.25
	}
	if quality > 2.0 {
		quality = 2.0
	}
	s.renderQuality = quality
}

// SetViewport updates the camera position and viewport size for culling.
func (s *ShadowSystem) SetViewport(cameraX, cameraY float64, width, height int) {
	s.cameraX = cameraX
	s.cameraY = cameraY
	s.viewportW = width
	s.viewportH = height
	s.viewportSet = true
}

// Update processes shadow-casting entities (no per-frame updates needed).
func (s *ShadowSystem) Update(entities []*Entity, deltaTime float64) {
	// Shadow system doesn't need per-frame updates
	// All rendering happens in RenderShadows
}

// RenderShadows renders shadows for all shadow-casting entities.
// This should be called after the main render pass but before lighting.
// screen: target render surface
// lightX, lightY: position of light source casting shadows
// lightRadius: effective range of the light
// Returns a shadow buffer image that can be composited with the scene.
func (s *ShadowSystem) RenderShadows(screen *ebiten.Image, lightX, lightY, lightRadius float64) *ebiten.Image {
	if !s.enabled {
		return nil
	}

	// Initialize shadow buffer if needed
	if s.shadowBuffer == nil {
		w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
		s.shadowBuffer = ebiten.NewImage(w, h)
	}

	// Clear shadow buffer
	s.shadowBuffer.Clear()

	// Collect shadow casters in viewport
	casters := s.collectShadowCasters(lightX, lightY, lightRadius)

	// Render shadows for each caster
	for _, caster := range casters {
		s.renderShadowForEntity(s.shadowBuffer, caster, lightX, lightY)
	}

	return s.shadowBuffer
}

// collectShadowCasters finds all entities that should cast shadows from this light.
func (s *ShadowSystem) collectShadowCasters(lightX, lightY, lightRadius float64) []*shadowCaster {
	casters := make([]*shadowCaster, 0, s.maxShadows)

	entities := s.world.GetEntities()
	for _, entity := range entities {
		// Get shadow component
		shadowComp, hasShadow := entity.GetComponent("shadow")
		if !hasShadow {
			continue
		}
		shadow, ok := shadowComp.(*ShadowComponent)
		if !ok || !shadow.Enabled || !shadow.CastsShadow {
			continue
		}

		// Get position component
		posComp, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Distance check: skip if outside light radius
		dx := pos.X - lightX
		dy := pos.Y - lightY
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > lightRadius+shadow.Radius {
			continue
		}

		// Viewport culling
		if s.viewportSet {
			if pos.X+shadow.Radius < s.cameraX ||
				pos.X-shadow.Radius > s.cameraX+float64(s.viewportW) ||
				pos.Y+shadow.Radius < s.cameraY ||
				pos.Y-shadow.Radius > s.cameraY+float64(s.viewportH) {
				continue
			}
		}

		casters = append(casters, &shadowCaster{
			shadow:   shadow,
			position: pos,
			x:        pos.X,
			y:        pos.Y,
		})

		// Limit shadows
		if len(casters) >= s.maxShadows {
			break
		}
	}

	return casters
}

// renderShadowForEntity renders shadow for a single entity.
func (s *ShadowSystem) renderShadowForEntity(target *ebiten.Image, caster *shadowCaster, lightX, lightY float64) {
	switch caster.shadow.ShadowType {
	case ShadowTypeHard:
		s.renderHardShadow(target, caster, lightX, lightY)
	case ShadowTypeSoft:
		s.renderSoftShadow(target, caster, lightX, lightY)
	case ShadowTypeContact:
		s.renderContactShadow(target, caster, lightX, lightY)
	}
}

// renderHardShadow renders a hard-edged shadow using ray-casting.
func (s *ShadowSystem) renderHardShadow(target *ebiten.Image, caster *shadowCaster, lightX, lightY float64) {
	// Calculate shadow vector (away from light)
	dx := caster.x - lightX
	dy := caster.y - lightY
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.1 {
		return // Entity is at light source
	}

	// Normalize direction
	dirX := dx / dist
	dirY := dy / dist

	// Calculate shadow length (based on distance from light)
	shadowLength := caster.shadow.Radius * 3.0

	// Draw shadow as a stretched circle/ellipse
	shadowX := caster.x + dirX*shadowLength*0.5
	shadowY := caster.y + dirY*shadowLength*0.5

	// Create shadow image (simple approach: filled rectangle)
	shadowW := int(caster.shadow.Radius * 2)
	shadowH := int(shadowLength)
	if shadowW < 1 || shadowH < 1 {
		return
	}

	// Create shadow rectangle
	shadowImg := ebiten.NewImage(shadowW, shadowH)
	shadowImg.Fill(caster.shadow.Color)

	// Draw shadow with rotation
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-float64(shadowW)/2, -float64(shadowH)/2)
	angle := math.Atan2(dirY, dirX) + math.Pi/2 // Rotate to align with shadow direction
	opts.GeoM.Rotate(angle)
	opts.GeoM.Translate(shadowX, shadowY)

	// Apply opacity
	opts.ColorScale.ScaleAlpha(float32(caster.shadow.Opacity))

	target.DrawImage(shadowImg, opts)
}

// renderSoftShadow renders a soft-edged shadow with penumbra.
func (s *ShadowSystem) renderSoftShadow(target *ebiten.Image, caster *shadowCaster, lightX, lightY float64) {
	// For now, use hard shadow with slightly lower opacity
	// Full soft shadow implementation would require blur/gradient rendering
	tempCaster := *caster
	tempShadow := *caster.shadow
	tempShadow.Opacity *= 0.7
	tempCaster.shadow = &tempShadow
	s.renderHardShadow(target, &tempCaster, lightX, lightY)

	// TODO: Implement proper soft shadow with penumbra gradients
}

// renderContactShadow renders a ground contact shadow (entity touching ground).
func (s *ShadowSystem) renderContactShadow(target *ebiten.Image, caster *shadowCaster, lightX, lightY float64) {
	// Contact shadows are small, dark ellipses at the entity's feet
	// They don't extend far but provide ground contact visual cues

	// Shadow position slightly offset based on light direction
	dx := caster.x - lightX
	dy := caster.y - lightY
	dist := math.Sqrt(dx*dx + dy*dy)
	if dist < 0.1 {
		dist = 1.0
	}

	dirX := dx / dist
	dirY := dy / dist

	// Small offset from entity center
	offsetDist := caster.shadow.Height * 0.5
	shadowX := caster.x + dirX*offsetDist
	shadowY := caster.y + dirY*offsetDist + caster.shadow.Radius*0.7 // Offset down

	// Create small elliptical shadow
	shadowW := int(caster.shadow.Radius * 1.5)
	shadowH := int(caster.shadow.Radius * 0.5)
	if shadowW < 1 || shadowH < 1 {
		return
	}

	shadowImg := ebiten.NewImage(shadowW, shadowH)

	// Fill with gradient (darker in center)
	shadowImg.Fill(caster.shadow.Color)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-float64(shadowW)/2, -float64(shadowH)/2)
	opts.GeoM.Translate(shadowX, shadowY)
	opts.ColorScale.ScaleAlpha(float32(caster.shadow.Opacity))

	target.DrawImage(shadowImg, opts)
}

// RenderAmbientOcclusion renders ambient occlusion for entities.
// This creates subtle darkening at corners and contact points.
func (s *ShadowSystem) RenderAmbientOcclusion(screen *ebiten.Image) {
	if !s.enabled {
		return
	}

	entities := s.world.GetEntities()
	for _, entity := range entities {
		// Get AO component
		aoComp, hasAO := entity.GetComponent("ambient_occlusion")
		if !hasAO {
			continue
		}
		ao, ok := aoComp.(*AmbientOcclusionComponent)
		if !ok || !ao.Enabled {
			continue
		}

		// Get position
		posComp, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}
		pos, ok := posComp.(*PositionComponent)
		if !ok {
			continue
		}

		// Viewport culling
		if s.viewportSet {
			if pos.X+ao.Radius < s.cameraX ||
				pos.X-ao.Radius > s.cameraX+float64(s.viewportW) ||
				pos.Y+ao.Radius < s.cameraY ||
				pos.Y-ao.Radius > s.cameraY+float64(s.viewportH) {
				continue
			}
		}

		s.renderAOForEntity(screen, ao, pos)
	}
}

// renderAOForEntity renders ambient occlusion for a single entity.
func (s *ShadowSystem) renderAOForEntity(target *ebiten.Image, ao *AmbientOcclusionComponent, pos *PositionComponent) {
	// Simple AO: render a dark, soft-edged circle around the entity
	aoRadius := int(ao.Radius)
	if aoRadius < 1 {
		return
	}

	aoImg := ebiten.NewImage(aoRadius*2, aoRadius*2)

	// Fill with dark color
	aoColor := color.RGBA{0, 0, 0, uint8(ao.Intensity * 128)}
	aoImg.Fill(aoColor)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(-float64(aoRadius), -float64(aoRadius))
	opts.GeoM.Translate(pos.X, pos.Y)

	target.DrawImage(aoImg, opts)

	// Corner darkening: add small dark spots at cardinal directions
	if ao.CornerDarkening {
		cornerRadius := int(ao.Radius * 0.3)
		if cornerRadius < 1 {
			return
		}

		cornerImg := ebiten.NewImage(cornerRadius*2, cornerRadius*2)
		cornerColor := color.RGBA{0, 0, 0, uint8(ao.CornerAmount * 192)}
		cornerImg.Fill(cornerColor)

		// Four corners (cardinal directions)
		offsets := []struct{ x, y float64 }{
			{ao.Radius, 0},
			{-ao.Radius, 0},
			{0, ao.Radius},
			{0, -ao.Radius},
		}

		for _, offset := range offsets {
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(-float64(cornerRadius), -float64(cornerRadius))
			opts.GeoM.Translate(pos.X+offset.x, pos.Y+offset.y)
			target.DrawImage(cornerImg, opts)
		}
	}
}
