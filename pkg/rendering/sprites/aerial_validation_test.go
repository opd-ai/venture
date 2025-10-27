package sprites

import (
	"math"
	"testing"
)

// TestAerialTemplate_ProportionConsistency verifies that all aerial templates
// maintain the standard 35/50/15 (head/torso/legs) proportion ratios.
// This ensures visual consistency across genres and directions.
func TestAerialTemplate_ProportionConsistency(t *testing.T) {
	const tolerance = 0.05 // 5% tolerance for minor variations

	tests := []struct {
		name      string
		template  func(Direction) AnatomicalTemplate
		direction Direction
	}{
		// Base template
		{"base_up", HumanoidAerialTemplate, DirUp},
		{"base_down", HumanoidAerialTemplate, DirDown},
		{"base_left", HumanoidAerialTemplate, DirLeft},
		{"base_right", HumanoidAerialTemplate, DirRight},

		// Genre templates
		{"fantasy_up", FantasyHumanoidAerial, DirUp},
		{"fantasy_down", FantasyHumanoidAerial, DirDown},
		{"scifi_up", SciFiHumanoidAerial, DirUp},
		{"scifi_down", SciFiHumanoidAerial, DirDown},
		{"horror_up", HorrorHumanoidAerial, DirUp},
		{"horror_down", HorrorHumanoidAerial, DirDown},
		{"cyberpunk_up", CyberpunkHumanoidAerial, DirUp},
		{"cyberpunk_down", CyberpunkHumanoidAerial, DirDown},
		{"postapoc_up", PostApocHumanoidAerial, DirUp},
		{"postapoc_down", PostApocHumanoidAerial, DirDown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := tt.template(tt.direction)

			// Extract body part specs
			head, hasHead := template.BodyPartLayout[PartHead]
			torso, hasTorso := template.BodyPartLayout[PartTorso]
			legs, hasLegs := template.BodyPartLayout[PartLegs]

			if !hasHead || !hasTorso || !hasLegs {
				t.Fatal("Template missing required body parts")
			}

			// Check head proportion (should be ~0.35 ± tolerance)
			expectedHead := 0.35
			if math.Abs(head.RelativeHeight-expectedHead) > tolerance {
				t.Errorf("Head height %.2f outside range %.2f ± %.2f",
					head.RelativeHeight, expectedHead, tolerance)
			}

			// Check torso proportion (should be ~0.50 ± tolerance)
			expectedTorso := 0.50
			if math.Abs(torso.RelativeHeight-expectedTorso) > tolerance {
				t.Errorf("Torso height %.2f outside range %.2f ± %.2f",
					torso.RelativeHeight, expectedTorso, tolerance)
			}

			// Check legs proportion (should be ~0.15 ± tolerance)
			expectedLegs := 0.15
			if math.Abs(legs.RelativeHeight-expectedLegs) > tolerance {
				t.Errorf("Legs height %.2f outside range %.2f ± %.2f",
					legs.RelativeHeight, expectedLegs, tolerance)
			}

			// Verify proportions add up to approximately 1.0 (100%)
			total := head.RelativeHeight + torso.RelativeHeight + legs.RelativeHeight
			if math.Abs(total-1.0) > tolerance {
				t.Errorf("Total proportions %.2f != 1.0 (±%.2f)", total, tolerance)
			}
		})
	}
}

