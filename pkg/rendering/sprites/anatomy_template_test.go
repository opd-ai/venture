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
		{"eyes", PartEyes, "eyes"},
		{"mouth", PartMouth, "mouth"},
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

// TestEnhancedHumanoidTemplate tests the Phase 15.1 enhanced humanoid template.
func TestEnhancedHumanoidTemplate(t *testing.T) {
	template := EnhancedHumanoidTemplate()

	if template.Name != "enhanced_humanoid" {
		t.Errorf("Template name = %v, want 'enhanced_humanoid'", template.Name)
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

	// Verify Phase 15.1 pixel dimensions are set correctly
	// Head should be 4×4 pixels
	headSpec := template.BodyPartLayout[PartHead]
	if headSpec.PreferredPixelSize == nil {
		t.Fatal("Head PreferredPixelSize should not be nil")
	}
	if headSpec.PreferredPixelSize.Width != 4 {
		t.Errorf("Head width = %d, want 4", headSpec.PreferredPixelSize.Width)
	}
	if headSpec.PreferredPixelSize.Height != 4 {
		t.Errorf("Head height = %d, want 4", headSpec.PreferredPixelSize.Height)
	}

	// Torso should be 4×6 pixels
	torsoSpec := template.BodyPartLayout[PartTorso]
	if torsoSpec.PreferredPixelSize == nil {
		t.Fatal("Torso PreferredPixelSize should not be nil")
	}
	if torsoSpec.PreferredPixelSize.Width != 4 {
		t.Errorf("Torso width = %d, want 4", torsoSpec.PreferredPixelSize.Width)
	}
	if torsoSpec.PreferredPixelSize.Height != 6 {
		t.Errorf("Torso height = %d, want 6", torsoSpec.PreferredPixelSize.Height)
	}

	// Legs should be 4×8 pixels
	legsSpec := template.BodyPartLayout[PartLegs]
	if legsSpec.PreferredPixelSize == nil {
		t.Fatal("Legs PreferredPixelSize should not be nil")
	}
	if legsSpec.PreferredPixelSize.Width != 4 {
		t.Errorf("Legs width = %d, want 4", legsSpec.PreferredPixelSize.Width)
	}
	if legsSpec.PreferredPixelSize.Height != 8 {
		t.Errorf("Legs height = %d, want 8", legsSpec.PreferredPixelSize.Height)
	}

	// Arms should have pixel dimensions set
	armsSpec := template.BodyPartLayout[PartArms]
	if armsSpec.PreferredPixelSize == nil {
		t.Fatal("Arms PreferredPixelSize should not be nil")
	}
	if armsSpec.PreferredPixelSize.Width < 1 || armsSpec.PreferredPixelSize.Height < 1 {
		t.Errorf("Arms dimensions invalid: %dx%d", armsSpec.PreferredPixelSize.Width, armsSpec.PreferredPixelSize.Height)
	}

	// Shadow should not have pixel dimensions (scales with sprite)
	shadowSpec := template.BodyPartLayout[PartShadow]
	if shadowSpec.PreferredPixelSize != nil {
		t.Error("Shadow should not have PreferredPixelSize (should scale with sprite)")
	}

	// Verify GetEffectiveWidth/Height return pixel dimensions
	const spriteSize = 28
	headWidth := headSpec.GetEffectiveWidth(spriteSize)
	headHeight := headSpec.GetEffectiveHeight(spriteSize)
	if headWidth != 4 {
		t.Errorf("Head effective width = %d, want 4", headWidth)
	}
	if headHeight != 4 {
		t.Errorf("Head effective height = %d, want 4", headHeight)
	}

	torsoWidth := torsoSpec.GetEffectiveWidth(spriteSize)
	torsoHeight := torsoSpec.GetEffectiveHeight(spriteSize)
	if torsoWidth != 4 {
		t.Errorf("Torso effective width = %d, want 4", torsoWidth)
	}
	if torsoHeight != 6 {
		t.Errorf("Torso effective height = %d, want 6", torsoHeight)
	}

	legsWidth := legsSpec.GetEffectiveWidth(spriteSize)
	legsHeight := legsSpec.GetEffectiveHeight(spriteSize)
	if legsWidth != 4 {
		t.Errorf("Legs effective width = %d, want 4", legsWidth)
	}
	if legsHeight != 8 {
		t.Errorf("Legs effective height = %d, want 8", legsHeight)
	}

	// Verify proportions remain constant across different sprite sizes
	const largeSpriteSize = 64
	headWidthLarge := headSpec.GetEffectiveWidth(largeSpriteSize)
	headHeightLarge := headSpec.GetEffectiveHeight(largeSpriteSize)
	if headWidthLarge != 4 {
		t.Errorf("Head effective width at 64x64 = %d, want 4 (pixel-perfect should be constant)", headWidthLarge)
	}
	if headHeightLarge != 4 {
		t.Errorf("Head effective height at 64x64 = %d, want 4 (pixel-perfect should be constant)", headHeightLarge)
	}
}

// TestDetailedHumanoidTemplate tests the Phase 15.1 detailed template with facial features.
func TestDetailedHumanoidTemplate(t *testing.T) {
	template := DetailedHumanoidTemplate()

	if template.Name != "detailed_humanoid" {
		t.Errorf("Template name = %v, want 'detailed_humanoid'", template.Name)
	}

	// Verify all expected body parts including facial features
	expectedParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartArms, PartHead, PartEyes, PartMouth}
	for _, part := range expectedParts {
		if _, exists := template.BodyPartLayout[part]; !exists {
			t.Errorf("Missing body part: %v", part.String())
		}
	}

	// Verify eyes have correct pixel dimensions (2×1 pixels)
	eyesSpec := template.BodyPartLayout[PartEyes]
	if eyesSpec.PreferredPixelSize == nil {
		t.Fatal("Eyes PreferredPixelSize should not be nil")
	}
	if eyesSpec.PreferredPixelSize.Width != 2 {
		t.Errorf("Eyes width = %d, want 2", eyesSpec.PreferredPixelSize.Width)
	}
	if eyesSpec.PreferredPixelSize.Height != 1 {
		t.Errorf("Eyes height = %d, want 1", eyesSpec.PreferredPixelSize.Height)
	}

	// Verify mouth has correct pixel dimensions (2×1 pixels)
	mouthSpec := template.BodyPartLayout[PartMouth]
	if mouthSpec.PreferredPixelSize == nil {
		t.Fatal("Mouth PreferredPixelSize should not be nil")
	}
	if mouthSpec.PreferredPixelSize.Width != 2 {
		t.Errorf("Mouth width = %d, want 2", mouthSpec.PreferredPixelSize.Width)
	}
	if mouthSpec.PreferredPixelSize.Height != 1 {
		t.Errorf("Mouth height = %d, want 1", mouthSpec.PreferredPixelSize.Height)
	}

	// Verify facial features have higher Z-index than head
	headZ := template.BodyPartLayout[PartHead].ZIndex
	eyesZ := eyesSpec.ZIndex
	mouthZ := mouthSpec.ZIndex

	if eyesZ <= headZ {
		t.Errorf("Eyes Z-index (%d) should be above head Z-index (%d)", eyesZ, headZ)
	}
	if mouthZ <= headZ {
		t.Errorf("Mouth Z-index (%d) should be above head Z-index (%d)", mouthZ, headZ)
	}

	// Verify eyes are positioned above mouth
	eyesY := eyesSpec.RelativeY
	mouthY := mouthSpec.RelativeY
	if eyesY >= mouthY {
		t.Errorf("Eyes Y position (%f) should be above (less than) mouth Y position (%f)", eyesY, mouthY)
	}

	// Verify GetEffectiveWidth/Height work correctly for facial features
	const spriteSize = 28
	eyesWidth := eyesSpec.GetEffectiveWidth(spriteSize)
	eyesHeight := eyesSpec.GetEffectiveHeight(spriteSize)
	if eyesWidth != 2 {
		t.Errorf("Eyes effective width = %d, want 2", eyesWidth)
	}
	if eyesHeight != 1 {
		t.Errorf("Eyes effective height = %d, want 1", eyesHeight)
	}

	mouthWidth := mouthSpec.GetEffectiveWidth(spriteSize)
	mouthHeight := mouthSpec.GetEffectiveHeight(spriteSize)
	if mouthWidth != 2 {
		t.Errorf("Mouth effective width = %d, want 2", mouthWidth)
	}
	if mouthHeight != 1 {
		t.Errorf("Mouth effective height = %d, want 1", mouthHeight)
	}

	// Verify base template dimensions are preserved (head, torso, legs)
	headSpec := template.BodyPartLayout[PartHead]
	if headSpec.PreferredPixelSize == nil || headSpec.PreferredPixelSize.Width != 4 || headSpec.PreferredPixelSize.Height != 4 {
		t.Error("Head dimensions should be preserved from EnhancedHumanoidTemplate (4×4)")
	}

	torsoSpec := template.BodyPartLayout[PartTorso]
	if torsoSpec.PreferredPixelSize == nil || torsoSpec.PreferredPixelSize.Width != 4 || torsoSpec.PreferredPixelSize.Height != 6 {
		t.Error("Torso dimensions should be preserved from EnhancedHumanoidTemplate (4×6)")
	}

	legsSpec := template.BodyPartLayout[PartLegs]
	if legsSpec.PreferredPixelSize == nil || legsSpec.PreferredPixelSize.Width != 4 || legsSpec.PreferredPixelSize.Height != 8 {
		t.Error("Legs dimensions should be preserved from EnhancedHumanoidTemplate (4×8)")
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

// TestHumanoidAerialTemplate tests aerial-view humanoid templates (Phase 1).
func TestHumanoidAerialTemplate(t *testing.T) {
	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}

	for _, dir := range directions {
		t.Run(string(dir), func(t *testing.T) {
			template := HumanoidAerialTemplate(dir)

			// Verify template name
			expectedName := "humanoid_aerial_" + string(dir)
			if template.Name != expectedName {
				t.Errorf("Template name = %s, want %s", template.Name, expectedName)
			}

			// Verify all required aerial parts are present
			requiredParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartArms, PartHead}
			for _, part := range requiredParts {
				if _, exists := template.BodyPartLayout[part]; !exists {
					t.Errorf("Missing required part: %s", part.String())
				}
			}

			// Verify aerial proportions (35/50/15 for head/torso/legs)
			headSpec := template.BodyPartLayout[PartHead]
			torsoSpec := template.BodyPartLayout[PartTorso]
			legsSpec := template.BodyPartLayout[PartLegs]

			// Head should be ~35% (±0.02 tolerance)
			if headSpec.RelativeHeight < 0.33 || headSpec.RelativeHeight > 0.37 {
				t.Errorf("Aerial head height = %f, want 0.33-0.37 (35%% ±2%%)", headSpec.RelativeHeight)
			}

			// Torso should be ~50% (±0.03 tolerance)
			if torsoSpec.RelativeHeight < 0.47 || torsoSpec.RelativeHeight > 0.53 {
				t.Errorf("Aerial torso height = %f, want 0.47-0.53 (50%% ±3%%)", torsoSpec.RelativeHeight)
			}

			// Legs should be ~15% (±0.02 tolerance)
			if legsSpec.RelativeHeight < 0.13 || legsSpec.RelativeHeight > 0.17 {
				t.Errorf("Aerial legs height = %f, want 0.13-0.17 (15%% ±2%%)", legsSpec.RelativeHeight)
			}

			// Verify shadow is ellipse for aerial depth perception
			shadowSpec := template.BodyPartLayout[PartShadow]
			hasEllipse := false
			for _, shape := range shadowSpec.ShapeTypes {
				if shape == shapes.ShapeEllipse {
					hasEllipse = true
					break
				}
			}
			if !hasEllipse {
				t.Error("Aerial shadow should use ellipse shape")
			}
		})
	}
}

