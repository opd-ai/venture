// Package postprocess provides post-processing effects for rendered scenes.
package postprocess

import "image/color"

// FantasyPreset returns post-processing settings for fantasy genre.
// Characteristics: Warm colors, high saturation, soft vignette.
func FantasyPreset() Preset {
	return Preset{
		Name:        "Fantasy",
		Description: "Warm, saturated colors with soft vignette for medieval fantasy atmosphere",
		Config: Config{
			MotionBlur: DefaultMotionBlurConfig(),
			DepthBlur:  DefaultDepthBlurConfig(),
			ColorGrading: ColorGradingConfig{
				Enabled:     true,
				Saturation:  1.2,
				Contrast:    1.1,
				Brightness:  0.05,
				Temperature: 0.15, // Warm
				Tint:        0.0,
			},
			Vignette: VignetteConfig{
				Enabled:     true,
				Intensity:   0.4,
				Softness:    0.7,
				InnerRadius: 0.5,
				OuterRadius: 1.3,
				Color:       color.RGBA{20, 15, 10, 255}, // Warm dark brown
			},
			ChromaticAberration: ChromaticAberrationConfig{
				Enabled:    false,
				Intensity:  0.0,
				Samples:    3,
				DirectionX: 1.0,
				DirectionY: 0.0,
			},
		},
	}
}

// SciFiPreset returns post-processing settings for sci-fi genre.
// Characteristics: Cool colors, high contrast, subtle chromatic aberration.
func SciFiPreset() Preset {
	return Preset{
		Name:        "Sci-Fi",
		Description: "Cool, high-contrast colors with chromatic aberration for futuristic technology",
		Config: Config{
			MotionBlur: DefaultMotionBlurConfig(),
			DepthBlur:  DefaultDepthBlurConfig(),
			ColorGrading: ColorGradingConfig{
				Enabled:     true,
				Saturation:  1.15,
				Contrast:    1.25,
				Brightness:  0.0,
				Temperature: -0.2,  // Cool
				Tint:        -0.05, // Slight green
			},
			Vignette: VignetteConfig{
				Enabled:     true,
				Intensity:   0.35,
				Softness:    0.5,
				InnerRadius: 0.4,
				OuterRadius: 1.2,
				Color:       color.RGBA{0, 5, 10, 255}, // Cool dark blue
			},
			ChromaticAberration: ChromaticAberrationConfig{
				Enabled:    true,
				Intensity:  0.15,
				Samples:    3,
				DirectionX: 1.0,
				DirectionY: 0.0,
			},
		},
	}
}

// HorrorPreset returns post-processing settings for horror genre.
// Characteristics: Desaturated, low brightness, strong vignette.
func HorrorPreset() Preset {
	return Preset{
		Name:        "Horror",
		Description: "Desaturated, dark atmosphere with strong vignette for horror and dread",
		Config: Config{
			MotionBlur: DefaultMotionBlurConfig(),
			DepthBlur:  DefaultDepthBlurConfig(),
			ColorGrading: ColorGradingConfig{
				Enabled:     true,
				Saturation:  0.6,
				Contrast:    1.3,
				Brightness:  -0.15,
				Temperature: -0.1, // Slightly cool
				Tint:        0.05, // Slight magenta
			},
			Vignette: VignetteConfig{
				Enabled:     true,
				Intensity:   0.7,
				Softness:    0.4,
				InnerRadius: 0.3,
				OuterRadius: 1.0,
				Color:       color.RGBA{5, 0, 5, 255}, // Very dark purple
			},
			ChromaticAberration: ChromaticAberrationConfig{
				Enabled:    false,
				Intensity:  0.0,
				Samples:    3,
				DirectionX: 1.0,
				DirectionY: 0.0,
			},
		},
	}
}

// CyberpunkPreset returns post-processing settings for cyberpunk genre.
// Characteristics: High saturation, neon tints, chromatic aberration, harsh contrast.
func CyberpunkPreset() Preset {
	return Preset{
		Name:        "Cyberpunk",
		Description: "High saturation neon colors with chromatic aberration for dystopian cyberpunk",
		Config: Config{
			MotionBlur: DefaultMotionBlurConfig(),
			DepthBlur:  DefaultDepthBlurConfig(),
			ColorGrading: ColorGradingConfig{
				Enabled:     true,
				Saturation:  1.4,
				Contrast:    1.35,
				Brightness:  -0.05,
				Temperature: 0.0,
				Tint:        0.1, // Magenta tint for neon
			},
			Vignette: VignetteConfig{
				Enabled:     true,
				Intensity:   0.45,
				Softness:    0.3,
				InnerRadius: 0.4,
				OuterRadius: 1.1,
				Color:       color.RGBA{10, 0, 15, 255}, // Dark magenta
			},
			ChromaticAberration: ChromaticAberrationConfig{
				Enabled:    true,
				Intensity:  0.25,
				Samples:    4,
				DirectionX: 1.0,
				DirectionY: 0.0,
			},
		},
	}
}

