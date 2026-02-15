package engine

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// createTestEbitenImage creates a small ebiten.Image with a colored circle for testing.
func createTestEbitenImage(w, h int) *ebiten.Image {
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	cx, cy := w/2, h/2
	r := w / 3
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= r*r {
				rgba.SetRGBA(x, y, color.RGBA{R: 180, G: 120, B: 80, A: 255})
			}
		}
	}
	return ebiten.NewImageFromImage(rgba)
}

func TestNewSpriteFinalizerSystem(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteFinalizerSystem(world, 42)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
	if sys.seed != 42 {
		t.Errorf("expected seed 42, got %d", sys.seed)
	}
}

func TestSpriteFinalizerSystem_SkipsNilSprite(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteFinalizerSystem(world, 42)

	entity := NewEntity(1)
	// No sprite component — should not panic
	sys.Update([]*Entity{entity}, 0.016)
}

func TestSpriteFinalizerSystem_SkipsInvisibleSprite(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteFinalizerSystem(world, 42)

	entity := NewEntity(1)
	sprite := &EbitenSprite{
		Image:   createTestEbitenImage(16, 16),
		Visible: false,
		Width:   16,
		Height:  16,
	}
	entity.AddComponent(sprite)

	sys.Update([]*Entity{entity}, 0.016)
	if sprite.Finalized {
		t.Error("invisible sprite should not be finalized")
	}
}

func TestSpriteFinalizerSystem_FinalizesVisibleSprite(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteFinalizerSystem(world, 42)

	entity := NewEntity(1)
	sprite := &EbitenSprite{
		Image:   createTestEbitenImage(16, 16),
		Visible: true,
		Width:   16,
		Height:  16,
	}
	entity.AddComponent(sprite)

	sys.Update([]*Entity{entity}, 0.016)
	if !sprite.Finalized {
		t.Error("visible sprite should be finalized after Update")
	}
}

func TestSpriteFinalizerSystem_SkipsAlreadyFinalized(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteFinalizerSystem(world, 42)

	entity := NewEntity(1)
	sprite := &EbitenSprite{
		Image:     createTestEbitenImage(16, 16),
		Visible:   true,
		Width:     16,
		Height:    16,
		Finalized: true,
	}
	entity.AddComponent(sprite)

	originalImg := sprite.Image
	sys.Update([]*Entity{entity}, 0.016)

	// Image should not be replaced since it was already finalized
	if sprite.Image != originalImg {
		t.Error("already finalized sprite should not be re-processed")
	}
}

func TestSpriteFinalizerSystem_ReFinalizesAfterReset(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteFinalizerSystem(world, 42)

	entity := NewEntity(1)
	sprite := &EbitenSprite{
		Image:   createTestEbitenImage(16, 16),
		Visible: true,
		Width:   16,
		Height:  16,
	}
	entity.AddComponent(sprite)

	// First pass — finalizes
	sys.Update([]*Entity{entity}, 0.016)
	if !sprite.Finalized {
		t.Fatal("expected finalized after first pass")
	}

	// Simulate sprite regeneration
	sprite.Image = createTestEbitenImage(16, 16)
	sprite.Finalized = false

	// Second pass — should re-finalize
	sys.Update([]*Entity{entity}, 0.016)
	if !sprite.Finalized {
		t.Error("expected re-finalized after reset")
	}
}

func TestSpriteFinalizerSystem_ProcessesDirectionalSprites(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteFinalizerSystem(world, 42)

	entity := NewEntity(1)
	sprite := &EbitenSprite{
		Image:   createTestEbitenImage(16, 16),
		Visible: true,
		Width:   16,
		Height:  16,
		DirectionalImages: map[int]*ebiten.Image{
			0: createTestEbitenImage(16, 16),
			1: createTestEbitenImage(16, 16),
		},
	}
	entity.AddComponent(sprite)

	sys.Update([]*Entity{entity}, 0.016)
	if !sprite.Finalized {
		t.Error("expected finalized")
	}
}

func TestSpriteFinalizerSystem_GetProcessedCount(t *testing.T) {
	world := NewWorld()
	sys := NewSpriteFinalizerSystem(world, 42)

	entity := NewEntity(1)
	sprite := &EbitenSprite{
		Image:   createTestEbitenImage(16, 16),
		Visible: true,
		Width:   16,
		Height:  16,
	}
	entity.AddComponent(sprite)
	world.AddEntity(entity)

	sys.Update([]*Entity{entity}, 0.016)
	count := sys.GetProcessedCount()
	if count != 1 {
		t.Errorf("expected processed count 1, got %d", count)
	}
}
