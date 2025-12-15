package engine

import (
	"sync"
	"testing"
)

func TestResourceNodeComponent_Type(t *testing.T) {
	node := NewResourceNodeComponent(ResourceTypeOre, "mountain")
	if got := node.Type(); got != "resource_node" {
		t.Errorf("Type() = %v, want resource_node", got)
	}
}

func TestResourceNodeComponent_NewResourceNodeComponent(t *testing.T) {
	tests := []struct {
		name         string
		resourceType ResourceType
		biome        string
		wantTool     ToolType
	}{
		{"ore node", ResourceTypeOre, "mountain", ToolTypePickaxe},
		{"wood node", ResourceTypeWood, "forest", ToolTypeAxe},
		{"herb node", ResourceTypeHerb, "plains", ToolTypeSickle},
		{"gem node", ResourceTypeGem, "cave", ToolTypePickaxe},
		{"fiber node", ResourceTypeFiber, "grassland", ToolTypeSickle},
		{"essence node", ResourceTypeEssence, "magical", ToolTypeStaff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewResourceNodeComponent(tt.resourceType, tt.biome)

			if node.ResourceType != tt.resourceType {
				t.Errorf("ResourceType = %v, want %v", node.ResourceType, tt.resourceType)
			}
			if node.BiomeType != tt.biome {
				t.Errorf("BiomeType = %v, want %v", node.BiomeType, tt.biome)
			}
			if node.RequiredTool != tt.wantTool {
				t.Errorf("RequiredTool = %v, want %v", node.RequiredTool, tt.wantTool)
			}
			if node.Quantity != 3 {
				t.Errorf("Quantity = %v, want 3", node.Quantity)
			}
			if node.MaxQuantity != 3 {
				t.Errorf("MaxQuantity = %v, want 3", node.MaxQuantity)
			}
			if node.IsDepleted {
				t.Error("New node should not be depleted")
			}
		})
	}
}

func TestResourceNodeComponent_CanHarvest(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(*ResourceNodeComponent)
		skillLevel int
		want       bool
	}{
		{
			name:       "can harvest with sufficient skill",
			setup:      func(n *ResourceNodeComponent) {},
			skillLevel: 10,
			want:       true,
		},
		{
			name:       "cannot harvest with insufficient skill",
			setup:      func(n *ResourceNodeComponent) { n.MinSkillLevel = 20 },
			skillLevel: 10,
			want:       false,
		},
		{
			name: "cannot harvest depleted node",
			setup: func(n *ResourceNodeComponent) {
				n.Quantity = 0
				n.IsDepleted = true
			},
			skillLevel: 10,
			want:       false,
		},
		{
			name:       "can harvest at exact skill level",
			setup:      func(n *ResourceNodeComponent) { n.MinSkillLevel = 10 },
			skillLevel: 10,
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewResourceNodeComponent(ResourceTypeOre, "mountain")
			tt.setup(node)

			if got := node.CanHarvest(tt.skillLevel); got != tt.want {
				t.Errorf("CanHarvest(%d) = %v, want %v", tt.skillLevel, got, tt.want)
			}
		})
	}
}

func TestResourceNodeComponent_Harvest(t *testing.T) {
	t.Run("successful harvest", func(t *testing.T) {
		node := NewResourceNodeComponent(ResourceTypeOre, "mountain")
		initialQty := node.Quantity

		if !node.Harvest() {
			t.Error("Harvest() should return true")
		}
		if node.Quantity != initialQty-1 {
			t.Errorf("Quantity = %v, want %v", node.Quantity, initialQty-1)
		}
	})

	t.Run("harvest until depleted", func(t *testing.T) {
		node := NewResourceNodeComponent(ResourceTypeOre, "mountain")
		node.Quantity = 1

		if !node.Harvest() {
			t.Error("Harvest() should succeed on last harvest")
		}
		if !node.IsDepleted {
			t.Error("Node should be depleted after last harvest")
		}
		if node.Quantity != 0 {
			t.Errorf("Quantity = %v, want 0", node.Quantity)
		}
	})

	t.Run("cannot harvest depleted node", func(t *testing.T) {
		node := NewResourceNodeComponent(ResourceTypeOre, "mountain")
		node.Quantity = 0
		node.IsDepleted = true

		if node.Harvest() {
			t.Error("Harvest() should return false for depleted node")
		}
	})
}

