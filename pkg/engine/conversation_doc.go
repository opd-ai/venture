// Package engine provides the conversation management system for multi-party NPC dialogs.
//
// The conversation manager enables multiple players to interact with NPCs simultaneously
// while maintaining proper turn-taking, message ordering, and conflict resolution.
//
// # Multi-Party Conversations
//
// Conversations support one NPC and multiple players. All participants can see the
// conversation history in timestamp order. Messages are tagged with sender ID, sequence
// number, and server timestamp for consistent ordering across clients.
//
//	cm := engine.NewConversationManager(world)
//	conv, _ := cm.StartConversation(npcID, []uint64{player1, player2, player3})
//	cm.AddMessage(conv.ID, player1, "Player1", "Hello NPC!")
//
// # Turn-Taking and Queue Management
//
// NPCs process dialog requests one at a time using FIFO queuing. When multiple players
// submit requests simultaneously, they are queued and processed in order. The queue has
// a configurable maximum size (default 5) to prevent overflow.
//
//	req, _ := cm.QueueDialogRequest(npcID, playerID, "What quests do you have?")
//	processed, _ := cm.ProcessNextDialogRequest(npcID)
//	// Generate NPC response
//	cm.CompleteDialogRequest(npcID, req.RequestID, response, nil)
//
// # Timeout Handling
//
// Active dialog requests have a timeout (default 30 seconds). If a request is not
// completed within the timeout window, it is automatically failed and the next request
// is processed. This prevents stuck conversations from blocking the queue.
//
//	cm.requestTimeout = 30 * time.Second
//	// Request automatically fails after 30 seconds if not completed
//
// # Message Ordering
//
// Messages in a conversation are ordered by sequence number and timestamp. Clients may
// receive messages out of order due to network conditions, but can reorder them using
// the provided sequence numbers and timestamps.
//
//	messages, _ := cm.GetConversationMessages(convID)
//	// Messages are sorted by sequence number
//	for i, msg := range messages {
//	    fmt.Printf("[%d] %s: %s\n", msg.SequenceNum, msg.SenderName, msg.Content)
//	}
//
// # Conflict Resolution
//
// When multiple players attempt to interact with the same NPC simultaneously:
//   - Requests are queued in arrival order (FIFO)
//   - Only one request is active at a time
//   - Subsequent requests wait in queue
//   - Queue overflow (>5 pending) rejects new requests
//   - Timed-out requests are auto-completed with error
//
// # Performance Considerations
//
// The conversation manager is optimized for high-concurrency scenarios:
//   - Concurrent-safe with RWMutex protection
//   - O(1) conversation lookup by ID
//   - O(1) queue operations (append/pop)
//   - Periodic cleanup of stale conversations (>1 hour inactive)
//   - Minimal allocations in hot paths
//
// Typical performance:
//   - 1000 messages added: <10ms
//   - 100 concurrent queue operations: <5ms
//   - 50 active conversations: <5MB memory
//
// # Integration with NPCDialogSystem
//
// The conversation manager integrates with NPCDialogSystem to provide multi-party
// dialog capabilities:
//
//	dialogSys := engine.NewNPCDialogSystem(world, seed)
//	convID, _ := dialogSys.StartMultiPartyConversation(npcID, playerIDs)
//	respChan, _ := dialogSys.QueuePlayerInput(npcID, playerID, input)
//	dialogSys.ProcessQueuedDialogs(deltaTime) // In update loop
//
// # Cleanup and Maintenance
//
// Conversations should be periodically cleaned up to prevent memory leaks:
//
//	// In game loop or periodic task
//	removed := cm.CleanupStaleConversations()
//	// Removes conversations with >1 hour inactivity
//
// # Thread Safety
//
// All methods are thread-safe and can be called concurrently from multiple goroutines.
// The conversation manager uses fine-grained locking to minimize contention.
package engine
