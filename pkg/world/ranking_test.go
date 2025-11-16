package world

import (
	"testing"
)

func TestNewRankingManager(t *testing.T) {
	rm := NewRankingManager()
	if rm == nil {
		t.Fatal("expected non-nil RankingManager")
	}
	if rm.GetServerCount() != 0 {
		t.Errorf("expected 0 servers, got %d", rm.GetServerCount())
	}
}

func TestRegisterServer(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")

	if rm.GetServerCount() != 1 {
		t.Errorf("expected 1 server, got %d", rm.GetServerCount())
	}

	rank, exists := rm.GetServerRank("server1")
	if !exists {
		t.Fatal("expected server1 to exist")
	}
	if rank.ServerID != "server1" {
		t.Errorf("expected ServerID 'server1', got '%s'", rank.ServerID)
	}
}

func TestUpdatePopulation(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.UpdatePopulation("server1", 100)

	rank, _ := rm.GetServerRank("server1")
	if rank.Population != 100 {
		t.Errorf("expected Population 100, got %d", rank.Population)
	}
	if rank.TotalScore == 0.0 {
		t.Error("expected TotalScore to be calculated")
	}
}

func TestUpdateEconomicPower(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.UpdateEconomicPower("server1", 50000)

	rank, _ := rm.GetServerRank("server1")
	if rank.EconomicPower != 50000 {
		t.Errorf("expected EconomicPower 50000, got %d", rank.EconomicPower)
	}
}

func TestUpdateMilitaryStrength(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.UpdateMilitaryStrength("server1", 5)

	rank, _ := rm.GetServerRank("server1")
	if rank.MilitaryStrength != 5 {
		t.Errorf("expected MilitaryStrength 5, got %d", rank.MilitaryStrength)
	}
}

func TestUpdateDiplomaticInfluence(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.UpdateDiplomaticInfluence("server1", 3)

	rank, _ := rm.GetServerRank("server1")
	if rank.DiplomaticInfluence != 3 {
		t.Errorf("expected DiplomaticInfluence 3, got %d", rank.DiplomaticInfluence)
	}
}

func TestGetLeaderboardPopulation(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.RegisterServer("server2")
	rm.RegisterServer("server3")

	rm.UpdatePopulation("server1", 100)
	rm.UpdatePopulation("server2", 200)
	rm.UpdatePopulation("server3", 150)

	leaderboard := rm.GetLeaderboard(LeaderboardPopulation, 0)
	if len(leaderboard) != 3 {
		t.Fatalf("expected 3 servers in leaderboard, got %d", len(leaderboard))
	}

	if leaderboard[0].ServerID != "server2" {
		t.Errorf("expected server2 at position 1, got '%s'", leaderboard[0].ServerID)
	}
	if leaderboard[1].ServerID != "server3" {
		t.Errorf("expected server3 at position 2, got '%s'", leaderboard[1].ServerID)
	}
	if leaderboard[2].ServerID != "server1" {
		t.Errorf("expected server1 at position 3, got '%s'", leaderboard[2].ServerID)
	}
}

func TestGetLeaderboardEconomic(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.RegisterServer("server2")

	rm.UpdateEconomicPower("server1", 100000)
	rm.UpdateEconomicPower("server2", 50000)

	leaderboard := rm.GetLeaderboard(LeaderboardEconomic, 0)
	if leaderboard[0].ServerID != "server1" {
		t.Errorf("expected server1 at position 1, got '%s'", leaderboard[0].ServerID)
	}
}

func TestGetLeaderboardMilitary(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.RegisterServer("server2")

	rm.UpdateMilitaryStrength("server1", 10)
	rm.UpdateMilitaryStrength("server2", 5)

	leaderboard := rm.GetLeaderboard(LeaderboardMilitary, 0)
	if leaderboard[0].ServerID != "server1" {
		t.Errorf("expected server1 at position 1, got '%s'", leaderboard[0].ServerID)
	}
}

