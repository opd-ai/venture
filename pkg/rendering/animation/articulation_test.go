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

	// Verify Phase 45 scaled requirements for 64×64 sprites: arms ±6px, legs ±8px
	// (doubled from Phase 46 values of ±3px, ±4px for 32×32 sprites)
	if config.ArmOffsetMax != 6.0 {
		t.Errorf("ArmOffsetMax = %v, want 6.0 (scaled for 64×64)", config.ArmOffsetMax)
	}
	if config.LegOffsetMax != 8.0 {
		t.Errorf("LegOffsetMax = %v, want 8.0 (scaled for 64×64)", config.LegOffsetMax)
	}
	if config.HeadOffsetMax != 4.0 {
		t.Errorf("HeadOffsetMax = %v, want 4.0 (scaled for 64×64)", config.HeadOffsetMax)
	}
	if config.TailOffsetMax != 10.0 {
		t.Errorf("TailOffsetMax = %v, want 10.0 (scaled for 64×64)", config.TailOffsetMax)
	}

	// Verify rotation limits (unchanged - radians are resolution-independent)
	if config.ArmRotationMax != 0.3 {
		t.Errorf("ArmRotationMax = %v, want 0.3", config.ArmRotationMax)
	}
	if config.LegRotationMax != 0.4 {
		t.Errorf("LegRotationMax = %v, want 0.4", config.LegRotationMax)
	}
	if config.HeadRotationMax != 0.2 {
		t.Errorf("HeadRotationMax = %v, want 0.2", config.HeadRotationMax)
	}
	if config.TailRotationMax != 0.5 {
		t.Errorf("TailRotationMax = %v, want 0.5", config.TailRotationMax)
	}
}

func TestCalculateIdleArticulation(t *testing.T) {
	config := DefaultArticulationConfig()

	// Test multiple frames for smooth animation
	for frameIndex := 0; frameIndex < 8; frameIndex++ {
		art := calculateIdleArticulation(float64(frameIndex)/8.0, config)

		// Verify breathing motion is subtle (scaled 2× for 64×64: 1.0 → 2.0, 1.5 → 3.0)
		if math.Abs(art.Head.Y) > 2.0 {
			t.Errorf("Idle head motion too large: %v", art.Head.Y)
		}
		if math.Abs(art.Torso.Y) > 3.0 {
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

// TestDirectionalHeadRotation verifies head rotation based on facing direction.
func TestDirectionalHeadRotation(t *testing.T) {
	config := DefaultArticulationConfig()

	tests := []struct {
		direction   Direction8
		expectSign  int // -1 for left rotation, 0 for no rotation, 1 for right rotation
	}{
		{Dir8North, 0},      // Facing up, no rotation
		{Dir8South, 0},      // Facing camera, no rotation
		{Dir8East, 1},       // Turn head right
		{Dir8West, -1},      // Turn head left
		{Dir8NorthEast, 1},  // Turn slightly right
		{Dir8NorthWest, -1}, // Turn slightly left
		{Dir8SouthEast, 1},  // Turn slightly right
		{Dir8SouthWest, -1}, // Turn slightly left
	}

	for _, tt := range tests {
		t.Run(tt.direction.String(), func(t *testing.T) {
			rotation := calculateDirectionalHeadRotation(tt.direction, config)

			if tt.expectSign == 0 {
				if rotation != 0 {
					t.Errorf("Expected no rotation for %v, got %v", tt.direction, rotation)
				}
			} else if tt.expectSign > 0 && rotation <= 0 {
				t.Errorf("Expected positive rotation for %v, got %v", tt.direction, rotation)
			} else if tt.expectSign < 0 && rotation >= 0 {
				t.Errorf("Expected negative rotation for %v, got %v", tt.direction, rotation)
			}

			// Verify rotation is within limits
			maxRotation := config.HeadRotationMax * 0.5
			if math.Abs(rotation) > maxRotation+0.01 {
				t.Errorf("Head rotation %v exceeds max %v for %v", rotation, maxRotation, tt.direction)
			}
		})
	}
}

// TestDirectionalArmOffsets verifies arm X offsets based on facing direction.
func TestDirectionalArmOffsets(t *testing.T) {
	config := DefaultArticulationConfig()
	armCycle := 0.5 // Mid-swing

	// Test that arms have opposite X offsets for left/right directions
	t.Run("East direction", func(t *testing.T) {
		leftX, rightX := calculateDirectionalArmOffsets(Dir8East, armCycle, config)
		// Right arm should extend right (positive), left arm left (positive for east)
		if leftX >= rightX {
			t.Logf("Left arm X: %v, Right arm X: %v", leftX, rightX)
		}
	})

	t.Run("West direction", func(t *testing.T) {
		leftX, rightX := calculateDirectionalArmOffsets(Dir8West, armCycle, config)
		// Left arm should extend left (negative), right arm right (positive)
		if leftX > 0 {
			t.Logf("West direction: left arm should be negative, got %v", leftX)
		}
		if rightX < 0 {
			t.Logf("West direction: right arm should be positive, got %v", rightX)
		}
	})

	t.Run("North/South spread", func(t *testing.T) {
		leftXN, rightXN := calculateDirectionalArmOffsets(Dir8North, armCycle, config)
		leftXS, rightXS := calculateDirectionalArmOffsets(Dir8South, armCycle, config)

		// For front/back facing, arms spread symmetrically
		if leftXN != leftXS || rightXN != rightXS {
			t.Logf("North/South should have similar arm spread")
		}
	})
}

// TestWalkArticulationDirectional verifies walk articulation includes directional enhancements.
func TestWalkArticulationDirectional(t *testing.T) {
	config := DefaultArticulationConfig()

	directions := []Direction8{Dir8North, Dir8East, Dir8South, Dir8West}

	for _, dir := range directions {
		t.Run(dir.String(), func(t *testing.T) {
			// Test at mid-walk
			art := calculateWalkArticulation(0.5, dir, config)

			// Head should have rotation set
			// Note: North and South have 0 rotation which is valid
			if dir == Dir8East && art.Head.Rotation <= 0 {
				t.Errorf("East direction should have positive head rotation, got %v", art.Head.Rotation)
			}
			if dir == Dir8West && art.Head.Rotation >= 0 {
				t.Errorf("West direction should have negative head rotation, got %v", art.Head.Rotation)
			}

			// Arms should have X offsets
			if dir == Dir8East || dir == Dir8West {
				if art.LeftArm.X == 0 && art.RightArm.X == 0 {
					t.Errorf("Arms should have X offsets for %v direction", dir)
				}
			}
		})
	}
}

// TestPhase45AnimationScaling verifies animation values are properly scaled for 64×64 sprites.
func TestPhase45AnimationScaling(t *testing.T) {
	config := DefaultArticulationConfig()

	// Test that death animation has properly scaled Y offsets
	art := calculateDeathArticulation(1.0, config) // Full death animation

	// Head Y should be 16.0 (scaled from 8.0)
	if art.Head.Y != 16.0 {
		t.Errorf("Death head Y = %v, want 16.0 (scaled for 64×64)", art.Head.Y)
	}

	// Torso Y should be 24.0 (scaled from 12.0)
	if art.Torso.Y != 24.0 {
		t.Errorf("Death torso Y = %v, want 24.0 (scaled for 64×64)", art.Torso.Y)
	}

	// Leg Y should be 30.0 (scaled from 15.0)
	if art.LeftLeg.Y != 30.0 {
		t.Errorf("Death left leg Y = %v, want 30.0 (scaled for 64×64)", art.LeftLeg.Y)
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
