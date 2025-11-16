package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

func main() {
	mode := flag.String("mode", "pathfind", "Mode: pathfind, delivery, spawn, integration")
	verbose := flag.Bool("v", false, "Verbose output")
	flag.Parse()

	switch *mode {
	case "pathfind":
		testPathfinding(*verbose)
	case "delivery":
		testCourierDelivery(*verbose)
	case "spawn":
		testSpawning(*verbose)
	case "integration":
		testIntegration(*verbose)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		fmt.Println("Available modes: pathfind, delivery, spawn, integration")
	}
}

func testPathfinding(verbose bool) {
	fmt.Println("=== Testing Courier Pathfinding ===")

	world := engine.NewWorld()
	mailSys := engine.NewMailSystem(world)
	courierSys := engine.NewCourierSystem(world, mailSys)

	// Create server graph
	graph := map[string][]string{
		"Fantasy":   {"SciFi", "Horror"},
		"SciFi":     {"Fantasy", "Cyberpunk"},
		"Horror":    {"Fantasy", "PostApoc"},
		"Cyberpunk": {"SciFi", "PostApoc"},
		"PostApoc":  {"Horror", "Cyberpunk"},
	}
	courierSys.SetServerGraph(graph)

	tests := []struct {
		from       string
		to         string
		expectHops int
	}{
		{"Fantasy", "Fantasy", 1},   // Same server
		{"Fantasy", "SciFi", 2},     // Direct connection
		{"Fantasy", "Cyberpunk", 3}, // 2 hops: Fantasy → SciFi → Cyberpunk
		{"Fantasy", "PostApoc", 3},  // 2 hops: Fantasy → Horror → PostApoc
		{"SciFi", "PostApoc", 3},    // 2 hops: SciFi → Cyberpunk → PostApoc
	}

	for i, tt := range tests {
		courierID := courierSys.SpawnCourierNPC(10, 10, fmt.Sprintf("Courier%d", i))
		world.Update(0.0)

		err := courierSys.AssignDeliveryToCourier(courierID, fmt.Sprintf("msg-%d", i), tt.from, tt.to)
		if err != nil {
			fmt.Printf("  ✗ Test %d (%s → %s): Assignment failed - %v\n", i+1, tt.from, tt.to, err)
			continue
		}

		_, _, _, totalHops, err := courierSys.GetCourierStatus(courierID)
		if err != nil {
			fmt.Printf("  ✗ Test %d (%s → %s): Status check failed - %v\n", i+1, tt.from, tt.to, err)
			continue
		}

		if totalHops != tt.expectHops {
			fmt.Printf("  ✗ Test %d (%s → %s): Expected %d hops, got %d\n", i+1, tt.from, tt.to, tt.expectHops, totalHops)
		} else {
			if verbose {
				estimatedTime, _ := courierSys.EstimateDeliveryTime(courierID)
				fmt.Printf("  ✓ Test %d (%s → %s): %d hops, ~%.0f seconds\n", i+1, tt.from, tt.to, totalHops, estimatedTime)
			} else {
				fmt.Printf("  ✓ Test %d (%s → %s): %d hops\n", i+1, tt.from, tt.to, totalHops)
			}
		}
	}

	fmt.Println("=== Test Complete ===")
}