func TestResourceNodeComponent_UpdateRespawn(t *testing.T) {
	t.Run("respawn timer decreases", func(t *testing.T) {
		node := NewResourceNodeComponent(ResourceTypeOre, "mountain")
		node.IsDepleted = true
		node.RespawnTimer = 10.0

		if node.UpdateRespawn(5.0) {
			t.Error("UpdateRespawn should return false when timer not complete")
		}
		if node.RespawnTimer != 5.0 {
			t.Errorf("RespawnTimer = %v, want 5.0", node.RespawnTimer)
		}
	})

	t.Run("respawn completes", func(t *testing.T) {
		node := NewResourceNodeComponent(ResourceTypeOre, "mountain")
		node.IsDepleted = true
		node.RespawnTimer = 5.0

		if !node.UpdateRespawn(10.0) {
			t.Error("UpdateRespawn should return true when respawn completes")
		}
		if node.IsDepleted {
			t.Error("Node should no longer be depleted")
		}
		if node.Quantity != node.MaxQuantity {
			t.Errorf("Quantity = %v, want %v", node.Quantity, node.MaxQuantity)
		}
	})

	t.Run("no update when not depleted", func(t *testing.T) {
		node := NewResourceNodeComponent(ResourceTypeOre, "mountain")

		if node.UpdateRespawn(100.0) {
			t.Error("UpdateRespawn should return false when not depleted")
		}
	})
}

func TestResourceNodeComponent_GetRespawnProgress(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*ResourceNodeComponent)
		want      float64
		tolerance float64
	}{
		{
			name:      "not depleted returns 1.0",
			setup:     func(n *ResourceNodeComponent) {},
			want:      1.0,
			tolerance: 0.001,
		},
		{
			name: "half respawned",
			setup: func(n *ResourceNodeComponent) {
				n.IsDepleted = true
				n.RespawnTime = 100.0
				n.RespawnTimer = 50.0
			},
			want:      0.5,
			tolerance: 0.001,
		},
		{
			name: "just depleted",
			setup: func(n *ResourceNodeComponent) {
				n.IsDepleted = true
				n.RespawnTime = 100.0
				n.RespawnTimer = 100.0
			},
			want:      0.0,
			tolerance: 0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewResourceNodeComponent(ResourceTypeOre, "mountain")
			tt.setup(node)

			got := node.GetRespawnProgress()
			if got < tt.want-tt.tolerance || got > tt.want+tt.tolerance {
				t.Errorf("GetRespawnProgress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceNodeComponent_Serialization(t *testing.T) {
	node := NewResourceNodeComponent(ResourceTypeHerb, "plains")
	node.Quantity = 2
	node.IsDepleted = false

	data, err := node.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	node2 := &ResourceNodeComponent{}
	if err := node2.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if node2.ResourceType != node.ResourceType {
		t.Errorf("ResourceType = %v, want %v", node2.ResourceType, node.ResourceType)
	}
	if node2.Quantity != node.Quantity {
		t.Errorf("Quantity = %v, want %v", node2.Quantity, node.Quantity)
	}
	if node2.BiomeType != node.BiomeType {
		t.Errorf("BiomeType = %v, want %v", node2.BiomeType, node.BiomeType)
	}
}

func TestResourceNodeComponent_Concurrent(t *testing.T) {
	node := NewResourceNodeComponent(ResourceTypeOre, "mountain")
	node.Quantity = 100
	node.MaxQuantity = 100

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				node.Harvest()
				node.CanHarvest(10)
				node.GetQuantity()
			}
		}()
	}
	wg.Wait()
}

