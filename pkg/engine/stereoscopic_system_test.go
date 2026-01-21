package engine

import (
	"testing"
)

func TestNewStereoscopicSystem(t *testing.T) {
	world := &World{}
	sys := NewStereoscopicSystem(world)

	if sys == nil {
		t.Fatal("NewStereoscopicSystem returned nil")
	}

	if !sys.IsEnabled() {
		t.Error("Expected system to be enabled by default")
	}

	if sys.GetRenderPhase() != RenderPhaseIdle {
		t.Errorf("Expected render phase %s, got %s", RenderPhaseIdle, sys.GetRenderPhase())
	}

	if sys.GetFrameCount() != 0 {
		t.Errorf("Expected frame count 0, got %d", sys.GetFrameCount())
	}
}

func TestStereoscopicSystem_SetEnabled(t *testing.T) {
	world := &World{}
	sys := NewStereoscopicSystem(world)

	sys.SetEnabled(true)
	if !sys.IsEnabled() {
		t.Error("Expected system to be enabled")
	}

	sys.SetEnabled(false)
	if sys.IsEnabled() {
		t.Error("Expected system to be disabled")
	}
}

func TestStereoscopicSystem_Update_Disabled(t *testing.T) {
	world := &World{}
	sys := NewStereoscopicSystem(world)
	sys.SetEnabled(false) // Explicitly disable for this test

	// Create entity with stereoscopic component
	entity := NewEntity(1)
	stereo := NewStereoscopicComponent()
	stereo.SetEnabled(true)
	entity.AddComponent(stereo)

	// System is disabled, should not process
	sys.Update([]*Entity{entity}, 0.016)

	if sys.GetFrameCount() != 0 {
		t.Error("Expected no frames processed when system disabled")
	}
}

func TestStereoscopicSystem_Update_Enabled(t *testing.T) {
	world := &World{}
	sys := NewStereoscopicSystem(world)
	// System is enabled by default

	// Create entity with stereoscopic component
	entity := NewEntity(1)
	stereo := NewStereoscopicComponent()
	stereo.SetEnabled(true)
	entity.AddComponent(stereo)

	// System is enabled, should process
	sys.Update([]*Entity{entity}, 0.016)

	if sys.GetFrameCount() != 1 {
		t.Errorf("Expected 1 frame processed, got %d", sys.GetFrameCount())
	}

	// Process more frames
	sys.Update([]*Entity{entity}, 0.016)
	sys.Update([]*Entity{entity}, 0.016)

	if sys.GetFrameCount() != 3 {
		t.Errorf("Expected 3 frames processed, got %d", sys.GetFrameCount())
	}
}

func TestStereoscopicSystem_Update_DisabledComponent(t *testing.T) {
	world := &World{}
	sys := NewStereoscopicSystem(world)
	sys.SetEnabled(true)

	// Create entity with disabled stereoscopic component
	entity := NewEntity(1)
	stereo := NewStereoscopicComponent()
	stereo.SetEnabled(false) // Component disabled
	entity.AddComponent(stereo)

	callbackCalled := false
	sys.SetLeftEyeCallback(func(offset float64) {
		callbackCalled = true
	})

	sys.Update([]*Entity{entity}, 0.016)

	// Callbacks should not be called for disabled components
	if callbackCalled {
		t.Error("Callback should not be called for disabled component")
	}
}

func TestStereoscopicSystem_Callbacks(t *testing.T) {
	world := &World{}
	sys := NewStereoscopicSystem(world)
	sys.SetEnabled(true)

	// Track callback invocations
	leftCalled := false
	rightCalled := false
	postCalled := false
	var leftOffset, rightOffset float64

	sys.SetLeftEyeCallback(func(offset float64) {
		leftCalled = true
		leftOffset = offset
	})

	sys.SetRightEyeCallback(func(offset float64) {
		rightCalled = true
		rightOffset = offset
	})

	sys.SetPostRenderCallback(func() {
		postCalled = true
	})

	// Create entity with stereoscopic component
	entity := NewEntity(1)
	stereo := NewStereoscopicComponent()
	stereo.SetEnabled(true)
	stereo.SetIPD(64.0)
	entity.AddComponent(stereo)

	sys.Update([]*Entity{entity}, 0.016)

	// All callbacks should be called
	if !leftCalled {
		t.Error("Left eye callback not called")
	}
	if !rightCalled {
		t.Error("Right eye callback not called")
	}
	if !postCalled {
		t.Error("Post render callback not called")
	}

	// Verify offsets
	expectedSeparation := (64.0 / 1000.0) / 2.0
	if leftOffset != -expectedSeparation {
		t.Errorf("Expected left offset %v, got %v", -expectedSeparation, leftOffset)
	}
	if rightOffset != expectedSeparation {
		t.Errorf("Expected right offset %v, got %v", expectedSeparation, rightOffset)
	}
}

