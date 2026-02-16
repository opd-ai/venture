// Package engine provides enhanced camera and visual feedback components - tests.
package engine

import (
	"math"
	"testing"
)

// TestScreenShakeComponent_New tests component creation.
func TestScreenShakeComponent_New(t *testing.T) {
	shake := NewScreenShakeComponent()

	if shake == nil {
		t.Fatal("NewScreenShakeComponent() returned nil")
	}

	if shake.Type() != "screenShake" {
		t.Errorf("expected type 'screenShake', got '%s'", shake.Type())
	}

	if shake.Active {
		t.Error("new shake should not be active")
	}

	if shake.Frequency != 15.0 {
		t.Errorf("expected default frequency 15.0, got %.2f", shake.Frequency)
	}
}

// TestScreenShakeComponent_TriggerShake tests shake triggering.
func TestScreenShakeComponent_TriggerShake(t *testing.T) {
	tests := []struct {
		name      string
		intensity float64
		duration  float64
		wantErr   bool
	}{
		{"valid shake", 5.0, 0.3, false},
		{"zero intensity", 0.0, 0.5, false},
		{"negative intensity", -1.0, 0.3, true},
		{"zero duration", 5.0, 0.0, true},
		{"negative duration", 5.0, -0.1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shake := NewScreenShakeComponent()
			err := shake.TriggerShake(tt.intensity, tt.duration)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantErr {
				if !shake.Active {
					t.Error("shake should be active after trigger")
				}
				if shake.Intensity != tt.intensity {
					t.Errorf("expected intensity %.2f, got %.2f", tt.intensity, shake.Intensity)
				}
				if shake.Duration != tt.duration {
					t.Errorf("expected duration %.2f, got %.2f", tt.duration, shake.Duration)
				}
			}
		})
	}
}

// TestScreenShakeComponent_StackingShakes tests shake stacking behavior.
func TestScreenShakeComponent_StackingShakes(t *testing.T) {
	shake := NewScreenShakeComponent()

	// First shake
	err := shake.TriggerShake(5.0, 0.3)
	if err != nil {
		t.Fatalf("TriggerShake failed: %v", err)
	}

	// Simulate time passing
	shake.Elapsed = 0.1

	// Second shake (higher intensity)
	err = shake.TriggerShake(8.0, 0.2)
	if err != nil {
		t.Fatalf("TriggerShake failed: %v", err)
	}

	// Should take higher intensity
	if shake.Intensity != 8.0 {
		t.Errorf("expected intensity 8.0 from stacking, got %.2f", shake.Intensity)
	}

	// Duration should extend
	if shake.Duration < 0.3 {
		t.Errorf("expected duration extended to at least 0.3, got %.2f", shake.Duration)
	}
}

// TestScreenShakeComponent_IsShaking tests shake status.
func TestScreenShakeComponent_IsShaking(t *testing.T) {
	shake := NewScreenShakeComponent()

	if shake.IsShaking() {
		t.Error("new shake should not be shaking")
	}

	shake.TriggerShake(5.0, 0.3)
	if !shake.IsShaking() {
		t.Error("shake should be active after trigger")
	}

	// Simulate shake completion
	shake.Elapsed = 0.4
	if shake.IsShaking() {
		t.Error("shake should not be active after duration")
	}
}

// TestScreenShakeComponent_GetProgress tests progress calculation.
func TestScreenShakeComponent_GetProgress(t *testing.T) {
	shake := NewScreenShakeComponent()
	shake.TriggerShake(5.0, 1.0)

	tests := []struct {
		elapsed  float64
		expected float64
	}{
		{0.0, 0.0},
		{0.25, 0.25},
		{0.5, 0.5},
		{0.75, 0.75},
		{1.0, 1.0},
		{1.5, 1.0}, // Clamped at 1.0
	}

	for _, tt := range tests {
		shake.Elapsed = tt.elapsed
		progress := shake.GetProgress()

		if math.Abs(progress-tt.expected) > 0.001 {
			t.Errorf("elapsed %.2f: expected progress %.2f, got %.2f", tt.elapsed, tt.expected, progress)
		}
	}
}

// TestScreenShakeComponent_GetCurrentIntensity tests intensity decay.
func TestScreenShakeComponent_GetCurrentIntensity(t *testing.T) {
	shake := NewScreenShakeComponent()
	shake.TriggerShake(10.0, 1.0)

	tests := []struct {
		elapsed  float64
		expected float64
	}{
		{0.0, 10.0}, // Start: 100% intensity
		{0.5, 5.0},  // Middle: 50% intensity
		{0.75, 2.5}, // 75%: 25% intensity
		{1.0, 0.0},  // End: 0% intensity
	}

	for _, tt := range tests {
		shake.Elapsed = tt.elapsed
		intensity := shake.GetCurrentIntensity()

		if math.Abs(intensity-tt.expected) > 0.001 {
			t.Errorf("elapsed %.2f: expected intensity %.2f, got %.2f", tt.elapsed, tt.expected, intensity)
		}
	}
}

// TestScreenShakeComponent_CalculateOffset tests offset calculation.
func TestScreenShakeComponent_CalculateOffset(t *testing.T) {
	shake := NewScreenShakeComponent()
	shake.TriggerShake(10.0, 1.0)

	// Calculate offset at different times
	shake.Elapsed = 0.0
	shake.CalculateOffset()

	// Should have some offset
	if shake.OffsetX == 0 && shake.OffsetY == 0 {
		t.Error("expected non-zero offset at shake start")
	}

	// Store initial offset
	initialX, initialY := shake.OffsetX, shake.OffsetY

	// Advance time
	shake.Elapsed = 0.1
	shake.CalculateOffset()

	// Offset should change (sine wave)
	if shake.OffsetX == initialX && shake.OffsetY == initialY {
		t.Error("offset should change over time")
	}

	// At end, offset should be zero
	shake.Elapsed = 1.0
	shake.CalculateOffset()
	if shake.OffsetX != 0 || shake.OffsetY != 0 {
		t.Errorf("expected zero offset at end, got (%.2f, %.2f)", shake.OffsetX, shake.OffsetY)
	}
}

