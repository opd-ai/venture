package animation

import "math"

// BodyPart represents a specific part of an entity's body for articulation.
type BodyPart int

const (
	BodyPartHead BodyPart = iota
	BodyPartTorso
	BodyPartLeftArm
	BodyPartRightArm
	BodyPartLeftLeg
	BodyPartRightLeg
	BodyPartTail // For quadrupeds
)

// Animation multiplier constants for consistency and maintainability.
const (
	// armSwingMultiplier controls the horizontal arm swing during directional walking.
	armSwingMultiplier = 2.0

	// runningHeadRotationMultiplier amplifies head rotation during running.
	runningHeadRotationMultiplier = 1.2

	// runningArmSwingMultiplier amplifies arm swing during running.
	runningArmSwingMultiplier = 1.3
)

// String returns the string representation of the body part.
func (b BodyPart) String() string {
	switch b {
	case BodyPartHead:
		return "head"
	case BodyPartTorso:
		return "torso"
	case BodyPartLeftArm:
		return "left_arm"
	case BodyPartRightArm:
		return "right_arm"
	case BodyPartLeftLeg:
		return "left_leg"
	case BodyPartRightLeg:
		return "right_leg"
	case BodyPartTail:
		return "tail"
	default:
		return "unknown"
	}
}

// ArticulationOffset represents the positional offset of a body part.
type ArticulationOffset struct {
	X        float64 // Horizontal offset in pixels
	Y        float64 // Vertical offset in pixels
	Rotation float64 // Rotation in radians
}

// Articulation holds articulation data for all body parts in a frame.
type Articulation struct {
	Head     ArticulationOffset
	Torso    ArticulationOffset
	LeftArm  ArticulationOffset
	RightArm ArticulationOffset
	LeftLeg  ArticulationOffset
	RightLeg ArticulationOffset
	Tail     ArticulationOffset // Optional, for quadrupeds
}

// ArticulationConfig defines constraints for body part movement.
type ArticulationConfig struct {
	// Maximum pixel offsets for each body part
	ArmOffsetMax  float64 // Default: 3px
	LegOffsetMax  float64 // Default: 4px
	HeadOffsetMax float64 // Default: 2px
	TailOffsetMax float64 // Default: 5px

	// Rotation limits in radians
	ArmRotationMax  float64 // Default: 0.3 radians (~17 degrees)
	LegRotationMax  float64 // Default: 0.4 radians (~23 degrees)
	HeadRotationMax float64 // Default: 0.2 radians (~11 degrees)
	TailRotationMax float64 // Default: 0.5 radians (~29 degrees)
}

// DefaultArticulationConfig returns the default articulation configuration
// scaled for 64×64 sprites (Phase 45 migration).
// Original Phase 46 values (arms ±3px, legs ±4px) are doubled for 64×64.
func DefaultArticulationConfig() ArticulationConfig {
	return ArticulationConfig{
		ArmOffsetMax:    6.0,  // Scaled from 3px (32×32) to 6px (64×64)
		LegOffsetMax:    8.0,  // Scaled from 4px (32×32) to 8px (64×64)
		HeadOffsetMax:   4.0,  // Scaled from 2px (32×32) to 4px (64×64)
		TailOffsetMax:   10.0, // Scaled from 5px (32×32) to 10px (64×64)
		ArmRotationMax:  0.3,
		LegRotationMax:  0.4,
		HeadRotationMax: 0.2,
		TailRotationMax: 0.5,
	}
}

// CalculateArticulation computes body part articulation for an animation frame.
// state: animation state (idle, walk, run, attack, etc.)
// frameIndex: current frame number (0-7 for 8-frame animation)
// frameCount: total number of frames in animation
// direction: 8-directional facing
// config: articulation constraints
func CalculateArticulation(state string, frameIndex, frameCount int, direction Direction8, config ArticulationConfig) Articulation {
	t := float64(frameIndex) / float64(frameCount)
	art := Articulation{}

	switch state {
	case "idle":
		art = calculateIdleArticulation(t, config)
	case "walk":
		art = calculateWalkArticulation(t, direction, config)
	case "run":
		art = calculateRunArticulation(t, direction, config)
	case "attack":
		art = calculateAttackArticulation(t, direction, config)
	case "hit":
		art = calculateHitArticulation(t, config)
	case "death":
		art = calculateDeathArticulation(t, config)
	case "cast":
		art = calculateCastArticulation(t, config)
	case "jump":
		art = calculateJumpArticulation(t, config)
	default:
		// No articulation for unknown states
	}

	return art
}

