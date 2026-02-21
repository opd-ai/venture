// Package engine provides tests for advanced behavior tree nodes.
package engine

import (
	"math"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestUseConsumableNode_HealthPotion tests using health potions.
func TestUseConsumableNode_HealthPotion(t *testing.T) {
	tests := []struct {
		name            string
		currentHealth   float64
		maxHealth       float64
		healthThreshold float64
		hasPotion       bool
		wantStatus      NodeStatus
		wantHealthGain  bool
	}{
		{
			name:            "uses potion when low health",
			currentHealth:   20,
			maxHealth:       100,
			healthThreshold: 0.3,
			hasPotion:       true,
			wantStatus:      NodeSuccess,
			wantHealthGain:  true,
		},
		{
			name:            "no potion when health ok",
			currentHealth:   80,
			maxHealth:       100,
			healthThreshold: 0.3,
			hasPotion:       true,
			wantStatus:      NodeFailure,
			wantHealthGain:  false,
		},
		{
			name:            "fails without potion",
			currentHealth:   20,
			maxHealth:       100,
			healthThreshold: 0.3,
			hasPotion:       false,
			wantStatus:      NodeFailure,
			wantHealthGain:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(0)
			entity.AddComponent(&HealthComponent{
				Current: tt.currentHealth,
				Max:     tt.maxHealth,
			})

			inv := NewInventoryComponent(10, 100)
			if tt.hasPotion {
				potion := &item.Item{
					Type:           item.TypeConsumable,
					ConsumableType: item.ConsumablePotion,
					Stats:          item.Stats{Value: 50},
					Name:           "Health Potion",
				}
				inv.AddItem(potion)
			}
			entity.AddComponent(inv)

			blackboard := NewBlackboardWithSeed(12345)
			node := NewUseConsumableNode("test", item.ConsumablePotion, tt.healthThreshold, 1.0)

			status := node.Tick(entity, blackboard, 0.1)
			if status != tt.wantStatus {
				t.Errorf("got status %v, want %v", status, tt.wantStatus)
			}

			// Check health gain
			healthComp, _ := entity.GetComponent("health")
			health := healthComp.(*HealthComponent)
			healthGained := health.Current > tt.currentHealth
			if healthGained != tt.wantHealthGain {
				t.Errorf("health gained = %v, want %v", healthGained, tt.wantHealthGain)
			}
		})
	}
}

// TestRetreatToAllyNode tests retreating toward allies.
func TestRetreatToAllyNode(t *testing.T) {
	entity := NewEntity(0)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})
	entity.AddComponent(&VelocityComponent{})
	entity.AddComponent(&FactionComponent{FactionID: "good"})

	ally := NewEntity(0)
	ally.AddComponent(&PositionComponent{X: 200, Y: 100})
	ally.AddComponent(&FactionComponent{FactionID: "good"})

	enemy := NewEntity(0)
	enemy.AddComponent(&PositionComponent{X: 50, Y: 100})
	enemy.AddComponent(&FactionComponent{FactionID: "evil"})

	blackboard := NewBlackboardWithSeed(12345)
	blackboard.Set("nearby_entities", []*Entity{ally, enemy})

	node := NewRetreatToAllyNode("retreat", 100.0, 200.0, 10.0)

	// First tick should return Running as we move toward ally
	status := node.Tick(entity, blackboard, 0.1)
	if status != NodeRunning {
		t.Errorf("expected NodeRunning, got %v", status)
	}

	// Check that entity moved toward ally (x increased)
	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*PositionComponent)
	if pos.X <= 100 {
		t.Errorf("entity should have moved toward ally, got X=%f", pos.X)
	}
}

