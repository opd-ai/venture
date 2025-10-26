package sprites

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// TestGenerateComposite_Basic tests basic composite sprite generation.
func TestGenerateComposite_Basic(t *testing.T) {
	gen := NewGenerator()

	pal, err := palette.NewGenerator().Generate("fantasy", 12345)
	if err != nil {
		t.Fatalf("Failed to generate palette: %v", err)
	}

	config := CompositeConfig{
		BaseConfig: Config{
			Type:       SpriteEntity,
			Width:      28,
			Height:     28,
			Seed:       12345,
			GenreID:    "fantasy",
			Complexity: 0.5,
			Palette:    pal,
		},
		Layers: []LayerConfig{
			{
				Type:      LayerBody,
				ZIndex:    10,
				OffsetX:   0,
				OffsetY:   0,
				Scale:     1.0,
				Visible:   true,
				Seed:      12345,
				ShapeType: shapes.ShapeCircle,
			},
		},
		Equipment:     []EquipmentVisual{},
		StatusEffects: []StatusEffect{},
	}

	img, err := gen.GenerateComposite(config)
	if err != nil {
		t.Fatalf("Failed to generate composite: %v", err)
	}

	if img == nil {
		t.Fatal("Expected non-nil image")
	}

	bounds := img.Bounds()
	if bounds.Dx() != 28 || bounds.Dy() != 28 {
		t.Errorf("Expected 28x28 image, got %dx%d", bounds.Dx(), bounds.Dy())
	}
}

// TestGenerateComposite_MultipleLayers tests multi-layer composition.
func TestGenerateComposite_MultipleLayers(t *testing.T) {
	gen := NewGenerator()

	pal, _ := palette.NewGenerator().Generate("fantasy", 12345)

	config := CompositeConfig{
		BaseConfig: Config{
			Type:       SpriteEntity,
			Width:      28,
			Height:     28,
			Seed:       12345,
			Palette:    pal,
			Complexity: 0.5,
		},
		Layers: []LayerConfig{
			{Type: LayerBody, ZIndex: 10, Scale: 1.0, Visible: true, Seed: 12345, ShapeType: shapes.ShapeCircle},
			{Type: LayerHead, ZIndex: 20, Scale: 1.0, Visible: true, Seed: 12346, ShapeType: shapes.ShapeCircle, OffsetY: -8},
			{Type: LayerLegs, ZIndex: 5, Scale: 1.0, Visible: true, Seed: 12347, ShapeType: shapes.ShapeRectangle, OffsetY: 8},
		},
		Equipment:     []EquipmentVisual{},
		StatusEffects: []StatusEffect{},
	}

	img, err := gen.GenerateComposite(config)
	if err != nil {
		t.Fatalf("Failed to generate composite with multiple layers: %v", err)
	}

	if img == nil {
		t.Fatal("Expected non-nil image")
	}
}

// TestGenerateComposite_WithEquipment tests equipment layer generation.
func TestGenerateComposite_WithEquipment(t *testing.T) {
	gen := NewGenerator()

	pal, _ := palette.NewGenerator().Generate("fantasy", 12345)

	config := CompositeConfig{
		BaseConfig: Config{
			Type:       SpriteEntity,
			Width:      28,
			Height:     28,
			Seed:       12345,
			Palette:    pal,
			Complexity: 0.5,
		},
		Layers: []LayerConfig{
			{Type: LayerBody, ZIndex: 10, Scale: 1.0, Visible: true, Seed: 12345, ShapeType: shapes.ShapeCircle},
		},
		Equipment: []EquipmentVisual{
			{
				Slot:   "weapon",
				ItemID: "sword_001",
				Seed:   54321,
				Layer:  LayerWeapon,
				Params: make(map[string]interface{}),
			},
		},
		StatusEffects: []StatusEffect{},
	}

	img, err := gen.GenerateComposite(config)
	if err != nil {
		t.Fatalf("Failed to generate composite with equipment: %v", err)
	}

	if img == nil {
		t.Fatal("Expected non-nil image")
	}
}

