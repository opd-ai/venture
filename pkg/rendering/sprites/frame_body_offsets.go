// Package sprites provides per-frame body part positional offsets for animated
// top-down entity sprites. Instead of applying uniform geometric transforms to
// a static base sprite, each animation frame receives per-body-part (dx, dy)
// and scale adjustments so that legs visibly alternate during walk, arms extend
// during attack, and the torso subtly expands/contracts during idle breathing.
// All calculations are deterministic given (state, frameIndex, frameCount).
package sprites

import (
	"math"
)

// FramePartOffset describes the positional and scale adjustment for a single
// body part in a single animation frame. Values are in normalized sprite
// coordinates (0.0–1.0 range, relative to sprite width/height).
type FramePartOffset struct {
	DX    float64 // Horizontal shift (fraction of sprite width)
	DY    float64 // Vertical shift (fraction of sprite height)
	Scale float64 // Uniform scale multiplier (1.0 = unchanged)
}

// FrameOffsetMap maps BodyPart → FramePartOffset for one animation frame.
type FrameOffsetMap map[BodyPart]FramePartOffset

// ComputeFrameOffsets returns per-body-part positional offsets for the given
// animation state and frame. frameIndex is 0-based, frameCount is total frames
// in the cycle (typically 8). The returned map only contains entries for body
// parts that actually move; absent parts should render at their default position.
func ComputeFrameOffsets(state string, frameIndex, frameCount int) FrameOffsetMap {
	if frameCount <= 0 {
		return nil
	}
	t := float64(frameIndex) / float64(frameCount) // 0.0 – ~0.875 for 8 frames
	switch state {
	case "idle":
		return idleOffsets(t)
	case "walk":
		return walkOffsets(t)
	case "run":
		return runOffsets(t)
	case "attack":
		return attackOffsets(t)
	case "cast":
		return castOffsets(t)
	case "hit":
		return hitOffsets(t)
	case "death":
		return deathOffsets(t, frameIndex, frameCount)
	default:
		return idleOffsets(t)
	}
}

// --- State-specific offset generators ---

// idleOffsets produces subtle breathing: torso scale oscillation, slight head bob.
func idleOffsets(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	breath := math.Sin(phase)

	return FrameOffsetMap{
		PartTorso: {
			DX:    0,
			DY:    breath * 0.005,
			Scale: 1.0 + breath*0.015, // ±1.5% scale oscillation
		},
		PartHead: {
			DX:    0,
			DY:    breath * -0.008, // Head bobs opposite to torso expansion
			Scale: 1.0,
		},
		PartArms: {
			DX:    breath * 0.004, // Arms sway slightly outward
			DY:    breath * 0.003,
			Scale: 1.0,
		},
		PartShadow: {
			DX:    0,
			DY:    0,
			Scale: 1.0 + breath*0.01, // Shadow tracks torso size
		},
	}
}

// walkOffsets animates a top-down walking gait: legs alternate forward/backward,
// torso rotates slightly, arms swing opposite to legs.
func walkOffsets(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	legSwing := math.Sin(phase)
	armSwing := -legSwing // Arms oppose legs

	return FrameOffsetMap{
		PartLegs: {
			DX:    legSwing * 0.04, // Legs spread side-to-side from above
			DY:    math.Abs(legSwing) * -0.015,
			Scale: 1.0,
		},
		PartTorso: {
			DX:    0,
			DY:    math.Abs(math.Sin(phase*2)) * -0.01, // Subtle vertical bounce (2x freq)
			Scale: 1.0,
		},
		PartArms: {
			DX:    armSwing * 0.05, // Arms swing wider than legs
			DY:    armSwing * 0.02,
			Scale: 1.0,
		},
		PartHead: {
			DX:    legSwing * 0.008, // Tiny head sway
			DY:    math.Abs(math.Sin(phase*2)) * -0.01,
			Scale: 1.0,
		},
		PartShadow: {
			DX:    0,
			DY:    0,
			Scale: 1.0 + math.Abs(math.Sin(phase*2))*0.02,
		},
	}
}

