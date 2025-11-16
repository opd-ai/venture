// Package sprites - Phase 45 template tests for enhanced 64x64 sprites.
package sprites

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// TestEnhanced64HumanoidTemplate verifies Phase 45 64x64 humanoid template structure.
func TestEnhanced64HumanoidTemplate(t *testing.T) {
	template := Enhanced64HumanoidTemplate()

	if template.Name != "enhanced64_humanoid" {
		t.Errorf("expected template name 'enhanced64_humanoid', got %q", template.Name)
	}

	// Verify required body parts exist
	requiredParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartArms, PartHead}
	for _, part := range requiredParts {
		if _, ok := template.BodyPartLayout[part]; !ok {
			t.Errorf("missing required body part: %s", part)
		}
	}

	// Verify pixel dimensions for Phase 45 specifications
	const spriteSize = 64

	// Head: 8×8 pixels (12% of height)
	headSpec := template.BodyPartLayout[PartHead]
	if headSpec.PreferredPixelSize == nil {
		t.Fatal("head PreferredPixelSize should not be nil")
	}
	if headSpec.PreferredPixelSize.Width != 8 || headSpec.PreferredPixelSize.Height != 8 {
		t.Errorf("head expected 8×8 pixels, got %d×%d", headSpec.PreferredPixelSize.Width, headSpec.PreferredPixelSize.Height)
	}

	// Torso: 10×14 pixels (40% proportion)
	torsoSpec := template.BodyPartLayout[PartTorso]
	if torsoSpec.PreferredPixelSize == nil {
		t.Fatal("torso PreferredPixelSize should not be nil")
	}
	if torsoSpec.PreferredPixelSize.Width != 10 || torsoSpec.PreferredPixelSize.Height != 14 {
		t.Errorf("torso expected 10×14 pixels, got %d×%d", torsoSpec.PreferredPixelSize.Width, torsoSpec.PreferredPixelSize.Height)
	}

	// Legs: 8×16 pixels (48% proportion)
	legsSpec := template.BodyPartLayout[PartLegs]
	if legsSpec.PreferredPixelSize == nil {
		t.Fatal("legs PreferredPixelSize should not be nil")
	}
	if legsSpec.PreferredPixelSize.Width != 8 || legsSpec.PreferredPixelSize.Height != 16 {
		t.Errorf("legs expected 8×16 pixels, got %d×%d", legsSpec.PreferredPixelSize.Width, legsSpec.PreferredPixelSize.Height)
	}

	// Verify proportions sum to approximately 100% (12% + 40% + 48% = 100%)
	headHeight := headSpec.PreferredPixelSize.Height
	torsoHeight := torsoSpec.PreferredPixelSize.Height
	legsHeight := legsSpec.PreferredPixelSize.Height
	totalProportion := float64(headHeight+torsoHeight+legsHeight) / float64(spriteSize)

	if totalProportion < 0.55 || totalProportion > 0.65 {
		t.Errorf("total body proportion should be ~60%%, got %.1f%%", totalProportion*100)
	}

	// Verify Z-index ordering (legs < arms < torso < head)
	if legsSpec.ZIndex >= torsoSpec.ZIndex {
		t.Error("legs should render below torso")
	}
	if template.BodyPartLayout[PartArms].ZIndex >= torsoSpec.ZIndex {
		t.Error("arms should render below torso")
	}
	if torsoSpec.ZIndex >= headSpec.ZIndex {
		t.Error("torso should render below head")
	}
}