// TestScreenShakeComponent_Reset tests reset functionality.
func TestScreenShakeComponent_Reset(t *testing.T) {
	shake := NewScreenShakeComponent()
	shake.TriggerShake(5.0, 0.3)
	shake.Elapsed = 0.1
	shake.CalculateOffset()

	shake.Reset()

	if shake.Active {
		t.Error("shake should not be active after reset")
	}
	if shake.Elapsed != 0 {
		t.Error("elapsed should be 0 after reset")
	}
	if shake.OffsetX != 0 || shake.OffsetY != 0 {
		t.Error("offsets should be 0 after reset")
	}
}

// TestHitStopComponent_New tests component creation.
func TestHitStopComponent_New(t *testing.T) {
	hitStop := NewHitStopComponent()

	if hitStop == nil {
		t.Fatal("NewHitStopComponent() returned nil")
	}

	if hitStop.Type() != "hitStop" {
		t.Errorf("expected type 'hitStop', got '%s'", hitStop.Type())
	}

	if hitStop.Active {
		t.Error("new hit-stop should not be active")
	}

	if hitStop.TimeScale != 0.0 {
		t.Errorf("expected default time scale 0.0, got %.2f", hitStop.TimeScale)
	}
}

// TestHitStopComponent_TriggerHitStop tests hit-stop triggering.
func TestHitStopComponent_TriggerHitStop(t *testing.T) {
	tests := []struct {
		name      string
		duration  float64
		timeScale float64
		wantErr   bool
	}{
		{"valid hit-stop", 0.1, 0.0, false},
		{"slow motion", 0.2, 0.1, false},
		{"zero duration", 0.0, 0.0, true},
		{"negative duration", -0.1, 0.0, true},
		{"negative time scale", 0.1, -0.1, true},
		{"time scale > 1", 0.1, 1.5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hitStop := NewHitStopComponent()
			err := hitStop.TriggerHitStop(tt.duration, tt.timeScale)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.wantErr {
				if !hitStop.Active {
					t.Error("hit-stop should be active after trigger")
				}
				if hitStop.Duration != tt.duration {
					t.Errorf("expected duration %.2f, got %.2f", tt.duration, hitStop.Duration)
				}
				if hitStop.TimeScale != tt.timeScale {
					t.Errorf("expected time scale %.2f, got %.2f", tt.timeScale, hitStop.TimeScale)
				}
			}
		})
	}
}

// TestHitStopComponent_StackingHitStops tests hit-stop stacking.
func TestHitStopComponent_StackingHitStops(t *testing.T) {
	hitStop := NewHitStopComponent()

	// First hit-stop
	err := hitStop.TriggerHitStop(0.1, 0.1)
	if err != nil {
		t.Fatalf("TriggerHitStop failed: %v", err)
	}

	// Simulate time passing
	hitStop.Elapsed = 0.05

	// Second hit-stop (more dramatic)
	err = hitStop.TriggerHitStop(0.15, 0.0)
	if err != nil {
		t.Fatalf("TriggerHitStop failed: %v", err)
	}

	// Should take lower time scale (more dramatic)
	if hitStop.TimeScale != 0.0 {
		t.Errorf("expected time scale 0.0 from stacking, got %.2f", hitStop.TimeScale)
	}

	// Duration should extend
	if hitStop.Duration < 0.15 {
		t.Errorf("expected duration extended to at least 0.15, got %.2f", hitStop.Duration)
	}
}

// TestHitStopComponent_IsActive tests active status.
func TestHitStopComponent_IsActive(t *testing.T) {
	hitStop := NewHitStopComponent()

	if hitStop.IsActive() {
		t.Error("new hit-stop should not be active")
	}

	hitStop.TriggerHitStop(0.1, 0.0)
	if !hitStop.IsActive() {
		t.Error("hit-stop should be active after trigger")
	}

	// Simulate completion
	hitStop.Elapsed = 0.15
	if hitStop.IsActive() {
		t.Error("hit-stop should not be active after duration")
	}
}

// TestHitStopComponent_GetTimeScale tests time scale retrieval.
func TestHitStopComponent_GetTimeScale(t *testing.T) {
	hitStop := NewHitStopComponent()

	// Not active: should return 1.0 (normal time)
	if hitStop.GetTimeScale() != 1.0 {
		t.Errorf("expected time scale 1.0 when inactive, got %.2f", hitStop.GetTimeScale())
	}

	// Active: should return set time scale
	hitStop.TriggerHitStop(0.1, 0.05)
	if hitStop.GetTimeScale() != 0.05 {
		t.Errorf("expected time scale 0.05 when active, got %.2f", hitStop.GetTimeScale())
	}

	// After duration: should return 1.0 again
	hitStop.Elapsed = 0.15
	if hitStop.GetTimeScale() != 1.0 {
		t.Errorf("expected time scale 1.0 after duration, got %.2f", hitStop.GetTimeScale())
	}
}

// TestHitStopComponent_Reset tests reset functionality.
func TestHitStopComponent_Reset(t *testing.T) {
	hitStop := NewHitStopComponent()
	hitStop.TriggerHitStop(0.1, 0.0)
	hitStop.Elapsed = 0.05

	hitStop.Reset()

	if hitStop.Active {
		t.Error("hit-stop should not be active after reset")
	}
	if hitStop.Elapsed != 0 {
		t.Error("elapsed should be 0 after reset")
	}
	if hitStop.GetTimeScale() != 1.0 {
		t.Errorf("time scale should be 1.0 after reset, got %.2f", hitStop.GetTimeScale())
	}
}

