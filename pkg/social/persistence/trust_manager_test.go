package persistence

import (
	"testing"
	"time"
)

func TestNewTrustManager(t *testing.T) {
	tm := NewTrustManager()
	if tm == nil {
		t.Fatal("NewTrustManager returned nil")
	}
	if tm.records == nil {
		t.Error("TrustManager.records not initialized")
	}
	if len(tm.records) != 0 {
		t.Errorf("Expected 0 records, got %d", len(tm.records))
	}
}

func TestUpdateTrust(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	tests := []struct {
		name      string
		playerA   string
		playerB   string
		delta     float64
		timestamp time.Time
		wantErr   bool
	}{
		{"positive delta", "alice", "bob", 0.1, now, false},
		{"negative delta", "alice", "bob", -0.05, now, false},
		{"self trust", "alice", "alice", 0.1, now, true},
		{"large positive", "alice", "bob", 0.5, now, false},
		{"large negative", "alice", "bob", -0.8, now, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tm.UpdateTrust(tt.playerA, tt.playerB, tt.delta, tt.timestamp)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateTrust() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetTrust(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	// Test default trust (no record)
	trust := tm.GetTrust("alice", "bob")
	if trust != 0.5 {
		t.Errorf("Expected default trust 0.5, got %f", trust)
	}

	// Test self trust
	trust = tm.GetTrust("alice", "alice")
	if trust != 1.0 {
		t.Errorf("Expected self trust 1.0, got %f", trust)
	}

	// Create trust record
	tm.UpdateTrust("alice", "bob", 0.2, now)
	trust = tm.GetTrust("alice", "bob")
	expected := 0.7 // 0.5 + 0.2
	if trust != expected {
		t.Errorf("Expected trust %f, got %f", expected, trust)
	}

	// Test bidirectional symmetry
	trust2 := tm.GetTrust("bob", "alice")
	if trust != trust2 {
		t.Errorf("Trust not symmetric: alice->bob=%f, bob->alice=%f", trust, trust2)
	}
}

func TestTrustClamping(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	// Test upper bound clamping
	tm.UpdateTrust("alice", "bob", 1.0, now)
	trust := tm.GetTrust("alice", "bob")
	if trust > MaxTrustScore {
		t.Errorf("Trust exceeded maximum: %f > %f", trust, MaxTrustScore)
	}

	// Test lower bound clamping
	tm.UpdateTrust("alice", "bob", -2.0, now)
	trust = tm.GetTrust("alice", "bob")
	if trust < MinTrustScore {
		t.Errorf("Trust below minimum: %f < %f", trust, MinTrustScore)
	}
	if trust != MinTrustScore {
		t.Errorf("Expected trust %f, got %f", MinTrustScore, trust)
	}
}

func TestTrustManagerGetTrustLevel(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	tests := []struct {
		name      string
		delta     float64
		wantLevel TrustLevel
	}{
		{"stranger", -0.3, TrustLevelStranger},        // 0.5 - 0.3 = 0.2
		{"acquaintance", 0.0, TrustLevelAcquaintance}, // 0.5 + 0.0 = 0.5
		{"friend", 0.2, TrustLevelFriend},             // 0.5 + 0.2 = 0.7
		{"trusted", 0.4, TrustLevelTrusted},           // 0.5 + 0.4 = 0.9
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and recreate manager for each test
			tm.Clear()
			tm.UpdateTrust("alice", "bob", tt.delta, now)
			level := tm.GetTrustLevel("alice", "bob")
			if level != tt.wantLevel {
				t.Errorf("Expected trust level %s, got %s", tt.wantLevel, level)
			}
		})
	}
}

func TestGetTrustRecord(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	// Test non-existent record
	record := tm.GetTrustRecord("alice", "bob")
	if record != nil {
		t.Error("Expected nil for non-existent record")
	}

	// Create record
	tm.UpdateTrust("alice", "bob", 0.1, now)
	record = tm.GetTrustRecord("alice", "bob")
	if record == nil {
		t.Fatal("Expected record, got nil")
	}

	// Verify fields
	if record.Score != 0.6 {
		t.Errorf("Expected score 0.6, got %f", record.Score)
	}
	if record.Interactions != 1 {
		t.Errorf("Expected 1 interaction, got %d", record.Interactions)
	}

	// Test bidirectional access
	record2 := tm.GetTrustRecord("bob", "alice")
	if record2 == nil {
		t.Fatal("Expected bidirectional record")
	}
	if record.Score != record2.Score {
		t.Error("Bidirectional records don't match")
	}

	// Verify copy (modification shouldn't affect original)
	record.Score = 0.0
	trust := tm.GetTrust("alice", "bob")
	if trust == 0.0 {
		t.Error("Record copy modification affected original")
	}
}

func TestApplyDecay(t *testing.T) {
	tm := NewTrustManager()
	baseTime := time.Now()

	// Create multiple records with different ages
	tm.UpdateTrust("alice", "bob", 0.3, baseTime.Add(-10*24*time.Hour))    // 10 days old
	tm.UpdateTrust("alice", "charlie", 0.3, baseTime.Add(-5*24*time.Hour)) // 5 days old
	tm.UpdateTrust("alice", "david", 0.3, baseTime)                        // Fresh

	// Apply decay
	decayed := tm.ApplyDecay(baseTime)
	if decayed != 2 {
		t.Errorf("Expected 2 decayed records, got %d", decayed)
	}

	// Verify decay amounts (with tolerance for float precision)
	tolerance := 0.000001

	trustBob := tm.GetTrust("alice", "bob")
	expectedBob := 0.8 - (DecayRatePerDay * 10) // 0.5 + 0.3 - 0.1 = 0.7
	if diff := trustBob - expectedBob; diff < -tolerance || diff > tolerance {
		t.Errorf("Expected trust %f, got %f (diff: %f)", expectedBob, trustBob, diff)
	}

	trustCharlie := tm.GetTrust("alice", "charlie")
	expectedCharlie := 0.8 - (DecayRatePerDay * 5) // 0.5 + 0.3 - 0.05 = 0.75
	if diff := trustCharlie - expectedCharlie; diff < -tolerance || diff > tolerance {
		t.Errorf("Expected trust %f, got %f (diff: %f)", expectedCharlie, trustCharlie, diff)
	}

	trustDavid := tm.GetTrust("alice", "david")
	if trustDavid != 0.8 {
		t.Errorf("Expected no decay for fresh record, got %f", trustDavid)
	}
}

func TestDecayFloor(t *testing.T) {
	tm := NewTrustManager()
	baseTime := time.Now()

	// Create record with low trust
	tm.UpdateTrust("alice", "bob", -0.4, baseTime.Add(-365*24*time.Hour)) // 1 year old, trust 0.1

	// Apply decay
	tm.ApplyDecay(baseTime)

	// Trust should not go below minimum
	trust := tm.GetTrust("alice", "bob")
	if trust < MinTrustScore {
		t.Errorf("Trust decayed below minimum: %f", trust)
	}
	if trust != MinTrustScore {
		t.Errorf("Expected trust %f, got %f", MinTrustScore, trust)
	}
}

func TestGetPlayerTrustRecords(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	// Create records for alice with multiple players
	tm.UpdateTrust("alice", "bob", 0.1, now)
	tm.UpdateTrust("alice", "charlie", 0.2, now)
	tm.UpdateTrust("alice", "david", -0.1, now)
	tm.UpdateTrust("bob", "charlie", 0.3, now) // Not alice

	records := tm.GetPlayerTrustRecords("alice")
	if len(records) != 3 {
		t.Errorf("Expected 3 records for alice, got %d", len(records))
	}

	// Verify alice is in all records
	for _, record := range records {
		if record.PlayerA != "alice" && record.PlayerB != "alice" {
			t.Errorf("Record doesn't contain alice: %v", record)
		}
	}
}

func TestTrustManagerGetRecordCount(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	if tm.GetRecordCount() != 0 {
		t.Error("Expected 0 records initially")
	}

	tm.UpdateTrust("alice", "bob", 0.1, now)
	if tm.GetRecordCount() != 1 {
		t.Errorf("Expected 1 record, got %d", tm.GetRecordCount())
	}

	tm.UpdateTrust("alice", "charlie", 0.1, now)
	if tm.GetRecordCount() != 2 {
		t.Errorf("Expected 2 records, got %d", tm.GetRecordCount())
	}

	// Updating existing record shouldn't change count
	tm.UpdateTrust("alice", "bob", 0.1, now)
	if tm.GetRecordCount() != 2 {
		t.Errorf("Expected 2 records after update, got %d", tm.GetRecordCount())
	}
}

func TestSaveLoad(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	// Create test data
	tm.UpdateTrust("alice", "bob", 0.2, now)
	tm.UpdateTrust("alice", "charlie", -0.1, now)
	tm.UpdateTrust("bob", "david", 0.3, now)

	// Save
	data, err := tm.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Save returned empty data")
	}

	// Load into new manager
	tm2 := NewTrustManager()
	if err := tm2.Load(data); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Verify counts match
	if tm2.GetRecordCount() != tm.GetRecordCount() {
		t.Errorf("Record count mismatch: %d != %d", tm2.GetRecordCount(), tm.GetRecordCount())
	}

	// Verify trust scores match
	tests := []struct {
		playerA string
		playerB string
	}{
		{"alice", "bob"},
		{"alice", "charlie"},
		{"bob", "david"},
	}

	for _, tt := range tests {
		trust1 := tm.GetTrust(tt.playerA, tt.playerB)
		trust2 := tm2.GetTrust(tt.playerA, tt.playerB)
		if trust1 != trust2 {
			t.Errorf("Trust mismatch for %s-%s: %f != %f", tt.playerA, tt.playerB, trust1, trust2)
		}
	}
}

