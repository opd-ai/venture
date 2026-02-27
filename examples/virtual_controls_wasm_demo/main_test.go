package main

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/engine"
)

// TestNewGame verifies game initialization.
func TestNewGame(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"Creates game with input provider"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			if g == nil {
				t.Fatal("NewGame() returned nil")
			}
			if g.inputProvider == nil {
				t.Error("inputProvider should not be nil")
			}
			if g.firstTouchFrame != -1 {
				t.Errorf("firstTouchFrame = %d, want -1", g.firstTouchFrame)
			}
			if g.frameCount != 0 {
				t.Errorf("frameCount = %d, want 0", g.frameCount)
			}
			if g.touchDetected {
				t.Error("touchDetected should be false initially")
			}
			if g.controlsVisible {
				t.Error("controlsVisible should be false initially")
			}
		})
	}
}

// TestUpdate_NoTouch verifies update without touch input.
func TestUpdate_NoTouch(t *testing.T) {
	tests := []struct {
		name       string
		frames     int
		wantFrames int
	}{
		{"Single frame", 1, 1},
		{"Multiple frames", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			stubInput := &engine.StubInput{}
			g.inputProvider = stubInput

			for i := 0; i < tt.frames; i++ {
				if err := g.Update(); err != nil {
					t.Fatalf("Update() error = %v", err)
				}
			}

			if g.frameCount != tt.wantFrames {
				t.Errorf("frameCount = %d, want %d", g.frameCount, tt.wantFrames)
			}
			if g.touchDetected {
				t.Error("touchDetected should remain false without touch")
			}
			if g.controlsVisible {
				t.Error("controlsVisible should remain false without touch")
			}
			if g.firstTouchFrame != -1 {
				t.Errorf("firstTouchFrame = %d, want -1", g.firstTouchFrame)
			}
		})
	}
}

// TestUpdate_FirstTouch verifies first touch detection (Gap #3 fix validation).
func TestUpdate_FirstTouch(t *testing.T) {
	tests := []struct {
		name              string
		touchFrame        int // Frame when touch is detected
		wantFirstFrame    int
		wantControlsFrame int // Frame when controls become visible
	}{
		{"Touch on frame 1", 1, 1, 1},
		{"Touch on frame 5", 5, 5, 5},
		{"Touch on frame 10", 10, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			stubInput := &engine.StubInput{
				TouchPositions: make(map[ebiten.TouchID]struct{ X, Y int }),
			}
			g.inputProvider = stubInput

			// Run frames before touch
			for i := 1; i < tt.touchFrame; i++ {
				if err := g.Update(); err != nil {
					t.Fatalf("Update() error = %v", err)
				}
				if g.touchDetected {
					t.Errorf("touchDetected at frame %d, want false", i)
				}
			}

			// Simulate touch on target frame
			stubInput.TouchPositions[0] = struct{ X, Y int }{100, 200}
			if err := g.Update(); err != nil {
				t.Fatalf("Update() error = %v", err)
			}

			// Verify first touch detected on same frame
			if !g.touchDetected {
				t.Error("touchDetected should be true after touch")
			}
			if g.firstTouchFrame != tt.wantFirstFrame {
				t.Errorf("firstTouchFrame = %d, want %d", g.firstTouchFrame, tt.wantFirstFrame)
			}
			if !g.controlsVisible {
				t.Error("controlsVisible should be true after first touch (0-frame delay)")
			}
			if g.frameCount != tt.wantControlsFrame {
				t.Errorf("controls became visible at frame %d, want %d (same frame as touch)", g.frameCount, tt.wantControlsFrame)
			}

			// Verify state persists on subsequent frames
			if err := g.Update(); err != nil {
				t.Fatalf("Update() error = %v", err)
			}
			if !g.touchDetected {
				t.Error("touchDetected should persist")
			}
			if !g.controlsVisible {
				t.Error("controlsVisible should persist")
			}
		})
	}
}

// TestUpdate_MultipleTouch verifies multiple simultaneous touches.
func TestUpdate_MultipleTouch(t *testing.T) {
	g := NewGame()
	stubInput := &engine.StubInput{
		TouchPositions: make(map[ebiten.TouchID]struct{ X, Y int }),
	}
	g.inputProvider = stubInput

	// Add multiple touches
	stubInput.TouchPositions[0] = struct{ X, Y int }{100, 100}
	stubInput.TouchPositions[1] = struct{ X, Y int }{200, 200}
	stubInput.TouchPositions[2] = struct{ X, Y int }{300, 300}

	if err := g.Update(); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify first touch detected
	if !g.touchDetected {
		t.Error("touchDetected should be true with multiple touches")
	}
	if g.firstTouchFrame != 1 {
		t.Errorf("firstTouchFrame = %d, want 1", g.firstTouchFrame)
	}
	if !g.controlsVisible {
		t.Error("controlsVisible should be true with multiple touches")
	}

	// Verify touch IDs available
	touchIDs := g.inputProvider.GetTouchIDs()
	if len(touchIDs) != 3 {
		t.Errorf("GetTouchIDs() returned %d touches, want 3", len(touchIDs))
	}
}