// runOffsets: exaggerated walk with more bounce and wider strides.
func runOffsets(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	legSwing := math.Sin(phase)
	armSwing := -legSwing

	return FrameOffsetMap{
		PartLegs: {
			DX:    legSwing * 0.06,
			DY:    math.Abs(legSwing) * -0.025,
			Scale: 1.0 + math.Abs(legSwing)*0.03,
		},
		PartTorso: {
			DX:    0,
			DY:    math.Abs(math.Sin(phase*2)) * -0.02,
			Scale: 1.0 + math.Abs(math.Sin(phase*2))*0.01,
		},
		PartArms: {
			DX:    armSwing * 0.07,
			DY:    armSwing * 0.03,
			Scale: 1.0,
		},
		PartHead: {
			DX:    legSwing * 0.012,
			DY:    math.Abs(math.Sin(phase*2)) * -0.015,
			Scale: 1.0,
		},
		PartShadow: {
			DX:    0,
			DY:    0,
			Scale: 1.0 + math.Abs(math.Sin(phase*2))*0.03,
		},
	}
}

// attackOffsets: weapon arm lunges forward, body shifts in attack direction,
// torso slightly turns, legs brace.
func attackOffsets(t float64) FrameOffsetMap {
	// Three phases: wind-up (0–0.25), strike (0.25–0.5), recover (0.5–1.0)
	var armDX, armDY, torsoDX, legScale float64
	if t < 0.25 {
		// Wind-up: arm draws back
		p := t / 0.25
		armDX = -p * 0.06
		armDY = -p * 0.02
		torsoDX = -p * 0.015
		legScale = 1.0 + p*0.02
	} else if t < 0.5 {
		// Strike: arm thrusts forward
		p := (t - 0.25) / 0.25
		armDX = -0.06 + p*0.18 // From -0.06 to +0.12
		armDY = -0.02 + p*0.04
		torsoDX = -0.015 + p*0.05
		legScale = 1.02 - p*0.02
	} else {
		// Recover: return to neutral
		p := (t - 0.5) / 0.5
		ease := 1.0 - (1.0-p)*(1.0-p) // Quadratic ease out
		armDX = 0.12 * (1.0 - ease)
		armDY = 0.02 * (1.0 - ease)
		torsoDX = 0.035 * (1.0 - ease)
		legScale = 1.0
	}

	return FrameOffsetMap{
		PartArms: {
			DX:    armDX,
			DY:    armDY,
			Scale: 1.0,
		},
		PartTorso: {
			DX:    torsoDX,
			DY:    0,
			Scale: 1.0,
		},
		PartWeapon: {
			DX:    armDX * 1.2,
			DY:    armDY * 1.2,
			Scale: 1.0,
		},
		PartLegs: {
			DX:    0,
			DY:    0,
			Scale: legScale,
		},
		PartHead: {
			DX:    torsoDX * 0.5,
			DY:    0,
			Scale: 1.0,
		},
	}
}

// castOffsets: arms raise upward, body lifts slightly, magical "gathering" pose.
func castOffsets(t float64) FrameOffsetMap {
	// Gather (0–0.4), hold (0.4–0.7), release (0.7–1.0)
	var armDY, headDY, torsoScale float64
	if t < 0.4 {
		p := t / 0.4
		armDY = -p * 0.06 // Arms raise
		headDY = -p * 0.01
		torsoScale = 1.0 + p*0.02
	} else if t < 0.7 {
		armDY = -0.06
		headDY = -0.01
		torsoScale = 1.02
	} else {
		p := (t - 0.7) / 0.3
		ease := p * p // Quadratic ease in for snap release
		armDY = -0.06 + ease*0.08 // Arms thrust down past neutral
		headDY = -0.01 + ease*0.015
		torsoScale = 1.02 - ease*0.03
	}

	return FrameOffsetMap{
		PartArms: {
			DX:    0,
			DY:    armDY,
			Scale: 1.0 + math.Abs(armDY)*2, // Arms expand as they raise
		},
		PartHead: {
			DX:    0,
			DY:    headDY,
			Scale: 1.0,
		},
		PartTorso: {
			DX:    0,
			DY:    0,
			Scale: torsoScale,
		},
	}
}

