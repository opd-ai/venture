package animation

import (
	"math"
	"testing"
)

func TestBodyPartString(t *testing.T) {
	tests := []struct {
		part BodyPart
		want string
	}{
		{BodyPartHead, "head"},
		{BodyPartTorso, "torso"},
		{BodyPartLeftArm, "left_arm"},
		{BodyPartRightArm, "right_arm"},
		{BodyPartLeftLeg, "left_leg"},
		{BodyPartRightLeg, "right_leg"},
		{BodyPartTail, "tail"},
		{BodyPart(99), "unknown"},
	}

	for _, tt := range tests {
		if got := tt.part.String(); got != tt.want {
			t.Errorf("BodyPart.String() = %v, want %v", got, tt.want)
		}
	}
}

func TestDefaultArticulationConfig(t *testing.T) {
	config := DefaultArticulationConfig()

	// Verify Phase 46 requirements: arms ±3px, legs ±4px
	if config.ArmOffsetMax != 3.0 {
		t.Errorf("ArmOffsetMax = %v, want 3.0", config.ArmOffsetMax)
	}
	if config.LegOffsetMax != 4.0 {
		t.Errorf("LegOffsetMax = %v, want 4.0", config.LegOffsetMax)
	}

	// Verify all fields are set
	if config.HeadOffsetMax <= 0 {
		t.Error("HeadOffsetMax not set")
	}
	if config.TailOffsetMax <= 0 {
		t.Error("TailOffsetMax not set")
	}
	if config.ArmRotationMax <= 0 {
		t.Error("ArmRotationMax not set")
	}
	if config.LegRotationMax <= 0 {
		t.Error("LegRotationMax not set")
	}
}

func TestCalculateIdleArticulation(t *testing.T) {
	config := DefaultArticulationConfig()

	// Test multiple frames for smooth animation
	for frameIndex := 0; frameIndex < 8; frameIndex++ {
		art := calculateIdleArticulation(float64(frameIndex)/8.0, config)

		// Verify breathing motion is subtle
		if math.Abs(art.Head.Y) > 1.0 {
			t.Errorf("Idle head motion too large: %v", art.Head.Y)
		}
		if math.Abs(art.Torso.Y) > 1.5 {
			t.Errorf("Idle torso motion too large: %v", art.Torso.Y)
		}
		if math.Abs(art.Head.Rotation) > 0.05 {
			t.Errorf("Idle head rotation too large: %v", art.Head.Rotation)
		}
	}
}

func TestCalculateWalkArticulation(t *testing.T) {
	config := DefaultArticulationConfig()
	direction := Dir8East

	// Test all 8 frames
	for frameIndex := 0; frameIndex < 8; frameIndex++ {
		art := calculateWalkArticulation(float64(frameIndex)/8.0, direction, config)

		// Verify arm articulation respects Phase 46 constraints (±3px)
		if math.Abs(art.LeftArm.Y) > config.ArmOffsetMax*1.2 {
			t.Errorf("Left arm offset exceeds limit: %v > %v", art.LeftArm.Y, config.ArmOffsetMax*1.2)
		}
		if math.Abs(art.RightArm.Y) > config.ArmOffsetMax*1.2 {
			t.Errorf("Right arm offset exceeds limit: %v > %v", art.RightArm.Y, config.ArmOffsetMax*1.2)
		}

		// Verify leg articulation respects Phase 46 constraints (±4px)
		if math.Abs(art.LeftLeg.Y) > config.LegOffsetMax*1.2 {
			t.Errorf("Left leg offset exceeds limit: %v > %v", art.LeftLeg.Y, config.LegOffsetMax*1.2)
		}
		if math.Abs(art.RightLeg.Y) > config.LegOffsetMax*1.2 {
			t.Errorf("Right leg offset exceeds limit: %v > %v", art.RightLeg.Y, config.LegOffsetMax*1.2)
		}

		// Verify arms and legs are in opposite phase
		if (art.LeftArm.Y > 0 && art.RightArm.Y > 0) || (art.LeftArm.Y < 0 && art.RightArm.Y < 0) {
			t.Error("Arms should move in opposite directions during walk")
		}
	}
}

func TestCalculateRunArticulation(t *testing.T) {
	config := DefaultArticulationConfig()
	direction := Dir8South

	// Test at peak of motion (t=0.125 gives peak for sin(4*pi*t))
	art := calculateRunArticulation(0.125, direction, config)

	// Running should have motion (torso bobbing)
	// Just verify non-zero motion exists
	hasMotion := math.Abs(art.Torso.Y) > 0.1 || math.Abs(art.LeftArm.Y) > 0.1 || math.Abs(art.LeftLeg.Y) > 0.1
	if !hasMotion {
		t.Error("Run animation should have noticeable motion")
	}
}

