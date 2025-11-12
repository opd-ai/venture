package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

func TestCompanionInventoryComponent_Type(t *testing.T) {
	comp := NewCompanionInventoryComponent(10, 100.0)
	if comp.Type() != "companioninventory" {
		t.Errorf("expected 'companioninventory', got '%s'", comp.Type())
	}
}

func TestNewCompanionInventoryComponent(t *testing.T) {
	comp := NewCompanionInventoryComponent(10, 100.0)

	if comp.MaxItems != 10 {
		t.Errorf("expected MaxItems 10, got %d", comp.MaxItems)
	}
	if comp.MaxWeight != 100.0 {
		t.Errorf("expected MaxWeight 100.0, got %f", comp.MaxWeight)
	}
	if len(comp.Items) != 0 {
		t.Errorf("expected empty items, got %d", len(comp.Items))
	}
	if comp.FetchRadius != 100.0 {
		t.Errorf("expected FetchRadius 100.0, got %f", comp.FetchRadius)
	}
}

func TestCompanionInventoryComponent_AddItem(t *testing.T) {
	tests := []struct {
		name      string
		maxItems  int
		maxWeight float64
		items     []*item.Item
		wantCount int
	}{
		{
			name:      "add single item",
			maxItems:  10,
			maxWeight: 100.0,
			items: []*item.Item{
				{Name: "Item1", Stats: item.Stats{Weight: 5.0}},
			},
			wantCount: 1,
		},
		{
			name:      "add multiple items",
			maxItems:  10,
			maxWeight: 100.0,
			items: []*item.Item{
				{Name: "Item1", Stats: item.Stats{Weight: 5.0}},
				{Name: "Item2", Stats: item.Stats{Weight: 10.0}},
				{Name: "Item3", Stats: item.Stats{Weight: 3.0}},
			},
			wantCount: 3,
		},
		{
			name:      "exceed max items",
			maxItems:  2,
			maxWeight: 100.0,
			items: []*item.Item{
				{Name: "Item1", Stats: item.Stats{Weight: 5.0}},
				{Name: "Item2", Stats: item.Stats{Weight: 5.0}},
				{Name: "Item3", Stats: item.Stats{Weight: 5.0}},
			},
			wantCount: 2,
		},
		{
			name:      "exceed max weight",
			maxItems:  10,
			maxWeight: 20.0,
			items: []*item.Item{
				{Name: "Item1", Stats: item.Stats{Weight: 10.0}},
				{Name: "Item2", Stats: item.Stats{Weight: 9.0}},
				{Name: "Item3", Stats: item.Stats{Weight: 5.0}},
			},
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewCompanionInventoryComponent(tt.maxItems, tt.maxWeight)

			for _, itm := range tt.items {
				comp.AddItem(itm)
			}

			if comp.GetItemCount() != tt.wantCount {
				t.Errorf("expected %d items, got %d", tt.wantCount, comp.GetItemCount())
			}
		})
	}
}

func TestCompanionInventoryComponent_GetCurrentWeight(t *testing.T) {
	comp := NewCompanionInventoryComponent(10, 100.0)

	comp.AddItem(&item.Item{Name: "Item1", Stats: item.Stats{Weight: 5.5}})
	comp.AddItem(&item.Item{Name: "Item2", Stats: item.Stats{Weight: 10.3}})
	comp.AddItem(&item.Item{Name: "Item3", Stats: item.Stats{Weight: 3.2}})

	expected := 19.0
	actual := comp.GetCurrentWeight()

	if actual < expected-0.01 || actual > expected+0.01 {
		t.Errorf("expected weight ~%f, got %f", expected, actual)
	}
}