func TestStereoscopicSystem_CallbackOrder(t *testing.T) {
	world := &World{}
	sys := NewStereoscopicSystem(world)
	sys.SetEnabled(true)

	// Track callback order
	order := []string{}

	sys.SetLeftEyeCallback(func(offset float64) {
		order = append(order, "left")
	})

	sys.SetRightEyeCallback(func(offset float64) {
		order = append(order, "right")
	})

	sys.SetPostRenderCallback(func() {
		order = append(order, "post")
	})

	entity := NewEntity(1)
	stereo := NewStereoscopicComponent()
	stereo.SetEnabled(true)
	entity.AddComponent(stereo)

	sys.Update([]*Entity{entity}, 0.016)

	// Verify order: left -> right -> post
	if len(order) != 3 {
		t.Fatalf("Expected 3 callbacks, got %d", len(order))
	}
	if order[0] != "left" {
		t.Error("Expected left callback first")
	}
	if order[1] != "right" {
		t.Error("Expected right callback second")
	}
	if order[2] != "post" {
		t.Error("Expected post callback third")
	}
}

func TestStereoscopicSystem_RenderPhase(t *testing.T) {
	world := &World{}
	sys := NewStereoscopicSystem(world)
	sys.SetEnabled(true)

	// Track render phase during callbacks
	phases := []string{}

	sys.SetLeftEyeCallback(func(offset float64) {
		phases = append(phases, sys.GetRenderPhase())
	})

	sys.SetRightEyeCallback(func(offset float64) {
		phases = append(phases, sys.GetRenderPhase())
	})

	sys.SetPostRenderCallback(func() {
		phases = append(phases, sys.GetRenderPhase())
	})

	entity := NewEntity(1)
	stereo := NewStereoscopicComponent()
	stereo.SetEnabled(true)
	entity.AddComponent(stereo)

	sys.Update([]*Entity{entity}, 0.016)

	// Verify phases
	if len(phases) != 3 {
		t.Fatalf("Expected 3 phases recorded, got %d", len(phases))
	}
	if phases[0] != RenderPhaseLeftEye {
		t.Errorf("Expected %s, got %s", RenderPhaseLeftEye, phases[0])
	}
	if phases[1] != RenderPhaseRightEye {
		t.Errorf("Expected %s, got %s", RenderPhaseRightEye, phases[1])
	}
	if phases[2] != RenderPhaseComposite {
		t.Errorf("Expected %s, got %s", RenderPhaseComposite, phases[2])
	}

	// After update, phase should be idle
	if sys.GetRenderPhase() != RenderPhaseIdle {
		t.Errorf("Expected %s after update, got %s", RenderPhaseIdle, sys.GetRenderPhase())
	}
}

func TestStereoscopicSystem_MultipleEntities(t *testing.T) {
	world := &World{}
	sys := NewStereoscopicSystem(world)
	sys.SetEnabled(true)

	leftCallCount := 0
	sys.SetLeftEyeCallback(func(offset float64) {
		leftCallCount++
	})

	// Create multiple entities
	entities := make([]*Entity, 3)
	for i := 0; i < 3; i++ {
		entities[i] = NewEntity(uint64(i + 1))
		stereo := NewStereoscopicComponent()
		stereo.SetEnabled(true)
		entities[i].AddComponent(stereo)
	}

	sys.Update(entities, 0.016)

	// Should process all 3 entities
	if leftCallCount != 3 {
		t.Errorf("Expected 3 left eye callbacks, got %d", leftCallCount)
	}
}

func TestStereoscopicSystem_GetStats(t *testing.T) {
	world := &World{}
	sys := NewStereoscopicSystem(world)
	sys.SetEnabled(true)

	entity := NewEntity(1)
	stereo := NewStereoscopicComponent()
	stereo.SetEnabled(true)
	entity.AddComponent(stereo)

	// Process some frames
	for i := 0; i < 10; i++ {
		sys.Update([]*Entity{entity}, 0.016)
	}

	stats := sys.GetStats()
	if stats.FrameCount != 10 {
		t.Errorf("Expected frame count 10, got %d", stats.FrameCount)
	}
}

func TestCalculateStereoProjection(t *testing.T) {
	tests := []struct {
		name        string
		eyeOffset   float64
		convergence float64
		fov         float64
	}{
		{"basic", 0.032, 10.0, 1.0},
		{"narrow fov", 0.032, 10.0, 0.5},
		{"wide fov", 0.032, 10.0, 2.0},
		{"near convergence", 0.032, 1.0, 1.0},
		{"far convergence", 0.032, 100.0, 1.0},
		{"zero convergence handled", 0.032, 0.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shift := CalculateStereoProjection(tt.eyeOffset, tt.convergence, tt.fov)

			// Basic sanity checks
			if tt.eyeOffset > 0 && shift <= 0 {
				t.Error("Positive eye offset should produce positive shift")
			}
			if tt.eyeOffset < 0 && shift >= 0 {
				t.Error("Negative eye offset should produce negative shift")
			}
		})
	}
}

