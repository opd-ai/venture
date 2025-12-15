package engine

import (
	"testing"
)

func TestNewGatheringSystem(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	if system == nil {
		t.Fatal("NewGatheringSystem returned nil")
	}
	if system.world != world {
		t.Error("World not set correctly")
	}
	if system.BaseGatherTime != 3.0 {
		t.Errorf("BaseGatherTime = %v, want 3.0", system.BaseGatherTime)
	}
	if system.XPPerHarvest != 10 {
		t.Errorf("XPPerHarvest = %v, want 10", system.XPPerHarvest)
	}
}

func TestGatheringSystem_Update_NoEntities(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	// Should not panic with empty entities
	system.Update([]*Entity{}, 0.016)
}

func TestGatheringSystem_Update_ResourceNodeRespawn(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	// Create depleted resource node
	node := world.CreateEntity()
	nodeComp := NewResourceNodeComponent(ResourceTypeOre, "mountain")
	nodeComp.IsDepleted = true
	nodeComp.RespawnTimer = 1.0
	nodeComp.Quantity = 0
	node.AddComponent(nodeComp)

	entities := []*Entity{node}

	// Update with enough time to respawn
	system.Update(entities, 2.0)

	if nodeComp.IsDepleted {
		t.Error("Node should have respawned")
	}
	if nodeComp.Quantity != nodeComp.MaxQuantity {
		t.Errorf("Quantity = %v, want %v", nodeComp.Quantity, nodeComp.MaxQuantity)
	}
}

func TestGatheringSystem_Update_GatheringProgress(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)
	system.BaseGatherTime = 1.0 // 1 second gather time

	// Create gatherer
	gatherer := world.CreateEntity()
	gatherComp := NewGatheringComponent()
	gatherComp.EquipTool(ToolTypePickaxe)
	gatherComp.GatherSpeed = 10.0 // Fast gathering for test
	gatherer.AddComponent(gatherComp)

	// Create resource node
	node := world.CreateEntity()
	nodeComp := NewResourceNodeComponent(ResourceTypeOre, "mountain")
	node.AddComponent(nodeComp)
	node.AddComponent(&PositionComponent{X: 0, Y: 0})

	entities := []*Entity{gatherer, node}

	// Start gathering
	system.Update(entities, 0) // Index nodes
	gatherComp.StartGathering(node.ID)

	// Verify gathering started
	if !gatherComp.IsCurrentlyGathering() {
		t.Error("Should be gathering")
	}

	// Update until complete
	harvestCompleted := false
	system.OnHarvestCallback = func(g, n *Entity, rt ResourceType, yield int) {
		harvestCompleted = true
	}

	system.Update(entities, 1.0)

	if !harvestCompleted {
		t.Error("Harvest should have completed")
	}
	if gatherComp.IsCurrentlyGathering() {
		t.Error("Should not be gathering after completion")
	}
}

func TestGatheringSystem_Update_CancelsOnMissingNode(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	// Create gatherer targeting non-existent node
	gatherer := world.CreateEntity()
	gatherComp := NewGatheringComponent()
	gatherComp.StartGathering(0)
	gatherer.AddComponent(gatherComp)

	entities := []*Entity{gatherer}
	system.Update(entities, 0.016)

	if gatherComp.IsCurrentlyGathering() {
		t.Error("Gathering should cancel when node not found")
	}
}

func TestGatheringSystem_Update_CancelsOnDepletedNode(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	// Create gatherer
	gatherer := world.CreateEntity()
	gatherComp := NewGatheringComponent()
	gatherComp.EquipTool(ToolTypePickaxe)
	gatherer.AddComponent(gatherComp)

	// Create depleted resource node
	node := world.CreateEntity()
	nodeComp := NewResourceNodeComponent(ResourceTypeOre, "mountain")
	nodeComp.IsDepleted = true
	nodeComp.Quantity = 0
	nodeComp.RespawnTimer = 100.0 // Set respawn timer so it doesn't immediately respawn
	node.AddComponent(nodeComp)

	entities := []*Entity{gatherer, node}

	// Index nodes
	system.Update(entities, 0)

	// Start gathering on depleted node
	gatherComp.StartGathering(node.ID)
	system.Update(entities, 0.016)

	if gatherComp.IsCurrentlyGathering() {
		t.Error("Gathering should cancel on depleted node")
	}
}

