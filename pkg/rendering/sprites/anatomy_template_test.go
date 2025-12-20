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

	// Verify Phase 45 proportions (head 12%, torso 40%, legs 48%)
	headSpec := template.BodyPartLayout[PartHead]
	if headSpec.PreferredPixelSize == nil {
		t.Error("Head should have PreferredPixelSize for Phase 45 proportions")
	} else if headSpec.PreferredPixelSize.Height > 6 {
		t.Errorf("Head height too large for 12%% proportion: %d, want <= 6", headSpec.PreferredPixelSize.Height)
	}

	legsSpec := template.BodyPartLayout[PartLegs]
	if legsSpec.PreferredPixelSize == nil {
		t.Error("Legs should have PreferredPixelSize for Phase 45 proportions")
	} else if legsSpec.PreferredPixelSize.Height < 10 {
		t.Errorf("Legs height too small for 48%% proportion: %d, want >= 10", legsSpec.PreferredPixelSize.Height)
	}

	torsoSpec := template.BodyPartLayout[PartTorso]
	if torsoSpec.PreferredPixelSize == nil {
		t.Error("Torso should have PreferredPixelSize for Phase 45 proportions")
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

	// Verify top-down orientation (rotation = 0 for top-down view)
	torsoRotation := template.BodyPartLayout[PartTorso].Rotation
	if torsoRotation != 0 {
		t.Errorf("Torso rotation = %v, want 0 (top-down view)", torsoRotation)
	}

	// Verify head is at top (low Y value) for top-down perspective
	headY := template.BodyPartLayout[PartHead].RelativeY
	if headY > 0.5 {
		t.Errorf("Head Y position too low: %v, want < 0.5 (top half for top-down)", headY)
	}

	// Verify shadow is at bottom
	shadowY := template.BodyPartLayout[PartShadow].RelativeY
	if shadowY < 0.8 {
		t.Errorf("Shadow Y position too high: %v, want > 0.8 (bottom)", shadowY)
	}
}