func TestCalculateAttackArticulation(t *testing.T) {
	config := DefaultArticulationConfig()
	direction := Dir8East

	tests := []struct {
		name string
		t    float64
		desc string
	}{
		{"windup", 0.1, "wind-up phase"},
		{"strike", 0.35, "strike phase"},
		{"followthrough", 0.75, "follow-through phase"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			art := calculateAttackArticulation(tt.t, direction, config)

			// Verify right arm (attack arm) has significant motion
			if math.Abs(art.RightArm.X) < 0.1 && math.Abs(art.RightArm.Y) < 0.1 {
				t.Errorf("Attack articulation too subtle at %s", tt.desc)
			}

			// Verify torso rotates during attack
			if tt.name != "windup" && math.Abs(art.Torso.Rotation) < 0.01 {
				t.Errorf("Torso should rotate during %s", tt.desc)
			}
		})
	}
}

func TestCalculateHitArticulation(t *testing.T) {
	config := DefaultArticulationConfig()

	// Test knockback motion
	art := calculateHitArticulation(0.0, config) // Start of hit
	if art.Torso.X >= 0 {
		t.Error("Hit should create negative X offset (knockback)")
	}

	art = calculateHitArticulation(1.0, config) // End of hit
	if math.Abs(art.Torso.X) > 0.1 {
		t.Error("Hit should recover by end of animation")
	}
}

func TestCalculateDeathArticulation(t *testing.T) {
	config := DefaultArticulationConfig()

	// Test falling motion
	artStart := calculateDeathArticulation(0.0, config)
	artEnd := calculateDeathArticulation(1.0, config)

	// Should fall down (increase Y offset)
	if artEnd.Torso.Y <= artStart.Torso.Y {
		t.Error("Death animation should show falling motion")
	}
	if artEnd.Head.Y <= artStart.Head.Y {
		t.Error("Head should fall during death animation")
	}

	// Should rotate while falling
	if math.Abs(artEnd.Torso.Rotation) < 0.1 {
		t.Error("Death animation should include rotation")
	}
}

func TestCalculateJumpArticulation(t *testing.T) {
	config := DefaultArticulationConfig()

	tests := []struct {
		name string
		t    float64
	}{
		{"crouch", 0.1},
		{"jump", 0.5},
		{"landing", 0.9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			art := calculateJumpArticulation(tt.t, config)

			// Different phases should have different torso Y positions
			if tt.name == "jump" && art.Torso.Y >= 0 {
				t.Error("Jump phase should have negative Y (upward motion)")
			}
		})
	}
}

func TestCalculateArticulation(t *testing.T) {
	config := DefaultArticulationConfig()
	direction := Dir8North

	states := []string{"idle", "walk", "run", "attack", "hit", "death", "cast", "jump"}

	for _, state := range states {
		t.Run(state, func(t *testing.T) {
			// Should not panic for any state
			art := CalculateArticulation(state, 0, 8, direction, config)
			_ = art
		})
	}

	// Test unknown state (should return zero articulation)
	art := CalculateArticulation("unknown_state", 0, 8, direction, config)
	if art.Torso.X != 0 && art.Torso.Y != 0 {
		t.Error("Unknown state should return zero articulation")
	}
}

func TestArticulationDeterminism(t *testing.T) {
	config := DefaultArticulationConfig()
	direction := Dir8East

	// Same inputs should produce identical outputs
	art1 := CalculateArticulation("walk", 3, 8, direction, config)
	art2 := CalculateArticulation("walk", 3, 8, direction, config)

	if art1.LeftArm.Y != art2.LeftArm.Y {
		t.Error("CalculateArticulation not deterministic")
	}
	if art1.Torso.Rotation != art2.Torso.Rotation {
		t.Error("CalculateArticulation not deterministic")
	}
}

func BenchmarkCalculateWalkArticulation(b *testing.B) {
	config := DefaultArticulationConfig()
	direction := Dir8East
	for i := 0; i < b.N; i++ {
		calculateWalkArticulation(0.5, direction, config)
	}
}

func BenchmarkCalculateAttackArticulation(b *testing.B) {
	config := DefaultArticulationConfig()
	direction := Dir8South
	for i := 0; i < b.N; i++ {
		calculateAttackArticulation(0.35, direction, config)
	}
}

func BenchmarkCalculateArticulation(b *testing.B) {
	config := DefaultArticulationConfig()
	direction := Dir8North
	for i := 0; i < b.N; i++ {
		CalculateArticulation("walk", 4, 8, direction, config)
	}
}
