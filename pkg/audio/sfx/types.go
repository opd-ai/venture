package sfx

// EffectType represents different types of sound effects.
// Use these constants to specify which sound effect to generate.
// Originally from: generator.go
type EffectType string

// Sound effect type constants define the categories of procedurally generated sounds.
// Each type has distinct characteristics optimized for its gameplay context.
const (
	// EffectImpact generates short, punchy sounds for collisions and melee hits.
	// Typical duration: 0.05-0.15 seconds. Sharp attack, quick decay.
	EffectImpact EffectType = "impact"

	// EffectExplosion generates longer, rumbling sounds for destruction effects.
	// Typical duration: 0.5-1.0 seconds. Multiple layers with noise component.
	EffectExplosion EffectType = "explosion"

	// EffectMagic generates ethereal, mystical sounds for spellcasting.
	// Typical duration: 0.2-0.5 seconds. Sweeping frequencies with reverb-like tail.
	EffectMagic EffectType = "magic"

	// EffectLaser generates synthetic beam sounds for ranged energy weapons.
	// Typical duration: 0.1-0.3 seconds. High-frequency sweep with clean waveforms.
	EffectLaser EffectType = "laser"

	// EffectPickup generates pleasant notification sounds for item collection.
	// Typical duration: 0.1-0.2 seconds. Rising pitch with bright timbre.
	EffectPickup EffectType = "pickup"

	// EffectHit generates feedback sounds for damage taken or dealt.
	// Typical duration: 0.05-0.1 seconds. Similar to impact but more visceral.
	EffectHit EffectType = "hit"

	// EffectJump generates whoosh sounds for character movement actions.
	// Typical duration: 0.15-0.25 seconds. Rising then falling pitch envelope.
	EffectJump EffectType = "jump"

	// EffectDeath generates dramatic sounds for character defeat.
	// Typical duration: 0.8-1.2 seconds. Longest effect with complex envelope.
	EffectDeath EffectType = "death"

	// EffectPowerup generates triumphant sounds for gaining abilities or bonuses.
	// Typical duration: 0.3-0.5 seconds. Multiple rising tones layered together.
	EffectPowerup EffectType = "powerup"
)
