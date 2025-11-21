package network

import (
	"testing"
	"time"
)

// TestDesyncScenarios validates all 12 desync scenarios from Phase 64.3.

func TestDesyncScenario_Combat(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Server and client disagree on kill attribution
	// Player A (ID 100) vs Player B (ID 200)
	victimID := uint64(999)

	serverComponents := []ComponentData{
		{Type: "combat", Data: []byte{100}}, // killer ID 100
	}
	clientComponents := []ComponentData{
		{Type: "combat", Data: []byte{200}}, // killer ID 200
	}

	serverChecksum := detector.ComputeChecksum(victimID, 1000, serverComponents)
	clientChecksum := detector.ComputeChecksum(victimID, 1000, clientComponents)

	detected := detector.DetectDesync(DesyncCombat, victimID, clientChecksum.Hash, serverChecksum.Hash,
		"kill attribution: server=100, client=200")

	if !detected {
		t.Error("combat desync should be detected")
	}

	events := detector.GetEventsByType(DesyncCombat)
	if len(events) != 1 {
		t.Fatalf("expected 1 combat event, got %d", len(events))
	}
	if events[0].EntityID != victimID {
		t.Errorf("expected victim ID %d, got %d", victimID, events[0].EntityID)
	}
}

func TestDesyncScenario_Inventory(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Item counts mismatch (server has 10, client has 5)
	playerID := uint64(123)

	serverComponents := []ComponentData{
		{Type: "inventory", Data: []byte{10}}, // 10 items
	}
	clientComponents := []ComponentData{
		{Type: "inventory", Data: []byte{5}}, // 5 items
	}

	serverChecksum := detector.ComputeChecksum(playerID, 2000, serverComponents)
	clientChecksum := detector.ComputeChecksum(playerID, 2000, clientComponents)

	detected := detector.DetectDesync(DesyncInventory, playerID, clientChecksum.Hash, serverChecksum.Hash,
		"item count: server=10, client=5")

	if !detected {
		t.Error("inventory desync should be detected")
	}

	events := detector.GetEventsByType(DesyncInventory)
	if len(events) != 1 {
		t.Errorf("expected 1 inventory event, got %d", len(events))
	}
}

func TestDesyncScenario_Position(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Player location differs by >1 tile (32px)
	playerID := uint64(456)

	// Server position: (100, 200)
	serverComponents := []ComponentData{
		{Type: "position", Data: []byte{100, 0, 200, 0}}, // X=100, Y=200
	}
	// Client position: (150, 200) - 50px difference = ~1.5 tiles
	clientComponents := []ComponentData{
		{Type: "position", Data: []byte{150, 0, 200, 0}}, // X=150, Y=200
	}

	serverChecksum := detector.ComputeChecksum(playerID, 3000, serverComponents)
	clientChecksum := detector.ComputeChecksum(playerID, 3000, clientComponents)

	detected := detector.DetectDesync(DesyncPosition, playerID, clientChecksum.Hash, serverChecksum.Hash,
		"position: server=(100,200), client=(150,200), delta=50px")

	if !detected {
		t.Error("position desync should be detected")
	}

	events := detector.GetEventsByType(DesyncPosition)
	if len(events) != 1 {
		t.Errorf("expected 1 position event, got %d", len(events))
	}
}

func TestDesyncScenario_Entity(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Entity exists on client but not server
	entityID := uint64(777)

	// Client has entity data
	clientComponents := []ComponentData{
		{Type: "position", Data: []byte{10, 20}},
		{Type: "health", Data: []byte{100}},
	}
	// Server has no data (nil/empty)
	serverComponents := []ComponentData{}

	clientChecksum := detector.ComputeChecksum(entityID, 4000, clientComponents)
	serverChecksum := detector.ComputeChecksum(entityID, 4000, serverComponents)

	detected := detector.DetectDesync(DesyncEntity, entityID, clientChecksum.Hash, serverChecksum.Hash,
		"entity exists on client but not server")

	if !detected {
		t.Error("entity desync should be detected")
	}

	events := detector.GetEventsByType(DesyncEntity)
	if len(events) != 1 {
		t.Errorf("expected 1 entity event, got %d", len(events))
	}
}

