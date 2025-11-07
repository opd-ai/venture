// Package postprocess provides post-processing effects for rendered scenes.
package postprocess

import (
	"image"
	"image/color"
	"math"
)

// Processor manages post-processing effects and applies them to images.
type Processor struct {
	config Config
}

// NewProcessor creates a new post-processing processor with default configuration.
func NewProcessor() *Processor {
	return &Processor{
		config: DefaultConfig(),
	}
}

// NewProcessorWithConfig creates a new post-processing processor with custom configuration.
func NewProcessorWithConfig(config Config) *Processor {
	return &Processor{
		config: config,
	}
}

// SetConfig sets the post-processing configuration.
func (p *Processor) SetConfig(config Config) {
	p.config = config
}

// GetConfig returns the current post-processing configuration.
func (p *Processor) GetConfig() Config {
	return p.config
}

// ApplyAll applies all enabled post-processing effects in the recommended order.
// velocityMap is optional (for motion blur), depthMap is optional (for depth blur).
// If both motion blur and depth blur are enabled, only motion blur is applied
// since they are mutually exclusive blur effects.
func (p *Processor) ApplyAll(img *image.RGBA, velocityMap *VelocityMap, depthMap *image.RGBA) *image.RGBA {
	result := img

	// Apply blur effects (motion blur takes priority over depth blur)
	if p.config.MotionBlur.Enabled && velocityMap != nil {
		result = p.ApplyMotionBlur(result, velocityMap)
	} else if p.config.DepthBlur.Enabled && depthMap != nil {
		result = p.ApplyDepthBlur(result, depthMap)
	}

	// Apply color grading
	if p.config.ColorGrading.Enabled {
		result = p.ApplyColorGrading(result)
	}

	// Apply vignette
	if p.config.Vignette.Enabled {
		result = p.ApplyVignette(result)
	}

	// Apply chromatic aberration
	if p.config.ChromaticAberration.Enabled {
		result = p.ApplyChromaticAberration(result)
	}

	return result
}

// clamp restricts a value to a specified range.
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// clampUint8 restricts a value to uint8 range (0-255).
func clampUint8(value float64) uint8 {
	if value < 0 {
		return 0
	}
	if value > 255 {
		return 255
	}
	return uint8(value)
}

// lerp performs linear interpolation between two values.
func lerp(a, b, t float64) float64 {
	return a + (b-a)*t
}

// smoothstep performs smooth interpolation using Hermite polynomial.
func smoothstep(edge0, edge1, x float64) float64 {
	t := clamp((x-edge0)/(edge1-edge0), 0.0, 1.0)
	return t * t * (3.0 - 2.0*t)
}

// rgbToHSL converts RGB color to HSL color space.
func rgbToHSL(r, g, b float64) (h, s, l float64) {
	max := math.Max(math.Max(r, g), b)
	min := math.Min(math.Min(r, g), b)
	l = (max + min) / 2.0

	if max == min {
		// Achromatic (gray)
		h = 0
		s = 0
	} else {
		d := max - min
		if l > 0.5 {
			s = d / (2.0 - max - min)
		} else {
			s = d / (max + min)
		}

		switch max {
		case r:
			h = (g - b) / d
			if g < b {
				h += 6.0
			}
		case g:
			h = (b-r)/d + 2.0
		case b:
			h = (r-g)/d + 4.0
		}
		h /= 6.0
	}

	return h, s, l
}

// hslToRGB converts HSL color to RGB color space.
func hslToRGB(h, s, l float64) (r, g, b float64) {
	if s == 0 {
		// Achromatic (gray)
		r = l
		g = l
		b = l
	} else {
		var q float64
		if l < 0.5 {
			q = l * (1 + s)
		} else {
			q = l + s - l*s
		}
		p := 2*l - q

		r = hueToRGB(p, q, h+1.0/3.0)
		g = hueToRGB(p, q, h)
		b = hueToRGB(p, q, h-1.0/3.0)
	}

	return r, g, b
}

