package legendary

import (
	"bytes"
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/world/raids"
)

func TestNewQuestManager(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	if mgr == nil {
		t.Fatal("expected manager to be created")
	}

	if mgr.generator == nil {
		t.Error("expected generator to be initialized")
	}

	if mgr.serverValidation == nil {
		t.Error("expected server validation to be initialized")
	}

	if mgr.rewardCatalog == nil {
		t.Error("expected reward catalog to be initialized")
	}
}

func TestQuestManager_GenerateQuest(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      50,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"player_level":    50,
			"servers_visited": 5,
		},
	}

	quest, err := mgr.GenerateQuest("player1", 12345, params)
	if err != nil {
		t.Fatalf("failed to generate quest: %v", err)
	}

	if quest == nil {
		t.Fatal("expected quest to be generated")
	}

	if quest.ID == "" {
		t.Error("expected quest to have ID")
	}

	if len(quest.Phases) < 5 || len(quest.Phases) > 10 {
		t.Errorf("expected 5-10 phases, got %d", len(quest.Phases))
	}

	// Verify quest is tracked
	activeQuests := mgr.GetActiveQuests()
	if len(activeQuests) != 1 {
		t.Errorf("expected 1 active quest, got %d", len(activeQuests))
	}
}

func TestQuestManager_UpdatePhaseProgress(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, err := mgr.GenerateQuest("player1", 12345, params)
	if err != nil {
		t.Fatalf("failed to generate quest: %v", err)
	}

	// Update phase 0 to 50% progress
	err = mgr.UpdatePhaseProgress("player1", quest.ID, 0, 0.5)
	if err != nil {
		t.Fatalf("failed to update phase progress: %v", err)
	}

	// Verify progress
	progress := mgr.GetPlayerProgress("player1", quest.ID)
	if progress == nil {
		t.Fatal("expected progress to be tracked")
	}

	if progress.PhaseProgress != 0.5 {
		t.Errorf("expected progress 0.5, got %f", progress.PhaseProgress)
	}
}

func TestQuestManager_UpdatePhaseProgress_InvalidQuest(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	err := mgr.UpdatePhaseProgress("player1", "invalid_quest", 0, 0.5)
	if err == nil {
		t.Error("expected error for invalid quest")
	}
}

func TestQuestManager_UpdatePhaseProgress_InvalidPhase(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, _ := mgr.GenerateQuest("player1", 12345, params)

	err := mgr.UpdatePhaseProgress("player1", quest.ID, 999, 0.5)
	if err == nil {
		t.Error("expected error for invalid phase index")
	}
}

func TestQuestManager_ValidateServerVisit(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, err := mgr.GenerateQuest("player1", 12345, params)
	if err != nil {
		t.Fatalf("failed to generate quest: %v", err)
	}

	// Find a required server from travel phases
	var serverID string
	for _, phase := range quest.Phases {
		if phase.Type == PhaseTravel && phase.Requirements != nil && len(phase.Requirements.ServersToVisit) > 0 {
			serverID = phase.Requirements.ServersToVisit[0]
			break
		}
	}

	if serverID == "" {
		t.Skip("no travel phase in generated quest")
	}

	// Validate server visit
	err = mgr.ValidateServerVisit("player1", quest.ID, serverID)
	if err != nil {
		t.Fatalf("failed to validate server visit: %v", err)
	}

	// Verify progress
	progress := mgr.GetPlayerProgress("player1", quest.ID)
	if !contains(progress.ServersVisited, serverID) {
		t.Error("expected server to be recorded in progress")
	}
}

func TestQuestManager_ValidateServerVisit_InvalidServer(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, _ := mgr.GenerateQuest("player1", 12345, params)

	err := mgr.ValidateServerVisit("player1", quest.ID, "invalid_server")
	if err == nil {
		t.Error("expected error for invalid server")
	}
}

func TestQuestManager_ValidateRaidCompletion(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, err := mgr.GenerateQuest("player1", 12345, params)
	if err != nil {
		t.Fatalf("failed to generate quest: %v", err)
	}

	// Find raid phase
	var raidTier raids.RaidTier
	found := false
	for _, phase := range quest.Phases {
		if phase.Type == PhaseRaid && phase.Requirements != nil && len(phase.Requirements.RaidEncounters) > 0 {
			raidTier = phase.Requirements.RaidEncounters[0].Tier
			found = true
			break
		}
	}

	if !found {
		t.Skip("no raid phase in generated quest")
	}

	// Validate raid completion
	err = mgr.ValidateRaidCompletion("player1", quest.ID, "raid123", raidTier)
	if err != nil {
		t.Fatalf("failed to validate raid completion: %v", err)
	}

	// Verify progress
	progress := mgr.GetPlayerProgress("player1", quest.ID)
	if !contains(progress.RaidsCompleted, "raid123") {
		t.Error("expected raid to be recorded in progress")
	}
}

