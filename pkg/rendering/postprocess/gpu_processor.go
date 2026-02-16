//go:build !headless

// Package postprocess provides post-processing effects for rendered scenes.
// gpu_processor.go provides GPU-accelerated post-processing using Ebiten shaders.
package postprocess

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/sirupsen/logrus"
)

// Kage shader sources for GPU-based post-processing effects.
// These are compiled at runtime by Ebiten.

// vignetteShaderSrc applies darkening at screen edges.
var vignetteShaderSrc = []byte(`
//kage:unit pixels

package main

var Intensity float
var Softness float
var InnerRadius float
var OuterRadius float
var VignetteR float
var VignetteG float
var VignetteB float
var ScreenWidth float
var ScreenHeight float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Sample the source texture
	clr := imageSrc0At(srcPos)
	
	// Calculate center and normalize coordinates
	centerX := ScreenWidth / 2.0
	centerY := ScreenHeight / 2.0
	maxRadius := sqrt(centerX*centerX + centerY*centerY)
	
	// Distance from center (normalized 0-1)
	dx := srcPos.x - centerX
	dy := srcPos.y - centerY
	dist := sqrt(dx*dx + dy*dy) / maxRadius
	
	// Calculate vignette strength
	vignetteStrength := 0.0
	if dist > InnerRadius {
		t := (dist - InnerRadius) / (OuterRadius - InnerRadius)
		t = clamp(t, 0.0, 1.0)
		
		// Smoothstep for soft edges
		if Softness > 0 {
			t = t * t * (3.0 - 2.0 * t)
		}
		vignetteStrength = t * Intensity
	}
	
	// Blend with vignette color
	r := mix(clr.r, VignetteR, vignetteStrength)
	g := mix(clr.g, VignetteG, vignetteStrength)
	b := mix(clr.b, VignetteB, vignetteStrength)
	
	return vec4(r, g, b, clr.a)
}
`)

// colorGradingShaderSrc applies saturation, contrast, brightness, temperature, tint.
var colorGradingShaderSrc = []byte(`
//kage:unit pixels

package main

var Saturation float
var Contrast float
var Brightness float
var Temperature float
var Tint float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	clr := imageSrc0At(srcPos)
	
	r := clr.r
	g := clr.g
	b := clr.b
	
	// Brightness adjustment
	if Brightness != 0.0 {
		r = clamp(r + Brightness, 0.0, 1.0)
		g = clamp(g + Brightness, 0.0, 1.0)
		b = clamp(b + Brightness, 0.0, 1.0)
	}
	
	// Contrast adjustment
	if Contrast != 1.0 {
		r = clamp((r - 0.5) * Contrast + 0.5, 0.0, 1.0)
		g = clamp((g - 0.5) * Contrast + 0.5, 0.0, 1.0)
		b = clamp((b - 0.5) * Contrast + 0.5, 0.0, 1.0)
	}
	
	// Saturation adjustment
	if Saturation != 1.0 {
		lum := 0.299*r + 0.587*g + 0.114*b
		r = clamp(mix(lum, r, Saturation), 0.0, 1.0)
		g = clamp(mix(lum, g, Saturation), 0.0, 1.0)
		b = clamp(mix(lum, b, Saturation), 0.0, 1.0)
	}
	
	// Temperature adjustment (cool = blue, warm = orange)
	if Temperature < 0 {
		strength := -Temperature
		r = clamp(r - strength*0.2, 0.0, 1.0)
		b = clamp(b + strength*0.3, 0.0, 1.0)
	} else if Temperature > 0 {
		strength := Temperature
		r = clamp(r + strength*0.3, 0.0, 1.0)
		g = clamp(g + strength*0.15, 0.0, 1.0)
		b = clamp(b - strength*0.2, 0.0, 1.0)
	}
	
	// Tint adjustment (green vs magenta)
	if Tint < 0 {
		strength := -Tint
		r = clamp(r - strength*0.1, 0.0, 1.0)
		g = clamp(g + strength*0.2, 0.0, 1.0)
		b = clamp(b - strength*0.1, 0.0, 1.0)
	} else if Tint > 0 {
		strength := Tint
		r = clamp(r + strength*0.2, 0.0, 1.0)
		g = clamp(g - strength*0.15, 0.0, 1.0)
		b = clamp(b + strength*0.2, 0.0, 1.0)
	}
	
	return vec4(r, g, b, clr.a)
}
`)

