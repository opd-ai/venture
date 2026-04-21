package guild_vehicle

import (
	"fmt"
	"testing"
)

// stubMembershipValidator is a test double for MembershipValidator.
type stubMembershipValidator struct {
	members map[string]map[string]bool // guildID -> playerID -> bool
}

func newStubMembershipValidator() *stubMembershipValidator {
	return &stubMembershipValidator{members: make(map[string]map[string]bool)}
}

func (s *stubMembershipValidator) addMember(guildID, playerID string) {
	if s.members[guildID] == nil {
		s.members[guildID] = make(map[string]bool)
	}
	s.members[guildID][playerID] = true
}

func (s *stubMembershipValidator) IsMember(guildID, playerID string) bool {
	return s.members[guildID][playerID]
}

// stubVehicleSyncer records sync calls for assertion.
type stubVehicleSyncer struct {
	synced  map[uint64]*GuildVehicleFleetComponent
	cleared map[uint64]bool
}

func newStubVehicleSyncer() *stubVehicleSyncer {
	return &stubVehicleSyncer{
		synced:  make(map[uint64]*GuildVehicleFleetComponent),
		cleared: make(map[uint64]bool),
	}
}

func (s *stubVehicleSyncer) SyncVehicleFleetComponent(vehicleID uint64, comp *GuildVehicleFleetComponent) {
	s.synced[vehicleID] = comp
}

func (s *stubVehicleSyncer) ClearVehicleFleetComponent(vehicleID uint64) {
	s.cleared[vehicleID] = true
	delete(s.synced, vehicleID)
}

// stubStructureDamager records damage calls for assertion.
type stubStructureDamager struct {
	calls []damageCall
}

type damageCall struct {
	territoryID string
	structureID string
	damage      float64
}

func (s *stubStructureDamager) DamageStructure(territoryID, structureID string, damage float64) error {
	s.calls = append(s.calls, damageCall{territoryID, structureID, damage})
	return nil
}

// TestGrantAccess_MembershipValidation ensures non-members are rejected.
func TestGrantAccess_MembershipValidation(t *testing.T) {
	m := NewFleetManager()
	validator := newStubMembershipValidator()
	validator.addMember("guild1", "player1")
	m.SetMembershipValidator(validator)

	// Create fleet and add vehicle
	if err := m.AddVehicle("guild1", 100, "fleet1"); err != nil {
		t.Fatalf("AddVehicle: %v", err)
	}

	// Non-member must be rejected
	if err := m.GrantAccess("guild1", 100, "nonmember"); err == nil {
		t.Error("expected error for non-member, got nil")
	}

	// Guild member must be accepted
	if err := m.GrantAccess("guild1", 100, "player1"); err != nil {
		t.Errorf("GrantAccess for member: %v", err)
	}

	if !m.CheckAccess("guild1", 100, "player1") {
		t.Error("member should have access after GrantAccess")
	}
}

// TestGrantAccess_NoValidator allows all players when validator is nil.
func TestGrantAccess_NoValidator(t *testing.T) {
	m := NewFleetManager()
	if err := m.AddVehicle("guild1", 100, "fleet1"); err != nil {
		t.Fatalf("AddVehicle: %v", err)
	}
	if err := m.GrantAccess("guild1", 100, "anyone"); err != nil {
		t.Errorf("GrantAccess without validator: %v", err)
	}
}

// TestVehicleSyncer_AddAndRemove verifies syncer is called on add and remove.
func TestVehicleSyncer_AddAndRemove(t *testing.T) {
	m := NewFleetManager()
	syncer := newStubVehicleSyncer()
	m.SetVehicleSyncer(syncer)

	if err := m.AddVehicle("guild1", 42, "fleet1"); err != nil {
		t.Fatalf("AddVehicle: %v", err)
	}

	comp, ok := syncer.synced[42]
	if !ok {
		t.Fatal("syncer was not called for vehicle 42")
	}
	if comp.GuildID != "guild1" || comp.FleetID != "fleet1" {
		t.Errorf("comp = %+v; want GuildID=guild1 FleetID=fleet1", comp)
	}

	if err := m.RemoveVehicle("guild1", 42, "fleet1"); err != nil {
		t.Fatalf("RemoveVehicle: %v", err)
	}

	if !syncer.cleared[42] {
		t.Error("syncer.ClearVehicleFleetComponent was not called for vehicle 42")
	}
}

// TestVehicleSyncer_SiegeType verifies siege type is propagated in component.
func TestVehicleSyncer_SiegeType(t *testing.T) {
	m := NewFleetManager()
	syncer := newStubVehicleSyncer()
	m.SetVehicleSyncer(syncer)

	if err := m.AddVehicleWithType("guild1", 99, "siege_fleet", SiegeCatapult, 200); err != nil {
		t.Fatalf("AddVehicleWithType: %v", err)
	}

	comp, ok := syncer.synced[99]
	if !ok {
		t.Fatal("syncer was not called")
	}
	if comp.SiegeType != SiegeCatapult {
		t.Errorf("SiegeType = %v; want SiegeCatapult", comp.SiegeType)
	}
}

