package network

import (
	"sync"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewBufferStats(t *testing.T) {
	tests := []struct {
		name     string
		buffName string
		capacity int
		hasLog   bool
	}{
		{
			name:     "basic stats without logger",
			buffName: "test_buffer",
			capacity: 256,
			hasLog:   false,
		},
		{
			name:     "stats with logger",
			buffName: "monitored_buffer",
			capacity: 512,
			hasLog:   true,
		},
		{
			name:     "small buffer",
			buffName: "small",
			capacity: 10,
			hasLog:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logger *logrus.Entry
			if tt.hasLog {
				logger = logrus.NewEntry(logrus.New())
			}

			stats := NewBufferStats(tt.buffName, tt.capacity, logger)

			if stats.Name != tt.buffName {
				t.Errorf("Name = %v, want %v", stats.Name, tt.buffName)
			}
			if stats.Capacity != tt.capacity {
				t.Errorf("Capacity = %v, want %v", stats.Capacity, tt.capacity)
			}
			if stats.GetSent() != 0 {
				t.Errorf("Initial sent = %v, want 0", stats.GetSent())
			}
			if stats.GetDropped() != 0 {
				t.Errorf("Initial dropped = %v, want 0", stats.GetDropped())
			}
			if stats.GetCurrentSize() != 0 {
				t.Errorf("Initial size = %v, want 0", stats.GetCurrentSize())
			}
		})
	}
}

func TestBufferStats_RecordSend(t *testing.T) {
	stats := NewBufferStats("test", 100, nil)

	// Send 10 messages
	for i := 0; i < 10; i++ {
		stats.RecordSend()
	}

	if got := stats.GetSent(); got != 10 {
		t.Errorf("GetSent() = %v, want 10", got)
	}
	if got := stats.GetCurrentSize(); got != 10 {
		t.Errorf("GetCurrentSize() = %v, want 10", got)
	}
}

func TestBufferStats_RecordReceive(t *testing.T) {
	stats := NewBufferStats("test", 100, nil)

	// Send 10, receive 5
	for i := 0; i < 10; i++ {
		stats.RecordSend()
	}
	for i := 0; i < 5; i++ {
		stats.RecordReceive()
	}

	if got := stats.GetSent(); got != 10 {
		t.Errorf("GetSent() = %v, want 10", got)
	}
	if got := stats.GetCurrentSize(); got != 5 {
		t.Errorf("GetCurrentSize() = %v, want 5", got)
	}
}

func TestBufferStats_RecordDrop(t *testing.T) {
	stats := NewBufferStats("test", 100, nil)

	// Drop 3 messages
	for i := 0; i < 3; i++ {
		stats.RecordDrop()
	}

	if got := stats.GetDropped(); got != 3 {
		t.Errorf("GetDropped() = %v, want 3", got)
	}
}