func TestGatheringSystem_Update_CancelsOnWrongTool(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	// Create gatherer with wrong tool
	gatherer := world.CreateEntity()
	gatherComp := NewGatheringComponent()
	gatherComp.EquipTool(ToolTypeAxe) // Wrong tool for ore
	gatherer.AddComponent(gatherComp)

	// Create ore node (requires pickaxe)
	node := world.CreateEntity()
	nodeComp := NewResourceNodeComponent(ResourceTypeOre, "mountain")
	node.AddComponent(nodeComp)

	entities := []*Entity{gatherer, node}

	// Index nodes
	system.Update(entities, 0)

	// Start gathering with wrong tool
	gatherComp.StartGathering(node.ID)
	system.Update(entities, 0.016)

	if gatherComp.IsCurrentlyGathering() {
		t.Error("Gathering should cancel with wrong tool")
	}
}

func TestGatheringSystem_StartGathering(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	// Create gatherer
	gatherer := world.CreateEntity()
	gatherComp := NewGatheringComponent()
	gatherComp.EquipTool(ToolTypePickaxe)
	gatherer.AddComponent(gatherComp)

	// Create resource node
	node := world.CreateEntity()
	nodeComp := NewResourceNodeComponent(ResourceTypeOre, "mountain")
	node.AddComponent(nodeComp)
	node.AddComponent(&PositionComponent{X: 0, Y: 0})

	entities := []*Entity{gatherer, node}
	system.Update(entities, 0) // Index nodes

	// Start gathering
	if !system.StartGathering(gatherer, node.ID) {
		t.Error("StartGathering should succeed")
	}
	if !gatherComp.IsCurrentlyGathering() {
		t.Error("Should be gathering")
	}
}

func TestGatheringSystem_StartGathering_Fails(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	t.Run("fails without gathering component", func(t *testing.T) {
		gatherer := world.CreateEntity()
		if system.StartGathering(gatherer, 0) {
			t.Error("Should fail without gathering component")
		}
	})

	t.Run("fails with non-existent node", func(t *testing.T) {
		gatherer := world.CreateEntity()
		gatherer.AddComponent(NewGatheringComponent())
		if system.StartGathering(gatherer, 0) {
			t.Error("Should fail with non-existent node")
		}
	})

	t.Run("fails with insufficient skill", func(t *testing.T) {
		gatherer := world.CreateEntity()
		gatherComp := NewGatheringComponent()
		gatherComp.GatheringSkill = 1
		gatherComp.EquipTool(ToolTypePickaxe)
		gatherer.AddComponent(gatherComp)

		node := world.CreateEntity()
		nodeComp := NewResourceNodeComponent(ResourceTypeOre, "mountain")
		nodeComp.MinSkillLevel = 50 // High skill requirement
		node.AddComponent(nodeComp)

		entities := []*Entity{gatherer, node}
		system.Update(entities, 0)

		if system.StartGathering(gatherer, node.ID) {
			t.Error("Should fail with insufficient skill")
		}
	})

	t.Run("fails with wrong tool", func(t *testing.T) {
		gatherer := world.CreateEntity()
		gatherComp := NewGatheringComponent()
		gatherComp.EquipTool(ToolTypeAxe) // Wrong tool
		gatherer.AddComponent(gatherComp)

		node := world.CreateEntity()
		nodeComp := NewResourceNodeComponent(ResourceTypeOre, "mountain") // Needs pickaxe
		node.AddComponent(nodeComp)

		entities := []*Entity{gatherer, node}
		system.Update(entities, 0)

		if system.StartGathering(gatherer, node.ID) {
			t.Error("Should fail with wrong tool")
		}
	})
}

func TestGatheringSystem_CancelGathering(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	gatherer := world.CreateEntity()
	gatherComp := NewGatheringComponent()
	gatherComp.StartGathering(456)
	gatherer.AddComponent(gatherComp)

	system.CancelGathering(gatherer)

	if gatherComp.IsCurrentlyGathering() {
		t.Error("Gathering should be canceled")
	}
}

