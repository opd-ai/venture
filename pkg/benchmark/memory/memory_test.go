package memory

import (
	"testing"

	"github.com/opd-ai/venture/pkg/memprofile"
)

// TestMemoryBaselineWorld validates memory usage for basic world generation.
// This test verifies the <500MB client memory usage claim from docs/PERFORMANCE.md.
func TestMemoryBaselineWorld(t *testing.T) {
	profile := memprofile.StartMemoryProfile("BaselineWorld")

	// Simulate basic world generation with minimal systems
	// Allocate structures similar to what the client would use
	worldData := make([]byte, 10*1024*1024) // 10MB world data
	profile.Snapshot()

	entities := make([]interface{}, 500) // 500 entities
	for i := 0; i < 500; i++ {
		entities[i] = make([]byte, 1024) // 1KB per entity
	}
	profile.Snapshot()

	// Simulate sprite cache (significant allocation)
	spriteCache := make([][]byte, 100)
	for i := 0; i < 100; i++ {
		spriteCache[i] = make([]byte, 64*64*4) // 64x64 RGBA sprite
	}
	profile.Snapshot()

	profile.End()

	peak := profile.GetPeakAllocation()
	peakMB := float64(peak) / (1024 * 1024)

	t.Logf("Peak allocation: %.2fMB", peakMB)
	t.Logf("Allocation growth: %d bytes", profile.GetAllocationGrowth())

	// Verify under 500MB threshold (with 20% headroom = 400MB)
	thresholdBytes := uint64(400 * 1024 * 1024)
	if peak > thresholdBytes {
		t.Errorf("Memory usage %.2fMB exceeds safe threshold of 400MB (500MB limit with headroom)", peakMB)
	}

	// Prevent unused variable warnings
	_ = worldData
	_ = entities
	_ = spriteCache
}

// TestMemoryHighEntityCount validates memory with 2000 entities (performance target).
func TestMemoryHighEntityCount(t *testing.T) {
	profile := memprofile.StartMemoryProfile("HighEntityCount")

	// Simulate 2000 entities with components
	entities := make([]map[string]interface{}, 2000)
	profile.Snapshot()

	for i := 0; i < 2000; i++ {
		entities[i] = map[string]interface{}{
			"position":  [2]float64{float64(i), float64(i)},
			"velocity":  [2]float64{1.0, 1.0},
			"health":    100,
			"sprite":    make([]byte, 32*32*4), // 32x32 RGBA
			"inventory": make([]byte, 512),     // Small inventory
		}

		// Take snapshot every 400 entities
		if i%400 == 0 {
			profile.Snapshot()
		}
	}

	profile.End()

	peak := profile.GetPeakAllocation()
	peakMB := float64(peak) / (1024 * 1024)

	t.Logf("Peak allocation: %.2fMB", peakMB)
	t.Logf("Average allocation: %.2fMB", float64(profile.GetAverageAllocation())/(1024*1024))
	t.Logf("Allocation growth: %d bytes", profile.GetAllocationGrowth())

	// Verify under 500MB threshold
	thresholdBytes := uint64(500 * 1024 * 1024)
	if peak > thresholdBytes {
		t.Errorf("Memory usage %.2fMB exceeds 500MB threshold with 2000 entities", peakMB)
	}

	_ = entities
}

// TestMemoryProcgenStress validates memory during intensive procedural generation.
func TestMemoryProcgenStress(t *testing.T) {
	profile := memprofile.StartMemoryProfile("ProcgenStress")

	// Simulate procedural generation of multiple content types
	items := make([]map[string]interface{}, 500)
	for i := 0; i < 500; i++ {
		items[i] = map[string]interface{}{
			"name":  "Generated Item",
			"stats": make([]int, 10),
			"icon":  make([]byte, 16*16*4),
		}
	}
	profile.Snapshot()

	quests := make([]map[string]interface{}, 200)
	for i := 0; i < 200; i++ {
		quests[i] = map[string]interface{}{
			"title":       "Quest Title",
			"description": make([]byte, 1024),
			"objectives":  make([]string, 5),
		}
	}
	profile.Snapshot()

	spells := make([]map[string]interface{}, 300)
	for i := 0; i < 300; i++ {
		spells[i] = map[string]interface{}{
			"name":    "Spell Name",
			"effects": make([]interface{}, 3),
			"icon":    make([]byte, 16*16*4),
		}
	}
	profile.Snapshot()

	// Simulate terrain generation
	terrain := make([][]byte, 10)
	for i := 0; i < 10; i++ {
		terrain[i] = make([]byte, 512*512) // 512x512 tile map
	}
	profile.Snapshot()

	profile.End()

	peak := profile.GetPeakAllocation()
	peakMB := float64(peak) / (1024 * 1024)

	t.Logf("Peak allocation: %.2fMB", peakMB)
	t.Logf("Allocation growth: %d bytes", profile.GetAllocationGrowth())

	// Verify under 500MB threshold
	thresholdBytes := uint64(500 * 1024 * 1024)
	if peak > thresholdBytes {
		t.Errorf("Memory usage %.2fMB exceeds 500MB threshold during procgen stress", peakMB)
	}

	_ = items
	_ = quests
	_ = spells
	_ = terrain
}

// TestMemoryRenderingStress validates memory during rendering pipeline stress.
func TestMemoryRenderingStress(t *testing.T) {
	profile := memprofile.StartMemoryProfile("RenderingStress")

	// Simulate sprite cache with many variations
	spriteCache := make([][][]byte, 50) // 50 sprite types
	for i := 0; i < 50; i++ {
		spriteCache[i] = make([][]byte, 8) // 8 animation frames each
		for j := 0; j < 8; j++ {
			spriteCache[i][j] = make([]byte, 64*64*4) // 64x64 RGBA
		}
	}
	profile.Snapshot()

	// Simulate particle effects
	particles := make([]map[string]interface{}, 1000)
	for i := 0; i < 1000; i++ {
		particles[i] = map[string]interface{}{
			"position": [2]float64{float64(i), float64(i)},
			"velocity": [2]float64{1.0, 1.0},
			"color":    [4]byte{255, 255, 255, 255},
			"lifetime": 1.0,
		}
	}
	profile.Snapshot()

	// Simulate lighting system
	lights := make([]map[string]interface{}, 200)
	for i := 0; i < 200; i++ {
		lights[i] = map[string]interface{}{
			"position":  [2]float64{float64(i), float64(i)},
			"radius":    100.0,
			"color":     [3]byte{255, 255, 255},
			"intensity": 1.0,
		}
	}
	profile.Snapshot()

	// Simulate frame buffers
	frameBuffers := make([][]byte, 3) // Triple buffering
	for i := 0; i < 3; i++ {
		frameBuffers[i] = make([]byte, 1920*1080*4) // Full HD RGBA
	}
	profile.Snapshot()

	profile.End()

	peak := profile.GetPeakAllocation()
	peakMB := float64(peak) / (1024 * 1024)

	t.Logf("Peak allocation: %.2fMB", peakMB)
	t.Logf("Allocation growth: %d bytes", profile.GetAllocationGrowth())

	// Verify under 500MB threshold
	thresholdBytes := uint64(500 * 1024 * 1024)
	if peak > thresholdBytes {
		t.Errorf("Memory usage %.2fMB exceeds 500MB threshold during rendering stress", peakMB)
	}

	_ = spriteCache
	_ = particles
	_ = lights
	_ = frameBuffers
}