// TestCalculateShakeIntensity tests shake intensity calculation.
func TestCalculateShakeIntensity(t *testing.T) {
	tests := []struct {
		name         string
		damage       float64
		maxHP        float64
		scaleFactor  float64
		minIntensity float64
		maxIntensity float64
		expected     float64
	}{
		{"10 damage to 100 HP", 10, 100, 10, 1, 20, 1},    // 10/100*10 = 1 → clamped to min 1
		{"50 damage to 100 HP", 50, 100, 10, 1, 20, 5},    // 50/100*10 = 5
		{"100 damage to 100 HP", 100, 100, 10, 1, 20, 10}, // 100/100*10 = 10
		{"200 damage to 100 HP", 200, 100, 10, 1, 20, 20}, // 200/100*10 = 20 → clamped to max 20
		{"small damage", 1, 100, 10, 2, 20, 2},            // 1/100*10 = 0.1 → clamped to min 2
		{"zero max HP", 50, 0, 10, 1, 20, 5},              // Uses default 100
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateShakeIntensity(tt.damage, tt.maxHP, tt.scaleFactor, tt.minIntensity, tt.maxIntensity)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

// TestCalculateShakeDuration tests shake duration calculation.
func TestCalculateShakeDuration(t *testing.T) {
	tests := []struct {
		name               string
		intensity          float64
		baseDuration       float64
		additionalDuration float64
		maxIntensity       float64
		expected           float64
	}{
		{"low intensity", 5, 0.1, 0.2, 20, 0.15},    // 0.1 + (5/20)*0.2 = 0.15
		{"medium intensity", 10, 0.1, 0.2, 20, 0.2}, // 0.1 + (10/20)*0.2 = 0.2
		{"high intensity", 15, 0.1, 0.2, 20, 0.25},  // 0.1 + (15/20)*0.2 = 0.25
		{"max intensity", 20, 0.1, 0.2, 20, 0.3},    // 0.1 + (20/20)*0.2 = 0.3
		{"over max", 30, 0.1, 0.2, 20, 0.3},         // Clamped to max
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateShakeDuration(tt.intensity, tt.baseDuration, tt.additionalDuration, tt.maxIntensity)
			if math.Abs(result-tt.expected) > 0.001 {
				t.Errorf("expected %.3f, got %.3f", tt.expected, result)
			}
		})
	}
}

// Benchmark tests

func BenchmarkScreenShakeComponent_TriggerShake(b *testing.B) {
	shake := NewScreenShakeComponent()
	for i := 0; i < b.N; i++ {
		shake.TriggerShake(5.0, 0.3)
	}
}

func BenchmarkScreenShakeComponent_CalculateOffset(b *testing.B) {
	shake := NewScreenShakeComponent()
	shake.TriggerShake(5.0, 0.3)

	for i := 0; i < b.N; i++ {
		shake.Elapsed = float64(i) * 0.016 // 60 FPS simulation
		shake.CalculateOffset()
	}
}

func BenchmarkHitStopComponent_TriggerHitStop(b *testing.B) {
	hitStop := NewHitStopComponent()
	for i := 0; i < b.N; i++ {
		hitStop.TriggerHitStop(0.1, 0.0)
	}
}

// TestCameraSystem_IsVisible_NoCamera tests visibility when no camera is active.
// This is a regression test for the bug where IsVisible incorrectly culled sprites
// when activeCamera was nil.
func TestCameraSystem_IsVisible_NoCamera(t *testing.T) {
	cameraSystem := NewCameraSystem(800, 600)

	// Test entities at various world positions with no camera
	tests := []struct {
		name   string
		worldX float64
		worldY float64
		radius float64
		want   bool
	}{
		{"origin", 0, 0, 32, true},
		{"far positive", 10000, 10000, 32, true},
		{"far negative", -10000, -10000, 32, true},
		{"large radius", 5000, 5000, 500, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cameraSystem.IsVisible(tt.worldX, tt.worldY, tt.radius)
			if got != tt.want {
				t.Errorf("IsVisible(%v, %v, %v) = %v, want %v",
					tt.worldX, tt.worldY, tt.radius, got, tt.want)
			}
		})
	}
}

// TestCameraSystem_IsVisible_WithCamera tests visibility with an active camera.
func TestCameraSystem_IsVisible_WithCamera(t *testing.T) {
	cameraSystem := NewCameraSystem(800, 600)

	// Create camera entity
	cameraEntity := NewEntity(1)
	cameraComp := NewCameraComponent()
	cameraComp.X = 400 // Camera at center of 800x600 screen
	cameraComp.Y = 300
	cameraComp.Zoom = 1.0
	cameraEntity.AddComponent(cameraComp)

	pos := &PositionComponent{X: 400, Y: 300}
	cameraEntity.AddComponent(pos)

	cameraSystem.SetActiveCamera(cameraEntity)

	tests := []struct {
		name   string
		worldX float64
		worldY float64
		radius float64
		want   bool
	}{
		{"camera center", 400, 300, 32, true},
		{"visible left edge", 50, 300, 32, true},
		{"visible right edge", 750, 300, 32, true},
		{"visible top edge", 400, 50, 32, true},
		{"visible bottom edge", 400, 550, 32, true},
		{"far off screen left", -500, 300, 32, false},
		{"far off screen right", 1500, 300, 32, false},
		{"far off screen top", 400, -500, 32, false},
		{"far off screen bottom", 400, 1200, 32, false},
		{"with margin left", -20, 300, 32, true},  // Within margin (radius*2 = 64)
		{"with margin right", 820, 300, 32, true}, // Within margin
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cameraSystem.IsVisible(tt.worldX, tt.worldY, tt.radius)
			if got != tt.want {
				t.Errorf("IsVisible(%v, %v, %v) = %v, want %v",
					tt.worldX, tt.worldY, tt.radius, got, tt.want)
			}
		})
	}
}

