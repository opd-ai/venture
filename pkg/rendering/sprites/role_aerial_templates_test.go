package sprites

import (
	"testing"
)

// TestMapEntityTypeToRole verifies mapping from entity type strings to visual roles.
func TestMapEntityTypeToRole(t *testing.T) {
	tests := []struct {
		name       string
		entityType string
		wantRole   VisualRole
	}{
		{"mage", "mage", RoleMage},
		{"elementalist→mage", "elementalist", RoleMage},
		{"necromancer→mage", "necromancer", RoleMage},
		{"enchanter→mage", "enchanter", RoleMage},
		{"warrior", "warrior", RoleWarrior},
		{"berserker→warrior", "berserker", RoleWarrior},
		{"knight", "knight", RoleKnight},
		{"paladin→knight", "paladin", RoleKnight},
		{"rogue", "rogue", RoleRogue},
		{"assassin→rogue", "assassin", RoleRogue},
		{"ninja→rogue", "ninja", RoleRogue},
		{"merchant", "merchant", RoleMerchant},
		{"ranger", "ranger", RoleRanger},
		{"cleric→priest", "cleric", RolePriest},
		{"druid→priest", "druid", RolePriest},
		{"bard→priest", "bard", RolePriest},
		{"priest", "priest", RolePriest},
		{"unknown", "zombie", ""},
		{"humanoid", "humanoid", ""},
		{"player", "player", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapEntityTypeToRole(tt.entityType)
			if got != tt.wantRole {
				t.Errorf("MapEntityTypeToRole(%q) = %q, want %q", tt.entityType, got, tt.wantRole)
			}
		})
	}
}

// TestSelectRoleAerialTemplate verifies each role produces a named template.
func TestSelectRoleAerialTemplate(t *testing.T) {
	roles := []VisualRole{RoleMage, RoleWarrior, RoleKnight, RoleRogue, RoleMerchant, RoleRanger, RolePriest}
	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}

	for _, role := range roles {
		for _, dir := range directions {
			t.Run(string(role)+"_"+string(dir), func(t *testing.T) {
				tmpl := SelectRoleAerialTemplate(role, dir)
				if tmpl.Name == "" {
					t.Error("template name is empty")
				}
				if len(tmpl.BodyPartLayout) < 4 {
					t.Errorf("template has only %d parts, want at least 4", len(tmpl.BodyPartLayout))
				}
			})
		}
	}
}

// TestSelectRoleAerialTemplate_Fallback verifies unknown role falls back gracefully.
func TestSelectRoleAerialTemplate_Fallback(t *testing.T) {
	tmpl := SelectRoleAerialTemplate("unknown", DirDown)
	if tmpl.Name == "" {
		t.Error("fallback template should have a name")
	}
	if len(tmpl.BodyPartLayout) < 4 {
		t.Errorf("fallback template should have at least 4 parts, got %d", len(tmpl.BodyPartLayout))
	}
}

// TestRoleTemplateDistinctSilhouettes verifies role templates have meaningfully
// different torso proportions so silhouettes are distinguishable.
func TestRoleTemplateDistinctSilhouettes(t *testing.T) {
	dir := DirDown
	templates := map[VisualRole]AnatomicalTemplate{
		RoleMage:     MageAerialTemplate(dir),
		RoleWarrior:  WarriorAerialTemplate(dir),
		RoleKnight:   KnightAerialTemplate(dir),
		RoleRogue:    RogueAerialTemplate(dir),
		RoleMerchant: MerchantAerialTemplate(dir),
		RoleRanger:   RangerAerialTemplate(dir),
		RolePriest:   PriestAerialTemplate(dir),
	}

	type key struct{ a, b VisualRole }
	checked := map[key]bool{}

	for roleA, tmplA := range templates {
		for roleB, tmplB := range templates {
			if roleA == roleB {
				continue
			}
			k := key{roleA, roleB}
			kr := key{roleB, roleA}
			if checked[k] || checked[kr] {
				continue
			}
			checked[k] = true

			torsoA := tmplA.BodyPartLayout[PartTorso]
			torsoB := tmplB.BodyPartLayout[PartTorso]

			// At least one dimension should differ by >5%
			wDiff := roleTestAbs64(torsoA.RelativeWidth - torsoB.RelativeWidth)
			hDiff := roleTestAbs64(torsoA.RelativeHeight - torsoB.RelativeHeight)
			if wDiff < 0.02 && hDiff < 0.02 {
				t.Errorf("roles %s and %s have nearly identical torso proportions (w:%.3f/%.3f h:%.3f/%.3f)",
					roleA, roleB, torsoA.RelativeWidth, torsoB.RelativeWidth,
					torsoA.RelativeHeight, torsoB.RelativeHeight)
			}
		}
	}
}

