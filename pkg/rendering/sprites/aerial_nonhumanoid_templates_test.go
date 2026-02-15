package sprites

import (
	"testing"
)

// TestAerialNonhumanoidTemplates_AllTypes verifies all nonhumanoid aerial templates
// produce valid body layouts with correct naming and required parts.
func TestAerialNonhumanoidTemplates_AllTypes(t *testing.T) {
	tests := []struct {
		name         string
		factory      func(Direction) AnatomicalTemplate
		direction    Direction
		wantName     string
		requiredPart BodyPart
	}{
		{"quadruped_up", QuadrupedAerialTemplate, DirUp, "quadruped_aerial_up", PartTorso},
		{"quadruped_down", QuadrupedAerialTemplate, DirDown, "quadruped_aerial_down", PartTail},
		{"quadruped_left", QuadrupedAerialTemplate, DirLeft, "quadruped_aerial_left", PartHead},
		{"quadruped_right", QuadrupedAerialTemplate, DirRight, "quadruped_aerial_right", PartLegs},
		{"serpentine_up", SerpentineAerialTemplate, DirUp, "serpentine_aerial_up", PartTail},
		{"serpentine_down", SerpentineAerialTemplate, DirDown, "serpentine_aerial_down", PartHead},
		{"serpentine_left", SerpentineAerialTemplate, DirLeft, "serpentine_aerial_left", PartTorso},
		{"serpentine_right", SerpentineAerialTemplate, DirRight, "serpentine_aerial_right", PartEyes},
		{"arachnid_up", ArachnidAerialTemplate, DirUp, "arachnid_aerial_up", PartArms},
		{"arachnid_down", ArachnidAerialTemplate, DirDown, "arachnid_aerial_down", PartLegs},
		{"arachnid_left", ArachnidAerialTemplate, DirLeft, "arachnid_aerial_left", PartEyes},
		{"arachnid_right", ArachnidAerialTemplate, DirRight, "arachnid_aerial_right", PartHead},
		{"blob_up", BlobAerialTemplate, DirUp, "blob_aerial_up", PartTorso},
		{"blob_down", BlobAerialTemplate, DirDown, "blob_aerial_down", PartHead},
		{"blob_left", BlobAerialTemplate, DirLeft, "blob_aerial_left", PartEyes},
		{"blob_right", BlobAerialTemplate, DirRight, "blob_aerial_right", PartShadow},
		{"flying_up", FlyingAerialTemplate, DirUp, "flying_aerial_up", PartWings},
		{"flying_down", FlyingAerialTemplate, DirDown, "flying_aerial_down", PartTail},
		{"flying_left", FlyingAerialTemplate, DirLeft, "flying_aerial_left", PartArms},
		{"flying_right", FlyingAerialTemplate, DirRight, "flying_aerial_right", PartHead},
		{"mechanical_up", MechanicalAerialTemplate, DirUp, "mechanical_aerial_up", PartEyes},
		{"mechanical_down", MechanicalAerialTemplate, DirDown, "mechanical_aerial_down", PartArms},
		{"mechanical_left", MechanicalAerialTemplate, DirLeft, "mechanical_aerial_left", PartLegs},
		{"mechanical_right", MechanicalAerialTemplate, DirRight, "mechanical_aerial_right", PartTorso},
		{"undead_up", UndeadAerialTemplate, DirUp, "undead_aerial_up", PartHead},
		{"undead_down", UndeadAerialTemplate, DirDown, "undead_aerial_down", PartTorso},
		{"undead_left", UndeadAerialTemplate, DirLeft, "undead_aerial_left", PartArms},
		{"undead_right", UndeadAerialTemplate, DirRight, "undead_aerial_right", PartShadow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := tt.factory(tt.direction)
			if tmpl.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", tmpl.Name, tt.wantName)
			}
			if len(tmpl.BodyPartLayout) == 0 {
				t.Fatal("BodyPartLayout is empty")
			}
			if _, ok := tmpl.BodyPartLayout[tt.requiredPart]; !ok {
				t.Errorf("missing required part %v", tt.requiredPart)
			}
			// All templates must have a shadow
			if _, ok := tmpl.BodyPartLayout[PartShadow]; !ok {
				t.Error("missing shadow part")
			}
			// All templates must have a torso
			if _, ok := tmpl.BodyPartLayout[PartTorso]; !ok {
				t.Error("missing torso part")
			}
		})
	}
}

