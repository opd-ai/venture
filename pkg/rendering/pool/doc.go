// Package pool provides object pooling for frequently allocated rendering resources.
// The pool package implements sync.Pool-based pooling for Ebiten images and other
// rendering resources. This reduces allocation pressure and improves garbage collection
// performance by reusing objects instead of creating new ones.
//
// Key features:
//   - Multiple pools for common sprite sizes (28x28, 32x32, 64x64)
//   - Automatic size-based pool selection
//   - Zero-allocation retrieval from pools
//   - Automatic cleanup and reset on return
//   - Thread-safe operations via sync.Pool
//
// Usage:
//
//	// Get an image from the pool
//	img := pool.GetImage(32, 32)
//	defer pool.PutImage(img)
//
//	// Use the image
//	// ... draw operations ...
//
//	// Image is automatically returned to pool on defer
//
// Performance considerations:
//   - Always return pooled objects when done
//   - Prefer pooled sizes (28x28, 32x32, 64x64) for best performance
//   - Non-standard sizes create new images (not pooled)
//   - Pool automatically grows/shrinks based on demand
package pool
