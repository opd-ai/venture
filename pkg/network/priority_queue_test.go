package network

import (
	"testing"
)

func TestPriorityConstants(t *testing.T) {
	// Verify priority ordering
	if PriorityCritical <= PriorityHigh {
		t.Errorf("Expected PriorityCritical (%d) > PriorityHigh (%d)", PriorityCritical, PriorityHigh)
	}
	if PriorityHigh <= PriorityNormal {
		t.Errorf("Expected PriorityHigh (%d) > PriorityNormal (%d)", PriorityHigh, PriorityNormal)
	}
	if PriorityNormal <= PriorityLow {
		t.Errorf("Expected PriorityNormal (%d) > PriorityLow (%d)", PriorityNormal, PriorityLow)
	}
}

func TestNewStateUpdatePriorityQueue(t *testing.T) {
	pq := NewStateUpdatePriorityQueue(10)
	
	if pq == nil {
		t.Fatal("Expected non-nil priority queue")
	}
	
	if pq.Cap() != 10 {
		t.Errorf("Expected capacity 10, got %d", pq.Cap())
	}
	
	if pq.Len() != 0 {
		t.Errorf("Expected empty queue, got length %d", pq.Len())
	}
	
	if !pq.IsEmpty() {
		t.Error("Expected IsEmpty() to be true")
	}
	
	if pq.IsFull() {
		t.Error("Expected IsFull() to be false")
	}
}

func TestPriorityQueue_PushPop(t *testing.T) {
	pq := NewStateUpdatePriorityQueue(10)
	
	// Push a normal priority update
	update1 := &StateUpdate{
		EntityID: 1,
		Priority: PriorityNormal,
	}
	
	if !pq.Push(update1) {
		t.Fatal("Failed to push update")
	}
	
	if pq.Len() != 1 {
		t.Errorf("Expected length 1, got %d", pq.Len())
	}
	
	// Pop and verify
	popped := pq.Pop()
	if popped == nil {
		t.Fatal("Expected non-nil popped update")
	}
	
	if popped.EntityID != 1 {
		t.Errorf("Expected EntityID 1, got %d", popped.EntityID)
	}
	
	if pq.Len() != 0 {
		t.Errorf("Expected empty queue after pop, got length %d", pq.Len())
	}
}

func TestPriorityQueue_Ordering(t *testing.T) {
	tests := []struct {
		name      string
		priorities []uint8
		expected   []uint8
	}{
		{
			name:       "ascending priorities",
			priorities: []uint8{PriorityLow, PriorityNormal, PriorityHigh, PriorityCritical},
			expected:   []uint8{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow},
		},
		{
			name:       "descending priorities",
			priorities: []uint8{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow},
			expected:   []uint8{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow},
		},
		{
			name:       "mixed priorities",
			priorities: []uint8{PriorityNormal, PriorityCritical, PriorityLow, PriorityHigh},
			expected:   []uint8{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow},
		},
		{
			name:       "same priorities",
			priorities: []uint8{PriorityNormal, PriorityNormal, PriorityNormal},
			expected:   []uint8{PriorityNormal, PriorityNormal, PriorityNormal},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pq := NewStateUpdatePriorityQueue(10)
			
			// Push all updates
			for i, priority := range tt.priorities {
				update := &StateUpdate{
					EntityID: uint64(i),
					Priority: priority,
				}
				if !pq.Push(update) {
					t.Fatalf("Failed to push update %d", i)
				}
			}
			
			// Pop all and verify ordering
			for i, expectedPriority := range tt.expected {
				popped := pq.Pop()
				if popped == nil {
					t.Fatalf("Expected update at position %d, got nil", i)
				}
				if popped.Priority != expectedPriority {
					t.Errorf("At position %d: expected priority %d, got %d", i, expectedPriority, popped.Priority)
				}
			}
			
			// Verify empty
			if !pq.IsEmpty() {
				t.Errorf("Expected empty queue, got length %d", pq.Len())
			}
		})
	}
}