func TestQuestManager_ValidateCraftingCompletion(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, err := mgr.GenerateQuest("player1", 12345, params)
	if err != nil {
		t.Fatalf("failed to generate quest: %v", err)
	}

	// Find crafting phase
	found := false
	for _, phase := range quest.Phases {
		if phase.Type == PhaseCraft {
			found = true
			break
		}
	}

	if !found {
		t.Skip("no crafting phase in generated quest")
	}

	// Validate crafting with sufficient quality (Master tier = 4)
	err = mgr.ValidateCraftingCompletion("player1", quest.ID, "item123", 4)
	if err != nil {
		t.Fatalf("failed to validate crafting: %v", err)
	}

	// Verify progress
	progress := mgr.GetPlayerProgress("player1", quest.ID)
	if progress.MaterialsGathered["item123"] != 1 {
		t.Error("expected material to be recorded in progress")
	}
}

func TestQuestManager_ValidateCraftingCompletion_InsufficientQuality(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.9,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, err := mgr.GenerateQuest("player1", 12345, params)
	if err != nil {
		t.Fatalf("failed to generate quest: %v", err)
	}

	// Find crafting phase
	found := false
	for _, phase := range quest.Phases {
		if phase.Type == PhaseCraft {
			found = true
			break
		}
	}

	if !found {
		t.Skip("no crafting phase in generated quest")
	}

	// Validate crafting with insufficient quality (Basic tier = 1)
	err = mgr.ValidateCraftingCompletion("player1", quest.ID, "item123", 1)
	if err == nil {
		t.Error("expected error for insufficient crafting quality")
	}
}

func TestQuestManager_CompleteQuest(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, err := mgr.GenerateQuest("player1", 12345, params)
	if err != nil {
		t.Fatalf("failed to generate quest: %v", err)
	}

	// Complete all phases
	for i := range quest.Phases {
		err = mgr.UpdatePhaseProgress("player1", quest.ID, i, 1.0)
		if err != nil {
			t.Fatalf("failed to update phase %d: %v", i, err)
		}
	}

	// Complete quest
	rewards, err := mgr.CompleteQuest("player1", quest.ID)
	if err != nil {
		t.Fatalf("failed to complete quest: %v", err)
	}

	if rewards == nil {
		t.Fatal("expected rewards to be granted")
	}

	if rewards.Gold != quest.Rewards.Gold {
		t.Errorf("expected gold %d, got %d", quest.Rewards.Gold, rewards.Gold)
	}
}

func TestQuestManager_CompleteQuest_Incomplete(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, _ := mgr.GenerateQuest("player1", 12345, params)

	// Try to complete without finishing phases
	_, err := mgr.CompleteQuest("player1", quest.ID)
	if err == nil {
		t.Error("expected error when completing incomplete quest")
	}
}

func TestQuestManager_SaveLoad(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, _ := mgr.GenerateQuest("player1", 12345, params)
	mgr.UpdatePhaseProgress("player1", quest.ID, 0, 0.75)

	// Save state
	var buf bytes.Buffer
	err := mgr.Save(&buf)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	// Create new manager and load
	mgr2 := NewQuestManager(raidMgr)
	err = mgr2.Load(&buf)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}

	// Verify state
	progress := mgr2.GetPlayerProgress("player1", quest.ID)
	if progress == nil {
		t.Fatal("expected progress to be restored")
	}

	if progress.PhaseProgress != 0.75 {
		t.Errorf("expected progress 0.75, got %f", progress.PhaseProgress)
	}
}

func TestServerValidator_RecordVisit(t *testing.T) {
	sv := NewServerValidator()

	err := sv.RecordVisit("player1", "server1")
	if err != nil {
		t.Fatalf("failed to record visit: %v", err)
	}

	servers := sv.GetVisitedServers("player1")
	if len(servers) != 1 {
		t.Errorf("expected 1 server, got %d", len(servers))
	}

	if servers[0] != "server1" {
		t.Errorf("expected server1, got %s", servers[0])
	}
}

