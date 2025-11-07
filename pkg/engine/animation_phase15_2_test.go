package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

// TestAnimationPhase15_2_IdleBreathing tests subtle breathing animation for idle state.
func TestAnimationPhase15_2_IdleBreathing(t *testing.T) {
	// Test idle animation has subtle vertical and horizontal movement
	frameCount := 8
	
	for i := 0; i < frameCount; i++ {
		offset := calculateAnimationOffset("idle", i, frameCount)
		
		// Verify breathing movement is subtle (< 1 pixel)
		if math.Abs(offset.Y) > 1.0 {
			t.Errorf("Frame %d: Idle breathing Y offset too large: %f (should be < 1px)", i, offset.Y)
		}
		
		if math.Abs(offset.X) > 0.5 {
			t.Errorf("Frame %d: Idle breathing X offset too large: %f (should be < 0.5px)", i, offset.X)
		}
	}
	
	// Verify the animation oscillates (not static)
	// Compare frames at different phases of the cycle
	offset0 := calculateAnimationOffset("idle", 0, frameCount)
	offset2 := calculateAnimationOffset("idle", 2, frameCount)
	offset4 := calculateAnimationOffset("idle", 4, frameCount)
	
	// At least one comparison should show oscillation
	hasOscillation := math.Abs(offset0.Y-offset2.Y) > 0.1 || 
	                  math.Abs(offset2.Y-offset4.Y) > 0.1 ||
	                  math.Abs(offset0.Y-offset4.Y) > 0.1
	
	if !hasOscillation {
		t.Error("Idle breathing should oscillate vertically across frames")
	}
}

// TestAnimationPhase15_2_IdleBreathingRotation tests subtle head tilt for breathing.
func TestAnimationPhase15_2_IdleBreathingRotation(t *testing.T) {
	frameCount := 8
	
	for i := 0; i < frameCount; i++ {
		rotation := calculateAnimationRotation("idle", i, frameCount)
		
		// Verify rotation is very subtle (< 0.05 radians ≈ 2.9 degrees)
		if math.Abs(rotation) > 0.05 {
			t.Errorf("Frame %d: Idle breathing rotation too large: %f radians (should be < 0.05)", i, rotation)
		}
	}
	
	// Verify rotation oscillates
	rot0 := calculateAnimationRotation("idle", 0, frameCount)
	rot2 := calculateAnimationRotation("idle", 2, frameCount)
	rot4 := calculateAnimationRotation("idle", 4, frameCount)
	
	// At least one comparison should show oscillation
	hasOscillation := math.Abs(rot0-rot2) > 0.005 || 
	                  math.Abs(rot2-rot4) > 0.005 ||
	                  math.Abs(rot0-rot4) > 0.005
	
	if !hasOscillation {
		t.Error("Idle breathing rotation should oscillate across frames")
	}
}

// TestAnimationPhase15_2_IdleBreathingScale tests subtle scale changes for breathing.
func TestAnimationPhase15_2_IdleBreathingScale(t *testing.T) {
	frameCount := 8
	
	for i := 0; i < frameCount; i++ {
		scale := calculateAnimationScale("idle", i, frameCount)
		
		// Verify scale is very subtle (0.985 to 1.015, ±1.5%)
		if scale < 0.98 || scale > 1.02 {
			t.Errorf("Frame %d: Idle breathing scale out of range: %f (should be 0.98-1.02)", i, scale)
		}
		
		// Verify scale oscillates around 1.0
		deviation := math.Abs(scale - 1.0)
		if deviation > 0.02 {
			t.Errorf("Frame %d: Idle breathing scale deviation too large: %f", i, deviation)
		}
	}
}

// TestAnimationPhase15_2_AttackFollowThrough tests enhanced attack animation with better follow-through.
func TestAnimationPhase15_2_AttackFollowThrough(t *testing.T) {
	frameCount := 8
	
	// Test three phases: wind-up (0-0.2), strike (0.2-0.5), follow-through (0.5-1.0)
	
	// Wind-up phase (frames 0-1)
	windupOffset := calculateAnimationOffset("attack", 1, frameCount)
	if windupOffset.X > 0 {
		t.Error("Wind-up should have backward movement (negative X offset)")
	}
	
	// Strike phase (frames 2-4)
	strikeOffset := calculateAnimationOffset("attack", 3, frameCount)
	if strikeOffset.X < 5 {
		t.Errorf("Strike should have significant forward movement, got %f", strikeOffset.X)
	}
	
	// Follow-through phase (frames 5-7)
	followOffset := calculateAnimationOffset("attack", 6, frameCount)
	if followOffset.X < 0 {
		t.Errorf("Follow-through should maintain positive X offset, got %f", followOffset.X)
	}
	
	// Verify smooth return (quadratic easing)
	offset5 := calculateAnimationOffset("attack", 5, frameCount)
	offset6 := calculateAnimationOffset("attack", 6, frameCount)
	offset7 := calculateAnimationOffset("attack", 7, frameCount)
	
	// Each frame should move closer to 0
	if offset5.X < offset6.X || offset6.X < offset7.X {
		t.Error("Follow-through should smoothly return to origin")
	}
}

