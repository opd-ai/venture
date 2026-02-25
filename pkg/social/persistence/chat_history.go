package persistence

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

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

// ChangeType represents the type of change in the changelog
type ChangeType string

const (
	// ChangeTypeAdd indicates a message was added
	ChangeTypeAdd ChangeType = "add"
	// ChangeTypeDelete indicates a message was deleted
	ChangeTypeDelete ChangeType = "delete"
)

// ChangelogEntry records a single change to the chat history
type ChangelogEntry struct {
	Type      ChangeType `json:"type"`
	MessageID string     `json:"message_id"`
	Version   int        `json:"version"`
}

// ChatHistory manages persistent chat messages for a player
type ChatHistory struct {
	PlayerID     string            `json:"player_id"`
	Messages     []*Message        `json:"messages"`
	Changelog    []*ChangelogEntry `json:"changelog"`
	Version      int               `json:"version"`
	mu           sync.RWMutex      `json:"-"`
	timeProvider TimeProvider      `json:"-"` // TimeProvider for deterministic timestamps
}

// NewChatHistory creates a new chat history manager using real system time.
func NewChatHistory(playerID string) *ChatHistory {
	return NewChatHistoryWithTimeProvider(playerID, DefaultTimeProvider())
}

// NewChatHistoryWithTimeProvider creates a new chat history manager with a custom TimeProvider.
// Use this constructor in tests to inject a mock TimeProvider for deterministic timestamps.
func NewChatHistoryWithTimeProvider(playerID string, tp TimeProvider) *ChatHistory {
	return &ChatHistory{
		PlayerID:     playerID,
		Messages:     make([]*Message, 0, MaxMessagesPerPlayer),
		Changelog:    make([]*ChangelogEntry, 0, MaxChangelogSize),
		Version:      1,
		timeProvider: tp,
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

	// Record change in changelog
	c.appendToChangelog(&ChangelogEntry{
		Type:      ChangeTypeAdd,
		MessageID: msg.ID,
		Version:   c.Version,
	})

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

	// Filter out old messages and record deletions in changelog
	kept := make([]*Message, 0, len(c.Messages))
	var deletedIDs []string
	for _, msg := range c.Messages {
		if msg.Timestamp.After(cutoff) {
			kept = append(kept, msg)
		} else {
			deletedIDs = append(deletedIDs, msg.ID)
		}
	}

	c.Messages = kept
	deleted := originalCount - len(kept)

	if deleted > 0 {
		c.Version++
		// Record deletions in changelog
		for _, msgID := range deletedIDs {
			c.appendToChangelog(&ChangelogEntry{
				Type:      ChangeTypeDelete,
				MessageID: msgID,
				Version:   c.Version,
			})
		}
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

	// Preserve existing timeProvider before unmarshal overwrites
	existingTP := c.timeProvider

	if err := json.Unmarshal(buf.Bytes(), c); err != nil {
		return fmt.Errorf("failed to unmarshal chat history: %w", err)
	}

	// Restore or initialize timeProvider (not serialized)
	if existingTP != nil {
		c.timeProvider = existingTP
	} else {
		c.timeProvider = DefaultTimeProvider()
	}

	return nil
}

// appendToChangelog adds an entry to the changelog, maintaining the circular buffer limit.
// Must be called with c.mu held (Lock, not RLock).
func (c *ChatHistory) appendToChangelog(entry *ChangelogEntry) {
	c.Changelog = append(c.Changelog, entry)

	// Trim changelog to MaxChangelogSize (circular buffer)
	if len(c.Changelog) > MaxChangelogSize {
		excess := len(c.Changelog) - MaxChangelogSize
		c.Changelog = c.Changelog[excess:]
	}
}

// GetDelta computes the delta between this history and a given version.
//
// This method uses a changelog-based approach to accurately track message additions
// and deletions since fromVersion. Unlike the previous heuristic-based approach,
// this implementation maintains an ordered log of changes (up to MaxChangelogSize
// entries) and queries it to determine exactly which messages have been added or
// deleted since the requested version.
//
// For version 0, all messages are returned (full sync). If fromVersion is older
// than the oldest entry in the changelog, all messages are returned as a fallback.
//
// The caller should use [ApplyDelta] to merge results, which deduplicates by
// message ID to handle any edge cases gracefully.
func (c *ChatHistory) GetDelta(fromVersion int) []*Message {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if fromVersion >= c.Version {
		// Already up to date
		return nil
	}

	// If requesting from version 0, return all messages (full sync)
	if fromVersion == 0 {
		delta := make([]*Message, len(c.Messages))
		copy(delta, c.Messages)
		return delta
	}

	// Find the oldest version in the changelog
	oldestVersion := c.Version
	if len(c.Changelog) > 0 {
		oldestVersion = c.Changelog[0].Version
	}

	// If fromVersion is older than our changelog, fall back to full sync
	if fromVersion < oldestVersion {
		delta := make([]*Message, len(c.Messages))
		copy(delta, c.Messages)
		return delta
	}

	// Collect message IDs from changelog entries after fromVersion
	addedIDs := make(map[string]bool)
	deletedIDs := make(map[string]bool)

	for _, entry := range c.Changelog {
		if entry.Version > fromVersion {
			switch entry.Type {
			case ChangeTypeAdd:
				addedIDs[entry.MessageID] = true
				// If a message was deleted and then re-added, remove from deleted set
				delete(deletedIDs, entry.MessageID)
			case ChangeTypeDelete:
				deletedIDs[entry.MessageID] = true
				// If a message was added and then deleted in this range, remove from added set
				delete(addedIDs, entry.MessageID)
			}
		}
	}

	// Build delta from messages that were added and not subsequently deleted
	delta := make([]*Message, 0, len(addedIDs))
	for _, msg := range c.Messages {
		if addedIDs[msg.ID] && !deletedIDs[msg.ID] {
			delta = append(delta, msg)
		}
	}

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

// SetTimeProvider sets the time provider for the chat history.
// This is useful when loading a history from JSON and needing to inject a mock time provider.
func (c *ChatHistory) SetTimeProvider(tp TimeProvider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.timeProvider = tp
}
