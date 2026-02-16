package world

import (
	"testing"
	"time"
)

func TestEventType_String(t *testing.T) {
	tests := []struct {
		name      string
		eventType EventType
		want      string
	}{
		{"tournament", EventTournament, "Tournament"},
		{"server_vs_server", EventServerVsServer, "Server vs Server"},
		{"world_threat", EventWorldThreat, "World Threat"},
		{"seasonal_challenge", EventSeasonalChallenge, "Seasonal Challenge"},
		{"economic_crisis", EventEconomicCrisis, "Economic Crisis"},
		{"unknown", EventType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.eventType.String(); got != tt.want {
				t.Errorf("EventType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewEventManager(t *testing.T) {
	em := NewEventManager(12345)
	if em == nil {
		t.Fatal("expected non-nil EventManager")
	}
	if em.seed != 12345 {
		t.Errorf("expected seed 12345, got %d", em.seed)
	}
	if em.nextID != 1 {
		t.Errorf("expected nextID 1, got %d", em.nextID)
	}
}

func TestCreateTournament(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test Tournament", 3600, 4)

	if event == nil {
		t.Fatal("expected non-nil event")
	}
	if event.Type != EventTournament {
		t.Errorf("expected EventTournament, got %v", event.Type)
	}
	if event.Name != "Test Tournament" {
		t.Errorf("expected name 'Test Tournament', got '%s'", event.Name)
	}
	if !event.Active {
		t.Error("expected Active to be true")
	}
	if event.RequiredServers != 4 {
		t.Errorf("expected RequiredServers 4, got %d", event.RequiredServers)
	}
	if event.Goals["wins"] != 10 {
		t.Errorf("expected wins goal 10, got %d", event.Goals["wins"])
	}
}

func TestCreateServerVsServer(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateServerVsServer("Battle", "serverA", "serverB", 7200)

	if event == nil {
		t.Fatal("expected non-nil event")
	}
	if event.Type != EventServerVsServer {
		t.Errorf("expected EventServerVsServer, got %v", event.Type)
	}
	if len(event.Participants) != 2 {
		t.Errorf("expected 2 participants, got %d", len(event.Participants))
	}
	if event.Progress["serverA"] != 0 {
		t.Errorf("expected serverA progress 0, got %d", event.Progress["serverA"])
	}
	if event.Progress["serverB"] != 0 {
		t.Errorf("expected serverB progress 0, got %d", event.Progress["serverB"])
	}
}

func TestCreateWorldThreat(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateWorldThreat("Ancient Dragon", 10800, 5)

	if event == nil {
		t.Fatal("expected non-nil event")
	}
	if event.Type != EventWorldThreat {
		t.Errorf("expected EventWorldThreat, got %v", event.Type)
	}
	if event.RequiredServers != 5 {
		t.Errorf("expected RequiredServers 5, got %d", event.RequiredServers)
	}

	bossHealth := event.Goals["boss_damage"]
	if bossHealth <= 0 {
		t.Error("expected positive boss health")
	}
}

func TestCreateSeasonalChallenge(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateSeasonalChallenge("Winter", 86400)

	if event == nil {
		t.Fatal("expected non-nil event")
	}
	if event.Type != EventSeasonalChallenge {
		t.Errorf("expected EventSeasonalChallenge, got %v", event.Type)
	}
	if event.Name != "Winter Challenge" {
		t.Errorf("expected name 'Winter Challenge', got '%s'", event.Name)
	}
	if event.Goals["quests_completed"] != 50 {
		t.Errorf("expected quests_completed goal 50, got %d", event.Goals["quests_completed"])
	}
}

func TestCreateEconomicCrisis(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateEconomicCrisis("Market Crash", 14400)

	if event == nil {
		t.Fatal("expected non-nil event")
	}
	if event.Type != EventEconomicCrisis {
		t.Errorf("expected EventEconomicCrisis, got %v", event.Type)
	}
	if event.RequiredServers != 3 {
		t.Errorf("expected RequiredServers 3, got %d", event.RequiredServers)
	}
}

func TestRegisterParticipant(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test", 3600, 2)

	err := em.RegisterParticipant(event.ID, "server1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(event.Participants) != 1 {
		t.Errorf("expected 1 participant, got %d", len(event.Participants))
	}
	if event.Participants[0] != "server1" {
		t.Errorf("expected participant 'server1', got '%s'", event.Participants[0])
	}
}

func TestRegisterParticipant_Duplicate(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test", 3600, 2)

	em.RegisterParticipant(event.ID, "server1")
	err := em.RegisterParticipant(event.ID, "server1")

	if err == nil {
		t.Error("expected error for duplicate registration")
	}
}

func TestRegisterParticipant_NonExistent(t *testing.T) {
	em := NewEventManager(12345)
	err := em.RegisterParticipant("nonexistent", "server1")

	if err == nil {
		t.Error("expected error for non-existent event")
	}
}

func TestRegisterParticipant_Inactive(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test", 3600, 2)
	event.Active = false

	err := em.RegisterParticipant(event.ID, "server1")
	if err == nil {
		t.Error("expected error for inactive event")
	}
}

func TestUpdateProgress(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test", 3600, 2)
	em.RegisterParticipant(event.ID, "server1")

	err := em.UpdateProgress(event.ID, "server1", "wins", 5)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if event.Progress["server1:wins"] != 5 {
		t.Errorf("expected progress 5, got %d", event.Progress["server1:wins"])
	}
}

func TestUpdateProgress_MultipleUpdates(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test", 3600, 2)
	em.RegisterParticipant(event.ID, "server1")

	em.UpdateProgress(event.ID, "server1", "wins", 3)
	em.UpdateProgress(event.ID, "server1", "wins", 2)

	if event.Progress["server1:wins"] != 5 {
		t.Errorf("expected cumulative progress 5, got %d", event.Progress["server1:wins"])
	}
}

func TestUpdateProgress_Unregistered(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test", 3600, 2)

	err := em.UpdateProgress(event.ID, "server1", "wins", 5)
	if err == nil {
		t.Error("expected error for unregistered server")
	}
}

func TestCheckCompletion_WorldThreat(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateWorldThreat("Boss", 3600, 3)
	requiredDamage := event.Goals["boss_damage"]

	em.RegisterParticipant(event.ID, "server1")
	em.UpdateProgress(event.ID, "server1", "damage_dealt", requiredDamage)

	completed, err := em.CheckCompletion(event.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !completed {
		t.Error("expected event to be completed")
	}
	if event.Active {
		t.Error("expected event to be inactive after completion")
	}
}

func TestCheckCompletion_ServerVsServer(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateServerVsServer("Battle", "serverA", "serverB", 3600)
	requiredResources := event.Goals["resources"]

	em.UpdateProgress(event.ID, "serverA", "resources", requiredResources)

	completed, err := em.CheckCompletion(event.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !completed {
		t.Error("expected event to be completed")
	}
}

func TestCheckCompletion_Expired(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test", -1, 2)

	time.Sleep(10 * time.Millisecond)

	completed, err := em.CheckCompletion(event.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if completed {
		t.Error("expected expired event to not be completed")
	}
	if event.Active {
		t.Error("expected expired event to be inactive")
	}
}

func TestGetActiveEvents(t *testing.T) {
	em := NewEventManager(12345)
	em.CreateTournament("Event1", 3600, 2)
	em.CreateTournament("Event2", 3600, 2)
	inactive := em.CreateTournament("Event3", 3600, 2)
	inactive.Active = false

	active := em.GetActiveEvents()
	if len(active) != 2 {
		t.Errorf("expected 2 active events, got %d", len(active))
	}
}

func TestGetEventsByType(t *testing.T) {
	em := NewEventManager(12345)
	em.CreateTournament("Tournament1", 3600, 2)
	em.CreateTournament("Tournament2", 3600, 2)
	em.CreateWorldThreat("Threat1", 3600, 3)

	tournaments := em.GetEventsByType(EventTournament)
	if len(tournaments) != 2 {
		t.Errorf("expected 2 tournaments, got %d", len(tournaments))
	}

	threats := em.GetEventsByType(EventWorldThreat)
	if len(threats) != 1 {
		t.Errorf("expected 1 threat, got %d", len(threats))
	}
}

func TestGetEvent(t *testing.T) {
	em := NewEventManager(12345)
	created := em.CreateTournament("Test", 3600, 2)

	retrieved, err := em.GetEvent(created.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("expected ID %s, got %s", created.ID, retrieved.ID)
	}
}

func TestGetEvent_NonExistent(t *testing.T) {
	em := NewEventManager(12345)
	_, err := em.GetEvent("nonexistent")

	if err == nil {
		t.Error("expected error for non-existent event")
	}
}

func TestEventManager_Update(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test", -1, 2)

	time.Sleep(10 * time.Millisecond)
	em.Update(0.016)

	if event.Active {
		t.Error("expected expired event to be inactive after update")
	}
}

func TestGetEventCount(t *testing.T) {
	em := NewEventManager(12345)
	em.CreateTournament("Event1", 3600, 2)
	em.CreateTournament("Event2", 3600, 2)

	if em.GetEventCount() != 2 {
		t.Errorf("expected 2 events, got %d", em.GetEventCount())
	}
}

func TestGetActiveEventCount(t *testing.T) {
	em := NewEventManager(12345)
	em.CreateTournament("Event1", 3600, 2)
	inactive := em.CreateTournament("Event2", 3600, 2)
	inactive.Active = false

	if em.GetActiveEventCount() != 1 {
		t.Errorf("expected 1 active event, got %d", em.GetActiveEventCount())
	}
}

func TestCancelEvent(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test", 3600, 2)

	err := em.CancelEvent(event.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if event.Active {
		t.Error("expected event to be inactive after cancellation")
	}
}

func TestCancelEvent_NonExistent(t *testing.T) {
	em := NewEventManager(12345)
	err := em.CancelEvent("nonexistent")

	if err == nil {
		t.Error("expected error for non-existent event")
	}
}

func TestDeterministicWorldThreat(t *testing.T) {
	em1 := NewEventManager(12345)
	event1 := em1.CreateWorldThreat("Boss1", 3600, 3)

	em2 := NewEventManager(12345)
	event2 := em2.CreateWorldThreat("Boss2", 3600, 3)

	if event1.Goals["boss_damage"] != event2.Goals["boss_damage"] {
		t.Error("expected deterministic boss health generation with same seed")
	}
}

func TestMultipleServersProgress(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test", 3600, 4)

	servers := []string{"server1", "server2", "server3"}
	for _, server := range servers {
		em.RegisterParticipant(event.ID, server)
		em.UpdateProgress(event.ID, server, "wins", 5)
	}

	if len(event.Participants) != 3 {
		t.Errorf("expected 3 participants, got %d", len(event.Participants))
	}

	for _, server := range servers {
		if event.Progress[server+":wins"] != 5 {
			t.Errorf("expected progress 5 for %s, got %d", server, event.Progress[server+":wins"])
		}
	}
}

func TestCheckCompletion_Tournament(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateTournament("Test", 3600, 2)
	em.RegisterParticipant(event.ID, "server1")

	em.UpdateProgress(event.ID, "server1", "wins", 10)

	completed, err := em.CheckCompletion(event.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !completed {
		t.Error("expected tournament to be completed")
	}
}

func TestCheckCompletion_SeasonalChallenge(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateSeasonalChallenge("Test", 3600)
	em.RegisterParticipant(event.ID, "server1")

	// Update each goal independently using per-goal tracking
	em.UpdateProgress(event.ID, "server1", "quests_completed", 50)
	em.UpdateProgress(event.ID, "server1", "bosses_defeated", 5)

	completed, err := em.CheckCompletion(event.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if !completed {
		t.Error("expected seasonal challenge to be completed when all goals met")
	}
}

func TestCheckCompletion_SeasonalChallenge_Partial(t *testing.T) {
	em := NewEventManager(12345)
	event := em.CreateSeasonalChallenge("Test", 3600)
	em.RegisterParticipant(event.ID, "server1")

	// Only complete one of two goals
	em.UpdateProgress(event.ID, "server1", "quests_completed", 50)

	completed, err := em.CheckCompletion(event.ID)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if completed {
		t.Error("expected seasonal challenge to NOT be completed with only one goal met")
	}
}

func TestCheckCompletion_NonExistent(t *testing.T) {
	em := NewEventManager(12345)
	_, err := em.CheckCompletion("nonexistent")

	if err == nil {
		t.Error("expected error for non-existent event")
	}
}
