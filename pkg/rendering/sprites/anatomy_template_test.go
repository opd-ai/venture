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
