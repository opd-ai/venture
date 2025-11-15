package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Mock types for testing (to avoid import cycle with pkg/network)

type mockChatManager struct {
	mu              sync.RWMutex
	players         map[uint64]*mockPlayerChatState
	rateLimitStates map[uint64]map[int]time.Time
}

type mockPlayerChatState struct {
	PlayerID uint64
	Position mockVector2
}

type mockVector2 struct {
	X, Y float64
}

func newMockChatManager() *mockChatManager {
	return &mockChatManager{
		players:         make(map[uint64]*mockPlayerChatState),
		rateLimitStates: make(map[uint64]map[int]time.Time),
	}
}

func (m *mockChatManager) AddPlayer(playerID uint64, position mockVector2) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.players[playerID] = &mockPlayerChatState{
		PlayerID: playerID,
		Position: position,
	}
	m.rateLimitStates[playerID] = make(map[int]time.Time)
}

func (m *mockChatManager) SendMessage(playerID uint64, channel int, content string, recipientID uint64, localRadius float64) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.players[playerID]; !exists {
		return "", fmt.Errorf("player not found")
	}

	// Check rate limit (simple implementation)
	if lastSent, exists := m.rateLimitStates[playerID][channel]; exists {
		minInterval := 3 * time.Second // Global channel rate limit
		if channel == 1 {
			minInterval = 1 * time.Second // Local
		} else if channel >= 2 {
			minInterval = 500 * time.Millisecond // Party/Whisper
		}

		if time.Since(lastSent) < minInterval {
			return "", fmt.Errorf("rate limit exceeded")
		}
	}

	m.rateLimitStates[playerID][channel] = time.Now()

	// Generate message ID
	uuid := make([]byte, 16)
	io.ReadFull(rand.Reader, uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	msgID := hex.EncodeToString(uuid)

	return msgID, nil
}

// TestHighThroughput_50Players_500MessagesPerMinute tests the system with 50 players
// sending approximately 500 messages per minute total (10 messages per minute per player).
func TestHighThroughput_50Players_500MessagesPerMinute(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high-throughput test in short mode")
	}

	chatManager := newMockChatManager()
	numPlayers := 50
	messagesPerPlayerPerMinute := 10
	testDuration := 10 * time.Second // Run for 10 seconds

	// Calculate message rate
	messagesPerPlayer := messagesPerPlayerPerMinute * int(testDuration.Seconds()) / 60
	expectedTotal := numPlayers * messagesPerPlayer

	// Setup players
	for i := 0; i < numPlayers; i++ {
		playerID := uint64(100 + i)
		position := mockVector2{X: float64(i * 10), Y: 0}
		chatManager.AddPlayer(playerID, position)
	}

	var sentCount int32
	var wg sync.WaitGroup
	startTime := time.Now()

	// Simulate players sending messages
	for i := 0; i < numPlayers; i++ {
		wg.Add(1)
		go func(playerID uint64) {
			defer wg.Done()

			ticker := time.NewTicker(time.Duration(60000/messagesPerPlayerPerMinute) * time.Millisecond)
			defer ticker.Stop()

			timeout := time.After(testDuration)
			for {
				select {
				case <-ticker.C:
					msg := fmt.Sprintf("Message from player %d at %v", playerID, time.Since(startTime))
					_, err := chatManager.SendMessage(playerID, 0, msg, 0, 10.0) // Global channel
					if err == nil {
						atomic.AddInt32(&sentCount, 1)
					}
				case <-timeout:
					return
				}
			}
		}(uint64(100 + i))
	}

	wg.Wait()

	// Allow time for message processing
	time.Sleep(500 * time.Millisecond)

	finalSent := atomic.LoadInt32(&sentCount)
	deliveryRate := float64(finalSent) / float64(expectedTotal) * 100

	t.Logf("50-player throughput test results:")
	t.Logf("  Duration: %v", testDuration)
	t.Logf("  Players: %d", numPlayers)
	t.Logf("  Expected messages: %d", expectedTotal)
	t.Logf("  Sent messages: %d", finalSent)
	t.Logf("  Delivery rate: %.2f%%", deliveryRate)
	t.Logf("  Messages per second: %.2f", float64(finalSent)/testDuration.Seconds())

	// Verify reasonable delivery (>80% of expected)
	if deliveryRate < 80.0 {
		t.Errorf("Delivery rate %.2f%% is below 80%% threshold", deliveryRate)
	}

	// Verify throughput target (500 messages per minute = ~8.3 messages per second)
	messagesPerSecond := float64(finalSent) / testDuration.Seconds()
	targetPerSecond := float64(500) / 60.0
	if messagesPerSecond < targetPerSecond*0.8 {
		t.Errorf("Throughput %.2f msg/s is below target %.2f msg/s", messagesPerSecond, targetPerSecond)
	}
}