func TestDesyncScenario_Quest(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Quest state diverges (server=completed, client=incomplete)
	questID := uint64(888)

	serverComponents := []ComponentData{
		{Type: "quest", Data: []byte{2}}, // 2=completed
	}
	clientComponents := []ComponentData{
		{Type: "quest", Data: []byte{1}}, // 1=in_progress
	}

	serverChecksum := detector.ComputeChecksum(questID, 5000, serverComponents)
	clientChecksum := detector.ComputeChecksum(questID, 5000, clientComponents)

	detected := detector.DetectDesync(DesyncQuest, questID, clientChecksum.Hash, serverChecksum.Hash,
		"quest state: server=completed, client=in_progress")

	if !detected {
		t.Error("quest desync should be detected")
	}

	events := detector.GetEventsByType(DesyncQuest)
	if len(events) != 1 {
		t.Errorf("expected 1 quest event, got %d", len(events))
	}
}

func TestDesyncScenario_Guild(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Guild member list differs across servers
	guildID := uint64(111)

	// Server has 3 members
	serverComponents := []ComponentData{
		{Type: "guild", Data: []byte{1, 2, 3}}, // member IDs: 1,2,3
	}
	// Client has 2 members
	clientComponents := []ComponentData{
		{Type: "guild", Data: []byte{1, 2}}, // member IDs: 1,2
	}

	serverChecksum := detector.ComputeChecksum(guildID, 6000, serverComponents)
	clientChecksum := detector.ComputeChecksum(guildID, 6000, clientComponents)

	detected := detector.DetectDesync(DesyncGuild, guildID, clientChecksum.Hash, serverChecksum.Hash,
		"guild members: server=3, client=2")

	if !detected {
		t.Error("guild desync should be detected")
	}

	events := detector.GetEventsByType(DesyncGuild)
	if len(events) != 1 {
		t.Errorf("expected 1 guild event, got %d", len(events))
	}
}

func TestDesyncScenario_Housing(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Furniture placement mismatch
	housingID := uint64(222)

	// Server has furniture at (10, 20)
	serverComponents := []ComponentData{
		{Type: "furniture", Data: []byte{10, 20}},
	}
	// Client has furniture at (15, 20)
	clientComponents := []ComponentData{
		{Type: "furniture", Data: []byte{15, 20}},
	}

	serverChecksum := detector.ComputeChecksum(housingID, 7000, serverComponents)
	clientChecksum := detector.ComputeChecksum(housingID, 7000, clientComponents)

	detected := detector.DetectDesync(DesyncHousing, housingID, clientChecksum.Hash, serverChecksum.Hash,
		"furniture position: server=(10,20), client=(15,20)")

	if !detected {
		t.Error("housing desync should be detected")
	}

	events := detector.GetEventsByType(DesyncHousing)
	if len(events) != 1 {
		t.Errorf("expected 1 housing event, got %d", len(events))
	}
}

func TestDesyncScenario_Trade(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Ownership transfer fails mid-transaction
	itemID := uint64(333)

	// Server shows item owned by player 100
	serverComponents := []ComponentData{
		{Type: "ownership", Data: []byte{100}},
	}
	// Client shows item owned by player 200
	clientComponents := []ComponentData{
		{Type: "ownership", Data: []byte{200}},
	}

	serverChecksum := detector.ComputeChecksum(itemID, 8000, serverComponents)
	clientChecksum := detector.ComputeChecksum(itemID, 8000, clientComponents)

	detected := detector.DetectDesync(DesyncTrade, itemID, clientChecksum.Hash, serverChecksum.Hash,
		"item ownership: server=100, client=200")

	if !detected {
		t.Error("trade desync should be detected")
	}

	events := detector.GetEventsByType(DesyncTrade)
	if len(events) != 1 {
		t.Errorf("expected 1 trade event, got %d", len(events))
	}
}