// calculateDirectionalHeadRotation returns subtle head rotation based on facing direction.
// This provides visual indication of direction in top-down 64×64 sprites.
func calculateDirectionalHeadRotation(direction Direction8, config ArticulationConfig) float64 {
	// Subtle head rotation toward the direction of movement
	// Max rotation is about half the configured head rotation max
	maxRotation := config.HeadRotationMax * 0.5

	switch direction {
	case Dir8NorthWest:
		return -maxRotation * 0.7 // Turn head slightly left/up
	case Dir8North:
		return 0 // Facing up, no rotation needed
	case Dir8NorthEast:
		return maxRotation * 0.7 // Turn head slightly right/up
	case Dir8West:
		return -maxRotation // Turn head left
	case Dir8East:
		return maxRotation // Turn head right
	case Dir8SouthWest:
		return -maxRotation * 0.5 // Slight left turn
	case Dir8South:
		return 0 // Facing camera, no rotation
	case Dir8SouthEast:
		return maxRotation * 0.5 // Slight right turn
	default:
		return 0
	}
}

// calculateDirectionalArmOffsets returns X offsets for arms based on direction.
// Arms extend toward the movement direction for top-down visual clarity.
func calculateDirectionalArmOffsets(direction Direction8, armCycle float64, config ArticulationConfig) (leftArmX, rightArmX float64) {
	// Base arm offset scaled for 64×64 sprites
	baseOffset := config.ArmOffsetMax * 0.5

	switch direction {
	case Dir8West:
		// Moving left: leading arm extends left, trailing arm back
		leftArmX = -baseOffset - armCycle*armSwingMultiplier
		rightArmX = baseOffset + armCycle*armSwingMultiplier
	case Dir8East:
		// Moving right: leading arm extends right, trailing arm back
		leftArmX = baseOffset - armCycle*armSwingMultiplier
		rightArmX = -baseOffset + armCycle*armSwingMultiplier
	case Dir8North:
		// Moving up: arms spread slightly with swing
		leftArmX = -baseOffset * 0.5
		rightArmX = baseOffset * 0.5
	case Dir8South:
		// Moving down: arms spread slightly with swing
		leftArmX = -baseOffset * 0.5
		rightArmX = baseOffset * 0.5
	case Dir8NorthWest:
		leftArmX = -baseOffset * 0.8
		rightArmX = baseOffset * 0.3
	case Dir8NorthEast:
		leftArmX = -baseOffset * 0.3
		rightArmX = baseOffset * 0.8
	case Dir8SouthWest:
		leftArmX = -baseOffset * 0.7
		rightArmX = baseOffset * 0.4
	case Dir8SouthEast:
		leftArmX = -baseOffset * 0.4
		rightArmX = baseOffset * 0.7
	default:
		leftArmX = 0
		rightArmX = 0
	}

	return leftArmX, rightArmX
}

// calculateIdleArticulation creates subtle breathing animation.
// Scaled for 64×64 sprites (Phase 45 migration).
func calculateIdleArticulation(t float64, config ArticulationConfig) Articulation {
	breathCycle := math.Sin(t * 2 * math.Pi)

	return Articulation{
		Head: ArticulationOffset{
			Y:        breathCycle * 1.0,  // Subtle vertical head bob (scaled 2× for 64×64)
			Rotation: breathCycle * 0.02, // Tiny head tilt
		},
		Torso: ArticulationOffset{
			Y: breathCycle * 1.6, // Chest expansion/contraction (scaled 2× for 64×64)
		},
		LeftArm: ArticulationOffset{
			Y: breathCycle * 0.6, // Arms move slightly with breathing (scaled 2× for 64×64)
		},
		RightArm: ArticulationOffset{
			Y: breathCycle * 0.6,
		},
	}
}