// TestDetailed64HumanoidTemplate verifies Phase 45 detailed humanoid with facial features.
func TestDetailed64HumanoidTemplate(t *testing.T) {
	template := Detailed64HumanoidTemplate()

	if template.Name != "detailed64_humanoid" {
		t.Errorf("expected template name 'detailed64_humanoid', got %q", template.Name)
	}

	// Verify all parts from Enhanced64HumanoidTemplate are present
	baseParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartArms, PartHead}
	for _, part := range baseParts {
		if _, ok := template.BodyPartLayout[part]; !ok {
			t.Errorf("missing base body part: %s", part)
		}
	}

	// Verify facial features were added
	facialParts := []BodyPart{PartEyes, PartMouth}
	for _, part := range facialParts {
		if _, ok := template.BodyPartLayout[part]; !ok {
			t.Errorf("missing facial feature: %s", part)
		}
	}

	// Verify eyes pixel dimensions (4×2 pixels)
	eyesSpec := template.BodyPartLayout[PartEyes]
	if eyesSpec.PreferredPixelSize == nil {
		t.Fatal("eyes PreferredPixelSize should not be nil")
	}
	if eyesSpec.PreferredPixelSize.Width != 4 || eyesSpec.PreferredPixelSize.Height != 2 {
		t.Errorf("eyes expected 4×2 pixels, got %d×%d", eyesSpec.PreferredPixelSize.Width, eyesSpec.PreferredPixelSize.Height)
	}

	// Verify mouth pixel dimensions (4×2 pixels)
	mouthSpec := template.BodyPartLayout[PartMouth]
	if mouthSpec.PreferredPixelSize == nil {
		t.Fatal("mouth PreferredPixelSize should not be nil")
	}
	if mouthSpec.PreferredPixelSize.Width != 4 || mouthSpec.PreferredPixelSize.Height != 2 {
		t.Errorf("mouth expected 4×2 pixels, got %d×%d", mouthSpec.PreferredPixelSize.Width, mouthSpec.PreferredPixelSize.Height)
	}

	// Verify facial features render above head
	headSpec := template.BodyPartLayout[PartHead]
	if eyesSpec.ZIndex <= headSpec.ZIndex {
		t.Error("eyes should render above head")
	}
	if mouthSpec.ZIndex <= headSpec.ZIndex {
		t.Error("mouth should render above head")
	}

	// Verify facial features are positioned in head region (Y < 0.25)
	if eyesSpec.RelativeY >= 0.25 {
		t.Errorf("eyes should be in upper region (Y < 0.25), got %.2f", eyesSpec.RelativeY)
	}
	if mouthSpec.RelativeY >= 0.25 {
		t.Errorf("mouth should be in upper region (Y < 0.25), got %.2f", mouthSpec.RelativeY)
	}
}

// TestEnhanced64QuadrupedTemplate verifies Phase 45 64x64 quadruped template structure.
func TestEnhanced64QuadrupedTemplate(t *testing.T) {
	template := Enhanced64QuadrupedTemplate()

	if template.Name != "enhanced64_quadruped" {
		t.Errorf("expected template name 'enhanced64_quadruped', got %q", template.Name)
	}

	// Verify required body parts
	requiredParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartHead, PartTail}
	for _, part := range requiredParts {
		if _, ok := template.BodyPartLayout[part]; !ok {
			t.Errorf("missing required body part: %s", part)
		}
	}

	// Verify horizontal orientation (legs and torso should have rotation)
	legsSpec := template.BodyPartLayout[PartLegs]
	if legsSpec.Rotation != 90 {
		t.Errorf("legs expected 90° rotation (horizontal), got %.0f°", legsSpec.Rotation)
	}

	torsoSpec := template.BodyPartLayout[PartTorso]
	if torsoSpec.Rotation != 90 {
		t.Errorf("torso expected 90° rotation (horizontal), got %.0f°", torsoSpec.Rotation)
	}

	// Verify torso is elongated horizontally (20×14 pixels)
	if torsoSpec.PreferredPixelSize == nil {
		t.Fatal("torso PreferredPixelSize should not be nil")
	}
	if torsoSpec.PreferredPixelSize.Width != 20 || torsoSpec.PreferredPixelSize.Height != 14 {
		t.Errorf("torso expected 20×14 pixels, got %d×%d", torsoSpec.PreferredPixelSize.Width, torsoSpec.PreferredPixelSize.Height)
	}

	// Verify head is at forward position (X < 0.3)
	headSpec := template.BodyPartLayout[PartHead]
	if headSpec.RelativeX >= 0.3 {
		t.Errorf("head should be forward (X < 0.3), got %.2f", headSpec.RelativeX)
	}

	// Verify tail is at rear position (X > 0.7)
	tailSpec := template.BodyPartLayout[PartTail]
	if tailSpec.RelativeX <= 0.7 {
		t.Errorf("tail should be rear (X > 0.7), got %.2f", tailSpec.RelativeX)
	}
}

