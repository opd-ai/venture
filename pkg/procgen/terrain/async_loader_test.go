package terrain

import (
	"errors"
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// mockGenerator implements procgen.Generator for testing.
type mockGenerator struct {
	delay    time.Duration
	result   *Terrain
	err      error
	validate error
}

func (m *mockGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
}

func (m *mockGenerator) Validate(result interface{}) error {
	return m.validate
}

func TestAsyncLoader_NewAsyncLoader(t *testing.T) {
	tests := []struct {
		name       string
		logger     *logrus.Logger
		wantLogger bool
	}{
		{"with logger", logrus.New(), true},
		{"without logger", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := NewAsyncLoader(tt.logger)
			if loader == nil {
				t.Fatal("NewAsyncLoader returned nil")
			}
			if loader.done == nil {
				t.Error("done channel not initialized")
			}
			if (loader.logger != nil) != tt.wantLogger {
				t.Errorf("logger presence = %v, want %v", loader.logger != nil, tt.wantLogger)
			}
		})
	}
}

func TestAsyncLoader_StartGeneration_Success(t *testing.T) {
	terrain := &Terrain{
		Width:  80,
		Height: 50,
		Rooms:  []*Room{{X: 0, Y: 0, Width: 10, Height: 10}},
	}

	gen := &mockGenerator{
		delay:  10 * time.Millisecond,
		result: terrain,
	}

	loader := NewAsyncLoader(nil)
	params := procgen.GenerationParams{
		GenreID: "fantasy",
	}

	loader.StartGeneration(gen, 12345, params)

	// Progress should start at 0.1
	progress, err := loader.GetProgress()
	if err != nil {
		t.Errorf("GetProgress() error = %v, want nil", err)
	}
	if progress < 0.0 || progress > 0.2 {
		t.Errorf("initial progress = %v, want ~0.1", progress)
	}

	// Wait for completion
	result, err := loader.Wait()
	if err != nil {
		t.Fatalf("Wait() error = %v, want nil", err)
	}

	if result != terrain {
		t.Errorf("Wait() result = %v, want %v", result, terrain)
	}

	// Progress should be 1.0 when complete
	progress, err = loader.GetProgress()
	if err != nil {
		t.Errorf("GetProgress() after completion error = %v, want nil", err)
	}
	if progress != 1.0 {
		t.Errorf("final progress = %v, want 1.0", progress)
	}

	if !loader.IsDone() {
		t.Error("IsDone() = false, want true after completion")
	}
}

func TestAsyncLoader_StartGeneration_Error(t *testing.T) {
	expectedErr := errors.New("generation failed")
	gen := &mockGenerator{
		delay: 5 * time.Millisecond,
		err:   expectedErr,
	}

	loader := NewAsyncLoader(nil)
	params := procgen.GenerationParams{
		GenreID: "scifi",
	}

	loader.StartGeneration(gen, 54321, params)

	// Wait for completion
	result, err := loader.Wait()
	if err == nil {
		t.Fatal("Wait() error = nil, want error")
	}
	if err != expectedErr {
		t.Errorf("Wait() error = %v, want %v", err, expectedErr)
	}
	if result != nil {
		t.Errorf("Wait() result = %v, want nil on error", result)
	}

	// Progress should be 0.0 on error
	progress, errProgress := loader.GetProgress()
	if errProgress != expectedErr {
		t.Errorf("GetProgress() error = %v, want %v", errProgress, expectedErr)
	}
	if progress != 0.0 {
		t.Errorf("progress on error = %v, want 0.0", progress)
	}
}

func TestAsyncLoader_StartGeneration_InvalidType(t *testing.T) {
	// Generator returns wrong type
	gen := &mockGenerator{
		delay:  5 * time.Millisecond,
		result: nil, // Will return nil instead of *Terrain
	}

	// Override Generate to return invalid type
	invalidGen := &struct{ *mockGenerator }{gen}
	invalidGen.mockGenerator = gen

	loader := NewAsyncLoader(nil)
	params := procgen.GenerationParams{
		GenreID: "horror",
	}

	// Use a generator that returns an invalid type
	loader.StartGeneration(&invalidTypeGenerator{}, 99999, params)

	result, err := loader.Wait()
	if err == nil {
		t.Fatal("Wait() error = nil, want error for invalid type")
	}
	if result != nil {
		t.Errorf("Wait() result = %v, want nil for invalid type", result)
	}
}

// invalidTypeGenerator returns a non-Terrain type.
type invalidTypeGenerator struct{}