func testCourierDelivery(verbose bool) {
	fmt.Println("=== Testing Courier Delivery Behavior ===")

	world := engine.NewWorld()
	mailSys := engine.NewMailSystem(world)
	mailSys.SetDeliveryTime(1.0) // 1 second per hop for testing
	courierSys := engine.NewCourierSystem(world, mailSys)

	// Set up server graph
	graph := map[string][]string{
		"ServerA": {"ServerB"},
		"ServerB": {"ServerA", "ServerC"},
		"ServerC": {"ServerB"},
	}
	courierSys.SetServerGraph(graph)

	// Create sender and recipient
	sender := world.CreateEntity()
	sender.AddComponent(engine.NewMailComponent())
	recipient := world.CreateEntity()
	recipient.AddComponent(engine.NewMailComponent())
	world.Update(0.0)

	// Send mail (will spawn courier automatically)
	senderID := fmt.Sprintf("%d", sender.ID)
	recipientID := fmt.Sprintf("%d", recipient.ID)
	msg, err := mailSys.SendMail(senderID, recipientID, "Courier Test", "Testing courier delivery", nil, "ServerA", "ServerC")
	if err != nil {
		fmt.Printf("Error sending mail: %v\n", err)
		return
	}

	fmt.Printf("Message sent: ID=%s\n", msg.ID[:8])
	fmt.Printf("Route: ServerA → ServerB → ServerC (2 hops, 2 seconds)\n")

	// Spawn a courier to handle delivery
	courierID := courierSys.SpawnCourierNPC(100, 100, "TestCourier")
	world.Update(0.0)

	// Simulate courier delivery
	steps := 25 // Extra steps to ensure delivery
	for i := 0; i < steps; i++ {
		time.Sleep(100 * time.Millisecond)
		courierSys.Update(0.1)
		mailSys.Update(0.1)

		// Check courier status
		msgID, currentServer, progress, totalHops, err := courierSys.GetCourierStatus(courierID)
		if err == nil && msgID != "" {
			deliveryTime, _ := courierSys.EstimateDeliveryTime(courierID)
			if verbose || i%5 == 0 {
				fmt.Printf("  Courier: Server=%s, Progress=%d/%d, ETA=%.0fs\n",
					currentServer, progress, totalHops, deliveryTime)
			}
		} else if verbose && msgID == "" {
			fmt.Println("  Courier: Idle (delivery complete)")
		}

		// Check delivery progress
		courierPos := mailSys.GetCourierPosition(msg.ID)
		if courierPos != nil && (verbose || i%5 == 0) {
			fmt.Printf("  Mail: Progress=%.1f%%, Status=%s\n", courierPos.Progress*100, msg.GetStatus())
		}
	}

	// Check if delivered
	recipientMail, ok := recipient.GetComponent("mail")
	if !ok {
		fmt.Println("✗ Recipient has no mail component")
		return
	}
	mailComp := recipientMail.(*engine.MailComponent)

	if len(mailComp.Inbox) > 0 {
		fmt.Printf("✓ Message delivered successfully!\n")
		if verbose {
			fmt.Printf("  Subject: '%s'\n", mailComp.Inbox[0].Subject)
			fmt.Printf("  Body: '%s'\n", mailComp.Inbox[0].Body)
		}
	} else {
		fmt.Println("✗ Message not delivered")
	}

	fmt.Println("=== Test Complete ===")
}

