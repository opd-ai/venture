package engine

import (
	"testing"
)

// BenchmarkGetSpriteComponent benchmarks sprite component access in animation system.
// Tests the optimization of using cached GetSprite() vs generic GetComponent + type assertion.
func BenchmarkGetSpriteComponent(b *testing.B) {
	entity := NewEntity(1)
	sprite := NewSpriteComponent(32, 32, nil)
	entity.AddComponent(sprite)

	animSys := NewAnimationSystem(nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = animSys.getSpriteComponent(entity)
	}
}

// BenchmarkGetSpriteCached benchmarks the cached sprite getter for comparison.
func BenchmarkGetSpriteCached(b *testing.B) {
	entity := NewEntity(1)
	sprite := NewSpriteComponent(32, 32, nil)
	entity.AddComponent(sprite)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = entity.GetSprite()
	}
}