func TestDesyncScenario_Vehicle(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Mount/dismount state differs
	playerID := uint64(444)

	// Server shows player mounted
	serverComponents := []ComponentData{
		{Type: "vehicle", Data: []byte{1}}, // 1=mounted
	}
	// Client shows player dismounted
	clientComponents := []ComponentData{
		{Type: "vehicle", Data: []byte{0}}, // 0=dismounted
	}

	serverChecksum := detector.ComputeChecksum(playerID, 9000, serverComponents)
	clientChecksum := detector.ComputeChecksum(playerID, 9000, clientComponents)

	detected := detector.DetectDesync(DesyncVehicle, playerID, clientChecksum.Hash, serverChecksum.Hash,
		"vehicle state: server=mounted, client=dismounted")

	if !detected {
		t.Error("vehicle desync should be detected")
	}

	events := detector.GetEventsByType(DesyncVehicle)
	if len(events) != 1 {
		t.Errorf("expected 1 vehicle event, got %d", len(events))
	}
}

func TestDesyncScenario_Companion(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Loyalty/XP values drift
	companionID := uint64(555)

	// Server: loyalty=100, XP=5000
	serverComponents := []ComponentData{
		{Type: "companion", Data: []byte{100, 0, 136, 19}}, // loyalty=100, XP=5000
	}
	// Client: loyalty=95, XP=4800
	clientComponents := []ComponentData{
		{Type: "companion", Data: []byte{95, 0, 192, 18}}, // loyalty=95, XP=4800
	}

	serverChecksum := detector.ComputeChecksum(companionID, 10000, serverComponents)
	clientChecksum := detector.ComputeChecksum(companionID, 10000, clientComponents)

	detected := detector.DetectDesync(DesyncCompanion, companionID, clientChecksum.Hash, serverChecksum.Hash,
		"companion: server=(loyalty=100, xp=5000), client=(loyalty=95, xp=4800)")

	if !detected {
		t.Error("companion desync should be detected")
	}

	events := detector.GetEventsByType(DesyncCompanion)
	if len(events) != 1 {
		t.Errorf("expected 1 companion event, got %d", len(events))
	}
}

func TestDesyncScenario_Chunk(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Terrain modifications lost
	chunkID := uint64(666)

	// Server has modified terrain
	serverComponents := []ComponentData{
		{Type: "terrain", Data: []byte{1, 1, 1}}, // modified tiles
	}
	// Client has original terrain
	clientComponents := []ComponentData{
		{Type: "terrain", Data: []byte{0, 0, 0}}, // unmodified tiles
	}

	serverChecksum := detector.ComputeChecksum(chunkID, 11000, serverComponents)
	clientChecksum := detector.ComputeChecksum(chunkID, 11000, clientComponents)

	detected := detector.DetectDesync(DesyncChunk, chunkID, clientChecksum.Hash, serverChecksum.Hash,
		"terrain modifications: server=modified, client=original")

	if !detected {
		t.Error("chunk desync should be detected")
	}

	events := detector.GetEventsByType(DesyncChunk)
	if len(events) != 1 {
		t.Errorf("expected 1 chunk event, got %d", len(events))
	}
}

func TestDesyncScenario_Skill(t *testing.T) {
	detector := NewDesyncDetector()

	// Scenario: Skill points/unlocks mismatch
	playerID := uint64(777)

	// Server: 5 skill points, 3 unlocks
	serverComponents := []ComponentData{
		{Type: "skills", Data: []byte{5, 3}}, // points=5, unlocks=3
	}
	// Client: 3 skill points, 2 unlocks
	clientComponents := []ComponentData{
		{Type: "skills", Data: []byte{3, 2}}, // points=3, unlocks=2
	}

	serverChecksum := detector.ComputeChecksum(playerID, 12000, serverComponents)
	clientChecksum := detector.ComputeChecksum(playerID, 12000, clientComponents)

	detected := detector.DetectDesync(DesyncSkill, playerID, clientChecksum.Hash, serverChecksum.Hash,
		"skills: server=(points=5, unlocks=3), client=(points=3, unlocks=2)")

	if !detected {
		t.Error("skill desync should be detected")
	}

	events := detector.GetEventsByType(DesyncSkill)
	if len(events) != 1 {
		t.Errorf("expected 1 skill event, got %d", len(events))
	}
}

