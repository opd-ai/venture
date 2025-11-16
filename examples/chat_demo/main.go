package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/network"
)

// main demonstrates the chat system with E2E encryption and multiple channels.
func main() {
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	fmt.Println("=== Venture Chat System Demonstration ===")
	fmt.Println("Showcasing E2E encryption, channel routing, and rate limiting")

	// Create world for chat system
	world := engine.NewWorld()

	// Create chat system
	chatSystem := engine.NewChatSystem(world)

	// Create network chat manager for E2E encryption demo
	chatManager := network.NewChatManager()

	// Create player entities
	player1 := world.CreateEntity()
	player1.AddComponent(engine.NewChatComponent())
	player1.AddComponent(&engine.PositionComponent{X: 5, Y: 5})
	if *verbose {
		fmt.Printf("Created Player 1 (ID: %d) at position (5, 5)\n", player1.ID)
	}

	player2 := world.CreateEntity()
	player2.AddComponent(engine.NewChatComponent())
	player2.AddComponent(&engine.PositionComponent{X: 8, Y: 8})
	if *verbose {
		fmt.Printf("Created Player 2 (ID: %d) at position (8, 8)\n", player2.ID)
	}

	player3 := world.CreateEntity()
	player3.AddComponent(engine.NewChatComponent())
	player3.AddComponent(&engine.PositionComponent{X: 20, Y: 20})
	if *verbose {
		fmt.Printf("Created Player 3 (ID: %d) at position (20, 20)\n\n", player3.ID)
	}

	// Commit entities to world (required before accessing via GetEntity)
	world.Update(0.0)

	// Demonstrate E2E encryption
	demonstrateEncryption(chatManager, player1.ID, player2.ID, *verbose)

	// Demonstrate chat channels
	demonstrateChannels(chatSystem, world, player1.ID, player2.ID, player3.ID, *verbose)

	// Demonstrate rate limiting
	demonstrateRateLimiting(chatSystem, world, player1.ID, *verbose)

	// Demonstrate range limiting for local chat
	demonstrateRangeLimiting(chatSystem, world, player1.ID, player2.ID, player3.ID, *verbose)

	fmt.Println("\n=== Chat Demo Complete ===")
}

// demonstrateEncryption shows E2E encryption with Diffie-Hellman key exchange.
func demonstrateEncryption(cm *network.ChatManager, player1ID, player2ID uint64, verbose bool) {
	fmt.Println("--- Demonstrating E2E Encryption ---")

	// Get default DH parameters
	dhParams := network.DefaultDHParams()

	// Generate DH keypairs for both players
	player1DH, err := network.GenerateKeyPair(dhParams)
	if err != nil {
		log.Fatalf("Failed to generate Player 1 DH keypair: %v", err)
	}

	player2DH, err := network.GenerateKeyPair(dhParams)
	if err != nil {
		log.Fatalf("Failed to generate Player 2 DH keypair: %v", err)
	}

	if verbose {
		fmt.Printf("Player 1 DH public key: %x...\n", player1DH.PublicKey.Bytes()[:16])
		fmt.Printf("Player 2 DH public key: %x...\n", player2DH.PublicKey.Bytes()[:16])
	}

	// Compute shared secrets
	player1Secret, err := network.ComputeSharedSecret(player1DH.PrivateKey, player2DH.PublicKey, dhParams)
	if err != nil {
		log.Fatalf("Failed to compute Player 1 shared secret: %v", err)
	}

	player2Secret, err := network.ComputeSharedSecret(player2DH.PrivateKey, player1DH.PublicKey, dhParams)
	if err != nil {
		log.Fatalf("Failed to compute Player 2 shared secret: %v", err)
	}

	// Derive AES keys from shared secrets
	player1AESKey := network.DeriveAESKey(player1Secret)
	player2AESKey := network.DeriveAESKey(player2Secret)

	if verbose {
		fmt.Printf("Player 1 AES key: %x...\n", player1AESKey[:16])
		fmt.Printf("Player 2 AES key: %x...\n", player2AESKey[:16])
	}

	// Register encryption keys with chat manager
	cm.AddPlayer(player1ID, network.Vector2{X: 5, Y: 5}, player1AESKey)
	cm.AddPlayer(player2ID, network.Vector2{X: 8, Y: 8}, player2AESKey)

	// Encrypt a message
	plaintext := "Hello from Player 1! This message is encrypted."
	fmt.Printf("\nPlaintext: \"%s\"\n", plaintext)

	ciphertext, err := network.EncryptMessage(player1AESKey, []byte(plaintext))
	if err != nil {
		log.Fatalf("Encryption failed: %v", err)
	}

	fmt.Printf("Ciphertext (first 32 bytes): %x...\n", ciphertext[:32])

	// Decrypt the message
	decrypted, err := network.DecryptMessage(player2AESKey, ciphertext)
	if err != nil {
		log.Fatalf("Decryption failed: %v", err)
	}

	fmt.Printf("Decrypted: \"%s\"\n", string(decrypted))

	if string(decrypted) == plaintext {
		fmt.Println("✓ Encryption/decryption successful")
	} else {
		fmt.Println("✗ Decryption mismatch")
	}

	// Demonstrate that server cannot decrypt
	fmt.Println("\nServer attempting to decrypt (should fail):")
	wrongSecret := big.NewInt(12345) // Wrong shared secret
	wrongKey := network.DeriveAESKey(wrongSecret)
	_, err = network.DecryptMessage(wrongKey, ciphertext)
	if err != nil {
		fmt.Printf("✓ Server decryption failed (expected): %v\n", err)
	} else {
		fmt.Println("✗ Server should not be able to decrypt")
	}

	fmt.Println()
}

