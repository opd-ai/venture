package engine

import (
	"testing"
)

func TestNewCourierSystem(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	cs := NewCourierSystem(world, mailSystem)

	if cs == nil {
		t.Fatal("NewCourierSystem returned nil")
	}
	if cs.world != world {
		t.Error("CourierSystem world not set correctly")
	}
	if cs.mailSystem != mailSystem {
		t.Error("CourierSystem mailSystem not set correctly")
	}
	if cs.travelSpeed != 2.0 {
		t.Errorf("Expected default travel speed 2.0, got %f", cs.travelSpeed)
	}
}

func TestCourierSystem_SetServerGraph(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	graph := map[string][]string{
		"A": {"B", "C"},
		"B": {"A", "D"},
		"C": {"A"},
		"D": {"B"},
	}

	cs.SetServerGraph(graph)

	if len(cs.serverGraph) != len(graph) {
		t.Errorf("Expected %d servers in graph, got %d", len(graph), len(cs.serverGraph))
	}

	for server, neighbors := range graph {
		csNeighbors, exists := cs.serverGraph[server]
		if !exists {
			t.Errorf("Server %s not found in courier system graph", server)
			continue
		}
		if len(csNeighbors) != len(neighbors) {
			t.Errorf("Server %s: expected %d neighbors, got %d", server, len(neighbors), len(csNeighbors))
		}
	}
}

func TestCourierSystem_SetTravelSpeed(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	speeds := []float64{1.0, 5.0, 10.5}
	for _, speed := range speeds {
		cs.SetTravelSpeed(speed)
		if cs.travelSpeed != speed {
			t.Errorf("Expected travel speed %f, got %f", speed, cs.travelSpeed)
		}
	}
}

func TestCourierSystem_FindRoute_SameServer(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	route := cs.findRoute("ServerA", "ServerA")
	if len(route) != 1 {
		t.Errorf("Expected route length 1 for same server, got %d", len(route))
	}
	if route[0] != "ServerA" {
		t.Errorf("Expected route to contain ServerA, got %s", route[0])
	}
}

func TestCourierSystem_FindRoute_DirectConnection(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	graph := map[string][]string{
		"A": {"B"},
		"B": {"A"},
	}
	cs.SetServerGraph(graph)

	route := cs.findRoute("A", "B")
	if len(route) != 2 {
		t.Errorf("Expected route length 2, got %d", len(route))
	}
	if route[0] != "A" || route[1] != "B" {
		t.Errorf("Expected route [A, B], got %v", route)
	}
}

func TestCourierSystem_FindRoute_MultiHop(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	graph := map[string][]string{
		"A": {"B"},
		"B": {"A", "C"},
		"C": {"B", "D"},
		"D": {"C"},
	}
	cs.SetServerGraph(graph)

	route := cs.findRoute("A", "D")
	expectedLength := 4 // A -> B -> C -> D
	if len(route) != expectedLength {
		t.Errorf("Expected route length %d, got %d: %v", expectedLength, len(route), route)
	}
	if route[0] != "A" || route[len(route)-1] != "D" {
		t.Errorf("Expected route from A to D, got %v", route)
	}
}

func TestCourierSystem_FindRoute_NoConnection(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	graph := map[string][]string{
		"A": {"B"},
		"B": {"A"},
		"C": {"D"},
		"D": {"C"},
	}
	cs.SetServerGraph(graph)

	// A and C are not connected
	route := cs.findRoute("A", "C")
	// Should return direct route (which would fail in practice)
	if len(route) != 2 {
		t.Errorf("Expected fallback route length 2, got %d", len(route))
	}
}