// PostApocalypticPreset returns post-processing settings for post-apocalyptic genre.
// Characteristics: Dusty/brown tint, low saturation, harsh vignette.
func PostApocalypticPreset() Preset {
	return Preset{
		Name:        "Post-Apocalyptic",
		Description: "Dusty, desaturated wasteland atmosphere with brown tint and harsh vignette",
		Config: Config{
			MotionBlur: DefaultMotionBlurConfig(),
			DepthBlur:  DefaultDepthBlurConfig(),
			ColorGrading: ColorGradingConfig{
				Enabled:     true,
				Saturation:  0.7,
				Contrast:    1.2,
				Brightness:  -0.1,
				Temperature: 0.25, // Warm/dusty
				Tint:        -0.1, // Slight green for decay
			},
			Vignette: VignetteConfig{
				Enabled:     true,
				Intensity:   0.6,
				Softness:    0.5,
				InnerRadius: 0.35,
				OuterRadius: 1.15,
				Color:       color.RGBA{20, 15, 10, 255}, // Dusty brown
			},
			ChromaticAberration: ChromaticAberrationConfig{
				Enabled:    false,
				Intensity:  0.0,
				Samples:    3,
				DirectionX: 1.0,
				DirectionY: 0.0,
			},
		},
	}
}

// NeutralPreset returns minimal post-processing settings.
// Useful for testing or when no genre-specific effects are desired.
func NeutralPreset() Preset {
	return Preset{
		Name:        "Neutral",
		Description: "Minimal post-processing with neutral color grading",
		Config:      DefaultConfig(),
	}
}

// CinematicPreset returns dramatic post-processing for cutscenes.
// Characteristics: Strong vignette, depth blur, enhanced contrast.
func CinematicPreset() Preset {
	return Preset{
		Name:        "Cinematic",
		Description: "Dramatic cinematic look with strong vignette and depth of field",
		Config: Config{
			MotionBlur: DefaultMotionBlurConfig(),
			DepthBlur: DepthBlurConfig{
				Enabled:       true,
				FocalDistance: 0.5,
				FocalRange:    0.15,
				BlurStrength:  0.7,
				Samples:       9,
			},
			ColorGrading: ColorGradingConfig{
				Enabled:     true,
				Saturation:  1.1,
				Contrast:    1.3,
				Brightness:  0.0,
				Temperature: 0.05,
				Tint:        0.0,
			},
			Vignette: VignetteConfig{
				Enabled:     true,
				Intensity:   0.65,
				Softness:    0.6,
				InnerRadius: 0.35,
				OuterRadius: 1.0,
				Color:       color.RGBA{0, 0, 0, 255},
			},
			ChromaticAberration: ChromaticAberrationConfig{
				Enabled:    true,
				Intensity:  0.1,
				Samples:    3,
				DirectionX: 1.0,
				DirectionY: 0.0,
			},
		},
	}
}

// GetPresetByGenre returns the appropriate preset for a genre ID.
// Supports: "fantasy", "scifi", "horror", "cyberpunk", "postapoc", "cinematic", "neutral".
// Returns NeutralPreset() if genre is not recognized.
func GetPresetByGenre(genreID string) Preset {
	switch genreID {
	case "fantasy":
		return FantasyPreset()
	case "scifi", "sci-fi":
		return SciFiPreset()
	case "horror":
		return HorrorPreset()
	case "cyberpunk":
		return CyberpunkPreset()
	case "postapoc", "post-apocalyptic":
		return PostApocalypticPreset()
	case "cinematic":
		return CinematicPreset()
	case "neutral":
		return NeutralPreset()
	default:
		return NeutralPreset()
	}
}

// AllPresets returns all available presets as a slice.
func AllPresets() []Preset {
	return []Preset{
		FantasyPreset(),
		SciFiPreset(),
		HorrorPreset(),
		CyberpunkPreset(),
		PostApocalypticPreset(),
		NeutralPreset(),
		CinematicPreset(),
	}
}