// TestAerialDirectionalAsymmetry tests that aerial templates create directional asymmetry.
func TestAerialDirectionalAsymmetry(t *testing.T) {
	tests := []struct {
		direction      Direction
		checkAsymmetry func(*testing.T, AnatomicalTemplate)
	}{
		{
			direction: DirUp,
			checkAsymmetry: func(t *testing.T, template AnatomicalTemplate) {
				// Up: head centered, arms behind torso
				headSpec := template.BodyPartLayout[PartHead]
				if headSpec.RelativeX != 0.5 {
					t.Errorf("DirUp head should be centered (X=0.5), got %f", headSpec.RelativeX)
				}
				armsSpec := template.BodyPartLayout[PartArms]
				torsoSpec := template.BodyPartLayout[PartTorso]
				if armsSpec.ZIndex >= torsoSpec.ZIndex {
					t.Error("DirUp arms should be behind torso (lower ZIndex)")
				}
			},
		},
		{
			direction: DirDown,
			checkAsymmetry: func(t *testing.T, template AnatomicalTemplate) {
				// Down: head centered, arms in front of torso
				headSpec := template.BodyPartLayout[PartHead]
				if headSpec.RelativeX != 0.5 {
					t.Errorf("DirDown head should be centered (X=0.5), got %f", headSpec.RelativeX)
				}
				armsSpec := template.BodyPartLayout[PartArms]
				torsoSpec := template.BodyPartLayout[PartTorso]
				if armsSpec.ZIndex <= torsoSpec.ZIndex {
					t.Error("DirDown arms should be in front of torso (higher ZIndex)")
				}
			},
		},
		{
			direction: DirLeft,
			checkAsymmetry: func(t *testing.T, template AnatomicalTemplate) {
				// Left: head shifted left, arms rotated 270°
				headSpec := template.BodyPartLayout[PartHead]
				if headSpec.RelativeX >= 0.5 {
					t.Errorf("DirLeft head should be left of center (X<0.5), got %f", headSpec.RelativeX)
				}
				armsSpec := template.BodyPartLayout[PartArms]
				if armsSpec.Rotation != 270 {
					t.Errorf("DirLeft arms rotation = %f, want 270", armsSpec.Rotation)
				}
			},
		},
		{
			direction: DirRight,
			checkAsymmetry: func(t *testing.T, template AnatomicalTemplate) {
				// Right: head shifted right, arms rotated 90°
				headSpec := template.BodyPartLayout[PartHead]
				if headSpec.RelativeX <= 0.5 {
					t.Errorf("DirRight head should be right of center (X>0.5), got %f", headSpec.RelativeX)
				}
				armsSpec := template.BodyPartLayout[PartArms]
				if armsSpec.Rotation != 90 {
					t.Errorf("DirRight arms rotation = %f, want 90", armsSpec.Rotation)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.direction), func(t *testing.T) {
			template := HumanoidAerialTemplate(tt.direction)
			tt.checkAsymmetry(t, template)
		})
	}
}

// TestAerialGenreVariants tests genre-specific aerial templates.
func TestAerialGenreVariants(t *testing.T) {
	tests := []struct {
		name          string
		templateFunc  func(Direction) AnatomicalTemplate
		expectedName  string
		checkFeatures func(*testing.T, AnatomicalTemplate)
	}{
		{
			name:         "fantasy_aerial",
			templateFunc: FantasyHumanoidAerial,
			expectedName: "fantasy_aerial_down",
			checkFeatures: func(t *testing.T, template AnatomicalTemplate) {
				// Fantasy should have broader shoulders (wider torso)
				torsoSpec := template.BodyPartLayout[PartTorso]
				if torsoSpec.RelativeWidth < 0.64 {
					t.Errorf("Fantasy aerial should have broad shoulders (torso width >= 0.64), got %f", torsoSpec.RelativeWidth)
				}
				// Check for helmet shape in head
				headSpec := template.BodyPartLayout[PartHead]
				hasHelmetShape := false
				for _, shape := range headSpec.ShapeTypes {
					if shape == shapes.ShapeHexagon || shape == shapes.ShapeOctagon {
						hasHelmetShape = true
						break
					}
				}
				if !hasHelmetShape {
					t.Error("Fantasy aerial head should include helmet shapes (hexagon/octagon)")
				}
			},
		},
		{
			name:         "scifi_aerial",
			templateFunc: SciFiHumanoidAerial,
			expectedName: "scifi_aerial_down",
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
					t.Error("Sci-fi aerial should have angular torso shapes")
				}
				// Check for angular head
				headSpec := template.BodyPartLayout[PartHead]
				hasAngularHead := false
				for _, shape := range headSpec.ShapeTypes {
					if shape == shapes.ShapeOctagon || shape == shapes.ShapeHexagon {
						hasAngularHead = true
						break
					}
				}
				if !hasAngularHead {
					t.Error("Sci-fi aerial head should be angular (octagon/hexagon)")
				}
			},
		},
		{
			name:         "horror_aerial",
			templateFunc: HorrorHumanoidAerial,
			expectedName: "horror_aerial_down",
			checkFeatures: func(t *testing.T, template AnatomicalTemplate) {
				// Horror should maintain 35% head height for proportion consistency
				headSpec := template.BodyPartLayout[PartHead]
				if headSpec.RelativeHeight < 0.30 || headSpec.RelativeHeight > 0.40 {
					t.Errorf("Horror aerial head height should be ~0.35 (±0.05), got %f", headSpec.RelativeHeight)
				}
				// Horror aesthetic achieved through narrow width (elongated visual effect)
				if headSpec.RelativeWidth > 0.30 {
					t.Errorf("Horror aerial head should be narrow (width <= 0.30), got %f", headSpec.RelativeWidth)
				}
				// Check for reduced shadow opacity (ghostly effect)
				shadowSpec := template.BodyPartLayout[PartShadow]
				if shadowSpec.Opacity > 0.25 {
					t.Errorf("Horror aerial shadow should be faint (opacity <= 0.25), got %f", shadowSpec.Opacity)
				}
			},
		},
		{
			name:         "cyberpunk_aerial",
			templateFunc: CyberpunkHumanoidAerial,
			expectedName: "cyberpunk_aerial_down",
			checkFeatures: func(t *testing.T, template AnatomicalTemplate) {
				// Cyberpunk should have compact build
				torsoSpec := template.BodyPartLayout[PartTorso]
				if torsoSpec.RelativeHeight > 0.50 {
					t.Errorf("Cyberpunk aerial should have compact torso (height <= 0.50), got %f", torsoSpec.RelativeHeight)
				}
				// Check for neon glow overlay (armor part)
				if _, hasArmor := template.BodyPartLayout[PartArmor]; !hasArmor {
					t.Error("Cyberpunk aerial should have neon glow overlay (armor part)")
				}
			},
		},
		{
			name:         "postapoc_aerial",
			templateFunc: PostApocHumanoidAerial,
			expectedName: "postapoc_aerial_down",
			checkFeatures: func(t *testing.T, template AnatomicalTemplate) {
				// Post-apoc should have irregular shapes
				torsoSpec := template.BodyPartLayout[PartTorso]
				hasOrganic := false
				for _, shape := range torsoSpec.ShapeTypes {
					if shape == shapes.ShapeOrganic {
						hasOrganic = true
						break
					}
				}
				if !hasOrganic {
					t.Error("Post-apoc aerial should have organic/ragged torso shapes")
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

			// Verify all aerial parts present
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

// TestSelectAerialTemplate tests the aerial template dispatcher.
func TestSelectAerialTemplate(t *testing.T) {
	tests := []struct {
		name           string
		entityType     string
		genre          string
		direction      Direction
		expectedName   string
		shouldBeAerial bool
	}{
		{"humanoid_fantasy", "player", "fantasy", DirDown, "fantasy_aerial_down", true},
		{"humanoid_scifi", "humanoid", "scifi", DirUp, "scifi_aerial_up", true},
		{"humanoid_horror", "warrior", "horror", DirLeft, "horror_aerial_left", true},
		{"humanoid_cyberpunk", "knight", "cyberpunk", DirRight, "cyberpunk_aerial_right", true},
		{"humanoid_postapoc", "npc", "postapoc", DirDown, "postapoc_aerial_down", true},
		{"humanoid_unknown_genre", "player", "unknown", DirDown, "humanoid_aerial_down", true},
		{"non_humanoid_blob", "blob", "fantasy", DirDown, "blob", false},
		{"non_humanoid_quadruped", "wolf", "scifi", DirUp, "quadruped", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := SelectAerialTemplate(tt.entityType, tt.genre, tt.direction)

			if template.Name != tt.expectedName {
				t.Errorf("Template name = %s, want %s", template.Name, tt.expectedName)
			}

			// Verify aerial templates have proper proportions
			if tt.shouldBeAerial {
				torsoSpec := template.BodyPartLayout[PartTorso]
				// Aerial torsos should be ~50% height
				if torsoSpec.RelativeHeight < 0.45 || torsoSpec.RelativeHeight > 0.55 {
					t.Errorf("Aerial torso height should be ~50%%, got %f", torsoSpec.RelativeHeight)
				}
			}
		})
	}
}

// TestAerialTemplate_Determinism tests that aerial templates are deterministic.
func TestAerialTemplate_Determinism(t *testing.T) {
	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}

	for _, dir := range directions {
		t.Run(string(dir), func(t *testing.T) {
			// Generate template twice
			template1 := HumanoidAerialTemplate(dir)
			template2 := HumanoidAerialTemplate(dir)

			// Verify names match
			if template1.Name != template2.Name {
				t.Errorf("Template names differ: %s vs %s", template1.Name, template2.Name)
			}

			// Verify same number of parts
			if len(template1.BodyPartLayout) != len(template2.BodyPartLayout) {
				t.Errorf("Part count differs: %d vs %d", len(template1.BodyPartLayout), len(template2.BodyPartLayout))
			}

			// Verify all specs match
			for part, spec1 := range template1.BodyPartLayout {
				spec2, exists := template2.BodyPartLayout[part]
				if !exists {
					t.Errorf("Part %s missing in second generation", part.String())
					continue
				}

				// Compare all fields
				if spec1.RelativeX != spec2.RelativeX {
					t.Errorf("Part %s RelativeX differs: %f vs %f", part.String(), spec1.RelativeX, spec2.RelativeX)
				}
				if spec1.RelativeY != spec2.RelativeY {
					t.Errorf("Part %s RelativeY differs: %f vs %f", part.String(), spec1.RelativeY, spec2.RelativeY)
				}
				if spec1.ZIndex != spec2.ZIndex {
					t.Errorf("Part %s ZIndex differs: %d vs %d", part.String(), spec1.ZIndex, spec2.ZIndex)
				}
				if spec1.Rotation != spec2.Rotation {
					t.Errorf("Part %s Rotation differs: %f vs %f", part.String(), spec1.Rotation, spec2.Rotation)
				}
			}
		})
	}
}

// TestAerialProportions_Standard tests that all aerial templates follow standard proportions.
func TestAerialProportions_Standard(t *testing.T) {
	genres := []struct {
		name         string
		templateFunc func(Direction) AnatomicalTemplate
	}{
		{"base", HumanoidAerialTemplate},
		{"fantasy", FantasyHumanoidAerial},
		{"scifi", SciFiHumanoidAerial},
		{"horror", HorrorHumanoidAerial},
		{"cyberpunk", CyberpunkHumanoidAerial},
		{"postapoc", PostApocHumanoidAerial},
	}

	for _, g := range genres {
		t.Run(g.name, func(t *testing.T) {
			template := g.templateFunc(DirDown)

			// Head: 35% ± 5% tolerance (allow genre variation)
			headSpec := template.BodyPartLayout[PartHead]
			if headSpec.RelativeHeight < 0.28 || headSpec.RelativeHeight > 0.42 {
				t.Errorf("%s head height = %f, want 0.28-0.42 (35%% ±7%%)", g.name, headSpec.RelativeHeight)
			}

			// Torso: 50% ± 5% tolerance
			torsoSpec := template.BodyPartLayout[PartTorso]
			if torsoSpec.RelativeHeight < 0.45 || torsoSpec.RelativeHeight > 0.55 {
				t.Errorf("%s torso height = %f, want 0.45-0.55 (50%% ±5%%)", g.name, torsoSpec.RelativeHeight)
			}

			// Legs: 15% ± 3% tolerance (minimal from aerial view)
			legsSpec := template.BodyPartLayout[PartLegs]
			if legsSpec.RelativeHeight < 0.12 || legsSpec.RelativeHeight > 0.18 {
				t.Errorf("%s legs height = %f, want 0.12-0.18 (15%% ±3%%)", g.name, legsSpec.RelativeHeight)
			}

			// Verify reduced leg opacity for aerial perspective
			if legsSpec.Opacity > 0.85 {
				t.Errorf("%s legs should have reduced opacity for aerial view, got %f", g.name, legsSpec.Opacity)
			}
		})
	}
}

// BenchmarkAerialTemplates benchmarks aerial template generation performance.
func BenchmarkAerialTemplates(b *testing.B) {
	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}

	for _, dir := range directions {
		b.Run("base_"+string(dir), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = HumanoidAerialTemplate(dir)
			}
		})
	}
}