func TestCourierSystem_AssignDeliveryToCourier(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	cs := NewCourierSystem(world, mailSystem)

	graph := map[string][]string{
		"Server1": {"Server2"},
		"Server2": {"Server1", "Server3"},
		"Server3": {"Server2"},
	}
	cs.SetServerGraph(graph)

	// Create courier entity
	courierID := cs.SpawnCourierNPC(10, 20, "TestCourier")

	// Assign delivery
	messageID := "msg-12345"
	err := cs.AssignDeliveryToCourier(courierID, messageID, "Server1", "Server3")
	if err != nil {
		t.Fatalf("AssignDeliveryToCourier failed: %v", err)
	}

	// Verify courier has the assignment
	entity, _ := world.GetEntity(courierID)
	comp, _ := entity.GetComponent("courier")
	courier := comp.(*CourierComponent)

	if courier.CurrentMessageID != messageID {
		t.Errorf("Expected message ID %s, got %s", messageID, courier.CurrentMessageID)
	}
	if len(courier.CurrentRoute) != 3 {
		t.Errorf("Expected route length 3, got %d", len(courier.CurrentRoute))
	}
	if courier.RouteProgress != 0 {
		t.Errorf("Expected route progress 0, got %d", courier.RouteProgress)
	}
}

func TestCourierSystem_AssignDeliveryToCourier_CourierBusy(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	cs := NewCourierSystem(world, mailSystem)

	// Create courier entity
	courierID := cs.SpawnCourierNPC(10, 20, "BusyCourier")

	// Assign first delivery
	err := cs.AssignDeliveryToCourier(courierID, "msg-1", "A", "B")
	if err != nil {
		t.Fatalf("First assignment failed: %v", err)
	}

	// Try to assign second delivery (should fail)
	err = cs.AssignDeliveryToCourier(courierID, "msg-2", "A", "C")
	if err == nil {
		t.Error("Expected error when assigning to busy courier, got nil")
	}
}

func TestCourierSystem_AssignDeliveryToCourier_InvalidCourier(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	err := cs.AssignDeliveryToCourier(99999, "msg-123", "A", "B")
	if err == nil {
		t.Error("Expected error for invalid courier ID, got nil")
	}
}

func TestCourierSystem_FindAvailableCourier(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	// No couriers initially
	courierID := cs.FindAvailableCourier("Server1")
	if courierID != 0 {
		t.Errorf("Expected no available courier, got ID %d", courierID)
	}

	// Spawn courier
	spawned := cs.SpawnCourierNPC(10, 20, "TestCourier")
	world.UpdateEntities(0)

	// Should find courier
	courierID = cs.FindAvailableCourier("Server1")
	if courierID == 0 {
		t.Error("Expected to find available courier, got 0")
	}
	if courierID != spawned {
		t.Errorf("Expected courier ID %d, got %d", spawned, courierID)
	}

	// Assign delivery to make courier busy
	entity, _ := world.GetEntity(courierID)
	comp, _ := entity.GetComponent("courier")
	courier := comp.(*CourierComponent)
	courier.AssignDelivery("msg-123", []string{"A", "B"})

	// Should not find courier (busy)
	courierID = cs.FindAvailableCourier("Server1")
	if courierID != 0 {
		t.Errorf("Expected no available courier (busy), got ID %d", courierID)
	}
}

func TestCourierSystem_GetCourierStatus(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	courierID := cs.SpawnCourierNPC(10, 20, "TestCourier")

	// Test idle courier
	msgID, server, progress, totalHops, err := cs.GetCourierStatus(courierID)
	if err != nil {
		t.Fatalf("GetCourierStatus failed: %v", err)
	}
	if msgID != "" || server != "" || progress != 0 || totalHops != 0 {
		t.Error("Expected idle courier to return empty values")
	}

	// Assign delivery
	entity, _ := world.GetEntity(courierID)
	comp, _ := entity.GetComponent("courier")
	courier := comp.(*CourierComponent)
	route := []string{"Server1", "Server2", "Server3"}
	courier.AssignDelivery("msg-456", route)

	// Test active courier
	msgID, server, progress, totalHops, err = cs.GetCourierStatus(courierID)
	if err != nil {
		t.Fatalf("GetCourierStatus failed: %v", err)
	}
	if msgID != "msg-456" {
		t.Errorf("Expected message ID 'msg-456', got '%s'", msgID)
	}
	if server != "Server1" {
		t.Errorf("Expected current server 'Server1', got '%s'", server)
	}
	if progress != 0 {
		t.Errorf("Expected progress 0, got %d", progress)
	}
	if totalHops != 3 {
		t.Errorf("Expected total hops 3, got %d", totalHops)
	}
}