// chromaticAberrationShaderSrc separates color channels at screen edges.
var chromaticAberrationShaderSrc = []byte(`
//kage:unit pixels

package main

var Intensity float
var DirectionX float
var DirectionY float
var ScreenWidth float
var ScreenHeight float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Calculate center
	centerX := ScreenWidth / 2.0
	centerY := ScreenHeight / 2.0
	maxRadius := sqrt(centerX*centerX + centerY*centerY)
	
	// Distance from center (normalized 0-1)
	dx := srcPos.x - centerX
	dy := srcPos.y - centerY
	dist := sqrt(dx*dx + dy*dy) / maxRadius
	
	// Aberration strength increases with distance from center
	// Scale factor of 5.0 matches original chromaticAberrationScale
	aberrationStrength := dist * Intensity * 5.0
	
	// Red channel - shift in direction
	rx := srcPos.x + DirectionX * aberrationStrength
	ry := srcPos.y + DirectionY * aberrationStrength
	rc := imageSrc0At(vec2(rx, ry))
	
	// Green channel - center (no offset)
	gc := imageSrc0At(srcPos)
	
	// Blue channel - shift opposite direction
	bx := srcPos.x - DirectionX * aberrationStrength
	by := srcPos.y - DirectionY * aberrationStrength
	bc := imageSrc0At(vec2(bx, by))
	
	return vec4(rc.r, gc.g, bc.b, gc.a)
}
`)

// GPUProcessor provides GPU-accelerated post-processing using Ebiten shaders.
// This eliminates CPU-side pixel iteration that was causing 40-120ms overhead.
type GPUProcessor struct {
	config Config

	// Compiled shaders (lazily initialized)
	vignetteShader            *ebiten.Shader
	colorGradingShader        *ebiten.Shader
	chromaticAberrationShader *ebiten.Shader

	// Reusable output buffer (avoids per-frame GPU texture allocation)
	outputBuffer *ebiten.Image
	bufferWidth  int
	bufferHeight int

	// Shader compilation mutex
	mu sync.Mutex

	// Optional logger for error reporting
	logger *logrus.Logger
}

// NewGPUProcessor creates a new GPU-accelerated post-processor.
func NewGPUProcessor() *GPUProcessor {
	return &GPUProcessor{
		config: DefaultConfig(),
	}
}

// NewGPUProcessorWithConfig creates a GPU processor with custom configuration.
func NewGPUProcessorWithConfig(config Config) *GPUProcessor {
	return &GPUProcessor{
		config: config,
	}
}

// NewGPUProcessorWithLogger creates a GPU processor with custom configuration and logger.
// The logger is used to report shader compilation errors for debugging.
func NewGPUProcessorWithLogger(config Config, logger *logrus.Logger) *GPUProcessor {
	return &GPUProcessor{
		config: config,
		logger: logger,
	}
}

// SetConfig sets the post-processing configuration.
func (p *GPUProcessor) SetConfig(config Config) {
	p.config = config
}

// GetConfig returns the current configuration.
func (p *GPUProcessor) GetConfig() Config {
	return p.config
}

// PrecompileShaders eagerly compiles all GPU shaders during initialization.
// Call this during startup (e.g., loading screen) to avoid a 50-200ms stutter
// when post-processing effects are first activated during gameplay.
// Thread-safe; subsequent calls are no-ops if shaders are already compiled.
func (p *GPUProcessor) PrecompileShaders() error {
	return p.ensureShaders()
}

// ensureShaders compiles shaders if not already done. Thread-safe.
func (p *GPUProcessor) ensureShaders() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var err error

	if p.vignetteShader == nil {
		p.vignetteShader, err = ebiten.NewShader(vignetteShaderSrc)
		if err != nil {
			if p.logger != nil {
				p.logger.WithFields(logrus.Fields{
					"component": "GPUProcessor",
					"shader":    "vignette",
					"error":     err.Error(),
				}).Error("failed to compile vignette shader")
			}
			return err
		}
	}

	if p.colorGradingShader == nil {
		p.colorGradingShader, err = ebiten.NewShader(colorGradingShaderSrc)
		if err != nil {
			if p.logger != nil {
				p.logger.WithFields(logrus.Fields{
					"component": "GPUProcessor",
					"shader":    "color_grading",
					"error":     err.Error(),
				}).Error("failed to compile color grading shader")
			}
			return err
		}
	}

	if p.chromaticAberrationShader == nil {
		p.chromaticAberrationShader, err = ebiten.NewShader(chromaticAberrationShaderSrc)
		if err != nil {
			if p.logger != nil {
				p.logger.WithFields(logrus.Fields{
					"component": "GPUProcessor",
					"shader":    "chromatic_aberration",
					"error":     err.Error(),
				}).Error("failed to compile chromatic aberration shader")
			}
			return err
		}
	}

	return nil
}

