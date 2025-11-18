package persistence

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MaxMessagesPerPlayer is the maximum chat history size per player
const MaxMessagesPerPlayer = 1000

// MaxMessageAge is the maximum age for messages (30 days)
const MaxMessageAge = 30 * 24 * time.Hour

// Message represents a single chat message
type Message struct {
	ID        string    `json:"id"`
	Sender    string    `json:"sender"`
	Recipient string    `json:"recipient"`
	Channel   string    `json:"channel"` // "global", "guild", "whisper", "party"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
	Deleted   bool      `json:"deleted,omitempty"`
}

// ChatHistory manages persistent chat messages for a player
type ChatHistory struct {
	PlayerID string     `json:"player_id"`
	Messages []*Message `json:"messages"`
	Version  int        `json:"version"`
	mu       sync.RWMutex
}

// NewChatHistory creates a new chat history manager
func NewChatHistory(playerID string) *ChatHistory {
	return &ChatHistory{
		PlayerID: playerID,
		Messages: make([]*Message, 0, MaxMessagesPerPlayer),
		Version:  1,
	}
}

// AddMessage adds a new message to history
func (c *ChatHistory) AddMessage(msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message cannot be nil")
	}
	if msg.ID == "" {
		return fmt.Errorf("message ID cannot be empty")
	}
	if msg.Sender == "" {
		return fmt.Errorf("message sender cannot be empty")
	}
	if msg.Content == "" && !msg.Deleted {
		return fmt.Errorf("message content cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if message already exists (deduplicate)
	for _, existing := range c.Messages {
		if existing.ID == msg.ID {
			return nil // Already exists, silently ignore
		}
	}

	c.Messages = append(c.Messages, msg)
	c.Version++

	// Enforce max message limit (LRU)
	if len(c.Messages) > MaxMessagesPerPlayer {
		// Remove oldest messages
		excess := len(c.Messages) - MaxMessagesPerPlayer
		c.Messages = c.Messages[excess:]
	}

	return nil
}

// GetMessages returns all messages, optionally filtered
func (c *ChatHistory) GetMessages(filter *MessageFilter) []*Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if filter == nil {
		// Return all messages
		result := make([]*Message, len(c.Messages))
		copy(result, c.Messages)
		return result
	}

	result := make([]*Message, 0, len(c.Messages))
	for _, msg := range c.Messages {
		if filter.Matches(msg) {
			result = append(result, msg)
		}
	}
	return result
}

// DeleteOldMessages removes messages older than MaxMessageAge
func (c *ChatHistory) DeleteOldMessages(now time.Time) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	cutoff := now.Add(-MaxMessageAge)
	originalCount := len(c.Messages)

	// Filter out old messages
	kept := make([]*Message, 0, len(c.Messages))
	for _, msg := range c.Messages {
		if msg.Timestamp.After(cutoff) {
			kept = append(kept, msg)
		}
	}

	c.Messages = kept
	deleted := originalCount - len(kept)

	if deleted > 0 {
		c.Version++
	}

	return deleted
}

// Save serializes chat history with gzip compression
func (c *ChatHistory) Save() ([]byte, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Serialize to JSON
	data, err := json.Marshal(c)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal chat history: %w", err)
	}

	// Compress with gzip
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)
	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed to compress chat history: %w", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Load deserializes chat history from gzipped JSON
func (c *ChatHistory) Load(data []byte) error {
	// Decompress
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer reader.Close()

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(reader); err != nil {
		return fmt.Errorf("failed to decompress chat history: %w", err)
	}

	// Deserialize from JSON
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := json.Unmarshal(buf.Bytes(), c); err != nil {
		return fmt.Errorf("failed to unmarshal chat history: %w", err)
	}

	return nil
}

// GetDelta computes the delta between this history and a given version
func (c *ChatHistory) GetDelta(fromVersion int) []*Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if fromVersion >= c.Version {
		// Already up to date
		return nil
	}

	// For simplicity, we return all messages if the version gap is too large
	// In a production system, you'd maintain a changelog of added/deleted messages
	// For now, we return messages that are likely new based on timestamp estimation

	// If requesting from version 0, return all
	if fromVersion == 0 {
		delta := make([]*Message, len(c.Messages))
		copy(delta, c.Messages)
		return delta
	}

	// Estimate how many messages to return based on version difference
	versionDiff := c.Version - fromVersion
	if versionDiff <= 0 {
		return nil
	}

	// Return the last N messages where N = version difference
	// This is a simple heuristic - real delta would track individual changes
	startIdx := len(c.Messages) - versionDiff
	if startIdx < 0 {
		startIdx = 0
	}

	delta := make([]*Message, len(c.Messages[startIdx:]))
	copy(delta, c.Messages[startIdx:])
	return delta
}

// ApplyDelta merges delta messages into this history
func (c *ChatHistory) ApplyDelta(delta []*Message) error {
	if len(delta) == 0 {
		return nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Build a map of existing message IDs for deduplication
	existing := make(map[string]bool, len(c.Messages))
	for _, msg := range c.Messages {
		existing[msg.ID] = true
	}

	// Add new messages
	added := 0
	for _, msg := range delta {
		if !existing[msg.ID] {
			c.Messages = append(c.Messages, msg)
			existing[msg.ID] = true
			added++
		}
	}

	if added > 0 {
		// Sort messages by timestamp (assuming they arrive mostly in order)
		// For performance, we rely on messages being added chronologically
		// A production system would use a more sophisticated merge strategy

		// Enforce max message limit
		if len(c.Messages) > MaxMessagesPerPlayer {
			excess := len(c.Messages) - MaxMessagesPerPlayer
			c.Messages = c.Messages[excess:]
		}

		c.Version++
	}

	return nil
}

// GetVersion returns the current version number
func (c *ChatHistory) GetVersion() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Version
}

// MessageFilter defines criteria for filtering messages
type MessageFilter struct {
	Sender    string
	Recipient string
	Channel   string
	After     time.Time
	Before    time.Time
}

// Matches returns true if a message matches the filter criteria
func (f *MessageFilter) Matches(msg *Message) bool {
	if f.Sender != "" && msg.Sender != f.Sender {
		return false
	}
	if f.Recipient != "" && msg.Recipient != f.Recipient {
		return false
	}
	if f.Channel != "" && msg.Channel != f.Channel {
		return false
	}
	if !f.After.IsZero() && msg.Timestamp.Before(f.After) {
		return false
	}
	if !f.Before.IsZero() && msg.Timestamp.After(f.Before) {
		return false
	}
	return true
}
