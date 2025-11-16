package federation

import (
	"testing"
	"time"
)

func TestNewFederationState(t *testing.T) {
	fs := NewFederationState()

	if fs.ConnectedServers == nil {
		t.Error("ConnectedServers map not initialized")
	}
	if fs.PlayerCounts == nil {
		t.Error("PlayerCounts map not initialized")
	}
	if fs.MarketPrices == nil {
		t.Error("MarketPrices map not initialized")
	}
	if fs.ServerCount() != 0 {
		t.Errorf("Expected 0 servers, got %d", fs.ServerCount())
	}
}

func TestAddServer(t *testing.T) {
	fs := NewFederationState()

	info := &ServerInfo{
		ServerID:    "server1",
		ServerName:  "Test Server",
		Address:     "127.0.0.1:8080",
		Version:     "6.0.0",
		Features:    []string{"travel", "trade"},
		TrustLevel:  TrustVerified,
		PlayerCount: 10,
		Reputation:  0.8,
	}

	fs.AddServer(info)

	if fs.ServerCount() != 1 {
		t.Errorf("Expected 1 server, got %d", fs.ServerCount())
	}

	retrieved, exists := fs.GetServer("server1")
	if !exists {
		t.Fatal("Server not found after adding")
	}
	if retrieved.ServerName != "Test Server" {
		t.Errorf("Expected ServerName 'Test Server', got '%s'", retrieved.ServerName)
	}
	if !retrieved.IsOnline {
		t.Error("Server should be marked online after adding")
	}
	if retrieved.PlayerCount != 10 {
		t.Errorf("Expected PlayerCount 10, got %d", retrieved.PlayerCount)
	}
}

func TestRemoveServer(t *testing.T) {
	fs := NewFederationState()

	info := &ServerInfo{
		ServerID:   "server1",
		ServerName: "Test Server",
	}
	fs.AddServer(info)

	if fs.ServerCount() != 1 {
		t.Fatal("Server not added correctly")
	}

	fs.RemoveServer("server1")

	if fs.ServerCount() != 0 {
		t.Errorf("Expected 0 servers after removal, got %d", fs.ServerCount())
	}

	_, exists := fs.GetServer("server1")
	if exists {
		t.Error("Server still exists after removal")
	}
}

func TestUpdateServer(t *testing.T) {
	fs := NewFederationState()

	info := &ServerInfo{
		ServerID:    "server1",
		PlayerCount: 5,
	}
	fs.AddServer(info)

	// Update player count
	updated := fs.UpdateServer("server1", func(info *ServerInfo) {
		info.PlayerCount = 15
		info.Latency = 200
	})

	if !updated {
		t.Error("UpdateServer should return true for existing server")
	}

	retrieved, _ := fs.GetServer("server1")
	if retrieved.PlayerCount != 15 {
		t.Errorf("Expected PlayerCount 15, got %d", retrieved.PlayerCount)
	}
	if retrieved.Latency != 200 {
		t.Errorf("Expected Latency 200, got %d", retrieved.Latency)
	}

	// Try updating non-existent server
	updated = fs.UpdateServer("nonexistent", func(info *ServerInfo) {
		info.PlayerCount = 99
	})

	if updated {
		t.Error("UpdateServer should return false for non-existent server")
	}
}

func TestGetAllServers(t *testing.T) {
	fs := NewFederationState()

	fs.AddServer(&ServerInfo{ServerID: "server1", ServerName: "Server 1"})
	fs.AddServer(&ServerInfo{ServerID: "server2", ServerName: "Server 2"})
	fs.AddServer(&ServerInfo{ServerID: "server3", ServerName: "Server 3"})

	servers := fs.GetAllServers()
	if len(servers) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(servers))
	}

	// Verify returned list is a copy (not affected by modifications)
	servers[0].ServerName = "Modified"
	original, _ := fs.GetServer("server1")
	if original.ServerName == "Modified" {
		t.Error("GetAllServers should return copies, not references")
	}
}

