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

// TestHudDrawHealthBar_ZeroMaxHealth verifies no panic when health.Max is 0.
func TestHudDrawHealthBar_ZeroMaxHealth(t *testing.T) {
	hud := NewEbitenHUDSystem(800, 600)
	world := NewWorld()
	player := world.CreateEntity()
	player.AddComponent(&HealthComponent{Current: 0, Max: 0})
	hud.SetPlayerEntity(player)

	// Draw with nil screen should safely return (screen type assertion fails)
	hud.Draw(nil) // Should not panic
}

// TestHudDrawHealthBar_NilPlayerEntity verifies safe draw with nil player.
func TestHudDrawHealthBar_NilPlayerEntity(t *testing.T) {
	hud := NewEbitenHUDSystem(800, 600)
	// playerEntity is nil
	hud.Draw(nil) // Should not panic
}

// TestHudDrawHealthBar_MissingHealthComponent verifies safe draw without health.
func TestHudDrawHealthBar_MissingHealthComponent(t *testing.T) {
	hud := NewEbitenHUDSystem(800, 600)
	world := NewWorld()
	player := world.CreateEntity()
	hud.SetPlayerEntity(player)

	// drawHealthBar should safely return with no health component
	hud.drawHealthBar()
}

// mockTerritoryBonusProvider implements TerritoryBonusProvider for testing.
type mockTerritoryBonusProvider struct {
	resourceBonus float64
	xpBonus       float64
}

func (m *mockTerritoryBonusProvider) GetBonusesForGuild(guildID string) (float64, float64) {
	if guildID == "" {
		return 0, 0
	}
	return m.resourceBonus, m.xpBonus
}

// TestSetTerritoryBonusProvider verifies that territory bonus provider can be set.
func TestSetTerritoryBonusProvider(t *testing.T) {
	hud := NewEbitenHUDSystem(800, 600)

	// Initially no territory bonus provider
	if hud.territoryBonusProvider != nil {
		t.Error("Expected no territory bonus provider initially")
	}

	// Set territory bonus provider
	provider := &mockTerritoryBonusProvider{
		resourceBonus: 0.10, // 10% bonus
		xpBonus:       0.15, // 15% bonus
	}
	hud.SetTerritoryBonusProvider(provider)

	if hud.territoryBonusProvider == nil {
		t.Fatal("Expected territory bonus provider to be set")
	}

	resourceBonus, xpBonus := hud.territoryBonusProvider.GetBonusesForGuild("test_guild")
	if resourceBonus != 0.10 {
		t.Errorf("Expected resource bonus 0.10, got %v", resourceBonus)
	}
	if xpBonus != 0.15 {
		t.Errorf("Expected XP bonus 0.15, got %v", xpBonus)
	}

	// Clear territory bonus provider
	hud.SetTerritoryBonusProvider(nil)
	if hud.territoryBonusProvider != nil {
		t.Error("Expected territory bonus provider to be cleared")
	}
}

// TestTerritoryBonusesDisplay verifies territory bonuses are only shown with guild membership.
func TestTerritoryBonusesDisplay(t *testing.T) {
	hud := NewEbitenHUDSystem(800, 600)

	// Create a player entity
	world := NewWorld()
	player := world.CreateEntity()
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	hud.SetPlayerEntity(player)

	// Without territory provider, drawTerritoryBonuses should be safe to call
	hud.drawTerritoryBonuses() // Should not panic

	// With provider but no guild component
	provider := &mockTerritoryBonusProvider{
		resourceBonus: 0.10,
		xpBonus:       0.15,
	}
	hud.SetTerritoryBonusProvider(provider)
	hud.drawTerritoryBonuses() // Should not panic (no guild)

	// Add guild component
	player.AddComponent(&GuildComponent{GuildID: "test_guild"})
	// Note: We can't actually test rendering without Ebiten initialization,
	// but we can verify the method doesn't panic
	hud.drawTerritoryBonuses() // Should not panic
}

// TestTerritoryBonusesDisplay_NoBonuses verifies no display when bonuses are zero.
func TestTerritoryBonusesDisplay_NoBonuses(t *testing.T) {
	hud := NewEbitenHUDSystem(800, 600)

	// Create a player entity with guild
	world := NewWorld()
	player := world.CreateEntity()
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&GuildComponent{GuildID: "test_guild"})
	hud.SetPlayerEntity(player)

	// Provider with zero bonuses
	provider := &mockTerritoryBonusProvider{
		resourceBonus: 0,
		xpBonus:       0,
	}
	hud.SetTerritoryBonusProvider(provider)

	// Should safely return (no bonuses to display)
	hud.drawTerritoryBonuses() // Should not panic
}

// TestTerritoryBonusesDisplay_EmptyGuildID verifies no display for empty guild.
func TestTerritoryBonusesDisplay_EmptyGuildID(t *testing.T) {
	hud := NewEbitenHUDSystem(800, 600)

	// Create a player entity with empty guild ID
	world := NewWorld()
	player := world.CreateEntity()
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&GuildComponent{GuildID: ""})
	hud.SetPlayerEntity(player)

	provider := &mockTerritoryBonusProvider{
		resourceBonus: 0.10,
		xpBonus:       0.15,
	}
	hud.SetTerritoryBonusProvider(provider)

	// Should safely return (empty guild ID)
	hud.drawTerritoryBonuses() // Should not panic
}
