// Package palette provides procedural color palette generation for genre-based theming.
// Color palettes are generated deterministically from seeds and genre definitions,
// ensuring consistent visual styles across game sessions.
//
// # Phase 17.3: Time-of-Day Color Shifts
//
// The palette system supports dynamic time-of-day color modulation with four time states:
// - Dawn: Warm, soft tones with reduced saturation
// - Day: Bright, saturated colors (neutral baseline)
// - Dusk: Warm, rich tones with enhanced saturation
// - Night: Cool, desaturated colors with reduced lightness
//
// Time transitions are smooth with configurable 5-second interpolation periods,
// and intensity can be scaled from 0.0 (no effect) to 1.0 (full effect).
//
// # Basic Usage
//
//	gen := palette.NewGenerator()
//	palette, _ := gen.Generate("fantasy", 12345)
//
// # Time-of-Day Usage
//
//	timeConfig := palette.TimeConfig{
//	    CurrentTime: palette.TimeOfDayNight,
//	    TransitionProgress: 0.5,  // Halfway to next time
//	    IntensityMultiplier: 1.0, // Full effect
//	}
//	palette, _ := gen.GenerateWithTime("fantasy", 12345, timeConfig)
//
// # Advanced Usage
//
//	opts := palette.GenerationOptions{
//	    Harmony: palette.HarmonyTriadic,
//	    Mood: palette.MoodVibrant,
//	    Rarity: palette.RarityEpic,
//	}
//	palette, _ := gen.GenerateWithOptionsAndTime("scifi", 54321, opts, timeConfig)
//
// # Performance
//
// Time-of-day modulation adds <1% frame time overhead:
// - GetModulationForTime: ~2ns
// - ApplyTimeModulation: ~50µs for full palette
// - Color interpolation uses smooth step for natural transitions
//
// All generation is deterministic - same seed and parameters produce identical output.
package palette
