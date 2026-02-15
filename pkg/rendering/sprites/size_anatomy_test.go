package sprites

import (
	"math"
	"testing"
)

func TestParseSizeClass(t *testing.T) {
	tests := []struct {
		input string
		want  SizeClass
	}{
		{"tiny", SizeClassTiny},
		{"small", SizeClassSmall},
		{"medium", SizeClassMedium},
		{"large", SizeClassLarge},
		{"huge", SizeClassHuge},
		{"", SizeClassMedium},
		{"unknown", SizeClassMedium},
		{"TINY", SizeClassMedium}, // case-sensitive
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseSizeClass(tt.input)
			if got != tt.want {
				t.Errorf("ParseSizeClass(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestSizeClassString(t *testing.T) {
	tests := []struct {
		sc   SizeClass
		want string
	}{
		{SizeClassTiny, "tiny"},
		{SizeClassSmall, "small"},
		{SizeClassMedium, "medium"},
		{SizeClassLarge, "large"},
		{SizeClassHuge, "huge"},
		{SizeClass(99), "medium"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.sc.String(); got != tt.want {
				t.Errorf("SizeClass(%d).String() = %q, want %q", tt.sc, got, tt.want)
			}
		})
	}
}

func TestGetSizeScaleModifiers_AllClasses(t *testing.T) {
	classes := []SizeClass{SizeClassTiny, SizeClassSmall, SizeClassMedium, SizeClassLarge, SizeClassHuge}

	for _, sc := range classes {
		t.Run(sc.String(), func(t *testing.T) {
			mods := GetSizeScaleModifiers(sc)

			// All scales must be positive
			scales := []float64{
				mods.HeadWidthScale, mods.HeadHeightScale,
				mods.TorsoWidthScale, mods.TorsoHeightScale,
				mods.ArmWidthScale, mods.ArmHeightScale,
				mods.LegWidthScale, mods.LegHeightScale,
				mods.ShadowWidthScale, mods.ShadowOpacityScale,
			}
			for i, s := range scales {
				if s <= 0 {
					t.Errorf("scale[%d] = %f, must be positive", i, s)
				}
			}
		})
	}
}

func TestGetSizeScaleModifiers_HeadToTorsoRatio(t *testing.T) {
	// Key invariant: smaller creatures should have larger head-to-torso ratios
	// and larger creatures should have smaller ratios.
	tiny := GetSizeScaleModifiers(SizeClassTiny)
	medium := GetSizeScaleModifiers(SizeClassMedium)
	huge := GetSizeScaleModifiers(SizeClassHuge)

	tinyRatio := tiny.HeadWidthScale / tiny.TorsoWidthScale
	mediumRatio := medium.HeadWidthScale / medium.TorsoWidthScale
	hugeRatio := huge.HeadWidthScale / huge.TorsoWidthScale

	if tinyRatio <= mediumRatio {
		t.Errorf("tiny head/torso ratio (%f) should exceed medium (%f)", tinyRatio, mediumRatio)
	}
	if mediumRatio <= hugeRatio {
		t.Errorf("medium head/torso ratio (%f) should exceed huge (%f)", mediumRatio, hugeRatio)
	}
}

func TestGetSizeScaleModifiers_MediumIsBaseline(t *testing.T) {
	mods := GetSizeScaleModifiers(SizeClassMedium)
	if mods.HeadWidthScale != 1.0 || mods.TorsoWidthScale != 1.0 || mods.ArmWidthScale != 1.0 {
		t.Error("medium size should have 1.0x scales")
	}
}

func TestApplySizeScaling_MediumUnchanged(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)
	scaled := ApplySizeScaling(base, "medium")

	// Medium should return the template unmodified
	for part, spec := range base.BodyPartLayout {
		scaledSpec, ok := scaled.BodyPartLayout[part]
		if !ok {
			t.Errorf("missing part %v in scaled template", part)
			continue
		}
		if spec.RelativeWidth != scaledSpec.RelativeWidth {
			t.Errorf("part %v width changed for medium: %f → %f", part, spec.RelativeWidth, scaledSpec.RelativeWidth)
		}
		if spec.RelativeHeight != scaledSpec.RelativeHeight {
			t.Errorf("part %v height changed for medium: %f → %f", part, spec.RelativeHeight, scaledSpec.RelativeHeight)
		}
	}
}

func TestApplySizeScaling_EmptyStringIsMedium(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)
	scaled := ApplySizeScaling(base, "")

	if scaled.Name != base.Name {
		t.Errorf("empty string should return unmodified template, got name %q", scaled.Name)
	}
}