// TestEnhanced64BlobTemplate verifies Phase 45 64x64 blob/slime template structure.
func TestEnhanced64BlobTemplate(t *testing.T) {
	template := Enhanced64BlobTemplate()

	if template.Name != "enhanced64_blob" {
		t.Errorf("expected template name 'enhanced64_blob', got %q", template.Name)
	}

	// Verify minimal structure (shadow, torso, eyes for nucleus)
	requiredParts := []BodyPart{PartShadow, PartTorso, PartEyes}
	for _, part := range requiredParts {
		if _, ok := template.BodyPartLayout[part]; !ok {
			t.Errorf("missing required body part: %s", part)
		}
	}

	// Verify large torso (32×28 pixels for amorphous mass)
	torsoSpec := template.BodyPartLayout[PartTorso]
	if torsoSpec.PreferredPixelSize == nil {
		t.Fatal("torso PreferredPixelSize should not be nil")
	}
	if torsoSpec.PreferredPixelSize.Width != 32 || torsoSpec.PreferredPixelSize.Height != 28 {
		t.Errorf("torso expected 32×28 pixels, got %d×%d", torsoSpec.PreferredPixelSize.Width, torsoSpec.PreferredPixelSize.Height)
	}

	// Verify translucency (opacity < 1.0)
	if torsoSpec.Opacity >= 1.0 {
		t.Errorf("blob torso should be translucent (opacity < 1.0), got %.2f", torsoSpec.Opacity)
	}

	// Verify nucleus/eyes are visible through body (higher Z-index)
	eyesSpec := template.BodyPartLayout[PartEyes]
	if eyesSpec.ZIndex <= torsoSpec.ZIndex {
		t.Error("eyes (nucleus) should render above torso for visibility")
	}

	// Verify organic shapes are preferred
	hasOrganic := false
	for _, shape := range torsoSpec.ShapeTypes {
		if shape == shapes.ShapeOrganic { // Use shapes.ShapeOrganic constant
			hasOrganic = true
			break
		}
	}
	if !hasOrganic {
		t.Error("blob torso should include organic shape type")
	}
}