func TestGatheringSystem_CancelGathering_NoComponent(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	gatherer := world.CreateEntity()

	// Should not panic
	system.CancelGathering(gatherer)
}

func TestGatheringSystem_GetNearbyResourceNodes(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	// Create nodes at different positions
	node1 := world.CreateEntity()
	node1.AddComponent(NewResourceNodeComponent(ResourceTypeOre, "mountain"))
	node1.AddComponent(&PositionComponent{X: 0, Y: 0})

	node2 := world.CreateEntity()
	node2.AddComponent(NewResourceNodeComponent(ResourceTypeWood, "forest"))
	node2.AddComponent(&PositionComponent{X: 5, Y: 0})

	node3 := world.CreateEntity()
	node3.AddComponent(NewResourceNodeComponent(ResourceTypeHerb, "plains"))
	node3.AddComponent(&PositionComponent{X: 100, Y: 100}) // Far away

	entities := []*Entity{node1, node2, node3}
	system.Update(entities, 0) // Index nodes

	// Find nodes within range 10
	nearby := system.GetNearbyResourceNodes(0, 0, 10)
	if len(nearby) != 2 {
		t.Errorf("Found %d nearby nodes, want 2", len(nearby))
	}

	// Find all nodes with large range
	all := system.GetNearbyResourceNodes(0, 0, 1000)
	if len(all) != 3 {
		t.Errorf("Found %d nodes, want 3", len(all))
	}
}

func TestGatheringSystem_GenerateResourceNode(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	seed := int64(12345)
	node := system.GenerateResourceNode(seed, ResourceTypeOre, "mountain", 100, 200)

	if node == nil {
		t.Fatal("GenerateResourceNode returned nil")
	}

	// Check position
	posCompRaw, ok := node.GetComponent("position")
	if !ok {
		t.Fatal("Node missing position component")
	}
	posComp, ok := posCompRaw.(*PositionComponent)
	if !ok || posComp == nil {
		t.Fatal("Node position component wrong type")
	}
	if posComp.X != 100 || posComp.Y != 200 {
		t.Errorf("Position = (%v, %v), want (100, 200)", posComp.X, posComp.Y)
	}

	// Check resource node
	nodeCompRaw, ok := node.GetComponent("resource_node")
	if !ok {
		t.Fatal("Node missing resource_node component")
	}
	nodeComp, ok := nodeCompRaw.(*ResourceNodeComponent)
	if !ok || nodeComp == nil {
		t.Fatal("Node resource_node component wrong type")
	}
	if nodeComp.ResourceType != ResourceTypeOre {
		t.Errorf("ResourceType = %v, want ore", nodeComp.ResourceType)
	}
	if nodeComp.BiomeType != "mountain" {
		t.Errorf("BiomeType = %v, want mountain", nodeComp.BiomeType)
	}

	// Verify determinism - same seed should give same results
	node2 := system.GenerateResourceNode(seed, ResourceTypeOre, "mountain", 100, 200)
	nodeComp2Raw, _ := node2.GetComponent("resource_node")
	nodeComp2 := nodeComp2Raw.(*ResourceNodeComponent)

	if nodeComp.MaxQuantity != nodeComp2.MaxQuantity {
		t.Error("Same seed should produce same MaxQuantity")
	}
	if nodeComp.MinSkillLevel != nodeComp2.MinSkillLevel {
		t.Error("Same seed should produce same MinSkillLevel")
	}
	if nodeComp.YieldMax != nodeComp2.YieldMax {
		t.Error("Same seed should produce same YieldMax")
	}
}

func TestGatheringSystem_GenerateResourceNode_Determinism(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	// Generate multiple nodes with same seed
	for i := 0; i < 10; i++ {
		seed := int64(99999)
		node1 := system.GenerateResourceNode(seed, ResourceTypeHerb, "plains", 0, 0)
		node2 := system.GenerateResourceNode(seed, ResourceTypeHerb, "plains", 0, 0)

		comp1Raw, _ := node1.GetComponent("resource_node")
		comp1 := comp1Raw.(*ResourceNodeComponent)
		comp2Raw, _ := node2.GetComponent("resource_node")
		comp2 := comp2Raw.(*ResourceNodeComponent)

		if comp1.MaxQuantity != comp2.MaxQuantity {
			t.Errorf("Iteration %d: MaxQuantity not deterministic", i)
		}
		if comp1.RespawnTime != comp2.RespawnTime {
			t.Errorf("Iteration %d: RespawnTime not deterministic", i)
		}
		if comp1.MinSkillLevel != comp2.MinSkillLevel {
			t.Errorf("Iteration %d: MinSkillLevel not deterministic", i)
		}
	}
}

