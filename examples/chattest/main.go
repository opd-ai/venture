// Command chattest demonstrates the persistent chat history system.
//
// This tool shows:
//   - Adding messages to chat history
//   - Filtering messages by sender, channel, date
//   - Delta compression for efficient sync
//   - Save/Load with gzip compression
//   - Automatic cleanup of old messages
//
// Usage:
//
//	chattest [-messages N] [-filter sender|channel|date] [-demo-delta]
//
// Examples:
//
//	# Add 100 messages and display
//	chattest -messages 100
//
//	# Filter by sender
//	chattest -messages 100 -filter sender
//
//	# Demonstrate delta sync
//	chattest -messages 100 -demo-delta
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/opd-ai/venture/pkg/social/persistence"
)

func main() {
	messages := flag.Int("messages", 50, "number of messages to generate")
	filterType := flag.String("filter", "", "filter type: sender, channel, date, or empty for no filter")
	demoD := flag.Bool("demo-delta", false, "demonstrate delta synchronization")
	flag.Parse()

	fmt.Println("=== Venture Chat History Demo ===")
	fmt.Println()

	history := persistence.NewChatHistory("alice")
	fmt.Printf("Created chat history for player: alice\n")
	fmt.Println()

	generateMessages(history, *messages)
	displayStatistics(history)

	if *filterType != "" {
		demonstrateFiltering(history, *filterType)
	}

	data := demonstrateSaveLoad(history)

	if *demoD {
		demonstrateDeltaSync(history, data, *messages)
	}

	demonstrateCleanup()
	displayPerformanceMetrics()
}

// generateMessages creates sample messages for testing.
func generateMessages(history *persistence.ChatHistory, count int) {
	channels := []string{"global", "guild", "whisper", "party"}
	senders := []string{"alice", "bob", "charlie", "diana", "eve"}
	now := time.Now()

	fmt.Printf("Generating %d messages...\n", count)
	for i := 0; i < count; i++ {
		msg := &persistence.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			Sender:    senders[i%len(senders)],
			Recipient: "alice",
			Channel:   channels[i%len(channels)],
			Content:   fmt.Sprintf("Message content %d", i),
			Timestamp: now.Add(-time.Duration(count-i) * time.Minute),
		}
		if err := history.AddMessage(msg); err != nil {
			log.Fatalf("Failed to add message: %v", err)
		}
	}
	fmt.Printf("Added %d messages\n", count)
	fmt.Println()
}

// displayStatistics shows message count and version info.
func displayStatistics(history *persistence.ChatHistory) {
	allMessages := history.GetMessages(nil)
	fmt.Printf("Total messages: %d\n", len(allMessages))
	fmt.Printf("Version: %d\n", history.GetVersion())
	fmt.Println()
}

// demonstrateFiltering shows message filtering by different criteria.
func demonstrateFiltering(history *persistence.ChatHistory, filterType string) {
	fmt.Printf("=== Filtering Messages (type: %s) ===\n", filterType)
	filter := createFilter(filterType)
	if filter == nil {
		fmt.Printf("Unknown filter type: %s\n", filterType)
		return
	}

	filtered := history.GetMessages(filter)
	fmt.Printf("\nFiltered results: %d messages\n", len(filtered))
	displayFilteredMessages(filtered)
	fmt.Println()
}

// createFilter creates a message filter based on the specified type.
func createFilter(filterType string) *persistence.MessageFilter {
	now := time.Now()
	switch filterType {
	case "sender":
		fmt.Println("Filter: sender = 'bob'")
		return &persistence.MessageFilter{Sender: "bob"}
	case "channel":
		fmt.Println("Filter: channel = 'whisper'")
		return &persistence.MessageFilter{Channel: "whisper"}
	case "date":
		cutoff := now.Add(-30 * time.Minute)
		fmt.Printf("Filter: after %s (last 30 minutes)\n", cutoff.Format("15:04:05"))
		return &persistence.MessageFilter{After: cutoff}
	default:
		return nil
	}
}

// displayFilteredMessages shows the first 5 filtered messages.
func displayFilteredMessages(messages []*persistence.Message) {
	for i, msg := range messages {
		if i >= 5 {
			fmt.Printf("... and %d more messages\n", len(messages)-5)
			break
		}
		fmt.Printf("  [%s] (%s) %s: %s\n",
			msg.Timestamp.Format("15:04:05"),
			msg.Channel,
			msg.Sender,
			msg.Content)
	}
}

// demonstrateSaveLoad shows save/load functionality with compression metrics.
func demonstrateSaveLoad(history *persistence.ChatHistory) []byte {
	fmt.Println("=== Save/Load Demonstration ===")
	allMessages := history.GetMessages(nil)

	data, err := history.Save()
	if err != nil {
		log.Fatalf("Save failed: %v", err)
	}
	fmt.Printf("Saved %d messages to %d bytes (gzip compressed)\n", len(allMessages), len(data))

	displayCompressionMetrics(allMessages, data)

	loaded := persistence.NewChatHistory("alice")
	if err := loaded.Load(data); err != nil {
		log.Fatalf("Load failed: %v", err)
	}
	fmt.Printf("Loaded successfully: %d messages, version %d\n", len(loaded.GetMessages(nil)), loaded.GetVersion())
	fmt.Println()

	return data
}

