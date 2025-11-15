package engine

import (
	"testing"
	"time"
)

func TestConversationManager_StartConversation(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)

	tests := []struct {
		name      string
		npcID     uint64
		playerIDs []uint64
		wantErr   bool
	}{
		{
			name:      "single player conversation",
			npcID:     1,
			playerIDs: []uint64{100},
			wantErr:   false,
		},
		{
			name:      "multi-player conversation",
			npcID:     2,
			playerIDs: []uint64{100, 101, 102},
			wantErr:   false,
		},
		{
			name:      "empty player list",
			npcID:     3,
			playerIDs: []uint64{},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conv, err := cm.StartConversation(tt.npcID, tt.playerIDs)

			if (err != nil) != tt.wantErr {
				t.Errorf("StartConversation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if conv == nil {
					t.Error("StartConversation() returned nil conversation")
					return
				}

				if conv.NPCID != tt.npcID {
					t.Errorf("NPCID = %d, want %d", conv.NPCID, tt.npcID)
				}

				if len(conv.ParticipantIDs) != len(tt.playerIDs) {
					t.Errorf("ParticipantIDs length = %d, want %d", len(conv.ParticipantIDs), len(tt.playerIDs))
				}

				if conv.ID == "" {
					t.Error("Conversation ID is empty")
				}

				if conv.CreatedAt.IsZero() {
					t.Error("CreatedAt timestamp is zero")
				}
			}
		})
	}
}

func TestConversationManager_AddMessage(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)

	conv, _ := cm.StartConversation(1, []uint64{100})

	tests := []struct {
		name       string
		convID     string
		senderID   uint64
		senderName string
		content    string
		wantErr    bool
	}{
		{
			name:       "valid message",
			convID:     conv.ID,
			senderID:   100,
			senderName: "Player1",
			content:    "Hello NPC!",
			wantErr:    false,
		},
		{
			name:       "empty content",
			convID:     conv.ID,
			senderID:   100,
			senderName: "Player1",
			content:    "",
			wantErr:    false,
		},
		{
			name:       "invalid conversation ID",
			convID:     "nonexistent",
			senderID:   100,
			senderName: "Player1",
			content:    "Hello",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cm.AddMessage(tt.convID, tt.senderID, tt.senderName, tt.content)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConversationManager_GetConversation(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)

	conv, _ := cm.StartConversation(1, []uint64{100})

	tests := []struct {
		name    string
		convID  string
		wantErr bool
	}{
		{
			name:    "existing conversation",
			convID:  conv.ID,
			wantErr: false,
		},
		{
			name:    "nonexistent conversation",
			convID:  "nonexistent",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cm.GetConversation(tt.convID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetConversation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got == nil {
				t.Error("GetConversation() returned nil conversation")
			}
		})
	}
}

func TestConversationManager_MessageOrdering(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)

	conv, _ := cm.StartConversation(1, []uint64{100, 101})

	// Add messages in order
	messages := []string{"First", "Second", "Third", "Fourth", "Fifth"}
	for i, content := range messages {
		err := cm.AddMessage(conv.ID, uint64(100+i%2), "Player", content)
		if err != nil {
			t.Fatalf("AddMessage() failed: %v", err)
		}
		time.Sleep(1 * time.Millisecond) // Ensure distinct timestamps
	}

	// Retrieve messages
	retrieved, err := cm.GetConversationMessages(conv.ID)
	if err != nil {
		t.Fatalf("GetConversationMessages() failed: %v", err)
	}

	// Verify count
	if len(retrieved) != len(messages) {
		t.Errorf("Message count = %d, want %d", len(retrieved), len(messages))
	}

	// Verify order
	for i, msg := range retrieved {
		if msg.Content != messages[i] {
			t.Errorf("Message[%d] content = %s, want %s", i, msg.Content, messages[i])
		}

		if msg.SequenceNum != uint32(i) {
			t.Errorf("Message[%d] sequence = %d, want %d", i, msg.SequenceNum, i)
		}
	}

	// Verify timestamps are increasing
	for i := 1; i < len(retrieved); i++ {
		if retrieved[i].Timestamp.Before(retrieved[i-1].Timestamp) {
			t.Errorf("Message[%d] timestamp is before Message[%d]", i, i-1)
		}
	}
}

