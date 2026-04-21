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

// TestVehicleSyncer_StableFormationSlot verifies that formation slots are stable
// after a vehicle is removed and a new vehicle is added.
// Prior to this fix, len(fleet.Vehicles)-1 produced duplicate/shifting indices.
func TestVehicleSyncer_StableFormationSlot(t *testing.T) {
	m := NewFleetManager()
	syncer := newStubVehicleSyncer()
	m.SetVehicleSyncer(syncer)

	// Add three vehicles; expected slots 0, 1, 2
	for _, id := range []uint64{10, 11, 12} {
		if err := m.AddVehicle("g", id, "f"); err != nil {
			t.Fatalf("AddVehicle %d: %v", id, err)
		}
	}
	slot10 := syncer.synced[10].FormationPosition
	slot11 := syncer.synced[11].FormationPosition
	slot12 := syncer.synced[12].FormationPosition
	if slot10 != 0 || slot11 != 1 || slot12 != 2 {
		t.Fatalf("initial slots = %d,%d,%d; want 0,1,2", slot10, slot11, slot12)
	}

	// Remove vehicle 11; add vehicle 13
	if err := m.RemoveVehicle("g", 11, "f"); err != nil {
		t.Fatalf("RemoveVehicle: %v", err)
	}
	if err := m.AddVehicle("g", 13, "f"); err != nil {
		t.Fatalf("AddVehicle 13: %v", err)
	}

	// Slot for 13 must be 3 (not 1, which was vehicle 11's old slot)
	slot13 := syncer.synced[13].FormationPosition
	if slot13 != 3 {
		t.Errorf("slot for vehicle 13 = %d; want 3 (stable, non-reused)", slot13)
	}

	// Slots for 10 and 12 must remain unchanged
	if syncer.synced[10].FormationPosition != 0 {
		t.Errorf("slot for vehicle 10 changed to %d after re-add; want 0", syncer.synced[10].FormationPosition)
	}
	if syncer.synced[12].FormationPosition != 2 {
		t.Errorf("slot for vehicle 12 changed to %d after re-add; want 2", syncer.synced[12].FormationPosition)
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

// TestGetFormationOffsets_SingleVehicle verifies leader-only fleet returns one zero offset.
func TestGetFormationOffsets_SingleVehicle(t *testing.T) {
	m := NewFleetManager()
	if err := m.CreateFleet("g", "f", "cmd"); err != nil {
		t.Fatalf("CreateFleet: %v", err)
	}
	if err := m.AddVehicle("g", 1, "f"); err != nil {
		t.Fatalf("AddVehicle: %v", err)
	}
	_ = m.SetFormation("g", "f", FormationLine)

	offsets := m.GetFormationOffsets("g", "f")
	if len(offsets) != 1 {
		t.Fatalf("len(offsets) = %d; want 1", len(offsets))
	}
	if offsets[0].OffsetX != 0 || offsets[0].OffsetY != 0 {
		t.Errorf("single-vehicle offset = (%f,%f); want (0,0)", offsets[0].OffsetX, offsets[0].OffsetY)
	}
}

// TestGetFormationOffsets_TwoVehicles verifies minimal follower placement.
func TestGetFormationOffsets_TwoVehicles(t *testing.T) {
	formations := []FormationType{FormationLine, FormationWedge, FormationColumn, FormationCircle, FormationNone}
	for _, f := range formations {
		m := NewFleetManager()
		_ = m.CreateFleet("g", "f", "cmd")
		for i := uint64(1); i <= 2; i++ {
			_ = m.AddVehicle("g", i, "f")
		}
		_ = m.SetFormation("g", "f", f)

		offsets := m.GetFormationOffsets("g", "f")
		if len(offsets) != 2 {
			t.Errorf("formation %v: len(offsets) = %d; want 2", f, len(offsets))
			continue
		}
		if offsets[0].OffsetX != 0 || offsets[0].OffsetY != 0 {
			t.Errorf("formation %v: leader offset = (%f,%f); want (0,0)", f, offsets[0].OffsetX, offsets[0].OffsetY)
		}
	}
}

// TestGetFormationOffsets_LargeFleet verifies geometry scales for 8 vehicles.
func TestGetFormationOffsets_LargeFleet(t *testing.T) {
	const count = 8
	formations := []FormationType{FormationLine, FormationWedge, FormationColumn, FormationCircle}
	for _, f := range formations {
		m := NewFleetManager()
		_ = m.CreateFleet("g", "f", "cmd")
		for i := uint64(1); i <= count; i++ {
			_ = m.AddVehicle("g", i, "f")
		}
		_ = m.SetFormation("g", "f", f)

		offsets := m.GetFormationOffsets("g", "f")
		if len(offsets) != count {
			t.Errorf("formation %v: len(offsets) = %d; want %d", f, len(offsets), count)
			continue
		}
		for i, o := range offsets {
			if o.SlotIndex != i {
				t.Errorf("formation %v: offsets[%d].SlotIndex = %d; want %d", f, i, o.SlotIndex, i)
			}
		}
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

	const weaponDamage = 100.0
	if err := m.ApplySiegeDamage(77, "territory1", "wall1", weaponDamage); err != nil {
		t.Fatalf("ApplySiegeDamage: %v", err)
	}

	if len(damager.calls) != 1 {
		t.Fatalf("expected 1 damage call, got %d", len(damager.calls))
	}
	call := damager.calls[0]
	want := weaponDamage * SiegeCatapult.GetSiegeDamageMultiplier()
	if call.damage != want {
		t.Errorf("damage = %f; want %f (catapult multiplier)", call.damage, want)
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
