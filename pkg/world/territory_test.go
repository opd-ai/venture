package world

import (
	"testing"
	"time"
)

func TestNewTerritoryManager(t *testing.T) {
	tm := NewTerritoryManager()
	if tm == nil {
		t.Fatal("expected non-nil TerritoryManager")
	}
	if tm.captureRadius != 50.0 {
		t.Errorf("expected captureRadius 50.0, got %f", tm.captureRadius)
	}
	if tm.captureTimeBase != 60 {
		t.Errorf("expected captureTimeBase 60, got %d", tm.captureTimeBase)
	}
}

func TestCreateBorderZone(t *testing.T) {
	tm := NewTerritoryManager()
	zone := tm.CreateBorderZone("zone1", "serverA", "serverB", 3)

	if zone == nil {
		t.Fatal("expected non-nil BorderZone")
	}
	if zone.ZoneID != "zone1" {
		t.Errorf("expected ZoneID 'zone1', got '%s'", zone.ZoneID)
	}
	if zone.ServerA != "serverA" {
		t.Errorf("expected ServerA 'serverA', got '%s'", zone.ServerA)
	}
	if zone.ServerB != "serverB" {
		t.Errorf("expected ServerB 'serverB', got '%s'", zone.ServerB)
	}
	if len(zone.ControlPoints) != 3 {
		t.Errorf("expected 3 control points, got %d", len(zone.ControlPoints))
	}
	if zone.ResourceBonus != 0.10 {
		t.Errorf("expected ResourceBonus 0.10, got %f", zone.ResourceBonus)
	}
}

func TestGetZone(t *testing.T) {
	tm := NewTerritoryManager()
	tm.CreateBorderZone("zone1", "serverA", "serverB", 3)

	tests := []struct {
		name    string
		zoneID  string
		wantErr bool
	}{
		{"existing zone", "zone1", false},
		{"non-existent zone", "zone2", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zone, err := tm.GetZone(tt.zoneID)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if zone == nil {
					t.Error("expected non-nil zone")
				}
			}
		})
	}
}

func TestUpdateControlPoint(t *testing.T) {
	tm := NewTerritoryManager()
	zone := tm.CreateBorderZone("zone1", "serverA", "serverB", 3)

	cp := zone.ControlPoints[0]
	cp.LastUpdateTime = time.Now().Unix() - 10

	err := tm.UpdateControlPoint("zone1", 0, 1, 0, "factionA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cp.CaptureProgress <= 0.0 {
		t.Error("expected capture progress > 0")
	}
	if cp.CapturingFaction != "factionA" {
		t.Errorf("expected CapturingFaction 'factionA', got '%s'", cp.CapturingFaction)
	}
	if cp.AttackerCount != 1 {
		t.Errorf("expected AttackerCount 1, got %d", cp.AttackerCount)
	}
}

func TestUpdateControlPointWithDefenders(t *testing.T) {
	tm := NewTerritoryManager()
	zone := tm.CreateBorderZone("zone1", "serverA", "serverB", 3)

	cp := zone.ControlPoints[0]
	cp.CaptureProgress = 0.5
	cp.CapturingFaction = "factionA"
	cp.LastUpdateTime = time.Now().Unix() - 10

	err := tm.UpdateControlPoint("zone1", 0, 1, 2, "factionA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cp.CaptureProgress > 0.5 {
		t.Error("expected capture progress to decay with defenders present")
	}
	if cp.DefenderCount != 2 {
		t.Errorf("expected DefenderCount 2, got %d", cp.DefenderCount)
	}
}

func TestUpdateControlPointCompletion(t *testing.T) {
	tm := NewTerritoryManager()
	zone := tm.CreateBorderZone("zone1", "serverA", "serverB", 3)

	cp := zone.ControlPoints[0]
	cp.CaptureProgress = 0.9
	cp.LastUpdateTime = time.Now().Unix() - 60

	err := tm.UpdateControlPoint("zone1", 0, 1, 0, "factionA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cp.CaptureProgress != 1.0 {
		t.Errorf("expected full capture (1.0), got %f", cp.CaptureProgress)
	}
	if zone.OwnerFaction != "factionA" {
		t.Errorf("expected OwnerFaction 'factionA', got '%s'", zone.OwnerFaction)
	}
}

func TestUpdateControlPointInvalidIndex(t *testing.T) {
	tm := NewTerritoryManager()
	tm.CreateBorderZone("zone1", "serverA", "serverB", 3)

	err := tm.UpdateControlPoint("zone1", 10, 1, 0, "factionA")
	if err == nil {
		t.Error("expected error for invalid control point index")
	}
}