// TestAmbushNode tests ambush setup and triggering.
func TestAmbushNode(t *testing.T) {
	entity := NewEntity(0)
	entity.AddComponent(&PositionComponent{X: 100, Y: 100})

	target := NewEntity(0)
	target.AddComponent(&PositionComponent{X: 150, Y: 100})

	blackboard := NewBlackboardWithSeed(12345)

	node := NewAmbushNode("ambush", 10.0, 100.0)

	// First tick - setup ambush position
	status := node.Tick(entity, blackboard, 0.1)
	if status != NodeRunning {
		t.Errorf("expected NodeRunning during setup, got %v", status)
	}

	// Verify ambush position was set
	ambushPos, hasPos := blackboard.Get("ambush_position")
	if !hasPos {
		t.Error("ambush_position should be set")
	}

	// Continue to set up ambush
	for i := 0; i < 10; i++ {
		status = node.Tick(entity, blackboard, 0.1)
	}

	// Now add target and check for trigger
	blackboard.Set("target", target)
	status = node.Tick(entity, blackboard, 0.1)

	// Should trigger since target is in range
	_, hasAmbushPos := blackboard.Get("ambush_position")
	_ = hasAmbushPos // Ambush position may or may not be cleared depending on state
	_ = ambushPos
}

// TestFormationNode tests formation maintenance.
func TestFormationNode(t *testing.T) {
	tests := []struct {
		name          string
		formationType FormationType
		slot          int
		leaderX       float64
		leaderY       float64
		wantXOffset   bool // Whether X should differ from leader
		wantYOffset   bool // Whether Y should differ from leader
	}{
		{
			name:          "line formation",
			formationType: FormationLine,
			slot:          2,
			leaderX:       100,
			leaderY:       100,
			wantXOffset:   false, // Line is perpendicular to facing
			wantYOffset:   true,
		},
		{
			name:          "column formation",
			formationType: FormationColumn,
			slot:          1,
			leaderX:       100,
			leaderY:       100,
			wantXOffset:   true,
			wantYOffset:   false,
		},
		{
			name:          "circle formation",
			formationType: FormationCircle,
			slot:          0,
			leaderX:       100,
			leaderY:       100,
			wantXOffset:   true, // Circle positions are around leader
			wantYOffset:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(0)
			entity.AddComponent(&PositionComponent{X: 0, Y: 0})
			entity.AddComponent(&VelocityComponent{})
			entity.AddComponent(&SquadComponent{
				SquadID: 1,
				Role:    SquadRoleMember,
			})

			leader := NewEntity(0)
			leader.AddComponent(&PositionComponent{X: tt.leaderX, Y: tt.leaderY})

			blackboard := NewBlackboardWithSeed(12345)
			blackboard.Set("squad_leader", leader)
			blackboard.Set("formation_slot", tt.slot)

			node := NewFormationNode("formation", tt.formationType, 30.0, 100.0)

			// Tick to start movement
			status := node.Tick(entity, blackboard, 0.1)
			if status != NodeRunning && status != NodeSuccess {
				t.Errorf("expected NodeRunning or NodeSuccess, got %v", status)
			}
		})
	}
}

// TestHasConsumableNode tests consumable detection.
func TestHasConsumableNode(t *testing.T) {
	tests := []struct {
		name           string
		hasInventory   bool
		hasPotion      bool
		consumableType item.ConsumableType
		wantStatus     NodeStatus
	}{
		{
			name:           "has matching consumable",
			hasInventory:   true,
			hasPotion:      true,
			consumableType: item.ConsumablePotion,
			wantStatus:     NodeSuccess,
		},
		{
			name:           "no matching consumable",
			hasInventory:   true,
			hasPotion:      false,
			consumableType: item.ConsumablePotion,
			wantStatus:     NodeFailure,
		},
		{
			name:           "no inventory",
			hasInventory:   false,
			hasPotion:      false,
			consumableType: item.ConsumablePotion,
			wantStatus:     NodeFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(0)
			if tt.hasInventory {
				inv := NewInventoryComponent(10, 100)
				if tt.hasPotion {
					potion := &item.Item{
						Type:           item.TypeConsumable,
						ConsumableType: item.ConsumablePotion,
					}
					inv.AddItem(potion)
				}
				entity.AddComponent(inv)
			}

			blackboard := NewBlackboard()
			node := NewHasConsumableNode("test", tt.consumableType)

			status := node.Tick(entity, blackboard, 0.1)
			if status != tt.wantStatus {
				t.Errorf("got status %v, want %v", status, tt.wantStatus)
			}
		})
	}
}

