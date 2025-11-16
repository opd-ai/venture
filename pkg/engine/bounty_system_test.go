package engine

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewBountySystem(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	if bs == nil {
		t.Fatal("expected non-nil BountySystem")
	}
	if bs.GetActiveBountyCount() != 0 {
		t.Errorf("expected 0 active bounties, got %d", bs.GetActiveBountyCount())
	}
}

func TestCreateBounty(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bounty := bs.CreateBounty("serverA", "serverB", ObjectiveKill, "Kill 10 monsters", 100, 1, 3600)

	if bounty == nil {
		t.Fatal("expected non-nil bounty")
	}
	if bounty.IssuerServer != "serverA" {
		t.Errorf("expected IssuerServer 'serverA', got '%s'", bounty.IssuerServer)
	}
	if bounty.TargetServer != "serverB" {
		t.Errorf("expected TargetServer 'serverB', got '%s'", bounty.TargetServer)
	}
	if bounty.Objective != ObjectiveKill {
		t.Errorf("expected Objective Kill, got %v", bounty.Objective)
	}
	if bounty.Reward != 100 {
		t.Errorf("expected Reward 100, got %d", bounty.Reward)
	}
	if bounty.Difficulty != 1 {
		t.Errorf("expected Difficulty 1, got %d", bounty.Difficulty)
	}
}

func TestAcceptBounty(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bounty := bs.CreateBounty("serverA", "serverB", ObjectiveDeliver, "Deliver package", 50, 1, 3600)

	err := bs.AcceptBounty(bounty.ID, "player1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bounty.AcceptedBy != "player1" {
		t.Errorf("expected AcceptedBy 'player1', got '%s'", bounty.AcceptedBy)
	}
}

func TestAcceptBountyAlreadyAccepted(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bounty := bs.CreateBounty("serverA", "serverB", ObjectiveDeliver, "Deliver package", 50, 1, 3600)
	bs.AcceptBounty(bounty.ID, "player1")

	err := bs.AcceptBounty(bounty.ID, "player2")
	if err == nil {
		t.Error("expected error when accepting already accepted bounty")
	}
}

func TestAcceptBountyNotFound(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	err := bs.AcceptBounty("nonexistent", "player1")
	if err == nil {
		t.Error("expected error for non-existent bounty")
	}
}

func TestCompleteBounty(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bounty := bs.CreateBounty("serverA", "serverB", ObjectiveKill, "Kill 10 monsters", 100, 1, 3600)
	bs.AcceptBounty(bounty.ID, "player1")

	err := bs.CompleteBounty(bounty.ID, "player1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bounty.CompletedAt == 0 {
		t.Error("expected CompletedAt to be set")
	}
}

func TestCompleteBountyNotAccepted(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bounty := bs.CreateBounty("serverA", "serverB", ObjectiveKill, "Kill 10 monsters", 100, 1, 3600)

	err := bs.CompleteBounty(bounty.ID, "player1")
	if err == nil {
		t.Error("expected error when completing non-accepted bounty")
	}
}

func TestCompleteBountyWrongPlayer(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bounty := bs.CreateBounty("serverA", "serverB", ObjectiveKill, "Kill 10 monsters", 100, 1, 3600)
	bs.AcceptBounty(bounty.ID, "player1")

	err := bs.CompleteBounty(bounty.ID, "player2")
	if err == nil {
		t.Error("expected error when wrong player completes bounty")
	}
}

