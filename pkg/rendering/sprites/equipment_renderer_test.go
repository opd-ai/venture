package sprites

import (
	"image/color"
	"math/rand"
	"testing"
)

func TestNewEquipmentRenderer(t *testing.T) {
	r := NewEquipmentRenderer()
	if r == nil {
		t.Fatal("expected non-nil renderer")
	}
}

func TestEquipmentRenderer_ColorHelpers(t *testing.T) {
	r := NewEquipmentRenderer()

	tests := []struct {
		name string
		fn   func() color.RGBA
		want color.RGBA
	}{
		{
			"lighten_50",
			func() color.RGBA { return r.lighten(color.RGBA{R: 100, G: 100, B: 100, A: 255}, 50) },
			color.RGBA{R: 150, G: 150, B: 150, A: 255},
		},
		{
			"darken_50",
			func() color.RGBA { return r.darken(color.RGBA{R: 100, G: 100, B: 100, A: 255}, 50) },
			color.RGBA{R: 50, G: 50, B: 50, A: 255},
		},
		{
			"lighten_clamp_overflow",
			func() color.RGBA { return r.lighten(color.RGBA{R: 250, G: 250, B: 250, A: 255}, 100) },
			color.RGBA{R: 255, G: 255, B: 255, A: 255},
		},
		{
			"darken_clamp_underflow",
			func() color.RGBA { return r.darken(color.RGBA{R: 5, G: 5, B: 5, A: 255}, 100) },
			color.RGBA{R: 0, G: 0, B: 0, A: 255},
		},
		{
			"lighten_zero",
			func() color.RGBA { return r.lighten(color.RGBA{R: 128, G: 128, B: 128, A: 200}, 0) },
			color.RGBA{R: 128, G: 128, B: 128, A: 200},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn()
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEquipmentRenderer_BlendColor(t *testing.T) {
	r := NewEquipmentRenderer()
	tests := []struct {
		name string
		a, b color.RGBA
		t_   float64
		want color.RGBA
	}{
		{
			"half_blend",
			color.RGBA{R: 0, G: 0, B: 0, A: 255},
			color.RGBA{R: 200, G: 200, B: 200, A: 255},
			0.5,
			color.RGBA{R: 100, G: 100, B: 100, A: 255},
		},
		{
			"no_blend",
			color.RGBA{R: 100, G: 150, B: 200, A: 255},
			color.RGBA{R: 50, G: 50, B: 50, A: 255},
			0.0,
			color.RGBA{R: 100, G: 150, B: 200, A: 255},
		},
		{
			"full_blend",
			color.RGBA{R: 100, G: 150, B: 200, A: 255},
			color.RGBA{R: 50, G: 50, B: 50, A: 255},
			1.0,
			color.RGBA{R: 50, G: 50, B: 50, A: 255},
		},
		{
			"clamp_negative_t",
			color.RGBA{R: 100, G: 100, B: 100, A: 255},
			color.RGBA{R: 200, G: 200, B: 200, A: 255},
			-0.5,
			color.RGBA{R: 100, G: 100, B: 100, A: 255},
		},
		{
			"clamp_over_t",
			color.RGBA{R: 100, G: 100, B: 100, A: 255},
			color.RGBA{R: 200, G: 200, B: 200, A: 255},
			2.0,
			color.RGBA{R: 200, G: 200, B: 200, A: 255},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r.blendColor(tt.a, tt.b, tt.t_)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEquipmentRenderer_MaterialBaseColor(t *testing.T) {
	r := NewEquipmentRenderer()
	materials := []struct {
		name     string
		material MaterialType
	}{
		{"metal", MaterialMetal},
		{"leather", MaterialLeather},
		{"cloth", MaterialCloth},
		{"wood", MaterialWood},
		{"crystal", MaterialCrystal},
		{"energy", MaterialEnergy},
		{"unknown", MaterialType(99)},
	}

	for _, mt := range materials {
		t.Run(mt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(42))
			c := r.materialBaseColor(mt.material, rng)
			if c.A != 255 {
				t.Errorf("expected full alpha, got %d", c.A)
			}
		})
	}
}

func TestEquipmentRenderer_MaterialBaseColor_Deterministic(t *testing.T) {
	r := NewEquipmentRenderer()

	rng1 := rand.New(rand.NewSource(123))
	c1 := r.materialBaseColor(MaterialMetal, rng1)

	rng2 := rand.New(rand.NewSource(123))
	c2 := r.materialBaseColor(MaterialMetal, rng2)

	if c1 != c2 {
		t.Errorf("same seed should produce same color: %v != %v", c1, c2)
	}
}

func TestEquipmentRenderer_MaterialBaseColor_Variety(t *testing.T) {
	r := NewEquipmentRenderer()
	seen := make(map[color.RGBA]bool)

	for seed := int64(0); seed < 20; seed++ {
		rng := rand.New(rand.NewSource(seed))
		c := r.materialBaseColor(MaterialMetal, rng)
		seen[c] = true
	}

	if len(seen) < 5 {
		t.Errorf("expected variety in material colors, only got %d distinct values", len(seen))
	}
}

func TestEquipmentRenderer_EnchantmentColor(t *testing.T) {
	r := NewEquipmentRenderer()
	tests := []struct {
		name  string
		color string
	}{
		{"green", "green"},
		{"blue", "blue"},
		{"purple", "purple"},
		{"gold", "gold"},
		{"red", "red"},
		{"unknown", "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := r.enchantmentColor(tt.color)
			if c.A != 255 {
				t.Errorf("enchantment color %s: expected full alpha, got %d", tt.color, c.A)
			}
		})
	}
}

func TestEquipmentRenderer_EnchantmentColor_Distinct(t *testing.T) {
	r := NewEquipmentRenderer()
	colors := []string{"green", "blue", "purple", "gold", "red"}
	seen := make(map[color.RGBA]bool)

	for _, name := range colors {
		c := r.enchantmentColor(name)
		if seen[c] {
			t.Errorf("enchantment color %s is not distinct", name)
		}
		seen[c] = true
	}
}

func TestEquipmentRenderer_WithAlpha(t *testing.T) {
	r := NewEquipmentRenderer()
	tests := []struct {
		name  string
		base  color.RGBA
		alpha uint8
	}{
		{"full_to_half", color.RGBA{R: 100, G: 150, B: 200, A: 255}, 128},
		{"full_to_zero", color.RGBA{R: 100, G: 150, B: 200, A: 255}, 0},
		{"half_to_full", color.RGBA{R: 100, G: 150, B: 200, A: 128}, 255},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := r.withAlpha(tt.base, tt.alpha)
			if c.A != tt.alpha {
				t.Errorf("expected alpha %d, got %d", tt.alpha, c.A)
			}
			if c.R != tt.base.R || c.G != tt.base.G || c.B != tt.base.B {
				t.Errorf("RGB should be preserved: expected %d,%d,%d, got %d,%d,%d",
					tt.base.R, tt.base.G, tt.base.B, c.R, c.G, c.B)
			}
		})
	}
}

func TestEquipmentRenderer_MaterialDiffers(t *testing.T) {
	r := NewEquipmentRenderer()
	rng1 := rand.New(rand.NewSource(42))
	metal := r.materialBaseColor(MaterialMetal, rng1)

	rng2 := rand.New(rand.NewSource(42))
	leather := r.materialBaseColor(MaterialLeather, rng2)

	rng3 := rand.New(rand.NewSource(42))
	crystal := r.materialBaseColor(MaterialCrystal, rng3)

	if metal == leather {
		t.Error("metal and leather should produce different colors")
	}
	if metal == crystal {
		t.Error("metal and crystal should produce different colors")
	}
}

func TestMinInt(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 1},
		{5, 3, 3},
		{0, 0, 0},
		{-1, 1, -1},
	}
	for _, tt := range tests {
		if got := minInt(tt.a, tt.b); got != tt.want {
			t.Errorf("minInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestMaxInt_Equipment(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{1, 2, 2},
		{5, 3, 5},
		{0, 0, 0},
		{-1, 1, 1},
	}
	for _, tt := range tests {
		if got := maxInt(tt.a, tt.b); got != tt.want {
			t.Errorf("maxInt(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}
