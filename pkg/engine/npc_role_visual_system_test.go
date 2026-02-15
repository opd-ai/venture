package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/class/advanced"
)

// TestNewNpcRoleVisualSystem verifies system creation.
func TestNewNpcRoleVisualSystem(t *testing.T) {
	world := NewWorld()
	sys := NewNpcRoleVisualSystem(world, 12345)
	if sys == nil {
		t.Fatal("expected non-nil system")
	}
}

// TestNpcRoleVisualComponent_Type verifies component type identifier.
func TestNpcRoleVisualComponent_Type(t *testing.T) {
	c := NewNpcRoleVisualComponent("mage")
	if c.Type() != "npc_role_visual" {
		t.Errorf("got %q, want %q", c.Type(), "npc_role_visual")
	}
	if c.Role != "mage" {
		t.Errorf("got role %q, want %q", c.Role, "mage")
	}
}

// TestNpcRoleVisualSystem_InfersMerchant verifies merchant component detection.
func TestNpcRoleVisualSystem_InfersMerchant(t *testing.T) {
	world := NewWorld()
	sys := NewNpcRoleVisualSystem(world, 100)

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{Width: 32, Height: 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(NewMerchantComponent(20, MerchantFixed, 1.5))

	sys.Update([]*Entity{entity}, 3.0)

	comp, ok := entity.GetComponent("npc_role_visual")
	if !ok {
		t.Fatal("expected npc_role_visual component")
	}
	role := comp.(*NpcRoleVisualComponent).Role
	if role != "merchant" {
		t.Errorf("got role %q, want %q", role, "merchant")
	}
}

// TestNpcRoleVisualSystem_InfersFromClass verifies class-based role detection.
func TestNpcRoleVisualSystem_InfersFromClass(t *testing.T) {
	tests := []struct {
		name     string
		classID  advanced.ClassID
		wantRole string
	}{
		{"mage", advanced.ClassMage, "mage"},
		{"warrior", advanced.ClassWarrior, "warrior"},
		{"knight", advanced.ClassKnight, "knight"},
		{"rogue", advanced.ClassRogue, "rogue"},
		{"ranger", advanced.ClassRanger, "ranger"},
		{"cleric", advanced.ClassCleric, "priest"},
		{"bard", advanced.ClassBard, "priest"},
		{"necromancer", advanced.ClassNecromancer, "mage"},
		{"paladin", advanced.ClassPaladin, "knight"},
		{"assassin", advanced.ClassAssassin, "rogue"},
		{"berserker", advanced.ClassBerserker, "warrior"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewNpcRoleVisualSystem(world, 200)

			entity := world.CreateEntity()
			entity.AddComponent(&EbitenSprite{Width: 32, Height: 32})
			entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
			entity.AddComponent(&advanced.AdvancedClassComponent{PrimaryClass: tt.classID})

			sys.Update([]*Entity{entity}, 3.0)

			comp, ok := entity.GetComponent("npc_role_visual")
			if !ok {
				t.Fatal("expected npc_role_visual component")
			}
			role := comp.(*NpcRoleVisualComponent).Role
			if role != tt.wantRole {
				t.Errorf("got role %q, want %q", role, tt.wantRole)
			}
		})
	}
}

// TestNpcRoleVisualSystem_SkipsExisting verifies idempotence.
func TestNpcRoleVisualSystem_SkipsExisting(t *testing.T) {
	world := NewWorld()
	sys := NewNpcRoleVisualSystem(world, 300)

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{Width: 32, Height: 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(NewMerchantComponent(20, MerchantFixed, 1.5))
	entity.AddComponent(NewNpcRoleVisualComponent("warrior")) // pre-assigned

	sys.Update([]*Entity{entity}, 3.0)

	comp, _ := entity.GetComponent("npc_role_visual")
	role := comp.(*NpcRoleVisualComponent).Role
	if role != "warrior" {
		t.Errorf("should not overwrite existing role; got %q, want %q", role, "warrior")
	}
}

// TestNpcRoleVisualSystem_SkipsNonHumanoid verifies creatures are skipped.
func TestNpcRoleVisualSystem_SkipsNonHumanoid(t *testing.T) {
	world := NewWorld()
	sys := NewNpcRoleVisualSystem(world, 400)

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{Width: 32, Height: 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(&CreatureVisualComponent{Form: FormQuadruped})

	sys.Update([]*Entity{entity}, 3.0)

	_, ok := entity.GetComponent("npc_role_visual")
	if ok {
		t.Error("non-humanoid entity should not get role component")
	}
}

// TestNpcRoleVisualSystem_SkipsNoSprite verifies entities without sprites are skipped.
func TestNpcRoleVisualSystem_SkipsNoSprite(t *testing.T) {
	world := NewWorld()
	sys := NewNpcRoleVisualSystem(world, 500)

	entity := world.CreateEntity()
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})

	sys.Update([]*Entity{entity}, 3.0)

	_, ok := entity.GetComponent("npc_role_visual")
	if ok {
		t.Error("entity without sprite should not get role component")
	}
}