func TestGatheringComponent_Type(t *testing.T) {
	comp := NewGatheringComponent()
	if got := comp.Type(); got != "gathering" {
		t.Errorf("Type() = %v, want gathering", got)
	}
}

func TestGatheringComponent_NewGatheringComponent(t *testing.T) {
	comp := NewGatheringComponent()

	if comp.GatheringSkill != 1 {
		t.Errorf("GatheringSkill = %v, want 1", comp.GatheringSkill)
	}
	if comp.IsGathering {
		t.Error("New component should not be gathering")
	}
	if comp.GatherSpeed != 1.0 {
		t.Errorf("GatherSpeed = %v, want 1.0", comp.GatherSpeed)
	}
}

func TestGatheringComponent_StartStopGathering(t *testing.T) {
	comp := NewGatheringComponent()

	comp.StartGathering(123)
	if !comp.IsCurrentlyGathering() {
		t.Error("Should be gathering after StartGathering")
	}

	comp.StopGathering()
	if comp.IsCurrentlyGathering() {
		t.Error("Should not be gathering after StopGathering")
	}
	if comp.GetProgress() != 0 {
		t.Error("Progress should reset after StopGathering")
	}
}

func TestGatheringComponent_UpdateProgress(t *testing.T) {
	t.Run("progress increases", func(t *testing.T) {
		comp := NewGatheringComponent()
		comp.StartGathering(123)
		comp.GatherSpeed = 1.0

		// With base time of 2.0 and skill 1 (1.01 multiplier), 1 second = ~0.505
		complete := comp.UpdateProgress(1.0, 2.0)
		if complete {
			t.Error("Should not complete in 1 second with 2s base time")
		}
		if comp.GetProgress() <= 0 {
			t.Error("Progress should increase")
		}
	})

	t.Run("completes when progress reaches 1.0", func(t *testing.T) {
		comp := NewGatheringComponent()
		comp.StartGathering(123)
		comp.GatherSpeed = 10.0 // Fast gather

		complete := comp.UpdateProgress(1.0, 1.0)
		if !complete {
			t.Error("Should complete with high gather speed")
		}
	})

	t.Run("no progress when not gathering", func(t *testing.T) {
		comp := NewGatheringComponent()

		if comp.UpdateProgress(1.0, 1.0) {
			t.Error("Should not update when not gathering")
		}
	})
}

func TestGatheringComponent_CompleteGathering(t *testing.T) {
	comp := NewGatheringComponent()
	comp.StartGathering(123)

	comp.CompleteGathering(ResourceTypeOre)

	if comp.IsCurrentlyGathering() {
		t.Error("Should not be gathering after complete")
	}
	if comp.GetTotalHarvested(ResourceTypeOre) != 1 {
		t.Errorf("TotalHarvested = %v, want 1", comp.GetTotalHarvested(ResourceTypeOre))
	}
}

func TestGatheringComponent_AddXP(t *testing.T) {
	t.Run("add xp without level up", func(t *testing.T) {
		comp := NewGatheringComponent()

		leveledUp := comp.AddXP(50)
		if leveledUp {
			t.Error("Should not level up with 50 XP")
		}
	})

	t.Run("level up on sufficient xp", func(t *testing.T) {
		comp := NewGatheringComponent()

		leveledUp := comp.AddXP(150) // XP needed at level 1 is 150
		if !leveledUp {
			t.Error("Should level up with 150 XP at level 1")
		}
		if comp.GetSkillLevel() != 2 {
			t.Errorf("GatheringSkill = %v, want 2", comp.GetSkillLevel())
		}
	})

	t.Run("no level up at max level", func(t *testing.T) {
		comp := NewGatheringComponent()
		comp.GatheringSkill = 100

		leveledUp := comp.AddXP(10000)
		if leveledUp {
			t.Error("Should not level up at max level")
		}
		if comp.GetSkillLevel() != 100 {
			t.Errorf("GatheringSkill = %v, want 100", comp.GetSkillLevel())
		}
	})
}