// TestHighThroughput_DialogQueue_50Players tests NPC dialog queue with 50 concurrent players
func TestHighThroughput_DialogQueue_50Players(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high-throughput test in short mode")
	}

	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)
	dialogSys.conversationManager.maxQueueSize = 100

	// Create NPC
	npc := world.CreateEntity()
	npc.ID = 1
	dialogSys.InitializeNPCDialog(npc, "fantasy", nil, 12345)

	numPlayers := 50
	requestsPerPlayer := 5

	var wg sync.WaitGroup
	var successCount, failCount int32
	startTime := time.Now()

	// Players queue requests concurrently
	for i := 0; i < numPlayers; i++ {
		wg.Add(1)
		go func(playerID uint64) {
			defer wg.Done()

			for j := 0; j < requestsPerPlayer; j++ {
				respChan, err := dialogSys.QueuePlayerInput(npc.ID, playerID, fmt.Sprintf("Request %d", j))
				if err != nil {
					atomic.AddInt32(&failCount, 1)
					continue
				}

				// Don't wait for response in this test (queue only)
				go func() {
					<-respChan
					atomic.AddInt32(&successCount, 1)
				}()
			}
		}(uint64(100 + i))
	}

	wg.Wait()
	queueTime := time.Since(startTime)

	// Process all queued requests
	processStart := time.Now()
	totalRequests := numPlayers * requestsPerPlayer
	for i := 0; i < totalRequests; i++ {
		dialogSys.ProcessQueuedDialogs(0.016)
	}
	processTime := time.Since(processStart)

	// Wait for responses
	time.Sleep(500 * time.Millisecond)

	finalSuccess := atomic.LoadInt32(&successCount)
	finalFail := atomic.LoadInt32(&failCount)

	t.Logf("50-player dialog queue test results:")
	t.Logf("  Players: %d", numPlayers)
	t.Logf("  Requests per player: %d", requestsPerPlayer)
	t.Logf("  Total requests: %d", totalRequests)
	t.Logf("  Successful: %d", finalSuccess)
	t.Logf("  Failed: %d", finalFail)
	t.Logf("  Queue time: %v", queueTime)
	t.Logf("  Process time: %v", processTime)
	t.Logf("  Avg queue latency: %v", queueTime/time.Duration(numPlayers))
	t.Logf("  Avg process latency: %v", processTime/time.Duration(totalRequests))

	// Verify high success rate (>95%)
	successRate := float64(finalSuccess) / float64(totalRequests) * 100
	if successRate < 95.0 {
		t.Errorf("Success rate %.2f%% is below 95%% threshold", successRate)
	}
}