// hueToRGB is a helper function for HSL to RGB conversion.
func hueToRGB(p, q, t float64) float64 {
	if t < 0 {
		t += 1
	}
	if t > 1 {
		t -= 1
	}
	if t < 1.0/6.0 {
		return p + (q-p)*6*t
	}
	if t < 1.0/2.0 {
		return q
	}
	if t < 2.0/3.0 {
		return p + (q-p)*(2.0/3.0-t)*6
	}
	return p
}

// luminance calculates the perceived brightness of an RGB color.
func luminance(r, g, b float64) float64 {
	return 0.299*r + 0.587*g + 0.114*b
}

// sampleBilinear samples an image at non-integer coordinates using bilinear interpolation.
func sampleBilinear(img *image.RGBA, x, y float64) color.RGBA {
	bounds := img.Bounds()

	// Clamp to image bounds
	x = clamp(x, float64(bounds.Min.X), float64(bounds.Max.X-1))
	y = clamp(y, float64(bounds.Min.Y), float64(bounds.Max.Y-1))

	// Get integer coordinates and fractional parts
	x0 := int(math.Floor(x))
	y0 := int(math.Floor(y))
	x1 := int(math.Min(float64(x0+1), float64(bounds.Max.X-1)))
	y1 := int(math.Min(float64(y0+1), float64(bounds.Max.Y-1)))
	fx := x - float64(x0)
	fy := y - float64(y0)

	// Sample four corners
	c00 := img.RGBAAt(x0, y0)
	c10 := img.RGBAAt(x1, y0)
	c01 := img.RGBAAt(x0, y1)
	c11 := img.RGBAAt(x1, y1)

	// Bilinear interpolation
	r := (1-fx)*(1-fy)*float64(c00.R) + fx*(1-fy)*float64(c10.R) +
		(1-fx)*fy*float64(c01.R) + fx*fy*float64(c11.R)
	g := (1-fx)*(1-fy)*float64(c00.G) + fx*(1-fy)*float64(c10.G) +
		(1-fx)*fy*float64(c01.G) + fx*fy*float64(c11.G)
	b := (1-fx)*(1-fy)*float64(c00.B) + fx*(1-fy)*float64(c10.B) +
		(1-fx)*fy*float64(c01.B) + fx*fy*float64(c11.B)
	a := (1-fx)*(1-fy)*float64(c00.A) + fx*(1-fy)*float64(c10.A) +
		(1-fx)*fy*float64(c01.A) + fx*fy*float64(c11.A)

	return color.RGBA{
		R: clampUint8(r),
		G: clampUint8(g),
		B: clampUint8(b),
		A: clampUint8(a),
	}
}

// boxBlur applies a simple box blur to an image.
func boxBlur(img *image.RGBA, radius int) *image.RGBA {
	if radius < 1 {
		return img
	}

	bounds := img.Bounds()
	result := image.NewRGBA(bounds)

	// Horizontal pass
	temp := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b, a, count float64

			for dx := -radius; dx <= radius; dx++ {
				sx := x + dx
				if sx >= bounds.Min.X && sx < bounds.Max.X {
					c := img.RGBAAt(sx, y)
					r += float64(c.R)
					g += float64(c.G)
					b += float64(c.B)
					a += float64(c.A)
					count++
				}
			}

			temp.Set(x, y, color.RGBA{
				R: clampUint8(r / count),
				G: clampUint8(g / count),
				B: clampUint8(b / count),
				A: clampUint8(a / count),
			})
		}
	}

	// Vertical pass
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			var r, g, b, a, count float64

			for dy := -radius; dy <= radius; dy++ {
				sy := y + dy
				if sy >= bounds.Min.Y && sy < bounds.Max.Y {
					c := temp.RGBAAt(x, sy)
					r += float64(c.R)
					g += float64(c.G)
					b += float64(c.B)
					a += float64(c.A)
					count++
				}
			}

			result.Set(x, y, color.RGBA{
				R: clampUint8(r / count),
				G: clampUint8(g / count),
				B: clampUint8(b / count),
				A: clampUint8(a / count),
			})
		}
	}

	return result
}
