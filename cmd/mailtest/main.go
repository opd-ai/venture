package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

func main() {
	mode := flag.String("mode", "send", "Mode: send, delivery, stress")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	switch *mode {
	case "send":
		testSendMail(*verbose)
	case "delivery":
		testDelivery(*verbose)
	case "stress":
		testStress(*verbose)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		fmt.Println("Available modes: send, delivery, stress")
	}
}

func testSendMail(verbose bool) {
	fmt.Println("=== Testing Mail Sending ===")

	world := engine.NewWorld()
	mailSys := engine.NewMailSystem(world)

	sender := world.CreateEntity()
	sender.AddComponent(engine.NewMailComponent())

	recipient := world.CreateEntity()
	recipient.AddComponent(engine.NewMailComponent())

	world.Update(0.0)

	senderID := fmt.Sprintf("%d", sender.ID)
	recipientID := fmt.Sprintf("%d", recipient.ID)

	tests := []struct {
		name        string
		subject     string
		body        string
		attachments []uint64
		serverFrom  string
		serverTo    string
		expectErr   bool
	}{
		{"Valid same-server", "Hello", "Test message", nil, "server1", "server1", false},
		{"Valid cross-server", "Quest", "Reward attached", []uint64{123, 456}, "server1", "server2", false},
		{"Subject too long", string(make([]byte, 51)), "Body", nil, "server1", "server1", true},
		{"Body too long", "Subject", string(make([]byte, 501)), nil, "server1", "server1", true},
		{"Too many attachments", "Items", "6 items", []uint64{1, 2, 3, 4, 5, 6}, "server1", "server1", true},
	}

	for i, tt := range tests {
		msg, err := mailSys.SendMail(senderID, recipientID, tt.subject, tt.body, tt.attachments, tt.serverFrom, tt.serverTo)

		if tt.expectErr {
			if err == nil {
				fmt.Printf("  ✗ Test %d (%s): Expected error, got success\n", i+1, tt.name)
			} else {
				if verbose {
					fmt.Printf("  ✓ Test %d (%s): Correctly rejected - %v\n", i+1, tt.name, err)
				} else {
					fmt.Printf("  ✓ Test %d (%s): Correctly rejected\n", i+1, tt.name)
				}
			}
		} else {
			if err != nil {
				fmt.Printf("  ✗ Test %d (%s): Unexpected error - %v\n", i+1, tt.name, err)
			} else {
				status := msg.GetStatus()
				if verbose {
					fmt.Printf("  ✓ Test %d (%s): Success - ID=%s, Status=%s, Postage=%d\n",
						i+1, tt.name, msg.ID[:8], status, msg.Postage)
				} else {
					fmt.Printf("  ✓ Test %d (%s): Success (status=%s)\n", i+1, tt.name, status)
				}
			}
		}
	}

	fmt.Println("=== Test Complete ===")
}

func testDelivery(verbose bool) {
	fmt.Println("=== Testing Mail Delivery ===")

	world := engine.NewWorld()
	mailSys := engine.NewMailSystem(world)
	mailSys.SetDeliveryTime(1.0) // 1 second per hop for testing

	sender := world.CreateEntity()
	sender.AddComponent(engine.NewMailComponent())

	recipient := world.CreateEntity()
	recipient.AddComponent(engine.NewMailComponent())

	world.Update(0.0)

	senderID := fmt.Sprintf("%d", sender.ID)
	recipientID := fmt.Sprintf("%d", recipient.ID)

	// Send cross-server message (1 hop, 1 second delivery)
	msg, err := mailSys.SendMail(senderID, recipientID, "Test", "Cross-server delivery", nil, "server1", "server2")
	if err != nil {
		fmt.Printf("Error sending mail: %v\n", err)
		return
	}

	fmt.Printf("Message sent: ID=%s\n", msg.ID[:8])

	// Simulate delivery progress
	steps := 11 // Extra step to ensure delivery completes
	for i := 0; i < steps; i++ {
		time.Sleep(100 * time.Millisecond)
		mailSys.Update(0.1)

		courier := mailSys.GetCourierPosition(msg.ID)
		if courier != nil {
			progress := courier.Progress * 100
			if verbose || i < 10 {
				fmt.Printf("  Progress: %.1f%%\n", progress)
			}
		} else {
			if verbose {
				fmt.Println("  Courier cleared (delivery complete)")
			}
		}
	}

	// Check recipient inbox
	recipientMail, ok := recipient.GetComponent("mail")
	if !ok {
		fmt.Println("Error: Recipient has no mail component")
		return
	}
	mailComp := recipientMail.(*engine.MailComponent)

	if len(mailComp.Inbox) > 0 {
		deliveredMsg := mailComp.Inbox[0]
		fmt.Printf("✓ Message delivered to inbox: Subject='%s', Body='%s'\n",
			deliveredMsg.Subject, deliveredMsg.Body)
	} else {
		fmt.Println("✗ Message not in inbox")
	}

	fmt.Println("=== Test Complete ===")
}

func testStress(verbose bool) {
	fmt.Println("=== Stress Testing Mail System ===")

	world := engine.NewWorld()
	mailSys := engine.NewMailSystem(world)
	mailSys.SetDeliveryTime(0.1) // Fast delivery for stress test

	// Create 100 players
	playerCount := 100
	players := make([]*engine.Entity, playerCount)
	for i := 0; i < playerCount; i++ {
		player := world.CreateEntity()
		player.AddComponent(engine.NewMailComponent())
		players[i] = player
	}
	world.Update(0.0)

	// Send 1000 messages
	messageCount := 1000
	start := time.Now()

	for i := 0; i < messageCount; i++ {
		senderIdx := i % playerCount
		recipientIdx := (i + 1) % playerCount

		senderID := fmt.Sprintf("%d", players[senderIdx].ID)
		recipientID := fmt.Sprintf("%d", players[recipientIdx].ID)

		_, err := mailSys.SendMail(
			senderID,
			recipientID,
			fmt.Sprintf("Message %d", i),
			fmt.Sprintf("Body of message %d", i),
			nil,
			"server1",
			"server2",
		)

		if err != nil && verbose {
			fmt.Printf("Error sending message %d: %v\n", i, err)
		}
	}

	sendDuration := time.Since(start)
	fmt.Printf("Sent %d messages in %v (%.2f messages/sec)\n",
		messageCount, sendDuration, float64(messageCount)/sendDuration.Seconds())

	// Process delivery
	start = time.Now()
	for i := 0; i < 20; i++ { // 2 seconds total (0.1s delivery time × 20 updates)
		mailSys.Update(0.1)
	}
	deliveryDuration := time.Since(start)

	// Count delivered messages
	delivered := 0
	for _, player := range players {
		mailComp, ok := player.GetComponent("mail")
		if ok {
			delivered += len(mailComp.(*engine.MailComponent).Inbox)
		}
	}

	fmt.Printf("Delivered %d/%d messages in %v\n", delivered, messageCount, deliveryDuration)
	fmt.Printf("Delivery rate: %.2f messages/sec\n", float64(delivered)/deliveryDuration.Seconds())

	fmt.Println("=== Test Complete ===")
}
