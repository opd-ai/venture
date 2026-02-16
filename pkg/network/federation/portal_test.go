package federation

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// testPlayerComponent is a minimal component marking an entity as a player for tests.
type testPlayerComponent struct{}

func (t *testPlayerComponent) Type() string { return "player" }

func TestNewPortalSystem(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)

	if ps == nil {
		t.Fatal("NewPortalSystem() returned nil")
	}

	if ps.world != world {
		t.Error("world not set correctly")
	}

	if ps.federation != federation {
		t.Error("federation not set correctly")
	}
}

func TestPortalSystem_LocalTeleport(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()

	// Create player
	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})

	// Create local portal
	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer: "local",
		DestinationX:      500,
		DestinationY:      600,
	})
	portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process entity addition

	// Activate portal
	err := ps.ActivatePortal(player.ID, portal.ID, authMgr, transferMgr)
	if err != nil {
		t.Fatalf("ActivatePortal() error = %v", err)
	}

	// Verify player position changed
	posCompRaw, _ := player.GetComponent("position")
	posComp := posCompRaw.(*engine.PositionComponent)

	if posComp.X != 500 || posComp.Y != 600 {
		t.Errorf("Position = (%v, %v), want (500, 600)", posComp.X, posComp.Y)
	}
}

func TestPortalSystem_LocalTeleport_NoPosition(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()

	// Create player without position component
	player := world.CreateEntity()

	// Create portal
	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer: "local",
		DestinationX:      500,
		DestinationY:      600,
	})
	portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process entity addition

	// Should fail
	err := ps.ActivatePortal(player.ID, portal.ID, authMgr, transferMgr)
	if err == nil {
		t.Error("Expected error for player without position, got nil")
	}
}

func TestPortalSystem_CrossServerTransfer(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()

	// Create player with all required components
	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))
	world.Update(0) // Process entity addition

	// Create cross-server portal
	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer: "server-2",
		DestinationX:      500,
		DestinationY:      600,
	})
	portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process entity addition

	// Activate portal
	err := ps.ActivatePortal(player.ID, portal.ID, authMgr, transferMgr)
	if err != nil {
		t.Fatalf("ActivatePortal() error = %v", err)
	}

	// Verify transfer was prepared
	transfer, exists := transferMgr.GetTransfer(player.ID)
	if !exists {
		t.Error("Transfer not found")
	}

	if transfer.TargetID != "server-2" {
		t.Errorf("TargetID = %v, want server-2", transfer.TargetID)
	}
}

func TestPortalSystem_RequiredItem_Present(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()

	// Create player with key item
	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	player.AddComponent(&engine.HealthComponent{Current: 80, Max: 100})

	inv := engine.NewInventoryComponent(20, 100.0)
	keyItem := &item.Item{ID: "key-1", Name: "Portal Key"}
	inv.AddItem(keyItem)
	player.AddComponent(inv)

	// Create portal requiring key
	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer: "local",
		DestinationX:      500,
		DestinationY:      600,
		RequiredItem:      "Portal Key",
	})
	portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process entity addition

	// Should succeed
	err := ps.ActivatePortal(player.ID, portal.ID, authMgr, transferMgr)
	if err != nil {
		t.Errorf("ActivatePortal() error = %v", err)
	}
}

func TestPortalSystem_RequiredItem_Missing(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()

	// Create player without key item
	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	player.AddComponent(engine.NewInventoryComponent(20, 100.0))

	// Create portal requiring key
	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer: "local",
		DestinationX:      500,
		DestinationY:      600,
		RequiredItem:      "Portal Key",
	})
	portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process entity addition

	// Should fail
	err := ps.ActivatePortal(player.ID, portal.ID, authMgr, transferMgr)
	if err == nil {
		t.Error("Expected error for missing required item, got nil")
	}
}

func TestPortalSystem_PortalNotFound(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()

	player := world.CreateEntity()
	world.Update(0) // Process entity addition

	err := ps.ActivatePortal(player.ID, 99999, authMgr, transferMgr)
	if err == nil {
		t.Error("Expected error for portal not found, got nil")
	}
}

func TestPortalSystem_NotAPortal(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()

	player := world.CreateEntity()

	// Create entity without portal component
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	err := ps.ActivatePortal(player.ID, entity.ID, authMgr, transferMgr)
	if err == nil {
		t.Error("Expected error for non-portal entity, got nil")
	}
}

func TestPortalSystem_Update(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)

	// Create portal
	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer: "local",
		DestinationX:      500,
		DestinationY:      600,
	})
	portal.AddComponent(&engine.PositionComponent{X: 300, Y: 300})
	world.Update(0) // Process entity addition

	// Update should not crash
	ps.Update(0.016)
}