func testSpawning(verbose bool) {
	fmt.Println("=== Testing Post Office & Courier Spawning ===")

	world := engine.NewWorld()
	mailSys := engine.NewMailSystem(world)
	courierSys := engine.NewCourierSystem(world, mailSys)

	// Test courier spawning
	fmt.Println("\n1. Spawning Couriers:")
	courierNames := []string{"Swift", "Fast", "Quick"}
	courierIDs := make([]uint64, len(courierNames))

	for i, name := range courierNames {
		courierIDs[i] = courierSys.SpawnCourierNPC(float64(i*10), float64(i*10), name)
		world.Update(0.0)

		entity, exists := world.GetEntity(courierIDs[i])
		if !exists {
			fmt.Printf("  ✗ Courier '%s' not found\n", name)
			continue
		}

		// Verify components
		hasPosition := false
		hasCourier := false
		hasAI := false

		if _, ok := entity.GetComponent("position"); ok {
			hasPosition = true
		}
		if _, ok := entity.GetComponent("courier"); ok {
			hasCourier = true
		}
		if _, ok := entity.GetComponent("ai"); ok {
			hasAI = true
		}

		if hasPosition && hasCourier && hasAI {
			if verbose {
				fmt.Printf("  ✓ Courier '%s' (ID=%d): All components present\n", name, courierIDs[i])
			} else {
				fmt.Printf("  ✓ Courier '%s' spawned\n", name)
			}
		} else {
			fmt.Printf("  ✗ Courier '%s': Missing components (pos=%v, courier=%v, ai=%v)\n",
				name, hasPosition, hasCourier, hasAI)
		}
	}

	// Test post office spawning
	fmt.Println("\n2. Spawning Post Offices:")
	clerkNames := []string{"Bob", "Alice", "Charlie"}

	for i, clerkName := range clerkNames {
		buildingID, clerkID := courierSys.SpawnPostOffice(float64(i*20), float64(i*20), clerkName)
		world.Update(0.0)

		building, exists := world.GetEntity(buildingID)
		if !exists {
			fmt.Printf("  ✗ Post office building not found\n")
			continue
		}

		clerk, exists := world.GetEntity(clerkID)
		if !exists {
			fmt.Printf("  ✗ Clerk '%s' not found\n", clerkName)
			continue
		}

		// Verify building components
		_, hasPos := building.GetComponent("position")
		poComp, hasPostOffice := building.GetComponent("postoffice")

		// Verify clerk components
		_, hasClerkPos := clerk.GetComponent("position")
		clerkComp, hasClerkComponent := clerk.GetComponent("postoffice_clerk")

		if hasPos && hasPostOffice && hasClerkPos && hasClerkComponent {
			if verbose {
				po := poComp.(*engine.PostOfficeComponent)
				clerkData := clerkComp.(*engine.PostOfficeClerkComponent)
				fmt.Printf("  ✓ Post Office (Building=%d, Clerk=%d):\n", buildingID, clerkID)
				fmt.Printf("    Clerk: %s, Fee=%d, Office Fee=%d\n",
					po.ClerkName, clerkData.ServiceFee, po.ServiceFee)
			} else {
				fmt.Printf("  ✓ Post Office with clerk '%s' spawned\n", clerkName)
			}
		} else {
			fmt.Printf("  ✗ Post Office incomplete (building: pos=%v, po=%v; clerk: pos=%v, clerk=%v)\n",
				hasPos, hasPostOffice, hasClerkPos, hasClerkComponent)
		}
	}

	// Test finding available couriers
	fmt.Println("\n3. Finding Available Couriers:")
	availableID := courierSys.FindAvailableCourier("TestServer")
	if availableID != 0 {
		fmt.Printf("  ✓ Found available courier: ID=%d\n", availableID)
	} else {
		fmt.Println("  ✗ No available courier found")
	}

	fmt.Println("\n=== Test Complete ===")
}

