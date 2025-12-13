package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/world/raids"
)

func TestRaidLockoutComponent(t *testing.T) {
	tests := []struct {
		name       string
		tier       raids.RaidTier
		shouldLock bool
	}{
		{"normal tier", raids.TierNormal, true},
		{"heroic tier", raids.TierHeroic, true},
		{"mythic tier", raids.TierMythic, true},
		{"legendary tier", raids.TierLegendary, true},
		{"nightmare tier", raids.TierNightmare, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lockout := NewRaidLockoutComponent()

			// Player should not be locked out initially
			if lockout.IsLockedOut(tt.tier) {
				t.Errorf("player should not be locked out initially")
			}

			// Set lockout
			lockout.SetLockout("player1", tt.tier)

			// Player should now be locked out
			if !lockout.IsLockedOut(tt.tier) {
				t.Errorf("player should be locked out after SetLockout")
			}

			// Verify lockout exists in map
			lockoutData, exists := lockout.Lockouts[tt.tier]
			if !exists {
				t.Errorf("lockout data not found in map")
			}

			// Verify lockout properties
			if lockoutData.PlayerID != "player1" {
				t.Errorf("playerID = %s, want %s", lockoutData.PlayerID, "player1")
			}
			if lockoutData.Tier != tt.tier {
				t.Errorf("tier = %v, want %v", lockoutData.Tier, tt.tier)
			}

			// Verify reset time is ~7 days from now
			expectedReset := time.Now().Add(7 * 24 * time.Hour)
			timeDiff := lockoutData.NextReset.Sub(expectedReset)
			if timeDiff > time.Minute || timeDiff < -time.Minute {
				t.Errorf("reset time off by %v, should be ~7 days", timeDiff)
			}
		})
	}
}

func TestRaidLockoutComponentType(t *testing.T) {
	lockout := NewRaidLockoutComponent()
	if lockout.Type() != "raid_lockout" {
		t.Errorf("Type() = %s, want raid_lockout", lockout.Type())
	}
}

func TestRaidInstanceComponentType(t *testing.T) {
	instance := &RaidInstanceComponent{}
	if instance.Type() != "raid_instance" {
		t.Errorf("Type() = %s, want raid_instance", instance.Type())
	}
}

func TestRaidBossComponentType(t *testing.T) {
	boss := &RaidBossComponent{}
	if boss.Type() != "raid_boss" {
		t.Errorf("Type() = %s, want raid_boss", boss.Type())
	}
}

func TestRaidSystemCreation(t *testing.T) {
	world := NewWorld()
	system := NewRaidSystem(world, 12345)

	if system.world != world {
		t.Errorf("world reference not set correctly")
	}
	if system.generator == nil {
		t.Errorf("generator not initialized")
	}
	if system.instanceManager == nil {
		t.Errorf("instance manager not initialized")
	}
	if system.lockoutManager == nil {
		t.Errorf("lockout manager not initialized")
	}
	if system.cleanupInterval != 60.0 {
		t.Errorf("cleanupInterval = %f, want 60.0", system.cleanupInterval)
	}
}

func TestRaidSystemCreateInstance(t *testing.T) {
	world := NewWorld()
	system := NewRaidSystem(world, 12345)

	playerIDs := []string{"player1", "player2", "player3", "player4", "player5"}
	instanceID, entrance, err := system.CreateRaidInstance(
		raids.TierNormal,
		"group1",
		playerIDs,
		"fantasy",
		54321,
	)
	if err != nil {
		t.Fatalf("CreateRaidInstance failed: %v", err)
	}

	if instanceID == "" {
		t.Errorf("instanceID is empty")
	}

	if entrance == nil {
		t.Fatalf("entrance entity is nil")
	}

	// Verify entrance has required components
	if _, ok := entrance.GetComponent("position"); !ok {
		t.Errorf("entrance missing position component")
	}
	if _, ok := entrance.GetComponent("raid_instance"); !ok {
		t.Errorf("entrance missing raid_instance component")
	}
	if _, ok := entrance.GetComponent("sprite"); !ok {
		t.Errorf("entrance missing sprite component")
	}
	if _, ok := entrance.GetComponent("collision"); !ok {
		t.Errorf("entrance missing collision component")
	}

	// Verify instance component properties
	instanceCompRaw, ok := entrance.GetComponent("raid_instance")
	if !ok {
		t.Fatal("entrance missing raid_instance component")
	}
	instanceComp := instanceCompRaw.(*RaidInstanceComponent)
	if instanceComp.InstanceID != instanceID {
		t.Errorf("instanceID = %s, want %s", instanceComp.InstanceID, instanceID)
	}
	if instanceComp.Tier != raids.TierNormal {
		t.Errorf("tier = %v, want %v", instanceComp.Tier, raids.TierNormal)
	}
	if instanceComp.GroupID != "group1" {
		t.Errorf("groupID = %s, want group1", instanceComp.GroupID)
	}
	if len(instanceComp.PlayerIDs) != len(playerIDs) {
		t.Errorf("playerIDs length = %d, want %d", len(instanceComp.PlayerIDs), len(playerIDs))
	}
	if instanceComp.Completed {
		t.Errorf("instance should not be completed initially")
	}
	if instanceComp.RaidDungeon == nil {
		t.Errorf("raid dungeon is nil")
	}
}