func TestPortalSystem_SetManagers(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)

	if ps.authMgr != nil || ps.transferMgr != nil {
		t.Error("Managers should be nil initially")
	}

	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()
	ps.SetManagers(authMgr, transferMgr)

	if ps.authMgr != authMgr {
		t.Error("authMgr not set correctly")
	}
	if ps.transferMgr != transferMgr {
		t.Error("transferMgr not set correctly")
	}
}

func TestPortalSystem_Update_AutoActivation(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()
	ps.SetManagers(authMgr, transferMgr)

	// Create player near portal
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})

	// Create auto-activate local portal at same position
	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer:  "local",
		DestinationX:       500,
		DestinationY:       600,
		RequiresActivation: false,
	})
	portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	world.Update(0)

	// Update should auto-activate and teleport player
	ps.Update(0.016)

	posRaw, _ := player.GetComponent("position")
	pos := posRaw.(*engine.PositionComponent)
	if pos.X != 500 || pos.Y != 600 {
		t.Errorf("Player should have teleported to (500, 600), got (%v, %v)", pos.X, pos.Y)
	}
}

func TestPortalSystem_Update_RequiresActivation(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()
	ps.SetManagers(authMgr, transferMgr)

	// Create player near portal
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})

	// Create portal that requires manual activation
	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer:  "local",
		DestinationX:       500,
		DestinationY:       600,
		RequiresActivation: true,
	})
	portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	world.Update(0)

	// Update should NOT auto-activate
	ps.Update(0.016)

	posRaw, _ := player.GetComponent("position")
	pos := posRaw.(*engine.PositionComponent)
	if pos.X != 100 || pos.Y != 100 {
		t.Errorf("Player should NOT have teleported, got (%v, %v)", pos.X, pos.Y)
	}
}

func TestPortalSystem_Update_NoManagers(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	// Do NOT call SetManagers

	// Create player near portal
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})

	// Create auto-activate portal at same position
	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer:  "local",
		DestinationX:       500,
		DestinationY:       600,
		RequiresActivation: false,
	})
	portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	world.Update(0)

	// Update should skip activation due to nil managers (no crash)
	ps.Update(0.016)

	posRaw, _ := player.GetComponent("position")
	pos := posRaw.(*engine.PositionComponent)
	if pos.X != 100 || pos.Y != 100 {
		t.Errorf("Player should NOT have teleported without managers, got (%v, %v)", pos.X, pos.Y)
	}
}

func TestPortalSystem_Update_OutOfRange(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()
	ps.SetManagers(authMgr, transferMgr)

	// Create player far from portal
	player := world.CreateEntity()
	player.AddComponent(&testPlayerComponent{})
	player.AddComponent(&engine.PositionComponent{X: 0, Y: 0})

	// Create portal far away
	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer:  "local",
		DestinationX:       500,
		DestinationY:       600,
		RequiresActivation: false,
	})
	portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	world.Update(0)

	// Update should NOT activate - player is too far
	ps.Update(0.016)

	posRaw, _ := player.GetComponent("position")
	pos := posRaw.(*engine.PositionComponent)
	if pos.X != 0 || pos.Y != 0 {
		t.Errorf("Player should NOT have teleported from out of range, got (%v, %v)", pos.X, pos.Y)
	}
}

func TestCheckRequiredItem_NoInventory(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)

	player := world.CreateEntity()

	err := ps.checkRequiredItem(player.ID, "Portal Key")
	if err == nil {
		t.Error("Expected error for player without inventory, got nil")
	}
}

func TestCheckRequiredItem_PlayerNotFound(t *testing.T) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)

	err := ps.checkRequiredItem(99999, "Portal Key")
	if err == nil {
		t.Error("Expected error for player not found, got nil")
	}
}

// Benchmarks

func BenchmarkPortalSystem_LocalTeleport(b *testing.B) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)
	authMgr := NewAuthManager()
	transferMgr := NewTransferManager()

	player := world.CreateEntity()
	player.AddComponent(&engine.PositionComponent{X: 100, Y: 100})

	portal := world.CreateEntity()
	portal.AddComponent(&engine.PortalComponent{
		DestinationServer: "local",
		DestinationX:      500,
		DestinationY:      600,
	})
	portal.AddComponent(&engine.PositionComponent{X: 100, Y: 100})
	world.Update(0) // Process entity addition

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.ActivatePortal(player.ID, portal.ID, authMgr, transferMgr)
	}
}

func BenchmarkPortalSystem_CheckRequiredItem(b *testing.B) {
	world := engine.NewWorld()
	federation := NewFederationProtocol("server-1", testIdentity())
	ps := NewPortalSystem(world, federation)

	player := world.CreateEntity()
	inv := engine.NewInventoryComponent(20, 100.0)
	inv.AddItem(&item.Item{ID: "key-1", Name: "Portal Key"})
	inv.AddItem(&item.Item{ID: "sword-1", Name: "Sword"})
	inv.AddItem(&item.Item{ID: "potion-1", Name: "Potion"})
	player.AddComponent(inv)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ps.checkRequiredItem(player.ID, "Portal Key")
	}
}