// TestCameraSystem_IsVisible_NoComponent tests visibility when camera entity has no component.
func TestCameraSystem_IsVisible_NoComponent(t *testing.T) {
	cameraSystem := NewCameraSystem(800, 600)

	// Create entity without camera component
	entity := NewEntity(1)
	pos := &PositionComponent{X: 400, Y: 300}
	entity.AddComponent(pos)

	cameraSystem.SetActiveCamera(entity)

	// Should return true (all visible) since camera has no component
	got := cameraSystem.IsVisible(10000, 10000, 32)
	if !got {
		t.Error("IsVisible should return true when camera entity has no component")
	}
}

// TestSetCameraBoundsFromTerrain tests camera bounds calculation from terrain dimensions.
func TestSetCameraBoundsFromTerrain(t *testing.T) {
	tests := []struct {
		name            string
		terrainW        float64
		terrainH        float64
		screenW         int
		screenH         int
		zoom            float64
		wantMinX        float64
		wantMaxX        float64
		wantMinY        float64
		wantMaxY        float64
	}{
		{
			name:     "normal terrain zoom 1.0",
			terrainW: 2560, terrainH: 1600,
			screenW: 800, screenH: 600,
			zoom:     1.0,
			wantMinX: 400, wantMaxX: 2160,
			wantMinY: 300, wantMaxY: 1300,
		},
		{
			name:     "zoom 2.0 shrinks visible area",
			terrainW: 2560, terrainH: 1600,
			screenW: 800, screenH: 600,
			zoom:     2.0,
			wantMinX: 200, wantMaxX: 2360,
			wantMinY: 150, wantMaxY: 1450,
		},
		{
			name:     "terrain smaller than viewport in X centres",
			terrainW: 400, terrainH: 1600,
			screenW: 800, screenH: 600,
			zoom:     1.0,
			wantMinX: 200, wantMaxX: 200, // centred
			wantMinY: 300, wantMaxY: 1300,
		},
		{
			name:     "terrain smaller than viewport in both axes",
			terrainW: 400, terrainH: 300,
			screenW: 800, screenH: 600,
			zoom:     1.0,
			wantMinX: 200, wantMaxX: 200,
			wantMinY: 150, wantMaxY: 150,
		},
		{
			name:     "terrain exactly matches viewport",
			terrainW: 800, terrainH: 600,
			screenW: 800, screenH: 600,
			zoom:     1.0,
			wantMinX: 400, wantMaxX: 400, // centred
			wantMinY: 300, wantMaxY: 300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			camera := NewCameraComponent()
			camera.Zoom = tt.zoom
			SetCameraBoundsFromTerrain(camera, tt.terrainW, tt.terrainH, tt.screenW, tt.screenH)

			if math.Abs(camera.MinX-tt.wantMinX) > 0.001 {
				t.Errorf("MinX = %v, want %v", camera.MinX, tt.wantMinX)
			}
			if math.Abs(camera.MaxX-tt.wantMaxX) > 0.001 {
				t.Errorf("MaxX = %v, want %v", camera.MaxX, tt.wantMaxX)
			}
			if math.Abs(camera.MinY-tt.wantMinY) > 0.001 {
				t.Errorf("MinY = %v, want %v", camera.MinY, tt.wantMinY)
			}
			if math.Abs(camera.MaxY-tt.wantMaxY) > 0.001 {
				t.Errorf("MaxY = %v, want %v", camera.MaxY, tt.wantMaxY)
			}
		})
	}
}

// TestCameraSystem_BoundsClampingWithTerrain tests that camera position is clamped
// after setting terrain bounds.
func TestCameraSystem_BoundsClampingWithTerrain(t *testing.T) {
	cameraSystem := NewCameraSystem(800, 600)

	entity := NewEntity(1)
	cam := NewCameraComponent()
	cam.Zoom = 1.0
	cam.Smoothing = 0 // instant follow
	entity.AddComponent(cam)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	cameraSystem.SetActiveCamera(entity)

	// Set terrain bounds for an 80×50 tile map at 32px/tile = 2560×1600
	SetCameraBoundsFromTerrain(cam, 2560, 1600, 800, 600)

	// Position at origin should be clamped to MinX/MinY
	cameraSystem.Update([]*Entity{entity}, 1.0/60.0)
	if cam.X < cam.MinX || cam.Y < cam.MinY {
		t.Errorf("camera not clamped to min bounds: got (%v, %v), min (%v, %v)",
			cam.X, cam.Y, cam.MinX, cam.MinY)
	}

	// Position past max should be clamped
	pos := entity.GetPosition()
	pos.X = 3000
	pos.Y = 2000
	cameraSystem.Update([]*Entity{entity}, 1.0/60.0)
	if cam.X > cam.MaxX || cam.Y > cam.MaxY {
		t.Errorf("camera not clamped to max bounds: got (%v, %v), max (%v, %v)",
			cam.X, cam.Y, cam.MaxX, cam.MaxY)
	}
}