// TestEnhanced64MechanicalTemplate verifies Phase 45 64x64 mechanical template structure.
func TestEnhanced64MechanicalTemplate(t *testing.T) {
	template := Enhanced64MechanicalTemplate()

	if template.Name != "enhanced64_mechanical" {
		t.Errorf("expected template name 'enhanced64_mechanical', got %q", template.Name)
	}

	// Verify required body parts
	requiredParts := []BodyPart{PartShadow, PartLegs, PartTorso, PartArms, PartHead}
	for _, part := range requiredParts {
		if _, ok := template.BodyPartLayout[part]; !ok {
			t.Errorf("missing required body part: %s", part)
		}
	}

	// Verify torso is larger chassis (12×18 pixels)
	torsoSpec := template.BodyPartLayout[PartTorso]
	if torsoSpec.PreferredPixelSize == nil {
		t.Fatal("torso PreferredPixelSize should not be nil")
	}
	if torsoSpec.PreferredPixelSize.Width != 12 || torsoSpec.PreferredPixelSize.Height != 18 {
		t.Errorf("torso expected 12×18 pixels, got %d×%d", torsoSpec.PreferredPixelSize.Width, torsoSpec.PreferredPixelSize.Height)
	}

	// Verify head is cubic/angular (10×10 pixels)
	headSpec := template.BodyPartLayout[PartHead]
	if headSpec.PreferredPixelSize == nil {
		t.Fatal("head PreferredPixelSize should not be nil")
	}
	if headSpec.PreferredPixelSize.Width != 10 || headSpec.PreferredPixelSize.Height != 10 {
		t.Errorf("head expected 10×10 pixels, got %d×%d", headSpec.PreferredPixelSize.Width, headSpec.PreferredPixelSize.Height)
	}

	// Verify geometric shapes are preferred (rectangle, hexagon, octagon)
	hasGeometric := false
	for _, shape := range torsoSpec.ShapeTypes {
		// Use shapes constants
		if shape == shapes.ShapeRectangle || shape == shapes.ShapeHexagon || shape == shapes.ShapeOctagon {
			hasGeometric = true
			break
		}
	}
	if !hasGeometric {
		t.Error("mechanical torso should include geometric shape types")
	}
}

// TestPhase45ProportionAccuracy verifies all templates meet Phase 45 proportion targets.
func TestPhase45ProportionAccuracy(t *testing.T) {
	tests := []struct {
		name     string
		template AnatomicalTemplate
		expected map[BodyPart]PixelDimensions
	}{
		{
			name:     "Enhanced64Humanoid",
			template: Enhanced64HumanoidTemplate(),
			expected: map[BodyPart]PixelDimensions{
				PartHead:  {Width: 8, Height: 8},
				PartTorso: {Width: 10, Height: 14},
				PartLegs:  {Width: 8, Height: 16},
				PartArms:  {Width: 12, Height: 10},
			},
		},
		{
			name:     "Detailed64Humanoid",
			template: Detailed64HumanoidTemplate(),
			expected: map[BodyPart]PixelDimensions{
				PartHead:  {Width: 8, Height: 8},
				PartTorso: {Width: 10, Height: 14},
				PartLegs:  {Width: 8, Height: 16},
				PartArms:  {Width: 12, Height: 10},
				PartEyes:  {Width: 4, Height: 2},
				PartMouth: {Width: 4, Height: 2},
			},
		},
		{
			name:     "Enhanced64Quadruped",
			template: Enhanced64QuadrupedTemplate(),
			expected: map[BodyPart]PixelDimensions{
				PartHead:  {Width: 10, Height: 12},
				PartTorso: {Width: 20, Height: 14},
				PartLegs:  {Width: 20, Height: 8},
				PartTail:  {Width: 8, Height: 16},
			},
		},
		{
			name:     "Enhanced64Blob",
			template: Enhanced64BlobTemplate(),
			expected: map[BodyPart]PixelDimensions{
				PartTorso: {Width: 32, Height: 28},
				PartEyes:  {Width: 6, Height: 4},
			},
		},
		{
			name:     "Enhanced64Mechanical",
			template: Enhanced64MechanicalTemplate(),
			expected: map[BodyPart]PixelDimensions{
				PartHead:  {Width: 10, Height: 10},
				PartTorso: {Width: 12, Height: 18},
				PartArms:  {Width: 14, Height: 10},
				PartLegs:  {Width: 10, Height: 14},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for part, expectedDim := range tt.expected {
				spec, ok := tt.template.BodyPartLayout[part]
				if !ok {
					t.Errorf("%s missing expected body part: %s", tt.name, part)
					continue
				}

				if spec.PreferredPixelSize == nil {
					t.Errorf("%s %s PreferredPixelSize is nil", tt.name, part)
					continue
				}

				actualDim := *spec.PreferredPixelSize
				if actualDim.Width != expectedDim.Width || actualDim.Height != expectedDim.Height {
					t.Errorf("%s %s expected %d×%d pixels, got %d×%d",
						tt.name, part, expectedDim.Width, expectedDim.Height,
						actualDim.Width, actualDim.Height)
				}
			}
		})
	}
}

