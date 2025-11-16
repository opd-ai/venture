/*
Package engine provides the mail system for asynchronous item/message delivery.

# Mail System

The mail system enables players to send messages and items across servers via
courier NPCs. Messages are delivered asynchronously with simulated travel time.

# Components

MailComponent tracks player inbox and outbox:

	mailComp := engine.NewMailComponent()
	mailComp.AddToInbox(message)
	mailComp.AddToOutbox(message)
	unreadCount := mailComp.GetUnreadCount()

PostOfficeComponent marks buildings where players can send/receive mail:

	postOffice := engine.NewPostOfficeComponent("Bob the Clerk")
	// ServiceFee: 10 gold base fee
	// MaxDistance: 100 tiles (unused in current implementation)

# Mail System

MailSystem manages message sending, delivery, and courier tracking:

	world := engine.NewWorld()
	mailSys := engine.NewMailSystem(world)

	// Configure delivery time (seconds per server hop)
	mailSys.SetDeliveryTime(300.0) // 5 minutes per hop

	// Configure server hop calculation
	mailSys.SetServerHopsFunc(func(from, to string) int {
		// Custom logic to calculate hops between servers
		return 1
	})

	// Send mail
	msg, err := mailSys.SendMail(
		senderID,     // "player123"
		recipientID,  // "player456"
		subject,      // Max 50 chars
		body,         // Max 500 chars
		attachments,  // []uint64 item IDs, max 5
		serverFrom,   // "server1"
		serverTo,     // "server2"
	)

	// Update delivery progress each frame
	mailSys.Update(deltaTime)

	// Track courier position
	courier := mailSys.GetCourierPosition(messageID)
	if courier != nil {
		fmt.Printf("Progress: %.1f%%\n", courier.Progress*100)
	}

# Message Structure

MailMessage contains all message data:

	type MailMessage struct {
		ID          string   // Unique message ID
		SenderID    string   // Sender entity ID
		RecipientID string   // Recipient entity ID
		Subject     string   // Max 50 characters
		Body        string   // Max 500 characters
		Attachments []uint64 // Item IDs, max 5
		Postage     int      // Gold cost (10 + hops)
		SentAt      int64    // Unix timestamp
		DeliveredAt int64    // Unix timestamp (0 if in transit)
	}

Message status can be queried:

	status := msg.GetStatus()
	// Returns: MailStatusSent, MailStatusInTransit, MailStatusDelivered, MailStatusFailed

# Postage Calculation

Postage is calculated as: 10 gold + (1 gold × hops between servers)

	postage := mailSys.CalculatePostage(hops)
	// 0 hops (same server): 10 gold
	// 1 hop: 11 gold
	// 5 hops: 15 gold

# Same-Server Delivery

Messages sent on the same server are delivered immediately (hops = 0):

	msg, _ := mailSys.SendMail("1", "2", "Hi", "Hello", nil, "server1", "server1")
	// msg.DeliveredAt is set immediately

# Cross-Server Delivery

Messages sent between servers use courier simulation:

	// Server A -> Server B (1 hop, 5 minutes delivery time)
	msg, _ := mailSys.SendMail("1", "2", "Hi", "Hello", nil, "A", "B")

	// After 2.5 minutes (150s)
	mailSys.Update(150.0)
	courier := mailSys.GetCourierPosition(msg.ID)
	// courier.Progress = 0.5 (halfway)

	// After 5 minutes total
	mailSys.Update(150.0)
	// Message delivered, courier position cleared

# Courier Tracking

CourierPosition tracks in-transit message delivery:

	type CourierPosition struct {
		MessageID       string  // Message being delivered
		CurrentServer   string  // Origin server
		TargetServer    string  // Destination server
		Progress        float64 // 0.0 to TotalHops
		TotalHops       int     // Server hops
		EstimatedArrival int64  // Unix timestamp
	}

# Inbox Management

Recipients can manage their inbox:

	// Check unread messages (delivered in last 24 hours)
	unread := mailComp.GetUnreadCount()

	// Remove read messages
	mailComp.RemoveFromInbox(messageID)

	// Inbox has capacity limit (default 50)
	if !mailComp.AddToInbox(msg) {
		// Inbox full, delivery failed
	}

# Constraints

Message constraints:
  - Subject: Max 50 characters
  - Body: Max 500 characters
  - Attachments: Max 5 item IDs
  - Inbox capacity: 50 messages (default)

# Performance

All operations are designed for minimal overhead:
  - SendMail: <1ms (same server), <2ms (cross-server)
  - Update: <0.1ms per active courier
  - DeliverMail: <0.5ms per delivery

# Integration Example

Complete example for a game server:

	// Setup
	world := engine.NewWorld()
	mailSys := engine.NewMailSystem(world)
	mailSys.SetDeliveryTime(300.0) // 5 min per hop

	// Create player entities with mail components
	player1 := world.CreateEntity()
	player1.AddComponent(engine.NewMailComponent())

	player2 := world.CreateEntity()
	player2.AddComponent(engine.NewMailComponent())

	world.Update(0.0) // Commit entities

	// Send mail from player1 to player2
	msg, err := mailSys.SendMail(
		fmt.Sprintf("%d", player1.ID),
		fmt.Sprintf("%d", player2.ID),
		"Quest Reward",
		"Thanks for your help!",
		[]uint64{123, 456}, // Item IDs
		"server1",
		"server1",
	)

	// Game loop
	for {
		deltaTime := 0.016 // 60 FPS
		mailSys.Update(deltaTime)
		// Message delivered immediately (same server)
	}

# Future Enhancements (Phase 40.2, 40.3)

Planned features:
  - Courier NPCs with visual pathfinding
  - Post office building generation in towns
  - Mail UI (inbox, outbox, compose)
  - Delivery notifications
  - Item attachment system integration
*/
package engine