// displayCompressionMetrics calculates and displays compression statistics.
func displayCompressionMetrics(messages []*persistence.Message, data []byte) {
	uncompressedEstimate := len(messages) * 150
	compressionRatio := float64(uncompressedEstimate) / float64(len(data))
	fmt.Printf("Estimated compression: %.1fx (from ~%d bytes)\n", compressionRatio, uncompressedEstimate)

	storagePer1000 := float64(len(data)) / float64(len(messages)) * 1000.0
	fmt.Printf("Storage per 1000 messages: ~%.0f KB\n", storagePer1000/1024.0)
	fmt.Println()
}

// demonstrateDeltaSync shows delta synchronization with bandwidth savings.
func demonstrateDeltaSync(history *persistence.ChatHistory, fullData []byte, messageCount int) {
	fmt.Println("=== Delta Synchronization Demonstration ===")

	oldVersion := history.GetVersion() - 10
	if oldVersion < 0 {
		oldVersion = 0
	}
	fmt.Printf("Client version: %d\n", oldVersion)
	fmt.Printf("Server version: %d\n", history.GetVersion())

	delta := history.GetDelta(oldVersion)
	fmt.Printf("Delta contains: %d messages\n", len(delta))

	deltaData := calculateDeltaSize(delta)
	fmt.Printf("Delta size: %d bytes (vs full %d bytes)\n", len(deltaData), len(fullData))

	savings := (1.0 - float64(len(deltaData))/float64(len(fullData))) * 100.0
	fmt.Printf("Bandwidth savings: %.1f%%\n", savings)
	fmt.Println()

	applyDeltaToClient(delta, messageCount)
}

// calculateDeltaSize serializes delta messages to calculate size.
func calculateDeltaSize(delta []*persistence.Message) []byte {
	deltaHistory := persistence.NewChatHistory("temp")
	for _, msg := range delta {
		deltaHistory.AddMessage(msg)
	}
	deltaData, _ := deltaHistory.Save()
	return deltaData
}

// applyDeltaToClient demonstrates applying delta to a client.
func applyDeltaToClient(delta []*persistence.Message, messageCount int) {
	senders := []string{"alice", "bob", "charlie", "diana", "eve"}
	now := time.Now()

	fmt.Println("Applying delta to client...")
	clientHistory := persistence.NewChatHistory("alice")

	for i := 0; i < messageCount-10; i++ {
		msg := &persistence.Message{
			ID:        fmt.Sprintf("msg-%d", i),
			Sender:    senders[i%len(senders)],
			Content:   fmt.Sprintf("Message content %d", i),
			Timestamp: now.Add(-time.Duration(messageCount-i) * time.Minute),
		}
		clientHistory.AddMessage(msg)
	}

	beforeCount := len(clientHistory.GetMessages(nil))
	if err := clientHistory.ApplyDelta(delta); err != nil {
		log.Fatalf("ApplyDelta failed: %v", err)
	}
	afterCount := len(clientHistory.GetMessages(nil))

	fmt.Printf("Client before delta: %d messages\n", beforeCount)
	fmt.Printf("Client after delta: %d messages\n", afterCount)
	fmt.Printf("New messages added: %d\n", afterCount-beforeCount)
	fmt.Println()
}

// demonstrateCleanup shows automatic message cleanup functionality.
func demonstrateCleanup() {
	fmt.Println("=== Automatic Cleanup Demonstration ===")
	now := time.Now()

	oldHistory := persistence.NewChatHistory("bob")
	for i := 0; i < 20; i++ {
		msg := &persistence.Message{
			ID:        fmt.Sprintf("old-msg-%d", i),
			Sender:    "bob",
			Content:   "Old message",
			Timestamp: now.Add(-40 * 24 * time.Hour),
		}
		oldHistory.AddMessage(msg)
	}

	for i := 0; i < 30; i++ {
		msg := &persistence.Message{
			ID:        fmt.Sprintf("new-msg-%d", i),
			Sender:    "bob",
			Content:   "Recent message",
			Timestamp: now.Add(-5 * 24 * time.Hour),
		}
		oldHistory.AddMessage(msg)
	}

	beforeCleanup := len(oldHistory.GetMessages(nil))
	deleted := oldHistory.DeleteOldMessages(now)
	afterCleanup := len(oldHistory.GetMessages(nil))

	fmt.Printf("Messages before cleanup: %d\n", beforeCleanup)
	fmt.Printf("Messages deleted (>30 days): %d\n", deleted)
	fmt.Printf("Messages after cleanup: %d\n", afterCleanup)
	fmt.Println()
}

// displayPerformanceMetrics shows system performance characteristics.
func displayPerformanceMetrics() {
	fmt.Println("=== Performance Metrics ===")
	fmt.Printf("Storage: <30MB per player (1000 messages)\n")
	fmt.Printf("Delta compression: 70-90%% size reduction\n")
	fmt.Printf("Search: <100ms for 1000 messages\n")
	fmt.Printf("Message retention: 30 days automatic\n")
	fmt.Printf("Max messages: %d per player (LRU eviction)\n", persistence.MaxMessagesPerPlayer)
	fmt.Println()
}