// TestGenerateComposite_WithStatusEffects tests status effect overlays.
func TestGenerateComposite_WithStatusEffects(t *testing.T) {
	gen := NewGenerator()

	pal, _ := palette.NewGenerator().Generate("fantasy", 12345)

	config := CompositeConfig{
		BaseConfig: Config{
			Type:       SpriteEntity,
			Width:      28,
			Height:     28,
			Seed:       12345,
			Palette:    pal,
			Complexity: 0.5,
		},
		Layers: []LayerConfig{
			{Type: LayerBody, ZIndex: 10, Scale: 1.0, Visible: true, Seed: 12345, ShapeType: shapes.ShapeCircle},
		},
		Equipment: []EquipmentVisual{},
		StatusEffects: []StatusEffect{
			{
				Type:          "burning",
				Intensity:     0.8,
				Color:         "red",
				AnimSpeed:     1.0,
				ParticleCount: 10,
			},
		},
	}

	img, err := gen.GenerateComposite(config)
	if err != nil {
		t.Fatalf("Failed to generate composite with status effects: %v", err)
	}

	if img == nil {
		t.Fatal("Expected non-nil image")
	}
}

// TestGenerateComposite_LayerOrdering tests Z-index based layer ordering.
func TestGenerateComposite_LayerOrdering(t *testing.T) {
	gen := NewGenerator()

	pal, _ := palette.NewGenerator().Generate("fantasy", 12345)

	// Layers in non-sorted order
	config := CompositeConfig{
		BaseConfig: Config{
			Type:    SpriteEntity,
			Width:   28,
			Height:  28,
			Seed:    12345,
			Palette: pal,
		},
		Layers: []LayerConfig{
			{Type: LayerHead, ZIndex: 30, Scale: 1.0, Visible: true, Seed: 12346, ShapeType: shapes.ShapeCircle},
			{Type: LayerBody, ZIndex: 10, Scale: 1.0, Visible: true, Seed: 12345, ShapeType: shapes.ShapeCircle},
			{Type: LayerLegs, ZIndex: 5, Scale: 1.0, Visible: true, Seed: 12347, ShapeType: shapes.ShapeRectangle},
		},
		Equipment:     []EquipmentVisual{},
		StatusEffects: []StatusEffect{},
	}

	// Should render successfully despite unsorted input
	img, err := gen.GenerateComposite(config)
	if err != nil {
		t.Fatalf("Failed to generate composite: %v", err)
	}

	if img == nil {
		t.Fatal("Expected non-nil image")
	}
}

// TestGenerateComposite_InvisibleLayers tests layer visibility.
func TestGenerateComposite_InvisibleLayers(t *testing.T) {
	gen := NewGenerator()

	pal, _ := palette.NewGenerator().Generate("fantasy", 12345)

	config := CompositeConfig{
		BaseConfig: Config{
			Type:    SpriteEntity,
			Width:   28,
			Height:  28,
			Seed:    12345,
			Palette: pal,
		},
		Layers: []LayerConfig{
			{Type: LayerBody, ZIndex: 10, Scale: 1.0, Visible: true, Seed: 12345, ShapeType: shapes.ShapeCircle},
			{Type: LayerHead, ZIndex: 20, Scale: 1.0, Visible: false, Seed: 12346, ShapeType: shapes.ShapeCircle}, // Invisible
		},
		Equipment:     []EquipmentVisual{},
		StatusEffects: []StatusEffect{},
	}

	img, err := gen.GenerateComposite(config)
	if err != nil {
		t.Fatalf("Failed to generate composite: %v", err)
	}

	if img == nil {
		t.Fatal("Expected non-nil image")
	}
}