func TestCourierSystem_GetCourierStatus_InvalidCourier(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	_, _, _, _, err := cs.GetCourierStatus(99999)
	if err == nil {
		t.Error("Expected error for invalid courier ID, got nil")
	}
}

func TestCourierSystem_EstimateDeliveryTime(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	courierID := cs.SpawnCourierNPC(10, 20, "TestCourier")

	// Idle courier
	time, err := cs.EstimateDeliveryTime(courierID)
	if err != nil {
		t.Fatalf("EstimateDeliveryTime failed: %v", err)
	}
	if time != 0 {
		t.Errorf("Expected 0 time for idle courier, got %f", time)
	}

	// Assign 3-hop delivery
	entity, _ := world.GetEntity(courierID)
	comp, _ := entity.GetComponent("courier")
	courier := comp.(*CourierComponent)
	route := []string{"A", "B", "C", "D"} // 3 remaining hops from start
	courier.AssignDelivery("msg-789", route)

	time, err = cs.EstimateDeliveryTime(courierID)
	if err != nil {
		t.Fatalf("EstimateDeliveryTime failed: %v", err)
	}
	expectedTime := 3.0 * 300.0 // 3 hops × 5 minutes
	if time != expectedTime {
		t.Errorf("Expected delivery time %f, got %f", expectedTime, time)
	}

	// Advance courier one hop
	courier.AdvanceRoute()
	time, err = cs.EstimateDeliveryTime(courierID)
	if err != nil {
		t.Fatalf("EstimateDeliveryTime failed: %v", err)
	}
	expectedTime = 2.0 * 300.0 // 2 remaining hops
	if time != expectedTime {
		t.Errorf("Expected delivery time %f, got %f", expectedTime, time)
	}
}

func TestCourierSystem_SpawnCourierNPC(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)
	cs.SetTravelSpeed(5.0)

	courierID := cs.SpawnCourierNPC(100, 200, "FastCourier")

	if courierID == 0 {
		t.Fatal("SpawnCourierNPC returned 0")
	}

	entity, exists := world.GetEntity(courierID)
	if !exists {
		t.Fatal("Spawned courier entity not found in world")
	}

	// Check position component
	posComp, ok := entity.GetComponent("position")
	if !ok {
		t.Fatal("Courier missing position component")
	}
	pos := posComp.(*PositionComponent)
	if pos.X != 100 || pos.Y != 200 {
		t.Errorf("Expected position (100, 200), got (%f, %f)", pos.X, pos.Y)
	}

	// Check courier component
	courierComp, ok := entity.GetComponent("courier")
	if !ok {
		t.Fatal("Courier missing courier component")
	}
	courier := courierComp.(*CourierComponent)
	if courier.TravelSpeed != 5.0 {
		t.Errorf("Expected travel speed 5.0, got %f", courier.TravelSpeed)
	}

	// Check AI component
	if _, ok := entity.GetComponent("ai"); !ok {
		t.Error("Courier missing AI component")
	}

	// Check sprite component
	if _, ok := entity.GetComponent("sprite"); !ok {
		t.Error("Courier missing sprite component")
	}
}

func TestCourierSystem_SpawnPostOffice(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	buildingID, clerkID := cs.SpawnPostOffice(50, 75, "Bob")

	if buildingID == 0 {
		t.Fatal("SpawnPostOffice returned 0 for buildingID")
	}
	if clerkID == 0 {
		t.Fatal("SpawnPostOffice returned 0 for clerkID")
	}

	// Check building
	building, exists := world.GetEntity(buildingID)
	if !exists {
		t.Fatal("Post office building not found in world")
	}

	posComp, ok := building.GetComponent("position")
	if !ok {
		t.Fatal("Building missing position component")
	}
	pos := posComp.(*PositionComponent)
	if pos.X != 50 || pos.Y != 75 {
		t.Errorf("Expected building position (50, 75), got (%f, %f)", pos.X, pos.Y)
	}

	poComp, ok := building.GetComponent("postoffice")
	if !ok {
		t.Fatal("Building missing postoffice component")
	}
	po := poComp.(*PostOfficeComponent)
	if po.ClerkName != "Bob" {
		t.Errorf("Expected clerk name 'Bob', got '%s'", po.ClerkName)
	}

	// Check clerk
	clerk, exists := world.GetEntity(clerkID)
	if !exists {
		t.Fatal("Clerk not found in world")
	}

	clerkComp, ok := clerk.GetComponent("postoffice_clerk")
	if !ok {
		t.Fatal("Clerk missing postoffice_clerk component")
	}
	clerkData := clerkComp.(*PostOfficeClerkComponent)
	if clerkData.PostOfficeID != buildingID {
		t.Errorf("Expected clerk PostOfficeID %d, got %d", buildingID, clerkData.PostOfficeID)
	}
	if clerkData.GreetingDialogue == "" {
		t.Error("Clerk has empty greeting dialogue")
	}

	// Check clerk has interaction component
	if _, ok := clerk.GetComponent("contextaction"); !ok {
		t.Error("Clerk missing contextaction component")
	}
}