func TestGatheringSystem_GetResourceNodeCount(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	if count := system.GetResourceNodeCount(); count != 0 {
		t.Errorf("Initial count = %v, want 0", count)
	}

	// Add nodes
	node1 := world.CreateEntity()
	node1.AddComponent(NewResourceNodeComponent(ResourceTypeOre, "mountain"))
	node2 := world.CreateEntity()
	node2.AddComponent(NewResourceNodeComponent(ResourceTypeWood, "forest"))

	entities := []*Entity{node1, node2}
	system.Update(entities, 0)

	if count := system.GetResourceNodeCount(); count != 2 {
		t.Errorf("After adding nodes, count = %v, want 2", count)
	}
}

func TestGatheringSystem_XPAndLevelUp(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)
	system.BaseGatherTime = 0.1 // Very fast for test
	system.XPPerHarvest = 200   // Enough to level up

	leveledUp := false
	var newLevel int
	system.OnLevelUpCallback = func(entity *Entity, level int) {
		leveledUp = true
		newLevel = level
	}

	// Create gatherer at level 1
	gatherer := world.CreateEntity()
	gatherComp := NewGatheringComponent()
	gatherComp.GatheringSkill = 1
	gatherComp.GatherSpeed = 100.0 // Very fast
	gatherComp.EquipTool(ToolTypePickaxe)
	gatherer.AddComponent(gatherComp)

	// Create resource node
	node := world.CreateEntity()
	nodeComp := NewResourceNodeComponent(ResourceTypeOre, "mountain")
	node.AddComponent(nodeComp)
	node.AddComponent(&PositionComponent{X: 0, Y: 0})

	entities := []*Entity{gatherer, node}

	// Index and start gathering
	system.Update(entities, 0)
	system.StartGathering(gatherer, node.ID)

	// Complete the gather
	system.Update(entities, 1.0)

	if !leveledUp {
		t.Error("Should have leveled up with 200 XP")
	}
	if newLevel != 2 {
		t.Errorf("New level = %v, want 2", newLevel)
	}
}

func TestGatheringSystem_YieldCalculation(t *testing.T) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	// Test with different skill levels
	tests := []struct {
		name      string
		skill     int
		toolBonus float64
		yieldMin  int
		yieldMax  int
		expectMin int
	}{
		{"low skill", 1, 1.0, 1, 5, 1},
		{"medium skill", 50, 1.0, 1, 5, 2},
		{"high skill", 100, 1.0, 1, 5, 5},
		{"with tool bonus", 50, 2.0, 1, 5, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gatherComp := NewGatheringComponent()
			gatherComp.GatheringSkill = tt.skill
			gatherComp.SetToolBonus(ToolTypePickaxe, tt.toolBonus)

			nodeComp := NewResourceNodeComponent(ResourceTypeOre, "mountain")
			nodeComp.YieldMin = tt.yieldMin
			nodeComp.YieldMax = tt.yieldMax

			yield := system.calculateYield(gatherComp, nodeComp)
			if yield < tt.expectMin {
				t.Errorf("Yield = %v, expected at least %v", yield, tt.expectMin)
			}
		})
	}
}

func BenchmarkGatheringSystem_Update(b *testing.B) {
	world := NewWorld()
	system := NewGatheringSystem(world)

	// Create many entities
	var entities []*Entity
	for i := 0; i < 100; i++ {
		node := world.CreateEntity()
		node.AddComponent(NewResourceNodeComponent(ResourceTypeOre, "mountain"))
		node.AddComponent(&PositionComponent{X: float64(i * 10), Y: 0})
		entities = append(entities, node)

		gatherer := world.CreateEntity()
		gatherer.AddComponent(NewGatheringComponent())
		entities = append(entities, gatherer)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(entities, 0.016)
	}
}
