// Package sprites provides creature-type-specific per-frame body part offsets
// for animated top-down entity sprites. Nonhumanoid creatures (quadrupeds,
// serpents, arachnids, flying creatures, blobs, mechanical constructs, undead)
// receive movement patterns that match their body plan instead of defaulting
// to humanoid arm/leg swing. All calculations are deterministic given
// (entityType, state, frameIndex, frameCount).
package sprites

import (
	"math"
)

// ComputeCreatureFrameOffsets returns per-body-part positional offsets for the
// given animation state, frame, and creature type. For humanoid entities this
// delegates to the standard ComputeFrameOffsets. For nonhumanoid entities it
// selects creature-appropriate movement patterns that match the body plan
// defined in the aerial nonhumanoid templates.
func ComputeCreatureFrameOffsets(state string, frameIndex, frameCount int, entityType string) FrameOffsetMap {
	if frameCount <= 0 {
		return nil
	}
	if IsHumanoidEntity(entityType) {
		return ComputeFrameOffsets(state, frameIndex, frameCount)
	}
	t := float64(frameIndex) / float64(frameCount)

	switch creatureCategory(entityType) {
	case "quadruped":
		return quadrupedOffsets(state, t)
	case "serpentine":
		return serpentineOffsets(state, t)
	case "arachnid":
		return arachnidOffsets(state, t)
	case "flying":
		return flyingOffsets(state, t)
	case "blob":
		return blobOffsets(state, t)
	case "mechanical":
		return mechanicalOffsets(state, t)
	case "undead":
		return undeadOffsets(state, t)
	case "insect":
		return insectOffsets(state, t)
	case "multi_limbed":
		return multiLimbedOffsets(state, t)
	default:
		return ComputeFrameOffsets(state, frameIndex, frameCount)
	}
}

// creatureCategory maps entity type strings to broad creature categories.
func creatureCategory(entityType string) string {
	switch entityType {
	case "quadruped", "wolf", "bear", "animal", "beast", "horse":
		return "quadruped"
	case "serpentine", "snake", "worm", "tentacle", "wyrm":
		return "serpentine"
	case "arachnid", "spider", "scorpion":
		return "arachnid"
	case "insect", "beetle", "ant", "centipede", "mantis", "wasp", "moth", "crawler":
		return "insect"
	case "flying", "bird", "dragon", "bat", "wyvern":
		return "flying"
	case "blob", "slime", "amoeba", "ooze":
		return "blob"
	case "multi_limbed", "kraken", "octopus", "squid", "shoggoth", "abomination",
		"horror", "eldritch", "aberration", "hydra", "chimera":
		return "multi_limbed"
	case "mechanical", "robot", "golem", "construct", "android":
		return "mechanical"
	case "undead", "skeleton", "ghost", "zombie", "lich":
		return "undead"
	default:
		return ""
	}
}

// --- Quadruped offsets: trotting gait, body sway, tail wag ---

func quadrupedOffsets(state string, t float64) FrameOffsetMap {
	switch state {
	case "walk", "run":
		return quadrupedWalk(t, state == "run")
	case "attack":
		return quadrupedAttack(t)
	case "hit":
		return quadrupedHit(t)
	case "death":
		return quadrupedDeath(t)
	default:
		return quadrupedIdle(t)
	}
}

func quadrupedIdle(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	breath := math.Sin(phase)
	tailWag := math.Sin(phase * 2.5) // Tail wags faster than breathing

	return FrameOffsetMap{
		PartTorso: {
			DX:    0,
			DY:    breath * 0.004,
			Scale: 1.0 + breath*0.012,
		},
		PartHead: {
			DX:    breath * 0.006,
			DY:    breath * -0.005,
			Scale: 1.0,
		},
		PartLegs: {
			DX:    0,
			DY:    breath * 0.002,
			Scale: 1.0,
		},
		PartTail: {
			DX:    tailWag * 0.04,
			DY:    0,
			Scale: 1.0,
		},
		PartShadow: {
			DX:    0,
			DY:    0,
			Scale: 1.0 + breath*0.008,
		},
	}
}

func quadrupedWalk(t float64, running bool) FrameOffsetMap {
	phase := t * 2 * math.Pi
	amp := 1.0
	if running {
		amp = 1.6
	}
	// Diagonal trot: front-left+rear-right then front-right+rear-left
	legDiag := math.Sin(phase) * amp
	bodySway := math.Sin(phase) * 0.015 * amp
	bounce := math.Abs(math.Sin(phase*2)) * 0.012 * amp

	return FrameOffsetMap{
		PartLegs: {
			DX:    legDiag * 0.05,
			DY:    -bounce,
			Scale: 1.0 + math.Abs(legDiag)*0.02,
		},
		PartTorso: {
			DX:    bodySway,
			DY:    -bounce * 0.6,
			Scale: 1.0,
		},
		PartHead: {
			DX:    bodySway * 0.5,
			DY:    -bounce*0.8 - 0.005*amp,
			Scale: 1.0,
		},
		PartTail: {
			DX:    -bodySway * 2.5,
			DY:    math.Sin(phase*1.5) * 0.01,
			Scale: 1.0,
		},
		PartShadow: {
			DX:    0,
			DY:    0,
			Scale: 1.0 + bounce*0.5,
		},
	}
}