func TestGetTotalPlayers(t *testing.T) {
	fs := NewFederationState()

	fs.AddServer(&ServerInfo{ServerID: "server1", PlayerCount: 10})
	fs.AddServer(&ServerInfo{ServerID: "server2", PlayerCount: 15})
	fs.AddServer(&ServerInfo{ServerID: "server3", PlayerCount: 5})

	total := fs.GetTotalPlayers()
	if total != 30 {
		t.Errorf("Expected total players 30, got %d", total)
	}
}

func TestMarketPrices(t *testing.T) {
	fs := NewFederationState()

	// Add some market prices
	fs.UpdateMarketPrice("sword", 100.0)
	fs.UpdateMarketPrice("shield", 75.5)
	fs.UpdateMarketPrice("potion", 10.25)

	// Retrieve individual price
	price, exists := fs.GetMarketPrice("sword")
	if !exists {
		t.Error("Market price for sword should exist")
	}
	if price != 100.0 {
		t.Errorf("Expected sword price 100.0, got %f", price)
	}

	// Retrieve non-existent price
	_, exists = fs.GetMarketPrice("nonexistent")
	if exists {
		t.Error("Non-existent item should not have a price")
	}

	// Get all prices
	prices := fs.GetAllMarketPrices()
	if len(prices) != 3 {
		t.Errorf("Expected 3 market prices, got %d", len(prices))
	}
}

func TestCheckStaleServers(t *testing.T) {
	fs := NewFederationState()

	// Add servers
	now := time.Now()
	fs.AddServer(&ServerInfo{ServerID: "fresh"})
	fs.AddServer(&ServerInfo{ServerID: "stale"})

	// Manually update LastSeen for stale server
	fs.mu.Lock()
	if info, exists := fs.ConnectedServers["stale"]; exists {
		info.LastSeen = now.Add(-60 * time.Second)
	}
	fs.mu.Unlock()

	staleServers := fs.CheckStaleServers(30 * time.Second)

	if len(staleServers) != 1 {
		t.Errorf("Expected 1 stale server, got %d", len(staleServers))
		if len(staleServers) > 0 {
			t.Logf("Stale servers: %v", staleServers)
		}
		// Don't continue if we didn't get expected results
		return
	}
	if staleServers[0] != "stale" {
		t.Errorf("Expected stale server ID 'stale', got '%s'", staleServers[0])
	}

	// Verify stale server is marked offline
	info, _ := fs.GetServer("stale")
	if info.IsOnline {
		t.Error("Stale server should be marked offline")
	}

	// Fresh server should still be online
	info, _ = fs.GetServer("fresh")
	if !info.IsOnline {
		t.Error("Fresh server should remain online")
	}
}

func TestOnlineServerCount(t *testing.T) {
	fs := NewFederationState()

	fs.AddServer(&ServerInfo{ServerID: "online1", IsOnline: true})
	fs.AddServer(&ServerInfo{ServerID: "online2", IsOnline: true})
	fs.AddServer(&ServerInfo{ServerID: "offline1", IsOnline: false})

	// Manually set after adding to override AddServer's automatic IsOnline=true
	fs.UpdateServer("offline1", func(info *ServerInfo) {
		info.IsOnline = false
	})

	count := fs.OnlineServerCount()
	if count != 2 {
		t.Errorf("Expected 2 online servers, got %d", count)
	}
}

func TestHeartbeatTimestamp(t *testing.T) {
	fs := NewFederationState()

	before := time.Now()
	time.Sleep(10 * time.Millisecond)
	fs.UpdateHeartbeat()
	after := time.Now()

	lastHeartbeat := fs.GetLastHeartbeat()
	if lastHeartbeat.Before(before) || lastHeartbeat.After(after) {
		t.Error("Heartbeat timestamp not in expected range")
	}
}

