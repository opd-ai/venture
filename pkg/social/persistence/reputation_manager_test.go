package persistence

import (
	"testing"
	"time"
)

func TestNewReputationManager(t *testing.T) {
	rm := NewReputationManager()
	if rm == nil {
		t.Fatal("NewReputationManager returned nil")
	}
	if rm.records == nil {
		t.Error("ReputationManager.records not initialized")
	}
	if len(rm.records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(rm.records))
	}
}

func TestUpdateReputation(t *testing.T) {
	rm := NewReputationManager()
	now := time.Now()

	tests := []struct {
		name      string
		playerID  string
		category  ReputationCategory
		delta     float64
		timestamp time.Time
	}{
		{"trade positive", "alice", ReputationTrade, 10.0, now},
		{"trade negative", "alice", ReputationTrade, -5.0, now},
		{"combat positive", "alice", ReputationCombat, 20.0, now},
		{"social positive", "bob", ReputationSocial, 15.0, now},
		{"quest positive", "bob", ReputationQuest, 25.0, now},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := rm.UpdateReputation(tt.playerID, tt.category, tt.delta, tt.timestamp)
			if err != nil {
				t.Errorf("UpdateReputation() error = %v", err)
			}
		})
	}
}

func TestGetReputation(t *testing.T) {
	rm := NewReputationManager()
	now := time.Now()

	// Test default reputation (no record)
	rep := rm.GetReputation("alice", ReputationTrade)
	if rep != 0.0 {
		t.Errorf("Expected default reputation 0.0, got %f", rep)
	}

	// Create reputation record
	rm.UpdateReputation("alice", ReputationTrade, 10.0, now)
	rep = rm.GetReputation("alice", ReputationTrade)
	if rep != 10.0 {
		t.Errorf("Expected reputation 10.0, got %f", rep)
	}

	// Test multiple categories
	rm.UpdateReputation("alice", ReputationCombat, 20.0, now)
	repTrade := rm.GetReputation("alice", ReputationTrade)
	repCombat := rm.GetReputation("alice", ReputationCombat)

	if repTrade != 10.0 {
		t.Errorf("Expected trade reputation 10.0, got %f", repTrade)
	}
	if repCombat != 20.0 {
		t.Errorf("Expected combat reputation 20.0, got %f", repCombat)
	}
}

func TestGetTotalReputation(t *testing.T) {
	rm := NewReputationManager()
	now := time.Now()

	// Test default total (no record)
	total := rm.GetTotalReputation("alice")
	if total != 0.0 {
		t.Errorf("Expected default total 0.0, got %f", total)
	}

	// Create multiple category scores
	rm.UpdateReputation("alice", ReputationTrade, 10.0, now)
	rm.UpdateReputation("alice", ReputationCombat, 20.0, now)
	rm.UpdateReputation("alice", ReputationSocial, 30.0, now)
	rm.UpdateReputation("alice", ReputationQuest, 40.0, now)

	// Total should be average: (10+20+30+40)/4 = 25.0
	total = rm.GetTotalReputation("alice")
	expected := 25.0
	if total != expected {
		t.Errorf("Expected total reputation %f, got %f", expected, total)
	}
}

func TestGetReputationRecord(t *testing.T) {
	rm := NewReputationManager()
	now := time.Now()

	// Test non-existent record
	record := rm.GetReputationRecord("alice")
	if record != nil {
		t.Error("Expected nil for non-existent record")
	}

	// Create record with multiple categories
	rm.UpdateReputation("alice", ReputationTrade, 10.0, now)
	rm.UpdateReputation("alice", ReputationCombat, 20.0, now)

	record = rm.GetReputationRecord("alice")
	if record == nil {
		t.Fatal("Expected record, got nil")
	}

	// Verify fields
	if record.PlayerID != "alice" {
		t.Errorf("Expected PlayerID 'alice', got '%s'", record.PlayerID)
	}
	if len(record.Scores) != 2 {
		t.Errorf("Expected 2 category scores, got %d", len(record.Scores))
	}
	if record.TotalScore != 15.0 { // (10+20)/2
		t.Errorf("Expected total score 15.0, got %f", record.TotalScore)
	}

	// Verify copy (modification shouldn't affect original)
	record.TotalScore = 0.0
	total := rm.GetTotalReputation("alice")
	if total == 0.0 {
		t.Error("Record copy modification affected original")
	}
}

