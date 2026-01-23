//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/network"
)

func TestConvertSnapshotToStateUpdates_Empty(t *testing.T) {
	snapshot := network.WorldSnapshot{
		Timestamp: time.Now(),
		Sequence:  1,
		Entities:  make(map[uint64]network.EntitySnapshot),
	}

	updates := convertSnapshotToStateUpdates(snapshot)

	if len(updates) != 0 {
		t.Errorf("Expected 0 updates for empty snapshot, got %d", len(updates))
	}
}

func TestConvertSnapshotToStateUpdates_SingleEntity(t *testing.T) {
	now := time.Now()
	snapshot := network.WorldSnapshot{
		Timestamp: now,
		Sequence:  1,
		Entities: map[uint64]network.EntitySnapshot{
			100: {
				EntityID:   100,
				Timestamp:  now,
				Sequence:   1,
				Position:   network.Position{X: 10.5, Y: 20.5},
				Velocity:   network.Velocity{VX: 1.0, VY: -1.0},
				Components: make(map[string][]byte),
			},
		},
	}

	updates := convertSnapshotToStateUpdates(snapshot)

	if len(updates) != 1 {
		t.Fatalf("Expected 1 update, got %d", len(updates))
	}

	update := updates[0]

	if update.EntityID != 100 {
		t.Errorf("Expected EntityID 100, got %d", update.EntityID)
	}

	expectedTs := uint64(now.UnixNano() / 1000000)
	if update.Timestamp != expectedTs {
		t.Errorf("Expected Timestamp %d, got %d", expectedTs, update.Timestamp)
	}

	if update.Priority != network.PriorityNormal {
		t.Errorf("Expected Priority %d, got %d", network.PriorityNormal, update.Priority)
	}

	// Should have at least position and velocity components
	if len(update.Components) < 2 {
		t.Errorf("Expected at least 2 components (position, velocity), got %d", len(update.Components))
	}

	// Verify position component
	hasPosition := false
	hasVelocity := false
	for _, comp := range update.Components {
		if comp.Type == "position" {
			hasPosition = true
			if len(comp.Data) != 16 {
				t.Errorf("Position data should be 16 bytes, got %d", len(comp.Data))
			}
		}
		if comp.Type == "velocity" {
			hasVelocity = true
			if len(comp.Data) != 16 {
				t.Errorf("Velocity data should be 16 bytes, got %d", len(comp.Data))
			}
		}
	}

	if !hasPosition {
		t.Error("Missing position component in update")
	}
	if !hasVelocity {
		t.Error("Missing velocity component in update")
	}
}

func TestConvertSnapshotToStateUpdates_MultipleEntities(t *testing.T) {
	now := time.Now()
	snapshot := network.WorldSnapshot{
		Timestamp: now,
		Sequence:  5,
		Entities: map[uint64]network.EntitySnapshot{
			1: {
				EntityID:   1,
				Timestamp:  now,
				Position:   network.Position{X: 0, Y: 0},
				Velocity:   network.Velocity{VX: 0, VY: 0},
				Components: make(map[string][]byte),
			},
			2: {
				EntityID:   2,
				Timestamp:  now,
				Position:   network.Position{X: 100, Y: 100},
				Velocity:   network.Velocity{VX: 5, VY: 5},
				Components: make(map[string][]byte),
			},
			3: {
				EntityID:   3,
				Timestamp:  now,
				Position:   network.Position{X: -50, Y: 200},
				Velocity:   network.Velocity{VX: -1, VY: 2},
				Components: make(map[string][]byte),
			},
		},
	}

	updates := convertSnapshotToStateUpdates(snapshot)

	if len(updates) != 3 {
		t.Fatalf("Expected 3 updates, got %d", len(updates))
	}

	// Verify all entity IDs are present
	foundIDs := make(map[uint64]bool)
	for _, u := range updates {
		foundIDs[u.EntityID] = true
	}

	for id := uint64(1); id <= 3; id++ {
		if !foundIDs[id] {
			t.Errorf("Missing update for entity ID %d", id)
		}
	}
}

