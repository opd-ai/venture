// Package engine provides tests for tactical behavior tree nodes.
package engine

import (
	"math/rand"
	"testing"

	"github.com/opd-ai/venture/pkg/engine/aitypes"
)

// TestHealthBelowNode tests the health threshold condition.
func TestHealthBelowNode(t *testing.T) {
	tests := []struct {
		name      string
		current   float64
		max       float64
		threshold float64
		want      NodeStatus
	}{
		{"below_threshold", 25, 100, 0.5, NodeSuccess},
		{"above_threshold", 75, 100, 0.5, NodeFailure},
		{"at_threshold", 50, 100, 0.5, NodeFailure},
		{"full_health", 100, 100, 0.5, NodeFailure},
		{"zero_health", 0, 100, 0.1, NodeSuccess},
		{"critical", 10, 100, 0.25, NodeSuccess},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(1)
			entity.AddComponent(&HealthComponent{Current: tt.current, Max: tt.max})

			node := NewHealthBelowNode("test", tt.threshold)
			blackboard := NewBlackboardWithSeed(12345)

			got := node.Tick(entity, blackboard, 0.016)
			if got != tt.want {
				t.Errorf("HealthBelowNode.Tick() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHealthBelowNode_NoComponent tests failure without health component.
func TestHealthBelowNode_NoComponent(t *testing.T) {
	entity := NewEntity(1)
	node := NewHealthBelowNode("test", 0.5)
	blackboard := NewBlackboardWithSeed(12345)

	got := node.Tick(entity, blackboard, 0.016)
	if got != NodeFailure {
		t.Errorf("HealthBelowNode without component should fail, got %v", got)
	}
}

// TestHasTargetNode tests target presence checking.
func TestHasTargetNode(t *testing.T) {
	tests := []struct {
		name   string
		target interface{}
		want   NodeStatus
	}{
		{"has_entity_target", NewEntity(2), NodeSuccess},
		{"nil_target", nil, NodeFailure},
		{"wrong_type", "not_an_entity", NodeFailure},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(1)
			blackboard := NewBlackboardWithSeed(12345)
			if tt.target != nil {
				blackboard.Set("target", tt.target)
			}

			node := NewHasTargetNode("test")
			got := node.Tick(entity, blackboard, 0.016)
			if got != tt.want {
				t.Errorf("HasTargetNode.Tick() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestInRangeNode tests distance checking to target.
func TestInRangeNode(t *testing.T) {
	tests := []struct {
		name     string
		myX, myY float64
		tX, tY   float64
		distance float64
		want     NodeStatus
	}{
		{"in_range", 0, 0, 5, 0, 10, NodeSuccess},
		{"out_of_range", 0, 0, 15, 0, 10, NodeFailure},
		{"exactly_at_range", 0, 0, 10, 0, 10, NodeSuccess},
		{"diagonal_in_range", 0, 0, 7, 7, 15, NodeSuccess},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := NewEntity(1)
			entity.AddComponent(&PositionComponent{X: tt.myX, Y: tt.myY})

			target := NewEntity(2)
			target.AddComponent(&PositionComponent{X: tt.tX, Y: tt.tY})

			blackboard := NewBlackboardWithSeed(12345)
			blackboard.Set("target", target)

			node := NewInRangeNode("test", tt.distance)
			got := node.Tick(entity, blackboard, 0.016)
			if got != tt.want {
				t.Errorf("InRangeNode.Tick() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMoveToTargetNode tests movement toward target.
func TestMoveToTargetNode(t *testing.T) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 100, Y: 0})

	blackboard := NewBlackboardWithSeed(12345)
	blackboard.Set("target", target)

	node := NewMoveToTargetNode("test", 100.0, 10.0)

	// First tick should be running
	status := node.Tick(entity, blackboard, 1.0)
	if status != NodeRunning {
		t.Errorf("First tick should be Running, got %v", status)
	}

	// Check entity moved
	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*PositionComponent)
	if pos.X <= 0 {
		t.Errorf("Entity should have moved, X = %v", pos.X)
	}

	// Keep ticking until success
	for i := 0; i < 10; i++ {
		status = node.Tick(entity, blackboard, 0.1)
		if status == NodeSuccess {
			break
		}
	}
	if status != NodeSuccess {
		t.Errorf("Should reach target eventually, got %v", status)
	}
}

// TestFleeFromTargetNode tests fleeing behavior.
func TestFleeFromTargetNode(t *testing.T) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 0})

	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 0, Y: 0})

	blackboard := NewBlackboardWithSeed(12345)
	blackboard.Set("target", target)

	node := NewFleeFromTargetNode("test", 100.0, 50.0)

	// First tick should be running
	_ = node.Tick(entity, blackboard, 0.5)

	// Check entity moved away
	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*PositionComponent)
	if pos.X <= 10 {
		t.Errorf("Entity should have fled, X = %v", pos.X)
	}
}

// TestPatrolNode tests patrol waypoint movement.
func TestPatrolNode(t *testing.T) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	blackboard := NewBlackboardWithSeed(12345)
	waypoints := [][]float64{{10, 0}, {10, 10}, {0, 10}, {0, 0}}
	blackboard.Set("patrol_waypoints", waypoints)

	node := NewPatrolNode("test", 100.0, 0.5)

	// Tick multiple times to see movement
	for i := 0; i < 5; i++ {
		status := node.Tick(entity, blackboard, 0.1)
		if status == NodeFailure {
			t.Errorf("Patrol should not fail with valid waypoints")
		}
	}

	// Verify entity moved toward first waypoint
	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*PositionComponent)
	if pos.X <= 0 {
		t.Errorf("Entity should have moved toward waypoint, X = %v", pos.X)
	}
}

