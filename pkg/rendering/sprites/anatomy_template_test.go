package sprites

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// TestBodyPart_String tests string representation of body parts.
func TestBodyPart_String(t *testing.T) {
	tests := []struct {
		name string
		part BodyPart
		want string
	}{
		{"shadow", PartShadow, "shadow"},
		{"legs", PartLegs, "legs"},
		{"torso", PartTorso, "torso"},
		{"arms", PartArms, "arms"},
		{"head", PartHead, "head"},
		{"weapon", PartWeapon, "weapon"},
		{"shield", PartShield, "shield"},
		{"unknown", BodyPart(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.part.String()
			if got != tt.want {
				t.Errorf("BodyPart.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHumanoidTemplate tests the humanoid template structure.
func TestHumanoidTemplate(t *testing.T) {
	template := HumanoidTemplate()

	if template.Name != "humanoid" {
		t.Errorf("Template name = %v, want 'humanoid'", template.Name)
	}

	// Verify all expected body parts are present
	expectedParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartArms, PartHead}
	for _, part := range expectedParts {
		if _, exists := template.BodyPartLayout[part]; !exists {
			t.Errorf("Missing body part: %v", part.String())
		}
	}

	// Verify Z-index ordering (shadow < legs < arms < torso < head)
	shadowZ := template.BodyPartLayout[PartShadow].ZIndex
	legsZ := template.BodyPartLayout[PartLegs].ZIndex
	armsZ := template.BodyPartLayout[PartArms].ZIndex
	torsoZ := template.BodyPartLayout[PartTorso].ZIndex
	headZ := template.BodyPartLayout[PartHead].ZIndex

	if !(shadowZ < legsZ && legsZ < armsZ && armsZ < torsoZ && torsoZ < headZ) {
		t.Errorf("Z-index ordering incorrect: shadow=%d, legs=%d, arms=%d, torso=%d, head=%d",
			shadowZ, legsZ, armsZ, torsoZ, headZ)
	}

	// Verify shadow has low opacity
	shadowOpacity := template.BodyPartLayout[PartShadow].Opacity
	if shadowOpacity >= 0.5 {
		t.Errorf("Shadow opacity too high: %v, want < 0.5", shadowOpacity)
	}

	// Verify head is at top (low Y value)
	headY := template.BodyPartLayout[PartHead].RelativeY
	if headY > 0.5 {
		t.Errorf("Head Y position too low: %v, want < 0.5 (top half)", headY)
	}

	// Verify shadow is at bottom (high Y value)
	shadowY := template.BodyPartLayout[PartShadow].RelativeY
	if shadowY < 0.8 {
		t.Errorf("Shadow Y position too high: %v, want > 0.8 (bottom)", shadowY)
	}
}

// TestQuadrupedTemplate tests the quadruped template structure.
func TestQuadrupedTemplate(t *testing.T) {
	template := QuadrupedTemplate()

	if template.Name != "quadruped" {
		t.Errorf("Template name = %v, want 'quadruped'", template.Name)
	}

	// Verify expected parts
	expectedParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartHead}
	for _, part := range expectedParts {
		if _, exists := template.BodyPartLayout[part]; !exists {
			t.Errorf("Missing body part: %v", part.String())
		}
	}

	// Verify horizontal orientation (rotation = 90 for body)
	torsoRotation := template.BodyPartLayout[PartTorso].Rotation
	if torsoRotation != 90 {
		t.Errorf("Torso rotation = %v, want 90 (horizontal)", torsoRotation)
	}
}

// TestBlobTemplate tests the blob template structure.
func TestBlobTemplate(t *testing.T) {
	template := BlobTemplate()

	if template.Name != "blob" {
		t.Errorf("Template name = %v, want 'blob'", template.Name)
	}

	// Blobs should have minimal parts (shadow and torso only)
	if len(template.BodyPartLayout) > 2 {
		t.Errorf("Blob has too many parts: %d, expected 2 (shadow + torso)", len(template.BodyPartLayout))
	}

	// Verify torso uses organic/circular shapes
	torsoSpec := template.BodyPartLayout[PartTorso]
	hasOrganicShape := false
	for _, shapeType := range torsoSpec.ShapeTypes {
		if shapeType == shapes.ShapeOrganic || shapeType == shapes.ShapeCircle {
			hasOrganicShape = true
			break
		}
	}
	if !hasOrganicShape {
		t.Error("Blob torso should use organic or circle shapes")
	}
}

// TestMechanicalTemplate tests the mechanical template structure.
func TestMechanicalTemplate(t *testing.T) {
	template := MechanicalTemplate()

	if template.Name != "mechanical" {
		t.Errorf("Template name = %v, want 'mechanical'", template.Name)
	}

	// Verify geometric shapes are used (rectangles, hexagons, octagons)
	torsoSpec := template.BodyPartLayout[PartTorso]
	hasGeometricShape := false
	for _, shapeType := range torsoSpec.ShapeTypes {
		if shapeType == shapes.ShapeRectangle || shapeType == shapes.ShapeHexagon || shapeType == shapes.ShapeOctagon {
			hasGeometricShape = true
			break
		}
	}
	if !hasGeometricShape {
		t.Error("Mechanical torso should use geometric shapes")
	}
}

// TestFlyingTemplate tests the flying template structure.
func TestFlyingTemplate(t *testing.T) {
	template := FlyingTemplate()

	if template.Name != "flying" {
		t.Errorf("Template name = %v, want 'flying'", template.Name)
	}

	// Verify wings are present (using legs and arms parts)
	if _, hasLeftWing := template.BodyPartLayout[PartLegs]; !hasLeftWing {
		t.Error("Flying template missing left wing (PartLegs)")
	}
	if _, hasRightWing := template.BodyPartLayout[PartArms]; !hasRightWing {
		t.Error("Flying template missing right wing (PartArms)")
	}

	// Verify shadow has reduced opacity (flying creatures cast lighter shadows)
	shadowOpacity := template.BodyPartLayout[PartShadow].Opacity
	if shadowOpacity > 0.3 {
		t.Errorf("Flying shadow opacity too high: %v, want <= 0.3", shadowOpacity)
	}
}

// TestGetSortedParts tests Z-index sorting functionality.
func TestGetSortedParts(t *testing.T) {
	template := HumanoidTemplate()
	sortedParts := template.GetSortedParts()

	if len(sortedParts) != len(template.BodyPartLayout) {
		t.Errorf("Sorted parts count = %d, want %d", len(sortedParts), len(template.BodyPartLayout))
	}

	// Verify parts are sorted by Z-index
	for i := 1; i < len(sortedParts); i++ {
		if sortedParts[i-1].Spec.ZIndex > sortedParts[i].Spec.ZIndex {
			t.Errorf("Parts not sorted by Z-index at position %d: %d > %d",
				i, sortedParts[i-1].Spec.ZIndex, sortedParts[i].Spec.ZIndex)
		}
	}
}

// TestSelectTemplate tests template selection logic.
func TestSelectTemplate(t *testing.T) {
	tests := []struct {
		name         string
		entityType   string
		expectedName string
	}{
		{"humanoid direct", "humanoid", "humanoid"},
		{"player", "player", "humanoid"},
		{"npc", "npc", "humanoid"},
		{"knight", "knight", "humanoid"},
		{"mage", "mage", "humanoid"},
		{"warrior", "warrior", "humanoid"},
		{"quadruped direct", "quadruped", "quadruped"},
		{"wolf", "wolf", "quadruped"},
		{"bear", "bear", "quadruped"},
		{"animal", "animal", "quadruped"},
		{"blob direct", "blob", "blob"},
		{"slime", "slime", "blob"},
		{"amoeba", "amoeba", "blob"},
		{"mechanical direct", "mechanical", "mechanical"},
		{"robot", "robot", "mechanical"},
		{"golem", "golem", "mechanical"},
		{"flying direct", "flying", "flying"},
		{"bird", "bird", "flying"},
		{"dragon", "dragon", "flying"},
		{"unknown defaults to humanoid", "unknown_type", "humanoid"},
		{"empty defaults to humanoid", "", "humanoid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := SelectTemplate(tt.entityType)
			if template.Name != tt.expectedName {
				t.Errorf("SelectTemplate(%q) = %v, want %v", tt.entityType, template.Name, tt.expectedName)
			}
		})
	}
}

// TestPartSpecValidation tests that part specs have valid values.
func TestPartSpecValidation(t *testing.T) {
	templates := []AnatomicalTemplate{
		HumanoidTemplate(),
		QuadrupedTemplate(),
		BlobTemplate(),
		MechanicalTemplate(),
		FlyingTemplate(),
	}

	for _, template := range templates {
		t.Run(template.Name, func(t *testing.T) {
			for part, spec := range template.BodyPartLayout {
				// Verify relative positions are in valid range [0.0, 1.0]
				if spec.RelativeX < 0.0 || spec.RelativeX > 1.0 {
					t.Errorf("%s RelativeX out of range: %v", part.String(), spec.RelativeX)
				}
				if spec.RelativeY < 0.0 || spec.RelativeY > 1.0 {
					t.Errorf("%s RelativeY out of range: %v", part.String(), spec.RelativeY)
				}

				// Verify relative dimensions are in valid range (0.0, 1.0]
				if spec.RelativeWidth <= 0.0 || spec.RelativeWidth > 1.0 {
					t.Errorf("%s RelativeWidth out of range: %v", part.String(), spec.RelativeWidth)
				}
				if spec.RelativeHeight <= 0.0 || spec.RelativeHeight > 1.0 {
					t.Errorf("%s RelativeHeight out of range: %v", part.String(), spec.RelativeHeight)
				}

				// Verify opacity is in valid range [0.0, 1.0]
				if spec.Opacity < 0.0 || spec.Opacity > 1.0 {
					t.Errorf("%s Opacity out of range: %v", part.String(), spec.Opacity)
				}

				// Verify at least one shape type is specified
				if len(spec.ShapeTypes) == 0 {
					t.Errorf("%s has no shape types specified", part.String())
				}

				// Verify color role is not empty
				if spec.ColorRole == "" {
					t.Errorf("%s has empty color role", part.String())
				}
			}
		})
	}
}

// TestTemplateProportions tests that body part proportions are reasonable.
func TestTemplateProportions(t *testing.T) {
	template := HumanoidTemplate()

	// Check head proportions (should be ~25-35% of height)
	headHeight := template.BodyPartLayout[PartHead].RelativeHeight
	if headHeight < 0.20 || headHeight > 0.40 {
		t.Errorf("Head height proportion out of reasonable range: %v, want 0.20-0.40", headHeight)
	}

	// Check torso proportions (should be ~35-50% of height)
	torsoHeight := template.BodyPartLayout[PartTorso].RelativeHeight
	if torsoHeight < 0.30 || torsoHeight > 0.55 {
		t.Errorf("Torso height proportion out of reasonable range: %v, want 0.30-0.55", torsoHeight)
	}

	// Check legs proportions (should be ~25-40% of height)
	legsHeight := template.BodyPartLayout[PartLegs].RelativeHeight
	if legsHeight < 0.20 || legsHeight > 0.45 {
		t.Errorf("Legs height proportion out of reasonable range: %v, want 0.20-0.45", legsHeight)
	}
}

// BenchmarkTemplateSelection benchmarks template selection performance.
func BenchmarkTemplateSelection(b *testing.B) {
	entityTypes := []string{"humanoid", "quadruped", "blob", "mechanical", "flying"}

	for _, entityType := range entityTypes {
		b.Run(entityType, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = SelectTemplate(entityType)
			}
		})
	}
}

// BenchmarkGetSortedParts benchmarks part sorting performance.
func BenchmarkGetSortedParts(b *testing.B) {
	template := HumanoidTemplate()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = template.GetSortedParts()
	}
}

