//go:build !test
// +build !test

package sprites

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/shapes"
)

// TestItemRarity_String tests item rarity string representations.
func TestItemRarity_String(t *testing.T) {
	tests := []struct {
		rarity ItemRarity
		want   string
	}{
		{RarityCommon, "common"},
		{RarityUncommon, "uncommon"},
		{RarityRare, "rare"},
		{RarityEpic, "epic"},
		{RarityLegendary, "legendary"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.rarity.String()
			if got != tt.want {
				t.Errorf("ItemRarity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetRarityColorRole tests rarity color role mapping.
func TestGetRarityColorRole(t *testing.T) {
	tests := []struct {
		rarity ItemRarity
		want   string
	}{
		{RarityCommon, "primary"},
		{RarityUncommon, "secondary"},
		{RarityRare, "accent1"},
		{RarityEpic, "accent2"},
		{RarityLegendary, "accent3"},
	}

	for _, tt := range tests {
		t.Run(tt.rarity.String(), func(t *testing.T) {
			got := GetRarityColorRole(tt.rarity)
			if got != tt.want {
				t.Errorf("GetRarityColorRole(%v) = %v, want %v", tt.rarity, got, tt.want)
			}
		})
	}
}

// TestWeaponTemplates tests weapon item templates (Phase 5.4).
func TestWeaponTemplates(t *testing.T) {
	weapons := []struct {
		name     string
		itemType ItemType
		template func(ItemRarity) ItemTemplate
	}{
		{"sword", ItemSword, SwordTemplate},
		{"axe", ItemAxe, AxeTemplate},
		{"bow", ItemBow, BowTemplate},
		{"staff", ItemStaff, StaffTemplate},
		{"gun", ItemGun, GunTemplate},
	}

	for _, weapon := range weapons {
		t.Run(string(weapon.itemType), func(t *testing.T) {
			template := weapon.template(RarityCommon)

			// Verify template basics
			if template.Category != CategoryWeapon {
				t.Errorf("Category = %v, want %v", template.Category, CategoryWeapon)
			}
			if template.Type != weapon.itemType {
				t.Errorf("Type = %v, want %v", template.Type, weapon.itemType)
			}

			// Verify has parts
			if len(template.Parts) == 0 {
				t.Error("Template has no parts")
			}

			// Verify all parts have valid parameters
			for i, part := range template.Parts {
				if part.RelativeWidth <= 0 || part.RelativeHeight <= 0 {
					t.Errorf("Part %d has invalid dimensions", i)
				}
				if len(part.ShapeTypes) == 0 {
					t.Errorf("Part %d has no shape types", i)
				}
			}
		})
	}
}

// TestArmorTemplates tests armor item templates (Phase 5.4).
func TestArmorTemplates(t *testing.T) {
	template := HelmetTemplate(RarityCommon)

	if template.Category != CategoryArmor {
		t.Errorf("Category = %v, want %v", template.Category, CategoryArmor)
	}
	if template.Type != ItemHelmet {
		t.Errorf("Type = %v, want %v", template.Type, ItemHelmet)
	}
	if len(template.Parts) == 0 {
		t.Error("Helmet template has no parts")
	}
}

// TestConsumableTemplates tests consumable item templates (Phase 5.4).
func TestConsumableTemplates(t *testing.T) {
	consumables := []struct {
		name     string
		itemType ItemType
		template func(ItemRarity) ItemTemplate
	}{
		{"potion", ItemPotion, PotionTemplate},
		{"scroll", ItemScroll, ScrollTemplate},
	}

	for _, consumable := range consumables {
		t.Run(string(consumable.itemType), func(t *testing.T) {
			template := consumable.template(RarityCommon)

			if template.Category != CategoryConsumable {
				t.Errorf("Category = %v, want %v", template.Category, CategoryConsumable)
			}
			if template.Type != consumable.itemType {
				t.Errorf("Type = %v, want %v", template.Type, consumable.itemType)
			}
		})
	}
}

// TestAccessoryTemplates tests accessory item templates (Phase 5.4).
func TestAccessoryTemplates(t *testing.T) {
	template := RingTemplate(RarityCommon)

	if template.Category != CategoryAccessory {
		t.Errorf("Category = %v, want %v", template.Category, CategoryAccessory)
	}
	if template.Type != ItemRing {
		t.Errorf("Type = %v, want %v", template.Type, ItemRing)
	}
}

// TestQuestItemTemplates tests quest item templates (Phase 5.4).
func TestQuestItemTemplates(t *testing.T) {
	template := KeyTemplate(RarityCommon)

	if template.Category != CategoryQuest {
		t.Errorf("Category = %v, want %v", template.Category, CategoryQuest)
	}
	if template.Type != ItemKey {
		t.Errorf("Type = %v, want %v", template.Type, ItemKey)
	}
}

// TestRarityProgression tests that rarity increases visual complexity (Phase 5.4).
func TestRarityProgression(t *testing.T) {
	rarities := []ItemRarity{RarityCommon, RarityUncommon, RarityRare, RarityEpic, RarityLegendary}

	// Test with sword template
	prevParts := 0
	for _, rarity := range rarities {
		template := SwordTemplate(rarity)
		currentParts := len(template.Parts)

		// Higher rarities should have same or more parts (due to glow effects)
		if currentParts < prevParts {
			t.Errorf("Rarity %s has fewer parts (%d) than previous rarity (%d)",
				rarity.String(), currentParts, prevParts)
		}

		prevParts = currentParts
	}

	// Test with ring template
	prevParts = 0
	for _, rarity := range rarities {
		template := RingTemplate(rarity)
		currentParts := len(template.Parts)

		if currentParts < prevParts {
			t.Errorf("Ring rarity %s has fewer parts (%d) than previous rarity (%d)",
				rarity.String(), currentParts, prevParts)
		}

		prevParts = currentParts
	}
}

// TestSelectItemTemplate tests item template selection (Phase 5.4).
func TestSelectItemTemplate(t *testing.T) {
	tests := []struct {
		itemType ItemType
		rarity   ItemRarity
	}{
		{ItemSword, RarityCommon},
		{ItemAxe, RarityRare},
		{ItemBow, RarityEpic},
		{ItemStaff, RarityLegendary},
		{ItemGun, RarityUncommon},
		{ItemHelmet, RarityCommon},
		{ItemPotion, RarityRare},
		{ItemScroll, RarityUncommon},
		{ItemRing, RarityEpic},
		{ItemKey, RarityLegendary},
	}

	for _, tt := range tests {
		name := string(tt.itemType) + "_" + tt.rarity.String()
		t.Run(name, func(t *testing.T) {
			template := SelectItemTemplate(tt.itemType, tt.rarity)

			if template.Type != tt.itemType {
				t.Errorf("Template type = %v, want %v", template.Type, tt.itemType)
			}

			// Verify template name includes rarity
			if template.Name == "" {
				t.Error("Template has empty name")
			}
		})
	}
}

// TestItemPartValidation tests that all item parts have valid specifications.
func TestItemPartValidation(t *testing.T) {
	// Get all templates
	templates := []ItemTemplate{
		SwordTemplate(RarityLegendary),
		AxeTemplate(RarityLegendary),
		BowTemplate(RarityLegendary),
		StaffTemplate(RarityLegendary),
		GunTemplate(RarityLegendary),
		HelmetTemplate(RarityLegendary),
		PotionTemplate(RarityLegendary),
		ScrollTemplate(RarityLegendary),
		RingTemplate(RarityLegendary),
		KeyTemplate(RarityLegendary),
	}

	for _, template := range templates {
		t.Run(template.Name, func(t *testing.T) {
			for i, part := range template.Parts {
				// Validate dimensions
				if part.RelativeWidth <= 0 || part.RelativeWidth > 2.0 {
					t.Errorf("Part %d has invalid width: %f", i, part.RelativeWidth)
				}
				if part.RelativeHeight <= 0 || part.RelativeHeight > 2.0 {
					t.Errorf("Part %d has invalid height: %f", i, part.RelativeHeight)
				}

				// Validate opacity
				if part.Opacity < 0.0 || part.Opacity > 1.0 {
					t.Errorf("Part %d has invalid opacity: %f", i, part.Opacity)
				}

				// Validate has shapes
				if len(part.ShapeTypes) == 0 {
					t.Errorf("Part %d has no shape types", i)
				}

				// Validate shapes exist
				for _, shapeType := range part.ShapeTypes {
					if shapeType < shapes.ShapeCircle || shapeType > shapes.ShapeSkull {
						t.Errorf("Part %d has invalid shape type: %v", i, shapeType)
					}
				}
			}
		})
	}
}

// TestItemTemplateUniqueness tests that different item types are visually distinct.
func TestItemTemplateUniqueness(t *testing.T) {
	sword := SwordTemplate(RarityCommon)
	axe := AxeTemplate(RarityCommon)
	bow := BowTemplate(RarityCommon)

	// Swords and axes should have different primary shapes
	swordHasBlade := false
	for _, part := range sword.Parts {
		for _, shape := range part.ShapeTypes {
			if shape == shapes.ShapeBlade {
				swordHasBlade = true
			}
		}
	}
	if !swordHasBlade {
		t.Error("Sword template should include blade shape")
	}

	axeHasWedge := false
	for _, part := range axe.Parts {
		for _, shape := range part.ShapeTypes {
			if shape == shapes.ShapeWedge {
				axeHasWedge = true
			}
		}
	}
	if !axeHasWedge {
		t.Error("Axe template should include wedge shape")
	}

	bowHasCrescent := false
	for _, part := range bow.Parts {
		for _, shape := range part.ShapeTypes {
			if shape == shapes.ShapeCrescent {
				bowHasCrescent = true
			}
		}
	}
	if !bowHasCrescent {
		t.Error("Bow template should include crescent shape")
	}
}

// BenchmarkItemTemplates benchmarks item template generation (Phase 5.4).
func BenchmarkItemTemplates(b *testing.B) {
	items := []struct {
		name     string
		itemType ItemType
	}{
		{"sword", ItemSword},
		{"axe", ItemAxe},
		{"bow", ItemBow},
		{"staff", ItemStaff},
		{"gun", ItemGun},
		{"potion", ItemPotion},
	}

	for _, item := range items {
		b.Run(string(item.itemType), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = SelectItemTemplate(item.itemType, RarityCommon)
			}
		})
	}
}

// BenchmarkRarityVariants benchmarks rarity variant generation.
func BenchmarkRarityVariants(b *testing.B) {
	rarities := []ItemRarity{RarityCommon, RarityUncommon, RarityRare, RarityEpic, RarityLegendary}

	for _, rarity := range rarities {
		b.Run(rarity.String(), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				_ = SwordTemplate(rarity)
			}
		})
	}
}