// TestPlayerBoundsClamp tests that BoundsComponent clamps player position to terrain.
func TestPlayerBoundsClamp(t *testing.T) {
	tests := []struct {
		name       string
		inputX     float64
		inputY     float64
		boundsMinX float64
		boundsMinY float64
		boundsMaxX float64
		boundsMaxY float64
		wantX      float64
		wantY      float64
	}{
		{"inside bounds", 100, 100, 0, 0, 2560, 1600, 100, 100},
		{"at min edge", 0, 0, 0, 0, 2560, 1600, 0, 0},
		{"at max edge", 2560, 1600, 0, 0, 2560, 1600, 2560, 1600},
		{"past min X", -50, 100, 0, 0, 2560, 1600, 0, 100},
		{"past max X", 3000, 100, 0, 0, 2560, 1600, 2560, 100},
		{"past both max", 3000, 2000, 0, 0, 2560, 1600, 2560, 1600},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bounds := &BoundsComponent{
				MinX: tt.boundsMinX, MinY: tt.boundsMinY,
				MaxX: tt.boundsMaxX, MaxY: tt.boundsMaxY,
			}
			gotX, gotY := bounds.Clamp(tt.inputX, tt.inputY)
			if gotX != tt.wantX || gotY != tt.wantY {
				t.Errorf("Clamp(%v, %v) = (%v, %v), want (%v, %v)",
					tt.inputX, tt.inputY, gotX, gotY, tt.wantX, tt.wantY)
			}
		})
	}
}

// TestSetCameraBoundsFromTerrain_StoresTerrainDimensions verifies that
// SetCameraBoundsFromTerrain stores terrain pixel dimensions on the component.
func TestSetCameraBoundsFromTerrain_StoresTerrainDimensions(t *testing.T) {
	cam := NewCameraComponent()
	SetCameraBoundsFromTerrain(cam, 2560, 1600, 800, 600)

	if cam.TerrainWidthPx != 2560 {
		t.Errorf("TerrainWidthPx = %v, want 2560", cam.TerrainWidthPx)
	}
	if cam.TerrainHeightPx != 1600 {
		t.Errorf("TerrainHeightPx = %v, want 1600", cam.TerrainHeightPx)
	}
}

// TestCameraSystem_RecalculateBoundsOnResize tests that camera bounds are
// recalculated when screen dimensions change, as happens on mobile orientation
// changes. Uses table-driven tests for various zoom levels and resize scenarios.
func TestCameraSystem_RecalculateBoundsOnResize(t *testing.T) {
	tests := []struct {
		name          string
		terrainW      float64
		terrainH      float64
		initialW      int
		initialH      int
		newW          int
		newH          int
		zoom          float64
		wantMinX      float64
		wantMaxX      float64
		wantMinY      float64
		wantMaxY      float64
	}{
		{
			name:     "landscape to portrait orientation",
			terrainW: 2560, terrainH: 1600,
			initialW: 800, initialH: 600,
			newW: 600, newH: 800,
			zoom:     1.0,
			wantMinX: 300, wantMaxX: 2260,
			wantMinY: 400, wantMaxY: 1200,
		},
		{
			name:     "resize with zoom 2.0",
			terrainW: 2560, terrainH: 1600,
			initialW: 800, initialH: 600,
			newW: 1024, newH: 768,
			zoom:     2.0,
			wantMinX: 256, wantMaxX: 2304,
			wantMinY: 192, wantMaxY: 1408,
		},
		{
			name:     "terrain smaller than new viewport centres",
			terrainW: 400, terrainH: 300,
			initialW: 300, initialH: 200,
			newW: 800, newH: 600,
			zoom:     1.0,
			wantMinX: 200, wantMaxX: 200,
			wantMinY: 150, wantMaxY: 150,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewCameraSystem(tt.initialW, tt.initialH)
			entity := NewEntity(1)
			cam := NewCameraComponent()
			cam.Zoom = tt.zoom
			cam.Smoothing = 0
			entity.AddComponent(cam)
			entity.AddComponent(&PositionComponent{X: 0, Y: 0})
			sys.SetActiveCamera(entity)

			// Set initial bounds.
			SetCameraBoundsFromTerrain(cam, tt.terrainW, tt.terrainH, tt.initialW, tt.initialH)

			// Simulate screen resize.
			sys.ScreenWidth = tt.newW
			sys.ScreenHeight = tt.newH
			sys.RecalculateBounds()

			if math.Abs(cam.MinX-tt.wantMinX) > 0.001 {
				t.Errorf("MinX = %v, want %v", cam.MinX, tt.wantMinX)
			}
			if math.Abs(cam.MaxX-tt.wantMaxX) > 0.001 {
				t.Errorf("MaxX = %v, want %v", cam.MaxX, tt.wantMaxX)
			}
			if math.Abs(cam.MinY-tt.wantMinY) > 0.001 {
				t.Errorf("MinY = %v, want %v", cam.MinY, tt.wantMinY)
			}
			if math.Abs(cam.MaxY-tt.wantMaxY) > 0.001 {
				t.Errorf("MaxY = %v, want %v", cam.MaxY, tt.wantMaxY)
			}
		})
	}
}

// TestCameraSystem_RecalculateBoundsNoActiveCamera verifies that
// RecalculateBounds is safe to call when there is no active camera.
func TestCameraSystem_RecalculateBoundsNoActiveCamera(t *testing.T) {
	sys := NewCameraSystem(800, 600)
	// Should not panic.
	sys.RecalculateBounds()
}

// TestCameraSystem_RecalculateBoundsNoTerrainDims verifies that
// RecalculateBounds is a no-op when terrain dimensions were never set.
func TestCameraSystem_RecalculateBoundsNoTerrainDims(t *testing.T) {
	sys := NewCameraSystem(800, 600)
	entity := NewEntity(1)
	cam := NewCameraComponent()
	entity.AddComponent(cam)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	sys.SetActiveCamera(entity)

	origMinX := cam.MinX
	sys.RecalculateBounds()

	if cam.MinX != origMinX {
		t.Errorf("bounds changed without terrain dimensions: MinX was %v, now %v", origMinX, cam.MinX)
	}
}

