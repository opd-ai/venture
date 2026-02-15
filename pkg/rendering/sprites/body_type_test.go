package sprites

import (
	"testing"
)

func TestBodyType_String(t *testing.T) {
	tests := []struct {
		bt   BodyType
		want string
	}{
		{BodyTypeAverage, "average"},
		{BodyTypeStocky, "stocky"},
		{BodyTypeLean, "lean"},
		{BodyTypeMuscular, "muscular"},
		{BodyTypeHeavy, "heavy"},
		{BodyTypePetite, "petite"},
		{BodyTypeBroad, "broad"},
		{BodyTypeLanky, "lanky"},
		{BodyType(99), "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.bt.String(); got != tt.want {
				t.Errorf("BodyType(%d).String() = %q, want %q", tt.bt, got, tt.want)
			}
		})
	}
}

func TestBodyTypeCount(t *testing.T) {
	if BodyTypeCount != 8 {
		t.Errorf("BodyTypeCount = %d, want 8", BodyTypeCount)
	}
}

func TestGetBodyTypeModifiers_AllTypes(t *testing.T) {
	for bt := BodyType(0); bt < BodyTypeCount; bt++ {
		t.Run(bt.String(), func(t *testing.T) {
			mods := GetBodyTypeModifiers(bt)

			// All scale factors must be positive
			if mods.TorsoWidthScale <= 0 || mods.TorsoHeightScale <= 0 {
				t.Error("torso scale must be positive")
			}
			if mods.HeadWidthScale <= 0 || mods.HeadHeightScale <= 0 {
				t.Error("head scale must be positive")
			}
			if mods.ArmWidthScale <= 0 || mods.ArmHeightScale <= 0 {
				t.Error("arm scale must be positive")
			}
			if mods.LegWidthScale <= 0 || mods.LegHeightScale <= 0 {
				t.Error("leg scale must be positive")
			}
			if mods.ShadowWidthScale <= 0 {
				t.Error("shadow width scale must be positive")
			}

			// Average should be all 1.0
			if bt == BodyTypeAverage {
				if mods.TorsoWidthScale != 1.0 || mods.TorsoHeightScale != 1.0 {
					t.Error("average torso should be 1.0")
				}
				if mods.HeadWidthScale != 1.0 || mods.HeadHeightScale != 1.0 {
					t.Error("average head should be 1.0")
				}
			}
		})
	}
}

func TestGetBodyTypeModifiers_VisibleDifferences(t *testing.T) {
	avg := GetBodyTypeModifiers(BodyTypeAverage)
	minDiff := 0.10 // At least 10% difference from average to be visible

	for bt := BodyType(1); bt < BodyTypeCount; bt++ {
		t.Run(bt.String(), func(t *testing.T) {
			mods := GetBodyTypeModifiers(bt)
			// At least one significant body part must differ by >= minDiff
			torsoDiff := abs64(mods.TorsoWidthScale - avg.TorsoWidthScale)
			armDiff := abs64(mods.ArmWidthScale - avg.ArmWidthScale)
			legDiff := abs64(mods.LegWidthScale - avg.LegWidthScale)

			maxDiff := torsoDiff
			if armDiff > maxDiff {
				maxDiff = armDiff
			}
			if legDiff > maxDiff {
				maxDiff = legDiff
			}

			if maxDiff < minDiff {
				t.Errorf("body type %s max width diff from average = %.3f, want >= %.3f for visible difference", bt, maxDiff, minDiff)
			}
		})
	}
}

func TestDeriveBodyType_Deterministic(t *testing.T) {
	seeds := []int64{0, 1, 42, 12345, -999, 999999}
	for _, seed := range seeds {
		bt1 := DeriveBodyType(seed, "fantasy")
		bt2 := DeriveBodyType(seed, "fantasy")
		if bt1 != bt2 {
			t.Errorf("DeriveBodyType(%d, fantasy) not deterministic: %d != %d", seed, bt1, bt2)
		}
	}
}

func TestDeriveBodyType_Variety(t *testing.T) {
	// With 100 different seeds, we should see at least 4 different body types
	seen := make(map[BodyType]bool)
	for seed := int64(0); seed < 100; seed++ {
		bt := DeriveBodyType(seed*7919, "fantasy") // prime multiplier for spread
		seen[bt] = true
	}
	if len(seen) < 4 {
		t.Errorf("only %d body types seen across 100 seeds, want >= 4", len(seen))
	}
}

func TestDeriveBodyType_GenreInfluence(t *testing.T) {
	genres := []string{"fantasy", "horror", "cyberpunk", "sci-fi", "post-apocalyptic", ""}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			counts := make(map[BodyType]int)
			for seed := int64(0); seed < 200; seed++ {
				bt := DeriveBodyType(seed*31, genre)
				counts[bt]++
				if bt < 0 || bt >= BodyTypeCount {
					t.Errorf("invalid body type %d", bt)
				}
			}
			// Should produce at least 3 types with 200 samples
			if len(counts) < 3 {
				t.Errorf("genre %q only produced %d body types, want >= 3", genre, len(counts))
			}
		})
	}
}