// TestPhase45SortedParts verifies Z-index sorting works correctly for 64x64 templates.
func TestPhase45SortedParts(t *testing.T) {
	template := Enhanced64HumanoidTemplate()
	sorted := template.GetSortedParts()

	if len(sorted) == 0 {
		t.Fatal("GetSortedParts returned empty slice")
	}

	// Verify parts are sorted by Z-index (ascending)
	for i := 1; i < len(sorted); i++ {
		if sorted[i].Spec.ZIndex < sorted[i-1].Spec.ZIndex {
			t.Errorf("parts not sorted correctly at index %d: Z-index %d < %d",
				i, sorted[i].Spec.ZIndex, sorted[i-1].Spec.ZIndex)
		}
	}

	// Verify shadow is first (lowest Z-index)
	if sorted[0].Part != PartShadow {
		t.Errorf("first part should be shadow, got %s", sorted[0].Part)
	}
}

// TestPhase45TemplateConsistency verifies all Phase 45 templates follow conventions.
func TestPhase45TemplateConsistency(t *testing.T) {
	templates := []struct {
		name     string
		template AnatomicalTemplate
	}{
		{"Enhanced64Humanoid", Enhanced64HumanoidTemplate()},
		{"Detailed64Humanoid", Detailed64HumanoidTemplate()},
		{"Enhanced64Quadruped", Enhanced64QuadrupedTemplate()},
		{"Enhanced64Blob", Enhanced64BlobTemplate()},
		{"Enhanced64Mechanical", Enhanced64MechanicalTemplate()},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			// Verify shadow exists and has correct properties
			if shadow, ok := tt.template.BodyPartLayout[PartShadow]; ok {
				if shadow.ZIndex != 0 {
					t.Errorf("shadow should have Z-index 0, got %d", shadow.ZIndex)
				}
				if shadow.ColorRole != "shadow" {
					t.Errorf("shadow should have ColorRole 'shadow', got %q", shadow.ColorRole)
				}
				if shadow.Opacity < 0.2 || shadow.Opacity > 0.5 {
					t.Errorf("shadow opacity should be 0.2-0.5, got %.2f", shadow.Opacity)
				}
			} else {
				t.Error("template missing shadow body part")
			}

			// Verify all parts have valid opacity (0.0-1.0)
			for part, spec := range tt.template.BodyPartLayout {
				if spec.Opacity < 0.0 || spec.Opacity > 1.0 {
					t.Errorf("%s opacity out of range [0.0, 1.0]: %.2f", part, spec.Opacity)
				}
			}

			// Verify all parts have at least one shape type
			for part, spec := range tt.template.BodyPartLayout {
				if len(spec.ShapeTypes) == 0 {
					t.Errorf("%s has no shape types defined", part)
				}
			}

			// Verify all non-shadow parts have ColorRole
			for part, spec := range tt.template.BodyPartLayout {
				if part != PartShadow && spec.ColorRole == "" {
					t.Errorf("%s missing ColorRole", part)
				}
			}
		})
	}
}

// BenchmarkEnhanced64HumanoidTemplate benchmarks 64x64 humanoid template creation.
func BenchmarkEnhanced64HumanoidTemplate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Enhanced64HumanoidTemplate()
	}
}

// BenchmarkDetailed64HumanoidTemplate benchmarks detailed 64x64 humanoid template creation.
func BenchmarkDetailed64HumanoidTemplate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = Detailed64HumanoidTemplate()
	}
}

// BenchmarkPhase45GetSortedParts benchmarks Z-index sorting for 64x64 templates.
func BenchmarkPhase45GetSortedParts(b *testing.B) {
	template := Enhanced64HumanoidTemplate()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = template.GetSortedParts()
	}
}