// TestCameraSystem_CornerClamping verifies that when the player is at any map
// corner the camera clamps so zero void pixels are visible.
func TestCameraSystem_CornerClamping(t *testing.T) {
	tests := []struct {
		name     string
		zoom     float64
		terrainW float64
		terrainH float64
		screenW  int
		screenH  int
		// Player position at a corner
		playerX float64
		playerY float64
		// Expected screen coordinate of the terrain corner closest to the player
		cornerWorldX float64
		cornerWorldY float64
		wantScreenX  float64
		wantScreenY  float64
	}{
		// All 4 corners × zoom 0.5
		{
			name: "top-left corner zoom 0.5",
			zoom: 0.5, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 0, playerY: 0,
			cornerWorldX: 0, cornerWorldY: 0,
			wantScreenX: 0, wantScreenY: 0,
		},
		{
			name: "top-right corner zoom 0.5",
			zoom: 0.5, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 2560, playerY: 0,
			cornerWorldX: 2560, cornerWorldY: 0,
			wantScreenX: 800, wantScreenY: 0,
		},
		{
			name: "bottom-left corner zoom 0.5",
			zoom: 0.5, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 0, playerY: 1600,
			cornerWorldX: 0, cornerWorldY: 1600,
			wantScreenX: 0, wantScreenY: 600,
		},
		{
			name: "bottom-right corner zoom 0.5",
			zoom: 0.5, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 2560, playerY: 1600,
			cornerWorldX: 2560, cornerWorldY: 1600,
			wantScreenX: 800, wantScreenY: 600,
		},
		// All 4 corners × zoom 1.0
		{
			name: "top-left corner zoom 1.0",
			zoom: 1.0, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 0, playerY: 0,
			cornerWorldX: 0, cornerWorldY: 0,
			wantScreenX: 0, wantScreenY: 0,
		},
		{
			name: "top-right corner zoom 1.0",
			zoom: 1.0, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 2560, playerY: 0,
			cornerWorldX: 2560, cornerWorldY: 0,
			wantScreenX: 800, wantScreenY: 0,
		},
		{
			name: "bottom-left corner zoom 1.0",
			zoom: 1.0, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 0, playerY: 1600,
			cornerWorldX: 0, cornerWorldY: 1600,
			wantScreenX: 0, wantScreenY: 600,
		},
		{
			name: "bottom-right corner zoom 1.0",
			zoom: 1.0, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 2560, playerY: 1600,
			cornerWorldX: 2560, cornerWorldY: 1600,
			wantScreenX: 800, wantScreenY: 600,
		},
		// All 4 corners × zoom 2.0
		{
			name: "top-left corner zoom 2.0",
			zoom: 2.0, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 0, playerY: 0,
			cornerWorldX: 0, cornerWorldY: 0,
			wantScreenX: 0, wantScreenY: 0,
		},
		{
			name: "top-right corner zoom 2.0",
			zoom: 2.0, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 2560, playerY: 0,
			cornerWorldX: 2560, cornerWorldY: 0,
			wantScreenX: 800, wantScreenY: 0,
		},
		{
			name: "bottom-left corner zoom 2.0",
			zoom: 2.0, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 0, playerY: 1600,
			cornerWorldX: 0, cornerWorldY: 1600,
			wantScreenX: 0, wantScreenY: 600,
		},
		{
			name: "bottom-right corner zoom 2.0",
			zoom: 2.0, terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			playerX: 2560, playerY: 1600,
			cornerWorldX: 2560, cornerWorldY: 1600,
			wantScreenX: 800, wantScreenY: 600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewCameraSystem(tt.screenW, tt.screenH)
			entity := NewEntity(1)
			cam := NewCameraComponent()
			cam.Zoom = tt.zoom
			cam.Smoothing = 0 // instant follow
			entity.AddComponent(cam)
			entity.AddComponent(&PositionComponent{X: tt.playerX, Y: tt.playerY})
			sys.SetActiveCamera(entity)

			SetCameraBoundsFromTerrain(cam, tt.terrainW, tt.terrainH, tt.screenW, tt.screenH)

			// Run one update to position the camera
			sys.Update([]*Entity{entity}, 1.0/60.0)

			sx, sy := sys.WorldToScreen(tt.cornerWorldX, tt.cornerWorldY)
			if math.Abs(sx-tt.wantScreenX) > 0.5 {
				t.Errorf("WorldToScreen X for corner (%v,%v) = %v, want %v",
					tt.cornerWorldX, tt.cornerWorldY, sx, tt.wantScreenX)
			}
			if math.Abs(sy-tt.wantScreenY) > 0.5 {
				t.Errorf("WorldToScreen Y for corner (%v,%v) = %v, want %v",
					tt.cornerWorldX, tt.cornerWorldY, sy, tt.wantScreenY)
			}
		})
	}
}