func roleTestAbs64(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// TestRoleTemplateHeadVariation verifies head shapes differ between roles.
func TestRoleTemplateHeadVariation(t *testing.T) {
	mage := MageAerialTemplate(DirDown)
	rogue := RogueAerialTemplate(DirDown)

	mageHead := mage.BodyPartLayout[PartHead]
	rogueHead := rogue.BodyPartLayout[PartHead]

	// Mage has a much wider head (hat brim) than rogue
	if mageHead.RelativeWidth <= rogueHead.RelativeWidth {
		t.Errorf("mage head (%.3f) should be wider than rogue head (%.3f) due to hat",
			mageHead.RelativeWidth, rogueHead.RelativeWidth)
	}
}

// TestRoleTemplateKnightBulk verifies knight is the bulkiest template.
func TestRoleTemplateKnightBulk(t *testing.T) {
	knight := KnightAerialTemplate(DirDown)
	rogue := RogueAerialTemplate(DirDown)

	knightTorso := knight.BodyPartLayout[PartTorso]
	rogueTorso := rogue.BodyPartLayout[PartTorso]

	knightArea := knightTorso.RelativeWidth * knightTorso.RelativeHeight
	rogueArea := rogueTorso.RelativeWidth * rogueTorso.RelativeHeight

	if knightArea <= rogueArea {
		t.Errorf("knight torso area (%.4f) should be larger than rogue (%.4f)",
			knightArea, rogueArea)
	}
}

// TestMerchantHasBackpack verifies merchant template includes a pack.
func TestMerchantHasBackpack(t *testing.T) {
	m := MerchantAerialTemplate(DirDown)
	_, hasPack := m.BodyPartLayout[PartArmor]
	if !hasPack {
		t.Error("merchant template should have a backpack (using PartArmor slot)")
	}
}

// TestRangerHasQuiver verifies ranger template includes a quiver.
func TestRangerHasQuiver(t *testing.T) {
	r := RangerAerialTemplate(DirDown)
	quiver, hasQuiver := r.BodyPartLayout[PartArmor]
	if !hasQuiver {
		t.Error("ranger template should have a quiver (using PartArmor slot)")
	}
	// Quiver should be narrow and tall
	if quiver.RelativeWidth >= quiver.RelativeHeight {
		t.Errorf("quiver should be taller than wide (w:%.3f h:%.3f)",
			quiver.RelativeWidth, quiver.RelativeHeight)
	}
}

// TestAllDirectionsPopulated verifies all directions produce valid arm specs.
func TestAllDirectionsPopulated(t *testing.T) {
	directions := []Direction{DirUp, DirDown, DirLeft, DirRight}
	for _, dir := range directions {
		tmpl := WarriorAerialTemplate(dir)
		arms, ok := tmpl.BodyPartLayout[PartArms]
		if !ok {
			t.Errorf("direction %s missing arms", dir)
			continue
		}
		if arms.RelativeX == 0 && arms.RelativeY == 0 {
			t.Errorf("direction %s has zero arm position", dir)
		}
	}
}