// TestBodyPart_String_Phase52 tests new body part string representations.
func TestBodyPart_String_Phase52(t *testing.T) {
	tests := []struct {
		part BodyPart
		want string
	}{
		{PartHelmet, "helmet"},
		{PartArmor, "armor"},
		{PartTail, "tail"},
		{PartWings, "wings"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.part.String()
			if got != tt.want {
				t.Errorf("BodyPart.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHumanoidDirectionalTemplate tests directional variants (Phase 5.2).
func TestHumanoidDirectionalTemplate(t *testing.T) {
	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}

	for _, dir := range directions {
		t.Run(string(dir), func(t *testing.T) {
			template := HumanoidDirectionalTemplate(dir)

			// Verify template name
			expectedName := "humanoid_" + string(dir)
			if template.Name != expectedName {
				t.Errorf("Template name = %s, want %s", template.Name, expectedName)
			}

			// Verify has required parts
			requiredParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartArms, PartHead}
			for _, part := range requiredParts {
				if _, exists := template.BodyPartLayout[part]; !exists {
					t.Errorf("Missing required part: %s", part.String())
				}
			}

			// Verify arms positioning differs by direction
			armsSpec := template.BodyPartLayout[PartArms]
			switch dir {
			case DirLeft:
				if armsSpec.Rotation != 270 {
					t.Errorf("Left-facing arms rotation = %f, want 270", armsSpec.Rotation)
				}
			case DirRight:
				if armsSpec.Rotation != 90 {
					t.Errorf("Right-facing arms rotation = %f, want 90", armsSpec.Rotation)
				}
			}

			// Verify head positioning for left/right
			headSpec := template.BodyPartLayout[PartHead]
			switch dir {
			case DirLeft:
				if headSpec.RelativeX != 0.45 {
					t.Errorf("Left-facing head X = %f, want 0.45", headSpec.RelativeX)
				}
			case DirRight:
				if headSpec.RelativeX != 0.55 {
					t.Errorf("Right-facing head X = %f, want 0.55", headSpec.RelativeX)
				}
			}
		})
	}
}

// TestHumanoidWithEquipment tests equipment positioning (Phase 5.2).
func TestHumanoidWithEquipment(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		hasWeapon bool
		hasShield bool
	}{
		{"weapon_only_down", DirDown, true, false},
		{"shield_only_down", DirDown, false, true},
		{"both_down", DirDown, true, true},
		{"weapon_only_right", DirRight, true, false},
		{"shield_only_left", DirLeft, false, true},
		{"both_up", DirUp, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := HumanoidWithEquipment(tt.direction, tt.hasWeapon, tt.hasShield)

			// Verify weapon present when requested
			_, hasWeaponPart := template.BodyPartLayout[PartWeapon]
			if tt.hasWeapon && !hasWeaponPart {
				t.Error("Expected weapon part but not found")
			}
			if !tt.hasWeapon && hasWeaponPart {
				t.Error("Unexpected weapon part found")
			}

			// Verify shield present when requested
			_, hasShieldPart := template.BodyPartLayout[PartShield]
			if tt.hasShield && !hasShieldPart {
				t.Error("Expected shield part but not found")
			}
			if !tt.hasShield && hasShieldPart {
				t.Error("Unexpected shield part found")
			}

			// Verify weapon has appropriate shape
			if tt.hasWeapon {
				weaponSpec := template.BodyPartLayout[PartWeapon]
				hasBladeShape := false
				for _, shape := range weaponSpec.ShapeTypes {
					if shape == shapes.ShapeBlade {
						hasBladeShape = true
						break
					}
				}
				if !hasBladeShape {
					t.Error("Weapon should include ShapeBlade")
				}
			}

			// Verify shield has appropriate shape
			if tt.hasShield {
				shieldSpec := template.BodyPartLayout[PartShield]
				hasShieldShape := false
				for _, shape := range shieldSpec.ShapeTypes {
					if shape == shapes.ShapeShield {
						hasShieldShape = true
						break
					}
				}
				if !hasShieldShape {
					t.Error("Shield should include ShapeShield")
				}
			}
		})
	}
}

// TestGenreSpecificHumanoids tests genre-specific template variations (Phase 5.2).
func TestGenreSpecificHumanoids(t *testing.T) {
	tests := []struct {
		name          string
		templateFunc  func(Direction) AnatomicalTemplate
		expectedName  string
		checkFeatures func(*testing.T, AnatomicalTemplate)
	}{
		{
			name:         "fantasy",
			templateFunc: FantasyHumanoidTemplate,
			expectedName: "fantasy_humanoid_down",
			checkFeatures: func(t *testing.T, template AnatomicalTemplate) {
				// Fantasy should have broader shoulders
				torsoSpec := template.BodyPartLayout[PartTorso]
				if torsoSpec.RelativeWidth < 0.54 {
					t.Error("Fantasy humanoid should have broader shoulders")
				}
			},
		},
		{
			name:         "scifi",
			templateFunc: SciFiHumanoidTemplate,
			expectedName: "scifi_humanoid_down",
			checkFeatures: func(t *testing.T, template AnatomicalTemplate) {
				// Sci-fi should have angular shapes
				torsoSpec := template.BodyPartLayout[PartTorso]
				hasAngular := false
				for _, shape := range torsoSpec.ShapeTypes {
					if shape == shapes.ShapeHexagon || shape == shapes.ShapeOctagon {
						hasAngular = true
						break
					}
				}
				if !hasAngular {
					t.Error("Sci-fi humanoid should have angular shapes")
				}
			},
		},
		{
			name:         "horror",
			templateFunc: HorrorHumanoidTemplate,
			expectedName: "horror_humanoid_down",
			checkFeatures: func(t *testing.T, template AnatomicalTemplate) {
				// Horror should have elongated head
				headSpec := template.BodyPartLayout[PartHead]
				if headSpec.RelativeHeight <= 0.35 {
					t.Error("Horror humanoid should have elongated head")
				}
			},
		},
		{
			name:         "cyberpunk",
			templateFunc: CyberpunkHumanoidTemplate,
			expectedName: "cyberpunk_humanoid_down",
			checkFeatures: func(t *testing.T, template AnatomicalTemplate) {
				// Cyberpunk should have compact build
				torsoSpec := template.BodyPartLayout[PartTorso]
				if torsoSpec.RelativeHeight > 0.45 {
					t.Error("Cyberpunk humanoid should have compact torso")
				}
			},
		},
		{
			name:         "postapoc",
			templateFunc: PostApocHumanoidTemplate,
			expectedName: "postapoc_humanoid_down",
			checkFeatures: func(t *testing.T, template AnatomicalTemplate) {
				// Post-apoc should have irregular/organic shapes
				torsoSpec := template.BodyPartLayout[PartTorso]
				hasOrganic := false
				for _, shape := range torsoSpec.ShapeTypes {
					if shape == shapes.ShapeOrganic {
						hasOrganic = true
						break
					}
				}
				if !hasOrganic {
					t.Error("Post-apoc humanoid should have organic shapes")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := tt.templateFunc(DirDown)

			// Verify template name
			if template.Name != tt.expectedName {
				t.Errorf("Template name = %s, want %s", template.Name, tt.expectedName)
			}

			// Verify has all humanoid parts
			requiredParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartArms, PartHead}
			for _, part := range requiredParts {
				if _, exists := template.BodyPartLayout[part]; !exists {
					t.Errorf("Missing required part: %s", part.String())
				}
			}

			// Run genre-specific feature checks
			tt.checkFeatures(t, template)
		})
	}
}

// TestSelectHumanoidTemplate tests the genre-aware template selector (Phase 5.2).
func TestSelectHumanoidTemplate(t *testing.T) {
	tests := []struct {
		genre        string
		entityType   string
		direction    Direction
		expectedName string
	}{
		{"fantasy", "player", DirDown, "fantasy_humanoid_down"},
		{"scifi", "humanoid", DirUp, "scifi_humanoid_up"},
		{"horror", "warrior", DirLeft, "horror_humanoid_left"},
		{"cyberpunk", "knight", DirRight, "cyberpunk_humanoid_right"},
		{"postapoc", "npc", DirDown, "postapoc_humanoid_down"},
		{"unknown", "player", DirDown, "humanoid_down"},
		{"fantasy", "blob", DirDown, "blob"}, // Non-humanoid
	}

	for _, tt := range tests {
		name := tt.genre + "_" + tt.entityType + "_" + string(tt.direction)
		t.Run(name, func(t *testing.T) {
			template := SelectHumanoidTemplate(tt.genre, tt.entityType, tt.direction)

			if template.Name != tt.expectedName {
				t.Errorf("Template name = %s, want %s", template.Name, tt.expectedName)
			}
		})
	}
}

// BenchmarkDirectionalTemplates benchmarks directional template generation.
func BenchmarkDirectionalTemplates(b *testing.B) {
	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}

	for _, dir := range directions {
		b.Run(string(dir), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = HumanoidDirectionalTemplate(dir)
			}
		})
	}
}

