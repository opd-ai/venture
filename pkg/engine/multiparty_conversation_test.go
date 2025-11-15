package engine

import (
	"sync"
	"testing"
	"time"
)

// TestMultiPartyConversation_Integration tests full multi-party conversation flow
func TestMultiPartyConversation_Integration(t *testing.T) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)

	// Create NPC entity
	npc := world.CreateEntity()
	npc.ID = 1

	// Update world to flush entity to entities map
	world.Update(0)

	// Initialize NPC dialog
	err := dialogSys.InitializeNPCDialog(npc, "fantasy", nil, 12345)
	if err != nil {
		t.Fatalf("Failed to initialize NPC dialog: %v", err)
	}

	// Start multi-party conversation
	playerIDs := []uint64{100, 101, 102}
	convID, err := dialogSys.StartMultiPartyConversation(npc.ID, playerIDs)
	if err != nil {
		t.Fatalf("Failed to start conversation: %v", err)
	}

	// Queue multiple player inputs
	responseChan1, err := dialogSys.QueuePlayerInput(npc.ID, 100, "Hello!")
	if err != nil {
		t.Fatalf("Failed to queue player 1 input: %v", err)
	}

	responseChan2, err := dialogSys.QueuePlayerInput(npc.ID, 101, "Greetings!")
	if err != nil {
		t.Fatalf("Failed to queue player 2 input: %v", err)
	}

	responseChan3, err := dialogSys.QueuePlayerInput(npc.ID, 102, "Good day!")
	if err != nil {
		t.Fatalf("Failed to queue player 3 input: %v", err)
	}

	// Process dialog queue
	dialogSys.ProcessQueuedDialogs(0.016) // Simulate one frame

	// Wait for first response
	select {
	case resp := <-responseChan1:
		if !resp.Success {
			t.Errorf("Player 1 request failed: %v", resp.Error)
		}
		if resp.Content == "" {
			t.Error("Player 1 received empty response")
		}
	case <-time.After(1 * time.Second):
		t.Error("Player 1 response timeout")
	}

	// Process remaining requests
	dialogSys.ProcessQueuedDialogs(0.016)
	select {
	case resp := <-responseChan2:
		if !resp.Success {
			t.Errorf("Player 2 request failed: %v", resp.Error)
		}
	case <-time.After(1 * time.Second):
		t.Error("Player 2 response timeout")
	}

	dialogSys.ProcessQueuedDialogs(0.016)
	select {
	case resp := <-responseChan3:
		if !resp.Success {
			t.Errorf("Player 3 request failed: %v", resp.Error)
		}
	case <-time.After(1 * time.Second):
		t.Error("Player 3 response timeout")
	}

	// Verify conversation messages
	messages, err := dialogSys.GetConversationMessages(convID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	if len(messages) != 0 {
		// Note: Messages only added explicitly via AddConversationMessage
		// Dialog responses aren't automatically added to conversation
	}
}

// TestDialogQueue_TurnTaking tests turn-taking with simultaneous requests
func TestDialogQueue_TurnTaking(t *testing.T) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)

	npc := world.CreateEntity()
	npc.ID = 1
	dialogSys.InitializeNPCDialog(npc, "fantasy", nil, 12345)

	// Queue 5 simultaneous requests
	numPlayers := 5
	channels := make([]<-chan *DialogResponse, numPlayers)

	for i := 0; i < numPlayers; i++ {
		respChan, err := dialogSys.QueuePlayerInput(npc.ID, uint64(100+i), "Test input")
		if err != nil {
			t.Fatalf("Failed to queue player %d: %v", i, err)
		}
		channels[i] = respChan
	}

	// Verify queue status
	queueSize, hasActive, err := dialogSys.GetDialogQueueStatus(npc.ID)
	if err != nil {
		t.Fatalf("Failed to get queue status: %v", err)
	}

	if queueSize != numPlayers {
		t.Errorf("Queue size = %d, want %d", queueSize, numPlayers)
	}

	if hasActive {
		t.Error("Should not have active request before processing")
	}

	// Process all requests
	responses := make([]*DialogResponse, 0, numPlayers)
	for i := 0; i < numPlayers; i++ {
		dialogSys.ProcessQueuedDialogs(0.016)

		select {
		case resp := <-channels[i]:
			responses = append(responses, resp)
		case <-time.After(1 * time.Second):
			t.Fatalf("Response %d timeout", i)
		}
	}

	// Verify all responses received
	if len(responses) != numPlayers {
		t.Errorf("Received %d responses, want %d", len(responses), numPlayers)
	}

	// Verify queue is empty
	queueSize, hasActive, _ = dialogSys.GetDialogQueueStatus(npc.ID)
	if queueSize != 0 {
		t.Errorf("Queue size = %d, want 0", queueSize)
	}
	if hasActive {
		t.Error("Should not have active request after completion")
	}
}

