package trade

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestValidateTrust_LowTrust tests trust validation for low-trust scenarios.
func TestValidateTrust_LowTrust(t *testing.T) {
	world := engine.NewWorld()
	sys := NewTradeSystem(world)

	tests := []struct {
		name        string
		trustScore  float64
		items       []*item.Item
		wantErr     bool
		errContains string
	}{
		{
			name:       "low trust - max items exceeded",
			trustScore: 0.2, // Below lowTrustThreshold (0.3)
			items: []*item.Item{
				{Name: "item1", Rarity: item.RarityCommon},
				{Name: "item2", Rarity: item.RarityCommon},
				{Name: "item3", Rarity: item.RarityCommon},
				{Name: "item4", Rarity: item.RarityCommon},
				{Name: "item5", Rarity: item.RarityCommon},
				{Name: "item6", Rarity: item.RarityCommon}, // 6 items, over limit of 5
			},
			wantErr:     true,
			errContains: "low trust players can trade max",
		},
		{
			name:       "low trust - rare item forbidden",
			trustScore: 0.25,
			items: []*item.Item{
				{Name: "rare_sword", Rarity: item.RarityRare},
			},
			wantErr:     true,
			errContains: "low trust players cannot trade rare",
		},
		{
			name:       "low trust - epic item forbidden",
			trustScore: 0.1,
			items: []*item.Item{
				{Name: "epic_armor", Rarity: item.RarityEpic},
			},
			wantErr:     true,
			errContains: "low trust players cannot trade rare",
		},
		{
			name:       "low trust - legendary item forbidden",
			trustScore: 0.29,
			items: []*item.Item{
				{Name: "legendary_axe", Rarity: item.RarityLegendary},
			},
			wantErr:     true,
			errContains: "low trust players cannot trade rare",
		},
		{
			name:       "low trust - within limits common items",
			trustScore: 0.25,
			items: []*item.Item{
				{Name: "item1", Rarity: item.RarityCommon},
				{Name: "item2", Rarity: item.RarityCommon},
				{Name: "item3", Rarity: item.RarityCommon},
			},
			wantErr: false,
		},
		{
			name:       "low trust - within limits uncommon items",
			trustScore: 0.15,
			items: []*item.Item{
				{Name: "item1", Rarity: item.RarityUncommon},
				{Name: "item2", Rarity: item.RarityUncommon},
			},
			wantErr: false,
		},
		{
			name:       "low trust - exactly 5 items allowed",
			trustScore: 0.2,
			items: []*item.Item{
				{Name: "item1", Rarity: item.RarityCommon},
				{Name: "item2", Rarity: item.RarityCommon},
				{Name: "item3", Rarity: item.RarityCommon},
				{Name: "item4", Rarity: item.RarityCommon},
				{Name: "item5", Rarity: item.RarityCommon},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sys.validateTrust(tt.trustScore, tt.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTrust() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("validateTrust() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

// TestValidateTrust_MediumTrust tests trust validation for medium-trust scenarios.
func TestValidateTrust_MediumTrust(t *testing.T) {
	world := engine.NewWorld()
	sys := NewTradeSystem(world)

	tests := []struct {
		name        string
		trustScore  float64
		items       []*item.Item
		wantErr     bool
		errContains string
	}{
		{
			name:       "medium trust - legendary item forbidden",
			trustScore: 0.5, // Above lowTrustThreshold (0.3), below highTrustThreshold (0.7)
			items: []*item.Item{
				{Name: "legendary_weapon", Rarity: item.RarityLegendary},
			},
			wantErr:     true,
			errContains: "too low for legendary items",
		},
		{
			name:       "medium trust - rare item allowed",
			trustScore: 0.5,
			items: []*item.Item{
				{Name: "rare_item", Rarity: item.RarityRare},
			},
			wantErr: false,
		},
		{
			name:       "medium trust - epic item allowed",
			trustScore: 0.6,
			items: []*item.Item{
				{Name: "epic_item", Rarity: item.RarityEpic},
			},
			wantErr: false,
		},
		{
			name:       "medium trust - many common/uncommon items allowed",
			trustScore: 0.4,
			items: []*item.Item{
				{Name: "item1", Rarity: item.RarityCommon},
				{Name: "item2", Rarity: item.RarityUncommon},
				{Name: "item3", Rarity: item.RarityCommon},
				{Name: "item4", Rarity: item.RarityUncommon},
				{Name: "item5", Rarity: item.RarityCommon},
				{Name: "item6", Rarity: item.RarityCommon},
				{Name: "item7", Rarity: item.RarityCommon},
			},
			wantErr: false,
		},
		{
			name:       "medium trust - mix of rare and epic allowed",
			trustScore: 0.65,
			items: []*item.Item{
				{Name: "rare1", Rarity: item.RarityRare},
				{Name: "epic1", Rarity: item.RarityEpic},
				{Name: "rare2", Rarity: item.RarityRare},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sys.validateTrust(tt.trustScore, tt.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTrust() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errContains != "" {
				if !contains(err.Error(), tt.errContains) {
					t.Errorf("validateTrust() error = %v, want error containing %q", err, tt.errContains)
				}
			}
		})
	}
}

// TestValidateTrust_HighTrust tests trust validation for high-trust scenarios.
func TestValidateTrust_HighTrust(t *testing.T) {
	world := engine.NewWorld()
	sys := NewTradeSystem(world)

	tests := []struct {
		name       string
		trustScore float64
		items      []*item.Item
		wantErr    bool
	}{
		{
			name:       "high trust - legendary item allowed",
			trustScore: 0.80, // At highTrustThreshold (0.8)
			items: []*item.Item{
				{Name: "legendary_weapon", Rarity: item.RarityLegendary},
			},
			wantErr: false,
		},
		{
			name:       "high trust - multiple legendaries allowed",
			trustScore: 0.9,
			items: []*item.Item{
				{Name: "legendary1", Rarity: item.RarityLegendary},
				{Name: "legendary2", Rarity: item.RarityLegendary},
				{Name: "legendary3", Rarity: item.RarityLegendary},
			},
			wantErr: false,
		},
		{
			name:       "high trust - mix of all rarities allowed",
			trustScore: 1.0,
			items: []*item.Item{
				{Name: "common", Rarity: item.RarityCommon},
				{Name: "uncommon", Rarity: item.RarityUncommon},
				{Name: "rare", Rarity: item.RarityRare},
				{Name: "epic", Rarity: item.RarityEpic},
				{Name: "legendary", Rarity: item.RarityLegendary},
			},
			wantErr: false,
		},
		{
			name:       "high trust - many items allowed",
			trustScore: 0.8,
			items:      make([]*item.Item, 20), // 20 items
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Initialize items if needed
			for i := range tt.items {
				if tt.items[i] == nil {
					tt.items[i] = &item.Item{Name: "item", Rarity: item.RarityCommon}
				}
			}

			err := sys.validateTrust(tt.trustScore, tt.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTrust() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateTrust_EdgeCases tests edge cases for trust validation.
func TestValidateTrust_EdgeCases(t *testing.T) {
	world := engine.NewWorld()
	sys := NewTradeSystem(world)

	tests := []struct {
		name       string
		trustScore float64
		items      []*item.Item
		wantErr    bool
	}{
		{
			name:       "exactly at low trust threshold",
			trustScore: 0.3, // lowTrustThreshold
			items: []*item.Item{
				{Name: "rare_item", Rarity: item.RarityRare},
			},
			wantErr: false, // At threshold, considered medium trust
		},
		{
			name:       "exactly at high trust threshold",
			trustScore: 0.8, // highTrustThreshold
			items: []*item.Item{
				{Name: "legendary_item", Rarity: item.RarityLegendary},
			},
			wantErr: false, // At threshold, considered high trust
		},
		{
			name:       "zero trust score",
			trustScore: 0.0,
			items: []*item.Item{
				{Name: "common_item", Rarity: item.RarityCommon},
			},
			wantErr: false,
		},
		{
			name:       "empty item list",
			trustScore: 0.1,
			items:      []*item.Item{},
			wantErr:    false,
		},
		{
			name:       "nil item list handled gracefully",
			trustScore: 0.5,
			items:      nil,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sys.validateTrust(tt.trustScore, tt.items)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTrust() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestTransferTracker_RollbackFunc tests the rollback functionality.
func TestTransferTracker_RollbackFunc(t *testing.T) {
	// Create test inventories with large capacity
	proposerInv := &engine.InventoryComponent{
		Items:     []*item.Item{},
		MaxItems:  100,
		MaxWeight: 1000.0,
	}
	recipientInv := &engine.InventoryComponent{
		Items:     []*item.Item{},
		MaxItems:  100,
		MaxWeight: 1000.0,
	}

	// Add initial items to inventories
	initialProposerItem := &item.Item{ID: "proposer-initial", Name: "Proposer Initial"}
	initialRecipientItem := &item.Item{ID: "recipient-initial", Name: "Recipient Initial"}
	proposerInv.Items = append(proposerInv.Items, initialProposerItem)
	recipientInv.Items = append(recipientInv.Items, initialRecipientItem)

	// Create tracker and simulate transfers
	tracker := newTransferTracker()

	// Simulate removing items from proposer
	removedFromProposer := &item.Item{ID: "proposer-removed", Name: "Proposer Removed"}
	tracker.removedFromProposer = append(tracker.removedFromProposer, removedFromProposer)

	// Simulate removing items from recipient
	removedFromRecipient := &item.Item{ID: "recipient-removed", Name: "Recipient Removed"}
	tracker.removedFromRecipient = append(tracker.removedFromRecipient, removedFromRecipient)

	// Simulate adding items to recipient
	addedToRecipient := &item.Item{ID: "added-to-recipient", Name: "Added to Recipient"}
	recipientInv.Items = append(recipientInv.Items, addedToRecipient)
	tracker.addedToRecipient = append(tracker.addedToRecipient, addedToRecipient)

	// Simulate adding items to proposer
	addedToProposer := &item.Item{ID: "added-to-proposer", Name: "Added to Proposer"}
	proposerInv.Items = append(proposerInv.Items, addedToProposer)
	tracker.addedToProposer = append(tracker.addedToProposer, addedToProposer)

	// Verify pre-rollback state
	if len(proposerInv.Items) != 2 { // initial + added
		t.Errorf("Before rollback: proposer inventory = %d items, want 2", len(proposerInv.Items))
	}
	if len(recipientInv.Items) != 2 { // initial + added
		t.Errorf("Before rollback: recipient inventory = %d items, want 2", len(recipientInv.Items))
	}

	// Create and execute rollback function
	rollback := tracker.createRollbackFunc(proposerInv, recipientInv)
	rollback()

	// Verify rollback restored items
	if len(proposerInv.Items) != 2 { // initial + removed (restored)
		t.Errorf("After rollback: proposer inventory = %d items, want 2", len(proposerInv.Items))
	}
	if len(recipientInv.Items) != 2 { // initial + removed (restored)
		t.Errorf("After rollback: recipient inventory = %d items, want 2", len(recipientInv.Items))
	}

	// Verify specific items are present/absent
	hasProposerRemoved := false
	hasProposerAdded := false
	for _, itm := range proposerInv.Items {
		if itm.ID == "proposer-removed" {
			hasProposerRemoved = true
		}
		if itm.ID == "added-to-proposer" {
			hasProposerAdded = true
		}
	}

	if !hasProposerRemoved {
		t.Error("Rollback should restore removed items to proposer inventory")
	}
	if hasProposerAdded {
		t.Error("Rollback should remove added items from proposer inventory")
	}

	hasRecipientRemoved := false
	hasRecipientAdded := false
	for _, itm := range recipientInv.Items {
		if itm.ID == "recipient-removed" {
			hasRecipientRemoved = true
		}
		if itm.ID == "added-to-recipient" {
			hasRecipientAdded = true
		}
	}

	if !hasRecipientRemoved {
		t.Error("Rollback should restore removed items to recipient inventory")
	}
	if hasRecipientAdded {
		t.Error("Rollback should remove added items from recipient inventory")
	}
}

// TestTransferTracker_RollbackFunc_EmptyTransfers tests rollback with no transfers.
func TestTransferTracker_RollbackFunc_EmptyTransfers(t *testing.T) {
	proposerInv := &engine.InventoryComponent{}
	recipientInv := &engine.InventoryComponent{}

	// Add initial items
	initialItem1 := &item.Item{ID: "initial1", Name: "Initial 1"}
	initialItem2 := &item.Item{ID: "initial2", Name: "Initial 2"}
	proposerInv.AddItem(initialItem1)
	recipientInv.AddItem(initialItem2)

	// Create tracker with no transfers
	tracker := newTransferTracker()

	// Get initial counts
	initialProposerCount := len(proposerInv.Items)
	initialRecipientCount := len(recipientInv.Items)

	// Create and execute rollback (should be no-op)
	rollback := tracker.createRollbackFunc(proposerInv, recipientInv)
	rollback()

	// Verify inventories unchanged
	if len(proposerInv.Items) != initialProposerCount {
		t.Errorf("Rollback with no transfers changed proposer inventory: got %d items, want %d",
			len(proposerInv.Items), initialProposerCount)
	}
	if len(recipientInv.Items) != initialRecipientCount {
		t.Errorf("Rollback with no transfers changed recipient inventory: got %d items, want %d",
			len(recipientInv.Items), initialRecipientCount)
	}
}

// TestTransferTracker_NewTransferTracker tests tracker initialization.
func TestTransferTracker_NewTransferTracker(t *testing.T) {
	tracker := newTransferTracker()

	if tracker == nil {
		t.Fatal("newTransferTracker() returned nil")
	}

	// Verify all slices are nil or have zero length
	// (Go allows nil slices and they behave like empty slices)
	if len(tracker.removedFromProposer) != 0 {
		t.Errorf("removedFromProposer length = %d, want 0", len(tracker.removedFromProposer))
	}
	if len(tracker.removedFromRecipient) != 0 {
		t.Errorf("removedFromRecipient length = %d, want 0", len(tracker.removedFromRecipient))
	}
	if len(tracker.addedToRecipient) != 0 {
		t.Errorf("addedToRecipient length = %d, want 0", len(tracker.addedToRecipient))
	}
	if len(tracker.addedToProposer) != 0 {
		t.Errorf("addedToProposer length = %d, want 0", len(tracker.addedToProposer))
	}
}

// Helper function to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
