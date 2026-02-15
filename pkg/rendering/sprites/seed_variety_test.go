package sprites

import (
	"image/color"
	"math"
	"math/rand"
	"testing"
)

func TestGenerateAvatarTraits_Deterministic(t *testing.T) {
	// Same seed must produce identical traits
	t1 := GenerateAvatarTraits(12345)
	t2 := GenerateAvatarTraits(12345)

	if t1.SkinTone != t2.SkinTone {
		t.Errorf("SkinTone not deterministic: %v vs %v", t1.SkinTone, t2.SkinTone)
	}
	if t1.HairColor != t2.HairColor {
		t.Errorf("HairColor not deterministic: %v vs %v", t1.HairColor, t2.HairColor)
	}
	if t1.ClothingPrimary != t2.ClothingPrimary {
		t.Errorf("ClothingPrimary not deterministic")
	}
	if t1.ShoulderScale != t2.ShoulderScale {
		t.Errorf("ShoulderScale not deterministic: %f vs %f", t1.ShoulderScale, t2.ShoulderScale)
	}
}

func TestGenerateAvatarTraits_Variety(t *testing.T) {
	// Different seeds should produce different traits
	seeds := []int64{1, 2, 3, 100, 999, 5000, 123456}
	seen := make(map[color.RGBA]bool)
	for _, seed := range seeds {
		traits := GenerateAvatarTraits(seed)
		seen[traits.SkinTone] = true
		seen[traits.HairColor] = true
		seen[traits.ClothingPrimary] = true
	}
	// With 7 seeds * 3 colors = 21 possible, we should see significant variety
	if len(seen) < 10 {
		t.Errorf("insufficient variety: only %d unique colors from %d seeds", len(seen), len(seeds))
	}
}

func TestGenerateAvatarTraits_ValidRanges(t *testing.T) {
	tests := []struct {
		name string
		seed int64
	}{
		{"zero", 0},
		{"positive", 42},
		{"large", 9999999},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traits := GenerateAvatarTraits(tt.seed)

			if traits.ShoulderScale < 0.85 || traits.ShoulderScale > 1.15 {
				t.Errorf("ShoulderScale out of range: %f", traits.ShoulderScale)
			}
			if traits.HeadScale < 0.90 || traits.HeadScale > 1.10 {
				t.Errorf("HeadScale out of range: %f", traits.HeadScale)
			}
			if traits.HeightScale < 0.92 || traits.HeightScale > 1.08 {
				t.Errorf("HeightScale out of range: %f", traits.HeightScale)
			}
			if traits.SkinTone.A != 255 {
				t.Errorf("SkinTone alpha should be 255, got %d", traits.SkinTone.A)
			}
			if traits.HairColor.A != 255 {
				t.Errorf("HairColor alpha should be 255, got %d", traits.HairColor.A)
			}
		})
	}
}

func TestGenerateCreatureTraits(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		seed       int64
	}{
		{"spider", "spider", 42},
		{"serpent", "serpent", 42},
		{"dragon", "dragon", 100},
		{"blob", "blob", 200},
		{"undead", "undead", 300},
		{"quadruped", "quadruped", 400},
		{"unknown", "some_beast", 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traits := GenerateCreatureTraits(tt.seed, tt.entityType)

			// All colors should have full alpha
			if traits.SkinTone.A != 255 {
				t.Errorf("SkinTone alpha should be 255")
			}
			if traits.ClothingPrimary.A != 255 {
				t.Errorf("ClothingPrimary alpha should be 255")
			}

			// Proportions in valid range
			if traits.ShoulderScale < 0.85 || traits.ShoulderScale > 1.20 {
				t.Errorf("ShoulderScale out of range: %f", traits.ShoulderScale)
			}
		})
	}
}

func TestGenerateCreatureTraits_Deterministic(t *testing.T) {
	t1 := GenerateCreatureTraits(42, "spider")
	t2 := GenerateCreatureTraits(42, "spider")
	if t1.SkinTone != t2.SkinTone {
		t.Error("creature traits not deterministic")
	}
}

func TestColorForBodyPart(t *testing.T) {
	traits := GenerateAvatarTraits(42)

	tests := []struct {
		name     string
		role     string
		zIndex   int
		wantNil  bool
		wantDesc string
	}{
		{"head_secondary", "secondary", 15, false, "HairColor for head"},
		{"arms_secondary", "secondary", 8, false, "SkinTone for arms"},
		{"torso_primary", "primary", 10, false, "ClothingPrimary for torso"},
		{"legs_primary", "primary", 5, false, "LegColor for legs"},
		{"accent1", "accent1", 0, false, "ClothingSecondary"},
		{"accent2", "accent2", 0, false, "HairColor"},
		{"shadow", "shadow", 0, true, "nil for shadow"},
		{"highlight1", "highlight1", 0, true, "nil for highlight1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := traits.ColorForBodyPart(tt.role, tt.zIndex)
			if tt.wantNil && c != nil {
				t.Errorf("expected nil for %s, got %v", tt.wantDesc, c)
			}
			if !tt.wantNil && c == nil {
				t.Errorf("expected non-nil color for %s", tt.wantDesc)
			}
		})
	}
}