// TestAerialNonhumanoidTemplates_ZIndexOrdering ensures parts render in correct order.
func TestAerialNonhumanoidTemplates_ZIndexOrdering(t *testing.T) {
	factories := []struct {
		name    string
		factory func(Direction) AnatomicalTemplate
	}{
		{"quadruped", QuadrupedAerialTemplate},
		{"serpentine", SerpentineAerialTemplate},
		{"arachnid", ArachnidAerialTemplate},
		{"blob", BlobAerialTemplate},
		{"flying", FlyingAerialTemplate},
		{"mechanical", MechanicalAerialTemplate},
		{"undead", UndeadAerialTemplate},
	}

	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			tmpl := f.factory(DirUp)
			sorted := tmpl.GetSortedParts()
			for i := 1; i < len(sorted); i++ {
				if sorted[i].Spec.ZIndex < sorted[i-1].Spec.ZIndex {
					t.Errorf("ZIndex ordering broken: %v (z=%d) before %v (z=%d)",
						sorted[i-1].Part, sorted[i-1].Spec.ZIndex,
						sorted[i].Part, sorted[i].Spec.ZIndex)
				}
			}
			// Shadow must be first (ZIndex 0)
			if sorted[0].Part != PartShadow {
				t.Errorf("first sorted part should be shadow, got %v", sorted[0].Part)
			}
		})
	}
}

// TestAerialNonhumanoidTemplates_Proportions validates that torso dominates
// and parts fit within the sprite bounds.
func TestAerialNonhumanoidTemplates_Proportions(t *testing.T) {
	factories := []struct {
		name    string
		factory func(Direction) AnatomicalTemplate
	}{
		{"quadruped", QuadrupedAerialTemplate},
		{"serpentine", SerpentineAerialTemplate},
		{"arachnid", ArachnidAerialTemplate},
		{"blob", BlobAerialTemplate},
		{"flying", FlyingAerialTemplate},
		{"mechanical", MechanicalAerialTemplate},
		{"undead", UndeadAerialTemplate},
	}

	for _, f := range factories {
		for _, dir := range []Direction{DirUp, DirDown, DirLeft, DirRight} {
			t.Run(f.name+"_"+string(dir), func(t *testing.T) {
				tmpl := f.factory(dir)
				for part, spec := range tmpl.BodyPartLayout {
					// RelativeX/Y should be within [0, 1]
					if spec.RelativeX < 0 || spec.RelativeX > 1.0 {
						t.Errorf("part %v: RelativeX=%f out of range", part, spec.RelativeX)
					}
					if spec.RelativeY < -0.1 || spec.RelativeY > 1.1 {
						t.Errorf("part %v: RelativeY=%f out of range", part, spec.RelativeY)
					}
					// Opacity within [0, 1]
					if spec.Opacity < 0 || spec.Opacity > 1.0 {
						t.Errorf("part %v: Opacity=%f out of range", part, spec.Opacity)
					}
					// Must have at least one shape type
					if len(spec.ShapeTypes) == 0 {
						t.Errorf("part %v: no shape types", part)
					}
				}
				// Torso must be present and substantial
				torso := tmpl.BodyPartLayout[PartTorso]
				if torso.RelativeHeight < 0.20 {
					t.Errorf("torso height %f too small for creature template", torso.RelativeHeight)
				}
			})
		}
	}
}

// TestSelectNonhumanoidAerialTemplate verifies the routing function.
func TestSelectNonhumanoidAerialTemplate(t *testing.T) {
	tests := []struct {
		entityType string
		genre      string
		wantPrefix string
	}{
		{"wolf", "fantasy", "fantasy_quadruped_aerial"},
		{"spider", "horror", "horror_arachnid_aerial"},
		{"slime", "", "blob_aerial"},
		{"dragon", "scifi", "scifi_flying_aerial"},
		{"snake", "cyberpunk", "cyberpunk_serpentine_aerial"},
		{"robot", "postapoc", "postapoc_mechanical_aerial"},
		{"ghost", "", "undead_aerial"},
		{"unknown_creature", "", "quadruped_aerial"},
		{"bear", "fantasy", "fantasy_quadruped_aerial"},
		{"beetle", "horror", "horror_arachnid_aerial"},
		{"wyvern", "", "flying_aerial"},
		{"golem", "scifi", "scifi_mechanical_aerial"},
	}

	for _, tt := range tests {
		t.Run(tt.entityType+"_"+tt.genre, func(t *testing.T) {
			tmpl := SelectNonhumanoidAerialTemplate(tt.entityType, tt.genre, DirUp)
			if tmpl.Name == "" {
				t.Fatal("template name is empty")
			}
			if len(tmpl.BodyPartLayout) == 0 {
				t.Fatal("body part layout is empty")
			}
			// Name should contain the expected prefix
			found := false
			if len(tmpl.Name) >= len(tt.wantPrefix) {
				found = tmpl.Name[:len(tt.wantPrefix)] == tt.wantPrefix
			}
			if !found {
				t.Errorf("Name = %q, want prefix %q", tmpl.Name, tt.wantPrefix)
			}
		})
	}
}

