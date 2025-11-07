// Package palette provides time-of-day color modulation.
// Phase 17.3: Time-of-Day Color Shifts
// This file implements dynamic palette adjustments based on time of day
// with smooth transitions between time states.
package palette

import (
	"image/color"
	"math"
)

// GetModulationForTime returns the color modulation for a specific time of day.
// Each time period has characteristic color temperature and lighting adjustments.
func GetModulationForTime(timeOfDay TimeOfDay) ColorModulation {
	switch timeOfDay {
	case TimeOfDayDawn:
		// Dawn: Warm, soft tones with reduced saturation
		return ColorModulation{
			HueShift:             15.0, // Slight orange shift
			SaturationMultiplier: 0.85, // Reduced saturation
			LightnessOffset:      0.05, // Slightly brighter
			TemperatureShift:     0.4,  // Warm tones
		}
	case TimeOfDayDay:
		// Day: Bright, saturated colors (neutral baseline)
		return ColorModulation{
			HueShift:             0.0, // No hue shift
			SaturationMultiplier: 1.0, // Full saturation
			LightnessOffset:      0.0, // Standard lightness
			TemperatureShift:     0.0, // Neutral temperature
		}
	case TimeOfDayDusk:
		// Dusk: Warm, rich tones with enhanced saturation
		return ColorModulation{
			HueShift:             25.0,  // Orange/red shift
			SaturationMultiplier: 1.15,  // Enhanced saturation
			LightnessOffset:      -0.05, // Slightly darker
			TemperatureShift:     0.6,   // Very warm tones
		}
	case TimeOfDayNight:
		// Night: Cool, desaturated colors with reduced lightness
		return ColorModulation{
			HueShift:             -10.0, // Slight blue shift
			SaturationMultiplier: 0.6,   // Heavily desaturated
			LightnessOffset:      -0.25, // Much darker
			TemperatureShift:     -0.5,  // Cool tones
		}
	default:
		// Fallback to day (neutral)
		return GetModulationForTime(TimeOfDayDay)
	}
}

// InterpolateModulation smoothly interpolates between two ColorModulation values.
// progress is 0.0 (100% from) to 1.0 (100% to).
// Uses smooth step function for natural-feeling transitions.
func InterpolateModulation(from, to ColorModulation, progress float64) ColorModulation {
	// Clamp progress to [0, 1]
	if progress < 0.0 {
		progress = 0.0
	} else if progress > 1.0 {
		progress = 1.0
	}

	// Apply smooth step for easing (3t² - 2t³)
	t := progress * progress * (3.0 - 2.0*progress)

	return ColorModulation{
		HueShift:             from.HueShift + (to.HueShift-from.HueShift)*t,
		SaturationMultiplier: from.SaturationMultiplier + (to.SaturationMultiplier-from.SaturationMultiplier)*t,
		LightnessOffset:      from.LightnessOffset + (to.LightnessOffset-from.LightnessOffset)*t,
		TemperatureShift:     from.TemperatureShift + (to.TemperatureShift-from.TemperatureShift)*t,
	}
}

// GetModulationWithTransition returns the color modulation for the current time
// with smooth transition to the next time state based on progress.
func GetModulationWithTransition(config TimeConfig) ColorModulation {
	current := GetModulationForTime(config.CurrentTime)

	// If no transition in progress, return current modulation
	if config.TransitionProgress <= 0.0 {
		return scaleModulation(current, config.IntensityMultiplier)
	}

	// Determine next time state for transition
	nextTime := getNextTimeState(config.CurrentTime)
	next := GetModulationForTime(nextTime)

	// Interpolate between current and next
	interpolated := InterpolateModulation(current, next, config.TransitionProgress)

	return scaleModulation(interpolated, config.IntensityMultiplier)
}

// getNextTimeState returns the next time of day in the cycle.
func getNextTimeState(current TimeOfDay) TimeOfDay {
	switch current {
	case TimeOfDayDawn:
		return TimeOfDayDay
	case TimeOfDayDay:
		return TimeOfDayDusk
	case TimeOfDayDusk:
		return TimeOfDayNight
	case TimeOfDayNight:
		return TimeOfDayDawn
	default:
		return TimeOfDayDay
	}
}

// scaleModulation scales the modulation by intensity multiplier.
func scaleModulation(mod ColorModulation, intensity float64) ColorModulation {
	// Clamp intensity to [0, 1]
	if intensity < 0.0 {
		intensity = 0.0
	} else if intensity > 1.0 {
		intensity = 1.0
	}

	// Scale modulation effects towards neutral (day) values
	neutral := GetModulationForTime(TimeOfDayDay)

	return ColorModulation{
		HueShift:             neutral.HueShift + (mod.HueShift-neutral.HueShift)*intensity,
		SaturationMultiplier: neutral.SaturationMultiplier + (mod.SaturationMultiplier-neutral.SaturationMultiplier)*intensity,
		LightnessOffset:      neutral.LightnessOffset + (mod.LightnessOffset-neutral.LightnessOffset)*intensity,
		TemperatureShift:     neutral.TemperatureShift + (mod.TemperatureShift-neutral.TemperatureShift)*intensity,
	}
}