// ensureBuffer creates or resizes the output buffer if needed.
func (p *GPUProcessor) ensureBuffer(width, height int) {
	if p.outputBuffer == nil || p.bufferWidth != width || p.bufferHeight != height {
		if p.outputBuffer != nil {
			p.outputBuffer.Dispose()
		}
		p.outputBuffer = ebiten.NewImage(width, height)
		p.bufferWidth = width
		p.bufferHeight = height
	}
}

// ApplyAll applies all enabled effects using GPU shaders.
// This is the GPU-optimized replacement for the CPU-based Processor.ApplyAll.
// Returns the processed image. Input image is not modified.
// If shader compilation fails, the input image is returned unchanged and
// the error is logged if a logger was provided.
func (p *GPUProcessor) ApplyAll(input *ebiten.Image) *ebiten.Image {
	if input == nil {
		return nil
	}
	if !p.hasEnabledEffects() {
		return input
	}
	if err := p.ensureShaders(); err != nil {
		if p.logger != nil {
			p.logger.WithFields(logrus.Fields{
				"component": "GPUProcessor",
				"error":     err.Error(),
			}).Warn("shader compilation failed, returning unprocessed image")
		}
		return input
	}
	return p.applyEnabledEffects(input)
}

// hasEnabledEffects checks if any post-processing effects are enabled.
func (p *GPUProcessor) hasEnabledEffects() bool {
	hasColorGrading := p.config.ColorGrading.Enabled && !isNeutralColorGrading(p.config.ColorGrading)
	hasVignette := p.config.Vignette.Enabled && p.config.Vignette.Intensity > 0
	hasChromatic := p.config.ChromaticAberration.Enabled && p.config.ChromaticAberration.Intensity > 0
	return hasColorGrading || hasVignette || hasChromatic
}

// applyEnabledEffects applies all enabled effects in sequence to the input image.
func (p *GPUProcessor) applyEnabledEffects(input *ebiten.Image) *ebiten.Image {
	bounds := input.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	p.ensureBuffer(width, height)

	currentSrc := input
	needsCopy := false

	currentSrc, needsCopy = p.applyColorGradingIfEnabled(currentSrc, width, height, needsCopy)
	currentSrc, needsCopy = p.applyVignetteIfEnabled(currentSrc, width, height, needsCopy)
	currentSrc = p.applyChromaticAberrationIfEnabled(currentSrc, width, height, needsCopy)

	if currentSrc == p.outputBuffer {
		return p.outputBuffer
	}
	return input
}

// applyColorGradingIfEnabled applies color grading effect if enabled.
func (p *GPUProcessor) applyColorGradingIfEnabled(src *ebiten.Image, width, height int, needsCopy bool) (*ebiten.Image, bool) {
	hasColorGrading := p.config.ColorGrading.Enabled && !isNeutralColorGrading(p.config.ColorGrading)
	if !hasColorGrading {
		return src, needsCopy
	}
	p.outputBuffer.Clear()
	p.applyColorGrading(p.outputBuffer, src, width, height)
	return p.outputBuffer, true
}

// applyVignetteIfEnabled applies vignette effect if enabled.
func (p *GPUProcessor) applyVignetteIfEnabled(src *ebiten.Image, width, height int, needsCopy bool) (*ebiten.Image, bool) {
	hasVignette := p.config.Vignette.Enabled && p.config.Vignette.Intensity > 0
	if !hasVignette {
		return src, needsCopy
	}
	if needsCopy && src == p.outputBuffer {
		p.applyVignetteInPlace(p.outputBuffer, width, height)
		return p.outputBuffer, true
	}
	p.outputBuffer.Clear()
	p.applyVignette(p.outputBuffer, src, width, height)
	return p.outputBuffer, true
}

