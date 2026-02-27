//go:build !headless

// Package lighting provides dynamic lighting effects for rendered scenes.
// gpu_bloom.go provides GPU-accelerated bloom using Ebiten Kage shaders.
// This eliminates CPU-side pixel iteration that was causing 15-50ms overhead (V2 fix).
package lighting

import (
	"sync"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
	log "github.com/sirupsen/logrus"
)

// Kage shader sources for GPU-based bloom effects.

// bloomExtractShaderSrc extracts bright pixels above a threshold.
var bloomExtractShaderSrc = []byte(`
//kage:unit pixels

package main

var Threshold float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	clr := imageSrc0At(srcPos)
	
	// Calculate perceived brightness (luminance)
	brightness := 0.2126*clr.r + 0.7152*clr.g + 0.0722*clr.b
	
	// Keep pixel if above threshold, scale by excess brightness
	if brightness > Threshold {
		scale := (brightness - Threshold) / (1.0 - Threshold + 0.001)
		return vec4(clr.r * scale, clr.g * scale, clr.b * scale, clr.a)
	}
	
	return vec4(0.0, 0.0, 0.0, 0.0)
}
`)

// bloomBlurHShaderSrc applies horizontal Gaussian blur.
// Uses 9-tap Gaussian kernel for smooth results.
var bloomBlurHShaderSrc = []byte(`
//kage:unit pixels

package main

var Radius float
var ScreenWidth float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Gaussian weights for 9-tap kernel (sigma ~= radius/3)
	// Pre-computed: [0.0625, 0.1094, 0.1563, 0.1875, 0.1875, 0.1563, 0.1094, 0.0625]
	// Simplified to 5 weights with symmetry
	w0 := 0.227027  // center
	w1 := 0.194594  // +/- 1 step
	w2 := 0.121621  // +/- 2 steps
	w3 := 0.054054  // +/- 3 steps
	w4 := 0.016216  // +/- 4 steps
	
	step := Radius / 4.0
	
	// Sample with Gaussian weights
	result := imageSrc0At(srcPos) * w0
	
	result += imageSrc0At(vec2(srcPos.x + step*1.0, srcPos.y)) * w1
	result += imageSrc0At(vec2(srcPos.x - step*1.0, srcPos.y)) * w1
	
	result += imageSrc0At(vec2(srcPos.x + step*2.0, srcPos.y)) * w2
	result += imageSrc0At(vec2(srcPos.x - step*2.0, srcPos.y)) * w2
	
	result += imageSrc0At(vec2(srcPos.x + step*3.0, srcPos.y)) * w3
	result += imageSrc0At(vec2(srcPos.x - step*3.0, srcPos.y)) * w3
	
	result += imageSrc0At(vec2(srcPos.x + step*4.0, srcPos.y)) * w4
	result += imageSrc0At(vec2(srcPos.x - step*4.0, srcPos.y)) * w4
	
	return result
}
`)

// bloomBlurVShaderSrc applies vertical Gaussian blur.
var bloomBlurVShaderSrc = []byte(`
//kage:unit pixels

package main

var Radius float
var ScreenHeight float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Same Gaussian weights as horizontal pass
	w0 := 0.227027
	w1 := 0.194594
	w2 := 0.121621
	w3 := 0.054054
	w4 := 0.016216
	
	step := Radius / 4.0
	
	result := imageSrc0At(srcPos) * w0
	
	result += imageSrc0At(vec2(srcPos.x, srcPos.y + step*1.0)) * w1
	result += imageSrc0At(vec2(srcPos.x, srcPos.y - step*1.0)) * w1
	
	result += imageSrc0At(vec2(srcPos.x, srcPos.y + step*2.0)) * w2
	result += imageSrc0At(vec2(srcPos.x, srcPos.y - step*2.0)) * w2
	
	result += imageSrc0At(vec2(srcPos.x, srcPos.y + step*3.0)) * w3
	result += imageSrc0At(vec2(srcPos.x, srcPos.y - step*3.0)) * w3
	
	result += imageSrc0At(vec2(srcPos.x, srcPos.y + step*4.0)) * w4
	result += imageSrc0At(vec2(srcPos.x, srcPos.y - step*4.0)) * w4
	
	return result
}
`)