// TestAerialTemplate_ShadowConsistency verifies that shadow dimensions
// are appropriate for the body size across all aerial templates.
func TestAerialTemplate_ShadowConsistency(t *testing.T) {
	templates := []struct {
		name     string
		template func(Direction) AnatomicalTemplate
	}{
		{"base", HumanoidAerialTemplate},
		{"fantasy", FantasyHumanoidAerial},
		{"scifi", SciFiHumanoidAerial},
		{"horror", HorrorHumanoidAerial},
		{"cyberpunk", CyberpunkHumanoidAerial},
		{"postapoc", PostApocHumanoidAerial},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			template := tt.template(DirDown)

			shadow, hasShadow := template.BodyPartLayout[PartShadow]
			if !hasShadow {
				t.Fatal("Template missing shadow")
			}

			// Shadow should be at bottom of sprite (Y >= 0.85)
			if shadow.RelativeY < 0.85 {
				t.Errorf("Shadow Y position %.2f too high (expected >= 0.85)", shadow.RelativeY)
			}

			// Shadow width should be reasonable relative to torso
			torso := template.BodyPartLayout[PartTorso]
			if shadow.RelativeWidth > torso.RelativeWidth*1.2 {
				t.Errorf("Shadow width %.2f too large relative to torso %.2f",
					shadow.RelativeWidth, torso.RelativeWidth)
			}

			// Shadow height should be compressed (ellipse)
			if shadow.RelativeHeight > 0.20 {
				t.Errorf("Shadow height %.2f too tall (expected <= 0.20)", shadow.RelativeHeight)
			}

			// Shadow opacity should be semi-transparent (0.2 - 0.4)
			if shadow.Opacity < 0.15 || shadow.Opacity > 0.45 {
				t.Errorf("Shadow opacity %.2f outside valid range [0.15, 0.45]", shadow.Opacity)
			}
		})
	}
}

// TestAerialTemplate_ColorCoherence verifies that all body parts use
// consistent color role assignments across templates.
func TestAerialTemplate_ColorCoherence(t *testing.T) {
	validRoles := map[string]bool{
		"primary":   true,
		"secondary": true,
		"accent1":   true,
		"accent2":   true,
		"accent3":   true,
		"shadow":    true,
	}

	templates := []struct {
		name     string
		template func(Direction) AnatomicalTemplate
	}{
		{"base", HumanoidAerialTemplate},
		{"fantasy", FantasyHumanoidAerial},
		{"scifi", SciFiHumanoidAerial},
		{"horror", HorrorHumanoidAerial},
		{"cyberpunk", CyberpunkHumanoidAerial},
		{"postapoc", PostApocHumanoidAerial},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			template := tt.template(DirDown)

			for part, spec := range template.BodyPartLayout {
				if !validRoles[spec.ColorRole] {
					t.Errorf("Part %s has invalid color role: %s", part, spec.ColorRole)
				}
			}

			// Check standard color assignments
			if head, ok := template.BodyPartLayout[PartHead]; ok {
				// Head should typically use secondary (skin/clothing)
				// Exception: Cyberpunk uses accent1 for tech glow
				if tt.name != "cyberpunk" && head.ColorRole != "secondary" {
					t.Errorf("Head uses color role %s, expected 'secondary' (or 'accent1' for cyberpunk)",
						head.ColorRole)
				}
			}

			if torso, ok := template.BodyPartLayout[PartTorso]; ok {
				if torso.ColorRole != "primary" {
					t.Errorf("Torso uses color role %s, expected 'primary'", torso.ColorRole)
				}
			}

			if arms, ok := template.BodyPartLayout[PartArms]; ok {
				if arms.ColorRole != "secondary" {
					t.Errorf("Arms use color role %s, expected 'secondary'", arms.ColorRole)
				}
			}

			if legs, ok := template.BodyPartLayout[PartLegs]; ok {
				if legs.ColorRole != "primary" {
					t.Errorf("Legs use color role %s, expected 'primary'", legs.ColorRole)
				}
			}

			if shadow, ok := template.BodyPartLayout[PartShadow]; ok {
				if shadow.ColorRole != "shadow" {
					t.Errorf("Shadow uses color role %s, expected 'shadow'", shadow.ColorRole)
				}
			}
		})
	}
}