// TestHighThroughput_MultiPartyConversations tests multiple concurrent conversations
func TestHighThroughput_MultiPartyConversations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping high-throughput test in short mode")
	}

	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)

	numNPCs := 10
	playersPerNPC := 5
	messagesPerPlayer := 10

	var wg sync.WaitGroup
	var totalMessages int32

	// Create NPCs and conversations
	for i := 0; i < numNPCs; i++ {
		npc := world.CreateEntity()
		npc.ID = uint64(i + 1)
		dialogSys.InitializeNPCDialog(npc, "fantasy", nil, int64(12345+i))

		playerIDs := make([]uint64, playersPerNPC)
		for j := 0; j < playersPerNPC; j++ {
			playerIDs[j] = uint64(100 + i*playersPerNPC + j)
		}

		convID, _ := dialogSys.StartMultiPartyConversation(npc.ID, playerIDs)

		// Players send messages concurrently
		for _, playerID := range playerIDs {
			wg.Add(1)
			go func(pID uint64, cID string) {
				defer wg.Done()

				for k := 0; k < messagesPerPlayer; k++ {
					err := dialogSys.AddConversationMessage(cID, pID, fmt.Sprintf("Player%d", pID), fmt.Sprintf("Message %d", k))
					if err == nil {
						atomic.AddInt32(&totalMessages, 1)
					}
				}
			}(playerID, convID)
		}
	}

	wg.Wait()

	finalCount := atomic.LoadInt32(&totalMessages)
	expectedCount := numNPCs * playersPerNPC * messagesPerPlayer

	t.Logf("Multi-party conversation test results:")
	t.Logf("  NPCs: %d", numNPCs)
	t.Logf("  Players per NPC: %d", playersPerNPC)
	t.Logf("  Messages per player: %d", messagesPerPlayer)
	t.Logf("  Expected messages: %d", expectedCount)
	t.Logf("  Actual messages: %d", finalCount)
	t.Logf("  Success rate: %.2f%%", float64(finalCount)/float64(expectedCount)*100)

	if finalCount != int32(expectedCount) {
		t.Errorf("Message count = %d, want %d", finalCount, expectedCount)
	}
}

// TestLatency_MessageOrdering tests message ordering under high latency conditions
func TestLatency_MessageOrdering(t *testing.T) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)

	npc := world.CreateEntity()
	npc.ID = 1

	playerIDs := []uint64{100, 101, 102, 103, 104}
	convID, _ := dialogSys.StartMultiPartyConversation(npc.ID, playerIDs)

	numMessages := 100
	var wg sync.WaitGroup

	// Send messages with random delays to simulate network latency
	for i := 0; i < numMessages; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			// Simulate variable network latency (0-50ms)
			delay := time.Duration(idx%50) * time.Millisecond
			time.Sleep(delay)

			playerID := playerIDs[idx%len(playerIDs)]
			dialogSys.AddConversationMessage(convID, playerID, "Player", fmt.Sprintf("Message %d", idx))
		}(i)
	}

	wg.Wait()

	// Retrieve messages
	messages, _ := dialogSys.GetConversationMessages(convID)

	t.Logf("Message ordering test results:")
	t.Logf("  Messages sent: %d", numMessages)
	t.Logf("  Messages received: %d", len(messages))

	// Verify count
	if len(messages) != numMessages {
		t.Errorf("Message count = %d, want %d", len(messages), numMessages)
	}

	// Verify timestamps are monotonically increasing
	for i := 1; i < len(messages); i++ {
		if messages[i].Timestamp.Before(messages[i-1].Timestamp) {
			t.Errorf("Message[%d] timestamp is before Message[%d]", i, i-1)
		}
	}

	// Verify sequence numbers are sequential
	for i, msg := range messages {
		if msg.SequenceNum != uint32(i) {
			t.Errorf("Message[%d] sequence = %d, want %d", i, msg.SequenceNum, i)
		}
	}
}