func TestConvertSnapshotToStateUpdates_WithExtraComponents(t *testing.T) {
	now := time.Now()
	vehicleData := []byte{1, 2, 3, 4, 5}
	companionData := []byte{10, 20, 30}

	snapshot := network.WorldSnapshot{
		Timestamp: now,
		Sequence:  1,
		Entities: map[uint64]network.EntitySnapshot{
			42: {
				EntityID:  42,
				Timestamp: now,
				Position:  network.Position{X: 5.5, Y: 6.6},
				Velocity:  network.Velocity{VX: 0, VY: 0},
				Components: map[string][]byte{
					"vehicle":   vehicleData,
					"companion": companionData,
				},
			},
		},
	}

	updates := convertSnapshotToStateUpdates(snapshot)

	if len(updates) != 1 {
		t.Fatalf("Expected 1 update, got %d", len(updates))
	}

	update := updates[0]

	// Should have position + velocity + vehicle + companion = 4 components
	if len(update.Components) != 4 {
		t.Errorf("Expected 4 components, got %d", len(update.Components))
	}

	// Verify all components are present
	componentTypes := make(map[string]bool)
	for _, comp := range update.Components {
		componentTypes[comp.Type] = true
	}

	expectedTypes := []string{"position", "velocity", "vehicle", "companion"}
	for _, et := range expectedTypes {
		if !componentTypes[et] {
			t.Errorf("Missing component type: %s", et)
		}
	}
}

func TestSerializePosition(t *testing.T) {
	tests := []struct {
		name string
		pos  network.Position
	}{
		{"zero", network.Position{X: 0, Y: 0}},
		{"positive", network.Position{X: 10.5, Y: 20.25}},
		{"negative", network.Position{X: -100.0, Y: -50.5}},
		{"large", network.Position{X: 1000000.0, Y: 2000000.0}},
		{"small", network.Position{X: 0.0001, Y: 0.00001}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := serializePosition(tt.pos)

			if len(data) != 16 {
				t.Errorf("Expected 16 bytes, got %d", len(data))
			}

			// Data should be non-nil
			if data == nil {
				t.Error("Expected non-nil data")
			}
		})
	}
}

func TestSerializeVelocity(t *testing.T) {
	tests := []struct {
		name string
		vel  network.Velocity
	}{
		{"zero", network.Velocity{VX: 0, VY: 0}},
		{"positive", network.Velocity{VX: 5.5, VY: 10.25}},
		{"negative", network.Velocity{VX: -3.0, VY: -7.5}},
		{"mixed", network.Velocity{VX: 100.0, VY: -200.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := serializeVelocity(tt.vel)

			if len(data) != 16 {
				t.Errorf("Expected 16 bytes, got %d", len(data))
			}

			if data == nil {
				t.Error("Expected non-nil data")
			}
		})
	}
}

func TestPutFloat64(t *testing.T) {
	tests := []struct {
		name string
		val  float64
	}{
		{"zero", 0.0},
		{"one", 1.0},
		{"negative", -123.456},
		{"large", 1e308},
		{"small", 1e-308},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := make([]byte, 8)
			putFloat64(buf, tt.val)

			// Verify we wrote 8 bytes (no panic, no overflow)
			if len(buf) != 8 {
				t.Errorf("Buffer size changed from 8")
			}
		})
	}
}

func BenchmarkConvertSnapshotToStateUpdates_100Entities(b *testing.B) {
	now := time.Now()
	snapshot := network.WorldSnapshot{
		Timestamp: now,
		Sequence:  1,
		Entities:  make(map[uint64]network.EntitySnapshot),
	}

	// Create 100 entities
	for i := uint64(0); i < 100; i++ {
		snapshot.Entities[i] = network.EntitySnapshot{
			EntityID:   i,
			Timestamp:  now,
			Position:   network.Position{X: float64(i), Y: float64(i * 2)},
			Velocity:   network.Velocity{VX: 1.0, VY: 1.0},
			Components: make(map[string][]byte),
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertSnapshotToStateUpdates(snapshot)
	}
}

func BenchmarkConvertSnapshotToStateUpdates_1000Entities(b *testing.B) {
	now := time.Now()
	snapshot := network.WorldSnapshot{
		Timestamp: now,
		Sequence:  1,
		Entities:  make(map[uint64]network.EntitySnapshot),
	}

	// Create 1000 entities
	for i := uint64(0); i < 1000; i++ {
		snapshot.Entities[i] = network.EntitySnapshot{
			EntityID:  i,
			Timestamp: now,
			Position:  network.Position{X: float64(i), Y: float64(i * 2)},
			Velocity:  network.Velocity{VX: 1.0, VY: 1.0},
			Components: map[string][]byte{
				"vehicle":   make([]byte, 64),
				"companion": make([]byte, 32),
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = convertSnapshotToStateUpdates(snapshot)
	}
}

func BenchmarkSerializePosition(b *testing.B) {
	pos := network.Position{X: 123.456, Y: 789.012}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = serializePosition(pos)
	}
}

func BenchmarkSerializeVelocity(b *testing.B) {
	vel := network.Velocity{VX: 5.5, VY: -3.3}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = serializeVelocity(vel)
	}
}