func quadrupedAttack(t float64) FrameOffsetMap {
	// Lunge: body crouches, then springs forward, head strikes
	var headDY, bodyDY, legScale float64
	if t < 0.3 {
		p := t / 0.3
		headDY = p * 0.02 // Head dips during crouch
		bodyDY = p * 0.015
		legScale = 1.0 + p*0.04 // Legs brace wide
	} else if t < 0.5 {
		p := (t - 0.3) / 0.2
		headDY = 0.02 - p*0.08 // Head lunges forward
		bodyDY = 0.015 - p*0.04
		legScale = 1.04 - p*0.02
	} else {
		p := (t - 0.5) / 0.5
		ease := 1.0 - (1.0-p)*(1.0-p)
		headDY = -0.06 * (1.0 - ease)
		bodyDY = -0.025 * (1.0 - ease)
		legScale = 1.02 - ease*0.02
	}

	return FrameOffsetMap{
		PartHead:  {DX: 0, DY: headDY, Scale: 1.0},
		PartTorso: {DX: 0, DY: bodyDY, Scale: 1.0},
		PartLegs:  {DX: 0, DY: 0, Scale: legScale},
		PartTail:  {DX: 0, DY: -bodyDY * 0.5, Scale: 1.0},
	}
}

func quadrupedHit(t float64) FrameOffsetMap {
	var bodyDX, headDX float64
	if t < 0.2 {
		p := t / 0.2
		bodyDX = p * 0.06
		headDX = p * 0.08
	} else {
		p := (t - 0.2) / 0.8
		ease := 1.0 - (1.0-p)*(1.0-p)
		bodyDX = 0.06 * (1.0 - ease)
		headDX = 0.08 * (1.0 - ease)
	}
	return FrameOffsetMap{
		PartTorso: {DX: bodyDX, DY: 0, Scale: 1.0},
		PartHead:  {DX: headDX, DY: 0.01, Scale: 1.0},
		PartLegs:  {DX: bodyDX * 0.4, DY: 0, Scale: 1.0},
		PartTail:  {DX: bodyDX * 0.6, DY: 0, Scale: 1.0},
	}
}

func quadrupedDeath(t float64) FrameOffsetMap {
	collapse := t * t
	return FrameOffsetMap{
		PartTorso:  {DX: collapse * 0.03, DY: collapse * 0.02, Scale: 1.0 + collapse*0.15},
		PartHead:   {DX: collapse * 0.04, DY: collapse * 0.03, Scale: 1.0 - collapse*0.2},
		PartLegs:   {DX: collapse * 0.05, DY: collapse * 0.01, Scale: 1.0 + collapse*0.1},
		PartTail:   {DX: -collapse * 0.03, DY: collapse * 0.02, Scale: 1.0 - collapse*0.3},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0 + collapse*0.2},
	}
}

// --- Serpentine offsets: sinusoidal undulation, head weaving ---

func serpentineOffsets(state string, t float64) FrameOffsetMap {
	switch state {
	case "walk", "run":
		return serpentineSlither(t, state == "run")
	case "attack":
		return serpentineStrike(t)
	case "hit":
		return serpentineHit(t)
	case "death":
		return serpentineDeath(t)
	default:
		return serpentineIdle(t)
	}
}

func serpentineIdle(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	weave := math.Sin(phase * 1.5)

	return FrameOffsetMap{
		PartHead: {
			DX:    weave * 0.03,
			DY:    math.Sin(phase)*0.005 - 0.002,
			Scale: 1.0,
		},
		PartTorso: {
			DX:    weave * 0.015,
			DY:    math.Sin(phase+math.Pi/3) * 0.004,
			Scale: 1.0 + math.Sin(phase)*0.01,
		},
		PartLegs: { // Tail segment
			DX:    -weave * 0.025,
			DY:    math.Sin(phase+2*math.Pi/3) * 0.003,
			Scale: 1.0,
		},
	}
}

func serpentineSlither(t float64, fast bool) FrameOffsetMap {
	phase := t * 2 * math.Pi
	amp := 0.06
	if fast {
		amp = 0.09
	}
	// Three-segment S-curve undulation: head, body, tail phase-shifted
	headWave := math.Sin(phase) * amp
	bodyWave := math.Sin(phase+math.Pi*0.6) * amp * 0.8
	tailWave := math.Sin(phase+math.Pi*1.2) * amp * 0.6

	return FrameOffsetMap{
		PartHead: {
			DX:    headWave,
			DY:    -math.Abs(headWave) * 0.15,
			Scale: 1.0,
		},
		PartTorso: {
			DX:    bodyWave,
			DY:    -math.Abs(bodyWave) * 0.1,
			Scale: 1.0,
		},
		PartLegs: { // Tail
			DX:    tailWave,
			DY:    math.Abs(tailWave) * 0.1,
			Scale: 1.0,
		},
	}
}

func serpentineStrike(t float64) FrameOffsetMap {
	// Coil (0–0.35), strike (0.35–0.55), retract (0.55–1.0)
	var headDY, bodyDX float64
	if t < 0.35 {
		p := t / 0.35
		headDY = p * 0.03     // Head rises
		bodyDX = p * p * 0.04 // Body coils
	} else if t < 0.55 {
		p := (t - 0.35) / 0.2
		headDY = 0.03 - p*0.12 // Head strikes downward
		bodyDX = 0.04 - p*0.06
	} else {
		p := (t - 0.55) / 0.45
		ease := 1.0 - (1.0-p)*(1.0-p)
		headDY = -0.09 * (1.0 - ease)
		bodyDX = -0.02 * (1.0 - ease)
	}
	return FrameOffsetMap{
		PartHead:  {DX: 0, DY: headDY, Scale: 1.0 + math.Abs(headDY)*0.5},
		PartTorso: {DX: bodyDX, DY: 0, Scale: 1.0},
		PartLegs:  {DX: -bodyDX * 0.5, DY: 0, Scale: 1.0}, // Tail anchors
	}
}