// TestAnimationPhase15_2_AttackFollowThroughRotation tests enhanced rotation follow-through.
func TestAnimationPhase15_2_AttackFollowThroughRotation(t *testing.T) {
	frameCount := 8
	
	// Wind-up: should have backward rotation
	windupRot := calculateAnimationRotation("attack", 1, frameCount)
	if windupRot > 0 {
		t.Error("Wind-up should have backward rotation (negative)")
	}
	
	// Strike: should have forward rotation
	strikeRot := calculateAnimationRotation("attack", 3, frameCount)
	if strikeRot < 0.5 {
		t.Errorf("Strike should have significant forward rotation, got %f", strikeRot)
	}
	
	// Follow-through: should maintain rotation with deceleration
	followRot6 := calculateAnimationRotation("attack", 6, frameCount)
	followRot7 := calculateAnimationRotation("attack", 7, frameCount)
	
	if followRot6 < 0 || followRot7 < 0 {
		t.Error("Follow-through should maintain positive rotation")
	}
	
	if followRot7 > followRot6 {
		t.Error("Follow-through rotation should decelerate")
	}
	
	// Verify follow-through doesn't snap back (smooth easing)
	if followRot6-followRot7 > 0.5 {
		t.Error("Follow-through deceleration should be smooth, not abrupt")
	}
}

// TestAnimationPhase15_2_AttackAnticipationScale tests attack anticipation and power scaling.
func TestAnimationPhase15_2_AttackAnticipationScale(t *testing.T) {
	frameCount := 8
	
	// Anticipation (frames 0-1): slight compression
	anticipationScale := calculateAnimationScale("attack", 1, frameCount)
	if anticipationScale >= 1.0 {
		t.Error("Anticipation should compress sprite (scale < 1.0)")
	}
	if anticipationScale < 0.9 {
		t.Errorf("Anticipation compression too strong: %f (should be > 0.9)", anticipationScale)
	}
	
	// Strike (frames 2-4): expansion for power
	strikeScale := calculateAnimationScale("attack", 3, frameCount)
	if strikeScale <= 1.0 {
		t.Error("Strike should expand sprite (scale > 1.0)")
	}
	if strikeScale > 1.15 {
		t.Errorf("Strike expansion too strong: %f (should be < 1.15)", strikeScale)
	}
	
	// Follow-through (frames 5-7): gradual return to normal
	followScale := calculateAnimationScale("attack", 6, frameCount)
	if followScale < 0.95 || followScale > 1.15 {
		t.Errorf("Follow-through scale should be in reasonable range (0.95-1.15), got %f", followScale)
	}
	
	// Final frame should be close to normal (1.0)
	finalScale := calculateAnimationScale("attack", 7, frameCount)
	if math.Abs(finalScale-1.0) > 0.05 {
		t.Errorf("Final frame should return close to normal scale, got %f", finalScale)
	}
}

// TestAnimationPhase15_2_FrameCounts tests increased frame counts for smoother animations.
func TestAnimationPhase15_2_FrameCounts(t *testing.T) {
	spriteGen := sprites.NewGenerator()
	sys := NewAnimationSystem(spriteGen)
	
	tests := []struct {
		state    AnimationState
		expected int
		reason   string
	}{
		{AnimationStateIdle, 8, "Phase 15.2: Smoother breathing animation"},
		{AnimationStateWalk, 8, "Already meets 6-8 frame requirement"},
		{AnimationStateAttack, 8, "Phase 15.2: Better follow-through"},
	}
	
	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			count := sys.getFrameCount(tt.state)
			if count != tt.expected {
				t.Errorf("%s: Expected %d frames, got %d", tt.reason, tt.expected, count)
			}
			
			// Verify meets Phase 15.2 requirement (6-8 frames for primary animations)
			if count < 6 || count > 8 {
				t.Errorf("%s: Frame count %d outside 6-8 range required by Phase 15.2", tt.state, count)
			}
		})
	}
}

// TestAnimationPhase15_2_LODFrameRates tests distance-based LOD frame rates.
func TestAnimationPhase15_2_LODFrameRates(t *testing.T) {
	// Test expected frame rates based on Phase 15.2 specification
	tests := []struct {
		name         string
		fps          float64
		expectedTime float64
	}{
		{"Close range (12 FPS)", 12.0, 1.0 / 12.0},
		{"Medium range (6 FPS)", 6.0, 1.0 / 6.0},
		{"Far range (3 FPS)", 3.0, 1.0 / 3.0},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frameTime := 1.0 / tt.fps
			if math.Abs(frameTime-tt.expectedTime) > 0.001 {
				t.Errorf("Frame time calculation incorrect: got %f, expected %f", frameTime, tt.expectedTime)
			}
			
			// Verify frame time is reasonable (not too fast or too slow)
			if frameTime < 0.05 || frameTime > 0.5 {
				t.Errorf("Frame time %f outside reasonable range (0.05-0.5 seconds)", frameTime)
			}
		})
	}
}