// calculateWalkArticulation creates natural walking motion with arm and leg swing.
// Enhanced with directional head rotation and arm position changes for 64×64 sprites.
func calculateWalkArticulation(t float64, direction Direction8, config ArticulationConfig) Articulation {
	// Walking cycle: 8 frames for full left-right-left-right stride
	leftLegCycle := math.Sin(t * 2 * math.Pi)
	rightLegCycle := -leftLegCycle // Opposite phase
	armCycle := math.Sin(t * 2 * math.Pi)

	// Adjust amplitudes based on direction (side views show more motion)
	amplitudeMod := 1.0
	if direction == Dir8West || direction == Dir8East {
		amplitudeMod = 1.2 // Side view shows more articulation
	} else if direction == Dir8North || direction == Dir8South {
		amplitudeMod = 0.8 // Front/back view shows less articulation
	}

	// Calculate directional head rotation for direction indication
	headRotation := calculateDirectionalHeadRotation(direction, config)

	// Calculate arm X offsets based on direction for top-down visual clarity
	leftArmXOffset, rightArmXOffset := calculateDirectionalArmOffsets(direction, armCycle, config)

	return Articulation{
		Head: ArticulationOffset{
			Y:        math.Sin(t*4*math.Pi) * 1.0, // Head bobs twice per stride (scaled 2× for 64×64)
			Rotation: headRotation,                // Directional head rotation for direction indication
		},
		Torso: ArticulationOffset{
			Y:        math.Sin(t*4*math.Pi) * 1.6, // Torso bob (scaled 2× for 64×64)
			Rotation: math.Sin(t*2*math.Pi) * 0.05 * amplitudeMod, // Subtle torso rotation
		},
		LeftArm: ArticulationOffset{
			X:        leftArmXOffset, // Directional arm positioning
			Y:        armCycle * config.ArmOffsetMax * amplitudeMod,
			Rotation: armCycle * config.ArmRotationMax * amplitudeMod,
		},
		RightArm: ArticulationOffset{
			X:        rightArmXOffset, // Directional arm positioning
			Y:        -armCycle * config.ArmOffsetMax * amplitudeMod,
			Rotation: -armCycle * config.ArmRotationMax * amplitudeMod,
		},
		LeftLeg: ArticulationOffset{
			Y:        leftLegCycle * config.LegOffsetMax * amplitudeMod,
			Rotation: leftLegCycle * config.LegRotationMax * amplitudeMod,
		},
		RightLeg: ArticulationOffset{
			Y:        rightLegCycle * config.LegOffsetMax * amplitudeMod,
			Rotation: rightLegCycle * config.LegRotationMax * amplitudeMod,
		},
	}
}

// calculateRunArticulation creates exaggerated running motion.
// Enhanced with directional head rotation and arm position changes for 64×64 sprites.
func calculateRunArticulation(t float64, direction Direction8, config ArticulationConfig) Articulation {
	// Running has more exaggerated motion than walking (1.5x amplitude)
	amplitudeMod := 1.5
	if direction == Dir8West || direction == Dir8East {
		amplitudeMod = 1.8
	} else if direction == Dir8North || direction == Dir8South {
		amplitudeMod = 1.2
	}

	leftLegCycle := math.Sin(t * 2 * math.Pi)
	rightLegCycle := -leftLegCycle
	armCycle := math.Sin(t * 2 * math.Pi)

	// Calculate directional head rotation (more pronounced when running)
	headRotation := calculateDirectionalHeadRotation(direction, config) * runningHeadRotationMultiplier

	// Calculate arm X offsets based on direction (exaggerated for running)
	leftArmXOffset, rightArmXOffset := calculateDirectionalArmOffsets(direction, armCycle, config)
	leftArmXOffset *= runningArmSwingMultiplier  // More pronounced arm swing when running
	rightArmXOffset *= runningArmSwingMultiplier

	return Articulation{
		Head: ArticulationOffset{
			Y:        math.Sin(t*4*math.Pi) * 2.0, // Head bob (scaled 2× for 64×64)
			Rotation: headRotation + math.Sin(t*2*math.Pi)*0.08, // Directional + dynamic rotation
		},
		Torso: ArticulationOffset{
			Y:        math.Sin(t*4*math.Pi) * 3.0, // Torso bob (scaled 2× for 64×64)
			Rotation: math.Sin(t*2*math.Pi) * 0.1 * amplitudeMod,
		},
		LeftArm: ArticulationOffset{
			X:        leftArmXOffset, // Directional arm positioning
			Y:        armCycle * config.ArmOffsetMax * amplitudeMod,
			Rotation: armCycle * config.ArmRotationMax * 1.5,
		},
		RightArm: ArticulationOffset{
			X:        rightArmXOffset, // Directional arm positioning
			Y:        -armCycle * config.ArmOffsetMax * amplitudeMod,
			Rotation: -armCycle * config.ArmRotationMax * 1.5,
		},
		LeftLeg: ArticulationOffset{
			Y:        leftLegCycle * config.LegOffsetMax * amplitudeMod,
			Rotation: leftLegCycle * config.LegRotationMax * 1.5,
		},
		RightLeg: ArticulationOffset{
			Y:        rightLegCycle * config.LegOffsetMax * amplitudeMod,
			Rotation: rightLegCycle * config.LegRotationMax * 1.5,
		},
	}
}