// hitOffsets: entity recoils — body pushed back, head snaps, legs buckle.
func hitOffsets(t float64) FrameOffsetMap {
	// Impact (0–0.2), recoil (0.2–0.5), stagger (0.5–1.0)
	var bodyDX, headDX, legDY float64
	if t < 0.2 {
		p := t / 0.2
		bodyDX = -p * 0.08 // Pushed back
		headDX = -p * 0.10
		legDY = p * 0.015
	} else if t < 0.5 {
		p := (t - 0.2) / 0.3
		bodyDX = -0.08 + p*0.06
		headDX = -0.10 + p*0.08
		legDY = 0.015 * (1.0 - p*0.5)
	} else {
		p := (t - 0.5) / 0.5
		bodyDX = -0.02 * (1.0 - p)
		headDX = -0.02 * (1.0 - p)
		legDY = 0.0075 * (1.0 - p)
	}

	return FrameOffsetMap{
		PartTorso: {DX: bodyDX, DY: 0, Scale: 1.0},
		PartHead:  {DX: headDX, DY: 0, Scale: 1.0},
		PartArms:  {DX: bodyDX * 0.8, DY: 0, Scale: 1.0},
		PartLegs:  {DX: 0, DY: legDY, Scale: 1.0},
	}
}

// deathOffsets: entity collapses — progressive flattening and spreading.
func deathOffsets(t float64, frameIndex, frameCount int) FrameOffsetMap {
	// Smooth collapse: body widens and flattens, limbs spread
	collapse := t * t // Accelerating collapse
	finalFrame := frameIndex >= frameCount-1

	headScale := 1.0 - collapse*0.3
	torsoScaleX := 1.0 + collapse*0.2 // Widen
	torsoScaleY := 1.0 - collapse*0.4 // Flatten
	_ = torsoScaleY                    // Y scale applied via DY

	legSpread := collapse * 0.08

	headDY := collapse * 0.03
	torsoDY := collapse * 0.02

	if finalFrame {
		headScale = 0.7
		torsoDY = 0.02
	}

	return FrameOffsetMap{
		PartHead: {
			DX:    0,
			DY:    headDY,
			Scale: headScale,
		},
		PartTorso: {
			DX:    0,
			DY:    torsoDY,
			Scale: torsoScaleX,
		},
		PartArms: {
			DX:    collapse * 0.06,
			DY:    collapse * 0.02,
			Scale: 1.0 - collapse*0.2,
		},
		PartLegs: {
			DX:    legSpread,
			DY:    collapse * 0.01,
			Scale: 1.0 + collapse*0.1,
		},
		PartShadow: {
			DX:    0,
			DY:    0,
			Scale: 1.0 + collapse*0.3, // Shadow grows as entity spreads
		},
	}
}

// applyFrameOffsetsToSpec returns a modified copy of spec with frame-specific
// positional and scale adjustments applied. DX/DY are normalized fractions of
// sprite dimensions, added to the spec's RelativeX/RelativeY. Scale multiplies
// RelativeWidth and RelativeHeight uniformly.
func applyFrameOffsetsToSpec(spec PartSpec, part BodyPart, offsets FrameOffsetMap, spriteW, spriteH int) PartSpec {
	fo, ok := offsets[part]
	if !ok {
		return spec
	}
	mod := spec
	mod.RelativeX += fo.DX
	mod.RelativeY += fo.DY
	if fo.Scale != 0 && fo.Scale != 1.0 {
		mod.RelativeWidth *= fo.Scale
		mod.RelativeHeight *= fo.Scale
	}
	return mod
}