func TestPriorityQueue_CapacityLimit(t *testing.T) {
	capacity := 5
	pq := NewStateUpdatePriorityQueue(capacity)
	
	// Fill the queue
	for i := 0; i < capacity; i++ {
		update := &StateUpdate{
			EntityID: uint64(i),
			Priority: PriorityNormal,
		}
		if !pq.Push(update) {
			t.Fatalf("Failed to push update %d (within capacity)", i)
		}
	}
	
	if !pq.IsFull() {
		t.Error("Expected IsFull() to be true")
	}
	
	// Try to push one more (should fail)
	overflow := &StateUpdate{
		EntityID: 999,
		Priority: PriorityCritical,
	}
	if pq.Push(overflow) {
		t.Error("Expected push to fail when queue is full")
	}
	
	if pq.Len() != capacity {
		t.Errorf("Expected length %d, got %d", capacity, pq.Len())
	}
}

func TestPriorityQueue_PopEmpty(t *testing.T) {
	pq := NewStateUpdatePriorityQueue(10)
	
	popped := pq.Pop()
	if popped != nil {
		t.Error("Expected nil when popping from empty queue")
	}
}

func TestPriorityQueue_CriticalBeforeNormal(t *testing.T) {
	pq := NewStateUpdatePriorityQueue(10)
	
	// Push normal priority first
	normal := &StateUpdate{
		EntityID: 1,
		Priority: PriorityNormal,
	}
	pq.Push(normal)
	
	// Push critical priority second
	critical := &StateUpdate{
		EntityID: 2,
		Priority: PriorityCritical,
	}
	pq.Push(critical)
	
	// Critical should come out first
	first := pq.Pop()
	if first == nil || first.EntityID != 2 {
		t.Errorf("Expected critical update (EntityID=2) first, got %v", first)
	}
	
	// Normal should come out second
	second := pq.Pop()
	if second == nil || second.EntityID != 1 {
		t.Errorf("Expected normal update (EntityID=1) second, got %v", second)
	}
}

func TestPriorityQueue_CustomPriorities(t *testing.T) {
	pq := NewStateUpdatePriorityQueue(10)
	
	// Test with custom priority values
	updates := []*StateUpdate{
		{EntityID: 1, Priority: 50},
		{EntityID: 2, Priority: 100},
		{EntityID: 3, Priority: 150},
		{EntityID: 4, Priority: 75},
	}
	
	for _, update := range updates {
		pq.Push(update)
	}
	
	// Should come out in descending priority order
	expected := []uint64{3, 2, 4, 1} // EntityIDs in priority order
	for i, expectedID := range expected {
		popped := pq.Pop()
		if popped == nil {
			t.Fatalf("Expected update at position %d, got nil", i)
		}
		if popped.EntityID != expectedID {
			t.Errorf("At position %d: expected EntityID %d, got %d", i, expectedID, popped.EntityID)
		}
	}
}

// BenchmarkPriorityQueue_Push benchmarks push operations
func BenchmarkPriorityQueue_Push(b *testing.B) {
	pq := NewStateUpdatePriorityQueue(10000)
	update := &StateUpdate{
		EntityID: 1,
		Priority: PriorityNormal,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Push(update)
		if pq.IsFull() {
			// Reset when full
			pq = NewStateUpdatePriorityQueue(10000)
		}
	}
}

// BenchmarkPriorityQueue_Pop benchmarks pop operations
func BenchmarkPriorityQueue_Pop(b *testing.B) {
	pq := NewStateUpdatePriorityQueue(b.N)
	
	// Pre-fill the queue
	for i := 0; i < b.N; i++ {
		update := &StateUpdate{
			EntityID: uint64(i),
			Priority: PriorityNormal,
		}
		pq.Push(update)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Pop()
	}
}

// BenchmarkPriorityQueue_PushPop benchmarks mixed operations
func BenchmarkPriorityQueue_PushPop(b *testing.B) {
	pq := NewStateUpdatePriorityQueue(1000)
	update := &StateUpdate{
		EntityID: 1,
		Priority: PriorityNormal,
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pq.Push(update)
		pq.Pop()
	}
}