func serpentineHit(t float64) FrameOffsetMap {
	var dx float64
	if t < 0.15 {
		dx = (t / 0.15) * 0.07
	} else {
		p := (t - 0.15) / 0.85
		dx = 0.07 * math.Cos(p*math.Pi*3) * (1.0 - p) // Wobble/recoil
	}
	return FrameOffsetMap{
		PartHead:  {DX: dx * 1.2, DY: 0.01, Scale: 1.0},
		PartTorso: {DX: dx, DY: 0, Scale: 1.0},
		PartLegs:  {DX: dx * 0.5, DY: 0, Scale: 1.0},
	}
}

func serpentineDeath(t float64) FrameOffsetMap {
	collapse := t * t
	// Serpent curls and goes limp
	curl := math.Sin(collapse * math.Pi * 2)
	return FrameOffsetMap{
		PartHead:  {DX: curl * 0.04 * (1 - collapse), DY: collapse * 0.03, Scale: 1.0 - collapse*0.25},
		PartTorso: {DX: curl * 0.02 * (1 - collapse), DY: collapse * 0.02, Scale: 1.0 + collapse*0.1},
		PartLegs:  {DX: -curl * 0.03 * (1 - collapse), DY: collapse * 0.01, Scale: 1.0 - collapse*0.15},
	}
}

// --- Arachnid offsets: skittering legs, body bobbing ---

func arachnidOffsets(state string, t float64) FrameOffsetMap {
	switch state {
	case "walk", "run":
		return arachnidSkitter(t, state == "run")
	case "attack":
		return arachnidAttack(t)
	case "hit":
		return arachnidHit(t)
	case "death":
		return arachnidDeath(t)
	default:
		return arachnidIdle(t)
	}
}

func arachnidIdle(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	pulse := math.Sin(phase)
	// Legs shift subtly, pedipalps twitch
	legTwitch := math.Sin(phase*3) * 0.008

	return FrameOffsetMap{
		PartTorso: {
			DX:    0,
			DY:    pulse * 0.003,
			Scale: 1.0 + pulse*0.008,
		},
		PartHead: {
			DX:    legTwitch,
			DY:    pulse * -0.004,
			Scale: 1.0,
		},
		PartLegs: {
			DX:    legTwitch * 2,
			DY:    0,
			Scale: 1.0 + math.Abs(pulse)*0.01,
		},
	}
}

func arachnidSkitter(t float64, fast bool) FrameOffsetMap {
	phase := t * 2 * math.Pi
	freq := 3.0
	if fast {
		freq = 5.0
	}
	// Rapid alternating leg groups (tripod gait for insects/spiders)
	legGroup := math.Sin(phase * freq)
	bodyBob := math.Abs(math.Sin(phase*freq*2)) * 0.01

	return FrameOffsetMap{
		PartLegs: {
			DX:    legGroup * 0.04,
			DY:    -bodyBob,
			Scale: 1.0 + math.Abs(legGroup)*0.03,
		},
		PartTorso: {
			DX:    legGroup * 0.008,
			DY:    -bodyBob * 0.7,
			Scale: 1.0,
		},
		PartHead: {
			DX:    -legGroup * 0.01,
			DY:    -bodyBob * 0.5,
			Scale: 1.0,
		},
	}
}

func arachnidAttack(t float64) FrameOffsetMap {
	// Rear up (0–0.3), lunge (0.3–0.5), settle (0.5–1.0)
	var torsoScale, headDY, legDX float64
	if t < 0.3 {
		p := t / 0.3
		torsoScale = 1.0 + p*0.08 // Body rears up
		headDY = -p * 0.04        // Head lifts
		legDX = p * 0.03          // Legs spread
	} else if t < 0.5 {
		p := (t - 0.3) / 0.2
		torsoScale = 1.08 - p*0.04
		headDY = -0.04 + p*0.08 // Head strikes down
		legDX = 0.03 - p*0.04   // Legs thrust forward
	} else {
		p := (t - 0.5) / 0.5
		ease := 1.0 - (1.0-p)*(1.0-p)
		torsoScale = 1.04 - ease*0.04
		headDY = 0.04 * (1.0 - ease)
		legDX = -0.01 * (1.0 - ease)
	}
	return FrameOffsetMap{
		PartTorso: {DX: 0, DY: 0, Scale: torsoScale},
		PartHead:  {DX: 0, DY: headDY, Scale: 1.0},
		PartLegs:  {DX: legDX, DY: 0, Scale: 1.0 + math.Abs(legDX)*0.5},
	}
}

func arachnidHit(t float64) FrameOffsetMap {
	var bodyDY float64
	if t < 0.2 {
		bodyDY = (t / 0.2) * 0.04 // Body slams down
	} else {
		p := (t - 0.2) / 0.8
		bodyDY = 0.04 * (1.0 - p*p)
	}
	return FrameOffsetMap{
		PartTorso: {DX: 0, DY: bodyDY, Scale: 1.0 - bodyDY*2},
		PartHead:  {DX: 0, DY: bodyDY * 0.5, Scale: 1.0},
		PartLegs:  {DX: bodyDY * 0.5, DY: 0, Scale: 1.0 + bodyDY*3},
	}
}

func arachnidDeath(t float64) FrameOffsetMap {
	collapse := t * t
	// Legs curl inward (classic dead-bug pose from above)
	return FrameOffsetMap{
		PartTorso:  {DX: 0, DY: collapse * 0.01, Scale: 1.0 - collapse*0.1},
		PartHead:   {DX: 0, DY: collapse * 0.02, Scale: 1.0 - collapse*0.2},
		PartLegs:   {DX: 0, DY: 0, Scale: 1.0 - collapse*0.35}, // Legs contract inward
		PartShadow: {DX: 0, DY: 0, Scale: 1.0 - collapse*0.3},
	}
}