func TestConversationManager_QueueDialogRequest(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)

	tests := []struct {
		name        string
		npcID       uint64
		playerID    uint64
		playerInput string
		queueSize   int
		wantErr     bool
	}{
		{
			name:        "first request",
			npcID:       1,
			playerID:    100,
			playerInput: "Hello",
			queueSize:   0,
			wantErr:     false,
		},
		{
			name:        "second request",
			npcID:       1,
			playerID:    101,
			playerInput: "Hi there",
			queueSize:   1,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := cm.QueueDialogRequest(tt.npcID, tt.playerID, tt.playerInput)

			if (err != nil) != tt.wantErr {
				t.Errorf("QueueDialogRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if req == nil {
					t.Error("QueueDialogRequest() returned nil request")
					return
				}

				if req.PlayerID != tt.playerID {
					t.Errorf("PlayerID = %d, want %d", req.PlayerID, tt.playerID)
				}

				if req.PlayerInput != tt.playerInput {
					t.Errorf("PlayerInput = %s, want %s", req.PlayerInput, tt.playerInput)
				}

				if req.RequestID == "" {
					t.Error("RequestID is empty")
				}

				if req.ResponseChan == nil {
					t.Error("ResponseChan is nil")
				}
			}
		})
	}
}

func TestConversationManager_QueueDialogRequest_Full(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)
	cm.maxQueueSize = 3

	npcID := uint64(1)

	// Fill queue
	for i := 0; i < 3; i++ {
		_, err := cm.QueueDialogRequest(npcID, uint64(100+i), "Request")
		if err != nil {
			t.Fatalf("QueueDialogRequest() failed: %v", err)
		}
	}

	// Attempt to exceed queue size
	_, err := cm.QueueDialogRequest(npcID, 200, "Overflow")
	if err == nil {
		t.Error("QueueDialogRequest() should fail when queue is full")
	}
}

func TestConversationManager_ProcessNextDialogRequest(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)

	npcID := uint64(1)

	// Queue multiple requests
	req1, _ := cm.QueueDialogRequest(npcID, 100, "First")
	req2, _ := cm.QueueDialogRequest(npcID, 101, "Second")

	// Process first request
	processed, err := cm.ProcessNextDialogRequest(npcID)
	if err != nil {
		t.Fatalf("ProcessNextDialogRequest() failed: %v", err)
	}

	if processed == nil {
		t.Fatal("ProcessNextDialogRequest() returned nil")
	}

	if processed.RequestID != req1.RequestID {
		t.Errorf("Processed request ID = %s, want %s", processed.RequestID, req1.RequestID)
	}

	// Attempt to process while request active
	processed2, err := cm.ProcessNextDialogRequest(npcID)
	if err != nil {
		t.Fatalf("ProcessNextDialogRequest() failed: %v", err)
	}

	if processed2 != nil {
		t.Error("ProcessNextDialogRequest() should return nil when request is active")
	}

	// Complete first request
	err = cm.CompleteDialogRequest(npcID, req1.RequestID, "Response", nil)
	if err != nil {
		t.Fatalf("CompleteDialogRequest() failed: %v", err)
	}

	// Process second request
	processed3, err := cm.ProcessNextDialogRequest(npcID)
	if err != nil {
		t.Fatalf("ProcessNextDialogRequest() failed: %v", err)
	}

	if processed3 == nil {
		t.Fatal("ProcessNextDialogRequest() returned nil")
	}

	if processed3.RequestID != req2.RequestID {
		t.Errorf("Processed request ID = %s, want %s", processed3.RequestID, req2.RequestID)
	}
}