func TestApplyBodyTypeToTemplate_Average(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)
	result := ApplyBodyTypeToTemplate(base, BodyTypeAverage)

	// Average should return template unchanged (same name)
	if result.Name != base.Name {
		t.Errorf("average body type changed template name: %q -> %q", base.Name, result.Name)
	}
}

func TestApplyBodyTypeToTemplate_Stocky(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)
	result := ApplyBodyTypeToTemplate(base, BodyTypeStocky)

	// Stocky should have wider torso
	baseTorso := base.BodyPartLayout[PartTorso]
	resultTorso := result.BodyPartLayout[PartTorso]
	if resultTorso.RelativeWidth <= baseTorso.RelativeWidth {
		t.Errorf("stocky torso should be wider: base=%.3f result=%.3f",
			baseTorso.RelativeWidth, resultTorso.RelativeWidth)
	}

	// Name should be modified
	if result.Name == base.Name {
		t.Error("body type should modify template name")
	}
}

func TestApplyBodyTypeToTemplate_Lean(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)
	result := ApplyBodyTypeToTemplate(base, BodyTypeLean)

	baseTorso := base.BodyPartLayout[PartTorso]
	resultTorso := result.BodyPartLayout[PartTorso]
	if resultTorso.RelativeWidth >= baseTorso.RelativeWidth {
		t.Errorf("lean torso should be narrower: base=%.3f result=%.3f",
			baseTorso.RelativeWidth, resultTorso.RelativeWidth)
	}
}

func TestApplyBodyTypeToTemplate_AllTypes(t *testing.T) {
	base := EnhancedHumanoidAerialTemplate(DirDown)
	for bt := BodyType(0); bt < BodyTypeCount; bt++ {
		t.Run(bt.String(), func(t *testing.T) {
			result := ApplyBodyTypeToTemplate(base, bt)
			// Must have same body parts
			if len(result.BodyPartLayout) != len(base.BodyPartLayout) {
				t.Errorf("body part count changed: %d -> %d",
					len(base.BodyPartLayout), len(result.BodyPartLayout))
			}
			// All parts must have valid dimensions
			for part, spec := range result.BodyPartLayout {
				if spec.RelativeWidth <= 0 {
					t.Errorf("part %v has invalid width %.3f", part, spec.RelativeWidth)
				}
				if spec.RelativeHeight <= 0 {
					t.Errorf("part %v has invalid height %.3f", part, spec.RelativeHeight)
				}
			}
		})
	}
}

func TestApplyBodyTypeToTemplate_PreferredPixelSize(t *testing.T) {
	base := EnhancedHumanoidAerialTemplate(DirDown)
	result := ApplyBodyTypeToTemplate(base, BodyTypeHeavy)

	// Enhanced template has PreferredPixelSize; heavy should scale it up
	baseTorso := base.BodyPartLayout[PartTorso]
	resultTorso := result.BodyPartLayout[PartTorso]
	if baseTorso.PreferredPixelSize != nil && resultTorso.PreferredPixelSize != nil {
		if resultTorso.PreferredPixelSize.Width <= baseTorso.PreferredPixelSize.Width {
			t.Errorf("heavy torso pixel width should increase: base=%d result=%d",
				baseTorso.PreferredPixelSize.Width, resultTorso.PreferredPixelSize.Width)
		}
	}
}

func TestScalePixelDim(t *testing.T) {
	tests := []struct {
		name     string
		original int
		factor   float64
		want     int
	}{
		{"identity", 5, 1.0, 5},
		{"scale_up", 5, 1.4, 7},
		{"scale_down", 5, 0.6, 3},
		{"clamp_min", 1, 0.1, 1},
		{"round_up", 5, 1.09, 5},
		{"round_half", 5, 1.1, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scalePixelDim(tt.original, tt.factor)
			if got != tt.want {
				t.Errorf("scalePixelDim(%d, %.2f) = %d, want %d", tt.original, tt.factor, got, tt.want)
			}
		})
	}
}

func TestGenreBodyTypeWeights_Valid(t *testing.T) {
	genres := []string{"fantasy", "horror", "cyberpunk", "sci-fi", "post-apocalyptic", "unknown", ""}
	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			weights := genreBodyTypeWeights(genre)
			total := 0.0
			for i, w := range weights {
				if w <= 0 {
					t.Errorf("weight[%d] = %.2f, must be positive", i, w)
				}
				total += w
			}
			if total <= 0 {
				t.Error("total weight must be positive")
			}
		})
	}
}

func abs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func BenchmarkDeriveBodyType(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DeriveBodyType(int64(i), "fantasy")
	}
}

func BenchmarkApplyBodyTypeToTemplate(b *testing.B) {
	base := HumanoidAerialTemplate(DirDown)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ApplyBodyTypeToTemplate(base, BodyType(i%int(BodyTypeCount)))
	}
}