// calculateAttackArticulation creates attack motion with wind-up and follow-through.
// Scaled for 64×64 sprites (Phase 45 migration).
func calculateAttackArticulation(t float64, direction Direction8, config ArticulationConfig) Articulation {
	art := Articulation{}

	if t < 0.2 {
		// Wind-up phase (0-20%)
		windupT := t / 0.2
		art.RightArm = ArticulationOffset{
			X:        -windupT * 4.0, // Scaled 2× for 64×64
			Y:        -windupT * config.ArmOffsetMax,
			Rotation: -windupT * config.ArmRotationMax * 2,
		}
		art.Torso = ArticulationOffset{
			Rotation: -windupT * 0.15,
		}
	} else if t < 0.5 {
		// Strike phase (20-50%)
		strikeT := (t - 0.2) / 0.3
		art.RightArm = ArticulationOffset{
			X:        -4.0 + strikeT*20.0, // Scaled 2× for 64×64
			Y:        -config.ArmOffsetMax + strikeT*config.ArmOffsetMax*2,
			Rotation: -config.ArmRotationMax*2 + strikeT*config.ArmRotationMax*4,
		}
		art.LeftArm = ArticulationOffset{
			X: -strikeT * 3.0, // Scaled 2× for 64×64
			Y: strikeT * config.ArmOffsetMax * 0.5,
		}
		art.Torso = ArticulationOffset{
			Rotation: -0.15 + strikeT*0.35,
		}
	} else {
		// Follow-through phase (50-100%)
		followT := (t - 0.5) / 0.5
		easedT := 1.0 - (1.0-followT)*(1.0-followT) // Quadratic ease-out
		art.RightArm = ArticulationOffset{
			X:        16.0 - easedT*16.0, // Scaled 2× for 64×64
			Y:        config.ArmOffsetMax - easedT*config.ArmOffsetMax,
			Rotation: config.ArmRotationMax*2 - easedT*config.ArmRotationMax*2,
		}
		art.LeftArm = ArticulationOffset{
			X: -3.0 + easedT*3.0, // Scaled 2× for 64×64
			Y: config.ArmOffsetMax*0.5 - easedT*config.ArmOffsetMax*0.5,
		}
		art.Torso = ArticulationOffset{
			Rotation: 0.2 - easedT*0.2,
		}
	}

	return art
}

// calculateHitArticulation creates knockback reaction.
// Scaled for 64×64 sprites (Phase 45 migration).
func calculateHitArticulation(t float64, config ArticulationConfig) Articulation {
	// Recoil motion
	recoil := 1.0 - t

	return Articulation{
		Head: ArticulationOffset{
			X:        -recoil * config.HeadOffsetMax * 2,
			Rotation: recoil * config.HeadRotationMax,
		},
		Torso: ArticulationOffset{
			X:        -recoil * 6.0, // Scaled 2× for 64×64
			Rotation: -recoil * 0.15,
		},
		LeftArm: ArticulationOffset{
			X:        -recoil * config.ArmOffsetMax,
			Y:        recoil * config.ArmOffsetMax * 0.5,
			Rotation: recoil * config.ArmRotationMax,
		},
		RightArm: ArticulationOffset{
			X:        -recoil * config.ArmOffsetMax,
			Y:        -recoil * config.ArmOffsetMax * 0.5,
			Rotation: -recoil * config.ArmRotationMax,
		},
	}
}