// TestAerialTemplate_DirectionalAsymmetry verifies that different directions
// produce visually distinct templates (head offset, arm positioning).
func TestAerialTemplate_DirectionalAsymmetry(t *testing.T) {
	templates := []struct {
		name     string
		template func(Direction) AnatomicalTemplate
	}{
		{"base", HumanoidAerialTemplate},
		{"fantasy", FantasyHumanoidAerial},
		{"scifi", SciFiHumanoidAerial},
		{"horror", HorrorHumanoidAerial},
		{"cyberpunk", CyberpunkHumanoidAerial},
		{"postapoc", PostApocHumanoidAerial},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			up := tt.template(DirUp)
			down := tt.template(DirDown)
			left := tt.template(DirLeft)
			right := tt.template(DirRight)

			// Check head X offset for left/right
			headLeft := left.BodyPartLayout[PartHead]
			headRight := right.BodyPartLayout[PartHead]

			if headLeft.RelativeX == headRight.RelativeX {
				t.Error("Left and right directions have same head X position (no asymmetry)")
			}

			// Left should be < 0.5, right should be > 0.5
			if headLeft.RelativeX >= 0.5 {
				t.Errorf("Left head X %.2f should be < 0.5", headLeft.RelativeX)
			}
			if headRight.RelativeX <= 0.5 {
				t.Errorf("Right head X %.2f should be > 0.5", headRight.RelativeX)
			}

			// Check arm Z-index differences between up/down
			armsUp := up.BodyPartLayout[PartArms]
			armsDown := down.BodyPartLayout[PartArms]

			if armsUp.ZIndex == armsDown.ZIndex {
				t.Error("Up and down directions have same arm Z-index (no depth differentiation)")
			}

			// Arms should be behind torso (Z < 10) when facing up
			if armsUp.ZIndex >= 10 {
				t.Errorf("Arms facing up should be behind torso (Z=%d >= 10)", armsUp.ZIndex)
			}

			// Arms should be in front of torso (Z > 10) when facing down
			if armsDown.ZIndex <= 10 {
				t.Errorf("Arms facing down should be in front of torso (Z=%d <= 10)", armsDown.ZIndex)
			}
		})
	}
}

// TestAerialTemplate_ZIndexOrdering verifies that body parts have logical
// layering (shadow < legs < torso, etc.) for proper depth rendering.
func TestAerialTemplate_ZIndexOrdering(t *testing.T) {
	templates := []struct {
		name     string
		template func(Direction) AnatomicalTemplate
	}{
		{"base", HumanoidAerialTemplate},
		{"fantasy", FantasyHumanoidAerial},
		{"scifi", SciFiHumanoidAerial},
		{"horror", HorrorHumanoidAerial},
		{"cyberpunk", CyberpunkHumanoidAerial},
		{"postapoc", PostApocHumanoidAerial},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			template := tt.template(DirDown)

			shadow := template.BodyPartLayout[PartShadow]
			legs := template.BodyPartLayout[PartLegs]
			torso := template.BodyPartLayout[PartTorso]
			head := template.BodyPartLayout[PartHead]

			// Shadow should be at bottom layer
			if shadow.ZIndex != 0 {
				t.Errorf("Shadow Z-index %d should be 0 (bottom layer)", shadow.ZIndex)
			}

			// Legs < Torso < Head
			if legs.ZIndex >= torso.ZIndex {
				t.Errorf("Legs Z-index %d should be < torso Z-index %d",
					legs.ZIndex, torso.ZIndex)
			}

			if torso.ZIndex >= head.ZIndex {
				t.Errorf("Torso Z-index %d should be < head Z-index %d",
					torso.ZIndex, head.ZIndex)
			}

			// Head should be top layer (typically Z=15)
			if head.ZIndex < 14 {
				t.Errorf("Head Z-index %d too low (expected >= 14)", head.ZIndex)
			}
		})
	}
}