func TestReputationDecay(t *testing.T) {
	rm := NewReputationManager()
	baseTime := time.Now()

	// Create records with different ages
	rm.UpdateReputation("alice", ReputationTrade, 100.0, baseTime.Add(-10*24*time.Hour))
	rm.UpdateReputation("alice", ReputationCombat, 100.0, baseTime.Add(-5*24*time.Hour))
	rm.UpdateReputation("alice", ReputationSocial, 100.0, baseTime) // Fresh

	// Apply decay
	decayed := rm.ApplyDecay(baseTime)
	if decayed != 2 {
		t.Errorf("Expected 2 decayed scores, got %d", decayed)
	}

	// Verify decay amounts
	repTrade := rm.GetReputation("alice", ReputationTrade)
	expectedTrade := 100.0 - (DecayRatePerDay * 10) // 100 - 0.1 = 99.9
	if repTrade != expectedTrade {
		t.Errorf("Expected trade reputation %f, got %f", expectedTrade, repTrade)
	}

	repCombat := rm.GetReputation("alice", ReputationCombat)
	expectedCombat := 100.0 - (DecayRatePerDay * 5) // 100 - 0.05 = 99.95
	if repCombat != expectedCombat {
		t.Errorf("Expected combat reputation %f, got %f", expectedCombat, repCombat)
	}

	repSocial := rm.GetReputation("alice", ReputationSocial)
	if repSocial != 100.0 {
		t.Errorf("Expected no decay for fresh record, got %f", repSocial)
	}
}

func TestReputationDecayFloor(t *testing.T) {
	rm := NewReputationManager()
	baseTime := time.Now()

	// Create record with very low reputation that should decay to floor
	rm.UpdateReputation("alice", ReputationTrade, 1.0, baseTime.Add(-365*24*time.Hour)) // 1 year old, starts at 1.0

	// Apply decay (1.0 - (365 * 0.01) = 1.0 - 3.65 = -2.65, clamped to 0.0)
	rm.ApplyDecay(baseTime)

	rep := rm.GetReputation("alice", ReputationTrade)
	if rep < 0.0 {
		t.Errorf("Reputation decayed below 0: %f", rep)
	}
	if rep != 0.0 {
		t.Errorf("Expected reputation 0.0 (clamped), got %f", rep)
	}
}

func TestReputationDecayTotalRecalculation(t *testing.T) {
	rm := NewReputationManager()
	baseTime := time.Now()

	// Create multiple category scores
	rm.UpdateReputation("alice", ReputationTrade, 100.0, baseTime.Add(-10*24*time.Hour))
	rm.UpdateReputation("alice", ReputationCombat, 100.0, baseTime.Add(-10*24*time.Hour))

	// Apply decay
	rm.ApplyDecay(baseTime)

	// Verify total recalculated correctly
	expectedPerCategory := 100.0 - (DecayRatePerDay * 10) // 99.9
	expectedTotal := expectedPerCategory                  // Both categories have same value

	total := rm.GetTotalReputation("alice")
	if total != expectedTotal {
		t.Errorf("Expected total %f, got %f", expectedTotal, total)
	}
}

func TestReputationGetRecordCount(t *testing.T) {
	rm := NewReputationManager()
	now := time.Now()

	if rm.GetRecordCount() != 0 {
		t.Error("Expected 0 records initially")
	}

	rm.UpdateReputation("alice", ReputationTrade, 10.0, now)
	if rm.GetRecordCount() != 1 {
		t.Errorf("Expected 1 record, got %d", rm.GetRecordCount())
	}

	// Adding another category for same player shouldn't increase record count
	rm.UpdateReputation("alice", ReputationCombat, 20.0, now)
	if rm.GetRecordCount() != 1 {
		t.Errorf("Expected 1 record, got %d", rm.GetRecordCount())
	}

	// Adding record for different player should increase count
	rm.UpdateReputation("bob", ReputationTrade, 10.0, now)
	if rm.GetRecordCount() != 2 {
		t.Errorf("Expected 2 records, got %d", rm.GetRecordCount())
	}
}

func TestReputationSaveLoad(t *testing.T) {
	rm := NewReputationManager()
	now := time.Now()

	// Create test data
	rm.UpdateReputation("alice", ReputationTrade, 50.0, now)
	rm.UpdateReputation("alice", ReputationCombat, 75.0, now)
	rm.UpdateReputation("bob", ReputationSocial, 100.0, now)
	rm.UpdateReputation("bob", ReputationQuest, 25.0, now)

	// Save
	data, err := rm.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Save returned empty data")
	}

	// Load into new manager
	rm2 := NewReputationManager()
	if err := rm2.Load(data); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify counts match
	if rm2.GetRecordCount() != rm.GetRecordCount() {
		t.Errorf("Record count mismatch: %d != %d", rm2.GetRecordCount(), rm.GetRecordCount())
	}

	// Verify reputation scores match
	tests := []struct {
		playerID string
		category ReputationCategory
		expected float64
	}{
		{"alice", ReputationTrade, 50.0},
		{"alice", ReputationCombat, 75.0},
		{"bob", ReputationSocial, 100.0},
		{"bob", ReputationQuest, 25.0},
	}

	for _, tt := range tests {
		rep1 := rm.GetReputation(tt.playerID, tt.category)
		rep2 := rm2.GetReputation(tt.playerID, tt.category)
		if rep1 != rep2 {
			t.Errorf("Reputation mismatch for %s/%s: %f != %f", tt.playerID, tt.category, rep1, rep2)
		}
	}

	// Verify total reputation match
	totalAlice1 := rm.GetTotalReputation("alice")
	totalAlice2 := rm2.GetTotalReputation("alice")
	if totalAlice1 != totalAlice2 {
		t.Errorf("Total reputation mismatch for alice: %f != %f", totalAlice1, totalAlice2)
	}
}