// BenchmarkAerialGenreTemplates benchmarks genre-specific aerial template generation.
func BenchmarkAerialGenreTemplates(b *testing.B) {
	genres := []struct {
		name         string
		templateFunc func(Direction) AnatomicalTemplate
	}{
		{"fantasy", FantasyHumanoidAerial},
		{"scifi", SciFiHumanoidAerial},
		{"horror", HorrorHumanoidAerial},
		{"cyberpunk", CyberpunkHumanoidAerial},
		{"postapoc", PostApocHumanoidAerial},
	}

	for _, g := range genres {
		b.Run(g.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = g.templateFunc(DirDown)
			}
		})
	}
}

// BenchmarkSelectAerialTemplate benchmarks the aerial template dispatcher.
func BenchmarkSelectAerialTemplate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SelectAerialTemplate("player", "fantasy", DirDown)
	}
}

// TestPixelDimensions tests the PixelDimensions struct.
func TestPixelDimensions(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"head 4x4", 4, 4},
		{"torso 4x6", 4, 6},
		{"legs 4x8", 4, 8},
		{"boss head 8x8", 8, 8},
		{"zero dimensions", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pd := PixelDimensions{
				Width:  tt.width,
				Height: tt.height,
			}

			if pd.Width != tt.width {
				t.Errorf("PixelDimensions.Width = %d, want %d", pd.Width, tt.width)
			}
			if pd.Height != tt.height {
				t.Errorf("PixelDimensions.Height = %d, want %d", pd.Height, tt.height)
			}
		})
	}
}

