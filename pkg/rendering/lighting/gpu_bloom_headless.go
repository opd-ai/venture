//go:build headless

// Package lighting provides dynamic lighting effects for rendered scenes.
// gpu_bloom_headless.go provides stub GPUBloom for headless (no display) environments.
package lighting

import "github.com/hajimehoshi/ebiten/v2"

// GPUBloom is a stub for headless builds where GPU shaders are unavailable.
type GPUBloom struct {
	config BloomConfig
}

// NewGPUBloom creates a stub GPU bloom processor for headless environments.
func NewGPUBloom() *GPUBloom {
	return &GPUBloom{config: DefaultBloomConfig()}
}

// NewGPUBloomWithConfig creates a stub GPU bloom processor with custom configuration.
func NewGPUBloomWithConfig(config BloomConfig) *GPUBloom {
	return &GPUBloom{config: config}
}

// SetConfig updates the bloom configuration.
func (b *GPUBloom) SetConfig(config BloomConfig) {
	b.config = config
}

// GetConfig returns the current bloom configuration.
func (b *GPUBloom) GetConfig() BloomConfig {
	return b.config
}

// ApplyToBuffer is a no-op in headless mode.
func (b *GPUBloom) ApplyToBuffer(buffer *ebiten.Image) {}

// Dispose is a no-op in headless mode.
func (b *GPUBloom) Dispose() {}
