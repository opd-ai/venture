package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/parallel"
	"github.com/opd-ai/venture/pkg/rendering/pool"
)

// ImagePoolAdapter adapts pkg/rendering/pool.ImagePool to ImagePoolProvider interface.
type ImagePoolAdapter struct {
	pool *pool.ImagePool
}

// NewImagePoolAdapter creates an adapter for the image pool.
func NewImagePoolAdapter(p *pool.ImagePool) *ImagePoolAdapter {
	return &ImagePoolAdapter{pool: p}
}

// GetImage retrieves an image from the pool.
func (a *ImagePoolAdapter) GetImage(width, height int) *ebiten.Image {
	return a.pool.GetImage(width, height)
}

// PutImage returns an image to the pool.
func (a *ImagePoolAdapter) PutImage(img *ebiten.Image) {
	a.pool.PutImage(img)
}

// ParallelRendererAdapter adapts pkg/rendering/parallel.WorkerPool to ParallelRendererProvider interface.
type ParallelRendererAdapter struct {
	pool *parallel.WorkerPool
}

// NewParallelRendererAdapter creates an adapter for the parallel renderer.
func NewParallelRendererAdapter(p *parallel.WorkerPool) *ParallelRendererAdapter {
	return &ParallelRendererAdapter{pool: p}
}

// Start launches the worker goroutines.
func (a *ParallelRendererAdapter) Start() {
	a.pool.Start()
}

// Stop gracefully shuts down the worker pool.
func (a *ParallelRendererAdapter) Stop() {
	a.pool.Stop()
}

// IsRunning returns true if the worker pool is running.
func (a *ParallelRendererAdapter) IsRunning() bool {
	return a.pool.IsRunning()
}