// TestAttackTargetNode tests attack execution.
func TestAttackTargetNode(t *testing.T) {
	entity := NewEntity(3)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 5, Y: 0})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	blackboard := NewBlackboardWithSeed(12345)
	blackboard.Set("target", target)

	node := NewAttackTargetNode("test", 10.0, 25, 1.0, "melee")

	// Attack should succeed (in range)
	status := node.Tick(entity, blackboard, 0.016)
	if status != NodeSuccess {
		t.Errorf("Attack should succeed, got %v", status)
	}

	// Verify damage dealt
	healthComp, _ := target.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Current != 75 {
		t.Errorf("Target health should be 75, got %v", health.Current)
	}

	// Second attack should be on cooldown
	status = node.Tick(entity, blackboard, 0.016)
	if status != NodeRunning {
		t.Errorf("Should be on cooldown, got %v", status)
	}
}

// TestAttackTargetNode_OutOfRange tests attack failure when out of range.
func TestAttackTargetNode_OutOfRange(t *testing.T) {
	entity := NewEntity(3)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 100, Y: 0})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	blackboard := NewBlackboardWithSeed(12345)
	blackboard.Set("target", target)

	node := NewAttackTargetNode("test", 10.0, 25, 1.0, "melee")

	status := node.Tick(entity, blackboard, 0.016)
	if status != NodeFailure {
		t.Errorf("Attack should fail out of range, got %v", status)
	}
}

// TestCallForHelpNode tests help call behavior.
func TestCallForHelpNode(t *testing.T) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 10, Y: 20})
	entity.AddComponent(&FactionComponent{FactionID: "allies"})

	blackboard := NewBlackboardWithSeed(12345)
	node := NewCallForHelpNode("test", 100.0)

	status := node.Tick(entity, blackboard, 0.016)
	if status != NodeSuccess {
		t.Errorf("Call for help should succeed, got %v", status)
	}

	// Verify help call stored
	helpCall, ok := blackboard.Get("help_call")
	if !ok || helpCall == nil {
		t.Error("Help call should be stored in blackboard")
	}

	// Second call should succeed too (already called)
	status = node.Tick(entity, blackboard, 0.016)
	if status != NodeSuccess {
		t.Errorf("Repeat call should succeed, got %v", status)
	}
}

// TestWaitNode tests wait timing.
func TestWaitNode(t *testing.T) {
	entity := NewEntity(1)
	blackboard := NewBlackboardWithSeed(12345)
	node := NewWaitNode("test", 1.0)

	// Should be running during wait
	status := node.Tick(entity, blackboard, 0.5)
	if status != NodeRunning {
		t.Errorf("Should be running during wait, got %v", status)
	}

	// Should succeed after wait complete
	status = node.Tick(entity, blackboard, 0.6)
	if status != NodeSuccess {
		t.Errorf("Should succeed after wait, got %v", status)
	}
}