// bloomCompositeShaderSrc combines original with blurred bloom (additive).
var bloomCompositeShaderSrc = []byte(`
//kage:unit pixels

package main

var Intensity float

func Fragment(dstPos vec4, srcPos vec2, color vec4) vec4 {
	// Sample original (image 0) and bloom (image 1)
	original := imageSrc0At(srcPos)
	bloom := imageSrc1At(srcPos)
	
	// Additive blend with intensity control
	r := clamp(original.r + bloom.r * Intensity, 0.0, 1.0)
	g := clamp(original.g + bloom.g * Intensity, 0.0, 1.0)
	b := clamp(original.b + bloom.b * Intensity, 0.0, 1.0)
	
	return vec4(r, g, b, original.a)
}
`)

// GPUBloom provides GPU-accelerated bloom using Ebiten shaders.
// This eliminates CPU-side pixel iteration that was causing 15-50ms overhead.
type GPUBloom struct {
	config BloomConfig

	// Compiled shaders (lazily initialized)
	extractShader   *ebiten.Shader
	blurHShader     *ebiten.Shader
	blurVShader     *ebiten.Shader
	compositeShader *ebiten.Shader

	// Intermediate buffers (reused across frames)
	brightBuffer *ebiten.Image // Extracted bright pixels
	blurHBuffer  *ebiten.Image // Horizontal blur result
	blurVBuffer  *ebiten.Image // Vertical blur result (final bloom)

	// Buffer dimensions
	bufferWidth  int
	bufferHeight int

	// Metrics
	shaderCompilationErrors uint64 // Count of shader compilation failures for observability

	// Shader compilation mutex
	mu sync.Mutex
}

// NewGPUBloom creates a new GPU-accelerated bloom processor.
func NewGPUBloom() *GPUBloom {
	return &GPUBloom{
		config: DefaultBloomConfig(),
	}
}

// NewGPUBloomWithConfig creates a GPU bloom processor with custom configuration.
func NewGPUBloomWithConfig(config BloomConfig) *GPUBloom {
	return &GPUBloom{
		config: config,
	}
}

// SetConfig updates the bloom configuration.
func (b *GPUBloom) SetConfig(config BloomConfig) {
	b.config = config
}

// GetConfig returns the current bloom configuration.
func (b *GPUBloom) GetConfig() BloomConfig {
	return b.config
}

// ensureShaders compiles shaders if not already done. Thread-safe.
func (b *GPUBloom) ensureShaders() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	var err error

	if b.extractShader == nil {
		b.extractShader, err = ebiten.NewShader(bloomExtractShaderSrc)
		if err != nil {
			return err
		}
	}

	if b.blurHShader == nil {
		b.blurHShader, err = ebiten.NewShader(bloomBlurHShaderSrc)
		if err != nil {
			return err
		}
	}

	if b.blurVShader == nil {
		b.blurVShader, err = ebiten.NewShader(bloomBlurVShaderSrc)
		if err != nil {
			return err
		}
	}

	if b.compositeShader == nil {
		b.compositeShader, err = ebiten.NewShader(bloomCompositeShaderSrc)
		if err != nil {
			return err
		}
	}

	return nil
}

// ensureBuffers creates or resizes intermediate buffers if needed.
func (b *GPUBloom) ensureBuffers(width, height int) {
	if b.brightBuffer == nil || b.bufferWidth != width || b.bufferHeight != height {
		if b.brightBuffer != nil {
			b.brightBuffer.Dispose()
		}
		if b.blurHBuffer != nil {
			b.blurHBuffer.Dispose()
		}
		if b.blurVBuffer != nil {
			b.blurVBuffer.Dispose()
		}

		b.brightBuffer = ebiten.NewImage(width, height)
		b.blurHBuffer = ebiten.NewImage(width, height)
		b.blurVBuffer = ebiten.NewImage(width, height)
		b.bufferWidth = width
		b.bufferHeight = height
	}
}