// TestAerialTemplate_GenreSpecificFeatures verifies that genre variants
// have their distinctive elements while maintaining base proportions.
func TestAerialTemplate_GenreSpecificFeatures(t *testing.T) {
	t.Run("fantasy_broader_shoulders", func(t *testing.T) {
		base := HumanoidAerialTemplate(DirDown)
		fantasy := FantasyHumanoidAerial(DirDown)

		baseTorso := base.BodyPartLayout[PartTorso]
		fantasyTorso := fantasy.BodyPartLayout[PartTorso]

		if fantasyTorso.RelativeWidth <= baseTorso.RelativeWidth {
			t.Errorf("Fantasy torso width %.2f should be > base width %.2f",
				fantasyTorso.RelativeWidth, baseTorso.RelativeWidth)
		}
	})

	t.Run("scifi_jetpack_when_facing_up", func(t *testing.T) {
		scifiUp := SciFiHumanoidAerial(DirUp)
		scifiDown := SciFiHumanoidAerial(DirDown)

		_, hasArmorUp := scifiUp.BodyPartLayout[PartArmor]
		_, hasArmorDown := scifiDown.BodyPartLayout[PartArmor]

		if !hasArmorUp {
			t.Error("Sci-fi template facing up should have jetpack (PartArmor)")
		}

		if hasArmorDown {
			t.Error("Sci-fi template facing down should not have jetpack visible")
		}
	})

	t.Run("horror_reduced_shadow", func(t *testing.T) {
		base := HumanoidAerialTemplate(DirDown)
		horror := HorrorHumanoidAerial(DirDown)

		baseShadow := base.BodyPartLayout[PartShadow]
		horrorShadow := horror.BodyPartLayout[PartShadow]

		if horrorShadow.Opacity >= baseShadow.Opacity {
			t.Errorf("Horror shadow opacity %.2f should be < base opacity %.2f",
				horrorShadow.Opacity, baseShadow.Opacity)
		}
	})

	t.Run("cyberpunk_neon_glow", func(t *testing.T) {
		cyberpunk := CyberpunkHumanoidAerial(DirDown)

		armor, hasArmor := cyberpunk.BodyPartLayout[PartArmor]
		if !hasArmor {
			t.Fatal("Cyberpunk template should have armor overlay for neon glow")
		}

		// Glow should be semi-transparent
		if armor.Opacity > 0.4 {
			t.Errorf("Cyberpunk glow opacity %.2f too high (expected <= 0.4)", armor.Opacity)
		}

		// Glow should use accent color
		if armor.ColorRole != "accent1" {
			t.Errorf("Cyberpunk glow uses %s, expected 'accent1'", armor.ColorRole)
		}
	})
}