func TestApplyProportions(t *testing.T) {
	traits := AvatarTraits{
		ShoulderScale: 1.10,
		HeadScale:     0.95,
		HeightScale:   1.05,
	}

	baseSpec := PartSpec{
		RelativeWidth:  0.50,
		RelativeHeight: 0.35,
	}

	// Head should scale both dimensions
	headSpec := traits.ApplyProportions(baseSpec, PartHead)
	if math.Abs(headSpec.RelativeWidth-0.50*0.95) > 0.001 {
		t.Errorf("head width: want %f, got %f", 0.50*0.95, headSpec.RelativeWidth)
	}
	if math.Abs(headSpec.RelativeHeight-0.35*0.95) > 0.001 {
		t.Errorf("head height: want %f, got %f", 0.35*0.95, headSpec.RelativeHeight)
	}

	// Torso should scale width only
	torsoSpec := traits.ApplyProportions(baseSpec, PartTorso)
	if math.Abs(torsoSpec.RelativeWidth-0.50*1.10) > 0.001 {
		t.Errorf("torso width: want %f, got %f", 0.50*1.10, torsoSpec.RelativeWidth)
	}
	if math.Abs(torsoSpec.RelativeHeight-0.35) > 0.001 {
		t.Errorf("torso height should be unchanged")
	}

	// Legs should scale height only
	legSpec := traits.ApplyProportions(baseSpec, PartLegs)
	if math.Abs(legSpec.RelativeHeight-0.35*1.05) > 0.001 {
		t.Errorf("leg height: want %f, got %f", 0.35*1.05, legSpec.RelativeHeight)
	}
}

func TestHSLToRGBA(t *testing.T) {
	tests := []struct {
		name    string
		h, s, l float64
		wantR   uint8
	}{
		{"red", 0, 1.0, 0.5, 255},
		{"black", 0, 0, 0, 0},
		{"white", 0, 0, 1.0, 255},
		{"green_ish", 120, 1.0, 0.5, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := hslToRGBA(tt.h, tt.s, tt.l)
			if c.R != tt.wantR {
				t.Errorf("R: want %d, got %d", tt.wantR, c.R)
			}
			if c.A != 255 {
				t.Error("alpha should always be 255")
			}
		})
	}
}

func TestDarkenRGBA(t *testing.T) {
	c := color.RGBA{R: 200, G: 100, B: 50, A: 255}
	d := darkenRGBA(c, 0.5)
	if d.R != 100 || d.G != 50 || d.B != 25 {
		t.Errorf("darken(0.5): want {100,50,25}, got {%d,%d,%d}", d.R, d.G, d.B)
	}
	if d.A != 255 {
		t.Error("alpha should be preserved")
	}
}

func TestIsHumanoidEntity(t *testing.T) {
	tests := []struct {
		entityType string
		want       bool
	}{
		{"player", true},
		{"npc", true},
		{"knight", true},
		{"mage", true},
		{"warrior", true},
		{"merchant", true},
		{"humanoid", true},
		{"spider", false},
		{"dragon", false},
		{"blob", false},
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.entityType, func(t *testing.T) {
			if got := IsHumanoidEntity(tt.entityType); got != tt.want {
				t.Errorf("IsHumanoidEntity(%q) = %v, want %v", tt.entityType, got, tt.want)
			}
		})
	}
}

func TestExtractAerialFlagDefault(t *testing.T) {
	// Default should now be true (aerial)
	config := Config{}
	if !extractAerialFlag(config) {
		t.Error("extractAerialFlag should return true by default")
	}

	// Explicit false overrides
	config.Custom = map[string]interface{}{"useAerial": false}
	if extractAerialFlag(config) {
		t.Error("explicit false should be respected")
	}

	// Explicit true works
	config.Custom = map[string]interface{}{"useAerial": true}
	if !extractAerialFlag(config) {
		t.Error("explicit true should be respected")
	}
}

func TestVaryColor(t *testing.T) {
	base := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	// Verify determinism
	rng1 := newTestRNG(42)
	rng2 := newTestRNG(42)
	c1 := varyColor(base, rng1, 10)
	c2 := varyColor(base, rng2, 10)
	if c1 != c2 {
		t.Errorf("varyColor not deterministic: %v vs %v", c1, c2)
	}
	// Verify alpha is preserved
	if c1.A != 255 {
		t.Error("alpha should be 255")
	}
}

func TestSelectEntityTemplate_AerialDefault(t *testing.T) {
	// With useAerial=true (now default), humanoids should get aerial templates
	tmpl := selectEntityTemplate("player", "fantasy", DirDown, false, false, true)
	if tmpl.Name == "" {
		t.Error("template should have a name")
	}
	// Non-humanoids should also get aerial templates
	tmpl = selectEntityTemplate("spider", "", DirDown, false, false, true)
	if tmpl.Name == "" {
		t.Error("non-humanoid template should have a name")
	}
}

func BenchmarkGenerateAvatarTraits(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GenerateAvatarTraits(int64(i))
	}
}

func BenchmarkGenerateCreatureTraits(b *testing.B) {
	types := []string{"spider", "serpent", "dragon", "blob", "undead"}
	for i := 0; i < b.N; i++ {
		GenerateCreatureTraits(int64(i), types[i%len(types)])
	}
}

func newTestRNG(seed int64) *rand.Rand {
	return rand.New(rand.NewSource(seed))
}
