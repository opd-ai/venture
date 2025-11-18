// Package persistence provides persistent social data structures for Venture.
//
// This package implements:
//   - Persistent trust scores that survive server restarts
//   - Reputation tracking with automatic decay over time
//   - Trust level tiers (Stranger, Acquaintance, Friend, Trusted)
//   - Trade limits based on trust levels
//   - Cross-server trust synchronization via federation
//   - Chat history with delta compression and message filtering
//   - Automatic message cleanup (30-day retention)
//   - LRU eviction (1000 messages per player maximum)
//
// # Trust Levels
//
// Trust scores range from 0.0 to 1.0 and are categorized into tiers:
//   - Stranger: 0.0-0.3 (can trade common items only)
//   - Acquaintance: 0.3-0.6 (can trade common + uncommon items)
//   - Friend: 0.6-0.8 (can trade up to rare items)
//   - Trusted: 0.8-1.0 (can trade all items including legendary)
//
// # Reputation Decay
//
// Trust scores decay over time at a rate of 0.01 per day of inactivity.
// This encourages active social engagement and prevents stale relationships.
//
// # Chat History
//
// Chat history maintains the last 1000 messages per player with automatic
// cleanup of messages older than 30 days. Messages are compressed with gzip
// and support delta synchronization for efficient reconnection. Message
// filtering allows searching by sender, recipient, channel, or date range.
//
// Storage efficiency: ~30KB per 1000 messages (70-90% compression ratio).
//
// # Usage Example
//
//	// Trust management
//	manager := persistence.NewTrustManager()
//	manager.UpdateTrust("player1", "player2", 0.05, time.Now())
//	level := manager.GetTrustLevel("player1", "player2")
//	if level >= persistence.TrustLevelFriend {
//	    // Allow rare item trade
//	}
//
//	// Chat history
//	history := persistence.NewChatHistory("player1")
//	msg := &persistence.Message{
//	    ID: "msg1",
//	    Sender: "player1",
//	    Recipient: "player2",
//	    Channel: "whisper",
//	    Content: "Hello!",
//	    Timestamp: time.Now(),
//	}
//	history.AddMessage(msg)
//
//	// Filter messages
//	filter := &persistence.MessageFilter{Sender: "player1", Channel: "whisper"}
//	messages := history.GetMessages(filter)
//
//	// Delta sync on reconnect
//	delta := history.GetDelta(lastKnownVersion)
//	// Send delta to client for efficient sync
//
//	// Save persistent data
//	data, _ := manager.Save()
//	ioutil.WriteFile("trust.json.gz", data, 0644)
//	chatData, _ := history.Save()
//	ioutil.WriteFile("chat.json.gz", chatData, 0644)
package persistence