// --- Flying offsets: wing flapping, hovering bob ---

func flyingOffsets(state string, t float64) FrameOffsetMap {
	switch state {
	case "walk", "run":
		return flyingGlide(t, state == "run")
	case "attack":
		return flyingDive(t)
	case "hit":
		return flyingHit(t)
	case "death":
		return flyingFall(t)
	default:
		return flyingHover(t)
	}
}

func flyingHover(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	bob := math.Sin(phase) * 0.015
	wingBeat := math.Sin(phase * 3) // Wing flap faster than body

	return FrameOffsetMap{
		PartWings: {
			DX:    0,
			DY:    wingBeat * 0.008,
			Scale: 1.0 + wingBeat*0.06, // Wings spread/contract
		},
		PartTorso: {
			DX:    0,
			DY:    bob,
			Scale: 1.0,
		},
		PartHead: {
			DX:    0,
			DY:    bob * 0.7,
			Scale: 1.0,
		},
		PartLegs: { // Talons/feet dangle
			DX:    0,
			DY:    bob*0.5 + math.Abs(bob)*0.3,
			Scale: 1.0,
		},
		PartTail: {
			DX:    math.Sin(phase*0.8) * 0.02,
			DY:    bob * 0.4,
			Scale: 1.0,
		},
		PartShadow: {
			DX:    0,
			DY:    0,
			Scale: 1.0 + bob*2, // Shadow shrinks/grows with altitude
		},
	}
}

func flyingGlide(t float64, fast bool) FrameOffsetMap {
	phase := t * 2 * math.Pi
	flapRate := 2.0
	if fast {
		flapRate = 4.0
	}
	wingBeat := math.Sin(phase * flapRate)
	forwardLean := 0.01
	if fast {
		forwardLean = 0.02
	}

	return FrameOffsetMap{
		PartWings: {
			DX:    0,
			DY:    wingBeat * 0.01,
			Scale: 1.0 + wingBeat*0.08,
		},
		PartTorso: {
			DX:    0,
			DY:    -forwardLean + math.Abs(wingBeat)*0.005,
			Scale: 1.0,
		},
		PartHead: {
			DX:    0,
			DY:    -forwardLean * 1.5,
			Scale: 1.0,
		},
		PartTail: {
			DX:    math.Sin(phase*0.7) * 0.015,
			DY:    forwardLean,
			Scale: 1.0,
		},
		PartShadow: {
			DX:    0,
			DY:    0,
			Scale: 1.0 - forwardLean*3, // Shadow gets smaller when high
		},
	}
}

func flyingDive(t float64) FrameOffsetMap {
	// Wings fold (0–0.3), dive (0.3–0.6), pull up (0.6–1.0)
	var wingScale, bodyDY, headDY float64
	if t < 0.3 {
		p := t / 0.3
		wingScale = 1.0 - p*0.35 // Wings fold
		bodyDY = -p * 0.01
		headDY = -p * 0.02
	} else if t < 0.6 {
		p := (t - 0.3) / 0.3
		wingScale = 0.65 + p*0.1
		bodyDY = -0.01 + p*0.06 // Dive down
		headDY = -0.02 + p*0.08
	} else {
		p := (t - 0.6) / 0.4
		ease := 1.0 - (1.0-p)*(1.0-p)
		wingScale = 0.75 + ease*0.25 // Wings spread to brake
		bodyDY = 0.05 * (1.0 - ease)
		headDY = 0.06 * (1.0 - ease)
	}
	return FrameOffsetMap{
		PartWings:  {DX: 0, DY: 0, Scale: wingScale},
		PartTorso:  {DX: 0, DY: bodyDY, Scale: 1.0},
		PartHead:   {DX: 0, DY: headDY, Scale: 1.0},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0 + bodyDY*5},
	}
}

func flyingHit(t float64) FrameOffsetMap {
	// Tumble: rotation-like asymmetric offsets
	var wingDX, bodyDX float64
	if t < 0.2 {
		p := t / 0.2
		wingDX = p * 0.06
		bodyDX = p * 0.04
	} else {
		p := (t - 0.2) / 0.8
		decay := math.Exp(-p * 3)
		wingDX = 0.06 * math.Sin(p*math.Pi*4) * decay
		bodyDX = 0.04 * math.Sin(p*math.Pi*4) * decay
	}
	return FrameOffsetMap{
		PartWings: {DX: wingDX, DY: 0, Scale: 1.0 - math.Abs(wingDX)*2},
		PartTorso: {DX: bodyDX, DY: math.Abs(bodyDX) * 0.5, Scale: 1.0},
		PartHead:  {DX: bodyDX * 1.2, DY: 0, Scale: 1.0},
	}
}

func flyingFall(t float64) FrameOffsetMap {
	fall := t * t
	tumble := math.Sin(t*math.Pi*3) * (1.0 - fall)
	return FrameOffsetMap{
		PartWings:  {DX: tumble * 0.04, DY: 0, Scale: 1.0 - fall*0.5},
		PartTorso:  {DX: tumble * 0.02, DY: fall * 0.04, Scale: 1.0},
		PartHead:   {DX: tumble * 0.03, DY: fall * 0.05, Scale: 1.0 - fall*0.15},
		PartTail:   {DX: -tumble * 0.03, DY: fall * 0.03, Scale: 1.0},
		PartShadow: {DX: 0, DY: 0, Scale: 0.5 + fall*0.8}, // Shadow grows as creature falls to ground
	}
}

// --- Blob offsets: amoebic pulsation, asymmetric morphing ---

