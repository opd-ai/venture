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
//   - Persistent image storage with gallery management
//   - Image deduplication and LRU eviction (100 images, 50MB max per player)
//   - Support for PNG and JPEG formats with automatic compression
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
// Decay can be applied automatically using StartAutomaticDecay():
//
//	manager := persistence.NewTrustManager()
//	manager.StartAutomaticDecay(1 * time.Hour) // Check decay every hour
//	defer manager.StopAutomaticDecay()
//
// Alternatively, call ApplyDecay() manually at your preferred interval.
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
//	// Image gallery
//	gallery := persistence.NewImageGallery("player1")
//	img := createSomeImage() // Your image.Image
//	stored, _ := gallery.AddImage(img, "My Screenshot", persistence.ImageFormatJPEG, []string{"screenshot"})
//
//	// Retrieve images by tag
//	screenshots := gallery.GetImagesByTag("screenshot")
//
//	// Get lightweight metadata without image data
//	thumbnails := gallery.GetThumbnails()
//
//	// Save persistent data
//	data, _ := manager.Save()
//	ioutil.WriteFile("trust.json.gz", data, 0644)
//	chatData, _ := history.Save()
//	ioutil.WriteFile("chat.json.gz", chatData, 0644)
//	galleryData, _ := gallery.Save()
//	ioutil.WriteFile("gallery.json.gz", galleryData, 0644)
//
// # Thread Safety
//
// All public types in this package are safe for concurrent use:
//   - [ChatHistory]: Protected by sync.RWMutex; concurrent AddMessage/GetMessages/Save/Load are safe.
//   - [TrustManager]: Protected by sync.RWMutex; concurrent UpdateTrust/GetTrustLevel/Save/Load are safe.
//   - [ReputationManager]: Protected by sync.RWMutex; concurrent UpdateReputation/Save/Load are safe.
//   - [ImageGallery]: Protected by sync.RWMutex; concurrent AddImage/GetImage/Save/Load are safe.
//
// # Delta Synchronization
//
// [ChatHistory.GetDelta] uses a changelog-based approach to accurately track message
// additions and deletions since a given version. The implementation maintains an
// ordered log of changes (up to MaxChangelogSize entries) and queries it to determine
// exactly which messages have been added or deleted since the requested version.
//
// For version 0, all messages are returned (full sync). If fromVersion is older
// than the oldest entry in the changelog, all messages are returned as a fallback.
// This changelog-based implementation (added 2026-02-16) replaced the previous
// heuristic approach to provide accurate delta synchronization for multiplayer
// chat federation.
//
// # Constructor Patterns
//
// All managers in this package follow a dual constructor pattern:
//   - New<Type>(): Default constructor using RealTimeProvider (production)
//   - New<Type>WithTimeProvider(tp): Test constructor for injecting mock time
//
// This pattern is intentionally maintained for API stability rather than using
// functional options. The dual constructor approach provides:
//   - Clear separation between production and test usage
//   - Simpler API surface (no variadic options)
//   - Type-safe TimeProvider injection
//   - Backward compatibility with existing code
//
// Example production usage:
//
//	manager := persistence.NewTrustManager()  // Uses real time
//	history := persistence.NewChatHistory(playerID)
//
// Example test usage:
//
//	mockTime := &FixedTimeProvider{Time: time.Unix(1234567890, 0)}
//	manager := persistence.NewTrustManagerWithTimeProvider(mockTime)
//	history := persistence.NewChatHistoryWithTimeProvider(playerID, mockTime)
package persistence