// calculateDeathArticulation creates falling/collapsing motion.
// Scaled for 64×64 sprites (Phase 45 migration).
func calculateDeathArticulation(t float64, config ArticulationConfig) Articulation {
	// Collapse motion - scaled 2× for 64×64 sprites
	return Articulation{
		Head: ArticulationOffset{
			Y:        t * 16.0, // Scaled 2× for 64×64
			Rotation: t * math.Pi / 4, // 45 degree rotation
		},
		Torso: ArticulationOffset{
			Y:        t * 24.0, // Scaled 2× for 64×64
			Rotation: t * math.Pi / 3, // 60 degree rotation
		},
		LeftArm: ArticulationOffset{
			Y:        t * 20.0, // Scaled 2× for 64×64
			Rotation: t * math.Pi / 6,
		},
		RightArm: ArticulationOffset{
			Y:        t * 20.0, // Scaled 2× for 64×64
			Rotation: -t * math.Pi / 6,
		},
		LeftLeg: ArticulationOffset{
			Y:        t * 30.0, // Scaled 2× for 64×64
			Rotation: t * config.LegRotationMax * 2,
		},
		RightLeg: ArticulationOffset{
			Y:        t * 30.0, // Scaled 2× for 64×64
			Rotation: -t * config.LegRotationMax * 2,
		},
	}
}

// calculateCastArticulation creates spell-casting gesture.
// Scaled for 64×64 sprites (Phase 45 migration).
func calculateCastArticulation(t float64, config ArticulationConfig) Articulation {
	castCycle := math.Sin(t * 2 * math.Pi)

	return Articulation{
		Head: ArticulationOffset{
			Y:        -2.0, // Head slightly up (concentration) - scaled 2× for 64×64
			Rotation: castCycle * 0.05,
		},
		Torso: ArticulationOffset{
			Rotation: castCycle * 0.08,
		},
		LeftArm: ArticulationOffset{
			Y:        -config.ArmOffsetMax, // Arm raised
			Rotation: -config.ArmRotationMax * 1.5,
		},
		RightArm: ArticulationOffset{
			Y:        -config.ArmOffsetMax * 1.2, // Other arm raised higher
			Rotation: castCycle * config.ArmRotationMax,
		},
	}
}

// calculateJumpArticulation creates jumping motion with squash and stretch.
// Scaled for 64×64 sprites (Phase 45 migration).
func calculateJumpArticulation(t float64, config ArticulationConfig) Articulation {
	art := Articulation{}

	if t < 0.2 {
		// Crouch before jump
		crouchT := t / 0.2
		art.Torso = ArticulationOffset{
			Y: crouchT * 6.0, // Crouch down (scaled 2× for 64×64)
		}
		art.LeftLeg = ArticulationOffset{
			Y: crouchT * config.LegOffsetMax * 1.5,
		}
		art.RightLeg = ArticulationOffset{
			Y: crouchT * config.LegOffsetMax * 1.5,
		}
	} else if t < 0.8 {
		// Jump arc
		jumpT := (t - 0.2) / 0.6
		arc := -4.0 * (jumpT - jumpT*jumpT) // Parabolic arc
		art.Torso = ArticulationOffset{
			Y: arc * 30.0, // Scaled 2× for 64×64
		}
		art.LeftArm = ArticulationOffset{
			Y:        -arc * config.ArmOffsetMax,
			Rotation: -jumpT * config.ArmRotationMax,
		}
		art.RightArm = ArticulationOffset{
			Y:        -arc * config.ArmOffsetMax,
			Rotation: jumpT * config.ArmRotationMax,
		}
		art.LeftLeg = ArticulationOffset{
			Y:        arc * config.LegOffsetMax * 2,
			Rotation: jumpT * config.LegRotationMax,
		}
		art.RightLeg = ArticulationOffset{
			Y:        arc * config.LegOffsetMax * 2,
			Rotation: -jumpT * config.LegRotationMax,
		}
	} else {
		// Landing
		landT := (t - 0.8) / 0.2
		art.Torso = ArticulationOffset{
			Y: landT * 4.0, // Slight compression on landing (scaled 2× for 64×64)
		}
		art.LeftLeg = ArticulationOffset{
			Y: landT * config.LegOffsetMax,
		}
		art.RightLeg = ArticulationOffset{
			Y: landT * config.LegOffsetMax,
		}
	}

	return art
}