func TestRaidSystemEnterInstance(t *testing.T) {
	world := NewWorld()
	system := NewRaidSystem(world, 12345)

	// Create instance
	playerIDs := []string{"player1"}
	_, entrance, err := system.CreateRaidInstance(
		raids.TierNormal,
		"group1",
		playerIDs,
		"fantasy",
		54321,
	)
	if err != nil {
		t.Fatalf("CreateRaidInstance failed: %v", err)
	}

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&PlayerComponent{})

	// Enter instance
	err = system.EnterRaidInstance(player, entrance)
	if err != nil {
		t.Fatalf("EnterRaidInstance failed: %v", err)
	}

	// Verify player has lockout component
	if _, ok := player.GetComponent("raid_lockout"); !ok {
		t.Errorf("player should have lockout component after entering")
	}

	// Verify player position changed
	posRaw, ok := player.GetComponent("position")
	if !ok {
		t.Fatal("player missing position component")
	}
	pos := posRaw.(*PositionComponent)
	if pos.X == 0 && pos.Y == 0 {
		t.Errorf("player position should have changed")
	}
}

func TestRaidSystemEnterInstanceLockout(t *testing.T) {
	world := NewWorld()
	system := NewRaidSystem(world, 12345)

	// Create instance
	playerIDs := []string{"player1"}
	_, entrance, err := system.CreateRaidInstance(
		raids.TierNormal,
		"group1",
		playerIDs,
		"fantasy",
		54321,
	)
	if err != nil {
		t.Fatalf("CreateRaidInstance failed: %v", err)
	}

	// Create player with existing lockout
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 0, Y: 0})
	player.AddComponent(&PlayerComponent{})
	lockout := NewRaidLockoutComponent()
	lockout.SetLockout("player1", raids.TierNormal)
	player.AddComponent(lockout)

	// Try to enter instance (should fail)
	err = system.EnterRaidInstance(player, entrance)
	if err == nil {
		t.Errorf("EnterRaidInstance should fail for locked out player")
	}
}

func TestRaidSystemCompleteInstance(t *testing.T) {
	world := NewWorld()
	system := NewRaidSystem(world, 12345)

	// Create instance
	playerIDs := []string{"player1", "player2"}
	instanceID, entrance, err := system.CreateRaidInstance(
		raids.TierHeroic,
		"group1",
		playerIDs,
		"fantasy",
		54321,
	)
	if err != nil {
		t.Fatalf("CreateRaidInstance failed: %v", err)
	}

	// Create players
	player1 := world.CreateEntity()
	player1.AddComponent(&PlayerComponent{})
	player2 := world.CreateEntity()
	player2.AddComponent(&PlayerComponent{})

	players := []*Entity{player1, player2}

	// Complete instance
	err = system.CompleteRaidInstance(instanceID, players)
	if err != nil {
		t.Fatalf("CompleteRaidInstance failed: %v", err)
	}

	// Verify instance marked complete
	instanceCompRaw, ok := entrance.GetComponent("raid_instance")
	if !ok {
		t.Fatal("entrance missing raid_instance component")
	}
	instanceComp := instanceCompRaw.(*RaidInstanceComponent)
	if !instanceComp.Completed {
		t.Errorf("instance should be marked completed")
	}

	// Verify both players have lockouts
	for i, player := range players {
		lockoutComp, ok := player.GetComponent("raid_lockout")
		if !ok {
			t.Errorf("player%d should have lockout component", i+1)
			continue
		}
		lockout := lockoutComp.(*RaidLockoutComponent)
		if !lockout.IsLockedOut(raids.TierHeroic) {
			t.Errorf("player%d should be locked out of heroic tier", i+1)
		}
	}
}

