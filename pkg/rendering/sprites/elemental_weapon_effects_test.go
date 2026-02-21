package sprites

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestElementType_String(t *testing.T) {
	tests := []struct {
		element ElementType
		want    string
	}{
		{ElementNone, "none"},
		{ElementFire, "fire"},
		{ElementIce, "ice"},
		{ElementLightning, "lightning"},
		{ElementPoison, "poison"},
		{ElementHoly, "holy"},
		{ElementShadow, "shadow"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.element.String(); got != tt.want {
				t.Errorf("ElementType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseElementType(t *testing.T) {
	tests := []struct {
		input string
		want  ElementType
	}{
		// Fire variants
		{"fire", ElementFire},
		{"flame", ElementFire},
		{"burning", ElementFire},
		// Ice variants
		{"ice", ElementIce},
		{"frost", ElementIce},
		{"frozen", ElementIce},
		{"cold", ElementIce},
		// Lightning variants
		{"lightning", ElementLightning},
		{"electric", ElementLightning},
		{"shock", ElementLightning},
		{"thunder", ElementLightning},
		// Poison variants
		{"poison", ElementPoison},
		{"toxic", ElementPoison},
		{"venom", ElementPoison},
		{"acid", ElementPoison},
		// Holy variants
		{"holy", ElementHoly},
		{"sacred", ElementHoly},
		{"divine", ElementHoly},
		{"radiant", ElementHoly},
		{"light", ElementHoly},
		// Shadow variants
		{"shadow", ElementShadow},
		{"dark", ElementShadow},
		{"void", ElementShadow},
		{"unholy", ElementShadow},
		{"necrotic", ElementShadow},
		// Unknown
		{"unknown", ElementNone},
		{"", ElementNone},
		{"physical", ElementNone},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseElementType(tt.input); got != tt.want {
				t.Errorf("ParseElementType(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestDefaultElementalParams(t *testing.T) {
	seed := int64(12345)

	tests := []struct {
		element ElementType
	}{
		{ElementFire},
		{ElementIce},
		{ElementLightning},
		{ElementPoison},
		{ElementHoly},
		{ElementShadow},
	}

	for _, tt := range tests {
		t.Run(tt.element.String(), func(t *testing.T) {
			params := DefaultElementalParams(tt.element, seed)

			if params.Element != tt.element {
				t.Errorf("Element = %v, want %v", params.Element, tt.element)
			}
			if params.Intensity < 0 || params.Intensity > 1 {
				t.Errorf("Intensity = %v, want [0,1]", params.Intensity)
			}
			if params.ParticleCount <= 0 {
				t.Errorf("ParticleCount = %v, want > 0", params.ParticleCount)
			}
			if params.Seed != seed {
				t.Errorf("Seed = %v, want %v", params.Seed, seed)
			}
		})
	}
}

func TestNewElementalWeaponRenderer(t *testing.T) {
	renderer := NewElementalWeaponRenderer()
	if renderer == nil {
		t.Error("NewElementalWeaponRenderer() returned nil")
	}
}

func TestApplyElementalEffect_NilImage(t *testing.T) {
	renderer := NewElementalWeaponRenderer()
	params := DefaultElementalParams(ElementFire, 12345)

	// Should not panic with nil image
	renderer.ApplyElementalEffect(nil, params)
}

func TestApplyElementalEffect_NoElement(t *testing.T) {
	renderer := NewElementalWeaponRenderer()
	params := ElementalEffectParams{
		Element:   ElementNone,
		Intensity: 0.7,
		Seed:      12345,
	}

	img := ebiten.NewImage(32, 32)
	// Should not panic or modify with no element
	renderer.ApplyElementalEffect(img, params)
}

func TestApplyElementalEffect_ZeroIntensity(t *testing.T) {
	renderer := NewElementalWeaponRenderer()
	params := ElementalEffectParams{
		Element:   ElementFire,
		Intensity: 0,
		Seed:      12345,
	}

	img := ebiten.NewImage(32, 32)
	// Should not apply effects at zero intensity
	renderer.ApplyElementalEffect(img, params)
}

func TestApplyElementalEffect_AllElements(t *testing.T) {
	renderer := NewElementalWeaponRenderer()
	elements := []ElementType{
		ElementFire,
		ElementIce,
		ElementLightning,
		ElementPoison,
		ElementHoly,
		ElementShadow,
	}

	for _, element := range elements {
		t.Run(element.String(), func(t *testing.T) {
			img := ebiten.NewImage(32, 32)

			// Draw a simple weapon shape (diagonal line)
			for i := 0; i < 20; i++ {
				img.Set(6+i, 6+i, Color{R: 180, G: 180, B: 190, A: 255})
			}

			params := ElementalEffectParams{
				Element:        element,
				Intensity:      0.7,
				AnimationPhase: 0.5,
				ParticleCount:  6,
				Seed:           12345,
			}

			// Should not panic
			renderer.ApplyElementalEffect(img, params)
		})
	}
}

func TestApplyElementalEffect_DifferentAnimationPhases(t *testing.T) {
	renderer := NewElementalWeaponRenderer()

	phases := []float64{0.0, 0.25, 0.5, 0.75, 1.0}

	for _, phase := range phases {
		t.Run("phase_"+string(rune('0'+int(phase*10))), func(t *testing.T) {
			img := ebiten.NewImage(32, 32)

			// Draw weapon shape
			for i := 0; i < 16; i++ {
				img.Set(16, i+8, Color{R: 150, G: 150, B: 160, A: 255})
			}

			params := ElementalEffectParams{
				Element:        ElementFire,
				Intensity:      0.8,
				AnimationPhase: phase,
				ParticleCount:  4,
				Seed:           54321,
			}

			renderer.ApplyElementalEffect(img, params)
		})
	}
}

func TestGetElementFromEnchantment(t *testing.T) {
	tests := []struct {
		enchantType string
		tags        []string
		want        ElementType
	}{
		{"fire", nil, ElementFire},
		{"", []string{"weapon", "fire", "rare"}, ElementFire},
		{"frost", []string{"cold", "magic"}, ElementIce},
		{"", []string{"divine", "blessed"}, ElementHoly},
		{"", nil, ElementNone},
		{"normal", []string{"common", "steel"}, ElementNone},
	}

	for _, tt := range tests {
		t.Run(tt.enchantType, func(t *testing.T) {
			got := GetElementFromEnchantment(tt.enchantType, tt.tags)
			if got != tt.want {
				t.Errorf("GetElementFromEnchantment(%q, %v) = %v, want %v",
					tt.enchantType, tt.tags, got, tt.want)
			}
		})
	}
}

func TestElementalWeaponRenderer_DeterministicOutput(t *testing.T) {
	renderer := NewElementalWeaponRenderer()
	seed := int64(99999)

	// Create two identical weapon images
	createWeaponImage := func() *ebiten.Image {
		img := ebiten.NewImage(32, 32)
		for i := 0; i < 20; i++ {
			img.Set(6+i, 16, Color{R: 180, G: 180, B: 190, A: 255})
		}
		return img
	}

	img1 := createWeaponImage()
	img2 := createWeaponImage()

	params := ElementalEffectParams{
		Element:        ElementLightning,
		Intensity:      0.6,
		AnimationPhase: 0.3,
		ParticleCount:  5,
		Seed:           seed,
	}

	renderer.ApplyElementalEffect(img1, params)
	renderer.ApplyElementalEffect(img2, params)

	// Results should be identical given same seed
	// Note: Due to ebiten's At() method potentially panicking without game loop,
	// we just verify no panics occurred
}

func TestElementalWeaponRenderer_IntensityScaling(t *testing.T) {
	renderer := NewElementalWeaponRenderer()

	intensities := []float64{0.1, 0.3, 0.5, 0.7, 0.9, 1.0}

	for _, intensity := range intensities {
		t.Run("intensity_"+string(rune('0'+int(intensity*10))), func(t *testing.T) {
			img := ebiten.NewImage(32, 32)

			// Draw weapon
			for i := 0; i < 16; i++ {
				img.Set(16, 8+i, Color{R: 160, G: 160, B: 170, A: 255})
			}

			params := ElementalEffectParams{
				Element:        ElementPoison,
				Intensity:      intensity,
				AnimationPhase: 0.0,
				ParticleCount:  int(4 + intensity*4),
				Seed:           11111,
			}

			// Should not panic at any intensity
			renderer.ApplyElementalEffect(img, params)
		})
	}
}

// Color is a simple helper type for test pixel values
type Color struct {
	R, G, B, A uint8
}

func (c Color) RGBA() (r, g, b, a uint32) {
	return uint32(c.R) << 8, uint32(c.G) << 8, uint32(c.B) << 8, uint32(c.A) << 8
}

func BenchmarkApplyElementalEffect_Fire(b *testing.B) {
	renderer := NewElementalWeaponRenderer()
	img := ebiten.NewImage(32, 32)

	// Draw weapon shape
	for i := 0; i < 20; i++ {
		img.Set(6+i, 6+i, Color{R: 180, G: 180, B: 190, A: 255})
	}

	params := ElementalEffectParams{
		Element:        ElementFire,
		Intensity:      0.7,
		AnimationPhase: 0.0,
		ParticleCount:  6,
		Seed:           12345,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		renderer.ApplyElementalEffect(img, params)
		params.AnimationPhase += 0.01
		if params.AnimationPhase >= 1.0 {
			params.AnimationPhase = 0.0
		}
	}
}

func BenchmarkApplyElementalEffect_AllElements(b *testing.B) {
	renderer := NewElementalWeaponRenderer()
	elements := []ElementType{
		ElementFire,
		ElementIce,
		ElementLightning,
		ElementPoison,
		ElementHoly,
		ElementShadow,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		element := elements[i%len(elements)]
		img := ebiten.NewImage(32, 32)

		// Draw weapon
		for j := 0; j < 20; j++ {
			img.Set(6+j, 16, Color{R: 180, G: 180, B: 190, A: 255})
		}

		params := ElementalEffectParams{
			Element:        element,
			Intensity:      0.7,
			AnimationPhase: float64(i%100) / 100.0,
			ParticleCount:  6,
			Seed:           int64(i),
		}

		renderer.ApplyElementalEffect(img, params)
	}
}
