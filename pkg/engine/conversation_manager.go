package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"
)

// generateMessageID generates a unique message ID (UUID v4 format).
func generateMessageID() string {
	uuid := make([]byte, 16)
	_, err := io.ReadFull(rand.Reader, uuid)
	if err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}

	// Set version (4) and variant bits per RFC 4122
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return hex.EncodeToString(uuid)
}

// ConversationManager manages multi-party conversations between NPCs and players.
// It handles message ordering, turn-taking, and conflict resolution for simultaneous
// dialog requests.
type ConversationManager struct {
	mu sync.RWMutex

	// conversations tracks active conversations by ID
	conversations map[string]*Conversation

	// npcQueues tracks pending dialog requests per NPC entity ID
	npcQueues map[uint64]*DialogQueue

	// maxQueueSize limits pending requests per NPC (default 5)
	maxQueueSize int

	// requestTimeout is max time before auto-completing active request (default 30s)
	requestTimeout time.Duration

	// world reference for entity lookups
	world *World
}

// Conversation represents a multi-party conversation.
type Conversation struct {
	ID             string   // Unique conversation ID (UUID)
	NPCID          uint64   // NPC entity ID
	ParticipantIDs []uint64 // Player entity IDs
	Messages       []ConversationMessage
	CreatedAt      time.Time
	LastActivity   time.Time
}

// ConversationMessage represents a message in a conversation.
type ConversationMessage struct {
	MessageID   string    // Unique message ID
	SenderID    uint64    // Entity ID of sender (NPC or player)
	SenderName  string    // Display name
	Content     string    // Message content
	Timestamp   time.Time // Server timestamp for ordering
	SequenceNum uint32    // Sequence number within conversation
}

// DialogQueue manages pending dialog requests for an NPC.
type DialogQueue struct {
	mu            sync.Mutex
	npcID         uint64
	queue         []*DialogRequest
	activeRequest *DialogRequest
	maxSize       int
}

// DialogRequest represents a queued dialog request.
type DialogRequest struct {
	RequestID    string
	PlayerID     uint64
	PlayerInput  string
	QueuedAt     time.Time
	TimeoutAt    time.Time
	ResponseChan chan *DialogResponse
}

// DialogResponse represents the result of a dialog request.
type DialogResponse struct {
	RequestID string
	Content   string
	Success   bool
	Error     error
}

// NewConversationManager creates a new conversation manager.
func NewConversationManager(world *World) *ConversationManager {
	return &ConversationManager{
		conversations:  make(map[string]*Conversation),
		npcQueues:      make(map[uint64]*DialogQueue),
		maxQueueSize:   5,
		requestTimeout: 30 * time.Second,
		world:          world,
	}
}

// StartConversation creates or retrieves a conversation between an NPC and players.
func (cm *ConversationManager) StartConversation(npcID uint64, playerIDs []uint64) (*Conversation, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Generate conversation ID
	convID := fmt.Sprintf("conv-%d-%d", npcID, time.Now().UnixNano())

	// Create conversation
	conv := &Conversation{
		ID:             convID,
		NPCID:          npcID,
		ParticipantIDs: playerIDs,
		Messages:       make([]ConversationMessage, 0, 100),
		CreatedAt:      time.Now(),
		LastActivity:   time.Now(),
	}

	cm.conversations[convID] = conv

	// Initialize NPC queue if not exists
	if _, exists := cm.npcQueues[npcID]; !exists {
		cm.npcQueues[npcID] = &DialogQueue{
			npcID:   npcID,
			queue:   make([]*DialogRequest, 0, cm.maxQueueSize),
			maxSize: cm.maxQueueSize,
		}
	}

	return conv, nil
}

// AddMessage adds a message to a conversation with timestamp-based ordering.
func (cm *ConversationManager) AddMessage(convID string, senderID uint64, senderName string, content string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	conv, exists := cm.conversations[convID]
	if !exists {
		return fmt.Errorf("conversation %s not found", convID)
	}

	// Create message
	msg := ConversationMessage{
		MessageID:   generateMessageID(),
		SenderID:    senderID,
		SenderName:  senderName,
		Content:     content,
		Timestamp:   time.Now(),
		SequenceNum: uint32(len(conv.Messages)),
	}

	conv.Messages = append(conv.Messages, msg)
	conv.LastActivity = time.Now()

	return nil
}

// GetConversation retrieves a conversation by ID.
func (cm *ConversationManager) GetConversation(convID string) (*Conversation, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conv, exists := cm.conversations[convID]
	if !exists {
		return nil, fmt.Errorf("conversation %s not found", convID)
	}

	return conv, nil
}