// ApplyTimeModulation applies time-based color modulation to a palette.
// Returns a new palette with adjusted colors based on time of day.
func ApplyTimeModulation(palette *Palette, config TimeConfig) *Palette {
	if palette == nil {
		return nil
	}

	modulation := GetModulationWithTransition(config)

	// Create new palette with modulated colors
	result := &Palette{
		Primary:    modulateColor(palette.Primary, modulation),
		Secondary:  modulateColor(palette.Secondary, modulation),
		Background: modulateColor(palette.Background, modulation),
		Text:       modulateColor(palette.Text, modulation),
		Accent1:    modulateColor(palette.Accent1, modulation),
		Accent2:    modulateColor(palette.Accent2, modulation),
		Accent3:    modulateColor(palette.Accent3, modulation),
		Highlight1: modulateColor(palette.Highlight1, modulation),
		Highlight2: modulateColor(palette.Highlight2, modulation),
		Shadow1:    modulateColor(palette.Shadow1, modulation),
		Shadow2:    modulateColor(palette.Shadow2, modulation),
		Neutral:    modulateColor(palette.Neutral, modulation),
		Danger:     modulateColor(palette.Danger, modulation),
		Success:    modulateColor(palette.Success, modulation),
		Warning:    modulateColor(palette.Warning, modulation),
		Info:       modulateColor(palette.Info, modulation),
	}

	// Modulate additional colors
	result.Colors = make([]color.Color, len(palette.Colors))
	for i, c := range palette.Colors {
		result.Colors[i] = modulateColor(c, modulation)
	}

	return result
}

// modulateColor applies color modulation to a single color.
func modulateColor(c color.Color, mod ColorModulation) color.Color {
	if c == nil {
		return c
	}

	r, g, b, a := c.RGBA()

	// Convert to 0-255 range
	r8 := uint8(r >> 8)
	g8 := uint8(g >> 8)
	b8 := uint8(b >> 8)
	a8 := uint8(a >> 8)

	// Convert to HSL
	h, s, l := rgbToHSL(r8, g8, b8)

	// Apply modulation
	h = math.Mod(h+mod.HueShift+360.0, 360.0)
	s = clamp(s*mod.SaturationMultiplier, 0.0, 1.0)
	l = clamp(l+mod.LightnessOffset, 0.0, 1.0)

	// Apply temperature shift (affects hue based on current lightness)
	if mod.TemperatureShift != 0.0 {
		// Temperature shift is stronger in mid-tones
		tempIntensity := 1.0 - math.Abs(l-0.5)*2.0               // Peak at l=0.5
		tempShift := mod.TemperatureShift * 30.0 * tempIntensity // Up to ±30° hue shift
		h = math.Mod(h+tempShift+360.0, 360.0)
	}

	// Convert back to RGB
	r8, g8, b8 = hslToRGB(h, s, l)

	return color.RGBA{R: r8, G: g8, B: b8, A: a8}
}

// rgbToHSL converts RGB (0-255) to HSL (H: 0-360, S: 0-1, L: 0-1).
func rgbToHSL(r, g, b uint8) (h, s, l float64) {
	rf := float64(r) / 255.0
	gf := float64(g) / 255.0
	bf := float64(b) / 255.0

	max := math.Max(rf, math.Max(gf, bf))
	min := math.Min(rf, math.Min(gf, bf))

	l = (max + min) / 2.0

	if max == min {
		h = 0.0
		s = 0.0
	} else {
		delta := max - min

		if l > 0.5 {
			s = delta / (2.0 - max - min)
		} else {
			s = delta / (max + min)
		}

		switch max {
		case rf:
			h = (gf - bf) / delta
			if gf < bf {
				h += 6.0
			}
		case gf:
			h = (bf-rf)/delta + 2.0
		case bf:
			h = (rf-gf)/delta + 4.0
		}

		h *= 60.0
	}

	return h, s, l
}

// hslToRGB converts HSL (H: 0-360, S: 0-1, L: 0-1) to RGB (0-255).
func hslToRGB(h, s, l float64) (r, g, b uint8) {
	if s == 0.0 {
		// Achromatic (gray)
		val := uint8(l * 255.0)
		return val, val, val
	}

	var q float64
	if l < 0.5 {
		q = l * (1.0 + s)
	} else {
		q = l + s - l*s
	}

	p := 2.0*l - q

	hk := h / 360.0

	tr := hk + 1.0/3.0
	tg := hk
	tb := hk - 1.0/3.0

	r = uint8(hueToRGB(p, q, tr) * 255.0)
	g = uint8(hueToRGB(p, q, tg) * 255.0)
	b = uint8(hueToRGB(p, q, tb) * 255.0)

	return r, g, b
}
