package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
)

func TestCreatureAnatomyComponent_Type(t *testing.T) {
	comp := NewCreatureAnatomyComponent()
	if comp.Type() != "creature_anatomy" {
		t.Errorf("Type() = %v, want creature_anatomy", comp.Type())
	}
}

func TestCreatureAnatomyComponent_Defaults(t *testing.T) {
	comp := NewCreatureAnatomyComponent()

	if comp.AnatomyType != AnatomyHumanoid {
		t.Errorf("default AnatomyType = %v, want AnatomyHumanoid", comp.AnatomyType)
	}
	if comp.SizeModifier != 1.0 {
		t.Errorf("default SizeModifier = %v, want 1.0", comp.SizeModifier)
	}
	if comp.Assigned {
		t.Error("default Assigned should be false")
	}
}

func TestAnatomyType_String(t *testing.T) {
	tests := []struct {
		anatomy  AnatomyType
		expected string
	}{
		{AnatomyHumanoid, "humanoid"},
		{AnatomyQuadruped, "quadruped"},
		{AnatomySerpentine, "serpentine"},
		{AnatomyArachnid, "arachnid"},
		{AnatomyInsect, "insect"},
		{AnatomyFlying, "flying"},
		{AnatomyBlob, "blob"},
		{AnatomyMechanical, "mechanical"},
		{AnatomyUndead, "undead"},
		{AnatomyMultiLimbed, "multi_limbed"},
		{AnatomyType(99), "humanoid"}, // Unknown fallback
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.anatomy.String(); got != tt.expected {
				t.Errorf("AnatomyType(%d).String() = %v, want %v", tt.anatomy, got, tt.expected)
			}
		})
	}
}

func TestCreatureAnatomySystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureAnatomySystem(world, 12345)

	// Create entity with sprite and name
	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{})
	entity.AddComponent(&NameComponent{Name: "Giant Spider"})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Check anatomy was assigned
	comp, ok := entity.GetComponent("creature_anatomy")
	if !ok {
		t.Fatal("creature_anatomy component not added")
	}

	anatomy, ok := comp.(*CreatureAnatomyComponent)
	if !ok {
		t.Fatal("component should be *CreatureAnatomyComponent")
	}

	if !anatomy.Assigned {
		t.Error("Assigned should be true after Update")
	}

	// Spider should get arachnid anatomy
	if anatomy.AnatomyType != AnatomyArachnid {
		t.Errorf("AnatomyType = %v, want AnatomyArachnid for 'Giant Spider'", anatomy.AnatomyType)
	}
}

func TestCreatureAnatomySystem_SkipsNonSpriteEntities(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureAnatomySystem(world, 12345)

	// Entity without sprite
	entity := world.CreateEntity()
	entity.AddComponent(&NameComponent{Name: "Wolf"})

	entities := []*Entity{entity}
	sys.Update(entities, 0.016)

	// Should not have anatomy component
	if _, ok := entity.GetComponent("creature_anatomy"); ok {
		t.Error("entity without sprite should not get anatomy component")
	}
}

func TestCreatureAnatomySystem_DeterministicResults(t *testing.T) {
	// Same seed should give same results
	world1 := NewWorld()
	sys1 := NewCreatureAnatomySystem(world1, 42)
	entity1 := world1.CreateEntity()
	entity1.AddComponent(&EbitenSprite{})
	entity1.AddComponent(&NameComponent{Name: "Unknown Creature"})
	entity1.AddComponent(&TagComponent{Tags: []string{"monster"}})

	world2 := NewWorld()
	sys2 := NewCreatureAnatomySystem(world2, 42)
	entity2 := world2.CreateEntity()
	entity2.AddComponent(&EbitenSprite{})
	entity2.AddComponent(&NameComponent{Name: "Unknown Creature"})
	entity2.AddComponent(&TagComponent{Tags: []string{"monster"}})

	sys1.Update([]*Entity{entity1}, 0.016)
	sys2.Update([]*Entity{entity2}, 0.016)

	anatomy1, _ := entity1.GetComponent("creature_anatomy")
	anatomy2, _ := entity2.GetComponent("creature_anatomy")

	a1 := anatomy1.(*CreatureAnatomyComponent)
	a2 := anatomy2.(*CreatureAnatomyComponent)

	if a1.AnatomyType != a2.AnatomyType {
		t.Errorf("same seed produced different anatomy: %v vs %v",
			a1.AnatomyType, a2.AnatomyType)
	}
}