func TestServerValidator_MultipleVisits(t *testing.T) {
	sv := NewServerValidator()

	sv.RecordVisit("player1", "server1")
	sv.RecordVisit("player1", "server2")
	sv.RecordVisit("player1", "server3")

	servers := sv.GetVisitedServers("player1")
	if len(servers) != 3 {
		t.Errorf("expected 3 servers, got %d", len(servers))
	}
}

func TestServerValidator_RegisterFederatedServer(t *testing.T) {
	sv := NewServerValidator()

	sv.RegisterFederatedServer("server1")
	sv.RegisterFederatedServer("server2")

	if len(sv.federatedServers) != 2 {
		t.Errorf("expected 2 federated servers, got %d", len(sv.federatedServers))
	}
}

func TestRewardCatalog_ClaimReward(t *testing.T) {
	rc := NewRewardCatalog()

	err := rc.ClaimReward("player1", "reward1")
	if err != nil {
		t.Fatalf("failed to claim reward: %v", err)
	}

	if !rc.IsRewardClaimed("player1", "reward1") {
		t.Error("expected reward to be claimed")
	}
}

func TestRewardCatalog_ClaimReward_AlreadyClaimed(t *testing.T) {
	rc := NewRewardCatalog()

	rc.ClaimReward("player1", "reward1")
	err := rc.ClaimReward("player1", "reward1")
	if err == nil {
		t.Error("expected error when claiming reward twice")
	}
}

func TestRewardCatalog_GetAvailableRewards(t *testing.T) {
	rc := NewRewardCatalog()

	// Claim some rewards
	rc.ClaimReward("player1", "legendary_item_1")
	rc.ClaimReward("player1", "legendary_item_2")

	available := rc.GetAvailableRewards("player1")

	// Should have 48 available (50 total - 2 claimed)
	if len(available) != 48 {
		t.Errorf("expected 48 available rewards, got %d", len(available))
	}
}

func TestRewardCatalog_GeneratedRewards(t *testing.T) {
	rc := NewRewardCatalog()

	// Verify 50 legendary items were generated
	if len(rc.rewardPool) != 50 {
		t.Errorf("expected 50 rewards, got %d", len(rc.rewardPool))
	}

	// Verify all rewards are legendary tier
	for _, reward := range rc.rewardPool {
		if reward.Rarity != 3.0 {
			t.Errorf("expected rarity 3.0, got %f", reward.Rarity)
		}
		if !reward.Unique {
			t.Error("expected reward to be unique")
		}
	}
}

func TestQuestManager_GetStatistics(t *testing.T) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      50,
		GenreID:    "fantasy",
	}

	mgr.GenerateQuest("player1", 12345, params)
	mgr.GenerateQuest("player2", 12346, params)

	stats := mgr.GetStatistics()
	if stats == nil {
		t.Fatal("expected statistics to be returned")
	}

	if stats.TotalQuests != 2 {
		t.Errorf("expected 2 total quests, got %d", stats.TotalQuests)
	}

	if stats.ActiveQuests != 2 {
		t.Errorf("expected 2 active quests, got %d", stats.ActiveQuests)
	}
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Benchmarks

func BenchmarkQuestManager_GenerateQuest(b *testing.B) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      50,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.GenerateQuest("player1", int64(i), params)
	}
}

func BenchmarkQuestManager_UpdatePhaseProgress(b *testing.B) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, _ := mgr.GenerateQuest("player1", 12345, params)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.UpdatePhaseProgress("player1", quest.ID, 0, float64(i%100)/100.0)
	}
}

func BenchmarkQuestManager_ValidateServerVisit(b *testing.B) {
	raidMgr := raids.NewManager(12345, "fantasy")
	mgr := NewQuestManager(raidMgr)

	params := procgen.GenerationParams{
		Difficulty: 0.8,
		Depth:      50,
		GenreID:    "fantasy",
	}

	quest, _ := mgr.GenerateQuest("player1", 12345, params)

	// Find a required server
	var serverID string
	for _, phase := range quest.Phases {
		if phase.Type == PhaseTravel && phase.Requirements != nil && len(phase.Requirements.ServersToVisit) > 0 {
			serverID = phase.Requirements.ServersToVisit[0]
			break
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.ValidateServerVisit("player1", quest.ID, serverID)
	}
}

func BenchmarkRewardCatalog_GetAvailableRewards(b *testing.B) {
	rc := NewRewardCatalog()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rc.GetAvailableRewards("player1")
	}
}