// BenchmarkEquipmentTemplates benchmarks equipment template generation.
func BenchmarkEquipmentTemplates(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = HumanoidWithEquipment(DirDown, true, true)
	}
}

// BenchmarkGenreTemplates benchmarks genre-specific template generation.
func BenchmarkGenreTemplates(b *testing.B) {
	genres := []struct {
		name         string
		templateFunc func(Direction) AnatomicalTemplate
	}{
		{"fantasy", FantasyHumanoidTemplate},
		{"scifi", SciFiHumanoidTemplate},
		{"horror", HorrorHumanoidTemplate},
		{"cyberpunk", CyberpunkHumanoidTemplate},
		{"postapoc", PostApocHumanoidTemplate},
	}

	for _, g := range genres {
		b.Run(g.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = g.templateFunc(DirDown)
			}
		})
	}
}

// TestSerpentineTemplate tests the serpentine creature template (Phase 5.3).
func TestSerpentineTemplate(t *testing.T) {
	template := SerpentineTemplate()

	if template.Name != "serpentine" {
		t.Errorf("Template name = %s, want serpentine", template.Name)
	}

	// Verify has required parts
	requiredParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartHead}
	for _, part := range requiredParts {
		if _, exists := template.BodyPartLayout[part]; !exists {
			t.Errorf("Missing required part: %s", part.String())
		}
	}

	// Verify elongated body (torso taller than wide)
	torsoSpec := template.BodyPartLayout[PartTorso]
	if torsoSpec.RelativeHeight <= torsoSpec.RelativeWidth {
		t.Error("Serpentine torso should be taller than wide")
	}

	// Verify wedge-shaped head for snake appearance
	headSpec := template.BodyPartLayout[PartHead]
	hasWedge := false
	for _, shape := range headSpec.ShapeTypes {
		if shape == shapes.ShapeWedge {
			hasWedge = true
			break
		}
	}
	if !hasWedge {
		t.Error("Serpentine head should include wedge shape")
	}
}

