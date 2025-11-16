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
// as specified in Phase 46 (arms ±3px, legs ±4px).
func DefaultArticulationConfig() ArticulationConfig {
	return ArticulationConfig{
		ArmOffsetMax:    3.0,
		LegOffsetMax:    4.0,
		HeadOffsetMax:   2.0,
		TailOffsetMax:   5.0,
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

// calculateIdleArticulation creates subtle breathing animation.
func calculateIdleArticulation(t float64, config ArticulationConfig) Articulation {
	breathCycle := math.Sin(t * 2 * math.Pi)

	return Articulation{
		Head: ArticulationOffset{
			Y:        breathCycle * 0.5,  // Subtle vertical head bob
			Rotation: breathCycle * 0.02, // Tiny head tilt
		},
		Torso: ArticulationOffset{
			Y: breathCycle * 0.8, // Chest expansion/contraction
		},
		LeftArm: ArticulationOffset{
			Y: breathCycle * 0.3, // Arms move slightly with breathing
		},
		RightArm: ArticulationOffset{
			Y: breathCycle * 0.3,
		},
	}
}

// calculateWalkArticulation creates natural walking motion with arm and leg swing.
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

	return Articulation{
		Head: ArticulationOffset{
			Y: math.Sin(t*4*math.Pi) * 0.5, // Head bobs twice per stride
		},
		Torso: ArticulationOffset{
			Y:        math.Sin(t*4*math.Pi) * 0.8,
			Rotation: math.Sin(t*2*math.Pi) * 0.05 * amplitudeMod, // Subtle torso rotation
		},
		LeftArm: ArticulationOffset{
			Y:        armCycle * config.ArmOffsetMax * amplitudeMod,
			Rotation: armCycle * config.ArmRotationMax * amplitudeMod,
		},
		RightArm: ArticulationOffset{
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

	return Articulation{
		Head: ArticulationOffset{
			Y:        math.Sin(t*4*math.Pi) * 1.0,
			Rotation: math.Sin(t*2*math.Pi) * 0.08,
		},
		Torso: ArticulationOffset{
			Y:        math.Sin(t*4*math.Pi) * 1.5,
			Rotation: math.Sin(t*2*math.Pi) * 0.1 * amplitudeMod,
		},
		LeftArm: ArticulationOffset{
			Y:        armCycle * config.ArmOffsetMax * amplitudeMod,
			Rotation: armCycle * config.ArmRotationMax * 1.5,
		},
		RightArm: ArticulationOffset{
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
func calculateAttackArticulation(t float64, direction Direction8, config ArticulationConfig) Articulation {
	art := Articulation{}

	if t < 0.2 {
		// Wind-up phase (0-20%)
		windupT := t / 0.2
		art.RightArm = ArticulationOffset{
			X:        -windupT * 2.0,
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
			X:        -2.0 + strikeT*10.0,
			Y:        -config.ArmOffsetMax + strikeT*config.ArmOffsetMax*2,
			Rotation: -config.ArmRotationMax*2 + strikeT*config.ArmRotationMax*4,
		}
		art.LeftArm = ArticulationOffset{
			X: -strikeT * 1.5,
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
			X:        8.0 - easedT*8.0,
			Y:        config.ArmOffsetMax - easedT*config.ArmOffsetMax,
			Rotation: config.ArmRotationMax*2 - easedT*config.ArmRotationMax*2,
		}
		art.LeftArm = ArticulationOffset{
			X: -1.5 + easedT*1.5,
			Y: config.ArmOffsetMax*0.5 - easedT*config.ArmOffsetMax*0.5,
		}
		art.Torso = ArticulationOffset{
			Rotation: 0.2 - easedT*0.2,
		}
	}

	return art
}

// calculateHitArticulation creates knockback reaction.
func calculateHitArticulation(t float64, config ArticulationConfig) Articulation {
	// Recoil motion
	recoil := 1.0 - t

	return Articulation{
		Head: ArticulationOffset{
			X:        -recoil * config.HeadOffsetMax * 2,
			Rotation: recoil * config.HeadRotationMax,
		},
		Torso: ArticulationOffset{
			X:        -recoil * 3.0,
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
func calculateDeathArticulation(t float64, config ArticulationConfig) Articulation {
	// Collapse motion
	return Articulation{
		Head: ArticulationOffset{
			Y:        t * 8.0,
			Rotation: t * math.Pi / 4, // 45 degree rotation
		},
		Torso: ArticulationOffset{
			Y:        t * 12.0,
			Rotation: t * math.Pi / 3, // 60 degree rotation
		},
		LeftArm: ArticulationOffset{
			Y:        t * 10.0,
			Rotation: t * math.Pi / 6,
		},
		RightArm: ArticulationOffset{
			Y:        t * 10.0,
			Rotation: -t * math.Pi / 6,
		},
		LeftLeg: ArticulationOffset{
			Y:        t * 15.0,
			Rotation: t * config.LegRotationMax * 2,
		},
		RightLeg: ArticulationOffset{
			Y:        t * 15.0,
			Rotation: -t * config.LegRotationMax * 2,
		},
	}
}

// calculateCastArticulation creates spell-casting gesture.
func calculateCastArticulation(t float64, config ArticulationConfig) Articulation {
	castCycle := math.Sin(t * 2 * math.Pi)

	return Articulation{
		Head: ArticulationOffset{
			Y:        -1.0, // Head slightly up (concentration)
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
func calculateJumpArticulation(t float64, config ArticulationConfig) Articulation {
	art := Articulation{}

	if t < 0.2 {
		// Crouch before jump
		crouchT := t / 0.2
		art.Torso = ArticulationOffset{
			Y: crouchT * 3.0, // Crouch down
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
			Y: arc * 15.0,
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
			Y: landT * 2.0, // Slight compression on landing
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