// Apply applies GPU-accelerated bloom to the input image.
// The result is written directly to the dst image.
// If dst is nil, the bloom is composited back to the input image.
func (b *GPUBloom) Apply(input, dst *ebiten.Image) {
	if input == nil {
		return
	}

	if !b.config.Enabled || b.config.Intensity <= 0 {
		if dst != nil && dst != input {
			dst.DrawImage(input, nil)
		}
		return
	}

	// Ensure shaders are compiled
	if err := b.ensureShaders(); err != nil {
		atomic.AddUint64(&b.shaderCompilationErrors, 1)
		log.WithFields(log.Fields{
			"system":      "lighting",
			"subsys":      "gpu_bloom",
			"error":       err.Error(),
			"context":     "shader_compilation",
			"error_count": atomic.LoadUint64(&b.shaderCompilationErrors),
		}).Error("Shader compilation failed, falling back to passthrough")
		// Fallback: copy input to dst unchanged if shader compilation fails
		if dst != nil && dst != input {
			dst.DrawImage(input, nil)
		}
		return
	}

	bounds := input.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Ensure intermediate buffers exist
	b.ensureBuffers(width, height)

	// Step 1: Extract bright pixels
	b.brightBuffer.Clear()
	extractOpts := &ebiten.DrawRectShaderOptions{}
	extractOpts.Images[0] = input
	extractOpts.Uniforms = map[string]interface{}{
		"Threshold": float32(b.config.Threshold),
	}
	b.brightBuffer.DrawRectShader(width, height, b.extractShader, extractOpts)

	// Step 2: Horizontal blur
	b.blurHBuffer.Clear()
	blurHOpts := &ebiten.DrawRectShaderOptions{}
	blurHOpts.Images[0] = b.brightBuffer
	blurHOpts.Uniforms = map[string]interface{}{
		"Radius":      float32(b.config.Radius),
		"ScreenWidth": float32(width),
	}
	b.blurHBuffer.DrawRectShader(width, height, b.blurHShader, blurHOpts)

	// Step 3: Vertical blur
	b.blurVBuffer.Clear()
	blurVOpts := &ebiten.DrawRectShaderOptions{}
	blurVOpts.Images[0] = b.blurHBuffer
	blurVOpts.Uniforms = map[string]interface{}{
		"Radius":       float32(b.config.Radius),
		"ScreenHeight": float32(height),
	}
	b.blurVBuffer.DrawRectShader(width, height, b.blurVShader, blurVOpts)

	// Step 4: Composite bloom with original
	target := dst
	if target == nil {
		target = input
	}

	// If compositing to the same buffer, we need to handle it carefully
	if target == input {
		// Clear and composite in place
		compositeOpts := &ebiten.DrawRectShaderOptions{}
		compositeOpts.Images[0] = input
		compositeOpts.Images[1] = b.blurVBuffer
		compositeOpts.Uniforms = map[string]interface{}{
			"Intensity": float32(b.config.Intensity),
		}

		// Use blurHBuffer as temp (it's not needed anymore)
		b.blurHBuffer.Clear()
		b.blurHBuffer.DrawRectShader(width, height, b.compositeShader, compositeOpts)

		// Copy result back to input
		input.Clear()
		input.DrawImage(b.blurHBuffer, nil)
	} else {
		target.Clear()
		compositeOpts := &ebiten.DrawRectShaderOptions{}
		compositeOpts.Images[0] = input
		compositeOpts.Images[1] = b.blurVBuffer
		compositeOpts.Uniforms = map[string]interface{}{
			"Intensity": float32(b.config.Intensity),
		}
		target.DrawRectShader(width, height, b.compositeShader, compositeOpts)
	}
}

// ApplyToBuffer applies bloom directly to the provided buffer (in-place).
// This is optimized for the lighting system use case where we modify lightingBuffer.
func (b *GPUBloom) ApplyToBuffer(buffer *ebiten.Image) {
	b.Apply(buffer, nil)
}

// Dispose releases GPU resources held by the bloom processor.
// Call this when the processor is no longer needed.
func (b *GPUBloom) Dispose() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.brightBuffer != nil {
		b.brightBuffer.Dispose()
		b.brightBuffer = nil
	}
	if b.blurHBuffer != nil {
		b.blurHBuffer.Dispose()
		b.blurHBuffer = nil
	}
	if b.blurVBuffer != nil {
		b.blurVBuffer.Dispose()
		b.blurVBuffer = nil
	}
	// Note: Shaders don't have Dispose() in Ebiten - they're managed by runtime
}

// GetShaderCompilationErrors returns the total number of shader compilation failures.
// This metric is useful for monitoring graphics driver compatibility issues.
// Thread-safe via atomic operations.
func (b *GPUBloom) GetShaderCompilationErrors() uint64 {
	return atomic.LoadUint64(&b.shaderCompilationErrors)
}