// TestGenerateComposite_InvalidConfig tests validation.
func TestGenerateComposite_InvalidConfig(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name   string
		config CompositeConfig
	}{
		{
			name: "zero dimensions",
			config: CompositeConfig{
				BaseConfig: Config{Width: 0, Height: 0},
				Layers:     []LayerConfig{{Type: LayerBody}},
			},
		},
		{
			name: "no layers",
			config: CompositeConfig{
				BaseConfig: Config{Width: 28, Height: 28},
				Layers:     []LayerConfig{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gen.GenerateComposite(tt.config)
			if err == nil {
				t.Error("Expected error for invalid config")
			}
		})
	}
}

// TestLayerType_String tests layer type string representation.
func TestLayerType_String(t *testing.T) {
	tests := []struct {
		layerType LayerType
		expected  string
	}{
		{LayerBody, "body"},
		{LayerHead, "head"},
		{LayerLegs, "legs"},
		{LayerWeapon, "weapon"},
		{LayerArmor, "armor"},
		{LayerAccessory, "accessory"},
		{LayerEffect, "effect"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.layerType.String() != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, tt.layerType.String())
			}
		})
	}
}

// TestGetStatusEffectColor tests status effect color mapping.
func TestGetStatusEffectColor(t *testing.T) {
	gen := NewGenerator()

	tests := []string{
		"burning", "frozen", "poisoned", "stunned", "blessed", "cursed",
	}

	for _, effectType := range tests {
		t.Run(effectType, func(t *testing.T) {
			color := gen.getStatusEffectColor(effectType, "")
			if color == nil {
				t.Error("Expected non-nil color")
			}
		})
	}
}

// TestGetEquipmentShapeType tests equipment shape selection.
func TestGetEquipmentShapeType(t *testing.T) {
	gen := NewGenerator()
	rng := NewTestRNG(12345)

	tests := []string{"weapon", "armor", "accessory"}

	for _, slot := range tests {
		t.Run(slot, func(t *testing.T) {
			shapeType := gen.getEquipmentShapeType(slot, rng)
			if shapeType < 0 || shapeType > shapes.ShapeRing {
				t.Errorf("Invalid shape type: %d", shapeType)
			}
		})
	}
}

// BenchmarkGenerateComposite benchmarks composite sprite generation.
func BenchmarkGenerateComposite(b *testing.B) {
	gen := NewGenerator()
	pal, _ := palette.NewGenerator().Generate("fantasy", 12345)

	config := CompositeConfig{
		BaseConfig: Config{
			Type:    SpriteEntity,
			Width:   28,
			Height:  28,
			Seed:    12345,
			Palette: pal,
		},
		Layers: []LayerConfig{
			{Type: LayerBody, ZIndex: 10, Scale: 1.0, Visible: true, Seed: 12345, ShapeType: shapes.ShapeCircle},
			{Type: LayerHead, ZIndex: 20, Scale: 1.0, Visible: true, Seed: 12346, ShapeType: shapes.ShapeCircle},
		},
		Equipment:     []EquipmentVisual{},
		StatusEffects: []StatusEffect{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateComposite(config)
	}
}

// BenchmarkGenerateComposite_WithEquipment benchmarks with equipment.
func BenchmarkGenerateComposite_WithEquipment(b *testing.B) {
	gen := NewGenerator()
	pal, _ := palette.NewGenerator().Generate("fantasy", 12345)

	config := CompositeConfig{
		BaseConfig: Config{
			Type:    SpriteEntity,
			Width:   28,
			Height:  28,
			Seed:    12345,
			Palette: pal,
		},
		Layers: []LayerConfig{
			{Type: LayerBody, ZIndex: 10, Scale: 1.0, Visible: true, Seed: 12345, ShapeType: shapes.ShapeCircle},
		},
		Equipment: []EquipmentVisual{
			{Slot: "weapon", Seed: 54321, Layer: LayerWeapon},
			{Slot: "armor", Seed: 54322, Layer: LayerArmor},
		},
		StatusEffects: []StatusEffect{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.GenerateComposite(config)
	}
}

// Helper for testing
func NewTestRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