// TestArachnidTemplate tests the spider/insect template (Phase 5.3).
func TestArachnidTemplate(t *testing.T) {
	template := ArachnidTemplate()

	if template.Name != "arachnid" {
		t.Errorf("Template name = %s, want arachnid", template.Name)
	}

	// Verify has required parts including multi-leg representation
	requiredParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartHead, PartArms}
	for _, part := range requiredParts {
		if _, exists := template.BodyPartLayout[part]; !exists {
			t.Errorf("Missing required part: %s", part.String())
		}
	}

	// Verify wide leg spread (wider than body)
	legsSpec := template.BodyPartLayout[PartLegs]
	torsoSpec := template.BodyPartLayout[PartTorso]
	if legsSpec.RelativeWidth <= torsoSpec.RelativeWidth {
		t.Error("Arachnid legs should be wider than torso")
	}

	// Verify has lightning shape for leg appearance
	hasLightning := false
	for _, shape := range legsSpec.ShapeTypes {
		if shape == shapes.ShapeLightning {
			hasLightning = true
			break
		}
	}
	if !hasLightning {
		t.Error("Arachnid legs should include lightning shape for multi-leg appearance")
	}
}

// TestUndeadTemplate tests the undead creature template (Phase 5.3).
func TestUndeadTemplate(t *testing.T) {
	template := UndeadTemplate()

	if template.Name != "undead" {
		t.Errorf("Template name = %s, want undead", template.Name)
	}

	// Verify has all humanoid-like parts
	requiredParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartArms, PartHead}
	for _, part := range requiredParts {
		if _, exists := template.BodyPartLayout[part]; !exists {
			t.Errorf("Missing required part: %s", part.String())
		}
	}

	// Verify reduced opacity for ethereal appearance
	for part, spec := range template.BodyPartLayout {
		if part != PartShadow && spec.Opacity > 0.9 {
			t.Errorf("Undead part %s should have reduced opacity, got %f", part.String(), spec.Opacity)
		}
	}

	// Verify skull shape in head
	headSpec := template.BodyPartLayout[PartHead]
	hasSkull := false
	for _, shape := range headSpec.ShapeTypes {
		if shape == shapes.ShapeSkull {
			hasSkull = true
			break
		}
	}
	if !hasSkull {
		t.Error("Undead head should include skull shape")
	}

	// Verify thin limbs
	legsSpec := template.BodyPartLayout[PartLegs]
	if legsSpec.RelativeWidth > 0.30 {
		t.Errorf("Undead legs should be thin, got width %f", legsSpec.RelativeWidth)
	}
}

