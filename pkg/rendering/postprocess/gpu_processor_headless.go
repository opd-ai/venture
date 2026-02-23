//go:build headless

// Package postprocess provides post-processing effects for rendered scenes.
//
// gpu_processor_headless.go provides a stub GPUProcessor implementation for
// headless (no display) environments. This file is compiled instead of
// gpu_processor.go when the "headless" build tag is specified.
//
// # When to Use the Headless Build Tag
//
// The headless build tag should be used in the following scenarios:
//
//   - CI/CD pipelines: Server-based testing environments often lack a GPU
//     or display server (X11/Wayland). Running tests with the headless tag
//     avoids Ebiten shader compilation failures.
//
//   - Server-side code: The game server does not render graphics. Building
//     with -tags=headless prevents unnecessary GPU dependency checks.
//
//   - Automated testing: Unit tests that don't need actual post-processing
//     output can use the headless stub for faster, more reliable tests.
//
// # Build Examples
//
// Build with headless stub:
//
//	go build -tags=headless ./...
//
// Run tests with headless stub:
//
//	go test -tags=headless ./pkg/rendering/postprocess/...
//
// # Implications
//
//   - All post-processing effects become no-ops in headless mode
//   - ApplyAll returns the input image unchanged
//   - PrecompileShaders does nothing (no GPU shaders to compile)
//   - Configuration is still stored but has no visual effect
//   - Test coverage metrics will reflect the stub implementation
//
// This pattern is consistent with other Ebiten-dependent packages in the
// codebase (pkg/rendering/lighting/gpu_bloom_headless.go, pkg/engine headless
// stubs) which use the same build tag strategy for CI compatibility.
package postprocess

import "github.com/hajimehoshi/ebiten/v2"

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

// ApplyAll is a no-op in headless mode; returns input unchanged.
func (p *GPUProcessor) ApplyAll(input *ebiten.Image) *ebiten.Image {
	return input
}

// PrecompileShaders is a no-op in headless mode.
func (p *GPUProcessor) PrecompileShaders() error {
	return nil
}

// Dispose is a no-op in headless mode.
func (p *GPUProcessor) Dispose() {}