func blobOffsets(state string, t float64) FrameOffsetMap {
	switch state {
	case "walk", "run":
		return blobFlow(t, state == "run")
	case "attack":
		return blobEnvelope(t)
	case "hit":
		return blobSplash(t)
	case "death":
		return blobDissolve(t)
	default:
		return blobPulse(t)
	}
}

func blobPulse(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	// Asymmetric pulsation: X and Y scale at different rates
	pulseX := math.Sin(phase) * 0.04
	pulseY := math.Sin(phase+math.Pi/3) * 0.035

	return FrameOffsetMap{
		PartTorso: {
			DX:    pulseX * 0.3,
			DY:    pulseY * 0.3,
			Scale: 1.0 + math.Sin(phase)*0.03,
		},
		PartHead: { // "Core" or nucleus of the blob
			DX:    -pulseX * 0.2,
			DY:    -pulseY * 0.2,
			Scale: 1.0 + math.Sin(phase+math.Pi)*0.02,
		},
		PartLegs: { // Pseudopods / base
			DX:    pulseX * 0.5,
			DY:    pulseY * 0.4,
			Scale: 1.0 + math.Abs(math.Sin(phase*1.5))*0.04,
		},
		PartShadow: {
			DX:    0,
			DY:    0,
			Scale: 1.0 + math.Sin(phase)*0.025,
		},
	}
}

func blobFlow(t float64, fast bool) FrameOffsetMap {
	phase := t * 2 * math.Pi
	amp := 0.05
	if fast {
		amp = 0.08
	}
	// Flowing motion: body extends forward, trails behind
	extend := math.Sin(phase) * amp
	squish := math.Cos(phase) * amp * 0.5

	return FrameOffsetMap{
		PartTorso: {
			DX:    0,
			DY:    extend * 0.5,
			Scale: 1.0 - math.Abs(extend)*0.5, // Stretches thin when extending
		},
		PartHead: {
			DX:    squish * 0.3,
			DY:    extend,
			Scale: 1.0 + math.Abs(extend)*0.3,
		},
		PartLegs: {
			DX:    -squish * 0.3,
			DY:    -extend * 0.6,
			Scale: 1.0 + math.Abs(extend)*0.2,
		},
	}
}

func blobEnvelope(t float64) FrameOffsetMap {
	// Expand (0–0.4), engulf (0.4–0.7), retract (0.7–1.0)
	var scale, headDY float64
	if t < 0.4 {
		p := t / 0.4
		scale = 1.0 + p*0.15 // Expand
		headDY = -p * 0.02
	} else if t < 0.7 {
		p := (t - 0.4) / 0.3
		scale = 1.15 + p*0.05
		headDY = -0.02 + p*0.06
	} else {
		p := (t - 0.7) / 0.3
		ease := p * p
		scale = 1.2 - ease*0.2
		headDY = 0.04 * (1.0 - ease)
	}
	return FrameOffsetMap{
		PartTorso:  {DX: 0, DY: 0, Scale: scale},
		PartHead:   {DX: 0, DY: headDY, Scale: scale * 0.9},
		PartLegs:   {DX: 0, DY: -headDY * 0.5, Scale: scale * 0.8},
		PartShadow: {DX: 0, DY: 0, Scale: scale},
	}
}

func blobSplash(t float64) FrameOffsetMap {
	// Impact ripple effect
	var spread float64
	if t < 0.15 {
		spread = (t / 0.15) * 0.12
	} else {
		p := (t - 0.15) / 0.85
		spread = 0.12 * (1.0 - p) * math.Cos(p*math.Pi*2)
	}
	return FrameOffsetMap{
		PartTorso: {DX: 0, DY: 0, Scale: 1.0 + spread},
		PartHead:  {DX: spread * 0.5, DY: spread * 0.3, Scale: 1.0 - spread*0.5},
		PartLegs:  {DX: -spread * 0.5, DY: -spread * 0.3, Scale: 1.0 + spread*0.8},
	}
}

func blobDissolve(t float64) FrameOffsetMap {
	// Puddle: blob flattens and spreads
	spread := t
	return FrameOffsetMap{
		PartTorso:  {DX: 0, DY: spread * 0.02, Scale: 1.0 + spread*0.3},
		PartHead:   {DX: spread * 0.02, DY: spread * 0.03, Scale: 1.0 - spread*0.4},
		PartLegs:   {DX: -spread * 0.02, DY: spread * 0.01, Scale: 1.0 + spread*0.2},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0 + spread*0.5},
	}
}

// --- Mechanical offsets: rigid pistons, precise movement ---

func mechanicalOffsets(state string, t float64) FrameOffsetMap {
	switch state {
	case "walk", "run":
		return mechanicalStride(t, state == "run")
	case "attack":
		return mechanicalStrike(t)
	case "hit":
		return mechanicalRecoil(t)
	case "death":
		return mechanicalShutdown(t)
	default:
		return mechanicalIdle(t)
	}
}

func mechanicalIdle(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	// Mechanical idle: slight vibration, indicator pulse
	vibrate := math.Sin(phase*8) * 0.002

	return FrameOffsetMap{
		PartTorso: {
			DX:    vibrate,
			DY:    vibrate * 0.5,
			Scale: 1.0,
		},
		PartHead: {
			DX:    0,
			DY:    math.Sin(phase) * 0.003, // Scanner sweep
			Scale: 1.0,
		},
	}
}