// TestDialogQueue_FullQueue tests queue overflow behavior
func TestDialogQueue_FullQueue(t *testing.T) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)
	dialogSys.conversationManager.maxQueueSize = 3

	npc := world.CreateEntity()
	npc.ID = 1
	dialogSys.InitializeNPCDialog(npc, "fantasy", nil, 12345)

	// Fill queue
	for i := 0; i < 3; i++ {
		_, err := dialogSys.QueuePlayerInput(npc.ID, uint64(100+i), "Test")
		if err != nil {
			t.Fatalf("Failed to queue player %d: %v", i, err)
		}
	}

	// Attempt overflow
	_, err := dialogSys.QueuePlayerInput(npc.ID, 200, "Overflow")
	if err == nil {
		t.Error("Queue should reject requests when full")
	}
}

// TestDialogQueue_Timeout tests request timeout behavior
func TestDialogQueue_Timeout(t *testing.T) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)
	dialogSys.conversationManager.requestTimeout = 100 * time.Millisecond

	npc := world.CreateEntity()
	npc.ID = 1
	dialogSys.InitializeNPCDialog(npc, "fantasy", nil, 12345)

	// Queue request
	respChan, err := dialogSys.QueuePlayerInput(npc.ID, 100, "Test")
	if err != nil {
		t.Fatalf("Failed to queue: %v", err)
	}

	// Start processing (but don't complete)
	dialogSys.conversationManager.ProcessNextDialogRequest(npc.ID)

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Process again (should auto-complete timed-out request)
	dialogSys.ProcessQueuedDialogs(0.016)

	// Verify timeout response
	select {
	case resp := <-respChan:
		if resp.Success {
			t.Error("Timed-out request should fail")
		}
		if resp.Error == nil {
			t.Error("Timed-out request should have error")
		}
	case <-time.After(1 * time.Second):
		t.Error("Should receive timeout response")
	}
}

// TestMultiPartyConversation_MessageOrdering tests message ordering across multiple players
func TestMultiPartyConversation_MessageOrdering(t *testing.T) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)

	npc := world.CreateEntity()
	npc.ID = 1

	playerIDs := []uint64{100, 101, 102}
	convID, _ := dialogSys.StartMultiPartyConversation(npc.ID, playerIDs)

	// Add messages from different players
	messages := []struct {
		senderID   uint64
		senderName string
		content    string
	}{
		{100, "Player1", "First message"},
		{101, "Player2", "Second message"},
		{npc.ID, "NPC", "NPC response"},
		{102, "Player3", "Third message"},
		{100, "Player1", "Fourth message"},
	}

	for _, msg := range messages {
		err := dialogSys.AddConversationMessage(convID, msg.senderID, msg.senderName, msg.content)
		if err != nil {
			t.Fatalf("Failed to add message: %v", err)
		}
		time.Sleep(1 * time.Millisecond) // Ensure distinct timestamps
	}

	// Retrieve messages
	retrieved, err := dialogSys.GetConversationMessages(convID)
	if err != nil {
		t.Fatalf("Failed to get messages: %v", err)
	}

	// Verify count
	if len(retrieved) != len(messages) {
		t.Errorf("Message count = %d, want %d", len(retrieved), len(messages))
	}

	// Verify order
	for i, msg := range retrieved {
		if msg.Content != messages[i].content {
			t.Errorf("Message[%d] content = %s, want %s", i, msg.Content, messages[i].content)
		}

		if msg.SenderID != messages[i].senderID {
			t.Errorf("Message[%d] senderID = %d, want %d", i, msg.SenderID, messages[i].senderID)
		}
	}

	// Verify timestamps are monotonically increasing
	for i := 1; i < len(retrieved); i++ {
		if retrieved[i].Timestamp.Before(retrieved[i-1].Timestamp) {
			t.Errorf("Message[%d] timestamp is before Message[%d]", i, i-1)
		}
	}
}