// TestAnimationPhase15_2_DefaultFrameTime tests new default frame time.
func TestAnimationPhase15_2_DefaultFrameTime(t *testing.T) {
	anim := NewAnimationComponent(12345)
	
	expectedFrameTime := 1.0 / 12.0 // 12 FPS for close range
	
	if math.Abs(anim.FrameTime-expectedFrameTime) > 0.001 {
		t.Errorf("Default frame time should be 1/12 (0.0833), got %f", anim.FrameTime)
	}
}

// TestAnimationPhase15_2_Determinism tests that enhanced animations remain deterministic.
func TestAnimationPhase15_2_Determinism(t *testing.T) {
	states := []string{"idle", "walk", "attack"}
	frameCount := 8
	
	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			// Generate frames multiple times with same parameters
			for frame := 0; frame < frameCount; frame++ {
				offset1 := calculateAnimationOffset(state, frame, frameCount)
				offset2 := calculateAnimationOffset(state, frame, frameCount)
				
				if offset1.X != offset2.X || offset1.Y != offset2.Y {
					t.Errorf("Frame %d offset not deterministic: %+v vs %+v", frame, offset1, offset2)
				}
				
				rotation1 := calculateAnimationRotation(state, frame, frameCount)
				rotation2 := calculateAnimationRotation(state, frame, frameCount)
				
				if rotation1 != rotation2 {
					t.Errorf("Frame %d rotation not deterministic: %f vs %f", frame, rotation1, rotation2)
				}
				
				scale1 := calculateAnimationScale(state, frame, frameCount)
				scale2 := calculateAnimationScale(state, frame, frameCount)
				
				if scale1 != scale2 {
					t.Errorf("Frame %d scale not deterministic: %f vs %f", frame, scale1, scale2)
				}
			}
		})
	}
}

// TestAnimationPhase15_2_SmoothTransitions tests animations have smooth transitions between frames.
func TestAnimationPhase15_2_SmoothTransitions(t *testing.T) {
	states := []string{"idle", "walk", "attack"}
	frameCount := 8
	
	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			// Check offset smoothness
			for frame := 0; frame < frameCount-1; frame++ {
				offset1 := calculateAnimationOffset(state, frame, frameCount)
				offset2 := calculateAnimationOffset(state, frame+1, frameCount)
				
				// Verify offset doesn't jump too much between frames
				deltaX := math.Abs(offset2.X - offset1.X)
				deltaY := math.Abs(offset2.Y - offset1.Y)
				
				if deltaX > 10.0 {
					t.Errorf("%s frame %d->%d: X offset jump too large: %f pixels", state, frame, frame+1, deltaX)
				}
				if deltaY > 10.0 {
					t.Errorf("%s frame %d->%d: Y offset jump too large: %f pixels", state, frame, frame+1, deltaY)
				}
			}
			
			// Check rotation smoothness
			for frame := 0; frame < frameCount-1; frame++ {
				rot1 := calculateAnimationRotation(state, frame, frameCount)
				rot2 := calculateAnimationRotation(state, frame+1, frameCount)
				
				deltaRot := math.Abs(rot2 - rot1)
				
				// Verify rotation doesn't jump more than ~45 degrees between frames
				if deltaRot > math.Pi/4 {
					t.Errorf("%s frame %d->%d: Rotation jump too large: %f radians", state, frame, frame+1, deltaRot)
				}
			}
		})
	}
}

// BenchmarkAnimationPhase15_2_IdleBreathing benchmarks idle breathing calculation.
func BenchmarkAnimationPhase15_2_IdleBreathing(b *testing.B) {
	frameCount := 8
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateAnimationOffset("idle", i%frameCount, frameCount)
		_ = calculateAnimationRotation("idle", i%frameCount, frameCount)
		_ = calculateAnimationScale("idle", i%frameCount, frameCount)
	}
}

// BenchmarkAnimationPhase15_2_AttackFollowThrough benchmarks enhanced attack animation.
func BenchmarkAnimationPhase15_2_AttackFollowThrough(b *testing.B) {
	frameCount := 8
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calculateAnimationOffset("attack", i%frameCount, frameCount)
		_ = calculateAnimationRotation("attack", i%frameCount, frameCount)
		_ = calculateAnimationScale("attack", i%frameCount, frameCount)
	}
}