func TestBufferStats_Utilization(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		sends    int
		receives int
		want     float64
	}{
		{
			name:     "empty buffer",
			capacity: 100,
			sends:    0,
			receives: 0,
			want:     0.0,
		},
		{
			name:     "half full",
			capacity: 100,
			sends:    50,
			receives: 0,
			want:     0.5,
		},
		{
			name:     "80% full",
			capacity: 100,
			sends:    80,
			receives: 0,
			want:     0.8,
		},
		{
			name:     "full",
			capacity: 100,
			sends:    100,
			receives: 0,
			want:     1.0,
		},
		{
			name:     "with receives",
			capacity: 100,
			sends:    75,
			receives: 25,
			want:     0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := NewBufferStats("test", tt.capacity, nil)

			for i := 0; i < tt.sends; i++ {
				stats.RecordSend()
			}
			for i := 0; i < tt.receives; i++ {
				stats.RecordReceive()
			}

			got := stats.Utilization()
			if got != tt.want {
				t.Errorf("Utilization() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBufferStats_DropRate(t *testing.T) {
	tests := []struct {
		name    string
		sends   int
		drops   int
		want    float64
	}{
		{
			name:  "no messages",
			sends: 0,
			drops: 0,
			want:  0.0,
		},
		{
			name:  "no drops",
			sends: 100,
			drops: 0,
			want:  0.0,
		},
		{
			name:  "10% drop rate",
			sends: 90,
			drops: 10,
			want:  0.1,
		},
		{
			name:  "50% drop rate",
			sends: 50,
			drops: 50,
			want:  0.5,
		},
		{
			name:  "all drops",
			sends: 0,
			drops: 100,
			want:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := NewBufferStats("test", 100, nil)

			for i := 0; i < tt.sends; i++ {
				stats.RecordSend()
			}
			for i := 0; i < tt.drops; i++ {
				stats.RecordDrop()
			}

			got := stats.DropRate()
			if got != tt.want {
				t.Errorf("DropRate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBufferStats_Reset(t *testing.T) {
	stats := NewBufferStats("test", 100, nil)

	// Add some data
	for i := 0; i < 50; i++ {
		stats.RecordSend()
	}
	for i := 0; i < 10; i++ {
		stats.RecordDrop()
	}

	// Verify data exists
	if stats.GetSent() == 0 || stats.GetDropped() == 0 || stats.GetCurrentSize() == 0 {
		t.Fatal("Stats should have data before reset")
	}

	// Reset
	stats.Reset()

	// Verify all zeroed
	if got := stats.GetSent(); got != 0 {
		t.Errorf("After reset, GetSent() = %v, want 0", got)
	}
	if got := stats.GetDropped(); got != 0 {
		t.Errorf("After reset, GetDropped() = %v, want 0", got)
	}
	if got := stats.GetCurrentSize(); got != 0 {
		t.Errorf("After reset, GetCurrentSize() = %v, want 0", got)
	}
	if got := stats.Utilization(); got != 0.0 {
		t.Errorf("After reset, Utilization() = %v, want 0.0", got)
	}
}

func TestBufferStats_Snapshot(t *testing.T) {
	stats := NewBufferStats("test_buffer", 100, nil)

	// Add data
	for i := 0; i < 75; i++ {
		stats.RecordSend()
	}
	for i := 0; i < 25; i++ {
		stats.RecordReceive()
	}
	for i := 0; i < 10; i++ {
		stats.RecordDrop()
	}

	snapshot := stats.Snapshot()

	if snapshot.Name != "test_buffer" {
		t.Errorf("Snapshot.Name = %v, want test_buffer", snapshot.Name)
	}
	if snapshot.Capacity != 100 {
		t.Errorf("Snapshot.Capacity = %v, want 100", snapshot.Capacity)
	}
	if snapshot.CurrentSize != 50 {
		t.Errorf("Snapshot.CurrentSize = %v, want 50", snapshot.CurrentSize)
	}
	if snapshot.Sent != 75 {
		t.Errorf("Snapshot.Sent = %v, want 75", snapshot.Sent)
	}
	if snapshot.Dropped != 10 {
		t.Errorf("Snapshot.Dropped = %v, want 10", snapshot.Dropped)
	}
	if snapshot.Utilization != 0.5 {
		t.Errorf("Snapshot.Utilization = %v, want 0.5", snapshot.Utilization)
	}
	// DropRate = 10 / (75 + 10) = 10/85 ≈ 0.1176
	expectedDropRate := 10.0 / 85.0
	if snapshot.DropRate != expectedDropRate {
		t.Errorf("Snapshot.DropRate = %v, want %v", snapshot.DropRate, expectedDropRate)
	}
}

func TestBufferStats_ConcurrentAccess(t *testing.T) {
	stats := NewBufferStats("concurrent_test", 1000, nil)

	var wg sync.WaitGroup
	numGoroutines := 10
	operationsPerGoroutine := 100

	// Concurrent sends
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				stats.RecordSend()
			}
		}()
	}

	// Concurrent receives
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine; j++ {
				stats.RecordReceive()
			}
		}()
	}

	// Concurrent drops
	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < operationsPerGoroutine/10; j++ {
				stats.RecordDrop()
			}
		}()
	}

	wg.Wait()

	// Verify expected values
	expectedSent := uint64(numGoroutines * operationsPerGoroutine)
	if got := stats.GetSent(); got != expectedSent {
		t.Errorf("GetSent() = %v, want %v", got, expectedSent)
	}

	expectedDropped := uint64((numGoroutines / 2) * (operationsPerGoroutine / 10))
	if got := stats.GetDropped(); got != expectedDropped {
		t.Errorf("GetDropped() = %v, want %v", got, expectedDropped)
	}

	// Size should be sends - receives
	expectedSize := (numGoroutines * operationsPerGoroutine) - ((numGoroutines / 2) * operationsPerGoroutine)
	if got := stats.GetCurrentSize(); got != expectedSize {
		t.Errorf("GetCurrentSize() = %v, want %v", got, expectedSize)
	}
}

func TestBufferStats_WarnThreshold(t *testing.T) {
	// Create logger to capture warnings
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	entry := logrus.NewEntry(logger)

	stats := NewBufferStats("warn_test", 100, entry)

	// Fill to 79% - should not warn
	for i := 0; i < 79; i++ {
		stats.RecordSend()
	}

	// Fill to 80% - should warn (threshold is 0.8)
	stats.RecordSend()

	// Verify utilization is at threshold
	if got := stats.Utilization(); got < 0.8 {
		t.Errorf("Utilization() = %v, want >= 0.8", got)
	}

	// Send more - warning should not spam (lastWarnSize tracking)
	stats.RecordSend()
	stats.RecordSend()

	// Receive messages to drop below threshold
	for i := 0; i < 30; i++ {
		stats.RecordReceive()
	}

	// Send again to cross threshold - should warn again
	for i := 0; i < 30; i++ {
		stats.RecordSend()
	}
}