// TestAllDesyncTypes verifies all 12 desync types can be detected.
func TestAllDesyncTypes(t *testing.T) {
	detector := NewDesyncDetector()

	desyncTypes := []DesyncType{
		DesyncCombat, DesyncInventory, DesyncPosition, DesyncEntity,
		DesyncQuest, DesyncGuild, DesyncHousing, DesyncTrade,
		DesyncVehicle, DesyncCompanion, DesyncChunk, DesyncSkill,
	}

	hash1 := [32]byte{1}
	hash2 := [32]byte{2}

	for i, desyncType := range desyncTypes {
		detector.DetectDesync(desyncType, uint64(i), hash1, hash2, string(desyncType))
	}

	events := detector.GetEvents()
	if len(events) != 12 {
		t.Fatalf("expected 12 events (one per type), got %d", len(events))
	}

	// Verify each type was recorded
	for _, desyncType := range desyncTypes {
		typeEvents := detector.GetEventsByType(desyncType)
		if len(typeEvents) != 1 {
			t.Errorf("expected 1 event for type %v, got %d", desyncType, len(typeEvents))
		}
	}
}

// TestDesyncRecovery_AcceptanceCriteria validates Phase 64.3 acceptance criteria.
func TestDesyncRecovery_AcceptanceCriteria(t *testing.T) {
	detector := NewDesyncDetector()
	detector.SetDetectionDeadline(30 * time.Second)
	detector.SetRecoveryDeadline(10 * time.Second)

	// Test detection time <30 seconds
	hash1 := [32]byte{1}
	hash2 := [32]byte{2}

	start := time.Now()
	detected := detector.DetectDesync(DesyncCombat, 123, hash1, hash2, "test")
	detectionTime := time.Since(start)

	if !detected {
		t.Fatal("desync should be detected")
	}
	if detectionTime > 30*time.Second {
		t.Errorf("detection time %v exceeds 30 second target", detectionTime)
	}
	if detectionTime > 1*time.Millisecond {
		t.Errorf("detection time %v should be nearly instant", detectionTime)
	}

	// Test recovery time <10 seconds
	recoveryStart := time.Now()
	detector.RecordRecovery(DesyncCombat, 123, 5*time.Second)
	recoveryDuration := time.Since(recoveryStart)

	if recoveryDuration > 10*time.Second {
		t.Errorf("recovery marking time %v exceeds 10 second target", recoveryDuration)
	}

	events := detector.GetEvents()
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if !events[0].Recovered {
		t.Error("event should be marked as recovered")
	}
	if events[0].RecoveryTime != 5*time.Second {
		t.Errorf("expected recovery time 5s, got %v", events[0].RecoveryTime)
	}
}

// TestFalsePositiveRate validates <1% false positive target.
func TestFalsePositiveRate(t *testing.T) {
	detector := NewDesyncDetector()

	// Generate 1000 identical checksums - should not trigger desyncs
	components := []ComponentData{
		{Type: "position", Data: []byte{10, 20}},
	}

	falsePositives := 0
	for i := 0; i < 1000; i++ {
		checksum1 := detector.ComputeChecksum(1, 1000, components)
		checksum2 := detector.ComputeChecksum(1, 1000, components)

		if checksum1.Hash != checksum2.Hash {
			falsePositives++
		}
	}

	falsePositiveRate := float64(falsePositives) / 1000.0
	if falsePositiveRate > 0.01 {
		t.Errorf("false positive rate %.4f exceeds 1%% target", falsePositiveRate)
	}
	if falsePositives > 0 {
		t.Errorf("expected zero false positives, got %d", falsePositives)
	}
}