// TestCameraSystem_EdgeClamping verifies that when the player is along a single
// edge, that edge aligns flush with the corresponding screen edge.
func TestCameraSystem_EdgeClamping(t *testing.T) {
	tests := []struct {
		name       string
		zoom       float64
		playerX    float64
		playerY    float64
		probeX     float64 // world X to probe
		probeY     float64 // world Y to probe
		wantEdge   string  // which screen edge should align
		wantScreen float64 // expected screen coordinate on that axis
	}{
		{"left edge zoom 1.0", 1.0, 0, 800, 0, 800, "left", 0},
		{"right edge zoom 1.0", 1.0, 2560, 800, 2560, 800, "right", 800},
		{"top edge zoom 1.0", 1.0, 1280, 0, 1280, 0, "top", 0},
		{"bottom edge zoom 1.0", 1.0, 1280, 1600, 1280, 1600, "bottom", 600},
		{"left edge zoom 2.0", 2.0, 0, 800, 0, 800, "left", 0},
		{"right edge zoom 0.5", 0.5, 2560, 800, 2560, 800, "right", 800},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewCameraSystem(800, 600)
			entity := NewEntity(1)
			cam := NewCameraComponent()
			cam.Zoom = tt.zoom
			cam.Smoothing = 0
			entity.AddComponent(cam)
			entity.AddComponent(&PositionComponent{X: tt.playerX, Y: tt.playerY})
			sys.SetActiveCamera(entity)

			SetCameraBoundsFromTerrain(cam, 2560, 1600, 800, 600)
			sys.Update([]*Entity{entity}, 1.0/60.0)

			sx, sy := sys.WorldToScreen(tt.probeX, tt.probeY)

			switch tt.wantEdge {
			case "left", "right":
				if math.Abs(sx-tt.wantScreen) > 0.5 {
					t.Errorf("%s: screenX = %v, want %v", tt.wantEdge, sx, tt.wantScreen)
				}
			case "top", "bottom":
				if math.Abs(sy-tt.wantScreen) > 0.5 {
					t.Errorf("%s: screenY = %v, want %v", tt.wantEdge, sy, tt.wantScreen)
				}
			}
		})
	}
}

// TestCameraSystem_ZoomChangedRecalculatesBounds verifies that camera bounds
// are automatically recalculated when the zoom level changes.
func TestCameraSystem_ZoomChangedRecalculatesBounds(t *testing.T) {
	tests := []struct {
		name       string
		initialZ   float64
		newZ       float64
		terrainW   float64
		terrainH   float64
		screenW    int
		screenH    int
		wantMinX   float64
		wantMaxX   float64
	}{
		{
			name: "zoom in from 1.0 to 2.0",
			initialZ: 1.0, newZ: 2.0,
			terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			wantMinX: 200, wantMaxX: 2360,
		},
		{
			name: "zoom out from 1.0 to 0.5",
			initialZ: 1.0, newZ: 0.5,
			terrainW: 2560, terrainH: 1600, screenW: 800, screenH: 600,
			wantMinX: 800, wantMaxX: 1760,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewCameraSystem(tt.screenW, tt.screenH)
			entity := NewEntity(1)
			cam := NewCameraComponent()
			cam.Zoom = tt.initialZ
			cam.Smoothing = 0
			entity.AddComponent(cam)
			entity.AddComponent(&PositionComponent{X: 1280, Y: 800})
			sys.SetActiveCamera(entity)

			SetCameraBoundsFromTerrain(cam, tt.terrainW, tt.terrainH, tt.screenW, tt.screenH)
			sys.Update([]*Entity{entity}, 1.0/60.0)

			// Change zoom and update
			cam.Zoom = tt.newZ
			sys.Update([]*Entity{entity}, 1.0/60.0)

			if math.Abs(cam.MinX-tt.wantMinX) > 0.001 {
				t.Errorf("after zoom change: MinX = %v, want %v", cam.MinX, tt.wantMinX)
			}
			if math.Abs(cam.MaxX-tt.wantMaxX) > 0.001 {
				t.Errorf("after zoom change: MaxX = %v, want %v", cam.MaxX, tt.wantMaxX)
			}
		})
	}
}

// TestCameraSystem_InterpolationClamped verifies that interpolated camera
// positions (used in WorldToScreenInterpolated) are clamped to bounds.
func TestCameraSystem_InterpolationClamped(t *testing.T) {
	tests := []struct {
		name  string
		zoom  float64
		prevX float64
		prevY float64
		curX  float64
		curY  float64
		alpha float64
	}{
		{"prev outside bounds zoom 1.0", 1.0, 0, 0, 400, 300, 0.5},
		{"prev outside bounds zoom 2.0", 2.0, 0, 0, 200, 150, 0.5},
		{"prev outside bounds zoom 0.5", 0.5, 0, 0, 800, 600, 0.5},
		{"alpha near zero", 1.0, 0, 0, 400, 300, 0.1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewCameraSystem(800, 600)
			entity := NewEntity(1)
			cam := NewCameraComponent()
			cam.Zoom = tt.zoom
			cam.X = tt.curX
			cam.Y = tt.curY
			cam.PrevX = tt.prevX
			cam.PrevY = tt.prevY
			entity.AddComponent(cam)
			entity.AddComponent(&PositionComponent{X: tt.curX, Y: tt.curY})
			sys.SetActiveCamera(entity)

			SetCameraBoundsFromTerrain(cam, 2560, 1600, 800, 600)

			// Probe the top-left terrain corner. If interpolation is properly
			// clamped, the corner should appear at screen (0,0) or to the right/below.
			sx, sy := sys.WorldToScreenInterpolated(0, 0, tt.alpha)
			if sx < -0.5 {
				t.Errorf("terrain origin mapped to screenX=%v, expected >= 0 (no void)", sx)
			}
			if sy < -0.5 {
				t.Errorf("terrain origin mapped to screenY=%v, expected >= 0 (no void)", sy)
			}
		})
	}
}

// TestCameraSystem_BoundsZoomStored verifies that SetCameraBoundsFromTerrain
// stores the zoom level used for the bounds calculation.
func TestCameraSystem_BoundsZoomStored(t *testing.T) {
	cam := NewCameraComponent()
	cam.Zoom = 2.0
	SetCameraBoundsFromTerrain(cam, 2560, 1600, 800, 600)

	if cam.BoundsZoom != 2.0 {
		t.Errorf("BoundsZoom = %v, want 2.0", cam.BoundsZoom)
	}
}

