package engine

import (
	"testing"
)

// TestCreatureVisualComponent_Type verifies the component type string.
func TestCreatureVisualComponent_Type(t *testing.T) {
	c := &CreatureVisualComponent{Form: FormHumanoid}
	if c.Type() != "creature_visual" {
		t.Errorf("expected type creature_visual, got %s", c.Type())
	}
}

// TestClassifyCreatureForm_Tags tests tag-based creature form classification.
func TestClassifyCreatureForm_Tags(t *testing.T) {
	tests := []struct {
		name     string
		entName  string
		tags     []string
		expected CreatureForm
	}{
		{"undead tag", "Dark Warrior", []string{"undead", "medium"}, FormUndead},
		{"robotic tag", "Scout Unit", []string{"robotic", "fast"}, FormMechanical},
		{"wild tag", "Feral Beast", []string{"wild", "aggressive"}, FormQuadruped},
		{"human tag", "Street Runner", []string{"human", "augmented"}, FormHumanoid},
		{"spider keyword tag", "Lurker", []string{"spider", "fast"}, FormArachnid},
		{"dragon keyword tag", "Fire Breather", []string{"dragon", "boss"}, FormFlying},
		{"slime keyword tag", "Goo", []string{"slime"}, FormBlob},
		{"construct keyword tag", "Stone Thing", []string{"construct"}, FormMechanical},
		{"serpent keyword tag", "Hisser", []string{"serpent"}, FormSerpentine},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyCreatureForm(tt.entName, tt.tags)
			if got != tt.expected {
				t.Errorf("ClassifyCreatureForm(%q, %v) = %s, want %s",
					tt.entName, tt.tags, got, tt.expected)
			}
		})
	}
}

// TestClassifyCreatureForm_Name tests name-based creature form classification.
func TestClassifyCreatureForm_Name(t *testing.T) {
	tests := []struct {
		name     string
		entName  string
		tags     []string
		expected CreatureForm
	}{
		{"name with wolf", "Dire Wolf", nil, FormQuadruped},
		{"name with spider", "Giant Spider", nil, FormArachnid},
		{"name with dragon", "Ancient Dragon", nil, FormFlying},
		{"name with zombie", "Shambling Zombie", nil, FormUndead},
		{"name with slime", "Green Slime", nil, FormBlob},
		{"name with snake", "Pit Snake", nil, FormSerpentine},
		{"name with golem", "Iron Golem", nil, FormMechanical},
		{"name with bot", "War Bot", nil, FormMechanical},
		{"humanoid fallback", "Orc Warrior", nil, FormHumanoid},
		{"empty name", "", nil, FormHumanoid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyCreatureForm(tt.entName, tt.tags)
			if got != tt.expected {
				t.Errorf("ClassifyCreatureForm(%q, %v) = %s, want %s",
					tt.entName, tt.tags, got, tt.expected)
			}
		})
	}
}

// TestClassifyCreatureForm_TagPriority verifies tags override name-based classification.
func TestClassifyCreatureForm_TagPriority(t *testing.T) {
	// Name says "dragon" (flying) but tag says "robot" (mechanical) — tag wins.
	got := ClassifyCreatureForm("Cyber Dragon", []string{"robot"})
	if got != FormMechanical {
		t.Errorf("expected tags to take priority, got %s", got)
	}
}

// TestCreatureVisualClassifierSystem_Update tests automatic classification of enemies.
func TestCreatureVisualClassifierSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureVisualClassifierSystem(world)

	// Create an enemy entity without creature_visual
	enemy := world.CreateEntity()
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&ColliderComponent{Width: 32, Height: 32})
	enemy.AddComponent(&AttackComponent{Damage: 10})
	enemy.AddComponent(NewAnimationComponent(42))

	entities := []*Entity{enemy}
	sys.Update(entities, 0.016)

	// Should now have creature_visual component
	comp, ok := enemy.GetComponent("creature_visual")
	if !ok {
		t.Fatal("expected creature_visual component after classifier update")
	}
	cv := comp.(*CreatureVisualComponent)
	if cv.Form != FormHumanoid {
		t.Errorf("medium-sized generic enemy should be humanoid, got %s", cv.Form)
	}
	if cv.SizeClass != "medium" {
		t.Errorf("expected size class medium, got %s", cv.SizeClass)
	}
}