func TestMarketSyncTimestamp(t *testing.T) {
	fs := NewFederationState()

	before := time.Now()
	time.Sleep(10 * time.Millisecond)
	fs.UpdateMarketPrice("item1", 50.0)
	after := time.Now()

	lastSync := fs.GetLastMarketSync()
	if lastSync.Before(before) || lastSync.After(after) {
		t.Error("Market sync timestamp not in expected range")
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventAlliance, "Alliance"},
		{EventWar, "War"},
		{EventTreaty, "Treaty"},
		{EventEmbargo, "Embargo"},
		{EventTradePact, "TradePact"},
		{EventType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.eventType.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestNewSyncManager(t *testing.T) {
	fs := NewFederationState()
	sm := NewSyncManager(fs)

	if sm.state != fs {
		t.Error("SyncManager state not set correctly")
	}
	if sm.heartbeatInterval != 10*time.Second {
		t.Errorf("Expected heartbeat interval 10s, got %v", sm.heartbeatInterval)
	}
	if sm.marketSyncInterval != 60*time.Second {
		t.Errorf("Expected market sync interval 60s, got %v", sm.marketSyncInterval)
	}
	if sm.staleTimeout != 30*time.Second {
		t.Errorf("Expected stale timeout 30s, got %v", sm.staleTimeout)
	}
}

func TestSyncManagerIntervals(t *testing.T) {
	fs := NewFederationState()
	sm := NewSyncManager(fs)

	sm.SetHeartbeatInterval(5 * time.Second)
	sm.SetMarketSyncInterval(30 * time.Second)
	sm.SetStaleTimeout(15 * time.Second)

	if sm.heartbeatInterval != 5*time.Second {
		t.Errorf("Expected heartbeat interval 5s, got %v", sm.heartbeatInterval)
	}
	if sm.marketSyncInterval != 30*time.Second {
		t.Errorf("Expected market sync interval 30s, got %v", sm.marketSyncInterval)
	}
	if sm.staleTimeout != 15*time.Second {
		t.Errorf("Expected stale timeout 15s, got %v", sm.staleTimeout)
	}
}

func TestSyncManagerStartStop(t *testing.T) {
	fs := NewFederationState()
	sm := NewSyncManager(fs)

	// Use shorter intervals for faster testing
	sm.SetHeartbeatInterval(50 * time.Millisecond)

	sm.Start()
	time.Sleep(120 * time.Millisecond) // Wait for at least 2 heartbeats
	sm.Stop()

	// Verify heartbeat was updated
	lastHeartbeat := fs.GetLastHeartbeat()
	if time.Since(lastHeartbeat) > 200*time.Millisecond {
		t.Error("Heartbeat not updated during sync loop")
	}
}

func TestCreateHeartbeat(t *testing.T) {
	fs := NewFederationState()
	sm := NewSyncManager(fs)

	msg := sm.CreateHeartbeat("server1", 25)

	if msg.ServerID != "server1" {
		t.Errorf("Expected ServerID 'server1', got '%s'", msg.ServerID)
	}
	if msg.PlayerCount != 25 {
		t.Errorf("Expected PlayerCount 25, got %d", msg.PlayerCount)
	}
	if !msg.IsOnline {
		t.Error("Heartbeat should indicate server is online")
	}
	if time.Since(msg.Timestamp) > time.Second {
		t.Error("Heartbeat timestamp should be recent")
	}
}

func TestProcessHeartbeat(t *testing.T) {
	fs := NewFederationState()
	sm := NewSyncManager(fs)

	// Add server first
	fs.AddServer(&ServerInfo{
		ServerID:    "server1",
		PlayerCount: 10,
		Latency:     100,
	})

	// Process heartbeat with updated values
	msg := &HeartbeatMessage{
		ServerID:    "server1",
		Timestamp:   time.Now(),
		PlayerCount: 30,
		Latency:     200,
		IsOnline:    true,
	}

	sm.ProcessHeartbeat(msg)

	// Verify values were updated
	info, _ := fs.GetServer("server1")
	if info.PlayerCount != 30 {
		t.Errorf("Expected PlayerCount 30, got %d", info.PlayerCount)
	}
	if info.Latency != 200 {
		t.Errorf("Expected Latency 200, got %d", info.Latency)
	}
	if !info.IsOnline {
		t.Error("Server should be marked online")
	}
}

func TestCreateMarketSync(t *testing.T) {
	fs := NewFederationState()
	sm := NewSyncManager(fs)

	// Add some prices
	fs.UpdateMarketPrice("item1", 100.0)
	fs.UpdateMarketPrice("item2", 200.0)

	msg := sm.CreateMarketSync("server1")

	if msg.ServerID != "server1" {
		t.Errorf("Expected ServerID 'server1', got '%s'", msg.ServerID)
	}
	if len(msg.Prices) != 2 {
		t.Errorf("Expected 2 prices, got %d", len(msg.Prices))
	}
	if msg.Prices["item1"] != 100.0 {
		t.Errorf("Expected item1 price 100.0, got %f", msg.Prices["item1"])
	}
}

func TestProcessMarketSync(t *testing.T) {
	fs := NewFederationState()
	sm := NewSyncManager(fs)

	msg := &MarketSyncMessage{
		ServerID:  "server1",
		Timestamp: time.Now(),
		Prices: map[string]float64{
			"item1": 150.0,
			"item2": 250.0,
			"item3": 350.0,
		},
	}

	sm.ProcessMarketSync(msg)

	// Verify all prices were updated
	for itemID, expectedPrice := range msg.Prices {
		price, exists := fs.GetMarketPrice(itemID)
		if !exists {
			t.Errorf("Price for %s not found", itemID)
		}
		if price != expectedPrice {
			t.Errorf("Expected %s price %f, got %f", itemID, expectedPrice, price)
		}
	}
}

func TestConcurrentAccess(t *testing.T) {
	fs := NewFederationState()

	// Concurrent reads and writes
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			fs.AddServer(&ServerInfo{
				ServerID:    string(rune('A' + id)),
				PlayerCount: id * 10,
			})
			fs.GetAllServers()
			fs.GetTotalPlayers()
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify data integrity
	if fs.ServerCount() != 10 {
		t.Errorf("Expected 10 servers, got %d", fs.ServerCount())
	}
}

// Benchmarks

func BenchmarkAddServer(b *testing.B) {
	fs := NewFederationState()
	info := &ServerInfo{
		ServerID:    "server1",
		ServerName:  "Test Server",
		PlayerCount: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.AddServer(info)
	}
}

func BenchmarkUpdateServer(b *testing.B) {
	fs := NewFederationState()
	fs.AddServer(&ServerInfo{ServerID: "server1"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.UpdateServer("server1", func(info *ServerInfo) {
			info.PlayerCount++
		})
	}
}

func BenchmarkGetAllServers(b *testing.B) {
	fs := NewFederationState()
	for i := 0; i < 100; i++ {
		fs.AddServer(&ServerInfo{
			ServerID: string(rune('A' + (i % 26))),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.GetAllServers()
	}
}

func BenchmarkUpdateMarketPrice(b *testing.B) {
	fs := NewFederationState()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fs.UpdateMarketPrice("item1", float64(i))
	}
}

func BenchmarkProcessHeartbeat(b *testing.B) {
	fs := NewFederationState()
	sm := NewSyncManager(fs)
	fs.AddServer(&ServerInfo{ServerID: "server1"})

	msg := &HeartbeatMessage{
		ServerID:    "server1",
		Timestamp:   time.Now(),
		PlayerCount: 10,
		IsOnline:    true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.ProcessHeartbeat(msg)
	}
}

func BenchmarkProcessMarketSync(b *testing.B) {
	fs := NewFederationState()
	sm := NewSyncManager(fs)

	msg := &MarketSyncMessage{
		ServerID:  "server1",
		Timestamp: time.Now(),
		Prices: map[string]float64{
			"item1": 100.0,
			"item2": 200.0,
			"item3": 300.0,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm.ProcessMarketSync(msg)
	}
}