// TestNpcRoleVisualSystem_ThrottledScan verifies scan interval throttling.
func TestNpcRoleVisualSystem_ThrottledScan(t *testing.T) {
	world := NewWorld()
	sys := NewNpcRoleVisualSystem(world, 600)

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{Width: 32, Height: 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	entity.AddComponent(NewMerchantComponent(20, MerchantFixed, 1.5))

	// First update with small delta — should not scan
	sys.Update([]*Entity{entity}, 0.5)
	_, ok := entity.GetComponent("npc_role_visual")
	if ok {
		t.Error("should not scan within interval")
	}

	// Second update that crosses threshold
	sys.Update([]*Entity{entity}, 2.0)
	_, ok = entity.GetComponent("npc_role_visual")
	if !ok {
		t.Error("should scan after interval")
	}
}

// TestNpcRoleVisualSystem_StatsBasedInference verifies fallback to stats heuristic.
func TestNpcRoleVisualSystem_StatsBasedInference(t *testing.T) {
	tests := []struct {
		name     string
		stats    StatsComponent
		wantRole string
	}{
		{"high_magic→mage", StatsComponent{MagicPower: 50, Attack: 10}, "mage"},
		{"high_evasion→rogue", StatsComponent{MagicPower: 10, Attack: 10, Evasion: 0.4}, "rogue"},
		{"high_block→knight", StatsComponent{MagicPower: 10, Attack: 10, BlockChance: 0.4}, "knight"},
		{"high_attack→warrior", StatsComponent{MagicPower: 10, Attack: 50}, "warrior"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewNpcRoleVisualSystem(world, 700)

			entity := world.CreateEntity()
			entity.AddComponent(&EbitenSprite{Width: 32, Height: 32})
			entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
			entity.AddComponent(&tt.stats)

			sys.Update([]*Entity{entity}, 3.0)

			comp, ok := entity.GetComponent("npc_role_visual")
			if !ok {
				t.Fatal("expected npc_role_visual component")
			}
			role := comp.(*NpcRoleVisualComponent).Role
			if role != tt.wantRole {
				t.Errorf("got role %q, want %q", role, tt.wantRole)
			}
		})
	}
}

// TestNpcRoleVisualSystem_GetProcessedCount verifies counter.
func TestNpcRoleVisualSystem_GetProcessedCount(t *testing.T) {
	world := NewWorld()
	sys := NewNpcRoleVisualSystem(world, 800)

	e1 := world.CreateEntity()
	e1.AddComponent(&EbitenSprite{Width: 32, Height: 32})
	e1.AddComponent(&HealthComponent{Current: 100, Max: 100})
	e1.AddComponent(NewMerchantComponent(20, MerchantFixed, 1.5))

	e2 := world.CreateEntity()
	e2.AddComponent(&EbitenSprite{Width: 32, Height: 32})
	e2.AddComponent(&HealthComponent{Current: 50, Max: 50})
	e2.AddComponent(&advanced.AdvancedClassComponent{PrimaryClass: advanced.ClassMage})

	sys.Update([]*Entity{e1, e2}, 3.0)
	if got := sys.GetProcessedCount(); got != 2 {
		t.Errorf("got %d processed, want 2", got)
	}
}

// TestNpcRoleVisualSystem_NameBasedMerchant verifies merchant name detection.
func TestNpcRoleVisualSystem_NameBasedMerchant(t *testing.T) {
	world := NewWorld()
	sys := NewNpcRoleVisualSystem(world, 900)

	entity := world.CreateEntity()
	entity.AddComponent(&EbitenSprite{Width: 32, Height: 32})
	entity.AddComponent(&HealthComponent{Current: 100, Max: 100})
	mc := NewMerchantComponent(20, MerchantFixed, 1.5)
	mc.MerchantName = "The Wandering Trader"
	entity.AddComponent(mc)

	sys.Update([]*Entity{entity}, 3.0)

	comp, ok := entity.GetComponent("npc_role_visual")
	if !ok {
		t.Fatal("expected npc_role_visual component")
	}
	if comp.(*NpcRoleVisualComponent).Role != "merchant" {
		t.Errorf("merchant with name should still be classified as merchant")
	}
}
