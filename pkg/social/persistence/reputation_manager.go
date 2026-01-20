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

// ReputationScore tracks a player's reputation in a specific category
type ReputationScore struct {
	Category   ReputationCategory
	Score      float64
	LastUpdate time.Time
}

// ReputationRecord stores all reputation data for a player
type ReputationRecord struct {
	PlayerID   string
	Scores     map[ReputationCategory]*ReputationScore
	TotalScore float64
	LastUpdate time.Time
}

// ReputationManager manages persistent reputation scores for players
type ReputationManager struct {
	mu      sync.RWMutex
	records map[string]*ReputationRecord // key: playerID
}

// NewReputationManager creates a new ReputationManager
func NewReputationManager() *ReputationManager {
	return &ReputationManager{
		records: make(map[string]*ReputationRecord),
	}
}

// UpdateReputation modifies a player's reputation in a specific category
func (rm *ReputationManager) UpdateReputation(playerID string, category ReputationCategory, delta float64, timestamp time.Time) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	record, exists := rm.records[playerID]
	if !exists {
		record = &ReputationRecord{
			PlayerID:   playerID,
			Scores:     make(map[ReputationCategory]*ReputationScore),
			TotalScore: 0.0,
			LastUpdate: timestamp,
		}
		rm.records[playerID] = record
	}

	score, exists := record.Scores[category]
	if !exists {
		score = &ReputationScore{
			Category:   category,
			Score:      0.0,
			LastUpdate: timestamp,
		}
		record.Scores[category] = score
	}

	score.Score += delta
	score.LastUpdate = timestamp
	record.LastUpdate = timestamp

	// Recalculate total score (weighted average)
	total := 0.0
	for _, s := range record.Scores {
		total += s.Score
	}
	record.TotalScore = total / float64(len(record.Scores))

	return nil
}

// GetReputation returns the reputation score for a player in a specific category
func (rm *ReputationManager) GetReputation(playerID string, category ReputationCategory) float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	record, exists := rm.records[playerID]
	if !exists {
		return 0.0
	}

	score, exists := record.Scores[category]
	if !exists {
		return 0.0
	}

	return score.Score
}

// GetTotalReputation returns the overall reputation score for a player
func (rm *ReputationManager) GetTotalReputation(playerID string) float64 {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	record, exists := rm.records[playerID]
	if !exists {
		return 0.0
	}

	return record.TotalScore
}

// GetReputationRecord returns the complete reputation record for a player
func (rm *ReputationManager) GetReputationRecord(playerID string) *ReputationRecord {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	record, exists := rm.records[playerID]
	if !exists {
		return nil
	}

	// Return a deep copy
	copy := &ReputationRecord{
		PlayerID:   record.PlayerID,
		Scores:     make(map[ReputationCategory]*ReputationScore),
		TotalScore: record.TotalScore,
		LastUpdate: record.LastUpdate,
	}
	for cat, score := range record.Scores {
		scoreCopy := *score
		copy.Scores[cat] = &scoreCopy
	}

	return copy
}

// ApplyDecay reduces reputation scores based on time since last update
func (rm *ReputationManager) ApplyDecay(currentTime time.Time) int {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	decayed := 0
	for _, record := range rm.records {
		for _, score := range record.Scores {
			daysSinceUpdate := currentTime.Sub(score.LastUpdate).Hours() / 24.0
			if daysSinceUpdate > 1.0 {
				decay := DecayRatePerDay * daysSinceUpdate
				score.Score -= decay
				if score.Score < 0.0 {
					score.Score = 0.0
				}
				decayed++
			}
		}

		// Recalculate total
		total := 0.0
		count := 0
		for _, s := range record.Scores {
			total += s.Score
			count++
		}
		if count > 0 {
			record.TotalScore = total / float64(count)
		}
	}

	return decayed
}

// GetRecordCount returns the total number of reputation records
func (rm *ReputationManager) GetRecordCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return len(rm.records)
}

// Save serializes the reputation manager to compressed JSON
func (rm *ReputationManager) Save() ([]byte, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var records []*ReputationRecord
	for _, record := range rm.records {
		records = append(records, record)
	}

	jsonData, err := json.Marshal(records)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal reputation records: %w", err)
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(jsonData); err != nil {
		return nil, fmt.Errorf("failed to compress reputation data: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Load deserializes reputation records from compressed JSON
func (rm *ReputationManager) Load(data []byte) error {
	gzipReader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	jsonData, err := io.ReadAll(gzipReader)
	if err != nil {
		return fmt.Errorf("failed to decompress reputation data: %w", err)
	}

	var records []*ReputationRecord
	if err := json.Unmarshal(jsonData, &records); err != nil {
		return fmt.Errorf("failed to unmarshal reputation records: %w", err)
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.records = make(map[string]*ReputationRecord)
	for _, record := range records {
		rm.records[record.PlayerID] = record
	}

	return nil
}

// Clear removes all reputation records
func (rm *ReputationManager) Clear() {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.records = make(map[string]*ReputationRecord)
}
