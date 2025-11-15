package engine

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestDialogQueue_HighThroughput_50Players tests NPC dialog queue with 50 concurrent players
func TestDialogQueue_HighThroughput_50Players(t *testing.T) {
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

// TestMultiPartyConversations_HighThroughput tests multiple concurrent conversations
func TestMultiPartyConversations_HighThroughput(t *testing.T) {
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

// Benchmark tests
// BenchmarkDialogQueue_HighThroughput_50Players_Queueing benchmarks dialog queueing
func BenchmarkDialogQueue_HighThroughput_50Players_Queueing(b *testing.B) {
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

func BenchmarkMultiPartyConversation_MessageAdding(b *testing.B) {
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

func BenchmarkMultiPartyConversation_ConcurrentConversations(b *testing.B) {
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