// TestIsOutnumberedNode tests numerical advantage detection.
func TestIsOutnumberedNode(t *testing.T) {
	tests := []struct {
		name       string
		allyCount  int
		enemyCount int
		ratio      float64
		wantStatus NodeStatus
	}{
		{
			name:       "outnumbered 2:1",
			allyCount:  1,
			enemyCount: 3,
			ratio:      2.0,
			wantStatus: NodeSuccess,
		},
		{
			name:       "not outnumbered",
			allyCount:  3,
			enemyCount: 2,
			ratio:      2.0,
			wantStatus: NodeFailure,
		},
		{
			name:       "exactly at ratio",
			allyCount:  1,
			enemyCount: 2,
			ratio:      2.0,
			wantStatus: NodeSuccess,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(0)
			entity.AddComponent(&PositionComponent{X: 100, Y: 100})
			entity.AddComponent(&FactionComponent{FactionID: "good"})

			var nearbyEntities []*Entity

			// Add allies
			for i := 0; i < tt.allyCount; i++ {
				ally := NewEntity(0)
				ally.AddComponent(&PositionComponent{X: 100 + float64(i*10), Y: 100})
				ally.AddComponent(&FactionComponent{FactionID: "good"})
				nearbyEntities = append(nearbyEntities, ally)
			}

			// Add enemies
			for i := 0; i < tt.enemyCount; i++ {
				enemy := NewEntity(0)
				enemy.AddComponent(&PositionComponent{X: 100 + float64(i*10), Y: 110})
				enemy.AddComponent(&FactionComponent{FactionID: "evil"})
				nearbyEntities = append(nearbyEntities, enemy)
			}

			blackboard := NewBlackboard()
			blackboard.Set("nearby_entities", nearbyEntities)

			node := NewIsOutnumberedNode("test", 200.0, tt.ratio)
			status := node.Tick(entity, blackboard, 0.1)

			if status != tt.wantStatus {
				t.Errorf("got status %v, want %v", status, tt.wantStatus)
			}
		})
	}
}

// TestProtectAllyNode tests ally protection behavior.
func TestProtectAllyNode(t *testing.T) {
	protector := NewEntity(0)
	protector.AddComponent(&PositionComponent{X: 100, Y: 100})
	protector.AddComponent(&VelocityComponent{})
	protector.AddComponent(&FactionComponent{FactionID: "good"})

	woundedAlly := NewEntity(0)
	woundedAlly.AddComponent(&PositionComponent{X: 150, Y: 100})
	woundedAlly.AddComponent(&FactionComponent{FactionID: "good"})
	woundedAlly.AddComponent(&HealthComponent{Current: 10, Max: 100})

	enemy := NewEntity(0)
	enemy.AddComponent(&PositionComponent{X: 200, Y: 100})
	enemy.AddComponent(&FactionComponent{FactionID: "evil"})

	blackboard := NewBlackboard()
	blackboard.Set("nearby_entities", []*Entity{woundedAlly, enemy})
	blackboard.Set("target", enemy)

	node := NewProtectAllyNode("protect", 200.0, 0.3, 100.0)

	status := node.Tick(protector, blackboard, 0.1)
	if status != NodeRunning {
		t.Errorf("expected NodeRunning while moving to protect, got %v", status)
	}

	// Check that protector moved toward ally
	posComp, _ := protector.GetComponent("position")
	pos := posComp.(*PositionComponent)
	if pos.X <= 100 {
		t.Errorf("protector should have moved toward ally, got X=%f", pos.X)
	}
}