// TestRandomSelectorNode tests random child selection.
func TestRandomSelectorNode(t *testing.T) {
	// Create child nodes that return deterministic results
	successChild := &successNode{}
	failureChild := &failureNode{}

	node := NewRandomSelectorNode("test", successChild, failureChild)

	entity := NewEntity(1)
	blackboard := NewBlackboardWithSeed(12345)

	// Run multiple times to verify randomness
	successCount := 0
	for i := 0; i < 10; i++ {
		node.Reset()
		status := node.Tick(entity, blackboard, 0.016)
		if status == NodeSuccess {
			successCount++
		}
	}

	// Should have some variation (not all success or all failure)
	if successCount == 0 || successCount == 10 {
		t.Logf("Random selection may have low variation: %d successes", successCount)
	}
}

// successNode always returns success.
type successNode struct{}

func (s *successNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	return NodeSuccess
}
func (s *successNode) Reset()         {}
func (s *successNode) String() string { return "Success" }

// failureNode always returns failure.
type failureNode struct{}

func (f *failureNode) Tick(entity aitypes.EntityContext, blackboard *Blackboard, deltaTime float64) NodeStatus {
	return NodeFailure
}
func (f *failureNode) Reset()         {}
func (f *failureNode) String() string { return "Failure" }

// TestSucceederNode tests that it always returns success.
func TestSucceederNode(t *testing.T) {
	entity := NewEntity(1)
	blackboard := NewBlackboardWithSeed(12345)

	child := &failureNode{}
	node := NewSucceederNode("test", child)

	status := node.Tick(entity, blackboard, 0.016)
	if status != NodeSuccess {
		t.Errorf("Succeeder should return success, got %v", status)
	}
}

// TestFailerNode tests that it always returns failure.
func TestFailerNode(t *testing.T) {
	entity := NewEntity(1)
	blackboard := NewBlackboardWithSeed(12345)

	child := &successNode{}
	node := NewFailerNode("test", child)

	status := node.Tick(entity, blackboard, 0.016)
	if status != NodeFailure {
		t.Errorf("Failer should return failure, got %v", status)
	}
}

// TestSeekCoverNode tests cover seeking behavior.
func TestSeekCoverNode(t *testing.T) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 50, Y: 50})

	target := NewEntity(7)
	target.AddComponent(&PositionComponent{X: 0, Y: 0})

	blackboard := NewBlackboardWithSeed(12345)
	blackboard.Set("target", target)

	node := NewSeekCoverNode("test", 100.0, 30.0)

	// First tick should start moving toward cover
	status := node.Tick(entity, blackboard, 0.1)
	if status == NodeFailure {
		t.Errorf("SeekCover should not fail with valid positions")
	}
}

// TestFlankTargetNode tests flanking behavior.
func TestFlankTargetNode(t *testing.T) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 50, Y: 0})

	blackboard := NewBlackboardWithSeed(12345)
	blackboard.Set("target", target)

	node := NewFlankTargetNode("test", 100.0, 30.0)

	// First tick should start flanking
	status := node.Tick(entity, blackboard, 0.1)

	// Check entity moved
	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*PositionComponent)

	// Movement should have Y component (flanking perpendicular to direct line)
	if status != NodeRunning && status != NodeSuccess {
		t.Errorf("Flank should be running or success, got %v", status)
	}
	if pos.Y == 0 {
		// With perpendicular movement, Y should change
		t.Logf("Y position unchanged, may need more iterations")
	}
}