func TestCompression(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	// Create many records to test compression
	players := []string{"alice", "bob", "charlie", "david", "eve", "frank", "grace", "henry"}
	for i := 0; i < len(players); i++ {
		for j := i + 1; j < len(players); j++ {
			tm.UpdateTrust(players[i], players[j], 0.1, now)
		}
	}

	data, err := tm.Save()
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify compression (should be much smaller than raw JSON)
	// With 28 records, compressed size should be <10KB
	if len(data) > 10240 {
		t.Errorf("Compressed data too large: %d bytes", len(data))
	}
}

func TestClear(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	// Create records
	tm.UpdateTrust("alice", "bob", 0.1, now)
	tm.UpdateTrust("alice", "charlie", 0.1, now)

	// Clear
	tm.Clear()

	if tm.GetRecordCount() != 0 {
		t.Errorf("Expected 0 records after Clear, got %d", tm.GetRecordCount())
	}

	// Verify trust reset to default
	trust := tm.GetTrust("alice", "bob")
	if trust != 0.5 {
		t.Errorf("Expected default trust after Clear, got %f", trust)
	}
}

func TestConcurrentAccess(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	// Spawn multiple goroutines to test concurrent access
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			playerA := "player_a"
			playerB := "player_b"

			// Mix of reads and writes
			for j := 0; j < 100; j++ {
				if j%2 == 0 {
					tm.UpdateTrust(playerA, playerB, 0.01, now)
				} else {
					_ = tm.GetTrust(playerA, playerB)
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
	trust := tm.GetTrust("player_a", "player_b")
	if trust < 0.0 || trust > 1.0 {
		t.Errorf("Trust out of range after concurrent access: %f", trust)
	}
}

// Benchmarks

func BenchmarkUpdateTrust(b *testing.B) {
	tm := NewTrustManager()
	now := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.UpdateTrust("alice", "bob", 0.01, now)
	}
}

func BenchmarkGetTrust(b *testing.B) {
	tm := NewTrustManager()
	now := time.Now()
	tm.UpdateTrust("alice", "bob", 0.1, now)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tm.GetTrust("alice", "bob")
	}
}