// TestBossTemplate tests boss scaling (Phase 5.3).
func TestBossTemplate(t *testing.T) {
	tests := []struct {
		name      string
		baseFunc  func() AnatomicalTemplate
		scale     float64
		checkSize func(*testing.T, AnatomicalTemplate, AnatomicalTemplate)
	}{
		{
			name:     "boss_humanoid_2x",
			baseFunc: HumanoidTemplate,
			scale:    2.0,
			checkSize: func(t *testing.T, base, boss AnatomicalTemplate) {
				baseTorso := base.BodyPartLayout[PartTorso]
				bossTorso := boss.BodyPartLayout[PartTorso]
				if bossTorso.RelativeWidth < baseTorso.RelativeWidth*1.9 {
					t.Error("Boss torso should be approximately 2x wider")
				}
			},
		},
		{
			name:     "boss_quadruped_3x",
			baseFunc: QuadrupedTemplate,
			scale:    3.0,
			checkSize: func(t *testing.T, base, boss AnatomicalTemplate) {
				baseTorso := base.BodyPartLayout[PartTorso]
				bossTorso := boss.BodyPartLayout[PartTorso]
				if bossTorso.RelativeHeight < baseTorso.RelativeHeight*2.9 {
					t.Error("Boss torso should be approximately 3x taller")
				}
			},
		},
		{
			name:     "boss_blob_4x",
			baseFunc: BlobTemplate,
			scale:    4.0,
			checkSize: func(t *testing.T, base, boss AnatomicalTemplate) {
				baseTorso := base.BodyPartLayout[PartTorso]
				bossTorso := boss.BodyPartLayout[PartTorso]
				if bossTorso.RelativeWidth < baseTorso.RelativeWidth*3.9 {
					t.Error("Boss torso should be approximately 4x wider")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := tt.baseFunc()
			boss := BossTemplate(base, tt.scale)

			// Verify name prefix
			expectedName := "boss_" + base.Name
			if boss.Name != expectedName {
				t.Errorf("Boss name = %s, want %s", boss.Name, expectedName)
			}

			// Verify all parts copied
			if len(boss.BodyPartLayout) != len(base.BodyPartLayout) {
				t.Errorf("Boss part count = %d, want %d", len(boss.BodyPartLayout), len(base.BodyPartLayout))
			}

			// Run size check
			tt.checkSize(t, base, boss)
		})
	}
}

// TestApplyBossEnhancements tests boss detail enhancements (Phase 5.3).
func TestApplyBossEnhancements(t *testing.T) {
	base := HumanoidTemplate()
	enhanced := ApplyBossEnhancements(base)

	// Verify name prefix
	expectedName := "enhanced_" + base.Name
	if enhanced.Name != expectedName {
		t.Errorf("Enhanced name = %s, want %s", enhanced.Name, expectedName)
	}

	// Verify armor part added
	if _, hasArmor := enhanced.BodyPartLayout[PartArmor]; !hasArmor {
		t.Error("Enhanced boss should have armor part")
	}

	// Verify armor is larger than torso
	armorSpec := enhanced.BodyPartLayout[PartArmor]
	torsoSpec := enhanced.BodyPartLayout[PartTorso]
	if armorSpec.RelativeWidth <= torsoSpec.RelativeWidth {
		t.Error("Boss armor should be larger than torso")
	}

	// Verify armor has lower Z-index (behind torso)
	if armorSpec.ZIndex >= torsoSpec.ZIndex {
		t.Error("Boss armor should render behind torso")
	}
}

// TestSelectTemplate_Phase53 tests new monster archetypes (Phase 5.3).
func TestSelectTemplate_Phase53(t *testing.T) {
	tests := []struct {
		entityType   string
		expectedName string
	}{
		{"serpentine", "serpentine"},
		{"snake", "serpentine"},
		{"worm", "serpentine"},
		{"tentacle", "serpentine"},
		{"wyrm", "serpentine"},
		{"arachnid", "arachnid"},
		{"spider", "arachnid"},
		{"insect", "arachnid"},
		{"beetle", "arachnid"},
		{"undead", "undead"},
		{"skeleton", "undead"},
		{"ghost", "undead"},
		{"zombie", "undead"},
		{"lich", "undead"},
	}

	for _, tt := range tests {
		t.Run(tt.entityType, func(t *testing.T) {
			template := SelectTemplate(tt.entityType)
			if template.Name != tt.expectedName {
				t.Errorf("SelectTemplate(%s) name = %s, want %s", tt.entityType, template.Name, tt.expectedName)
			}
		})
	}
}

// BenchmarkMonsterTemplates benchmarks Phase 5.3 monster template generation.
func BenchmarkMonsterTemplates(b *testing.B) {
	templates := []struct {
		name string
		fn   func() AnatomicalTemplate
	}{
		{"serpentine", SerpentineTemplate},
		{"arachnid", ArachnidTemplate},
		{"undead", UndeadTemplate},
	}

	for _, tmpl := range templates {
		b.Run(tmpl.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = tmpl.fn()
			}
		})
	}
}

// BenchmarkBossScaling benchmarks boss template scaling.
func BenchmarkBossScaling(b *testing.B) {
	base := HumanoidTemplate()
	for i := 0; i < b.N; i++ {
		_ = BossTemplate(base, 2.5)
	}
}