func TestCompanionInventoryComponent_CanAddItem(t *testing.T) {
	tests := []struct {
		name        string
		currentLoad int
		maxItems    int
		maxWeight   float64
		itemWeight  float64
		want        bool
	}{
		{"can add to empty", 0, 10, 100.0, 5.0, true},
		{"can add within limits", 5, 10, 100.0, 5.0, true},
		{"cannot add at max items", 10, 10, 100.0, 5.0, false},
		{"cannot add exceeding weight", 0, 10, 20.0, 25.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewCompanionInventoryComponent(tt.maxItems, tt.maxWeight)

			// Fill inventory to current load
			for i := 0; i < tt.currentLoad; i++ {
				comp.AddItem(&item.Item{Name: "Filler", Stats: item.Stats{Weight: 1.0}})
			}

			newItem := &item.Item{Name: "Test", Stats: item.Stats{Weight: tt.itemWeight}}
			got := comp.CanAddItem(newItem)

			if got != tt.want {
				t.Errorf("CanAddItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompanionInventoryComponent_RemoveItem(t *testing.T) {
	comp := NewCompanionInventoryComponent(10, 100.0)

	item1 := &item.Item{Name: "Item1", Stats: item.Stats{Weight: 5.0}}
	item2 := &item.Item{Name: "Item2", Stats: item.Stats{Weight: 10.0}}
	item3 := &item.Item{Name: "Item3", Stats: item.Stats{Weight: 3.0}}

	comp.AddItem(item1)
	comp.AddItem(item2)
	comp.AddItem(item3)

	// Remove middle item
	removed := comp.RemoveItem(1)
	if removed != item2 {
		t.Errorf("expected to remove item2, got different item")
	}
	if comp.GetItemCount() != 2 {
		t.Errorf("expected 2 items after removal, got %d", comp.GetItemCount())
	}

	// Try to remove invalid index
	removed = comp.RemoveItem(10)
	if removed != nil {
		t.Errorf("expected nil for invalid index, got item")
	}
}

func TestCompanionInventoryComponent_RemoveItemByReference(t *testing.T) {
	comp := NewCompanionInventoryComponent(10, 100.0)

	item1 := &item.Item{Name: "Item1", Stats: item.Stats{Weight: 5.0}}
	item2 := &item.Item{Name: "Item2", Stats: item.Stats{Weight: 10.0}}
	item3 := &item.Item{Name: "Item3", Stats: item.Stats{Weight: 3.0}}
	notAdded := &item.Item{Name: "NotAdded", Stats: item.Stats{Weight: 1.0}}

	comp.AddItem(item1)
	comp.AddItem(item2)
	comp.AddItem(item3)

	// Remove by reference
	if !comp.RemoveItemByReference(item2) {
		t.Errorf("expected successful removal of item2")
	}
	if comp.GetItemCount() != 2 {
		t.Errorf("expected 2 items after removal, got %d", comp.GetItemCount())
	}

	// Try to remove non-existent item
	if comp.RemoveItemByReference(notAdded) {
		t.Errorf("expected false for non-existent item")
	}
}

func TestCompanionInventoryComponent_TransferToOwner(t *testing.T) {
	companion := NewCompanionInventoryComponent(10, 100.0)
	owner := NewInventoryComponent(5, 50.0)

	// Add items to companion
	companion.AddItem(&item.Item{Name: "Item1", Stats: item.Stats{Weight: 5.0}})
	companion.AddItem(&item.Item{Name: "Item2", Stats: item.Stats{Weight: 10.0}})
	companion.AddItem(&item.Item{Name: "Item3", Stats: item.Stats{Weight: 3.0}})

	// Transfer to owner
	untransferred := companion.TransferToOwner(owner)

	if len(untransferred) != 0 {
		t.Errorf("expected 0 untransferred items, got %d", len(untransferred))
	}
	if companion.GetItemCount() != 0 {
		t.Errorf("expected companion to have 0 items, got %d", companion.GetItemCount())
	}
	if owner.GetItemCount() != 3 {
		t.Errorf("expected owner to have 3 items, got %d", owner.GetItemCount())
	}
}

func TestCompanionInventoryComponent_TransferToOwner_OwnerFull(t *testing.T) {
	companion := NewCompanionInventoryComponent(10, 100.0)
	owner := NewInventoryComponent(2, 100.0)

	// Add items to companion
	companion.AddItem(&item.Item{Name: "Item1", Stats: item.Stats{Weight: 5.0}})
	companion.AddItem(&item.Item{Name: "Item2", Stats: item.Stats{Weight: 10.0}})
	companion.AddItem(&item.Item{Name: "Item3", Stats: item.Stats{Weight: 3.0}})

	// Transfer to owner (who can only hold 2 items)
	untransferred := companion.TransferToOwner(owner)

	if len(untransferred) != 1 {
		t.Errorf("expected 1 untransferred item, got %d", len(untransferred))
	}
	if companion.GetItemCount() != 1 {
		t.Errorf("expected companion to have 1 item, got %d", companion.GetItemCount())
	}
	if owner.GetItemCount() != 2 {
		t.Errorf("expected owner to have 2 items, got %d", owner.GetItemCount())
	}
}

func TestCompanionInventoryComponent_IsFull(t *testing.T) {
	tests := []struct {
		name      string
		maxItems  int
		maxWeight float64
		items     []*item.Item
		want      bool
	}{
		{
			name:      "empty is not full",
			maxItems:  10,
			maxWeight: 100.0,
			items:     []*item.Item{},
			want:      false,
		},
		{
			name:      "full by item count",
			maxItems:  2,
			maxWeight: 100.0,
			items: []*item.Item{
				{Name: "Item1", Stats: item.Stats{Weight: 5.0}},
				{Name: "Item2", Stats: item.Stats{Weight: 5.0}},
			},
			want: true,
		},
		{
			name:      "full by weight",
			maxItems:  10,
			maxWeight: 20.0,
			items: []*item.Item{
				{Name: "Item1", Stats: item.Stats{Weight: 20.0}},
			},
			want: true,
		},
		{
			name:      "not full",
			maxItems:  10,
			maxWeight: 100.0,
			items: []*item.Item{
				{Name: "Item1", Stats: item.Stats{Weight: 5.0}},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewCompanionInventoryComponent(tt.maxItems, tt.maxWeight)
			for _, itm := range tt.items {
				comp.AddItem(itm)
			}

			if got := comp.IsFull(); got != tt.want {
				t.Errorf("IsFull() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompanionInventoryComponent_Clear(t *testing.T) {
	comp := NewCompanionInventoryComponent(10, 100.0)

	comp.AddItem(&item.Item{Name: "Item1", Stats: item.Stats{Weight: 5.0}})
	comp.AddItem(&item.Item{Name: "Item2", Stats: item.Stats{Weight: 10.0}})

	comp.Clear()

	if comp.GetItemCount() != 0 {
		t.Errorf("expected 0 items after clear, got %d", comp.GetItemCount())
	}
	if comp.GetCurrentWeight() != 0.0 {
		t.Errorf("expected 0 weight after clear, got %f", comp.GetCurrentWeight())
	}
}