func TestIsPlayerInCaptureRange(t *testing.T) {
	tm := NewTerritoryManager()

	tests := []struct {
		name     string
		playerX  float64
		playerY  float64
		cpX      float64
		cpY      float64
		expected bool
	}{
		{"in range", 100.0, 100.0, 100.0, 100.0, true},
		{"at edge", 100.0, 100.0, 100.0, 150.0, true},
		{"out of range", 100.0, 100.0, 200.0, 200.0, false},
		{"close diagonal", 100.0, 100.0, 130.0, 130.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tm.IsPlayerInCaptureRange(tt.playerX, tt.playerY, tt.cpX, tt.cpY)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetControlledZones(t *testing.T) {
	tm := NewTerritoryManager()
	tm.CreateBorderZone("zone1", "serverA", "serverB", 3)
	tm.CreateBorderZone("zone2", "serverA", "serverC", 3)

	zone1, _ := tm.GetZone("zone1")
	zone1.OwnerFaction = "factionA"
	zone2, _ := tm.GetZone("zone2")
	zone2.OwnerFaction = "factionA"

	controlled := tm.GetControlledZones("factionA")
	if len(controlled) != 2 {
		t.Errorf("expected 2 controlled zones, got %d", len(controlled))
	}

	controlled = tm.GetControlledZones("factionB")
	if len(controlled) != 0 {
		t.Errorf("expected 0 controlled zones for factionB, got %d", len(controlled))
	}
}

func TestGetContestedZones(t *testing.T) {
	tm := NewTerritoryManager()
	zone := tm.CreateBorderZone("zone1", "serverA", "serverB", 3)
	tm.CreateBorderZone("zone2", "serverA", "serverC", 3)

	cp := zone.ControlPoints[0]
	cp.CaptureProgress = 0.5

	contested := tm.GetContestedZones()
	if len(contested) != 1 {
		t.Errorf("expected 1 contested zone, got %d", len(contested))
	}
}

func TestGetResourceBonus(t *testing.T) {
	tm := NewTerritoryManager()
	zone1 := tm.CreateBorderZone("zone1", "serverA", "serverB", 3)
	zone2 := tm.CreateBorderZone("zone2", "serverA", "serverC", 3)

	zone1.OwnerFaction = "factionA"
	zone2.OwnerFaction = "factionA"

	bonus := tm.GetResourceBonus("factionA")
	expected := 0.20
	if bonus != expected {
		t.Errorf("expected resource bonus %f, got %f", expected, bonus)
	}

	bonus = tm.GetResourceBonus("factionB")
	if bonus != 0.0 {
		t.Errorf("expected resource bonus 0.0 for factionB, got %f", bonus)
	}
}

func TestResetControlPoint(t *testing.T) {
	tm := NewTerritoryManager()
	zone := tm.CreateBorderZone("zone1", "serverA", "serverB", 3)

	cp := zone.ControlPoints[0]
	cp.CaptureProgress = 0.5
	cp.CapturingFaction = "factionA"
	cp.AttackerCount = 1
	cp.DefenderCount = 1

	err := tm.ResetControlPoint("zone1", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cp.CaptureProgress != 0.0 {
		t.Errorf("expected CaptureProgress 0.0, got %f", cp.CaptureProgress)
	}
	if cp.CapturingFaction != "" {
		t.Errorf("expected empty CapturingFaction, got '%s'", cp.CapturingFaction)
	}
	if cp.AttackerCount != 0 {
		t.Errorf("expected AttackerCount 0, got %d", cp.AttackerCount)
	}
	if cp.DefenderCount != 0 {
		t.Errorf("expected DefenderCount 0, got %d", cp.DefenderCount)
	}
}

func TestGetAllZones(t *testing.T) {
	tm := NewTerritoryManager()
	tm.CreateBorderZone("zone1", "serverA", "serverB", 3)
	tm.CreateBorderZone("zone2", "serverA", "serverC", 3)
	tm.CreateBorderZone("zone3", "serverB", "serverC", 3)

	zones := tm.GetAllZones()
	if len(zones) != 3 {
		t.Errorf("expected 3 zones, got %d", len(zones))
	}
}

func TestUpdateControlPoint_NegativeAttackers(t *testing.T) {
	tm := NewTerritoryManager()
	tm.CreateBorderZone("zone1", "serverA", "serverB", 3)

	err := tm.UpdateControlPoint("zone1", 0, -1, 0, "factionA")
	if err == nil {
		t.Error("expected error for negative attackers")
	}
}

func TestUpdateControlPoint_NegativeDefenders(t *testing.T) {
	tm := NewTerritoryManager()
	tm.CreateBorderZone("zone1", "serverA", "serverB", 3)

	err := tm.UpdateControlPoint("zone1", 0, 1, -1, "factionA")
	if err == nil {
		t.Error("expected error for negative defenders")
	}
}