// TestBossAerialTemplate_Scaling verifies that boss scaling correctly
// applies uniform scaling while preserving template structure.
func TestBossAerialTemplate_Scaling(t *testing.T) {
	tests := []struct {
		name  string
		scale float64
	}{
		{"boss_2.5x", 2.5},
		{"mini_boss_1.5x", 1.5},
		{"giant_boss_3.0x", 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := HumanoidAerialTemplate(DirDown)
			boss := BossAerialTemplate(base, tt.scale)

			// Verify all parts exist
			if len(boss.BodyPartLayout) != len(base.BodyPartLayout) {
				t.Fatalf("Boss has %d parts, base has %d", len(boss.BodyPartLayout), len(base.BodyPartLayout))
			}

			// Verify scaling was applied correctly to each part
			for part, baseSpec := range base.BodyPartLayout {
				bossSpec, exists := boss.BodyPartLayout[part]
				if !exists {
					t.Fatalf("Boss missing part: %s", part)
				}

				// Check width scaling
				expectedWidth := baseSpec.RelativeWidth * tt.scale
				if math.Abs(bossSpec.RelativeWidth-expectedWidth) > 0.001 {
					t.Errorf("Part %s width %.3f != expected %.3f",
						part, bossSpec.RelativeWidth, expectedWidth)
				}

				// Check height scaling
				expectedHeight := baseSpec.RelativeHeight * tt.scale
				if math.Abs(bossSpec.RelativeHeight-expectedHeight) > 0.001 {
					t.Errorf("Part %s height %.3f != expected %.3f",
						part, bossSpec.RelativeHeight, expectedHeight)
				}

				// Check position scaling (offsets from center)
				offsetX := baseSpec.RelativeX - 0.5
				offsetY := baseSpec.RelativeY - 0.5
				expectedX := 0.5 + (offsetX * tt.scale)
				expectedY := 0.5 + (offsetY * tt.scale)

				if math.Abs(bossSpec.RelativeX-expectedX) > 0.001 {
					t.Errorf("Part %s X position %.3f != expected %.3f",
						part, bossSpec.RelativeX, expectedX)
				}
				if math.Abs(bossSpec.RelativeY-expectedY) > 0.001 {
					t.Errorf("Part %s Y position %.3f != expected %.3f",
						part, bossSpec.RelativeY, expectedY)
				}

				// Verify non-scaled properties are preserved
				if bossSpec.ColorRole != baseSpec.ColorRole {
					t.Errorf("Part %s color role changed from %s to %s",
						part, baseSpec.ColorRole, bossSpec.ColorRole)
				}
				if bossSpec.ZIndex != baseSpec.ZIndex {
					t.Errorf("Part %s Z-index changed from %d to %d",
						part, baseSpec.ZIndex, bossSpec.ZIndex)
				}
				if bossSpec.Opacity != baseSpec.Opacity {
					t.Errorf("Part %s opacity changed from %.2f to %.2f",
						part, baseSpec.Opacity, bossSpec.Opacity)
				}
				if bossSpec.Rotation != baseSpec.Rotation {
					t.Errorf("Part %s rotation changed from %.0f to %.0f",
						part, baseSpec.Rotation, bossSpec.Rotation)
				}
			}
		})
	}
}

// TestBossAerialTemplate_ProportionPreservation verifies that boss scaling
// maintains the 35/50/15 proportion ratios.
func TestBossAerialTemplate_ProportionPreservation(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)
	boss := BossAerialTemplate(base, 2.5)

	const tolerance = 0.05

	head := boss.BodyPartLayout[PartHead]
	torso := boss.BodyPartLayout[PartTorso]
	legs := boss.BodyPartLayout[PartLegs]

	// Scaled heights
	scaledHeadHeight := head.RelativeHeight
	scaledTorsoHeight := torso.RelativeHeight
	scaledLegsHeight := legs.RelativeHeight

	// Ratios should remain 35:50:15 regardless of scaling
	// Compare ratios rather than absolute values
	baseHead := base.BodyPartLayout[PartHead]
	baseTorso := base.BodyPartLayout[PartTorso]
	baseLegs := base.BodyPartLayout[PartLegs]

	ratioHead := scaledHeadHeight / baseHead.RelativeHeight
	ratioTorso := scaledTorsoHeight / baseTorso.RelativeHeight
	ratioLegs := scaledLegsHeight / baseLegs.RelativeHeight

	// All ratios should be equal to the scale factor (2.5)
	expectedRatio := 2.5
	if math.Abs(ratioHead-expectedRatio) > tolerance {
		t.Errorf("Head scaling ratio %.2f != %.2f", ratioHead, expectedRatio)
	}
	if math.Abs(ratioTorso-expectedRatio) > tolerance {
		t.Errorf("Torso scaling ratio %.2f != %.2f", ratioTorso, expectedRatio)
	}
	if math.Abs(ratioLegs-expectedRatio) > tolerance {
		t.Errorf("Legs scaling ratio %.2f != %.2f", ratioLegs, expectedRatio)
	}
}