func TestCreatureAnatomySystem_KeywordMatching(t *testing.T) {
	tests := []struct {
		name        string
		entityName  string
		tags        []string
		wantAnatomy AnatomyType
	}{
		{"dragon by name", "Fire Dragon", nil, AnatomyFlying},
		{"wolf by name", "Dire Wolf", nil, AnatomyQuadruped},
		{"slime by name", "Green Slime", nil, AnatomyBlob},
		{"robot by name", "Combat Robot", nil, AnatomyMechanical},
		{"snake by name", "Giant Snake", nil, AnatomySerpentine},
		{"beetle by name", "Fire Beetle", nil, AnatomyInsect},
		{"spider by tag", "Crawler", []string{"spider"}, AnatomyArachnid},
		{"undead by name", "Skeletal Warrior", nil, AnatomyUndead},
		{"horror by name", "Shoggoth", nil, AnatomyMultiLimbed},
		{"orc humanoid", "Orc Warrior", nil, AnatomyHumanoid},
		{"merchant humanoid", "Merchant Smith", nil, AnatomyHumanoid},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCreatureAnatomySystem(world, 12345)

			entity := world.CreateEntity()
			entity.AddComponent(&EbitenSprite{})
			entity.AddComponent(&NameComponent{Name: tt.entityName})
			if tt.tags != nil {
				entity.AddComponent(&TagComponent{Tags: tt.tags})
			}

			sys.Update([]*Entity{entity}, 0.016)

			comp, _ := entity.GetComponent("creature_anatomy")
			anatomy := comp.(*CreatureAnatomyComponent)
			if anatomy.AnatomyType != tt.wantAnatomy {
				t.Errorf("AnatomyType = %v, want %v for '%s'",
					anatomy.AnatomyType, tt.wantAnatomy, tt.entityName)
			}
		})
	}
}

func TestCreatureAnatomySystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureAnatomySystem(world, 12345)

	sys.SetGenre("horror")

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{})
	entity.AddComponent(&NameComponent{Name: "Dark Spider"})

	sys.Update([]*Entity{entity}, 0.016)

	comp, _ := entity.GetComponent("creature_anatomy")
	anatomy := comp.(*CreatureAnatomyComponent)
	if anatomy.GenreVariant != "horror" {
		t.Errorf("GenreVariant = %v, want horror", anatomy.GenreVariant)
	}
}

func TestCreatureAnatomyComponent_GetTemplateSelector(t *testing.T) {
	comp := &CreatureAnatomyComponent{
		AnatomyType: AnatomyArachnid,
		SubType:     "spider",
	}

	if got := comp.GetTemplateSelector(); got != "spider" {
		t.Errorf("GetTemplateSelector() = %v, want spider", got)
	}

	// Without subtype, should return anatomy type string
	comp.SubType = ""
	if got := comp.GetTemplateSelector(); got != "arachnid" {
		t.Errorf("GetTemplateSelector() = %v, want arachnid", got)
	}
}

func TestCreatureAnatomyComponent_GetAerialTemplate(t *testing.T) {
	comp := &CreatureAnatomyComponent{
		AnatomyType:  AnatomyQuadruped,
		SubType:      "wolf",
		GenreVariant: "fantasy",
	}

	template := comp.GetAerialTemplate(sprites.DirDown)
	if template.Name == "" {
		t.Error("GetAerialTemplate should return a valid template")
	}
}

func TestCreatureTypeComponent_Type(t *testing.T) {
	comp := &CreatureTypeComponent{CreatureType: "wolf"}
	if comp.Type() != "creature_type" {
		t.Errorf("Type() = %v, want creature_type", comp.Type())
	}
}

func TestCreatureAnatomySystem_UpdateIdempotent(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureAnatomySystem(world, 12345)

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{})
	entity.AddComponent(&NameComponent{Name: "Giant Spider"})

	entities := []*Entity{entity}

	// First update
	sys.Update(entities, 0.016)
	comp1, _ := entity.GetComponent("creature_anatomy")
	anatomy1 := comp1.(*CreatureAnatomyComponent)
	type1 := anatomy1.AnatomyType

	// Second update should not change anything
	sys.Update(entities, 0.016)
	comp2, _ := entity.GetComponent("creature_anatomy")
	anatomy2 := comp2.(*CreatureAnatomyComponent)

	if anatomy2.AnatomyType != type1 {
		t.Errorf("second update changed anatomy: %v -> %v", type1, anatomy2.AnatomyType)
	}
}

func TestBuildKeywordMap(t *testing.T) {
	keywords := buildKeywordMap()

	// Verify key mappings exist
	tests := []struct {
		keyword string
		want    AnatomyType
	}{
		{"wolf", AnatomyQuadruped},
		{"spider", AnatomyArachnid},
		{"dragon", AnatomyFlying},
		{"slime", AnatomyBlob},
		{"robot", AnatomyMechanical},
		{"snake", AnatomySerpentine},
		{"beetle", AnatomyInsect},
		{"skeleton", AnatomyUndead},
		{"kraken", AnatomyMultiLimbed},
		{"orc", AnatomyHumanoid},
	}

	for _, tt := range tests {
		if got, ok := keywords[tt.keyword]; !ok {
			t.Errorf("keyword %q not found in map", tt.keyword)
		} else if got != tt.want {
			t.Errorf("keyword %q = %v, want %v", tt.keyword, got, tt.want)
		}
	}
}