func TestGatheringComponent_ToolManagement(t *testing.T) {
	comp := NewGatheringComponent()

	comp.EquipTool(ToolTypePickaxe)
	if comp.GetEquippedTool() != ToolTypePickaxe {
		t.Errorf("EquippedTool = %v, want pickaxe", comp.GetEquippedTool())
	}

	if !comp.HasCorrectTool(ToolTypePickaxe) {
		t.Error("Should have correct tool for pickaxe")
	}
	if comp.HasCorrectTool(ToolTypeAxe) {
		t.Error("Should not have correct tool for axe")
	}
	if !comp.HasCorrectTool(ToolTypeNone) {
		t.Error("Should always have 'none' tool")
	}
}

func TestGatheringComponent_ToolBonus(t *testing.T) {
	comp := NewGatheringComponent()

	// Default bonus is 1.0
	if bonus := comp.GetToolBonus(ToolTypePickaxe); bonus != 1.0 {
		t.Errorf("Default bonus = %v, want 1.0", bonus)
	}

	comp.SetToolBonus(ToolTypePickaxe, 1.5)
	if bonus := comp.GetToolBonus(ToolTypePickaxe); bonus != 1.5 {
		t.Errorf("Set bonus = %v, want 1.5", bonus)
	}
}

func TestGatheringComponent_Serialization(t *testing.T) {
	comp := NewGatheringComponent()
	comp.GatheringSkill = 25
	comp.EquipTool(ToolTypeAxe)
	comp.SetToolBonus(ToolTypeAxe, 1.5)

	data, err := comp.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	comp2 := &GatheringComponent{}
	if err := comp2.Deserialize(data); err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	if comp2.GatheringSkill != comp.GatheringSkill {
		t.Errorf("GatheringSkill = %v, want %v", comp2.GatheringSkill, comp.GatheringSkill)
	}
	if comp2.EquippedTool != comp.EquippedTool {
		t.Errorf("EquippedTool = %v, want %v", comp2.EquippedTool, comp.EquippedTool)
	}
}

func TestGatheringComponent_Concurrent(t *testing.T) {
	comp := NewGatheringComponent()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				comp.StartGathering(999)
				comp.UpdateProgress(0.1, 1.0)
				comp.GetProgress()
				comp.IsCurrentlyGathering()
				comp.StopGathering()
				comp.AddXP(1)
			}
		}(i)
	}
	wg.Wait()
}

func TestAllResourceTypes(t *testing.T) {
	types := AllResourceTypes()
	if len(types) != 6 {
		t.Errorf("AllResourceTypes() count = %v, want 6", len(types))
	}

	expected := map[ResourceType]bool{
		ResourceTypeOre:     true,
		ResourceTypeWood:    true,
		ResourceTypeHerb:    true,
		ResourceTypeGem:     true,
		ResourceTypeFiber:   true,
		ResourceTypeEssence: true,
	}

	for _, rt := range types {
		if !expected[rt] {
			t.Errorf("Unexpected resource type: %v", rt)
		}
	}
}

func TestRequiredToolForResource(t *testing.T) {
	tests := []struct {
		resource ResourceType
		want     ToolType
	}{
		{ResourceTypeOre, ToolTypePickaxe},
		{ResourceTypeWood, ToolTypeAxe},
		{ResourceTypeHerb, ToolTypeSickle},
		{ResourceTypeGem, ToolTypePickaxe},
		{ResourceTypeFiber, ToolTypeSickle},
		{ResourceTypeEssence, ToolTypeStaff},
		{ResourceType("unknown"), ToolTypeNone},
	}

	for _, tt := range tests {
		t.Run(string(tt.resource), func(t *testing.T) {
			if got := RequiredToolForResource(tt.resource); got != tt.want {
				t.Errorf("RequiredToolForResource(%v) = %v, want %v", tt.resource, got, tt.want)
			}
		})
	}
}
