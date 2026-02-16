//go:build headless

// Package postprocess provides post-processing effects for rendered scenes.
// gpu_processor_headless.go provides stub GPUProcessor for headless (no display) environments.
package postprocess

// GPUProcessor is a stub for headless builds where GPU shaders are unavailable.
type GPUProcessor struct {
	config Config
}

// NewGPUProcessor creates a stub GPU processor for headless environments.
func NewGPUProcessor() *GPUProcessor {
	return &GPUProcessor{config: DefaultConfig()}
}

// NewGPUProcessorWithConfig creates a stub GPU processor with custom configuration.
func NewGPUProcessorWithConfig(config Config) *GPUProcessor {
	return &GPUProcessor{config: config}
}

// SetConfig sets the post-processing configuration.
func (p *GPUProcessor) SetConfig(config Config) {
	p.config = config
}

// GetConfig returns the current configuration.
func (p *GPUProcessor) GetConfig() Config {
	return p.config
}

// PrecompileShaders is a no-op in headless mode.
func (p *GPUProcessor) PrecompileShaders() error {
	return nil
}

// Dispose is a no-op in headless mode.
func (p *GPUProcessor) Dispose() {}