func TestReputationCompression(t *testing.T) {
	rm := NewReputationManager()
	now := time.Now()

	// Create many records to test compression
	players := []string{"alice", "bob", "charlie", "david", "eve", "frank", "grace", "henry", "ivan", "judy"}
	categories := []ReputationCategory{ReputationTrade, ReputationCombat, ReputationSocial, ReputationQuest}

	for _, player := range players {
		for _, category := range categories {
			rm.UpdateReputation(player, category, 100.0, now)
		}
	}

	data, err := rm.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// With 10 players × 4 categories = 40 scores, compressed size should be <15KB
	if len(data) > 15360 {
		t.Errorf("Compressed data too large: %d bytes", len(data))
	}
}

func TestReputationClear(t *testing.T) {
	rm := NewReputationManager()
	now := time.Now()

	// Create records
	rm.UpdateReputation("alice", ReputationTrade, 10.0, now)
	rm.UpdateReputation("alice", ReputationCombat, 20.0, now)
	rm.UpdateReputation("bob", ReputationSocial, 30.0, now)

	// Clear
	rm.Clear()

	if rm.GetRecordCount() != 0 {
		t.Errorf("Expected 0 records after Clear, got %d", rm.GetRecordCount())
	}

	// Verify reputation reset to default
	rep := rm.GetReputation("alice", ReputationTrade)
	if rep != 0.0 {
		t.Errorf("Expected default reputation after Clear, got %f", rep)
	}

	total := rm.GetTotalReputation("alice")
	if total != 0.0 {
		t.Errorf("Expected default total after Clear, got %f", total)
	}
}

func TestReputationConcurrentAccess(t *testing.T) {
	rm := NewReputationManager()
	now := time.Now()

	// Spawn multiple goroutines to test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			playerID := "player"
			category := ReputationTrade

			// Mix of reads and writes
			for j := 0; j < 100; j++ {
				if j%2 == 0 {
					rm.UpdateReputation(playerID, category, 1.0, now)
				} else {
					_ = rm.GetReputation(playerID, category)
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify manager is still functional
	rep := rm.GetReputation("player", ReputationTrade)
	if rep < 0.0 {
		t.Errorf("Reputation below 0 after concurrent access: %f", rep)
	}
}

func TestReputationCategoryString(t *testing.T) {
	tests := []struct {
		category ReputationCategory
		expected string
	}{
		{ReputationTrade, "trade"},
		{ReputationCombat, "combat"},
		{ReputationSocial, "social"},
		{ReputationQuest, "quest"},
	}

	for _, tt := range tests {
		if string(tt.category) != tt.expected {
			t.Errorf("Category string mismatch: got %s, want %s", tt.category, tt.expected)
		}
	}
}

// Benchmarks

func BenchmarkUpdateReputation(b *testing.B) {
	rm := NewReputationManager()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.UpdateReputation("alice", ReputationTrade, 1.0, now)
	}
}

func BenchmarkGetReputation(b *testing.B) {
	rm := NewReputationManager()
	now := time.Now()
	rm.UpdateReputation("alice", ReputationTrade, 100.0, now)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rm.GetReputation("alice", ReputationTrade)
	}
}

func BenchmarkGetTotalReputation(b *testing.B) {
	rm := NewReputationManager()
	now := time.Now()

	categories := []ReputationCategory{ReputationTrade, ReputationCombat, ReputationSocial, ReputationQuest}
	for _, category := range categories {
		rm.UpdateReputation("alice", category, 100.0, now)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rm.GetTotalReputation("alice")
	}
}

func BenchmarkReputationApplyDecay(b *testing.B) {
	rm := NewReputationManager()
	baseTime := time.Now()

	// Create 10,000 reputation scores (2,500 players × 4 categories)
	for i := 0; i < 2500; i++ {
		playerID := "player_" + string(rune(i))
		categories := []ReputationCategory{ReputationTrade, ReputationCombat, ReputationSocial, ReputationQuest}
		for _, category := range categories {
			rm.UpdateReputation(playerID, category, 100.0, baseTime.Add(-10*24*time.Hour))
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.ApplyDecay(baseTime)
	}
}

func BenchmarkReputationSave(b *testing.B) {
	rm := NewReputationManager()
	now := time.Now()

	// Create 1000 players with 4 categories each
	for i := 0; i < 1000; i++ {
		playerID := "player_" + string(rune(i))
		categories := []ReputationCategory{ReputationTrade, ReputationCombat, ReputationSocial, ReputationQuest}
		for _, category := range categories {
			rm.UpdateReputation(playerID, category, 100.0, now)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := rm.Save()
		if err != nil {
			b.Fatal(err)
		}
	}
}