// TestGetFormationOffsets_Line checks that line formation spreads laterally.
func TestGetFormationOffsets_Line(t *testing.T) {
	m := NewFleetManager()
	if err := m.CreateFleet("g", "f", "cmd"); err != nil {
		t.Fatalf("CreateFleet: %v", err)
	}
	for i := uint64(1); i <= 3; i++ {
		if err := m.AddVehicle("g", i, "f"); err != nil {
			t.Fatalf("AddVehicle %d: %v", i, err)
		}
	}
	if err := m.SetFormation("g", "f", FormationLine); err != nil {
		t.Fatalf("SetFormation: %v", err)
	}

	offsets := m.GetFormationOffsets("g", "f")
	if len(offsets) != 3 {
		t.Fatalf("len(offsets) = %d; want 3", len(offsets))
	}

	// Leader slot must be at origin
	if offsets[0].OffsetX != 0 || offsets[0].OffsetY != 0 {
		t.Errorf("slot 0 offset = (%f,%f); want (0,0)", offsets[0].OffsetX, offsets[0].OffsetY)
	}

	// Non-leader slots must have non-zero X offset (lateral spreading)
	for _, o := range offsets[1:] {
		if o.OffsetY != 0 {
			t.Errorf("line formation: slot %d has OffsetY=%f; want 0", o.SlotIndex, o.OffsetY)
		}
		if o.OffsetX == 0 {
			t.Errorf("line formation: slot %d has OffsetX=0; want non-zero", o.SlotIndex)
		}
	}
}

// TestGetFormationOffsets_Column checks single-file behind leader.
func TestGetFormationOffsets_Column(t *testing.T) {
	m := NewFleetManager()
	if err := m.CreateFleet("g", "f", "cmd"); err != nil {
		t.Fatalf("CreateFleet: %v", err)
	}
	for i := uint64(1); i <= 3; i++ {
		if err := m.AddVehicle("g", i, "f"); err != nil {
			t.Fatalf("AddVehicle %d: %v", i, err)
		}
	}
	_ = m.SetFormation("g", "f", FormationColumn)

	offsets := m.GetFormationOffsets("g", "f")
	for _, o := range offsets[1:] {
		if o.OffsetX != 0 {
			t.Errorf("column slot %d: OffsetX=%f; want 0", o.SlotIndex, o.OffsetX)
		}
		if o.OffsetY >= 0 {
			t.Errorf("column slot %d: OffsetY=%f; want negative", o.SlotIndex, o.OffsetY)
		}
	}
}

// TestGetFormationOffsets_NonExistentFleet returns nil.
func TestGetFormationOffsets_NonExistentFleet(t *testing.T) {
	m := NewFleetManager()
	if offsets := m.GetFormationOffsets("nobody", "nofleet"); offsets != nil {
		t.Errorf("expected nil for missing fleet, got %v", offsets)
	}
}

// TestApplySiegeDamage_Basic verifies damage multiplier is applied.
func TestApplySiegeDamage_Basic(t *testing.T) {
	m := NewFleetManager()
	damager := &stubStructureDamager{}
	m.SetStructureDamager(damager)

	if err := m.AddVehicleWithType("guild1", 77, "fleet1", SiegeCatapult, 100); err != nil {
		t.Fatalf("AddVehicleWithType: %v", err)
	}

	if err := m.ApplySiegeDamage(77, "territory1", "wall1", 100.0); err != nil {
		t.Fatalf("ApplySiegeDamage: %v", err)
	}

	if len(damager.calls) != 1 {
		t.Fatalf("expected 1 damage call, got %d", len(damager.calls))
	}
	call := damager.calls[0]
	// SiegeCatapult multiplier = 5.0
	want := 100.0 * 5.0
	if call.damage != want {
		t.Errorf("damage = %f; want %f (catapult x5)", call.damage, want)
	}
	if call.territoryID != "territory1" || call.structureID != "wall1" {
		t.Errorf("wrong target: got %s/%s", call.territoryID, call.structureID)
	}
}

// TestApplySiegeDamage_NoVehicle returns error.
func TestApplySiegeDamage_NoVehicle(t *testing.T) {
	m := NewFleetManager()
	m.SetStructureDamager(&stubStructureDamager{})
	err := m.ApplySiegeDamage(999, "t", "s", 10)
	if err == nil {
		t.Error("expected error for unknown vehicle, got nil")
	}
}

// TestApplySiegeDamage_NoDamager returns error.
func TestApplySiegeDamage_NoDamager(t *testing.T) {
	m := NewFleetManager()
	if err := m.AddVehicle("g", 1, "f"); err != nil {
		t.Fatalf("AddVehicle: %v", err)
	}
	err := m.ApplySiegeDamage(1, "t", "s", 10)
	if err == nil {
		t.Error("expected error when no StructureDamager configured, got nil")
	}
}

// TestApplySiegeDamage_DamagerError propagates errors.
func TestApplySiegeDamage_DamagerError(t *testing.T) {
	m := NewFleetManager()
	errDamager := &errStructureDamager{}
	m.SetStructureDamager(errDamager)
	if err := m.AddVehicle("g", 5, "f"); err != nil {
		t.Fatalf("AddVehicle: %v", err)
	}
	if err := m.ApplySiegeDamage(5, "t", "s", 10); err == nil {
		t.Error("expected error from damager, got nil")
	}
}

type errStructureDamager struct{}

func (e *errStructureDamager) DamageStructure(_, _ string, _ float64) error {
	return fmt.Errorf("territory system offline")
}

// TestSetters_NilSafe verifies passing nil to setters does not panic.
func TestSetters_NilSafe(t *testing.T) {
	m := NewFleetManager()
	m.SetMembershipValidator(nil)
	m.SetVehicleSyncer(nil)
	m.SetStructureDamager(nil)

	// After nil setters, normal operations should work fine
	if err := m.AddVehicle("g", 1, "f"); err != nil {
		t.Errorf("AddVehicle after nil setters: %v", err)
	}
	if err := m.GrantAccess("g", 1, "player1"); err != nil {
		t.Errorf("GrantAccess after nil validator: %v", err)
	}
}