// demonstrateChannels shows message delivery across different channels.
func demonstrateChannels(cs *engine.ChatSystem, world *engine.World, player1ID, player2ID, player3ID uint64, verbose bool) {
	fmt.Println("--- Demonstrating Chat Channels ---")

	// Global channel (all players receive)
	fmt.Println("\nGlobal Channel:")
	err := cs.SendMessage(player1ID, engine.ChatGlobal, "Global announcement from Player 1", 0)
	if err != nil {
		log.Printf("Global message failed: %v", err)
	} else {
		fmt.Println("✓ Global message sent successfully")
	}

	// Local channel (range-based, only nearby players receive)
	fmt.Println("\nLocal Channel:")
	err = cs.SendMessage(player1ID, engine.ChatLocal, "Local message from Player 1", 0)
	if err != nil {
		log.Printf("Local message failed: %v", err)
	} else {
		fmt.Println("✓ Local message sent successfully")
	}

	// Whisper channel (direct message to specific player)
	fmt.Println("\nWhisper Channel:")
	err = cs.SendMessage(player1ID, engine.ChatWhisper, "Private whisper to Player 2", player2ID)
	if err != nil {
		log.Printf("Whisper failed: %v", err)
	} else {
		fmt.Println("✓ Whisper sent successfully")
	}

	// Attempt whisper without recipient (should fail)
	fmt.Println("\nWhisper without recipient (should fail):")
	err = cs.SendMessage(player1ID, engine.ChatWhisper, "Invalid whisper", 0)
	if err != nil {
		fmt.Printf("✓ Whisper correctly rejected: %v\n", err)
	} else {
		fmt.Println("✗ Whisper should require recipient ID")
	}

	if verbose {
		// Show message counts for each player
		fmt.Println("\nMessage counts:")
		for _, playerID := range []uint64{player1ID, player2ID, player3ID} {
			entity, _ := world.GetEntity(playerID)
			chatComp, _ := entity.GetComponent("chat")
			chat := chatComp.(*engine.ChatComponent)
			fmt.Printf("  Player %d: %d messages\n", playerID, len(chat.Messages))
		}
	}

	fmt.Println()
}

// demonstrateRateLimiting shows rate limit enforcement and mute mechanics.
func demonstrateRateLimiting(cs *engine.ChatSystem, world *engine.World, playerID uint64, verbose bool) {
	fmt.Println("--- Demonstrating Rate Limiting ---")

	// Global channel has 1 msg/3 sec limit
	fmt.Println("\nGlobal channel rate limit test (1 msg/3 sec):")

	// Send first message (should succeed)
	err := cs.SendMessage(playerID, engine.ChatGlobal, "First global message", 0)
	if err != nil {
		log.Printf("First message failed: %v", err)
	} else {
		fmt.Println("✓ First message sent successfully")
	}

	// Immediately send second message (should fail due to rate limit)
	err = cs.SendMessage(playerID, engine.ChatGlobal, "Second global message (too soon)", 0)
	if err != nil {
		fmt.Printf("✓ Second message correctly rate-limited: %v\n", err)
	} else {
		fmt.Println("✗ Second message should be rate-limited")
	}

	// Wait and send third message (should succeed)
	if verbose {
		fmt.Println("Waiting 3.1 seconds for rate limit cooldown...")
	}
	time.Sleep(3100 * time.Millisecond)

	err = cs.SendMessage(playerID, engine.ChatGlobal, "Third global message (after cooldown)", 0)
	if err != nil {
		log.Printf("Third message failed: %v", err)
	} else {
		fmt.Println("✓ Third message sent after cooldown")
	}

	// Demonstrate mute escalation by violating rate limits multiple times
	fmt.Println("\nMute escalation test (triggering violations):")

	entity, _ := world.GetEntity(playerID)
	chatComp, _ := entity.GetComponent("chat")
	chat := chatComp.(*engine.ChatComponent)

	// Manually apply violations to demonstrate mute escalation
	chat.ApplyMute(1) // First violation: 30s mute
	muteTime1 := chat.MuteExpiry.Sub(time.Now()).Seconds()
	fmt.Printf("✓ First violation: Muted for %.0fs\n", muteTime1)

	chat.ApplyMute(2) // Second violation: 60s mute
	muteTime2 := chat.MuteExpiry.Sub(time.Now()).Seconds()
	fmt.Printf("✓ Second violation: Muted for %.0fs (doubled)\n", muteTime2)

	chat.ApplyMute(3) // Third violation: 120s mute
	muteTime3 := chat.MuteExpiry.Sub(time.Now()).Seconds()
	fmt.Printf("✓ Third violation: Muted for %.0fs (doubled again)\n", muteTime3)

	// Clear mute for continued testing
	chat.MuteExpiry = time.Now()
	chat.ViolationCount = 0

	fmt.Println()
}