// TestPartSpec_GetEffectiveWidth tests effective width calculation.
func TestPartSpec_GetEffectiveWidth(t *testing.T) {
	tests := []struct {
		name        string
		spec        PartSpec
		spriteWidth int
		want        int
	}{
		{
			name: "with pixel dimensions",
			spec: PartSpec{
				RelativeWidth: 0.5,
				PreferredPixelSize: &PixelDimensions{
					Width:  4,
					Height: 4,
				},
			},
			spriteWidth: 28,
			want:        4, // Uses PreferredPixelSize, not RelativeWidth
		},
		{
			name: "without pixel dimensions",
			spec: PartSpec{
				RelativeWidth:      0.5,
				PreferredPixelSize: nil,
			},
			spriteWidth: 28,
			want:        14, // 28 * 0.5 = 14
		},
		{
			name: "relative width 0.35",
			spec: PartSpec{
				RelativeWidth:      0.35,
				PreferredPixelSize: nil,
			},
			spriteWidth: 28,
			want:        9, // 28 * 0.35 = 9.8, truncates to 9
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.GetEffectiveWidth(tt.spriteWidth)
			if got != tt.want {
				t.Errorf("GetEffectiveWidth() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestPartSpec_GetEffectiveHeight tests effective height calculation.
func TestPartSpec_GetEffectiveHeight(t *testing.T) {
	tests := []struct {
		name         string
		spec         PartSpec
		spriteHeight int
		want         int
	}{
		{
			name: "with pixel dimensions",
			spec: PartSpec{
				RelativeHeight: 0.5,
				PreferredPixelSize: &PixelDimensions{
					Width:  4,
					Height: 6,
				},
			},
			spriteHeight: 28,
			want:         6, // Uses PreferredPixelSize, not RelativeHeight
		},
		{
			name: "without pixel dimensions",
			spec: PartSpec{
				RelativeHeight:     0.5,
				PreferredPixelSize: nil,
			},
			spriteHeight: 28,
			want:         14, // 28 * 0.5 = 14
		},
		{
			name: "relative height 0.45",
			spec: PartSpec{
				RelativeHeight:     0.45,
				PreferredPixelSize: nil,
			},
			spriteHeight: 28,
			want:         12, // 28 * 0.45 = 12.6, truncates to 12
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.GetEffectiveHeight(tt.spriteHeight)
			if got != tt.want {
				t.Errorf("GetEffectiveHeight() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestPartSpec_ToPixelDimensions tests conversion from relative to pixel dimensions.
func TestPartSpec_ToPixelDimensions(t *testing.T) {
	tests := []struct {
		name         string
		spec         PartSpec
		spriteWidth  int
		spriteHeight int
		wantWidth    int
		wantHeight   int
	}{
		{
			name: "half size sprite",
			spec: PartSpec{
				RelativeWidth:  0.5,
				RelativeHeight: 0.5,
			},
			spriteWidth:  28,
			spriteHeight: 28,
			wantWidth:    14,
			wantHeight:   14,
		},
		{
			name: "head proportions",
			spec: PartSpec{
				RelativeWidth:  0.35,
				RelativeHeight: 0.35,
			},
			spriteWidth:  28,
			spriteHeight: 28,
			wantWidth:    9, // 28 * 0.35 = 9.8, truncates to 9
			wantHeight:   9, // 28 * 0.35 = 9.8, truncates to 9
		},
		{
			name: "torso proportions",
			spec: PartSpec{
				RelativeWidth:  0.50,
				RelativeHeight: 0.45,
			},
			spriteWidth:  28,
			spriteHeight: 28,
			wantWidth:    14, // 28 * 0.50 = 14
			wantHeight:   12, // 28 * 0.45 = 12.6, truncates to 12
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.spec.ToPixelDimensions(tt.spriteWidth, tt.spriteHeight)
			if got.Width != tt.wantWidth {
				t.Errorf("ToPixelDimensions().Width = %d, want %d", got.Width, tt.wantWidth)
			}
			if got.Height != tt.wantHeight {
				t.Errorf("ToPixelDimensions().Height = %d, want %d", got.Height, tt.wantHeight)
			}
		})
	}
}

// TestPartSpec_WithPixelDimensions tests adding pixel dimensions to existing spec.
func TestPartSpec_WithPixelDimensions(t *testing.T) {
	originalSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.25,
		RelativeWidth:  0.35,
		RelativeHeight: 0.35,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle},
		ZIndex:         15,
		ColorRole:      "secondary",
		Opacity:        1.0,
		Rotation:       0,
	}

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"head 4x4", 4, 4},
		{"torso 4x6", 4, 6},
		{"legs 4x8", 4, 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := originalSpec.WithPixelDimensions(tt.width, tt.height)

			// Check pixel dimensions are set correctly
			if got.PreferredPixelSize == nil {
				t.Fatal("PreferredPixelSize should not be nil")
			}
			if got.PreferredPixelSize.Width != tt.width {
				t.Errorf("PreferredPixelSize.Width = %d, want %d", got.PreferredPixelSize.Width, tt.width)
			}
			if got.PreferredPixelSize.Height != tt.height {
				t.Errorf("PreferredPixelSize.Height = %d, want %d", got.PreferredPixelSize.Height, tt.height)
			}

			// Check all other fields are preserved
			if got.RelativeX != originalSpec.RelativeX {
				t.Errorf("RelativeX = %f, want %f", got.RelativeX, originalSpec.RelativeX)
			}
			if got.RelativeY != originalSpec.RelativeY {
				t.Errorf("RelativeY = %f, want %f", got.RelativeY, originalSpec.RelativeY)
			}
			if got.ZIndex != originalSpec.ZIndex {
				t.Errorf("ZIndex = %d, want %d", got.ZIndex, originalSpec.ZIndex)
			}
			if got.ColorRole != originalSpec.ColorRole {
				t.Errorf("ColorRole = %s, want %s", got.ColorRole, originalSpec.ColorRole)
			}

			// Original spec should be unchanged (value receiver)
			if originalSpec.PreferredPixelSize != nil {
				t.Error("Original spec should not have PreferredPixelSize set")
			}
		})
	}
}

// TestNewPartSpecFromPixels tests creating spec from pixel dimensions.
func TestNewPartSpecFromPixels(t *testing.T) {
	tests := []struct {
		name      string
		width     int
		height    int
		shapeType shapes.ShapeType
		zIndex    int
		colorRole string
	}{
		{"head 4x4", 4, 4, shapes.ShapeCircle, 15, "secondary"},
		{"torso 4x6", 4, 6, shapes.ShapeRectangle, 10, "primary"},
		{"legs 4x8", 4, 8, shapes.ShapeCapsule, 5, "primary"},
		{"boss head 8x8", 8, 8, shapes.ShapeSkull, 15, "accent1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewPartSpecFromPixels(tt.width, tt.height, tt.shapeType, tt.zIndex, tt.colorRole)

			// Check pixel dimensions
			if got.PreferredPixelSize == nil {
				t.Fatal("PreferredPixelSize should not be nil")
			}
			if got.PreferredPixelSize.Width != tt.width {
				t.Errorf("PreferredPixelSize.Width = %d, want %d", got.PreferredPixelSize.Width, tt.width)
			}
			if got.PreferredPixelSize.Height != tt.height {
				t.Errorf("PreferredPixelSize.Height = %d, want %d", got.PreferredPixelSize.Height, tt.height)
			}

			// Check other fields
			if got.ZIndex != tt.zIndex {
				t.Errorf("ZIndex = %d, want %d", got.ZIndex, tt.zIndex)
			}
			if got.ColorRole != tt.colorRole {
				t.Errorf("ColorRole = %s, want %s", got.ColorRole, tt.colorRole)
			}
			if len(got.ShapeTypes) != 1 || got.ShapeTypes[0] != tt.shapeType {
				t.Errorf("ShapeTypes = %v, want [%v]", got.ShapeTypes, tt.shapeType)
			}

			// Check defaults
			if got.RelativeX != 0.5 {
				t.Errorf("RelativeX = %f, want 0.5", got.RelativeX)
			}
			if got.RelativeY != 0.5 {
				t.Errorf("RelativeY = %f, want 0.5", got.RelativeY)
			}
			if got.Opacity != 1.0 {
				t.Errorf("Opacity = %f, want 1.0", got.Opacity)
			}
			if got.Rotation != 0 {
				t.Errorf("Rotation = %f, want 0", got.Rotation)
			}

			// Check relative dimensions are calculated (as fallbacks)
			const typicalSize = 28.0
			expectedRelWidth := float64(tt.width) / typicalSize
			expectedRelHeight := float64(tt.height) / typicalSize
			if got.RelativeWidth != expectedRelWidth {
				t.Errorf("RelativeWidth = %f, want %f", got.RelativeWidth, expectedRelWidth)
			}
			if got.RelativeHeight != expectedRelHeight {
				t.Errorf("RelativeHeight = %f, want %f", got.RelativeHeight, expectedRelHeight)
			}
		})
	}
}

// TestPhase151EnhancedProportionalScaling demonstrates Phase 15.1 usage.
// This integration test verifies that the "head 4×4, torso 4×6, legs 4×8" specification
// from Phase 15.1 can be implemented using the new pixel dimension support.
func TestPhase151EnhancedProportionalScaling(t *testing.T) {
	// Create Phase 15.1 humanoid with exact pixel dimensions
	head := NewPartSpecFromPixels(4, 4, shapes.ShapeCircle, 15, "secondary")
	torso := NewPartSpecFromPixels(4, 6, shapes.ShapeRectangle, 10, "primary")
	legs := NewPartSpecFromPixels(4, 8, shapes.ShapeCapsule, 5, "primary")

	// Verify exact pixel dimensions
	if head.GetEffectiveWidth(28) != 4 {
		t.Errorf("head width = %d, want 4", head.GetEffectiveWidth(28))
	}
	if head.GetEffectiveHeight(28) != 4 {
		t.Errorf("head height = %d, want 4", head.GetEffectiveHeight(28))
	}

	if torso.GetEffectiveWidth(28) != 4 {
		t.Errorf("torso width = %d, want 4", torso.GetEffectiveWidth(28))
	}
	if torso.GetEffectiveHeight(28) != 6 {
		t.Errorf("torso height = %d, want 6", torso.GetEffectiveHeight(28))
	}

	if legs.GetEffectiveWidth(28) != 4 {
		t.Errorf("legs width = %d, want 4", legs.GetEffectiveWidth(28))
	}
	if legs.GetEffectiveHeight(28) != 8 {
		t.Errorf("legs height = %d, want 8", legs.GetEffectiveHeight(28))
	}

	// Verify dimensions work correctly at different sprite sizes
	// (pixel dimensions should be constant regardless of sprite size)
	if head.GetEffectiveWidth(32) != 4 {
		t.Errorf("head width at 32x32 = %d, want 4", head.GetEffectiveWidth(32))
	}
	if torso.GetEffectiveHeight(64) != 6 {
		t.Errorf("torso height at 64x64 = %d, want 6", torso.GetEffectiveHeight(64))
	}

	// Demonstrate backward compatibility: templates without PreferredPixelSize still work
	template := HumanoidTemplate()
	headSpec := template.BodyPartLayout[PartHead]

	// Should calculate from relative dimensions (no PreferredPixelSize set)
	expectedWidth := int(float64(28) * headSpec.RelativeWidth)
	if headSpec.GetEffectiveWidth(28) != expectedWidth {
		t.Errorf("legacy head width = %d, want %d", headSpec.GetEffectiveWidth(28), expectedWidth)
	}

	// Can upgrade existing template with pixel dimensions
	upgradedHead := headSpec.WithPixelDimensions(4, 4)
	if upgradedHead.GetEffectiveWidth(28) != 4 {
		t.Errorf("upgraded head width = %d, want 4", upgradedHead.GetEffectiveWidth(28))
	}
	// Original template unchanged
	if headSpec.PreferredPixelSize != nil {
		t.Error("Original template should not have PreferredPixelSize set")
	}
}

// TestNewPartSpecFromPixels_Validation tests input validation for pixel dimensions.
func TestNewPartSpecFromPixels_Validation(t *testing.T) {
	tests := []struct {
		name           string
		inputWidth     int
		inputHeight    int
		expectedWidth  int
		expectedHeight int
	}{
		{"zero width", 0, 4, 1, 4},
		{"zero height", 4, 0, 4, 1},
		{"both zero", 0, 0, 1, 1},
		{"negative width", -5, 4, 1, 4},
		{"negative height", 4, -3, 4, 1},
		{"both negative", -2, -8, 1, 1},
		{"valid dimensions", 4, 6, 4, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := NewPartSpecFromPixels(tt.inputWidth, tt.inputHeight, shapes.ShapeCircle, 15, "secondary")

			if spec.PreferredPixelSize == nil {
				t.Fatal("PreferredPixelSize should not be nil")
			}
			if spec.PreferredPixelSize.Width != tt.expectedWidth {
				t.Errorf("Width = %d, want %d (input was %d)", spec.PreferredPixelSize.Width, tt.expectedWidth, tt.inputWidth)
			}
			if spec.PreferredPixelSize.Height != tt.expectedHeight {
				t.Errorf("Height = %d, want %d (input was %d)", spec.PreferredPixelSize.Height, tt.expectedHeight, tt.inputHeight)
			}
		})
	}
}

// TestPartSpec_WithPixelDimensions_Validation tests input validation for WithPixelDimensions.
func TestPartSpec_WithPixelDimensions_Validation(t *testing.T) {
	baseSpec := PartSpec{
		RelativeX:      0.5,
		RelativeY:      0.25,
		RelativeWidth:  0.35,
		RelativeHeight: 0.35,
		ShapeTypes:     []shapes.ShapeType{shapes.ShapeCircle},
		ZIndex:         15,
		ColorRole:      "secondary",
		Opacity:        1.0,
		Rotation:       0,
	}

	tests := []struct {
		name           string
		inputWidth     int
		inputHeight    int
		expectedWidth  int
		expectedHeight int
	}{
		{"zero width", 0, 4, 1, 4},
		{"zero height", 4, 0, 4, 1},
		{"both zero", 0, 0, 1, 1},
		{"negative width", -5, 4, 1, 4},
		{"negative height", 4, -3, 4, 1},
		{"both negative", -2, -8, 1, 1},
		{"valid dimensions", 4, 6, 4, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := baseSpec.WithPixelDimensions(tt.inputWidth, tt.inputHeight)

			if spec.PreferredPixelSize == nil {
				t.Fatal("PreferredPixelSize should not be nil")
			}
			if spec.PreferredPixelSize.Width != tt.expectedWidth {
				t.Errorf("Width = %d, want %d (input was %d)", spec.PreferredPixelSize.Width, tt.expectedWidth, tt.inputWidth)
			}
			if spec.PreferredPixelSize.Height != tt.expectedHeight {
				t.Errorf("Height = %d, want %d (input was %d)", spec.PreferredPixelSize.Height, tt.expectedHeight, tt.inputHeight)
			}

			// Verify other fields unchanged
			if spec.RelativeX != baseSpec.RelativeX {
				t.Errorf("RelativeX changed: got %f, want %f", spec.RelativeX, baseSpec.RelativeX)
			}
			if spec.ZIndex != baseSpec.ZIndex {
				t.Errorf("ZIndex changed: got %d, want %d", spec.ZIndex, baseSpec.ZIndex)
			}
		})
	}
}

// BenchmarkGetEffectiveWidth benchmarks effective width calculation.
func BenchmarkGetEffectiveWidth(b *testing.B) {
	spec := PartSpec{
		RelativeWidth: 0.5,
		PreferredPixelSize: &PixelDimensions{
			Width:  4,
			Height: 4,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = spec.GetEffectiveWidth(28)
	}
}

// BenchmarkGetEffectiveWidthNoPixelDimensions benchmarks effective width without pixel dimensions.
func BenchmarkGetEffectiveWidthNoPixelDimensions(b *testing.B) {
	spec := PartSpec{
		RelativeWidth:      0.5,
		PreferredPixelSize: nil,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = spec.GetEffectiveWidth(28)
	}
}

// BenchmarkNewPartSpecFromPixels benchmarks creating spec from pixels.
func BenchmarkNewPartSpecFromPixels(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = NewPartSpecFromPixels(4, 4, shapes.ShapeCircle, 15, "secondary")
	}
}