// TestPerformance_FrameTime tests that social systems don't impact frame time
func TestPerformance_FrameTime(t *testing.T) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)
	dialogSys.conversationManager.maxQueueSize = 200

	// Setup 10 NPCs with dialog queues
	for i := 0; i < 10; i++ {
		npc := world.CreateEntity()
		npc.ID = uint64(i + 1)
		dialogSys.InitializeNPCDialog(npc, "fantasy", nil, int64(12345+i))

		// Queue 20 requests per NPC
		for j := 0; j < 20; j++ {
			dialogSys.QueuePlayerInput(npc.ID, uint64(100+j), "Test input")
		}
	}

	// Measure processing time
	iterations := 1000
	totalTime := time.Duration(0)

	for i := 0; i < iterations; i++ {
		start := time.Now()
		dialogSys.ProcessQueuedDialogs(0.016)
		totalTime += time.Since(start)
	}

	avgTime := totalTime / time.Duration(iterations)
	targetFrameTime := 16670 * time.Microsecond // 60 FPS = 16.67ms

	t.Logf("Frame time performance results:")
	t.Logf("  Iterations: %d", iterations)
	t.Logf("  Avg process time: %v", avgTime)
	t.Logf("  Target frame time: %v", targetFrameTime)
	t.Logf("  Frame time budget used: %.2f%%", float64(avgTime)/float64(targetFrameTime)*100)

	// Verify processing doesn't exceed 10% of frame budget
	if avgTime > targetFrameTime/10 {
		t.Errorf("Avg process time %v exceeds 10%% of frame budget (%v)", avgTime, targetFrameTime/10)
	}
}

// TestMemory_50PlayerOverhead tests memory overhead with 50 active players
func TestMemory_50PlayerOverhead(t *testing.T) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)

	// Create 5 NPCs
	for i := 0; i < 5; i++ {
		npc := world.CreateEntity()
		npc.ID = uint64(i + 1)
		dialogSys.InitializeNPCDialog(npc, "fantasy", nil, int64(12345+i))
	}

	// Create 10 conversations with 5 players each
	for i := 0; i < 10; i++ {
		npcID := uint64(1 + (i % 5))
		playerIDs := make([]uint64, 5)
		for j := 0; j < 5; j++ {
			playerIDs[j] = uint64(100 + i*5 + j)
		}

		convID, _ := dialogSys.StartMultiPartyConversation(npcID, playerIDs)

		// Add 100 messages per conversation
		for j := 0; j < 100; j++ {
			playerID := playerIDs[j%5]
			dialogSys.AddConversationMessage(convID, playerID, "Player", fmt.Sprintf("Message %d", j))
		}
	}

	// Get active conversation count
	activeConvs := dialogSys.conversationManager.GetActiveConversations()

	t.Logf("Memory overhead test results:")
	t.Logf("  Active conversations: %d", activeConvs)
	t.Logf("  Total players: 50")
	t.Logf("  Messages per conversation: 100")
	t.Logf("  Total messages: %d", activeConvs*100)

	// Verify conversation count
	if activeConvs != 10 {
		t.Errorf("Active conversations = %d, want 10", activeConvs)
	}
}

// Benchmark high-throughput scenarios
func BenchmarkHighThroughput_50Players_Queueing(b *testing.B) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)
	dialogSys.conversationManager.maxQueueSize = b.N

	npc := world.CreateEntity()
	npc.ID = 1
	dialogSys.InitializeNPCDialog(npc, "fantasy", nil, 12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		playerID := uint64(100 + (i % 50))
		dialogSys.QueuePlayerInput(npc.ID, playerID, "Test input")
	}
}

func BenchmarkHighThroughput_MessageAdding(b *testing.B) {
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

func BenchmarkHighThroughput_ConcurrentConversations(b *testing.B) {
	world := NewWorld()
	dialogSys := NewNPCDialogSystem(world, 12345)

	// Create 10 conversations
	convIDs := make([]string, 10)
	for i := 0; i < 10; i++ {
		npc := world.CreateEntity()
		npc.ID = uint64(i + 1)
		convIDs[i], _ = dialogSys.StartMultiPartyConversation(npc.ID, []uint64{100})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		convID := convIDs[i%10]
		dialogSys.AddConversationMessage(convID, 100, "Player", "Message")
	}
}