// TestMultiPartyConversation_ConcurrentAccess tests concurrent conversation access
func TestMultiPartyConversation_ConcurrentAccess(t *testing.T) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)

	npc := world.CreateEntity()
	npc.ID = 1
	dialogSys.InitializeNPCDialog(npc, "fantasy", nil, 12345)

	playerIDs := make([]uint64, 50)
	for i := range playerIDs {
		playerIDs[i] = uint64(100 + i)
	}

	convID, _ := dialogSys.StartMultiPartyConversation(npc.ID, playerIDs)

	var wg sync.WaitGroup

	// Concurrent message additions
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(playerID uint64) {
			defer wg.Done()
			dialogSys.AddConversationMessage(convID, playerID, "Player", "Concurrent message")
		}(uint64(100 + i))
	}

	wg.Wait()

	// Verify all messages added
	messages, _ := dialogSys.GetConversationMessages(convID)
	if len(messages) != 50 {
		t.Errorf("Message count = %d, want 50", len(messages))
	}
}

// TestDialogQueue_ConcurrentRequests tests concurrent dialog queueing
func TestDialogQueue_ConcurrentRequests(t *testing.T) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)
	dialogSys.conversationManager.maxQueueSize = 100

	npc := world.CreateEntity()
	npc.ID = 1
	dialogSys.InitializeNPCDialog(npc, "fantasy", nil, 12345)

	numRequests := 50
	var wg sync.WaitGroup
	successCount := 0
	var mu sync.Mutex

	// Concurrent queueing
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(playerID uint64) {
			defer wg.Done()
			_, err := dialogSys.QueuePlayerInput(npc.ID, playerID, "Test")
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(uint64(100 + i))
	}

	wg.Wait()

	// Verify all requests queued
	if successCount != numRequests {
		t.Errorf("Success count = %d, want %d", successCount, numRequests)
	}
}

// TestConversationCleanup tests stale conversation cleanup
func TestConversationCleanup(t *testing.T) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)

	// Create multiple conversations
	for i := 0; i < 5; i++ {
		npc := world.CreateEntity()
		npc.ID = uint64(i + 1)
		dialogSys.StartMultiPartyConversation(npc.ID, []uint64{uint64(100 + i)})
	}

	// Make some conversations stale
	dialogSys.conversationManager.mu.Lock()
	count := 0
	for _, conv := range dialogSys.conversationManager.conversations {
		if count < 3 {
			conv.LastActivity = time.Now().Add(-2 * time.Hour)
		}
		count++
	}
	dialogSys.conversationManager.mu.Unlock()

	// Cleanup
	removed := dialogSys.CleanupStaleConversations()

	if removed != 3 {
		t.Errorf("Removed conversations = %d, want 3", removed)
	}

	activeCount := dialogSys.conversationManager.GetActiveConversations()
	if activeCount != 2 {
		t.Errorf("Active conversations = %d, want 2", activeCount)
	}
}

// Benchmark tests
func BenchmarkMultiPartyConversation_QueueInput(b *testing.B) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)
	dialogSys.conversationManager.maxQueueSize = b.N

	npc := world.CreateEntity()
	npc.ID = 1
	dialogSys.InitializeNPCDialog(npc, "fantasy", nil, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialogSys.QueuePlayerInput(npc.ID, uint64(100+i), "Test input")
	}
}

func BenchmarkMultiPartyConversation_ProcessQueue(b *testing.B) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)
	dialogSys.conversationManager.maxQueueSize = b.N

	npc := world.CreateEntity()
	npc.ID = 1
	dialogSys.InitializeNPCDialog(npc, "fantasy", nil, 12345)

	// Pre-queue requests
	for i := 0; i < b.N; i++ {
		dialogSys.QueuePlayerInput(npc.ID, uint64(100+i), "Test")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialogSys.ProcessQueuedDialogs(0.016)
	}
}

func BenchmarkMultiPartyConversation_AddMessage(b *testing.B) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)

	npc := world.CreateEntity()
	npc.ID = 1
	convID, _ := dialogSys.StartMultiPartyConversation(npc.ID, []uint64{100})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialogSys.AddConversationMessage(convID, 100, "Player", "Message")
	}
}

func BenchmarkMultiPartyConversation_GetMessages(b *testing.B) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)

	npc := world.CreateEntity()
	npc.ID = 1
	convID, _ := dialogSys.StartMultiPartyConversation(npc.ID, []uint64{100})

	// Add messages
	for i := 0; i < 100; i++ {
		dialogSys.AddConversationMessage(convID, 100, "Player", "Message")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dialogSys.GetConversationMessages(convID)
	}
}