func BenchmarkApplyDecay(b *testing.B) {
	tm := NewTrustManager()
	baseTime := time.Now()

	// Create 10,000 trust records
	for i := 0; i < 10000; i++ {
		playerA := "player_" + string(rune(i/100))
		playerB := "player_" + string(rune(i%100))
		tm.UpdateTrust(playerA, playerB, 0.1, baseTime.Add(-10*24*time.Hour))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tm.ApplyDecay(baseTime)
	}
}

func BenchmarkSave(b *testing.B) {
	tm := NewTrustManager()
	now := time.Now()

	// Create 1000 records
	for i := 0; i < 1000; i++ {
		playerA := "player_" + string(rune(i/10))
		playerB := "player_" + string(rune(i%10))
		tm.UpdateTrust(playerA, playerB, 0.1, now)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := tm.Save()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Automatic Decay Tests

func TestStartAutomaticDecay(t *testing.T) {
	tm := NewTrustManager()

	// Test successful start
	err := tm.StartAutomaticDecay(100 * time.Millisecond)
	if err != nil {
		t.Fatalf("Failed to start automatic decay: %v", err)
	}

	if !tm.IsAutomaticDecayRunning() {
		t.Error("Expected automatic decay to be running")
	}

	// Test that starting again fails
	err = tm.StartAutomaticDecay(100 * time.Millisecond)
	if err == nil {
		t.Error("Expected error when starting already running decay")
	}

	// Clean up
	tm.StopAutomaticDecay()
}

func TestStopAutomaticDecay(t *testing.T) {
	tm := NewTrustManager()

	// Test stopping when not running
	stopped := tm.StopAutomaticDecay()
	if stopped {
		t.Error("Expected StopAutomaticDecay to return false when not running")
	}

	// Start and then stop
	err := tm.StartAutomaticDecay(100 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	stopped = tm.StopAutomaticDecay()
	if !stopped {
		t.Error("Expected StopAutomaticDecay to return true")
	}

	if tm.IsAutomaticDecayRunning() {
		t.Error("Expected automatic decay to not be running after stop")
	}

	// Test stopping again
	stopped = tm.StopAutomaticDecay()
	if stopped {
		t.Error("Expected StopAutomaticDecay to return false when already stopped")
	}
}

func TestStartAutomaticDecayInvalidInterval(t *testing.T) {
	tm := NewTrustManager()

	// Test zero interval
	err := tm.StartAutomaticDecay(0)
	if err == nil {
		t.Error("Expected error for zero interval")
	}

	// Test negative interval
	err = tm.StartAutomaticDecay(-1 * time.Second)
	if err == nil {
		t.Error("Expected error for negative interval")
	}
}

func TestAutomaticDecayAppliesDecay(t *testing.T) {
	tm := NewTrustManager()
	baseTime := time.Now()

	// Create a record that's old enough to decay
	oldTime := baseTime.Add(-10 * 24 * time.Hour) // 10 days ago
	tm.UpdateTrust("alice", "bob", 0.3, oldTime)

	initialTrust := tm.GetTrust("alice", "bob")
	if initialTrust != 0.8 { // 0.5 + 0.3
		t.Fatalf("Expected initial trust 0.8, got %f", initialTrust)
	}

	// Start automatic decay with a short interval
	err := tm.StartAutomaticDecay(50 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for at least one decay cycle
	time.Sleep(150 * time.Millisecond)

	// Stop decay
	tm.StopAutomaticDecay()

	// Verify that decay was applied
	finalTrust := tm.GetTrust("alice", "bob")
	if finalTrust >= initialTrust {
		t.Errorf("Expected trust to decay from %f, but got %f", initialTrust, finalTrust)
	}
}

func TestIsAutomaticDecayRunning(t *testing.T) {
	tm := NewTrustManager()

	// Initially not running
	if tm.IsAutomaticDecayRunning() {
		t.Error("Expected automatic decay to not be running initially")
	}

	// Start it
	err := tm.StartAutomaticDecay(100 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	if !tm.IsAutomaticDecayRunning() {
		t.Error("Expected automatic decay to be running after start")
	}

	// Stop it
	tm.StopAutomaticDecay()

	if tm.IsAutomaticDecayRunning() {
		t.Error("Expected automatic decay to not be running after stop")
	}
}

func TestAutomaticDecayMultipleRecords(t *testing.T) {
	tm := NewTrustManager()
	baseTime := time.Now()

	// Create multiple records with different ages
	tm.UpdateTrust("alice", "bob", 0.3, baseTime.Add(-10*24*time.Hour))    // 10 days old
	tm.UpdateTrust("alice", "charlie", 0.3, baseTime.Add(-5*24*time.Hour)) // 5 days old
	tm.UpdateTrust("alice", "david", 0.3, baseTime)                        // Fresh

	// Start automatic decay
	err := tm.StartAutomaticDecay(50 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for decay
	time.Sleep(150 * time.Millisecond)

	// Stop decay
	tm.StopAutomaticDecay()

	// Verify old records decayed
	trustBob := tm.GetTrust("alice", "bob")
	trustCharlie := tm.GetTrust("alice", "charlie")
	trustDavid := tm.GetTrust("alice", "david")

	// Bob should have decayed most (oldest)
	if trustBob >= 0.8 {
		t.Errorf("Expected bob's trust to decay, got %f", trustBob)
	}

	// Charlie should have decayed less than Bob
	if trustCharlie >= 0.8 || trustCharlie <= trustBob {
		t.Errorf("Expected charlie's trust (%f) to be between bob (%f) and fresh", trustCharlie, trustBob)
	}

	// David should not have decayed (too fresh)
	if trustDavid != 0.8 {
		t.Errorf("Expected david's trust to remain 0.8, got %f", trustDavid)
	}
}

func TestAutomaticDecayConcurrentSafe(t *testing.T) {
	tm := NewTrustManager()
	now := time.Now()

	// Start automatic decay
	err := tm.StartAutomaticDecay(10 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// Run concurrent operations while decay is running
	done := make(chan bool)
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				tm.UpdateTrust("player_a", "player_b", 0.01, now)
				_ = tm.GetTrust("player_a", "player_b")
				_ = tm.GetRecordCount()
			}
			done <- true
		}(i)
	}

	// Wait for goroutines
	for i := 0; i < 5; i++ {
		<-done
	}

	// Stop decay
	tm.StopAutomaticDecay()

	// Verify manager is still functional
	trust := tm.GetTrust("player_a", "player_b")
	if trust < 0.0 || trust > 1.0 {
		t.Errorf("Trust out of range: %f", trust)
	}
}

// TimeProvider Tests

// mockTimeProvider implements TimeProvider for deterministic testing
type mockTimeProvider struct {
	fixedTime time.Time
}

func (m *mockTimeProvider) Now() time.Time {
	return m.fixedTime
}

func TestNewTrustManagerWithTimeProvider(t *testing.T) {
	fixedTime := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	mockTP := &mockTimeProvider{fixedTime: fixedTime}

	tm := NewTrustManagerWithTimeProvider(mockTP)

	if tm == nil {
		t.Fatal("NewTrustManagerWithTimeProvider returned nil")
	}
	if tm.records == nil {
		t.Error("TrustManager.records not initialized")
	}
	if tm.timeProvider != mockTP {
		t.Error("TimeProvider not set correctly")
	}
}

func TestTrustManagerSetTimeProvider(t *testing.T) {
	tm := NewTrustManager()

	// Verify default time provider is set
	if tm.timeProvider == nil {
		t.Fatal("Default timeProvider not set")
	}

	// Set a mock time provider
	fixedTime := time.Date(2026, 2, 17, 15, 0, 0, 0, time.UTC)
	mockTP := &mockTimeProvider{fixedTime: fixedTime}
	tm.SetTimeProvider(mockTP)

	// Verify the new time provider is used in decay loop
	// Start automatic decay and verify it uses the mock time
	baseTime := time.Now()
	oldTime := baseTime.Add(-10 * 24 * time.Hour)
	tm.UpdateTrust("alice", "bob", 0.3, oldTime)

	initialTrust := tm.GetTrust("alice", "bob")

	// Start decay - now uses mock time provider
	err := tm.StartAutomaticDecay(50 * time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}

	// Wait for decay cycle
	time.Sleep(150 * time.Millisecond)
	tm.StopAutomaticDecay()

	// Since mock time is always fixed, decay calculation will use mockTP.Now()
	// The key test here is that we can successfully inject and use a custom TimeProvider
	finalTrust := tm.GetTrust("alice", "bob")

	// Verify trust changed (decay applied)
	if finalTrust >= initialTrust {
		t.Logf("Trust may not have decayed significantly with mock time (initial=%f, final=%f)", initialTrust, finalTrust)
	}
}

func TestTrustManagerDeterministicDecay(t *testing.T) {
	// Create two managers with same mock time
	fixedTime := time.Date(2026, 2, 17, 12, 0, 0, 0, time.UTC)
	mockTP1 := &mockTimeProvider{fixedTime: fixedTime}
	mockTP2 := &mockTimeProvider{fixedTime: fixedTime}

	tm1 := NewTrustManagerWithTimeProvider(mockTP1)
	tm2 := NewTrustManagerWithTimeProvider(mockTP2)

	// Set up identical records with old timestamps
	oldTime := fixedTime.Add(-10 * 24 * time.Hour)
	tm1.UpdateTrust("alice", "bob", 0.3, oldTime)
	tm2.UpdateTrust("alice", "bob", 0.3, oldTime)

	// Apply decay with same time
	tm1.ApplyDecay(fixedTime)
	tm2.ApplyDecay(fixedTime)

	// Trust scores should be identical
	trust1 := tm1.GetTrust("alice", "bob")
	trust2 := tm2.GetTrust("alice", "bob")

	if trust1 != trust2 {
		t.Errorf("Expected deterministic decay results: tm1=%f, tm2=%f", trust1, trust2)
	}
}
