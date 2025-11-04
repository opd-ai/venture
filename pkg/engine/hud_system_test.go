package engine

import (
	"testing"
	"time"
)

// mockNetworkClient implements NetworkClient interface for testing.
type mockNetworkClient struct {
	latency   time.Duration
	connected bool
}

func (m *mockNetworkClient) GetLatency() time.Duration {
	return m.latency
}

func (m *mockNetworkClient) IsConnected() bool {
	return m.connected
}

// TestSetNetworkClient verifies that network client can be set and retrieved.
func TestSetNetworkClient(t *testing.T) {
	hud := NewEbitenHUDSystem(800, 600)

	// Initially no network client
	if hud.networkClient != nil {
		t.Error("Expected no network client initially")
	}

	// Set network client
	client := &mockNetworkClient{
		latency:   50 * time.Millisecond,
		connected: true,
	}
	hud.SetNetworkClient(client)

	if hud.networkClient == nil {
		t.Fatal("Expected network client to be set")
	}

	if hud.networkClient.GetLatency() != 50*time.Millisecond {
		t.Errorf("Expected latency 50ms, got %v", hud.networkClient.GetLatency())
	}

	if !hud.networkClient.IsConnected() {
		t.Error("Expected client to be connected")
	}

	// Clear network client
	hud.SetNetworkClient(nil)
	if hud.networkClient != nil {
		t.Error("Expected network client to be cleared")
	}
}

// TestNetworkStatusDisplay verifies network status is only shown in multiplayer.
func TestNetworkStatusDisplay(t *testing.T) {
	hud := NewEbitenHUDSystem(800, 600)

	// Create a player entity with health component
	world := NewWorld()
	player := world.CreateEntity()
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	hud.SetPlayerEntity(player)

	// Without network client, drawNetworkStatus should be safe to call
	hud.drawNetworkStatus() // Should not panic

	// With network client but disconnected
	client := &mockNetworkClient{
		latency:   100 * time.Millisecond,
		connected: false,
	}
	hud.SetNetworkClient(client)
	hud.drawNetworkStatus() // Should not panic (not connected)

	// With connected client
	client.connected = true
	// Note: We can't actually test rendering without Ebiten initialization,
	// but we can verify the method doesn't panic
	hud.drawNetworkStatus() // Should not panic
}

// StubHUDSystem is a test implementation of UISystem for HUD testing.
type StubHUDSystem struct {
	UpdateCount  int
	DrawCount    int
	active       bool
	playerEntity *Entity
}

// NewStubHUDSystem creates a new stub HUD system for testing.
func NewStubHUDSystem() *StubHUDSystem {
	return &StubHUDSystem{
		active: true,
	}
}

// SetPlayerEntity sets the player entity (no-op for testing).
func (s *StubHUDSystem) SetPlayerEntity(entity *Entity) {
	s.playerEntity = entity
}

// Update is called every frame but stub doesn't need to do anything.
func (s *StubHUDSystem) Update(entities []*Entity, deltaTime float64) {
	s.UpdateCount++
}

// Draw increments the draw counter. Implements UISystem interface.
func (s *StubHUDSystem) Draw(screen interface{}) {
	s.DrawCount++
}

// IsActive returns whether the HUD is currently visible.
// Implements UISystem interface.
func (s *StubHUDSystem) IsActive() bool {
	return s.active
}

// SetActive sets whether the HUD is visible.
// Implements UISystem interface.
func (s *StubHUDSystem) SetActive(active bool) {
	s.active = active
}

// Compile-time check that StubHUDSystem implements UISystem
var _ UISystem = (*StubHUDSystem)(nil)