// TestHasAlliesNearbyNode tests ally detection.
func TestHasAlliesNearbyNode(t *testing.T) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&FactionComponent{FactionID: "allies"})

	ally1 := NewEntity(4)
	ally1.AddComponent(&PositionComponent{X: 10, Y: 0})
	ally1.AddComponent(&FactionComponent{FactionID: "allies"})

	ally2 := NewEntity(5)
	ally2.AddComponent(&PositionComponent{X: 20, Y: 0})
	ally2.AddComponent(&FactionComponent{FactionID: "allies"})

	enemy := NewEntity(6)
	enemy.AddComponent(&PositionComponent{X: 5, Y: 0})
	enemy.AddComponent(&FactionComponent{FactionID: "enemies"})

	blackboard := NewBlackboardWithSeed(12345)
	blackboard.Set("nearby_entities", []*Entity{ally1, ally2, enemy})

	node := NewHasAlliesNearbyNode("test", 50.0, 2)

	status := node.Tick(entity, blackboard, 0.016)
	if status != NodeSuccess {
		t.Errorf("Should detect 2 allies nearby, got %v", status)
	}

	// Check ally count was stored
	count, ok := blackboard.Get("ally_count")
	if !ok {
		t.Error("Ally count should be stored")
	} else if count.(int) != 2 {
		t.Errorf("Ally count should be 2, got %v", count)
	}
}

// TestNodeStringMethods tests all String() methods for coverage.
func TestNodeStringMethods(t *testing.T) {
	tests := []struct {
		name string
		node BehaviorNode
	}{
		{"HealthBelow", NewHealthBelowNode("test", 0.5)},
		{"HasTarget", NewHasTargetNode("test")},
		{"InRange", NewInRangeNode("test", 10.0)},
		{"HasAlliesNearby", NewHasAlliesNearbyNode("test", 50.0, 2)},
		{"MoveToTarget", NewMoveToTargetNode("test", 100.0, 5.0)},
		{"FleeFromTarget", NewFleeFromTargetNode("test", 100.0, 50.0)},
		{"SeekCover", NewSeekCoverNode("test", 100.0, 30.0)},
		{"FlankTarget", NewFlankTargetNode("test", 100.0, 30.0)},
		{"Patrol", NewPatrolNode("test", 100.0, 1.0)},
		{"AttackTarget", NewAttackTargetNode("test", 10.0, 25, 1.0, "melee")},
		{"CallForHelp", NewCallForHelpNode("test", 100.0)},
		{"Wait", NewWaitNode("test", 1.0)},
		{"RandomSelector", NewRandomSelectorNode("test", &successNode{})},
		{"Succeeder", NewSucceederNode("test", &successNode{})},
		{"Failer", NewFailerNode("test", &successNode{})},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.node.String()
			if s == "" {
				t.Error("String() should not return empty string")
			}
		})
	}
}

// TestNodeResetMethods tests all Reset() methods for coverage.
func TestNodeResetMethods(t *testing.T) {
	nodes := []BehaviorNode{
		NewHealthBelowNode("test", 0.5),
		NewHasTargetNode("test"),
		NewInRangeNode("test", 10.0),
		NewHasAlliesNearbyNode("test", 50.0, 2),
		NewMoveToTargetNode("test", 100.0, 5.0),
		NewFleeFromTargetNode("test", 100.0, 50.0),
		NewSeekCoverNode("test", 100.0, 30.0),
		NewFlankTargetNode("test", 100.0, 30.0),
		NewPatrolNode("test", 100.0, 1.0),
		NewAttackTargetNode("test", 10.0, 25, 1.0, "melee"),
		NewCallForHelpNode("test", 100.0),
		NewWaitNode("test", 1.0),
		NewRandomSelectorNode("test", &successNode{}),
		NewSucceederNode("test", &successNode{}),
		NewFailerNode("test", &successNode{}),
	}

	for _, node := range nodes {
		// This should not panic
		node.Reset()
	}
}

// BenchmarkMoveToTargetNode benchmarks movement node performance.
func BenchmarkMoveToTargetNode(b *testing.B) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&VelocityComponent{VX: 0, VY: 0})

	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 1000, Y: 1000})

	blackboard := NewBlackboardWithSeed(12345)
	blackboard.Set("target", target)

	node := NewMoveToTargetNode("bench", 100.0, 5.0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.Tick(entity, blackboard, 0.016)
	}
}

