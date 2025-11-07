// Package sprites provides equipment visual tests.
package sprites

import (
	"testing"
)

func TestGetMaterialTypeFromWeaponType(t *testing.T) {
	tests := []struct {
		name       string
		weaponType string
		genreID    string
		want       MaterialType
	}{
		{"sword is metal", "sword", "fantasy", MaterialMetal},
		{"axe is metal", "axe", "fantasy", MaterialMetal},
		{"bow is wood", "bow", "fantasy", MaterialWood},
		{"staff is wood", "staff", "fantasy", MaterialWood},
		{"wand is wood", "wand", "fantasy", MaterialWood},
		{"gun is metal", "gun", "sci-fi", MaterialMetal},
		{"dagger is metal", "dagger", "fantasy", MaterialMetal},
		{"unknown defaults to metal", "unknown", "fantasy", MaterialMetal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMaterialTypeFromWeaponType(tt.weaponType, tt.genreID)
			if got != tt.want {
				t.Errorf("GetMaterialTypeFromWeaponType(%q, %q) = %v, want %v", tt.weaponType, tt.genreID, got, tt.want)
			}
		})
	}
}

func TestGetMaterialTypeFromArmorType(t *testing.T) {
	tests := []struct {
		name      string
		armorType string
		genreID   string
		want      MaterialType
	}{
		{"helmet is metal", "helmet", "fantasy", MaterialMetal},
		{"chest is metal", "chest", "fantasy", MaterialMetal},
		{"legs is leather", "legs", "fantasy", MaterialLeather},
		{"boots is metal", "boots", "fantasy", MaterialMetal},
		{"gloves is metal", "gloves", "fantasy", MaterialMetal},
		{"shield is metal", "shield", "fantasy", MaterialMetal},
		{"unknown defaults to metal", "unknown", "fantasy", MaterialMetal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMaterialTypeFromArmorType(tt.armorType, tt.genreID)
			if got != tt.want {
				t.Errorf("GetMaterialTypeFromArmorType(%q, %q) = %v, want %v", tt.armorType, tt.genreID, got, tt.want)
			}
		})
	}
}

func TestGetMaterialTypeFromTags(t *testing.T) {
	tests := []struct {
		name    string
		tags    []string
		genreID string
		want    MaterialType
	}{
		{"metal tag", []string{"metal", "heavy"}, "fantasy", MaterialMetal},
		{"leather tag", []string{"leather", "light"}, "fantasy", MaterialLeather},
		{"cloth tag", []string{"cloth", "magical"}, "fantasy", MaterialCloth},
		{"wood tag", []string{"wooden", "ranged"}, "fantasy", MaterialWood},
		{"crystal tag", []string{"crystal", "magical"}, "fantasy", MaterialCrystal},
		{"energy tag", []string{"energy", "arcane"}, "sci-fi", MaterialEnergy},
		{"fantasy default", []string{"heavy"}, "fantasy", MaterialMetal},
		{"sci-fi default", []string{"advanced"}, "sci-fi", MaterialEnergy},
		{"horror default", []string{"rusty"}, "horror", MaterialWood},
		{"cyberpunk default", []string{"cyber"}, "cyberpunk", MaterialMetal},
		{"post-apocalyptic default", []string{"scavenged"}, "post-apocalyptic", MaterialLeather},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMaterialTypeFromTags(tt.tags, tt.genreID)
			if got != tt.want {
				t.Errorf("GetMaterialTypeFromTags(%v, %q) = %v, want %v", tt.tags, tt.genreID, got, tt.want)
			}
		})
	}
}

func TestGetDamageStateFromDurability(t *testing.T) {
	tests := []struct {
		name    string
		current int
		max     int
		want    DamageState
	}{
		{"pristine 100%", 100, 100, DamageStatePristine},
		{"pristine 99%", 99, 100, DamageStateWorn},
		{"worn 75%", 75, 100, DamageStateWorn},
		{"worn 50%", 50, 100, DamageStateWorn},
		{"damaged 49%", 49, 100, DamageStateDamaged},
		{"damaged 25%", 25, 100, DamageStateDamaged},
		{"broken 24%", 24, 100, DamageStateBroken},
		{"broken 0%", 0, 100, DamageStateBroken},
		{"zero max pristine", 0, 0, DamageStatePristine},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDamageStateFromDurability(tt.current, tt.max)
			if got != tt.want {
				t.Errorf("GetDamageStateFromDurability(%d, %d) = %v, want %v", tt.current, tt.max, got, tt.want)
			}
		})
	}
}