func TestCalculateStereoProjection_Deterministic(t *testing.T) {
	for i := 0; i < 100; i++ {
		s1 := CalculateStereoProjection(0.032, 10.0, 1.0)
		s2 := CalculateStereoProjection(0.032, 10.0, 1.0)
		if s1 != s2 {
			t.Errorf("CalculateStereoProjection not deterministic: %v != %v", s1, s2)
		}
	}
}

func TestCalculateViewportForEye(t *testing.T) {
	tests := []struct {
		name        string
		eye         string
		totalWidth  int
		totalHeight int
		expectX     int
		expectY     int
		expectW     int
		expectH     int
	}{
		{"left eye 1920x1080", EyeLeft, 1920, 1080, 0, 0, 960, 1080},
		{"right eye 1920x1080", EyeRight, 1920, 1080, 960, 0, 960, 1080},
		{"left eye 2560x1440", EyeLeft, 2560, 1440, 0, 0, 1280, 1440},
		{"right eye 2560x1440", EyeRight, 2560, 1440, 1280, 0, 1280, 1440},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			x, y, w, h := CalculateViewportForEye(tt.eye, tt.totalWidth, tt.totalHeight)

			if x != tt.expectX {
				t.Errorf("Expected x %d, got %d", tt.expectX, x)
			}
			if y != tt.expectY {
				t.Errorf("Expected y %d, got %d", tt.expectY, y)
			}
			if w != tt.expectW {
				t.Errorf("Expected width %d, got %d", tt.expectW, w)
			}
			if h != tt.expectH {
				t.Errorf("Expected height %d, got %d", tt.expectH, h)
			}
		})
	}
}

func TestApplyAsymmetricFrustum(t *testing.T) {
	// Test symmetric case (zero offset)
	left, right, top, bottom := ApplyAsymmetricFrustum(0.0, 0.1, 1.0, 16.0/9.0)

	// With zero offset, left and right should be symmetric
	if left != -right {
		t.Errorf("With zero offset, left (%v) should equal -right (%v)", left, right)
	}
	if top != -bottom {
		t.Errorf("Top (%v) should equal -bottom (%v)", top, bottom)
	}

	// Test with positive offset (right eye)
	leftR, rightR, _, _ := ApplyAsymmetricFrustum(0.032, 0.1, 1.0, 16.0/9.0)

	// Right eye frustum should be shifted right (both planes move in positive direction)
	// left plane becomes less negative (increases), right plane increases
	if leftR <= left {
		t.Errorf("Right eye left plane should be shifted right (increase): %v <= %v", leftR, left)
	}
	if rightR <= right {
		t.Errorf("Right eye right plane should be shifted right (increase): %v <= %v", rightR, right)
	}

	// Test with negative offset (left eye)
	leftL, rightL, _, _ := ApplyAsymmetricFrustum(-0.032, 0.1, 1.0, 16.0/9.0)

	// Left eye frustum should be shifted left (both planes move in negative direction)
	// left plane becomes more negative (decreases), right plane decreases
	if leftL >= left {
		t.Errorf("Left eye left plane should be shifted left (decrease): %v >= %v", leftL, left)
	}
	if rightL >= right {
		t.Errorf("Left eye right plane should be shifted left (decrease): %v >= %v", rightL, right)
	}
}

func TestStereoscopicSystem_ThreadSafety(t *testing.T) {
	world := &World{}
	sys := NewStereoscopicSystem(world)

	done := make(chan bool, 4)

	go func() {
		for i := 0; i < 1000; i++ {
			sys.SetEnabled(i%2 == 0)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = sys.IsEnabled()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = sys.GetRenderPhase()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = sys.GetFrameCount()
		}
		done <- true
	}()

	for i := 0; i < 4; i++ {
		<-done
	}
}

func BenchmarkStereoscopicSystem_Update(b *testing.B) {
	world := &World{}
	sys := NewStereoscopicSystem(world)
	sys.SetEnabled(true)

	entity := NewEntity(1)
	stereo := NewStereoscopicComponent()
	stereo.SetEnabled(true)
	entity.AddComponent(stereo)

	entities := []*Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}

func BenchmarkCalculateStereoProjection(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateStereoProjection(0.032, 10.0, 1.0)
	}
}

func BenchmarkCalculateViewportForEye(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateViewportForEye(EyeLeft, 1920, 1080)
		CalculateViewportForEye(EyeRight, 1920, 1080)
	}
}