func (g *invalidTypeGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	return "not a terrain", nil
}

func (g *invalidTypeGenerator) Validate(result interface{}) error {
	return nil
}

func TestAsyncLoader_GetProgress_Concurrent(t *testing.T) {
	terrain := &Terrain{Width: 100, Height: 100}
	gen := &mockGenerator{
		delay:  20 * time.Millisecond,
		result: terrain,
	}

	loader := NewAsyncLoader(nil)
	params := procgen.GenerationParams{GenreID: "cyberpunk"}

	loader.StartGeneration(gen, 77777, params)

	// Concurrently poll progress from multiple goroutines
	done := make(chan struct{})
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					progress, err := loader.GetProgress()
					if err != nil {
						t.Errorf("GetProgress() concurrent error = %v", err)
					}
					if progress < 0.0 || progress > 1.0 {
						t.Errorf("GetProgress() concurrent progress = %v, want 0.0-1.0", progress)
					}
				}
			}
		}()
	}

	// Wait for completion
	_, err := loader.Wait()
	close(done)

	if err != nil {
		t.Errorf("Wait() error = %v, want nil", err)
	}
}

func TestAsyncLoader_IsDone(t *testing.T) {
	terrain := &Terrain{Width: 60, Height: 40}
	gen := &mockGenerator{
		delay:  15 * time.Millisecond,
		result: terrain,
	}

	loader := NewAsyncLoader(nil)
	params := procgen.GenerationParams{GenreID: "fantasy"}

	// Before starting
	if loader.IsDone() {
		t.Error("IsDone() = true before StartGeneration, want false")
	}

	loader.StartGeneration(gen, 11111, params)

	// During generation (may or may not be done depending on timing)
	// Just verify it doesn't panic
	_ = loader.IsDone()

	// After completion
	loader.Wait()
	if !loader.IsDone() {
		t.Error("IsDone() = false after completion, want true")
	}
}

func TestAsyncLoader_GetResult(t *testing.T) {
	terrain := &Terrain{Width: 120, Height: 80}
	gen := &mockGenerator{
		delay:  10 * time.Millisecond,
		result: terrain,
	}

	loader := NewAsyncLoader(nil)
	params := procgen.GenerationParams{GenreID: "scifi"}

	// Before generation
	if result := loader.GetResult(); result != nil {
		t.Errorf("GetResult() before generation = %v, want nil", result)
	}

	loader.StartGeneration(gen, 22222, params)

	// Wait for completion
	loader.Wait()

	// After completion
	result := loader.GetResult()
	if result != terrain {
		t.Errorf("GetResult() = %v, want %v", result, terrain)
	}
}

func TestAsyncLoader_WithLogger(t *testing.T) {
	// Test that logger doesn't cause issues
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	terrain := &Terrain{Width: 90, Height: 60, Rooms: []*Room{{X: 5, Y: 5, Width: 20, Height: 15}}}
	gen := &mockGenerator{
		delay:  5 * time.Millisecond,
		result: terrain,
	}

	loader := NewAsyncLoader(logger)
	params := procgen.GenerationParams{
		GenreID:    "horror",
		Difficulty: 0.5,
	}

	loader.StartGeneration(gen, 33333, params)

	result, err := loader.Wait()
	if err != nil {
		t.Fatalf("Wait() with logger error = %v, want nil", err)
	}
	if result.Width != 90 || result.Height != 60 {
		t.Errorf("terrain dimensions = %dx%d, want 90x60", result.Width, result.Height)
	}
}

func BenchmarkAsyncLoader_SmallTerrain(b *testing.B) {
	terrain := &Terrain{Width: 40, Height: 30}
	gen := &mockGenerator{result: terrain}

	params := procgen.GenerationParams{GenreID: "fantasy"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader := NewAsyncLoader(nil)
		loader.StartGeneration(gen, int64(i), params)
		_, _ = loader.Wait()
	}
}

func BenchmarkAsyncLoader_ConcurrentProgress(b *testing.B) {
	terrain := &Terrain{Width: 80, Height: 50}
	gen := &mockGenerator{
		delay:  10 * time.Millisecond,
		result: terrain,
	}

	params := procgen.GenerationParams{GenreID: "scifi"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loader := NewAsyncLoader(nil)
		loader.StartGeneration(gen, int64(i), params)

		// Poll progress concurrently
		done := make(chan struct{})
		go func() {
			for {
				select {
				case <-done:
					return
				default:
					loader.GetProgress()
				}
			}
		}()

		loader.Wait()
		close(done)
	}
}