func TestGetBounty(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bounty := bs.CreateBounty("serverA", "serverB", ObjectiveExplore, "Explore region", 75, 2, 3600)

	retrieved, err := bs.GetBounty(bounty.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.ID != bounty.ID {
		t.Errorf("expected bounty ID '%s', got '%s'", bounty.ID, retrieved.ID)
	}
}

func TestGetAvailableBounties(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bs.CreateBounty("serverA", "serverB", ObjectiveKill, "Kill monsters", 100, 1, 3600)
	bounty2 := bs.CreateBounty("serverA", "serverC", ObjectiveDeliver, "Deliver item", 50, 1, 3600)
	bs.AcceptBounty(bounty2.ID, "player1")

	available := bs.GetAvailableBounties()
	if len(available) != 1 {
		t.Errorf("expected 1 available bounty, got %d", len(available))
	}
}

func TestGetBountiesByServer(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bs.CreateBounty("serverA", "serverB", ObjectiveKill, "Kill monsters", 100, 1, 3600)
	bs.CreateBounty("serverA", "serverC", ObjectiveDeliver, "Deliver item", 50, 1, 3600)
	bs.CreateBounty("serverB", "serverC", ObjectiveEscort, "Escort NPC", 75, 2, 3600)

	bounties := bs.GetBountiesByServer("serverA")
	if len(bounties) != 2 {
		t.Errorf("expected 2 bounties for serverA, got %d", len(bounties))
	}

	bounties = bs.GetBountiesByServer("serverB")
	if len(bounties) != 2 {
		t.Errorf("expected 2 bounties for serverB, got %d", len(bounties))
	}
}

func TestGetBountiesByDifficulty(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bs.CreateBounty("serverA", "serverB", ObjectiveKill, "Kill monsters", 100, 1, 3600)
	bs.CreateBounty("serverA", "serverC", ObjectiveDeliver, "Deliver item", 50, 1, 3600)
	bs.CreateBounty("serverB", "serverC", ObjectiveEscort, "Escort NPC", 75, 2, 3600)

	bounties := bs.GetBountiesByDifficulty(1)
	if len(bounties) != 2 {
		t.Errorf("expected 2 bounties with difficulty 1, got %d", len(bounties))
	}

	bounties = bs.GetBountiesByDifficulty(2)
	if len(bounties) != 1 {
		t.Errorf("expected 1 bounty with difficulty 2, got %d", len(bounties))
	}
}

func TestGetCompletionRate(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bounty1 := bs.CreateBounty("serverA", "serverB", ObjectiveKill, "Kill monsters", 100, 1, 3600)
	bounty2 := bs.CreateBounty("serverA", "serverC", ObjectiveDeliver, "Deliver item", 50, 1, 3600)

	bs.AcceptBounty(bounty1.ID, "player1")
	bs.AcceptBounty(bounty2.ID, "player2")
	bs.CompleteBounty(bounty1.ID, "player1")

	bs.Update(0.016)

	rate := bs.GetCompletionRate()
	if rate != 0.5 {
		t.Errorf("expected completion rate 0.5, got %f", rate)
	}
}

func TestBountyExpiration(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bs.CreateBounty("serverA", "serverB", ObjectiveKill, "Kill monsters", 100, 1, -1)

	if bs.GetActiveBountyCount() != 1 {
		t.Errorf("expected 1 active bounty, got %d", bs.GetActiveBountyCount())
	}

	bs.Update(0.016)

	if bs.GetActiveBountyCount() != 0 {
		t.Errorf("expected 0 active bounties after expiration, got %d", bs.GetActiveBountyCount())
	}
}

func TestCancelBounty(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	bs := NewBountySystem(world, logger)

	bounty := bs.CreateBounty("serverA", "serverB", ObjectiveKill, "Kill monsters", 100, 1, 3600)

	err := bs.CancelBounty(bounty.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bs.GetActiveBountyCount() != 0 {
		t.Errorf("expected 0 active bounties after cancellation, got %d", bs.GetActiveBountyCount())
	}
}

func TestObjectiveTypeString(t *testing.T) {
	tests := []struct {
		objective ObjectiveType
		expected  string
	}{
		{ObjectiveKill, "Kill"},
		{ObjectiveDeliver, "Deliver"},
		{ObjectiveEscort, "Escort"},
		{ObjectiveExplore, "Explore"},
		{ObjectiveCraft, "Craft"},
		{ObjectiveType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.objective.String()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBountyComponentType(t *testing.T) {
	bc := BountyComponent{}
	if bc.Type() != "bounty" {
		t.Errorf("expected component type 'bounty', got '%s'", bc.Type())
	}
}