// TestCameraSystem_SpawnInCorner verifies that camera clamping is applied
// immediately on the first frame when a player spawns at (0,0), before any
// input. The terrain origin must map to screen (0,0) — no void pixels.
func TestCameraSystem_SpawnInCorner(t *testing.T) {
	tests := []struct {
		name    string
		spawnX  float64
		spawnY  float64
		zoom    float64
		probeWX float64 // world X to probe (terrain corner)
		probeWY float64 // world Y to probe (terrain corner)
		wantSX  float64 // expected screen X
		wantSY  float64 // expected screen Y
	}{
		{
			name: "spawn at origin zoom 1.0",
			spawnX: 0, spawnY: 0, zoom: 1.0,
			probeWX: 0, probeWY: 0,
			wantSX: 0, wantSY: 0,
		},
		{
			name: "spawn at origin zoom 2.0",
			spawnX: 0, spawnY: 0, zoom: 2.0,
			probeWX: 0, probeWY: 0,
			wantSX: 0, wantSY: 0,
		},
		{
			name: "spawn at max corner zoom 1.0",
			spawnX: 2560, spawnY: 1600, zoom: 1.0,
			probeWX: 2560, probeWY: 1600,
			wantSX: 800, wantSY: 600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewCameraSystem(800, 600)
			entity := NewEntity(1)
			cam := NewCameraComponent()
			cam.Zoom = tt.zoom
			cam.Smoothing = 0 // instant follow, no delay
			entity.AddComponent(cam)
			entity.AddComponent(&PositionComponent{X: tt.spawnX, Y: tt.spawnY})
			sys.SetActiveCamera(entity)

			// Bounds are set before the first update, as setupCompletePlayerEntity does.
			SetCameraBoundsFromTerrain(cam, 2560, 1600, 800, 600)

			// First frame update — no prior player input.
			sys.Update([]*Entity{entity}, 1.0/60.0)

			sx, sy := sys.WorldToScreen(tt.probeWX, tt.probeWY)
			if math.Abs(sx-tt.wantSX) > 0.5 {
				t.Errorf("first frame: WorldToScreen X = %v, want %v", sx, tt.wantSX)
			}
			if math.Abs(sy-tt.wantSY) > 0.5 {
				t.Errorf("first frame: WorldToScreen Y = %v, want %v", sy, tt.wantSY)
			}
		})
	}
}

// TestCameraSystem_CenterOfMap verifies that when the player is near the
// center of a large map, the camera follows freely without any clamping
// distortion — the player maps to screen center.
func TestCameraSystem_CenterOfMap(t *testing.T) {
	tests := []struct {
		name    string
		playerX float64
		playerY float64
		zoom    float64
	}{
		{"exact center zoom 1.0", 1280, 800, 1.0},
		{"offset from center zoom 1.0", 1000, 700, 1.0},
		{"center zoom 2.0", 1280, 800, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewCameraSystem(800, 600)
			entity := NewEntity(1)
			cam := NewCameraComponent()
			cam.Zoom = tt.zoom
			cam.Smoothing = 0
			entity.AddComponent(cam)
			entity.AddComponent(&PositionComponent{X: tt.playerX, Y: tt.playerY})
			sys.SetActiveCamera(entity)

			SetCameraBoundsFromTerrain(cam, 2560, 1600, 800, 600)
			sys.Update([]*Entity{entity}, 1.0/60.0)

			// The player's world position should map to screen center.
			sx, sy := sys.WorldToScreen(tt.playerX, tt.playerY)
			halfW := float64(800) / 2
			halfH := float64(600) / 2
			if math.Abs(sx-halfW) > 0.5 {
				t.Errorf("player at center: screenX = %v, want %v", sx, halfW)
			}
			if math.Abs(sy-halfH) > 0.5 {
				t.Errorf("player at center: screenY = %v, want %v", sy, halfH)
			}
		})
	}
}

// TestCameraSystem_MapSmallerThanViewport verifies that when the map is smaller
// than the viewport on one or both axes, the camera centres the map and the
// player stays visible on screen regardless of position.
func TestCameraSystem_MapSmallerThanViewport(t *testing.T) {
	tests := []struct {
		name     string
		terrainW float64
		terrainH float64
		playerX  float64
		playerY  float64
	}{
		{"small map both axes, player at origin", 400, 300, 0, 0},
		{"small map both axes, player at max", 400, 300, 400, 300},
		{"small map X only, player at origin", 400, 1600, 0, 0},
		{"small map Y only, player mid", 2560, 300, 1280, 150},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sys := NewCameraSystem(800, 600)
			entity := NewEntity(1)
			cam := NewCameraComponent()
			cam.Zoom = 1.0
			cam.Smoothing = 0
			entity.AddComponent(cam)
			entity.AddComponent(&PositionComponent{X: tt.playerX, Y: tt.playerY})
			sys.SetActiveCamera(entity)

			SetCameraBoundsFromTerrain(cam, tt.terrainW, tt.terrainH, 800, 600)
			sys.Update([]*Entity{entity}, 1.0/60.0)

			// For a small-map axis the camera should centre on the terrain.
			// The terrain midpoint should be at screen centre on that axis.
			if tt.terrainW < 800 {
				midX := tt.terrainW / 2
				sx, _ := sys.WorldToScreen(midX, 0)
				if math.Abs(sx-400) > 0.5 {
					t.Errorf("small X: terrain midpoint at screenX=%v, want 400", sx)
				}
			}
			if tt.terrainH < 600 {
				midY := tt.terrainH / 2
				_, sy := sys.WorldToScreen(0, midY)
				if math.Abs(sy-300) > 0.5 {
					t.Errorf("small Y: terrain midpoint at screenY=%v, want 300", sy)
				}
			}
		})
	}
}