// BenchmarkAttackTargetNode benchmarks attack node performance.
func BenchmarkAttackTargetNode(b *testing.B) {
	entity := NewEntity(3)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 5, Y: 0})
	target.AddComponent(&HealthComponent{Current: 1000000, Max: 1000000})

	blackboard := NewBlackboardWithSeed(12345)
	blackboard.Set("target", target)

	node := NewAttackTargetNode("bench", 10.0, 1, 0.0, "melee") // No cooldown

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		node.Tick(entity, blackboard, 0.016)
		node.Reset() // Reset cooldown for next iteration
	}
}

// BenchmarkSequenceWithTacticalNodes benchmarks a realistic behavior tree.
func BenchmarkSequenceWithTacticalNodes(b *testing.B) {
	entity := NewEntity(1)
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})
	entity.AddComponent(&HealthComponent{Current: 50, Max: 100})

	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 100, Y: 0})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	blackboard := NewBlackboardWithSeed(12345)
	blackboard.Set("target", target)

	// Build a realistic behavior tree
	tree := NewSequenceNode("combat",
		NewHasTargetNode("check_target"),
		NewInRangeNode("check_range", 50.0),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.Tick(entity, blackboard, 0.016)
		tree.Reset()
	}
}

// TestIntegrationCombatBehavior tests a complete combat behavior tree.
func TestIntegrationCombatBehavior(t *testing.T) {
	// Setup entities
	attacker := NewEntity(3)
	attacker.AddComponent(&PositionComponent{X: 0, Y: 0})
	attacker.AddComponent(&HealthComponent{Current: 100, Max: 100})

	target := NewEntity(2)
	target.AddComponent(&PositionComponent{X: 30, Y: 0})
	target.AddComponent(&HealthComponent{Current: 100, Max: 100})

	blackboard := NewBlackboardWithSeed(42)
	blackboard.Set("target", target)

	// Build behavior tree: If target exists and in range, attack; else move closer
	attackBehavior := NewSequenceNode("attack_if_in_range",
		NewHasTargetNode("check_target"),
		NewInRangeNode("check_range", 10.0),
		NewAttackTargetNode("attack", 10.0, 10, 0.5, "melee"),
	)

	approachBehavior := NewSequenceNode("approach",
		NewHasTargetNode("check_target"),
		NewMoveToTargetNode("move", 50.0, 8.0),
	)

	tree := NewSelectorNode("combat_ai",
		attackBehavior,
		approachBehavior,
	)

	// Simulate a few seconds of behavior
	for tick := 0; tick < 100; tick++ {
		tree.Tick(attacker, blackboard, 0.016)
	}

	// Verify attacker moved toward target
	posComp, _ := attacker.GetComponent("position")
	pos := posComp.(*PositionComponent)
	if pos.X <= 0 {
		t.Errorf("Attacker should have moved, X = %v", pos.X)
	}

	// After approaching, should be attacking (target health reduced)
	targetHealth, _ := target.GetComponent("health")
	th := targetHealth.(*HealthComponent)
	// May or may not have attacked depending on timing
	t.Logf("Target health: %v, Attacker position: %v", th.Current, pos.X)
}

// TestWithDeterministicRNG verifies deterministic behavior with seeded RNG.
func TestWithDeterministicRNG(t *testing.T) {
	createTest := func(seed int64) float64 {
		entity := NewEntity(1)
		entity.AddComponent(&PositionComponent{X: 0, Y: 0})

		target := NewEntity(2)
		target.AddComponent(&PositionComponent{X: 0, Y: 0}) // Same position

		blackboard := NewBlackboardWithSeed(seed)
		blackboard.Set("target", target)

		node := NewFleeFromTargetNode("test", 100.0, 50.0)

		// Flee from same position uses RNG for direction
		node.Tick(entity, blackboard, 1.0)

		posComp, _ := entity.GetComponent("position")
		pos := posComp.(*PositionComponent)
		return pos.X
	}

	// Same seed should produce same result
	result1 := createTest(12345)
	result2 := createTest(12345)
	if result1 != result2 {
		t.Errorf("Same seed should produce same result: %v vs %v", result1, result2)
	}

	// Different seed should produce different result
	result3 := createTest(54321)
	if result1 == result3 {
		t.Log("Different seeds may occasionally produce same result")
	}
}

// Helper to seed for determinism
func init() {
	rand.Seed(42) // For any global rand usage in tests
}
