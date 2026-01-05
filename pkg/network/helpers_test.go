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

// TestSequenceLessThan tests the wrap-around-safe sequence comparison
func TestSequenceLessThan(t *testing.T) {
	tests := []struct {
		name   string
		seq1   uint32
		seq2   uint32
		expect bool
	}{
		{"equal sequences", 100, 100, false},
		{"normal comparison - seq1 less", 100, 200, true},
		{"normal comparison - seq1 greater", 200, 100, false},
		{"wrap-around - seq1 before wrap", 0xFFFFFFF0, 0x00000010, true},
		{"wrap-around - seq1 after wrap", 0x00000010, 0xFFFFFFF0, false},
		{"exactly at UINT32_MAX", 0xFFFFFFFF, 0x00000000, true},
		{"exactly at UINT32_MAX reversed", 0x00000000, 0xFFFFFFFF, false},
		{"large gap no wrap", 1000, 1000000000, true},
		{"large gap with wrap", 0xC0000000, 0x40000000, false}, // 3/4 range vs 1/4 range
		// At exactly half range (2^31), the difference equals the threshold, not less than
		// So we treat this as not less than to avoid ambiguity
		{"edge case - exactly half range", 0, 0x80000000, false},
		{"edge case - just over half range", 0, 0x80000001, false},
		{"edge case - just under half range", 0, 0x7FFFFFFF, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sequenceLessThan(tt.seq1, tt.seq2)
			if result != tt.expect {
				t.Errorf("sequenceLessThan(%d, %d) = %v, want %v", tt.seq1, tt.seq2, result, tt.expect)
			}
		})
	}
}

// TestSequenceDifference tests the wrap-around-safe sequence difference calculation
func TestSequenceDifference(t *testing.T) {
	tests := []struct {
		name   string
		newer  uint32
		older  uint32
		expect uint32
	}{
		{"no difference", 100, 100, 0},
		{"normal difference", 200, 100, 100},
		{"small difference", 10, 5, 5},
		{"large difference", 1000000, 1000, 999000},
		{"wrap-around difference", 10, 0xFFFFFFF0, 26},  // 10 - (max-15) = 26
		{"wrap at UINT32_MAX", 0, 0xFFFFFFFF, 1},        // 0 - max = 1
		{"wrap with larger gap", 100, 0xFFFFFF00, 356},  // 100 - (max-255) = 356
		{"exactly one wrap", 0x00000000, 0xFFFFFFFF, 1}, // 0 - max = 1
		{"multiple values after wrap", 0x00001000, 0xFFFFF000, 0x00002000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sequenceDifference(tt.newer, tt.older)
			if result != tt.expect {
				t.Errorf("sequenceDifference(%d, %d) = %d, want %d", tt.newer, tt.older, result, tt.expect)
			}
		})
	}
}

// TestSequenceInRange tests the wrap-around-safe range check
func TestSequenceInRange(t *testing.T) {
	tests := []struct {
		name     string
		seq      uint32
		ref      uint32
		rangeVal uint32
		expect   bool
	}{
		{"exact match", 100, 100, 10, true},
		{"within range ahead", 105, 100, 10, true},
		{"within range behind", 95, 100, 10, true},
		{"exactly at range limit ahead", 110, 100, 10, true},
		{"exactly at range limit behind", 90, 100, 10, true},
		{"just outside range ahead", 111, 100, 10, false},
		{"just outside range behind", 89, 100, 10, false},
		{"wrap-around within range", 5, 0xFFFFFFFB, 10, true},
		{"wrap-around outside range", 20, 0xFFFFFFFB, 10, false},
		{"large range", 5000, 1000, 10000, true},
		{"large range outside", 15000, 1000, 10000, false},
		{"zero range exact", 100, 100, 0, true},
		{"zero range different", 101, 100, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sequenceInRange(tt.seq, tt.ref, tt.rangeVal)
			if result != tt.expect {
				t.Errorf("sequenceInRange(%d, %d, %d) = %v, want %v", tt.seq, tt.ref, tt.rangeVal, result, tt.expect)
			}
		})
	}
}

// TestSequenceWrapAroundScenario tests a realistic wrap-around scenario
func TestSequenceWrapAroundScenario(t *testing.T) {
	// Simulate a server approaching and crossing UINT32_MAX
	// These sequences are chosen to actually have a difference of 1 between each
	sequences := []uint32{
		0xFFFFFFFC, // -4 from max
		0xFFFFFFFD, // -3 from max
		0xFFFFFFFE, // -2 from max
		0xFFFFFFFF, // max
		0x00000000, // wrapped to 0
		0x00000001, // +1 after wrap
		0x00000002, // +2 after wrap
	}

	// Verify each sequence is "less than" the next
	for i := 0; i < len(sequences)-1; i++ {
		if !sequenceLessThan(sequences[i], sequences[i+1]) {
			t.Errorf("Expected sequence %d (0x%X) < sequence %d (0x%X)",
				sequences[i], sequences[i], sequences[i+1], sequences[i+1])
		}

		// Verify difference is 1
		diff := sequenceDifference(sequences[i+1], sequences[i])
		if diff != 1 {
			t.Errorf("Expected difference of 1 between sequence %d and %d, got %d",
				sequences[i], sequences[i+1], diff)
		}
	}

	// Verify the first sequence is within range of the last (range of 10)
	if !sequenceInRange(sequences[len(sequences)-1], sequences[0], 10) {
		t.Errorf("Expected sequence %d to be within 10 of sequence %d",
			sequences[len(sequences)-1], sequences[0])
	}
}