func TestApplySizeScaling_TinyEnlargesHead(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)
	scaled := ApplySizeScaling(base, "tiny")

	baseHead := base.BodyPartLayout[PartHead]
	scaledHead := scaled.BodyPartLayout[PartHead]

	if scaledHead.RelativeWidth <= baseHead.RelativeWidth {
		t.Errorf("tiny head width %f should exceed base %f", scaledHead.RelativeWidth, baseHead.RelativeWidth)
	}
	if scaledHead.RelativeHeight <= baseHead.RelativeHeight {
		t.Errorf("tiny head height %f should exceed base %f", scaledHead.RelativeHeight, baseHead.RelativeHeight)
	}
}

func TestApplySizeScaling_TinyShrinksTorso(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)
	scaled := ApplySizeScaling(base, "tiny")

	baseTorso := base.BodyPartLayout[PartTorso]
	scaledTorso := scaled.BodyPartLayout[PartTorso]

	if scaledTorso.RelativeWidth >= baseTorso.RelativeWidth {
		t.Errorf("tiny torso width %f should be less than base %f", scaledTorso.RelativeWidth, baseTorso.RelativeWidth)
	}
}

func TestApplySizeScaling_HugeShrinkHead(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)
	scaled := ApplySizeScaling(base, "huge")

	baseHead := base.BodyPartLayout[PartHead]
	scaledHead := scaled.BodyPartLayout[PartHead]

	if scaledHead.RelativeWidth >= baseHead.RelativeWidth {
		t.Errorf("huge head width %f should be less than base %f", scaledHead.RelativeWidth, baseHead.RelativeWidth)
	}
}

func TestApplySizeScaling_HugeEnlargesTorso(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)
	scaled := ApplySizeScaling(base, "huge")

	baseTorso := base.BodyPartLayout[PartTorso]
	scaledTorso := scaled.BodyPartLayout[PartTorso]

	if scaledTorso.RelativeWidth <= baseTorso.RelativeWidth {
		t.Errorf("huge torso width %f should exceed base %f", scaledTorso.RelativeWidth, baseTorso.RelativeWidth)
	}
}

func TestApplySizeScaling_PreservesZIndex(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)

	for _, size := range []string{"tiny", "small", "large", "huge"} {
		t.Run(size, func(t *testing.T) {
			scaled := ApplySizeScaling(base, size)
			for part, baseSpec := range base.BodyPartLayout {
				scaledSpec := scaled.BodyPartLayout[part]
				if scaledSpec.ZIndex != baseSpec.ZIndex {
					t.Errorf("part %v ZIndex changed from %d to %d", part, baseSpec.ZIndex, scaledSpec.ZIndex)
				}
			}
		})
	}
}

func TestApplySizeScaling_PreservesPartCount(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)

	for _, size := range []string{"tiny", "small", "large", "huge"} {
		t.Run(size, func(t *testing.T) {
			scaled := ApplySizeScaling(base, size)
			if len(scaled.BodyPartLayout) != len(base.BodyPartLayout) {
				t.Errorf("part count changed: %d → %d", len(base.BodyPartLayout), len(scaled.BodyPartLayout))
			}
		})
	}
}

func TestApplySizeScaling_ShadowOpacityCapped(t *testing.T) {
	// Create template with high shadow opacity to test capping
	base := HumanoidAerialTemplate(DirDown)
	shadow := base.BodyPartLayout[PartShadow]
	shadow.Opacity = 0.50
	base.BodyPartLayout[PartShadow] = shadow

	scaled := ApplySizeScaling(base, "huge")
	scaledShadow := scaled.BodyPartLayout[PartShadow]

	if scaledShadow.Opacity > 0.55 {
		t.Errorf("shadow opacity %f should be capped at 0.55", scaledShadow.Opacity)
	}
}

func TestApplySizeScaling_PreferredPixelSize(t *testing.T) {
	base := EnhancedHumanoidAerialTemplate(DirDown)

	scaled := ApplySizeScaling(base, "huge")
	baseHead := base.BodyPartLayout[PartHead]
	scaledHead := scaled.BodyPartLayout[PartHead]

	if baseHead.PreferredPixelSize != nil && scaledHead.PreferredPixelSize != nil {
		// Huge should shrink head pixels
		if scaledHead.PreferredPixelSize.Width >= baseHead.PreferredPixelSize.Width {
			// Width might stay the same due to rounding at small pixel counts — acceptable
			if scaledHead.PreferredPixelSize.Width > baseHead.PreferredPixelSize.Width {
				t.Errorf("huge head pixel width %d should not exceed base %d",
					scaledHead.PreferredPixelSize.Width, baseHead.PreferredPixelSize.Width)
			}
		}
	}
}