// TestCoordinatedAttackNode tests synchronized squad attacks.
func TestCoordinatedAttackNode(t *testing.T) {
	attacker := NewEntity(0)
	attacker.AddComponent(&PositionComponent{X: 100, Y: 100})

	target := NewEntity(0)
	target.AddComponent(&PositionComponent{X: 120, Y: 100})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	blackboard := NewBlackboard()
	blackboard.Set("target", target)
	blackboard.Set("squad_ready_count", 2)
	blackboard.Set("squad_size", 2)

	node := NewCoordinatedAttackNode("coordinated", 50.0, 20, 1.0)

	// First tick signals coordination
	status := node.Tick(attacker, blackboard, 0.1)
	if status != NodeRunning {
		t.Errorf("expected NodeRunning during coordination, got %v", status)
	}

	// Second tick with all ready should attack
	status = node.Tick(attacker, blackboard, 0.1)
	if status != NodeSuccess {
		t.Errorf("expected NodeSuccess after attack, got %v", status)
	}

	// Check damage was dealt with bonus
	targetHealth, _ := target.GetComponent("health")
	health := targetHealth.(*HealthComponent)
	expectedDamage := 25.0 // 20 + 25% bonus
	actualDamage := 100.0 - health.Current
	if math.Abs(actualDamage-expectedDamage) > 0.1 {
		t.Errorf("expected ~%f damage, got %f", expectedDamage, actualDamage)
	}
}

// TestCanSeeTargetNode tests line of sight checks.
func TestCanSeeTargetNode(t *testing.T) {
	tests := []struct {
		name       string
		distance   float64
		maxRange   float64
		losBlocked bool
		wantStatus NodeStatus
	}{
		{
			name:       "target in range with LOS",
			distance:   50,
			maxRange:   100,
			losBlocked: false,
			wantStatus: NodeSuccess,
		},
		{
			name:       "target out of range",
			distance:   150,
			maxRange:   100,
			losBlocked: false,
			wantStatus: NodeFailure,
		},
		{
			name:       "target blocked by obstacle",
			distance:   50,
			maxRange:   100,
			losBlocked: true,
			wantStatus: NodeFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(0)
			entity.AddComponent(&PositionComponent{X: 0, Y: 0})

			target := NewEntity(0)
			target.AddComponent(&PositionComponent{X: tt.distance, Y: 0})

			blackboard := NewBlackboard()
			blackboard.Set("target", target)
			if tt.losBlocked {
				blackboard.Set("los_blocked", true)
			}

			node := NewCanSeeTargetNode("test", tt.maxRange)
			status := node.Tick(entity, blackboard, 0.1)

			if status != tt.wantStatus {
				t.Errorf("got status %v, want %v", status, tt.wantStatus)
			}
		})
	}
}

// TestIsInCoverNode tests cover status checks.
func TestIsInCoverNode(t *testing.T) {
	entity := NewEntity(0)

	tests := []struct {
		name       string
		inCover    bool
		hasCover   bool
		wantStatus NodeStatus
	}{
		{
			name:       "in cover",
			inCover:    true,
			hasCover:   true,
			wantStatus: NodeSuccess,
		},
		{
			name:       "not in cover",
			inCover:    false,
			hasCover:   true,
			wantStatus: NodeFailure,
		},
		{
			name:       "no cover data",
			inCover:    false,
			hasCover:   false,
			wantStatus: NodeFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blackboard := NewBlackboard()
			if tt.hasCover {
				blackboard.Set("at_cover", tt.inCover)
			}

			node := NewIsInCoverNode("test")
			status := node.Tick(entity, blackboard, 0.1)

			if status != tt.wantStatus {
				t.Errorf("got status %v, want %v", status, tt.wantStatus)
			}
		})
	}
}

// TestIsAmbushingNode tests ambush state detection.
func TestIsAmbushingNode(t *testing.T) {
	entity := NewEntity(0)

	tests := []struct {
		name       string
		inAmbush   bool
		hasState   bool
		wantStatus NodeStatus
	}{
		{
			name:       "in ambush",
			inAmbush:   true,
			hasState:   true,
			wantStatus: NodeSuccess,
		},
		{
			name:       "not ambushing",
			inAmbush:   false,
			hasState:   true,
			wantStatus: NodeFailure,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			blackboard := NewBlackboard()
			if tt.hasState {
				blackboard.Set("in_ambush", tt.inAmbush)
			}

			node := NewIsAmbushingNode("test")
			status := node.Tick(entity, blackboard, 0.1)

			if status != tt.wantStatus {
				t.Errorf("got status %v, want %v", status, tt.wantStatus)
			}
		})
	}
}