// demonstrateRangeLimiting shows local chat range restrictions.
func demonstrateRangeLimiting(cs *engine.ChatSystem, world *engine.World, player1ID, player2ID, player3ID uint64, verbose bool) {
	fmt.Println("--- Demonstrating Range Limiting ---")

	// Player 1 at (5, 5), Player 2 at (8, 8), Player 3 at (20, 20)
	// Default local radius is 10 tiles

	fmt.Println("\nLocal chat delivery based on distance:")

	// Get positions for distance calculations
	entity1, _ := world.GetEntity(player1ID)
	pos1, _ := entity1.GetComponent("position")
	p1 := pos1.(*engine.PositionComponent)

	entity2, _ := world.GetEntity(player2ID)
	pos2, _ := entity2.GetComponent("position")
	p2 := pos2.(*engine.PositionComponent)

	entity3, _ := world.GetEntity(player3ID)
	pos3, _ := entity3.GetComponent("position")
	p3 := pos3.(*engine.PositionComponent)

	// Calculate distances (squared distance for comparison)
	distSq12 := (p2.X-p1.X)*(p2.X-p1.X) + (p2.Y-p1.Y)*(p2.Y-p1.Y)
	distSq13 := (p3.X-p1.X)*(p3.X-p1.X) + (p3.Y-p1.Y)*(p3.Y-p1.Y)
	dist12 := math.Sqrt(distSq12)
	dist13 := math.Sqrt(distSq13)

	fmt.Printf("Player 1 → Player 2 distance: %.2f tiles\n", dist12)
	fmt.Printf("Player 1 → Player 3 distance: %.2f tiles\n", dist13)
	fmt.Printf("Local chat radius: 10 tiles\n\n")

	// Send local message from Player 1
	err := cs.SendMessage(player1ID, engine.ChatLocal, "Local message from Player 1", 0)
	if err != nil {
		log.Printf("Local message failed: %v", err)
	}

	// Check which players received the message
	chat2, _ := entity2.GetComponent("chat")
	chat3, _ := entity3.GetComponent("chat")

	player2Messages := chat2.(*engine.ChatComponent).Messages
	player3Messages := chat3.(*engine.ChatComponent).Messages

	// Filter for local messages
	player2LocalCount := countMessagesFromChannel(player2Messages, engine.ChatLocal)
	player3LocalCount := countMessagesFromChannel(player3Messages, engine.ChatLocal)

	if dist12 <= 10 {
		if player2LocalCount > 0 {
			fmt.Printf("✓ Player 2 received local message (within range: %.2f tiles)\n", dist12)
		} else {
			fmt.Printf("✗ Player 2 should have received message (within range)\n")
		}
	} else {
		if player2LocalCount == 0 {
			fmt.Printf("✓ Player 2 did not receive (out of range: %.2f tiles)\n", dist12)
		} else {
			fmt.Printf("✗ Player 2 should not receive (out of range)\n")
		}
	}

	if dist13 <= 10 {
		if player3LocalCount > 0 {
			fmt.Printf("✓ Player 3 received local message (within range: %.2f tiles)\n", dist13)
		} else {
			fmt.Printf("✗ Player 3 should have received message (within range)\n")
		}
	} else {
		if player3LocalCount == 0 {
			fmt.Printf("✓ Player 3 did not receive (out of range: %.2f tiles)\n", dist13)
		} else {
			fmt.Printf("✗ Player 3 should not receive (out of range)\n")
		}
	}

	// Demonstrate megaphone range extension
	fmt.Println("\nMegaphone range extension (10 → 30 tiles):")

	chat1, _ := entity1.GetComponent("chat")
	player1Chat := chat1.(*engine.ChatComponent)

	// Activate megaphone
	success := player1Chat.ActivateMegaphone()
	if !success {
		fmt.Println("✗ Failed to activate megaphone (no uses remaining)")
		return
	}
	effectiveRadius := player1Chat.GetEffectiveRadius()
	fmt.Printf("Effective radius after megaphone: %.0f tiles\n", effectiveRadius)

	if effectiveRadius == 30 {
		fmt.Println("✓ Megaphone correctly extends range to 30 tiles")
	} else {
		fmt.Printf("✗ Expected 30 tiles, got %.0f\n", effectiveRadius)
	}

	fmt.Println()
}

// countMessagesFromChannel counts messages from a specific channel.
func countMessagesFromChannel(messages []engine.ChatMessage, channel engine.ChatChannel) int {
	count := 0
	for _, msg := range messages {
		if msg.Channel == channel {
			count++
		}
	}
	return count
}