func mechanicalStride(t float64, fast bool) FrameOffsetMap {
	phase := t * 2 * math.Pi
	// Rigid piston-like leg movement (no organic smoothing)
	step := math.Floor(math.Sin(phase)*2+0.5) / 2 // Quantized to steps
	speed := 1.0
	if fast {
		speed = 1.5
	}

	return FrameOffsetMap{
		PartLegs: {
			DX:    step * 0.03 * speed,
			DY:    0,
			Scale: 1.0,
		},
		PartTorso: {
			DX:    0,
			DY:    -math.Abs(step) * 0.008 * speed,
			Scale: 1.0,
		},
		PartHead: {
			DX:    0,
			DY:    -math.Abs(step) * 0.005 * speed,
			Scale: 1.0,
		},
		PartArms: { // Manipulator arms
			DX:    -step * 0.02 * speed,
			DY:    0,
			Scale: 1.0,
		},
	}
}

func mechanicalStrike(t float64) FrameOffsetMap {
	// Wind-up (0–0.2), execute (0.2–0.4), retract (0.4–1.0)
	var armDX, bodyDX float64
	if t < 0.2 {
		p := t / 0.2
		armDX = -p * 0.04
		bodyDX = -p * 0.01
	} else if t < 0.4 {
		p := (t - 0.2) / 0.2
		armDX = -0.04 + p*0.14 // Fast, precise strike
		bodyDX = -0.01 + p*0.03
	} else {
		p := (t - 0.4) / 0.6
		armDX = 0.10 * (1.0 - p) // Smooth retract
		bodyDX = 0.02 * (1.0 - p)
	}
	return FrameOffsetMap{
		PartArms:  {DX: armDX, DY: 0, Scale: 1.0},
		PartTorso: {DX: bodyDX, DY: 0, Scale: 1.0},
		PartHead:  {DX: bodyDX * 0.5, DY: 0, Scale: 1.0},
	}
}

func mechanicalRecoil(t float64) FrameOffsetMap {
	// Rigid recoil with sparks (represented by scale flicker)
	var dx, flicker float64
	if t < 0.1 {
		dx = (t / 0.1) * 0.05
		flicker = 0.05
	} else {
		p := (t - 0.1) / 0.9
		dx = 0.05 * (1.0 - p)
		flicker = 0.05 * (1.0 - p) * float64(int(p*10)%2)
	}
	return FrameOffsetMap{
		PartTorso: {DX: dx, DY: 0, Scale: 1.0 + flicker},
		PartHead:  {DX: dx * 0.8, DY: 0, Scale: 1.0 - flicker},
		PartLegs:  {DX: dx * 0.3, DY: 0, Scale: 1.0},
	}
}

func mechanicalShutdown(t float64) FrameOffsetMap {
	// Power-down: slow tilt, then collapse
	var tilt, drop float64
	if t < 0.6 {
		p := t / 0.6
		tilt = p * 0.04 // Slow lean
		drop = p * p * 0.01
	} else {
		p := (t - 0.6) / 0.4
		tilt = 0.04 + p*0.02
		drop = 0.01 + p*p*0.03
	}
	return FrameOffsetMap{
		PartTorso: {DX: tilt, DY: drop, Scale: 1.0},
		PartHead:  {DX: tilt * 1.3, DY: drop * 1.5, Scale: 1.0 - drop*2},
		PartArms:  {DX: tilt * 0.5, DY: drop * 2, Scale: 1.0},
		PartLegs:  {DX: 0, DY: 0, Scale: 1.0},
	}
}

// --- Undead offsets: shambling, jerky, unsettling ---

func undeadOffsets(state string, t float64) FrameOffsetMap {
	switch state {
	case "walk", "run":
		return undeadShamble(t, state == "run")
	case "attack":
		return undeadLurch(t)
	case "hit":
		return undeadStagger(t)
	case "death":
		return undeadCollapse(t)
	default:
		return undeadSway(t)
	}
}

func undeadSway(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	// Unsettling, irregular sway with head lolling
	sway := math.Sin(phase) * 0.02
	headLoll := math.Sin(phase*1.7+0.5) * 0.025

	return FrameOffsetMap{
		PartTorso: {
			DX:    sway,
			DY:    math.Abs(sway) * 0.3,
			Scale: 1.0,
		},
		PartHead: {
			DX:    headLoll,
			DY:    math.Abs(headLoll) * 0.4,
			Scale: 1.0,
		},
		PartArms: {
			DX:    sway * 1.5, // Arms dangle loosely
			DY:    math.Abs(sway) * 0.6,
			Scale: 1.0,
		},
		PartLegs: {
			DX:    -sway * 0.3,
			DY:    0,
			Scale: 1.0,
		},
	}
}

func undeadShamble(t float64, frantic bool) FrameOffsetMap {
	phase := t * 2 * math.Pi
	speed := 1.0
	if frantic {
		speed = 1.8
	}
	// Uneven, lurching gait — one leg drags
	legA := math.Sin(phase) * 0.04 * speed
	legB := math.Sin(phase+0.4) * 0.025 * speed // Asymmetric — one leg lags
	sway := math.Sin(phase*0.8) * 0.025 * speed

	return FrameOffsetMap{
		PartLegs: {
			DX:    (legA + legB) * 0.5,
			DY:    -math.Abs(legA) * 0.15,
			Scale: 1.0,
		},
		PartTorso: {
			DX:    sway,
			DY:    -math.Abs(legA) * 0.08,
			Scale: 1.0,
		},
		PartHead: {
			DX:    sway*1.3 + math.Sin(phase*2.3)*0.008,
			DY:    math.Abs(sway) * 0.3,
			Scale: 1.0,
		},
		PartArms: {
			DX:    sway * 1.8,
			DY:    math.Abs(sway) * 0.5,
			Scale: 1.0,
		},
	}
}

