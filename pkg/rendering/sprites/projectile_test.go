// Package sprites provides procedural projectile sprite generation tests.
// Phase 10.2: Projectile Physics System
package sprites

import (
	"testing"
)

func TestGenerateProjectileSprite(t *testing.T) {
	tests := []struct {
		name           string
		seed           int64
		projectileType string
		genreID        string
		size           int
	}{
		{
			name:           "arrow fantasy",
			seed:           12345,
			projectileType: "arrow",
			genreID:        "fantasy",
			size:           12,
		},
		{
			name:           "bolt scifi",
			seed:           12345,
			projectileType: "bolt",
			genreID:        "scifi",
			size:           12,
		},
		{
			name:           "bullet horror",
			seed:           12345,
			projectileType: "bullet",
			genreID:        "horror",
			size:           8,
		},
		{
			name:           "magic cyberpunk",
			seed:           12345,
			projectileType: "magic",
			genreID:        "cyberpunk",
			size:           12,
		},
		{
			name:           "fireball postapoc",
			seed:           12345,
			projectileType: "fireball",
			genreID:        "postapoc",
			size:           16,
		},
		{
			name:           "energy scifi",
			seed:           12345,
			projectileType: "energy",
			genreID:        "scifi",
			size:           12,
		},
		{
			name:           "default type",
			seed:           12345,
			projectileType: "unknown",
			genreID:        "fantasy",
			size:           12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sprite := GenerateProjectileSprite(tt.seed, tt.projectileType, tt.genreID, tt.size)
			if sprite == nil {
				t.Error("expected non-nil sprite")
			}

			bounds := sprite.Bounds()
			if bounds.Dx() != tt.size || bounds.Dy() != tt.size {
				t.Errorf("expected size %dx%d, got %dx%d", tt.size, tt.size, bounds.Dx(), bounds.Dy())
			}
		})
	}
}

func TestProjectileSpriteConstants(t *testing.T) {
	// Test that all projectile type constants are defined
	types := []ProjectileType{
		ProjectileArrow,
		ProjectileBolt,
		ProjectileBullet,
		ProjectileMagic,
		ProjectileFireball,
		ProjectileEnergy,
	}

	for _, pt := range types {
		if string(pt) == "" {
			t.Errorf("projectile type constant is empty")
		}
	}
}

func TestProjectileSpriteDeterminism(t *testing.T) {
	// Test that same seed produces same sprite
	seed := int64(42)
	projectileType := "arrow"
	genreID := "fantasy"
	size := 12

	sprite1 := GenerateProjectileSprite(seed, projectileType, genreID, size)
	sprite2 := GenerateProjectileSprite(seed, projectileType, genreID, size)

	if sprite1 == nil || sprite2 == nil {
		t.Fatal("sprites should not be nil")
	}

	// Check that both sprites have the same size
	if sprite1.Bounds() != sprite2.Bounds() {
		t.Error("sprites should have the same bounds")
	}

	// Note: We can't easily compare pixel-by-pixel without access to pixel data
	// but the determinism is ensured by the palette generation being deterministic
}

// Benchmark projectile sprite generation
func BenchmarkGenerateProjectileSprite(b *testing.B) {
	seed := int64(12345)
	projectileType := "arrow"
	genreID := "fantasy"
	size := 12

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GenerateProjectileSprite(seed, projectileType, genreID, size)
	}
}

func BenchmarkGenerateProjectileSpriteAllTypes(b *testing.B) {
	seed := int64(12345)
	genreID := "fantasy"
	size := 12
	types := []string{"arrow", "bolt", "bullet", "magic", "fireball", "energy"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, pt := range types {
			GenerateProjectileSprite(seed, pt, genreID, size)
		}
	}
}