func TestCourierSystem_NotifyDeliveryComplete(t *testing.T) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	courierID := cs.SpawnCourierNPC(10, 20, "TestCourier")

	// Assign delivery
	entity, _ := world.GetEntity(courierID)
	comp, _ := entity.GetComponent("courier")
	courier := comp.(*CourierComponent)
	messageID := "msg-complete"
	courier.AssignDelivery(messageID, []string{"A", "B", "C"})

	// Verify courier is carrying mail
	if !courier.IsCarryingMail() {
		t.Fatal("Courier should be carrying mail before notification")
	}

	// Notify delivery complete
	cs.NotifyDeliveryComplete(messageID)

	// Verify courier is now idle
	if courier.IsCarryingMail() {
		t.Error("Courier should be idle after delivery complete notification")
	}
	if courier.CurrentMessageID != "" {
		t.Errorf("Expected empty message ID, got '%s'", courier.CurrentMessageID)
	}
}

func TestCourierSystem_Update_IdleCourier(t *testing.T) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	cs := NewCourierSystem(world, mailSystem)

	courierID := cs.SpawnCourierNPC(10, 20, "IdleCourier")

	// Update should not affect idle courier
	cs.Update(1.0)

	entity, _ := world.GetEntity(courierID)
	comp, _ := entity.GetComponent("courier")
	courier := comp.(*CourierComponent)

	if courier.IsCarryingMail() {
		t.Error("Idle courier should remain idle after update")
	}
}

// Benchmark tests
func BenchmarkCourierSystem_FindRoute(b *testing.B) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	// Create complex graph
	graph := map[string][]string{
		"A": {"B", "C"},
		"B": {"A", "D", "E"},
		"C": {"A", "F"},
		"D": {"B"},
		"E": {"B", "G"},
		"F": {"C", "G"},
		"G": {"E", "F"},
	}
	cs.SetServerGraph(graph)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.findRoute("A", "G")
	}
}

func BenchmarkCourierSystem_AssignDelivery(b *testing.B) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	cs := NewCourierSystem(world, mailSystem)

	graph := map[string][]string{
		"S1": {"S2"},
		"S2": {"S1", "S3"},
		"S3": {"S2"},
	}
	cs.SetServerGraph(graph)

	// Create courier
	courierID := cs.SpawnCourierNPC(10, 20, "BenchCourier")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Clear assignment between iterations
		entity, _ := world.GetEntity(courierID)
		comp, _ := entity.GetComponent("courier")
		courier := comp.(*CourierComponent)
		courier.CompleteDelivery()

		cs.AssignDeliveryToCourier(courierID, "msg-bench", "S1", "S3")
	}
}

func BenchmarkCourierSystem_Update(b *testing.B) {
	world := NewWorld()
	mailSystem := NewMailSystem(world)
	cs := NewCourierSystem(world, mailSystem)

	// Spawn 10 couriers
	for i := 0; i < 10; i++ {
		cs.SpawnCourierNPC(float64(i*10), float64(i*10), "Courier")
	}
	world.UpdateEntities(0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.Update(0.016) // 60 FPS delta
	}
}

func BenchmarkCourierSystem_SpawnCourier(b *testing.B) {
	world := NewWorld()
	cs := NewCourierSystem(world, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cs.SpawnCourierNPC(100, 200, "BenchCourier")
	}
}