func undeadLurch(t float64) FrameOffsetMap {
	// Shambling lurch attack: slow telegraphed then sudden
	var bodyDX, headDX, armDX float64
	if t < 0.5 {
		p := t / 0.5
		bodyDX = -p * p * 0.03 // Slow wind-up
		headDX = -p * p * 0.02
		armDX = -p * 0.04
	} else if t < 0.65 {
		p := (t - 0.5) / 0.15
		bodyDX = -0.03 + p*0.12 // Sudden lurch
		headDX = -0.02 + p*0.10
		armDX = -0.04 + p*0.14
	} else {
		p := (t - 0.65) / 0.35
		bodyDX = 0.09 * (1.0 - p)
		headDX = 0.08 * (1.0 - p)
		armDX = 0.10 * (1.0 - p)
	}
	return FrameOffsetMap{
		PartTorso: {DX: bodyDX, DY: 0, Scale: 1.0},
		PartHead:  {DX: headDX, DY: math.Abs(headDX) * 0.3, Scale: 1.0},
		PartArms:  {DX: armDX, DY: 0, Scale: 1.0},
		PartLegs:  {DX: bodyDX * 0.3, DY: 0, Scale: 1.0},
	}
}

func undeadStagger(t float64) FrameOffsetMap {
	// Hit causes wild staggering
	var dx, dy float64
	if t < 0.15 {
		p := t / 0.15
		dx = p * 0.08
		dy = p * 0.02
	} else {
		p := (t - 0.15) / 0.85
		decay := 1.0 - p
		dx = 0.08 * decay * math.Sin(p*math.Pi*5)
		dy = 0.02 * decay * math.Abs(math.Sin(p*math.Pi*5))
	}
	return FrameOffsetMap{
		PartTorso: {DX: dx, DY: dy, Scale: 1.0},
		PartHead:  {DX: dx * 1.3, DY: dy * 1.5, Scale: 1.0},
		PartArms:  {DX: dx * 1.6, DY: dy * 0.5, Scale: 1.0},
		PartLegs:  {DX: dx * 0.4, DY: 0, Scale: 1.0},
	}
}

func undeadCollapse(t float64) FrameOffsetMap {
	// Crumbling apart: limbs splay, body deflates
	collapse := t * t
	return FrameOffsetMap{
		PartTorso:  {DX: collapse * 0.02, DY: collapse * 0.03, Scale: 1.0 - collapse*0.15},
		PartHead:   {DX: collapse * 0.05, DY: collapse * 0.04, Scale: 1.0 - collapse*0.3},
		PartArms:   {DX: collapse * 0.06, DY: collapse * 0.02, Scale: 1.0 - collapse*0.25},
		PartLegs:   {DX: -collapse * 0.04, DY: collapse * 0.01, Scale: 1.0 + collapse*0.1},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0 - collapse*0.5},
	}
}

// --- Insect offsets: scuttling legs, antennae bob, abdomen sway ---

func insectOffsets(state string, t float64) FrameOffsetMap {
	switch state {
	case "walk", "run":
		return insectScuttle(t, state == "run")
	case "attack":
		return insectBite(t)
	case "hit":
		return insectFlinch(t)
	case "death":
		return insectFlip(t)
	default:
		return insectIdle(t)
	}
}

func insectIdle(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	// Antennae (head) bob and legs twitch subtly
	return FrameOffsetMap{
		PartHead:   {DX: math.Sin(phase*2) * 0.01, DY: math.Sin(phase) * 0.005, Scale: 1.0},
		PartArms:   {DX: 0, DY: 0, Scale: 1.0}, // thorax stays still
		PartTorso:  {DX: 0, DY: math.Sin(phase*0.5) * 0.003, Scale: 1.0 + math.Sin(phase)*0.01},
		PartLegs:   {DX: math.Sin(phase*3) * 0.008, DY: 0, Scale: 1.0},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0},
	}
}

func insectScuttle(t float64, fast bool) FrameOffsetMap {
	phase := t * 2 * math.Pi
	amp := 0.03
	if fast {
		amp = 0.05
	}
	// Rapid alternating leg movement
	legPhase := math.Sin(phase * 4)
	return FrameOffsetMap{
		PartHead:   {DX: math.Sin(phase*2) * amp * 0.3, DY: -amp * 0.2, Scale: 1.0},
		PartArms:   {DX: 0, DY: math.Sin(phase) * amp * 0.2, Scale: 1.0},
		PartTorso:  {DX: 0, DY: math.Sin(phase*2) * amp * 0.15, Scale: 1.0},
		PartLegs:   {DX: legPhase * amp, DY: math.Cos(phase*4) * amp * 0.5, Scale: 1.0},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0},
	}
}

func insectBite(t float64) FrameOffsetMap {
	// Head lunges forward, mandibles open
	lunge := math.Sin(t * math.Pi)
	return FrameOffsetMap{
		PartHead:   {DX: 0, DY: -lunge * 0.06, Scale: 1.0 + lunge*0.05},
		PartArms:   {DX: 0, DY: -lunge * 0.02, Scale: 1.0},
		PartTorso:  {DX: 0, DY: 0, Scale: 1.0},
		PartLegs:   {DX: lunge * 0.02, DY: 0, Scale: 1.0 + lunge*0.02},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0},
	}
}

func insectFlinch(t float64) FrameOffsetMap {
	recoil := math.Sin(t * math.Pi) * math.Exp(-t * 2)
	return FrameOffsetMap{
		PartHead:   {DX: recoil * 0.03, DY: recoil * 0.02, Scale: 1.0 - recoil*0.05},
		PartArms:   {DX: 0, DY: recoil * 0.01, Scale: 1.0},
		PartTorso:  {DX: 0, DY: recoil * 0.02, Scale: 1.0},
		PartLegs:   {DX: -recoil * 0.02, DY: 0, Scale: 1.0 + math.Abs(recoil)*0.03},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0},
	}
}