func TestConversationManager_CompleteDialogRequest(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)

	npcID := uint64(1)
	req, _ := cm.QueueDialogRequest(npcID, 100, "Test")
	cm.ProcessNextDialogRequest(npcID)

	tests := []struct {
		name      string
		npcID     uint64
		requestID string
		response  string
		err       error
		wantErr   bool
	}{
		{
			name:      "valid completion",
			npcID:     npcID,
			requestID: req.RequestID,
			response:  "NPC response",
			err:       nil,
			wantErr:   false,
		},
		{
			name:      "invalid request ID",
			npcID:     npcID,
			requestID: "invalid",
			response:  "Response",
			err:       nil,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cm.CompleteDialogRequest(tt.npcID, tt.requestID, tt.response, tt.err)

			if (err != nil) != tt.wantErr {
				t.Errorf("CompleteDialogRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConversationManager_DialogTimeout(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)
	cm.requestTimeout = 100 * time.Millisecond

	npcID := uint64(1)
	req, _ := cm.QueueDialogRequest(npcID, 100, "Test")

	// Process request
	processed, _ := cm.ProcessNextDialogRequest(npcID)
	if processed == nil {
		t.Fatal("ProcessNextDialogRequest() returned nil")
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Try to process next (should auto-complete timed-out request)
	processed2, err := cm.ProcessNextDialogRequest(npcID)
	if err != nil {
		t.Fatalf("ProcessNextDialogRequest() failed: %v", err)
	}

	// Should be nil (queue empty after timeout)
	if processed2 != nil {
		t.Error("ProcessNextDialogRequest() should return nil after timeout")
	}

	// Verify response channel received timeout error
	select {
	case resp := <-req.ResponseChan:
		if resp.Success {
			t.Error("Response should indicate failure")
		}
		if resp.Error == nil {
			t.Error("Response should have error set")
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("Response channel did not receive timeout response")
	}
}

func TestConversationManager_GetDialogQueueStatus(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)

	npcID := uint64(1)

	// Initially empty
	queueSize, hasActive, err := cm.GetDialogQueueStatus(npcID)
	if err == nil {
		t.Error("GetDialogQueueStatus() should fail for non-existent queue")
	}

	// Queue requests
	cm.QueueDialogRequest(npcID, 100, "First")
	cm.QueueDialogRequest(npcID, 101, "Second")

	queueSize, hasActive, err = cm.GetDialogQueueStatus(npcID)
	if err != nil {
		t.Fatalf("GetDialogQueueStatus() failed: %v", err)
	}

	if queueSize != 2 {
		t.Errorf("Queue size = %d, want 2", queueSize)
	}

	if hasActive {
		t.Error("Should not have active request yet")
	}

	// Process request
	cm.ProcessNextDialogRequest(npcID)

	queueSize, hasActive, err = cm.GetDialogQueueStatus(npcID)
	if err != nil {
		t.Fatalf("GetDialogQueueStatus() failed: %v", err)
	}

	if queueSize != 1 {
		t.Errorf("Queue size = %d, want 1", queueSize)
	}

	if !hasActive {
		t.Error("Should have active request")
	}
}

func TestConversationManager_CleanupStaleConversations(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)

	// Create conversations
	conv1, _ := cm.StartConversation(1, []uint64{100})
	conv2, _ := cm.StartConversation(2, []uint64{101})

	// Make conv1 stale
	cm.mu.Lock()
	cm.conversations[conv1.ID].LastActivity = time.Now().Add(-2 * time.Hour)
	cm.mu.Unlock()

	// Cleanup
	removed := cm.CleanupStaleConversations()

	if removed != 1 {
		t.Errorf("Removed conversations = %d, want 1", removed)
	}

	// Verify conv1 removed
	_, err := cm.GetConversation(conv1.ID)
	if err == nil {
		t.Error("Stale conversation should be removed")
	}

	// Verify conv2 exists
	_, err = cm.GetConversation(conv2.ID)
	if err != nil {
		t.Error("Active conversation should not be removed")
	}
}

func TestConversationManager_GetActiveConversations(t *testing.T) {
	world := NewWorld()
	cm := NewConversationManager(world)

	if count := cm.GetActiveConversations(); count != 0 {
		t.Errorf("Initial count = %d, want 0", count)
	}

	cm.StartConversation(1, []uint64{100})
	cm.StartConversation(2, []uint64{101})

	if count := cm.GetActiveConversations(); count != 2 {
		t.Errorf("Count after adding = %d, want 2", count)
	}
}

// Benchmark tests
func BenchmarkConversationManager_StartConversation(b *testing.B) {
	world := NewWorld()
	cm := NewConversationManager(world)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.StartConversation(uint64(i), []uint64{100, 101, 102})
	}
}

func BenchmarkConversationManager_AddMessage(b *testing.B) {
	world := NewWorld()
	cm := NewConversationManager(world)
	conv, _ := cm.StartConversation(1, []uint64{100})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.AddMessage(conv.ID, 100, "Player", "Test message")
	}
}

func BenchmarkConversationManager_QueueDialogRequest(b *testing.B) {
	world := NewWorld()
	cm := NewConversationManager(world)
	cm.maxQueueSize = 1000 // Large queue for benchmark

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cm.QueueDialogRequest(1, uint64(100+i), "Test input")
	}
}

func BenchmarkConversationManager_ProcessDialogRequest(b *testing.B) {
	world := NewWorld()
	cm := NewConversationManager(world)

	// Pre-queue requests
	for i := 0; i < b.N; i++ {
		cm.QueueDialogRequest(1, uint64(100+i), "Test")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req, _ := cm.ProcessNextDialogRequest(1)
		if req != nil {
			cm.CompleteDialogRequest(1, req.RequestID, "Response", nil)
		}
	}
}