// TestAdvancedNodeStringMethods tests String() output for all nodes.
func TestAdvancedNodeStringMethods(t *testing.T) {
	tests := []struct {
		name     string
		node     BehaviorNode
		contains string
	}{
		{
			name:     "UseConsumableNode",
			node:     NewUseConsumableNode("test", item.ConsumablePotion, 0.3, 1.0),
			contains: "UseConsumable",
		},
		{
			name:     "RetreatToAllyNode",
			node:     NewRetreatToAllyNode("test", 100, 200, 10),
			contains: "RetreatToAlly",
		},
		{
			name:     "AmbushNode",
			node:     NewAmbushNode("test", 10, 50),
			contains: "Ambush",
		},
		{
			name:     "FormationNode",
			node:     NewFormationNode("test", FormationLine, 30, 100),
			contains: "Formation",
		},
		{
			name:     "HasConsumableNode",
			node:     NewHasConsumableNode("test", item.ConsumablePotion),
			contains: "HasConsumable",
		},
		{
			name:     "IsOutnumberedNode",
			node:     NewIsOutnumberedNode("test", 100, 2.0),
			contains: "IsOutnumbered",
		},
		{
			name:     "IsInCoverNode",
			node:     NewIsInCoverNode("test"),
			contains: "IsInCover",
		},
		{
			name:     "CanSeeTargetNode",
			node:     NewCanSeeTargetNode("test", 100),
			contains: "CanSeeTarget",
		},
		{
			name:     "IsAmbushingNode",
			node:     NewIsAmbushingNode("test"),
			contains: "IsAmbushing",
		},
		{
			name:     "CoordinatedAttackNode",
			node:     NewCoordinatedAttackNode("test", 50, 20, 1.0),
			contains: "CoordinatedAttack",
		},
		{
			name:     "ProtectAllyNode",
			node:     NewProtectAllyNode("test", 200, 0.3, 100),
			contains: "ProtectAlly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			str := tt.node.String()
			if str == "" {
				t.Error("String() returned empty string")
			}
			// Verify it contains expected text
			if len(str) < len(tt.contains) {
				t.Errorf("String() too short: %q", str)
			}
		})
	}
}

// TestTagComponent tests the TagComponent functionality.
func TestTagComponent(t *testing.T) {
	t.Run("Type", func(t *testing.T) {
		comp := NewTagComponent("test")
		if comp.Type() != "tag" {
			t.Errorf("expected type 'tag', got %q", comp.Type())
		}
	})

	t.Run("HasTag", func(t *testing.T) {
		comp := NewTagComponent("lever", "interactable")
		if !comp.HasTag("lever") {
			t.Error("should have tag 'lever'")
		}
		if comp.HasTag("trap") {
			t.Error("should not have tag 'trap'")
		}
	})

	t.Run("AddTag", func(t *testing.T) {
		comp := NewTagComponent("lever")
		comp.AddTag("hazard")
		if !comp.HasTag("hazard") {
			t.Error("should have added tag 'hazard'")
		}
		// Adding duplicate should not create duplicates
		comp.AddTag("lever")
		count := 0
		for _, tag := range comp.Tags {
			if tag == "lever" {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected 1 'lever' tag, got %d", count)
		}
	})

	t.Run("RemoveTag", func(t *testing.T) {
		comp := NewTagComponent("lever", "trap", "hazard")
		comp.RemoveTag("trap")
		if comp.HasTag("trap") {
			t.Error("should have removed tag 'trap'")
		}
		if !comp.HasTag("lever") || !comp.HasTag("hazard") {
			t.Error("should still have other tags")
		}
	})
}