func TestApplySizeScaling_NonhumanoidTemplate(t *testing.T) {
	base := QuadrupedAerialTemplate(DirDown)
	scaled := ApplySizeScaling(base, "large")

	// Should work on nonhumanoid templates too
	if len(scaled.BodyPartLayout) == 0 {
		t.Error("size scaling should produce non-empty template for nonhumanoid")
	}

	// Name should reflect size
	if scaled.Name == base.Name {
		t.Error("scaled template name should differ from base")
	}
}

func TestApplySizeScaling_AllDirections(t *testing.T) {
	dirs := []Direction{DirUp, DirDown, DirLeft, DirRight}

	for _, dir := range dirs {
		for _, size := range []string{"tiny", "huge"} {
			t.Run(string(dir)+"_"+size, func(t *testing.T) {
				base := HumanoidAerialTemplate(dir)
				scaled := ApplySizeScaling(base, size)
				if len(scaled.BodyPartLayout) != len(base.BodyPartLayout) {
					t.Errorf("part count mismatch for %s %s", dir, size)
				}
			})
		}
	}
}

func TestApplySizeScaling_MonotonicProgression(t *testing.T) {
	// Head should shrink monotonically from Tiny → Huge
	// Torso should grow monotonically from Tiny → Huge
	sizes := []string{"tiny", "small", "medium", "large", "huge"}
	base := HumanoidAerialTemplate(DirDown)

	var prevHeadW, prevTorsoW float64
	for i, size := range sizes {
		scaled := ApplySizeScaling(base, size)
		headW := scaled.BodyPartLayout[PartHead].RelativeWidth
		torsoW := scaled.BodyPartLayout[PartTorso].RelativeWidth

		if i > 0 {
			if headW > prevHeadW+0.001 {
				t.Errorf("head width should not increase from %s to %s: %f → %f",
					sizes[i-1], size, prevHeadW, headW)
			}
			if torsoW < prevTorsoW-0.001 {
				t.Errorf("torso width should not decrease from %s to %s: %f → %f",
					sizes[i-1], size, prevTorsoW, torsoW)
			}
		}
		prevHeadW = headW
		prevTorsoW = torsoW
	}
}

func TestApplySizeScaling_TinyShapePreference(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)
	scaled := ApplySizeScaling(base, "tiny")

	head := scaled.BodyPartLayout[PartHead]
	// Tiny creatures should prefer circle heads (rounder = cuter)
	found := false
	for _, st := range head.ShapeTypes {
		if st == 0 { // ShapeCircle = 0
			found = true
			break
		}
	}
	if !found {
		t.Error("tiny size should prefer circle head shapes")
	}
}

func BenchmarkApplySizeScaling(b *testing.B) {
	base := HumanoidAerialTemplate(DirDown)
	sizes := []string{"tiny", "small", "medium", "large", "huge"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplySizeScaling(base, sizes[i%len(sizes)])
	}
}

func TestApplySizeScaling_DistinctSilhouettes(t *testing.T) {
	// Key test: different sizes must produce meaningfully different proportions.
	base := HumanoidAerialTemplate(DirDown)
	tiny := ApplySizeScaling(base, "tiny")
	huge := ApplySizeScaling(base, "huge")

	tinyHead := tiny.BodyPartLayout[PartHead]
	hugeHead := huge.BodyPartLayout[PartHead]
	tinyTorso := tiny.BodyPartLayout[PartTorso]
	hugeTorso := huge.BodyPartLayout[PartTorso]

	// Head width ratio between tiny and huge should differ by at least 50%
	headRatio := tinyHead.RelativeWidth / hugeHead.RelativeWidth
	if headRatio < 1.5 {
		t.Errorf("tiny/huge head width ratio %f should be ≥1.5 for distinct silhouettes", headRatio)
	}

	// Torso width ratio should also be dramatic
	torsoRatio := hugeTorso.RelativeWidth / tinyTorso.RelativeWidth
	if torsoRatio < 1.5 {
		t.Errorf("huge/tiny torso width ratio %f should be ≥1.5 for distinct silhouettes", torsoRatio)
	}

	// Total area difference should be significant
	tinyArea := tinyHead.RelativeWidth*tinyHead.RelativeHeight + tinyTorso.RelativeWidth*tinyTorso.RelativeHeight
	hugeArea := hugeHead.RelativeWidth*hugeHead.RelativeHeight + hugeTorso.RelativeWidth*hugeTorso.RelativeHeight
	areaDiff := math.Abs(hugeArea-tinyArea) / tinyArea
	if areaDiff < 0.2 {
		t.Errorf("total area difference %f should be ≥0.2 between tiny and huge", areaDiff)
	}
}