// TestBlobTemplate tests the blob template structure.
func TestBlobTemplate(t *testing.T) {
	template := BlobTemplate()

	if template.Name != "blob" {
		t.Errorf("Template name = %v, want 'blob'", template.Name)
	}

	// Blobs should have shadow, torso, and optional head for top-down view
	if len(template.BodyPartLayout) < 2 || len(template.BodyPartLayout) > 3 {
		t.Errorf("Blob has unexpected number of parts: %d, expected 2-3 (shadow + torso + optional head)", len(template.BodyPartLayout))
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

	// Verify shadow is at bottom for top-down perspective
	shadowY := template.BodyPartLayout[PartShadow].RelativeY
	if shadowY < 0.8 {
		t.Errorf("Shadow Y position too high: %v, want > 0.8 (bottom)", shadowY)
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

	// Verify head is at top (low Y value) for top-down perspective
	headY := template.BodyPartLayout[PartHead].RelativeY
	if headY > 0.5 {
		t.Errorf("Head Y position too low: %v, want < 0.5 (top half for top-down)", headY)
	}

	// Verify shadow is at bottom
	shadowY := template.BodyPartLayout[PartShadow].RelativeY
	if shadowY < 0.8 {
		t.Errorf("Shadow Y position too high: %v, want > 0.8 (bottom)", shadowY)
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

	// Verify head is at top (low Y value) for top-down perspective
	headY := template.BodyPartLayout[PartHead].RelativeY
	if headY > 0.5 {
		t.Errorf("Head Y position too low: %v, want < 0.5 (top half for top-down)", headY)
	}

	// Verify shadow is at bottom
	shadowY := template.BodyPartLayout[PartShadow].RelativeY
	if shadowY < 0.8 {
		t.Errorf("Shadow Y position too high: %v, want > 0.8 (bottom)", shadowY)
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
// Updated for Phase 45 top-down proportions: head 12%, torso 40%, legs 48%.
func TestTemplateProportions(t *testing.T) {
	template := HumanoidTemplate()

	// Check head proportions (Phase 45: ~10-15% of height)
	headHeight := template.BodyPartLayout[PartHead].RelativeHeight
	if headHeight < 0.08 || headHeight > 0.20 {
		t.Errorf("Head height proportion out of reasonable range: %v, want 0.08-0.20", headHeight)
	}

	// Check torso proportions (Phase 45: ~35-45% of height)
	torsoHeight := template.BodyPartLayout[PartTorso].RelativeHeight
	if torsoHeight < 0.30 || torsoHeight > 0.50 {
		t.Errorf("Torso height proportion out of reasonable range: %v, want 0.30-0.50", torsoHeight)
	}

	// Check legs proportions (Phase 45: ~40-55% of height)
	legsHeight := template.BodyPartLayout[PartLegs].RelativeHeight
	if legsHeight < 0.35 || legsHeight > 0.55 {
		t.Errorf("Legs height proportion out of reasonable range: %v, want 0.35-0.55", legsHeight)
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
				// Fantasy should have broader shoulders (8 pixels vs base 7 pixels)
				torsoSpec := template.BodyPartLayout[PartTorso]
				if torsoSpec.PreferredPixelSize == nil || torsoSpec.PreferredPixelSize.Width < 8 {
					t.Errorf("Fantasy aerial should have broad shoulders (torso width >= 8px), got %v", torsoSpec.PreferredPixelSize)
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
				// Phase 15.1: Horror uses enhanced template with pixel dimensions
				headSpec := template.BodyPartLayout[PartHead]
				if headSpec.PreferredPixelSize == nil {
					t.Error("Horror aerial should have pixel dimensions")
				} else {
					// Horror should have 5×5 pixel head (enhanced template)
					if headSpec.PreferredPixelSize.Height != 5 {
						t.Errorf("Horror aerial head height should be 5px, got %d", headSpec.PreferredPixelSize.Height)
					}
				}
				// Horror aesthetic achieved through skull shapes
				hasSkullShape := false
				for _, shape := range headSpec.ShapeTypes {
					if shape == shapes.ShapeSkull {
						hasSkullShape = true
						break
					}
				}
				if !hasSkullShape {
					t.Error("Horror aerial head should include skull shape")
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
		{"humanoid_unknown_genre", "player", "unknown", DirDown, "enhanced_humanoid_aerial_down", true}, // Phase 15.1: Uses enhanced template
		{"non_humanoid_blob", "blob", "fantasy", DirDown, "blob", false},
		{"non_humanoid_quadruped", "wolf", "scifi", DirUp, "quadruped", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := SelectAerialTemplate(tt.entityType, tt.genre, tt.direction)

			if template.Name != tt.expectedName {
				t.Errorf("Template name = %s, want %s", template.Name, tt.expectedName)
			}

			// Verify aerial templates have proper pixel dimensions
			if tt.shouldBeAerial {
				torsoSpec := template.BodyPartLayout[PartTorso]
				// Aerial torsos should have pixel dimensions (6-8 pixels for height)
				if torsoSpec.PreferredPixelSize == nil {
					t.Errorf("Aerial torso should have PreferredPixelSize")
				} else if torsoSpec.PreferredPixelSize.Height < 6 || torsoSpec.PreferredPixelSize.Height > 8 {
					t.Errorf("Aerial torso height should be 6-8px, got %d", torsoSpec.PreferredPixelSize.Height)
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
		{"base", EnhancedHumanoidAerialTemplate}, // Phase 15.1: Use enhanced template
		{"fantasy", FantasyHumanoidAerial},
		{"scifi", SciFiHumanoidAerial},
		{"horror", HorrorHumanoidAerial},
		{"cyberpunk", CyberpunkHumanoidAerial},
		{"postapoc", PostApocHumanoidAerial},
	}

	for _, g := range genres {
		t.Run(g.name, func(t *testing.T) {
			template := g.templateFunc(DirDown)

			// Head: 5×5 or 6×6 pixels (depending on genre)
			headSpec := template.BodyPartLayout[PartHead]
			if headSpec.PreferredPixelSize == nil {
				t.Errorf("%s head should have PreferredPixelSize", g.name)
			} else if headSpec.PreferredPixelSize.Width < 5 || headSpec.PreferredPixelSize.Width > 6 {
				t.Errorf("%s head width = %dpx, want 5-6px", g.name, headSpec.PreferredPixelSize.Width)
			}

			// Torso: 6-8 pixels height (depending on genre)
			torsoSpec := template.BodyPartLayout[PartTorso]
			if torsoSpec.PreferredPixelSize == nil {
				t.Errorf("%s torso should have PreferredPixelSize", g.name)
			} else if torsoSpec.PreferredPixelSize.Height < 6 || torsoSpec.PreferredPixelSize.Height > 8 {
				t.Errorf("%s torso height = %dpx, want 6-8px", g.name, torsoSpec.PreferredPixelSize.Height)
			}

			// Legs: 2 pixels height (minimal from aerial view)
			legsSpec := template.BodyPartLayout[PartLegs]
			if legsSpec.PreferredPixelSize == nil {
				t.Errorf("%s legs should have PreferredPixelSize", g.name)
			} else if legsSpec.PreferredPixelSize.Height != 2 {
				t.Errorf("%s legs height = %dpx, want 2px", g.name, legsSpec.PreferredPixelSize.Height)
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

	// OBSOLETE CODE REMOVED: Backward compatibility demonstration
	// Removed: Tests for templates without PreferredPixelSize
	// PRE-1.0: All templates use PreferredPixelSize - relative-only mode not supported
	// Post Phase 45: HumanoidTemplate now has PreferredPixelSize by default
	template := HumanoidTemplate()
	headSpec := template.BodyPartLayout[PartHead]

	// HumanoidTemplate now has PreferredPixelSize (Phase 45 update)
	if headSpec.PreferredPixelSize == nil {
		t.Error("HumanoidTemplate should have PreferredPixelSize set after Phase 45 update")
	}

	// Can create template with custom pixel dimensions
	upgradedHead := headSpec.WithPixelDimensions(4, 4)
	if upgradedHead.GetEffectiveWidth(28) != 4 {
		t.Errorf("upgraded head width = %d, want 4", upgradedHead.GetEffectiveWidth(28))
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

// ============================================================================
// Phase 15.1: Genre-Specific Anatomical Variations Tests
// ============================================================================

// TestApplyFantasyVariation tests fantasy genre variation application.
func TestApplyFantasyVariation(t *testing.T) {
	tests := []struct {
		name         string
		baseTemplate AnatomicalTemplate
		wantPrefix   string
	}{
		{"quadruped", QuadrupedTemplate(), "fantasy_quadruped"},
		{"blob", BlobTemplate(), "fantasy_blob"},
		{"mechanical", MechanicalTemplate(), "fantasy_mechanical"},
		{"flying", FlyingTemplate(), "fantasy_flying"},
		{"serpentine", SerpentineTemplate(), "fantasy_serpentine"},
		{"arachnid", ArachnidTemplate(), "fantasy_arachnid"},
		{"undead", UndeadTemplate(), "fantasy_undead"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyFantasyVariation(tt.baseTemplate)

			// Verify name has fantasy prefix
			if result.Name != tt.wantPrefix {
				t.Errorf("Name = %v, want %v", result.Name, tt.wantPrefix)
			}

			// Verify organic shapes are preferred
			for part, spec := range result.BodyPartLayout {
				if part == PartShadow {
					continue // Shadow unchanged
				}

				// Check that at least one organic shape is present
				hasOrganicShape := false
				organicShapes := []shapes.ShapeType{
					shapes.ShapeOrganic, shapes.ShapeBean, shapes.ShapeEllipse,
					shapes.ShapeCircle, shapes.ShapeCapsule, shapes.ShapeWave,
				}
				for _, shapeType := range spec.ShapeTypes {
					for _, organic := range organicShapes {
						if shapeType == organic {
							hasOrganicShape = true
							break
						}
					}
					if hasOrganicShape {
						break
					}
				}

				if !hasOrganicShape && len(spec.ShapeTypes) > 0 {
					// Some parts may use non-organic shapes if they're unique
					// This is acceptable as long as geometric shapes are avoided
					for _, shapeType := range spec.ShapeTypes {
						if shapeType == shapes.ShapeRectangle ||
							shapeType == shapes.ShapeHexagon ||
							shapeType == shapes.ShapeOctagon {
							t.Errorf("Part %v has geometric shape %v, should prefer organic",
								part.String(), shapeType.String())
						}
					}
				}
			}

			// Verify shadow remains unchanged
			if shadowSpec, hasShadow := result.BodyPartLayout[PartShadow]; hasShadow {
				baseSpec := tt.baseTemplate.BodyPartLayout[PartShadow]
				if len(shadowSpec.ShapeTypes) != len(baseSpec.ShapeTypes) {
					t.Error("Shadow shape types should not be modified")
				}
			}
		})
	}
}

// TestApplySciFiVariation tests sci-fi genre variation application.
func TestApplySciFiVariation(t *testing.T) {
	tests := []struct {
		name         string
		baseTemplate AnatomicalTemplate
		wantPrefix   string
	}{
		{"quadruped", QuadrupedTemplate(), "scifi_quadruped"},
		{"blob", BlobTemplate(), "scifi_blob"},
		{"mechanical", MechanicalTemplate(), "scifi_mechanical"},
		{"flying", FlyingTemplate(), "scifi_flying"},
		{"serpentine", SerpentineTemplate(), "scifi_serpentine"},
		{"arachnid", ArachnidTemplate(), "scifi_arachnid"},
		{"undead", UndeadTemplate(), "scifi_undead"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplySciFiVariation(tt.baseTemplate)

			// Verify name has scifi prefix
			if result.Name != tt.wantPrefix {
				t.Errorf("Name = %v, want %v", result.Name, tt.wantPrefix)
			}

			// Verify geometric shapes are preferred
			for part, spec := range result.BodyPartLayout {
				if part == PartShadow {
					continue // Shadow unchanged
				}

				// Check that geometric shapes are present or organic shapes avoided
				hasGeometricShape := false
				geometricShapes := []shapes.ShapeType{
					shapes.ShapeHexagon, shapes.ShapeOctagon, shapes.ShapeRectangle,
					shapes.ShapeTriangle, shapes.ShapeGear, shapes.ShapeCrystal,
				}
				for _, shapeType := range spec.ShapeTypes {
					for _, geometric := range geometricShapes {
						if shapeType == geometric {
							hasGeometricShape = true
							break
						}
					}
					if hasGeometricShape {
						break
					}
				}

				// Verify no organic shapes that should have been replaced
				for _, shapeType := range spec.ShapeTypes {
					if shapeType == shapes.ShapeOrganic || shapeType == shapes.ShapeBean {
						t.Errorf("Part %v still has organic shape %v, should be replaced with geometric",
							part.String(), shapeType.String())
					}
				}
			}
		})
	}
}

// TestApplyHorrorVariation tests horror genre variation application.
func TestApplyHorrorVariation(t *testing.T) {
	tests := []struct {
		name         string
		baseTemplate AnatomicalTemplate
		wantPrefix   string
	}{
		{"quadruped", QuadrupedTemplate(), "horror_quadruped"},
		{"blob", BlobTemplate(), "horror_blob"},
		{"mechanical", MechanicalTemplate(), "horror_mechanical"},
		{"flying", FlyingTemplate(), "horror_flying"},
		{"serpentine", SerpentineTemplate(), "horror_serpentine"},
		{"arachnid", ArachnidTemplate(), "horror_arachnid"},
		{"undead", UndeadTemplate(), "horror_undead"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyHorrorVariation(tt.baseTemplate)

			// Verify name has horror prefix
			if result.Name != tt.wantPrefix {
				t.Errorf("Name = %v, want %v", result.Name, tt.wantPrefix)
			}

			// Verify shadow opacity is reduced
			if shadowSpec, hasShadow := result.BodyPartLayout[PartShadow]; hasShadow {
				baseSpec := tt.baseTemplate.BodyPartLayout[PartShadow]
				if shadowSpec.Opacity >= baseSpec.Opacity {
					t.Errorf("Shadow opacity not reduced: got %v, want < %v",
						shadowSpec.Opacity, baseSpec.Opacity)
				}
			}

			// Verify head proportions are distorted (if head exists)
			if headSpec, hasHead := result.BodyPartLayout[PartHead]; hasHead {
				baseSpec := tt.baseTemplate.BodyPartLayout[PartHead]
				// Head should be elongated (taller, narrower)
				if headSpec.RelativeHeight <= baseSpec.RelativeHeight {
					t.Error("Head should be elongated (increased height)")
				}
				if headSpec.RelativeWidth >= baseSpec.RelativeWidth {
					t.Error("Head should be narrowed (decreased width)")
				}

				// Should prefer skull/organic shapes
				hasDistortedShape := false
				for _, shapeType := range headSpec.ShapeTypes {
					if shapeType == shapes.ShapeSkull || shapeType == shapes.ShapeOrganic {
						hasDistortedShape = true
						break
					}
				}
				if !hasDistortedShape && len(headSpec.ShapeTypes) > 0 {
					t.Error("Head should use skull or organic shapes for horror aesthetic")
				}
			}

			// Verify torso uses irregular shapes
			if torsoSpec, hasTorso := result.BodyPartLayout[PartTorso]; hasTorso {
				hasIrregular := false
				for _, shapeType := range torsoSpec.ShapeTypes {
					if shapeType == shapes.ShapeOrganic || shapeType == shapes.ShapeBean {
						hasIrregular = true
						break
					}
				}
				if !hasIrregular {
					t.Error("Torso should use irregular organic/bean shapes")
				}
			}

			// Verify limbs are elongated (if present)
			for _, part := range []BodyPart{PartLegs, PartArms} {
				if spec, exists := result.BodyPartLayout[part]; exists {
					baseSpec := tt.baseTemplate.BodyPartLayout[part]
					if spec.RelativeHeight <= baseSpec.RelativeHeight {
						t.Errorf("Part %v should be elongated (increased height)", part.String())
					}
				}
			}
		})
	}
}

// TestApplyCyberpunkVariation tests cyberpunk genre variation application.
func TestApplyCyberpunkVariation(t *testing.T) {
	tests := []struct {
		name         string
		baseTemplate AnatomicalTemplate
		wantPrefix   string
	}{
		{"quadruped", QuadrupedTemplate(), "cyberpunk_quadruped"},
		{"blob", BlobTemplate(), "cyberpunk_blob"},
		{"mechanical", MechanicalTemplate(), "cyberpunk_mechanical"},
		{"flying", FlyingTemplate(), "cyberpunk_flying"},
		{"serpentine", SerpentineTemplate(), "cyberpunk_serpentine"},
		{"arachnid", ArachnidTemplate(), "cyberpunk_arachnid"},
		{"undead", UndeadTemplate(), "cyberpunk_undead"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyCyberpunkVariation(tt.baseTemplate)

			// Verify name has cyberpunk prefix
			if result.Name != tt.wantPrefix {
				t.Errorf("Name = %v, want %v", result.Name, tt.wantPrefix)
			}

			// Verify angular/tech shapes are used
			for part, spec := range result.BodyPartLayout {
				if part == PartShadow {
					continue // Shadow unchanged
				}

				// No organic shapes should remain
				for _, shapeType := range spec.ShapeTypes {
					if shapeType == shapes.ShapeOrganic || shapeType == shapes.ShapeBean {
						t.Errorf("Part %v has organic shape %v, should be angular",
							part.String(), shapeType.String())
					}
				}

				// Head should use accent1 (tech glow)
				if part == PartHead && spec.ColorRole != "accent1" {
					t.Errorf("Head color role = %v, want accent1 for tech glow", spec.ColorRole)
				}
			}

			// Verify tech armor overlay is added (if torso exists)
			if _, hasTorso := tt.baseTemplate.BodyPartLayout[PartTorso]; hasTorso {
				armorSpec, hasArmor := result.BodyPartLayout[PartArmor]
				if !hasArmor {
					t.Error("Should have tech armor overlay when torso exists")
				} else {
					// Verify armor properties
					if armorSpec.ColorRole != "accent1" {
						t.Errorf("Armor color role = %v, want accent1", armorSpec.ColorRole)
					}
					if armorSpec.Opacity >= 0.5 {
						t.Errorf("Armor opacity too high: %v, want < 0.5 for glow effect",
							armorSpec.Opacity)
					}
					// Should use tech shapes
					hasTechShape := false
					for _, shapeType := range armorSpec.ShapeTypes {
						if shapeType == shapes.ShapeHexagon || shapeType == shapes.ShapeOctagon {
							hasTechShape = true
							break
						}
					}
					if !hasTechShape {
						t.Error("Armor should use hexagon/octagon tech shapes")
					}
				}
			}
		})
	}
}

// TestApplyPostApocVariation tests post-apocalyptic genre variation application.
func TestApplyPostApocVariation(t *testing.T) {
	tests := []struct {
		name         string
		baseTemplate AnatomicalTemplate
		wantPrefix   string
	}{
		{"quadruped", QuadrupedTemplate(), "postapoc_quadruped"},
		{"blob", BlobTemplate(), "postapoc_blob"},
		{"mechanical", MechanicalTemplate(), "postapoc_mechanical"},
		{"flying", FlyingTemplate(), "postapoc_flying"},
		{"serpentine", SerpentineTemplate(), "postapoc_serpentine"},
		{"arachnid", ArachnidTemplate(), "postapoc_arachnid"},
		{"undead", UndeadTemplate(), "postapoc_undead"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyPostApocVariation(tt.baseTemplate)

			// Verify name has postapoc prefix
			if result.Name != tt.wantPrefix {
				t.Errorf("Name = %v, want %v", result.Name, tt.wantPrefix)
			}

			// Verify rough/irregular shapes are preferred
			for part, spec := range result.BodyPartLayout {
				if part == PartShadow {
					continue // Shadow unchanged
				}

				// Check for rough shapes
				hasRoughShape := false
				roughShapes := []shapes.ShapeType{
					shapes.ShapeOrganic, shapes.ShapeRectangle,
					shapes.ShapeCapsule, shapes.ShapeBean,
				}
				for _, shapeType := range spec.ShapeTypes {
					for _, rough := range roughShapes {
						if shapeType == rough {
							hasRoughShape = true
							break
						}
					}
					if hasRoughShape {
						break
					}
				}

				// Verify smooth geometric shapes are avoided
				for _, shapeType := range spec.ShapeTypes {
					if shapeType == shapes.ShapeCircle && !hasRoughShape {
						t.Errorf("Part %v uses circle, should prefer organic shapes",
							part.String())
					}
				}
			}
		})
	}
}

// TestSelectTemplateWithGenre tests genre-aware template selection.
func TestSelectTemplateWithGenre(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		genre      string
		wantName   string
	}{
		// Fantasy variations
		{"fantasy_quadruped", "quadruped", "fantasy", "fantasy_quadruped"},
		{"fantasy_blob", "blob", "fantasy", "fantasy_blob"},
		{"fantasy_flying", "flying", "fantasy", "fantasy_flying"},

		// Sci-fi variations
		{"scifi_mechanical", "mechanical", "scifi", "scifi_mechanical"},
		{"scifi_arachnid", "spider", "scifi", "scifi_arachnid"},
		{"scifi_serpentine", "snake", "sci-fi", "scifi_serpentine"},

		// Horror variations
		{"horror_undead", "undead", "horror", "horror_undead"},
		{"horror_quadruped", "wolf", "horror", "horror_quadruped"},

		// Cyberpunk variations
		{"cyberpunk_mechanical", "robot", "cyberpunk", "cyberpunk_mechanical"},
		{"cyberpunk_flying", "dragon", "cyberpunk", "cyberpunk_flying"},

		// Post-apoc variations
		{"postapoc_blob", "slime", "postapoc", "postapoc_blob"},
		{"postapoc_arachnid", "insect", "post-apocalyptic", "postapoc_arachnid"},

		// No genre (base templates)
		{"base_quadruped", "quadruped", "", "quadruped"},
		{"base_blob", "blob", "unknown", "blob"},

		// Humanoid (returns default humanoid)
		{"humanoid", "humanoid", "fantasy", "humanoid"},
		{"player", "player", "scifi", "humanoid"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SelectTemplateWithGenre(tt.entityType, tt.genre)

			if result.Name != tt.wantName {
				t.Errorf("Template name = %v, want %v", result.Name, tt.wantName)
			}

			// Verify template has required body parts
			if len(result.BodyPartLayout) == 0 {
				t.Error("Template should have at least one body part")
			}

			// Verify shadow exists (all templates should have shadow)
			if _, hasShadow := result.BodyPartLayout[PartShadow]; !hasShadow {
				t.Error("Template should have shadow part")
			}
		})
	}
}

// TestGenreVariationDeterminism tests that genre variations are deterministic.
func TestGenreVariationDeterminism(t *testing.T) {
	baseTemplate := QuadrupedTemplate()

	tests := []struct {
		name      string
		applyFunc func(AnatomicalTemplate) AnatomicalTemplate
	}{
		{"fantasy", ApplyFantasyVariation},
		{"scifi", ApplySciFiVariation},
		{"horror", ApplyHorrorVariation},
		{"cyberpunk", ApplyCyberpunkVariation},
		{"postapoc", ApplyPostApocVariation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Apply variation twice
			result1 := tt.applyFunc(baseTemplate)
			result2 := tt.applyFunc(baseTemplate)

			// Verify same name
			if result1.Name != result2.Name {
				t.Errorf("Names differ: %v vs %v", result1.Name, result2.Name)
			}

			// Verify same number of parts
			if len(result1.BodyPartLayout) != len(result2.BodyPartLayout) {
				t.Errorf("Different part counts: %d vs %d",
					len(result1.BodyPartLayout), len(result2.BodyPartLayout))
			}

			// Verify each part is identical
			for part, spec1 := range result1.BodyPartLayout {
				spec2, exists := result2.BodyPartLayout[part]
				if !exists {
					t.Errorf("Part %v missing in second result", part.String())
					continue
				}

				// Compare key fields
				if spec1.RelativeX != spec2.RelativeX {
					t.Errorf("Part %v: RelativeX differs", part.String())
				}
				if spec1.RelativeY != spec2.RelativeY {
					t.Errorf("Part %v: RelativeY differs", part.String())
				}
				if spec1.RelativeWidth != spec2.RelativeWidth {
					t.Errorf("Part %v: RelativeWidth differs", part.String())
				}
				if spec1.RelativeHeight != spec2.RelativeHeight {
					t.Errorf("Part %v: RelativeHeight differs", part.String())
				}
				if spec1.ZIndex != spec2.ZIndex {
					t.Errorf("Part %v: ZIndex differs", part.String())
				}
				if spec1.Opacity != spec2.Opacity {
					t.Errorf("Part %v: Opacity differs", part.String())
				}
				if len(spec1.ShapeTypes) != len(spec2.ShapeTypes) {
					t.Errorf("Part %v: ShapeTypes length differs", part.String())
				}
			}
		})
	}
}

// TestGenreVariationShapePreservation tests that variations maintain valid shapes.
func TestGenreVariationShapePreservation(t *testing.T) {
	templates := []AnatomicalTemplate{
		QuadrupedTemplate(),
		BlobTemplate(),
		MechanicalTemplate(),
		FlyingTemplate(),
		SerpentineTemplate(),
		ArachnidTemplate(),
		UndeadTemplate(),
	}

	variations := []struct {
		name      string
		applyFunc func(AnatomicalTemplate) AnatomicalTemplate
	}{
		{"fantasy", ApplyFantasyVariation},
		{"scifi", ApplySciFiVariation},
		{"horror", ApplyHorrorVariation},
		{"cyberpunk", ApplyCyberpunkVariation},
		{"postapoc", ApplyPostApocVariation},
	}

	for _, template := range templates {
		for _, variation := range variations {
			t.Run(template.Name+"_"+variation.name, func(t *testing.T) {
				result := variation.applyFunc(template)

				// Verify all parts still have at least one shape type
				for part, spec := range result.BodyPartLayout {
					if len(spec.ShapeTypes) == 0 {
						t.Errorf("Part %v has no shape types after variation", part.String())
					}

					// Verify shape types are valid (not default zero value)
					for i, shapeType := range spec.ShapeTypes {
						if shapeType.String() == "unknown" {
							t.Errorf("Part %v has invalid shape type at index %d",
								part.String(), i)
						}
					}
				}

				// Verify Z-index ordering is maintained (shadow lowest)
				if shadowSpec, hasShadow := result.BodyPartLayout[PartShadow]; hasShadow {
					for part, spec := range result.BodyPartLayout {
						if part != PartShadow && spec.ZIndex <= shadowSpec.ZIndex {
							t.Errorf("Part %v has Z-index %d <= shadow Z-index %d",
								part.String(), spec.ZIndex, shadowSpec.ZIndex)
						}
					}
				}

				// Verify opacity is valid (0.0-1.0)
				for part, spec := range result.BodyPartLayout {
					if spec.Opacity < 0.0 || spec.Opacity > 1.0 {
						t.Errorf("Part %v has invalid opacity: %v (must be 0.0-1.0)",
							part.String(), spec.Opacity)
					}
				}
			})
		}
	}
}

// BenchmarkApplyFantasyVariation benchmarks fantasy variation application.
func BenchmarkApplyFantasyVariation(b *testing.B) {
	template := QuadrupedTemplate()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyFantasyVariation(template)
	}
}

// BenchmarkApplySciFiVariation benchmarks sci-fi variation application.
func BenchmarkApplySciFiVariation(b *testing.B) {
	template := MechanicalTemplate()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplySciFiVariation(template)
	}
}

// BenchmarkApplyHorrorVariation benchmarks horror variation application.
func BenchmarkApplyHorrorVariation(b *testing.B) {
	template := UndeadTemplate()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ApplyHorrorVariation(template)
	}
}

// BenchmarkSelectTemplateWithGenre benchmarks genre-aware template selection.
func BenchmarkSelectTemplateWithGenre(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SelectTemplateWithGenre("quadruped", "fantasy")
	}
}