func insectFlip(t float64) FrameOffsetMap {
	// Insect flips over on death, legs curl inward
	flip := t * t
	return FrameOffsetMap{
		PartHead:   {DX: flip * 0.04, DY: flip * 0.03, Scale: 1.0 - flip*0.2},
		PartArms:   {DX: 0, DY: flip * 0.02, Scale: 1.0 - flip*0.15},
		PartTorso:  {DX: 0, DY: flip * 0.01, Scale: 1.0 - flip*0.1},
		PartLegs:   {DX: 0, DY: 0, Scale: 1.0 - flip*0.4},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0 - flip*0.3},
	}
}

// --- Multi-limbed offsets: writhing tentacles, pulsing mass, asymmetric motion ---

func multiLimbedOffsets(state string, t float64) FrameOffsetMap {
	switch state {
	case "walk", "run":
		return multiLimbedCrawl(t, state == "run")
	case "attack":
		return multiLimbedGrab(t)
	case "hit":
		return multiLimbedRecoil(t)
	case "death":
		return multiLimbedCollapse(t)
	default:
		return multiLimbedWrithe(t)
	}
}

func multiLimbedWrithe(t float64) FrameOffsetMap {
	phase := t * 2 * math.Pi
	// Constant asymmetric writhing — each part moves on a different cycle
	return FrameOffsetMap{
		PartHead: {
			DX:    math.Sin(phase*1.7) * 0.015,
			DY:    math.Cos(phase*1.3) * 0.012,
			Scale: 1.0 + math.Sin(phase*0.8)*0.02,
		},
		PartArms: { // secondary mass
			DX:    math.Sin(phase*2.1+0.5) * 0.02,
			DY:    math.Cos(phase*1.9) * 0.018,
			Scale: 1.0 + math.Sin(phase*1.2)*0.03,
		},
		PartTorso: { // central body pulsates
			DX:    math.Sin(phase*0.7) * 0.008,
			DY:    math.Cos(phase*0.9) * 0.008,
			Scale: 1.0 + math.Sin(phase)*0.025,
		},
		PartLegs: { // tentacles sway
			DX:    math.Sin(phase*2.5) * 0.03,
			DY:    math.Cos(phase*2.3+0.7) * 0.025,
			Scale: 1.0 + math.Sin(phase*1.5)*0.02,
		},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0 + math.Sin(phase*0.5)*0.015},
	}
}

func multiLimbedCrawl(t float64, fast bool) FrameOffsetMap {
	phase := t * 2 * math.Pi
	amp := 0.04
	if fast {
		amp = 0.07
	}
	return FrameOffsetMap{
		PartHead:   {DX: math.Sin(phase*1.5) * amp * 0.4, DY: -amp * 0.3, Scale: 1.0},
		PartArms:   {DX: math.Cos(phase*2) * amp * 0.5, DY: math.Sin(phase*1.7) * amp * 0.3, Scale: 1.0},
		PartTorso:  {DX: 0, DY: math.Sin(phase) * amp * 0.2, Scale: 1.0 + math.Sin(phase)*0.02},
		PartLegs:   {DX: math.Sin(phase*3) * amp, DY: math.Cos(phase*2.5) * amp * 0.8, Scale: 1.0},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0},
	}
}

func multiLimbedGrab(t float64) FrameOffsetMap {
	// Tentacles lash outward then retract
	strike := math.Sin(t * math.Pi)
	retract := math.Max(0, math.Sin(t*math.Pi*2-math.Pi/2))
	return FrameOffsetMap{
		PartHead:   {DX: 0, DY: -strike * 0.03, Scale: 1.0 + strike*0.05},
		PartArms:   {DX: strike * 0.04, DY: -strike * 0.02, Scale: 1.0 + strike*0.08},
		PartTorso:  {DX: 0, DY: retract * 0.02, Scale: 1.0 + retract*0.04},
		PartLegs:   {DX: strike * 0.06, DY: strike * 0.05, Scale: 1.0 + strike*0.1},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0 + strike*0.05},
	}
}

func multiLimbedRecoil(t float64) FrameOffsetMap {
	shock := math.Sin(t*math.Pi) * math.Exp(-t*1.5)
	return FrameOffsetMap{
		PartHead:   {DX: shock * 0.04, DY: shock * 0.03, Scale: 1.0 - math.Abs(shock)*0.05},
		PartArms:   {DX: -shock * 0.03, DY: shock * 0.02, Scale: 1.0},
		PartTorso:  {DX: shock * 0.02, DY: shock * 0.02, Scale: 1.0 - math.Abs(shock)*0.03},
		PartLegs:   {DX: -shock * 0.05, DY: -shock * 0.03, Scale: 1.0 - math.Abs(shock)*0.04},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0},
	}
}

func multiLimbedCollapse(t float64) FrameOffsetMap {
	// Tentacles splay and go limp, body deflates
	collapse := t * t
	return FrameOffsetMap{
		PartHead:   {DX: collapse * 0.06, DY: collapse * 0.04, Scale: 1.0 - collapse*0.35},
		PartArms:   {DX: -collapse * 0.04, DY: collapse * 0.03, Scale: 1.0 - collapse*0.2},
		PartTorso:  {DX: 0, DY: collapse * 0.02, Scale: 1.0 - collapse*0.25},
		PartLegs:   {DX: collapse * 0.03, DY: collapse * 0.05, Scale: 1.0 + collapse*0.15},
		PartShadow: {DX: 0, DY: 0, Scale: 1.0 - collapse*0.4},
	}
}