// TestUpdate_TouchRelease verifies behavior when touch is released.
func TestUpdate_TouchRelease(t *testing.T) {
	g := NewGame()
	stubInput := &engine.StubInput{
		TouchPositions: make(map[ebiten.TouchID]struct{ X, Y int }),
	}
	g.inputProvider = stubInput

	// Initial touch
	stubInput.TouchPositions[0] = struct{ X, Y int }{100, 100}
	if err := g.Update(); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if !g.touchDetected {
		t.Error("touchDetected should be true after first touch")
	}

	// Release touch (remove from map)
	delete(stubInput.TouchPositions, 0)
	if err := g.Update(); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify detection state persists (first touch already occurred)
	if !g.touchDetected {
		t.Error("touchDetected should persist after touch release")
	}
	if !g.controlsVisible {
		t.Error("controlsVisible should persist after touch release")
	}

	// Verify no active touches
	touchIDs := g.inputProvider.GetTouchIDs()
	if len(touchIDs) != 0 {
		t.Errorf("GetTouchIDs() returned %d touches after release, want 0", len(touchIDs))
	}
}

// TestLayout verifies screen size.
func TestLayout(t *testing.T) {
	tests := []struct {
		name                        string
		outsideWidth, outsideHeight int
		wantWidth, wantHeight       int
	}{
		{"Standard size", 1024, 768, screenWidth, screenHeight},
		{"Different outside size", 1920, 1080, screenWidth, screenHeight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			w, h := g.Layout(tt.outsideWidth, tt.outsideHeight)
			if w != tt.wantWidth {
				t.Errorf("Layout() width = %d, want %d", w, tt.wantWidth)
			}
			if h != tt.wantHeight {
				t.Errorf("Layout() height = %d, want %d", h, tt.wantHeight)
			}
		})
	}
}

// TestTouchPosition verifies touch position retrieval through InputProvider.
func TestTouchPosition(t *testing.T) {
	tests := []struct {
		name  string
		id    ebiten.TouchID
		x, y  int
		wantX int
		wantY int
	}{
		{"Touch at origin", 0, 0, 0, 0, 0},
		{"Touch at (100, 200)", 1, 100, 200, 100, 200},
		{"Touch at (800, 600)", 2, 800, 600, 800, 600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGame()
			stubInput := &engine.StubInput{
				TouchPositions: make(map[ebiten.TouchID]struct{ X, Y int }),
			}
			g.inputProvider = stubInput

			// Set touch position
			stubInput.TouchPositions[tt.id] = struct{ X, Y int }{tt.x, tt.y}

			// Retrieve through interface
			x, y := g.inputProvider.GetTouchPosition(tt.id)
			if x != tt.wantX {
				t.Errorf("GetTouchPosition() x = %d, want %d", x, tt.wantX)
			}
			if y != tt.wantY {
				t.Errorf("GetTouchPosition() y = %d, want %d", y, tt.wantY)
			}
		})
	}
}

// TestTouchPosition_InvalidID verifies behavior with invalid touch ID.
func TestTouchPosition_InvalidID(t *testing.T) {
	g := NewGame()
	stubInput := &engine.StubInput{
		TouchPositions: make(map[ebiten.TouchID]struct{ X, Y int }),
	}
	g.inputProvider = stubInput

	// Add one touch
	stubInput.TouchPositions[0] = struct{ X, Y int }{100, 100}

	// Query different ID
	x, y := g.inputProvider.GetTouchPosition(999)
	if x != 0 || y != 0 {
		t.Errorf("GetTouchPosition(invalid) = (%d, %d), want (0, 0)", x, y)
	}
}

// BenchmarkUpdate_NoTouch benchmarks update without touch.
func BenchmarkUpdate_NoTouch(b *testing.B) {
	g := NewGame()
	stubInput := &engine.StubInput{}
	g.inputProvider = stubInput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Update()
	}
}

// BenchmarkUpdate_WithTouch benchmarks update with active touch.
func BenchmarkUpdate_WithTouch(b *testing.B) {
	g := NewGame()
	stubInput := &engine.StubInput{
		TouchPositions: make(map[ebiten.TouchID]struct{ X, Y int }),
	}
	stubInput.TouchPositions[0] = struct{ X, Y int }{100, 100}
	g.inputProvider = stubInput

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = g.Update()
	}
}
