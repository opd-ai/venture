package persistence

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

// TrustManager manages persistent trust scores between players
type TrustManager struct {
	mu      sync.RWMutex
	records map[string]*TrustRecord // key: "playerA:playerB" (sorted)

	// Automatic decay scheduling
	decayMu       sync.Mutex
	decayTicker   *time.Ticker
	decayStopChan chan struct{}
	decayRunning  bool
}

// NewTrustManager creates a new TrustManager
func NewTrustManager() *TrustManager {
	return &TrustManager{
		records: make(map[string]*TrustRecord),
	}
}

// makeKey creates a sorted key for two player IDs
func makeKey(playerA, playerB string) (string, string, string) {
	if playerA < playerB {
		return playerA, playerB, playerA + ":" + playerB
	}
	return playerB, playerA, playerB + ":" + playerA
}

// UpdateTrust modifies the trust score between two players
func (tm *TrustManager) UpdateTrust(playerA, playerB string, delta float64, timestamp time.Time) error {
	if playerA == playerB {
		return fmt.Errorf("cannot update trust with self")
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	sortedA, sortedB, key := makeKey(playerA, playerB)

	record, exists := tm.records[key]
	if !exists {
		record = &TrustRecord{
			PlayerA:      sortedA,
			PlayerB:      sortedB,
			Score:        0.5, // Start at neutral
			LastUpdate:   timestamp,
			Interactions: 0,
		}
		tm.records[key] = record
	}

	record.Score += delta
	if record.Score < MinTrustScore {
		record.Score = MinTrustScore
	}
	if record.Score > MaxTrustScore {
		record.Score = MaxTrustScore
	}

	record.LastUpdate = timestamp
	if delta > 0 {
		record.Interactions++
	}

	return nil
}

// GetTrust returns the trust score between two players
func (tm *TrustManager) GetTrust(playerA, playerB string) float64 {
	if playerA == playerB {
		return MaxTrustScore
	}

	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, _, key := makeKey(playerA, playerB)
	record, exists := tm.records[key]
	if !exists {
		return 0.5 // Default neutral trust
	}

	return record.Score
}

// GetTrustLevel returns the trust tier between two players
func (tm *TrustManager) GetTrustLevel(playerA, playerB string) TrustLevel {
	score := tm.GetTrust(playerA, playerB)
	return GetTrustLevel(score)
}

// GetTrustRecord returns the complete trust record between two players
func (tm *TrustManager) GetTrustRecord(playerA, playerB string) *TrustRecord {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, _, key := makeKey(playerA, playerB)
	record, exists := tm.records[key]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modification
	copy := *record
	return &copy
}

// ApplyDecay reduces trust scores based on time since last update
func (tm *TrustManager) ApplyDecay(currentTime time.Time) int {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	decayed := 0
	for _, record := range tm.records {
		daysSinceUpdate := currentTime.Sub(record.LastUpdate).Hours() / 24.0
		if daysSinceUpdate > 1.0 {
			decay := DecayRatePerDay * daysSinceUpdate
			record.Score -= decay
			if record.Score < MinTrustScore {
				record.Score = MinTrustScore
			}
			decayed++
		}
	}

	return decayed
}

// GetPlayerTrustRecords returns all trust records for a specific player
func (tm *TrustManager) GetPlayerTrustRecords(playerID string) []*TrustRecord {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var records []*TrustRecord
	for _, record := range tm.records {
		if record.PlayerA == playerID || record.PlayerB == playerID {
			copy := *record
			records = append(records, &copy)
		}
	}

	return records
}

// GetRecordCount returns the total number of trust records
func (tm *TrustManager) GetRecordCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.records)
}

// Save serializes the trust manager to compressed JSON
func (tm *TrustManager) Save() ([]byte, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	// Convert map to slice for JSON serialization
	var records []*TrustRecord
	for _, record := range tm.records {
		records = append(records, record)
	}

	jsonData, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal trust records: %w", err)
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(jsonData); err != nil {
		return nil, fmt.Errorf("failed to compress trust data: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Load deserializes trust records from compressed JSON
func (tm *TrustManager) Load(data []byte) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	jsonData, err := io.ReadAll(gzipReader)
	if err != nil {
		return fmt.Errorf("failed to decompress trust data: %w", err)
	}

	var records []*TrustRecord
	if err := json.Unmarshal(jsonData, &records); err != nil {
		return fmt.Errorf("failed to unmarshal trust records: %w", err)
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.records = make(map[string]*TrustRecord)
	for _, record := range records {
		_, _, key := makeKey(record.PlayerA, record.PlayerB)
		tm.records[key] = record
	}

	return nil
}

// Clear removes all trust records
func (tm *TrustManager) Clear() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.records = make(map[string]*TrustRecord)
}

// StartAutomaticDecay begins automatic trust score decay in a background goroutine.
// The interval parameter controls how often decay is applied (e.g., 1*time.Hour).
// Call StopAutomaticDecay() to stop the background processing.
func (tm *TrustManager) StartAutomaticDecay(interval time.Duration) error {
	tm.decayMu.Lock()
	defer tm.decayMu.Unlock()

	if tm.decayRunning {
		return fmt.Errorf("automatic decay already running")
	}

	if interval <= 0 {
		return fmt.Errorf("interval must be positive")
	}

	tm.decayTicker = time.NewTicker(interval)
	tm.decayStopChan = make(chan struct{})
	tm.decayRunning = true

	go tm.runDecayLoop()
	return nil
}

// StopAutomaticDecay stops the background decay processing.
// Returns true if decay was running and was stopped, false if it was not running.
func (tm *TrustManager) StopAutomaticDecay() bool {
	tm.decayMu.Lock()
	defer tm.decayMu.Unlock()

	if !tm.decayRunning {
		return false
	}

	close(tm.decayStopChan)
	tm.decayTicker.Stop()
	tm.decayRunning = false
	return true
}

// IsAutomaticDecayRunning returns whether automatic decay is currently active.
func (tm *TrustManager) IsAutomaticDecayRunning() bool {
	tm.decayMu.Lock()
	defer tm.decayMu.Unlock()
	return tm.decayRunning
}

// runDecayLoop is the background goroutine that applies decay at regular intervals.
func (tm *TrustManager) runDecayLoop() {
	for {
		select {
		case <-tm.decayStopChan:
			return
		case <-tm.decayTicker.C:
			tm.ApplyDecay(time.Now())
		}
	}
}