// applyChromaticAberrationIfEnabled applies chromatic aberration effect if enabled.
func (p *GPUProcessor) applyChromaticAberrationIfEnabled(src *ebiten.Image, width, height int, needsCopy bool) *ebiten.Image {
	hasChromatic := p.config.ChromaticAberration.Enabled && p.config.ChromaticAberration.Intensity > 0
	if !hasChromatic {
		return src
	}
	if needsCopy && src == p.outputBuffer {
		p.applyChromaticAberrationInPlace(p.outputBuffer, width, height)
		return p.outputBuffer
	}
	p.outputBuffer.Clear()
	p.applyChromaticAberration(p.outputBuffer, src, width, height)
	return p.outputBuffer
}

// isNeutralColorGrading checks if color grading settings are neutral (no effect).
func isNeutralColorGrading(cfg ColorGradingConfig) bool {
	return cfg.Saturation == 1.0 &&
		cfg.Contrast == 1.0 &&
		cfg.Brightness == 0.0 &&
		cfg.Temperature == 0.0 &&
		cfg.Tint == 0.0
}

// applyColorGrading applies color grading shader to dst from src.
func (p *GPUProcessor) applyColorGrading(dst, src *ebiten.Image, width, height int) {
	cfg := p.config.ColorGrading

	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = src
	opts.Uniforms = map[string]interface{}{
		"Saturation":  float32(cfg.Saturation),
		"Contrast":    float32(cfg.Contrast),
		"Brightness":  float32(cfg.Brightness),
		"Temperature": float32(cfg.Temperature),
		"Tint":        float32(cfg.Tint),
	}

	dst.DrawRectShader(width, height, p.colorGradingShader, opts)
}

// applyVignette applies vignette shader to dst from src.
func (p *GPUProcessor) applyVignette(dst, src *ebiten.Image, width, height int) {
	cfg := p.config.Vignette

	// Get vignette color (default black)
	vr, vg, vb := 0.0, 0.0, 0.0
	if cfg.Color != nil {
		r, g, b, _ := cfg.Color.RGBA()
		vr = float64(r) / 65535.0
		vg = float64(g) / 65535.0
		vb = float64(b) / 65535.0
	}

	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = src
	opts.Uniforms = map[string]interface{}{
		"Intensity":    float32(cfg.Intensity),
		"Softness":     float32(cfg.Softness),
		"InnerRadius":  float32(cfg.InnerRadius),
		"OuterRadius":  float32(cfg.OuterRadius),
		"VignetteR":    float32(vr),
		"VignetteG":    float32(vg),
		"VignetteB":    float32(vb),
		"ScreenWidth":  float32(width),
		"ScreenHeight": float32(height),
	}

	dst.DrawRectShader(width, height, p.vignetteShader, opts)
}

// applyVignetteInPlace re-applies vignette to the buffer (ping-pong workaround).
// Creates a temporary copy, clears buffer, then applies shader.
func (p *GPUProcessor) applyVignetteInPlace(buffer *ebiten.Image, width, height int) {
	// Create temporary image for ping-pong
	temp := ebiten.NewImage(width, height)
	defer temp.Dispose()

	// Copy current buffer to temp
	temp.DrawImage(buffer, nil)

	// Clear and apply vignette from temp
	buffer.Clear()
	p.applyVignette(buffer, temp, width, height)
}

// applyChromaticAberration applies chromatic aberration shader to dst from src.
func (p *GPUProcessor) applyChromaticAberration(dst, src *ebiten.Image, width, height int) {
	cfg := p.config.ChromaticAberration

	opts := &ebiten.DrawRectShaderOptions{}
	opts.Images[0] = src
	opts.Uniforms = map[string]interface{}{
		"Intensity":    float32(cfg.Intensity),
		"DirectionX":   float32(cfg.DirectionX),
		"DirectionY":   float32(cfg.DirectionY),
		"ScreenWidth":  float32(width),
		"ScreenHeight": float32(height),
	}

	dst.DrawRectShader(width, height, p.chromaticAberrationShader, opts)
}

// applyChromaticAberrationInPlace applies chromatic aberration with ping-pong.
func (p *GPUProcessor) applyChromaticAberrationInPlace(buffer *ebiten.Image, width, height int) {
	temp := ebiten.NewImage(width, height)
	defer temp.Dispose()

	temp.DrawImage(buffer, nil)
	buffer.Clear()
	p.applyChromaticAberration(buffer, temp, width, height)
}

// Dispose releases GPU resources held by the processor.
// Call this when the processor is no longer needed.
func (p *GPUProcessor) Dispose() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.outputBuffer != nil {
		p.outputBuffer.Dispose()
		p.outputBuffer = nil
	}
	// Note: Shaders don't have Dispose() in Ebiten - they're managed by runtime
}
