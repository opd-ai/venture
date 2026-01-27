package terrain

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
)

// TestAsyncLoader_Integration_BSP tests async loading with real BSP generator.
func TestAsyncLoader_Integration_BSP(t *testing.T) {
	gen := NewBSPGenerator()
	loader := NewAsyncLoader(nil)

	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Difficulty: 0.5,
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	loader.StartGeneration(gen, 12345, params)

	// Poll progress during generation
	progressUpdates := 0
	timeout := time.After(5 * time.Second)

	for !loader.IsDone() {
		select {
		case <-timeout:
			t.Fatal("async generation timed out after 5 seconds")
		default:
			progress, err := loader.GetProgress()
			if err != nil {
				t.Fatalf("GetProgress() error = %v", err)
			}
			if progress > 0.0 && progress < 1.0 {
				progressUpdates++
			}
			time.Sleep(1 * time.Millisecond)
		}
	}

	// Verify completion
	terrain, err := loader.Wait()
	if err != nil {
		t.Fatalf("Wait() error = %v", err)
	}

	if terrain == nil {
		t.Fatal("terrain is nil after successful generation")
	}

	if terrain.Width != 80 || terrain.Height != 50 {
		t.Errorf("terrain dimensions = %dx%d, want 80x50", terrain.Width, terrain.Height)
	}

	if len(terrain.Rooms) == 0 {
		t.Error("terrain has no rooms")
	}

	if progressUpdates == 0 {
		t.Log("warning: no progress updates captured (generation may be too fast)")
	}
}

// TestAsyncLoader_Integration_Cellular tests async loading with cellular automata generator.
func TestAsyncLoader_Integration_Cellular(t *testing.T) {
	gen := NewCellularGenerator()
	loader := NewAsyncLoader(nil)

	params := procgen.GenerationParams{
		GenreID:    "horror",
		Difficulty: 0.6,
		Custom: map[string]interface{}{
			"width":  100,
			"height": 80,
		},
	}

	loader.StartGeneration(gen, 54321, params)

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = loader.Wait()
	}()

	select {
	case <-done:
		// Success
	case <-time.After(10 * time.Second):
		t.Fatal("cellular generation timed out after 10 seconds")
	}

	terrain := loader.GetResult()
	if terrain == nil {
		t.Fatal("GetResult() returned nil after completion")
	}

	if terrain.Width != 100 || terrain.Height != 80 {
		t.Errorf("terrain dimensions = %dx%d, want 100x80", terrain.Width, terrain.Height)
	}
}

// TestAsyncLoader_Integration_Composite tests async loading with composite generator.
func TestAsyncLoader_Integration_Composite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping composite generation test in short mode")
	}

	gen := NewCompositeGenerator()
	loader := NewAsyncLoader(nil)

	params := procgen.GenerationParams{
		GenreID:    "scifi",
		Difficulty: 0.7,
		Custom: map[string]interface{}{
			"width":         120,
			"height":        100,
			"biomeCount":    3,
			"transitionWidth": 3,
		},
	}

	startTime := time.Now()
	loader.StartGeneration(gen, 99999, params)

	// Monitor progress
	var lastProgress float64
	for !loader.IsDone() {
		progress, err := loader.GetProgress()
		if err != nil {
			t.Fatalf("GetProgress() error = %v", err)
		}

		if progress < lastProgress {
			t.Errorf("progress went backwards: %v -> %v", lastProgress, progress)
		}
		lastProgress = progress

		time.Sleep(2 * time.Millisecond)
	}

	terrain, err := loader.Wait()
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Wait() error = %v", err)
	}

	if terrain == nil {
		t.Fatal("terrain is nil after successful generation")
	}

	if terrain.Width != 120 || terrain.Height != 100 {
		t.Errorf("terrain dimensions = %dx%d, want 120x100", terrain.Width, terrain.Height)
	}

	t.Logf("composite generation completed in %v", elapsed)

	// Verify progress reached 100%
	finalProgress, _ := loader.GetProgress()
	if finalProgress != 1.0 {
		t.Errorf("final progress = %v, want 1.0", finalProgress)
	}
}

// BenchmarkAsyncLoader_BSP benchmarks async loading with BSP generator.
func BenchmarkAsyncLoader_BSP(b *testing.B) {
	gen := NewBSPGenerator()
	params := procgen.GenerationParams{
		GenreID: "fantasy",
		Custom: map[string]interface{}{
			"width":  80,
			"height": 50,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader := NewAsyncLoader(nil)
		loader.StartGeneration(gen, int64(i), params)
		_, err := loader.Wait()
		if err != nil {
			b.Fatalf("generation failed: %v", err)
		}
	}
}

// BenchmarkAsyncLoader_Cellular benchmarks async loading with cellular generator.
func BenchmarkAsyncLoader_Cellular(b *testing.B) {
	gen := NewCellularGenerator()
	params := procgen.GenerationParams{
		GenreID: "horror",
		Custom: map[string]interface{}{
			"width":  100,
			"height": 80,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader := NewAsyncLoader(nil)
		loader.StartGeneration(gen, int64(i), params)
		_, err := loader.Wait()
		if err != nil {
			b.Fatalf("generation failed: %v", err)
		}
	}
}
