package network

import (
	"testing"
)

func TestNewStateUpdate(t *testing.T) {
	update := NewStateUpdate(123, 200)

	if update == nil {
		t.Fatal("Expected non-nil update")
	}
	if update.EntityID != 123 {
		t.Errorf("Expected EntityID 123, got %d", update.EntityID)
	}
	if update.Priority != 200 {
		t.Errorf("Expected Priority 200, got %d", update.Priority)
	}
}

func TestNewCriticalUpdate(t *testing.T) {
	update := NewCriticalUpdate(456)

	if update == nil {
		t.Fatal("Expected non-nil update")
	}
	if update.EntityID != 456 {
		t.Errorf("Expected EntityID 456, got %d", update.EntityID)
	}
	if update.Priority != PriorityCritical {
		t.Errorf("Expected Priority %d (PriorityCritical), got %d", PriorityCritical, update.Priority)
	}
}

func TestNewHighPriorityUpdate(t *testing.T) {
	update := NewHighPriorityUpdate(789)

	if update == nil {
		t.Fatal("Expected non-nil update")
	}
	if update.EntityID != 789 {
		t.Errorf("Expected EntityID 789, got %d", update.EntityID)
	}
	if update.Priority != PriorityHigh {
		t.Errorf("Expected Priority %d (PriorityHigh), got %d", PriorityHigh, update.Priority)
	}
}

func TestNewNormalUpdate(t *testing.T) {
	update := NewNormalUpdate(101112)

	if update == nil {
		t.Fatal("Expected non-nil update")
	}
	if update.EntityID != 101112 {
		t.Errorf("Expected EntityID 101112, got %d", update.EntityID)
	}
	if update.Priority != PriorityNormal {
		t.Errorf("Expected Priority %d (PriorityNormal), got %d", PriorityNormal, update.Priority)
	}
}

func TestNewLowPriorityUpdate(t *testing.T) {
	update := NewLowPriorityUpdate(131415)

	if update == nil {
		t.Fatal("Expected non-nil update")
	}
	if update.EntityID != 131415 {
		t.Errorf("Expected EntityID 131415, got %d", update.EntityID)
	}
	if update.Priority != PriorityLow {
		t.Errorf("Expected Priority %d (PriorityLow), got %d", PriorityLow, update.Priority)
	}
}

func TestHelperFunctions_PriorityValues(t *testing.T) {
	// Verify helper functions create updates with correct priority ordering
	critical := NewCriticalUpdate(1)
	high := NewHighPriorityUpdate(2)
	normal := NewNormalUpdate(3)
	low := NewLowPriorityUpdate(4)

	if critical.Priority <= high.Priority {
		t.Error("Expected critical priority > high priority")
	}
	if high.Priority <= normal.Priority {
		t.Error("Expected high priority > normal priority")
	}
	if normal.Priority <= low.Priority {
		t.Error("Expected normal priority > low priority")
	}
}