// TestCreatureVisualClassifierSystem_SkipsExisting verifies idempotence.
func TestCreatureVisualClassifierSystem_SkipsExisting(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureVisualClassifierSystem(world)

	enemy := world.CreateEntity()
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	enemy.AddComponent(&CreatureVisualComponent{
		Form:      FormFlying,
		SizeClass: "huge",
	})

	entities := []*Entity{enemy}
	sys.Update(entities, 0.016)

	comp, _ := enemy.GetComponent("creature_visual")
	cv := comp.(*CreatureVisualComponent)
	if cv.Form != FormFlying {
		t.Errorf("existing form should be preserved, got %s", cv.Form)
	}
}

// TestCreatureVisualClassifierSystem_SkipsNonEnemies verifies friendly entities are skipped.
func TestCreatureVisualClassifierSystem_SkipsNonEnemies(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureVisualClassifierSystem(world)

	player := world.CreateEntity()
	player.AddComponent(&TeamComponent{TeamID: 1})

	entities := []*Entity{player}
	sys.Update(entities, 0.016)

	if player.HasComponent("creature_visual") {
		t.Error("player entity should not receive creature_visual")
	}
}

// TestCreatureVisualClassifierSystem_TinyIsArachnid verifies tiny creatures get arachnid form.
func TestCreatureVisualClassifierSystem_TinyIsArachnid(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureVisualClassifierSystem(world)

	tiny := world.CreateEntity()
	tiny.AddComponent(&TeamComponent{TeamID: 2})
	tiny.AddComponent(&ColliderComponent{Width: 16, Height: 16})
	tiny.AddComponent(NewAnimationComponent(99))

	entities := []*Entity{tiny}
	sys.Update(entities, 0.016)

	comp, ok := tiny.GetComponent("creature_visual")
	if !ok {
		t.Fatal("expected creature_visual")
	}
	cv := comp.(*CreatureVisualComponent)
	if cv.Form != FormArachnid {
		t.Errorf("tiny enemy should be arachnid, got %s", cv.Form)
	}
}

// TestCreatureVisualClassifierSystem_HugeBossIsFlying verifies huge high-damage bosses.
func TestCreatureVisualClassifierSystem_HugeBossIsFlying(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureVisualClassifierSystem(world)

	boss := world.CreateEntity()
	boss.AddComponent(&TeamComponent{TeamID: 2})
	boss.AddComponent(&ColliderComponent{Width: 64, Height: 64})
	boss.AddComponent(&AttackComponent{Damage: 50})
	boss.AddComponent(NewAnimationComponent(123))

	entities := []*Entity{boss}
	sys.Update(entities, 0.016)

	comp, ok := boss.GetComponent("creature_visual")
	if !ok {
		t.Fatal("expected creature_visual")
	}
	cv := comp.(*CreatureVisualComponent)
	if cv.Form != FormFlying {
		t.Errorf("huge high-damage boss should be flying, got %s", cv.Form)
	}
}

// TestCreatureVisualClassifierSystem_MarksAnimDirty verifies animation dirty flag is set.
func TestCreatureVisualClassifierSystem_MarksAnimDirty(t *testing.T) {
	world := NewWorld()
	sys := NewCreatureVisualClassifierSystem(world)

	enemy := world.CreateEntity()
	enemy.AddComponent(&TeamComponent{TeamID: 2})
	anim := NewAnimationComponent(55)
	anim.Dirty = false
	enemy.AddComponent(anim)

	entities := []*Entity{enemy}
	sys.Update(entities, 0.016)

	if !anim.Dirty {
		t.Error("animation should be marked dirty after classification")
	}
}