// TestSelectAerialTemplate_NonhumanoidRouting verifies that SelectAerialTemplate
// now routes nonhumanoid types to aerial templates instead of falling back.
func TestSelectAerialTemplate_NonhumanoidRouting(t *testing.T) {
	tests := []struct {
		entityType  string
		wantContain string
	}{
		{"wolf", "aerial"},
		{"spider", "aerial"},
		{"slime", "aerial"},
		{"dragon", "aerial"},
		{"snake", "aerial"},
		{"robot", "aerial"},
		{"ghost", "aerial"},
		{"player", "aerial"},
		{"npc", "aerial"},
	}

	for _, tt := range tests {
		t.Run(tt.entityType, func(t *testing.T) {
			tmpl := SelectAerialTemplate(tt.entityType, "fantasy", DirUp)
			found := false
			for i := 0; i <= len(tmpl.Name)-len(tt.wantContain); i++ {
				if tmpl.Name[i:i+len(tt.wantContain)] == tt.wantContain {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("SelectAerialTemplate(%q) = %q, should contain %q",
					tt.entityType, tmpl.Name, tt.wantContain)
			}
		})
	}
}

// TestAerialNonhumanoidTemplates_DirectionDifference verifies that direction
// actually changes the template layout (head/tail positions differ).
func TestAerialNonhumanoidTemplates_DirectionDifference(t *testing.T) {
	factories := []struct {
		name    string
		factory func(Direction) AnatomicalTemplate
	}{
		{"quadruped", QuadrupedAerialTemplate},
		{"serpentine", SerpentineAerialTemplate},
		{"flying", FlyingAerialTemplate},
		{"mechanical", MechanicalAerialTemplate},
	}

	for _, f := range factories {
		t.Run(f.name, func(t *testing.T) {
			up := f.factory(DirUp)
			down := f.factory(DirDown)

			upHead := up.BodyPartLayout[PartHead]
			downHead := down.BodyPartLayout[PartHead]

			// Head position should differ between up and down
			if upHead.RelativeX == downHead.RelativeX && upHead.RelativeY == downHead.RelativeY {
				t.Error("head position identical for DirUp and DirDown — direction not applied")
			}
		})
	}
}

// TestAerialNonhumanoidTemplates_DistinctSilhouettes verifies each creature type
// has a meaningfully different body layout (not just renamed humanoid).
func TestAerialNonhumanoidTemplates_DistinctSilhouettes(t *testing.T) {
	templates := map[string]AnatomicalTemplate{
		"quadruped":  QuadrupedAerialTemplate(DirUp),
		"serpentine": SerpentineAerialTemplate(DirUp),
		"arachnid":   ArachnidAerialTemplate(DirUp),
		"blob":       BlobAerialTemplate(DirUp),
		"flying":     FlyingAerialTemplate(DirUp),
		"mechanical": MechanicalAerialTemplate(DirUp),
	}

	// Each pair should differ in at least part count or torso proportions
	names := make([]string, 0, len(templates))
	for k := range templates {
		names = append(names, k)
	}

	for i := 0; i < len(names); i++ {
		for j := i + 1; j < len(names); j++ {
			a := templates[names[i]]
			b := templates[names[j]]

			aTorso := a.BodyPartLayout[PartTorso]
			bTorso := b.BodyPartLayout[PartTorso]

			partCountDiff := len(a.BodyPartLayout) != len(b.BodyPartLayout)
			widthDiff := aTorso.RelativeWidth != bTorso.RelativeWidth
			heightDiff := aTorso.RelativeHeight != bTorso.RelativeHeight

			if !partCountDiff && !widthDiff && !heightDiff {
				t.Errorf("%s and %s have identical part count and torso proportions",
					names[i], names[j])
			}
		}
	}
}

// BenchmarkAerialNonhumanoidTemplates benchmarks template creation.
func BenchmarkAerialNonhumanoidTemplates(b *testing.B) {
	b.Run("QuadrupedAerial", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = QuadrupedAerialTemplate(DirUp)
		}
	})
	b.Run("SerpentineAerial", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = SerpentineAerialTemplate(DirUp)
		}
	})
	b.Run("ArachnidAerial", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = ArachnidAerialTemplate(DirUp)
		}
	})
	b.Run("BlobAerial", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = BlobAerialTemplate(DirUp)
		}
	})
	b.Run("FlyingAerial", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = FlyingAerialTemplate(DirUp)
		}
	})
	b.Run("MechanicalAerial", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = MechanicalAerialTemplate(DirUp)
		}
	})
	b.Run("SelectNonhumanoidAerial", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = SelectNonhumanoidAerialTemplate("spider", "horror", DirUp)
		}
	})
}