func TestGetEnchantmentFromRarity(t *testing.T) {
	tests := []struct {
		name          string
		rarity        string
		wantActive    bool
		wantColor     string
		wantIntensity float64
		wantParticles int
	}{
		{"common no enchantment", "common", false, "white", 0.0, 0},
		{"uncommon green", "uncommon", true, "green", 0.2, 2},
		{"rare blue", "rare", true, "blue", 0.4, 4},
		{"epic purple", "epic", true, "purple", 0.6, 8},
		{"legendary gold", "legendary", true, "gold", 0.8, 12},
		{"unknown no enchantment", "unknown", false, "white", 0.0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetEnchantmentFromRarity(tt.rarity)
			if got.Active != tt.wantActive {
				t.Errorf("Active = %v, want %v", got.Active, tt.wantActive)
			}
			if got.Color != tt.wantColor {
				t.Errorf("Color = %v, want %v", got.Color, tt.wantColor)
			}
			if got.Intensity != tt.wantIntensity {
				t.Errorf("Intensity = %v, want %v", got.Intensity, tt.wantIntensity)
			}
			if got.ParticleCount != tt.wantParticles {
				t.Errorf("ParticleCount = %v, want %v", got.ParticleCount, tt.wantParticles)
			}
			if tt.wantActive && got.PulseSpeed <= 0 {
				t.Errorf("PulseSpeed should be > 0 for active enchantment, got %v", got.PulseSpeed)
			}
		})
	}
}

func TestGetDetailLevelFromRarity(t *testing.T) {
	tests := []struct {
		name   string
		rarity string
		want   float64
	}{
		{"common low detail", "common", 0.3},
		{"uncommon medium-low", "uncommon", 0.4},
		{"rare medium", "rare", 0.6},
		{"epic high", "epic", 0.8},
		{"legendary max", "legendary", 1.0},
		{"unknown default", "unknown", 0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDetailLevelFromRarity(tt.rarity)
			if got != tt.want {
				t.Errorf("GetDetailLevelFromRarity(%q) = %v, want %v", tt.rarity, got, tt.want)
			}
		})
	}
}

func TestGetMaterialVisualProperties(t *testing.T) {
	tests := []struct {
		name     string
		material MaterialType
		wantType string
	}{
		{"metal properties", MaterialMetal, "grain"},
		{"leather properties", MaterialLeather, "grain"},
		{"cloth properties", MaterialCloth, "weave"},
		{"wood properties", MaterialWood, "grain"},
		{"crystal properties", MaterialCrystal, "dots"},
		{"energy properties", MaterialEnergy, "none"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMaterialVisualProperties(tt.material)
			if got.PatternType != tt.wantType {
				t.Errorf("PatternType = %v, want %v", got.PatternType, tt.wantType)
			}
			if got.Sheen < 0.0 || got.Sheen > 1.0 {
				t.Errorf("Sheen out of range: %v", got.Sheen)
			}
			if got.Roughness < 0.0 || got.Roughness > 1.0 {
				t.Errorf("Roughness out of range: %v", got.Roughness)
			}
			if got.Reflectivity < 0.0 || got.Reflectivity > 1.0 {
				t.Errorf("Reflectivity out of range: %v", got.Reflectivity)
			}
		})
	}
}

func TestGetDamageVisualEffects(t *testing.T) {
	tests := []struct {
		name             string
		state            DamageState
		wantOpacity      float64
		wantCrackDensity float64
	}{
		{"pristine no damage", DamageStatePristine, 1.0, 0.0},
		{"worn light damage", DamageStateWorn, 0.95, 0.1},
		{"damaged moderate", DamageStateDamaged, 0.85, 0.4},
		{"broken heavy", DamageStateBroken, 0.7, 0.7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetDamageVisualEffects(tt.state)
			if got.OpacityMultiplier != tt.wantOpacity {
				t.Errorf("OpacityMultiplier = %v, want %v", got.OpacityMultiplier, tt.wantOpacity)
			}
			if got.CrackDensity != tt.wantCrackDensity {
				t.Errorf("CrackDensity = %v, want %v", got.CrackDensity, tt.wantCrackDensity)
			}
			if got.ColorDarken < 0.0 || got.ColorDarken > 1.0 {
				t.Errorf("ColorDarken out of range: %v", got.ColorDarken)
			}
			if got.Dirtiness < 0.0 || got.Dirtiness > 1.0 {
				t.Errorf("Dirtiness out of range: %v", got.Dirtiness)
			}
		})
	}
}

func TestMaterialTypeString(t *testing.T) {
	tests := []struct {
		material MaterialType
		want     string
	}{
		{MaterialMetal, "metal"},
		{MaterialLeather, "leather"},
		{MaterialCloth, "cloth"},
		{MaterialWood, "wood"},
		{MaterialCrystal, "crystal"},
		{MaterialEnergy, "energy"},
		{MaterialType(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.material.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDamageStateString(t *testing.T) {
	tests := []struct {
		state DamageState
		want  string
	}{
		{DamageStatePristine, "pristine"},
		{DamageStateWorn, "worn"},
		{DamageStateDamaged, "damaged"},
		{DamageStateBroken, "broken"},
		{DamageState(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Benchmark tests for performance verification
func BenchmarkGetMaterialTypeFromTags(b *testing.B) {
	tags := []string{"metal", "heavy", "powerful"}
	for i := 0; i < b.N; i++ {
		GetMaterialTypeFromTags(tags, "fantasy")
	}
}

func BenchmarkGetDamageStateFromDurability(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetDamageStateFromDurability(50, 100)
	}
}

func BenchmarkGetEnchantmentFromRarity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetEnchantmentFromRarity("legendary")
	}
}

func BenchmarkGetMaterialVisualProperties(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetMaterialVisualProperties(MaterialMetal)
	}
}

func BenchmarkGetDamageVisualEffects(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetDamageVisualEffects(DamageStateDamaged)
	}
}