func testIntegration(verbose bool) {
	fmt.Println("=== Testing Full Mail + Courier Integration ===")

	world := engine.NewWorld()
	mailSys := engine.NewMailSystem(world)
	mailSys.SetDeliveryTime(0.5) // Fast delivery for testing
	courierSys := engine.NewCourierSystem(world, mailSys)

	// Set up multi-server network
	graph := map[string][]string{
		"Hub":   {"North", "South", "East", "West"},
		"North": {"Hub"},
		"South": {"Hub"},
		"East":  {"Hub"},
		"West":  {"Hub"},
	}
	courierSys.SetServerGraph(graph)

	// Create post offices in each server
	fmt.Println("\n1. Setting up post offices:")
	servers := []string{"Hub", "North", "South", "East", "West"}
	for i, server := range servers {
		buildingID, clerkID := courierSys.SpawnPostOffice(float64(i*30), float64(i*30), fmt.Sprintf("Clerk_%s", server))
		world.Update(0.0)
		if verbose {
			fmt.Printf("  ✓ Post office in %s: Building=%d, Clerk=%d\n", server, buildingID, clerkID)
		} else {
			fmt.Printf("  ✓ Post office in %s\n", server)
		}
	}

	// Spawn couriers
	fmt.Println("\n2. Spawning courier fleet:")
	courierCount := 5
	for i := 0; i < courierCount; i++ {
		courierID := courierSys.SpawnCourierNPC(float64(i*15), float64(i*15), fmt.Sprintf("Courier_%d", i+1))
		world.Update(0.0)
		if verbose {
			fmt.Printf("  ✓ Courier_%d spawned (ID=%d)\n", i+1, courierID)
		}
	}
	if !verbose {
		fmt.Printf("  ✓ %d couriers spawned\n", courierCount)
	}

	// Create players
	fmt.Println("\n3. Creating players:")
	playerCount := 10
	players := make([]*engine.Entity, playerCount)
	for i := 0; i < playerCount; i++ {
		player := world.CreateEntity()
		player.AddComponent(engine.NewMailComponent())
		players[i] = player
	}
	world.Update(0.0)
	fmt.Printf("  ✓ %d players created\n", playerCount)

	// Send mail between servers
	fmt.Println("\n4. Sending cross-server mail:")
	messageCount := 20
	sentMessages := 0
	for i := 0; i < messageCount; i++ {
		senderIdx := i % playerCount
		recipientIdx := (i + 1) % playerCount

		fromServer := servers[i%len(servers)]
		toServer := servers[(i+1)%len(servers)]

		senderID := fmt.Sprintf("%d", players[senderIdx].ID)
		recipientID := fmt.Sprintf("%d", players[recipientIdx].ID)

		msg, err := mailSys.SendMail(
			senderID,
			recipientID,
			fmt.Sprintf("Cross-Server %d", i),
			fmt.Sprintf("Message from %s to %s", fromServer, toServer),
			nil,
			fromServer,
			toServer,
		)

		if err != nil {
			if verbose {
				fmt.Printf("  ✗ Message %d failed: %v\n", i+1, err)
			}
		} else {
			sentMessages++
			if verbose {
				fmt.Printf("  ✓ Message %d: %s → %s (Postage=%d)\n", i+1, fromServer, toServer, msg.Postage)
			}
		}
	}
	fmt.Printf("  ✓ Sent %d/%d messages\n", sentMessages, messageCount)

	// Process deliveries
	fmt.Println("\n5. Processing deliveries:")
	start := time.Now()
	updateCount := 20
	for i := 0; i < updateCount; i++ {
		courierSys.Update(0.05)
		mailSys.Update(0.05)
		time.Sleep(50 * time.Millisecond)

		if verbose && i%5 == 0 {
			activeCouriers := 0
			entities := world.GetEntitiesWith("courier")
			for _, entity := range entities {
				comp, ok := entity.GetComponent("courier")
				if ok {
					courier := comp.(*engine.CourierComponent)
					if courier.IsCarryingMail() {
						activeCouriers++
					}
				}
			}
			fmt.Printf("  Update %d: Active couriers=%d\n", i+1, activeCouriers)
		}
	}
	duration := time.Since(start)

	// Count delivered messages
	fmt.Println("\n6. Checking delivery results:")
	deliveredCount := 0
	for i, player := range players {
		mailComp, ok := player.GetComponent("mail")
		if ok {
			inbox := mailComp.(*engine.MailComponent).Inbox
			if len(inbox) > 0 {
				deliveredCount += len(inbox)
				if verbose {
					fmt.Printf("  Player %d: %d messages in inbox\n", i+1, len(inbox))
				}
			}
		}
	}

	fmt.Printf("\n=== Results ===\n")
	fmt.Printf("Messages sent: %d\n", sentMessages)
	fmt.Printf("Messages delivered: %d\n", deliveredCount)
	fmt.Printf("Delivery rate: %.2f%%\n", float64(deliveredCount)/float64(sentMessages)*100)
	fmt.Printf("Processing time: %v\n", duration)
	fmt.Println("=== Test Complete ===")
}