func TestGetLeaderboardDiplomatic(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.RegisterServer("server2")

	rm.UpdateDiplomaticInfluence("server1", 2)
	rm.UpdateDiplomaticInfluence("server2", 5)

	leaderboard := rm.GetLeaderboard(LeaderboardDiplomatic, 0)
	if leaderboard[0].ServerID != "server2" {
		t.Errorf("expected server2 at position 1, got '%s'", leaderboard[0].ServerID)
	}
}

func TestGetLeaderboardOverall(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.RegisterServer("server2")

	rm.UpdatePopulation("server1", 100)
	rm.UpdateEconomicPower("server1", 50000)
	rm.UpdateMilitaryStrength("server1", 5)
	rm.UpdateDiplomaticInfluence("server1", 3)

	rm.UpdatePopulation("server2", 50)
	rm.UpdateEconomicPower("server2", 100000)

	leaderboard := rm.GetLeaderboard(LeaderboardOverall, 0)
	if len(leaderboard) != 2 {
		t.Fatalf("expected 2 servers, got %d", len(leaderboard))
	}
}

func TestGetLeaderboardLimit(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.RegisterServer("server2")
	rm.RegisterServer("server3")

	rm.UpdatePopulation("server1", 100)
	rm.UpdatePopulation("server2", 200)
	rm.UpdatePopulation("server3", 150)

	leaderboard := rm.GetLeaderboard(LeaderboardPopulation, 2)
	if len(leaderboard) != 2 {
		t.Errorf("expected 2 servers in limited leaderboard, got %d", len(leaderboard))
	}
}

func TestGetTopServer(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.RegisterServer("server2")

	rm.UpdatePopulation("server1", 100)
	rm.UpdatePopulation("server2", 200)

	top, exists := rm.GetTopServer(LeaderboardPopulation)
	if !exists {
		t.Fatal("expected top server to exist")
	}
	if top.ServerID != "server2" {
		t.Errorf("expected top server 'server2', got '%s'", top.ServerID)
	}
}

func TestGetTopServerEmpty(t *testing.T) {
	rm := NewRankingManager()
	_, exists := rm.GetTopServer(LeaderboardPopulation)
	if exists {
		t.Error("expected no top server in empty ranking")
	}
}

func TestGetServerPosition(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.RegisterServer("server2")
	rm.RegisterServer("server3")

	rm.UpdatePopulation("server1", 100)
	rm.UpdatePopulation("server2", 200)
	rm.UpdatePopulation("server3", 150)

	position := rm.GetServerPosition("server2", LeaderboardPopulation)
	if position != 1 {
		t.Errorf("expected position 1 for server2, got %d", position)
	}

	position = rm.GetServerPosition("server3", LeaderboardPopulation)
	if position != 2 {
		t.Errorf("expected position 2 for server3, got %d", position)
	}

	position = rm.GetServerPosition("server1", LeaderboardPopulation)
	if position != 3 {
		t.Errorf("expected position 3 for server1, got %d", position)
	}
}

func TestGetServerPositionNotFound(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")

	position := rm.GetServerPosition("nonexistent", LeaderboardPopulation)
	if position != -1 {
		t.Errorf("expected position -1 for non-existent server, got %d", position)
	}
}

func TestRemoveServer(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")
	rm.RegisterServer("server2")

	rm.RemoveServer("server1")

	if rm.GetServerCount() != 1 {
		t.Errorf("expected 1 server after removal, got %d", rm.GetServerCount())
	}

	_, exists := rm.GetServerRank("server1")
	if exists {
		t.Error("expected server1 to be removed")
	}
}

func TestLeaderboardTypeString(t *testing.T) {
	tests := []struct {
		leaderboard LeaderboardType
		expected    string
	}{
		{LeaderboardPopulation, "Population"},
		{LeaderboardEconomic, "Economic Power"},
		{LeaderboardMilitary, "Military Strength"},
		{LeaderboardDiplomatic, "Diplomatic Influence"},
		{LeaderboardOverall, "Overall Score"},
		{LeaderboardType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.leaderboard.String()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConcurrentUpdates(t *testing.T) {
	rm := NewRankingManager()
	rm.RegisterServer("server1")

	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			rm.UpdatePopulation("server1", i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			rm.GetServerRank("server1")
		}
		done <- true
	}()

	<-done
	<-done
}