// TestBossAerialTemplate_DirectionalAsymmetry verifies that boss scaling
// preserves directional asymmetry (head offsets, arm positioning).
func TestBossAerialTemplate_DirectionalAsymmetry(t *testing.T) {
	directions := []struct {
		dir  Direction
		name string
	}{
		{DirUp, "up"},
		{DirDown, "down"},
		{DirLeft, "left"},
		{DirRight, "right"},
	}

	for _, d := range directions {
		t.Run(d.name, func(t *testing.T) {
			base := HumanoidAerialTemplate(d.dir)
			boss := BossAerialTemplate(base, 2.5)

			baseHead := base.BodyPartLayout[PartHead]
			bossHead := boss.BodyPartLayout[PartHead]

			// Calculate expected boss head position
			offsetX := baseHead.RelativeX - 0.5
			expectedX := 0.5 + (offsetX * 2.5)

			if math.Abs(bossHead.RelativeX-expectedX) > 0.001 {
				t.Errorf("Boss head X %.3f != expected %.3f (base: %.3f, offset: %.3f)",
					bossHead.RelativeX, expectedX, baseHead.RelativeX, offsetX)
			}

			// Verify asymmetry is maintained
			if d.dir == DirLeft && bossHead.RelativeX >= 0.5 {
				t.Error("Boss facing left should have head X < 0.5")
			}
			if d.dir == DirRight && bossHead.RelativeX <= 0.5 {
				t.Error("Boss facing right should have head X > 0.5")
			}
		})
	}
}

// TestBossAerialTemplate_AllGenres verifies that boss scaling works
// correctly with all genre-specific aerial templates.
func TestBossAerialTemplate_AllGenres(t *testing.T) {
	genres := []struct {
		name     string
		template func(Direction) AnatomicalTemplate
	}{
		{"fantasy", FantasyHumanoidAerial},
		{"scifi", SciFiHumanoidAerial},
		{"horror", HorrorHumanoidAerial},
		{"cyberpunk", CyberpunkHumanoidAerial},
		{"postapoc", PostApocHumanoidAerial},
	}

	for _, g := range genres {
		t.Run(g.name, func(t *testing.T) {
			base := g.template(DirDown)
			boss := BossAerialTemplate(base, 2.5)

			// Verify boss has all base parts
			for part := range base.BodyPartLayout {
				if _, exists := boss.BodyPartLayout[part]; !exists {
					t.Errorf("Genre %s boss missing part %s", g.name, part)
				}
			}

			// Verify scaling was applied
			for part, baseSpec := range base.BodyPartLayout {
				bossSpec := boss.BodyPartLayout[part]
				if bossSpec.RelativeWidth <= baseSpec.RelativeWidth {
					t.Errorf("Genre %s part %s width not scaled", g.name, part)
				}
			}

			// Verify Z-index ordering is preserved
			shadow := boss.BodyPartLayout[PartShadow]
			legs := boss.BodyPartLayout[PartLegs]
			torso := boss.BodyPartLayout[PartTorso]
			head := boss.BodyPartLayout[PartHead]

			if shadow.ZIndex >= legs.ZIndex || legs.ZIndex >= torso.ZIndex || torso.ZIndex >= head.ZIndex {
				t.Errorf("Genre %s boss has invalid Z-index ordering", g.name)
			}
		})
	}
}

// TestBossAerialTemplate_InvalidScale verifies that invalid scale values
// are handled gracefully (defaulting to 1.0 for safety).
func TestBossAerialTemplate_InvalidScale(t *testing.T) {
	base := HumanoidAerialTemplate(DirDown)

	tests := []struct {
		name  string
		scale float64
	}{
		{"zero_scale", 0.0},
		{"negative_scale", -1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boss := BossAerialTemplate(base, tt.scale)

			// Should default to 1.0 scale (no scaling)
			for part, baseSpec := range base.BodyPartLayout {
				bossSpec := boss.BodyPartLayout[part]
				if math.Abs(bossSpec.RelativeWidth-baseSpec.RelativeWidth) > 0.001 {
					t.Errorf("Invalid scale %f should result in 1.0x scaling", tt.scale)
				}
			}
		})
	}
}