// GetConversationMessages returns messages in timestamp order.
func (cm *ConversationManager) GetConversationMessages(convID string) ([]ConversationMessage, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	conv, exists := cm.conversations[convID]
	if !exists {
		return nil, fmt.Errorf("conversation %s not found", convID)
	}

	// Messages are already ordered by sequence number (insertion order)
	// Additional timestamp-based sorting could be added here if needed

	return conv.Messages, nil
}

// QueueDialogRequest queues a dialog request for an NPC.
// Returns error if queue is full or NPC is not found.
func (cm *ConversationManager) QueueDialogRequest(npcID uint64, playerID uint64, playerInput string) (*DialogRequest, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Get or create NPC queue
	queue, exists := cm.npcQueues[npcID]
	if !exists {
		queue = &DialogQueue{
			npcID:   npcID,
			queue:   make([]*DialogRequest, 0, cm.maxQueueSize),
			maxSize: cm.maxQueueSize,
		}
		cm.npcQueues[npcID] = queue
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	// Check queue size
	if len(queue.queue) >= queue.maxSize {
		return nil, fmt.Errorf("NPC dialog queue full (max %d pending requests)", queue.maxSize)
	}

	// Create request
	req := &DialogRequest{
		RequestID:    fmt.Sprintf("req-%d-%d", npcID, time.Now().UnixNano()),
		PlayerID:     playerID,
		PlayerInput:  playerInput,
		QueuedAt:     time.Now(),
		TimeoutAt:    time.Now().Add(cm.requestTimeout),
		ResponseChan: make(chan *DialogResponse, 1),
	}

	// Add to queue
	queue.queue = append(queue.queue, req)

	return req, nil
}

// ProcessNextDialogRequest processes the next queued request for an NPC.
// Returns nil if queue is empty or another request is active.
func (cm *ConversationManager) ProcessNextDialogRequest(npcID uint64) (*DialogRequest, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	queue, exists := cm.npcQueues[npcID]
	if !exists {
		return nil, fmt.Errorf("no dialog queue for NPC %d", npcID)
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	// Check if request already active
	if queue.activeRequest != nil {
		// Check timeout
		if time.Now().After(queue.activeRequest.TimeoutAt) {
			// Auto-complete timed-out request
			queue.activeRequest.ResponseChan <- &DialogResponse{
				RequestID: queue.activeRequest.RequestID,
				Content:   "",
				Success:   false,
				Error:     fmt.Errorf("request timed out after %v", cm.requestTimeout),
			}
			close(queue.activeRequest.ResponseChan)
			queue.activeRequest = nil
		} else {
			return nil, nil // Request still active
		}
	}

	// Pop next request
	if len(queue.queue) == 0 {
		return nil, nil // Queue empty
	}

	req := queue.queue[0]
	queue.queue = queue.queue[1:]
	queue.activeRequest = req

	return req, nil
}

// CompleteDialogRequest marks an active request as complete and sends response.
func (cm *ConversationManager) CompleteDialogRequest(npcID uint64, requestID string, response string, err error) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	queue, exists := cm.npcQueues[npcID]
	if !exists {
		return fmt.Errorf("no dialog queue for NPC %d", npcID)
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	if queue.activeRequest == nil || queue.activeRequest.RequestID != requestID {
		return fmt.Errorf("request %s is not the active request", requestID)
	}

	// Send response
	queue.activeRequest.ResponseChan <- &DialogResponse{
		RequestID: requestID,
		Content:   response,
		Success:   err == nil,
		Error:     err,
	}
	close(queue.activeRequest.ResponseChan)

	// Clear active request
	queue.activeRequest = nil

	return nil
}

// GetDialogQueueStatus returns the queue status for an NPC.
func (cm *ConversationManager) GetDialogQueueStatus(npcID uint64) (queueSize int, hasActive bool, err error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	queue, exists := cm.npcQueues[npcID]
	if !exists {
		return 0, false, fmt.Errorf("no dialog queue for NPC %d", npcID)
	}

	queue.mu.Lock()
	defer queue.mu.Unlock()

	return len(queue.queue), queue.activeRequest != nil, nil
}

// CleanupStaleConversations removes conversations with no activity in the last hour.
func (cm *ConversationManager) CleanupStaleConversations() int {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cutoff := time.Now().Add(-1 * time.Hour)
	removed := 0

	for id, conv := range cm.conversations {
		if conv.LastActivity.Before(cutoff) {
			delete(cm.conversations, id)
			removed++
		}
	}

	return removed
}

// GetActiveConversations returns the number of active conversations.
func (cm *ConversationManager) GetActiveConversations() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return len(cm.conversations)
}