func TestRaidSystemUpdateBossMechanics(t *testing.T) {
	world := NewWorld()
	system := NewRaidSystem(world, 12345)

	// Create boss entity
	boss := world.CreateEntity()
	boss.AddComponent(&PositionComponent{X: 500, Y: 500})
	boss.AddComponent(&HealthComponent{Current: 1000, Max: 1000})

	mechanic := raids.BossMechanic{
		ID:       "test_mechanic",
		Name:     "Test Mechanic",
		Type:     raids.MechanicInstant,
		Cooldown: 5 * time.Second,
		Damage:   100,
		Radius:   200,
	}

	bossComp := &RaidBossComponent{
		BossIndex:     0,
		CurrentPhase:  0,
		Mechanics:     []raids.BossMechanic{mechanic},
		Phases:        []raids.BossPhase{},
		MechanicTimer: make(map[string]float64),
		InstanceID:    "test_instance",
	}
	boss.AddComponent(bossComp)

	// Create target player
	player := world.CreateEntity()
	player.AddComponent(&PositionComponent{X: 550, Y: 550}) // Within AoE range
	player.AddComponent(&HealthComponent{Current: 500, Max: 500})
	player.AddComponent(&PlayerComponent{})

	// Update system multiple times to trigger mechanic
	for i := 0; i < 10; i++ {
		system.Update(1.0) // 1 second per update
	}

	// Verify mechanic timer was initialized and updated
	timer, exists := bossComp.MechanicTimer["mechanic_0"]
	if !exists {
		t.Errorf("mechanic timer not initialized")
	}
	if timer < 0 || timer > 5.0 {
		t.Errorf("mechanic timer = %f, should be 0-5.0", timer)
	}
}

func TestRaidSystemUpdateBossPhases(t *testing.T) {
	world := NewWorld()
	system := NewRaidSystem(world, 12345)

	// Create boss entity
	boss := world.CreateEntity()
	boss.AddComponent(&PositionComponent{X: 500, Y: 500})
	boss.AddComponent(&HealthComponent{Current: 800, Max: 1000})

	phase1 := raids.BossPhase{
		Number:       1,
		HealthThresh: 0.75,
		Mechanics:    []string{"mechanic1"},
		AddSpawns:    2,
	}

	phase2 := raids.BossPhase{
		Number:       2,
		HealthThresh: 0.50,
		Mechanics:    []string{"mechanic2"},
		AddSpawns:    1,
	}

	bossComp := &RaidBossComponent{
		BossIndex:     0,
		CurrentPhase:  0,
		Mechanics:     []raids.BossMechanic{},
		Phases:        []raids.BossPhase{phase1, phase2},
		MechanicTimer: make(map[string]float64),
		InstanceID:    "test_instance",
	}
	boss.AddComponent(bossComp)

	// Update system
	system.Update(1.0)

	// Verify phase transition happened (800/1000 = 0.8, so phase 1 at 0.75 should trigger)
	if bossComp.CurrentPhase != 1 {
		t.Errorf("current phase = %d, want 1", bossComp.CurrentPhase)
	}
	if !bossComp.PhaseEntered {
		t.Errorf("PhaseEntered should be true")
	}

	// Reduce health to trigger phase 2
	healthCompRaw, ok := boss.GetComponent("health")
	if !ok {
		t.Fatal("boss missing health component")
	}
	healthComp := healthCompRaw.(*HealthComponent)
	healthComp.Current = 450 // 450/1000 = 0.45, below 0.50 threshold

	system.Update(1.0)

	// Verify phase 2 transition
	if bossComp.CurrentPhase != 2 {
		t.Errorf("current phase = %d, want 2", bossComp.CurrentPhase)
	}
}

func TestRaidSystemCleanupExpiredInstances(t *testing.T) {
	world := NewWorld()
	system := NewRaidSystem(world, 12345)

	// Create instance with past expiration
	entrance := world.CreateEntity()
	entrance.AddComponent(&RaidInstanceComponent{
		InstanceID: "expired_instance",
		ExpiresAt:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		Completed:  false,
	})

	// Verify entrance exists
	if _, ok := world.GetEntity(entrance.ID); !ok {
		t.Fatal("entrance entity not found before cleanup")
	}

	// Trigger cleanup
	system.cleanupExpiredInstances()

	// Verify entrance was removed
	if _, ok := world.GetEntity(entrance.ID); ok {
		t.Errorf("expired entrance should have been removed")
	}
}

func TestRaidSystemCleanupCompletedInstances(t *testing.T) {
	world := NewWorld()
	system := NewRaidSystem(world, 12345)

	// Create completed instance
	entrance := world.CreateEntity()
	entrance.AddComponent(&RaidInstanceComponent{
		InstanceID: "completed_instance",
		ExpiresAt:  time.Now().Add(4 * time.Hour), // Not expired
		Completed:  true,                          // But completed
	})

	// Verify entrance exists
	if _, ok := world.GetEntity(entrance.ID); !ok {
		t.Fatal("entrance entity not found before cleanup")
	}

	// Trigger cleanup
	system.cleanupExpiredInstances()

	// Verify entrance was removed
	if _, ok := world.GetEntity(entrance.ID); ok {
		t.Errorf("completed entrance should have been removed")
	}
}
