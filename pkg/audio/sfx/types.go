package sfx

// EffectType represents different types of sound effects.
// Use these constants to specify which sound effect to generate.
// Originally from: generator.go
type EffectType string

// Sound effect type constants.
// Originally from: generator.go
const (
	EffectImpact    EffectType = "impact"
	EffectExplosion EffectType = "explosion"
	EffectMagic     EffectType = "magic"
	EffectLaser     EffectType = "laser"
	EffectPickup    EffectType = "pickup"
	EffectHit       EffectType = "hit"
	EffectJump      EffectType = "jump"
	EffectDeath     EffectType = "death"
	EffectPowerup   EffectType = "powerup"
)
