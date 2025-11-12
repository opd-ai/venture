package engine

import (
	"math"

	"github.com/opd-ai/venture/pkg/procgen/item"
	"github.com/sirupsen/logrus"
)

// CompanionInventorySystem handles companion item fetching and carrying.
// Companions can automatically pick up items when AutoFetch is enabled,
// and can be commanded to fetch specific items.
type CompanionInventorySystem struct {
	world  *World
	logger *logrus.Entry
}

// NewCompanionInventorySystem creates a new companion inventory system.
func NewCompanionInventorySystem(world *World) *CompanionInventorySystem {
	return NewCompanionInventorySystemWithLogger(world, nil)
}

// NewCompanionInventorySystemWithLogger creates a system with a logger.
func NewCompanionInventorySystemWithLogger(world *World, logger *logrus.Logger) *CompanionInventorySystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "companion_inventory",
		})
		logEntry.Debug("companion inventory system created")
	}

	return &CompanionInventorySystem{
		world:  world,
		logger: logEntry,
	}
}

// Update processes companion inventory behavior.
func (s *CompanionInventorySystem) Update(deltaTime float64) {
	if s.world == nil {
		return
	}

	// Get all companions with inventory
	companions := s.world.GetEntitiesWith("companion", "companioninventory", "position")

	for _, companion := range companions {
		companionCompRaw, ok := companion.GetComponent("companion")
		if !ok {
			continue
		}
		companionComp := companionCompRaw.(*CompanionComponent)

		invCompRaw, ok := companion.GetComponent("companioninventory")
		if !ok {
			continue
		}
		invComp := invCompRaw.(*CompanionInventoryComponent)

		posCompRaw, ok := companion.GetComponent("position")
		if !ok {
			continue
		}
		posComp := posCompRaw.(*PositionComponent)

		// Only process if AutoFetch is enabled
		if !invComp.AutoFetch {
			continue
		}

		// Check if companion has room in inventory
		if invComp.IsFull() {
			continue
		}

		// Get owner to verify companion is valid
		owner, ok := s.world.GetEntity(companionComp.OwnerID)
		if !ok || owner == nil {
			continue
		}

		// Find nearby items within fetch radius
		s.fetchNearbyItems(companion, posComp, invComp)
	}
}

// fetchNearbyItems searches for and picks up items within fetch radius.
func (s *CompanionInventorySystem) fetchNearbyItems(companion *Entity, pos *PositionComponent, inv *CompanionInventoryComponent) {
	// Get all entities with item components in the world
	// Note: This is a simplified implementation. In a real game, you would use
	// spatial partitioning (quadtree) for efficiency.
	allEntities := s.world.GetEntitiesWith("item", "position")

	for _, entity := range allEntities {
		// Skip if inventory is full
		if inv.IsFull() {
			break
		}

		// Get item component
		itemCompRaw, ok := entity.GetComponent("item")
		if !ok {
			continue
		}
		itemComp := itemCompRaw.(*ItemComponent)

		// Get item position
		itemPosRaw, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		itemPos := itemPosRaw.(*PositionComponent)

		// Check if item is within fetch radius
		distance := s.distance(pos, itemPos)
		if distance > inv.FetchRadius {
			continue
		}

		// Try to pick up the item
		if inv.CanAddItem(itemComp.Item) {
			if inv.AddItem(itemComp.Item) {
				// Remove item entity from world
				s.world.RemoveEntity(entity.ID)

				if s.logger != nil {
					s.logger.WithFields(logrus.Fields{
						"companion": companion.ID,
						"item":      itemComp.Item.Name,
					}).Debug("companion fetched item")
				}
			}
		}
	}
}

// distance calculates the Euclidean distance between two positions.
func (s *CompanionInventorySystem) distance(a, b *PositionComponent) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// TransferItemsToOwner transfers all items from companion to owner's inventory.
// Returns the number of items successfully transferred.
func (s *CompanionInventorySystem) TransferItemsToOwner(companionID uint64) int {
	if s.world == nil {
		return 0
	}

	companion, ok := s.world.GetEntity(companionID)
	if !ok || companion == nil {
		return 0
	}

	companionCompRaw, ok := companion.GetComponent("companion")
	if !ok {
		return 0
	}
	companionComp := companionCompRaw.(*CompanionComponent)

	invCompRaw, ok := companion.GetComponent("companioninventory")
	if !ok {
		return 0
	}
	invComp := invCompRaw.(*CompanionInventoryComponent)

	// Get owner
	owner, ok := s.world.GetEntity(companionComp.OwnerID)
	if !ok || owner == nil {
		return 0
	}

	ownerInvRaw, ok := owner.GetComponent("inventory")
	if !ok {
		return 0
	}
	ownerInv := ownerInvRaw.(*InventoryComponent)

	// Transfer items
	initialCount := invComp.GetItemCount()
	untransferred := invComp.TransferToOwner(ownerInv)
	transferred := initialCount - len(untransferred)

	if s.logger != nil && transferred > 0 {
		s.logger.WithFields(logrus.Fields{
			"companion":   companionID,
			"transferred": transferred,
			"remaining":   len(untransferred),
		}).Debug("transferred items to owner")
	}

	return transferred
}

// ItemComponent wraps an item for entity representation.
type ItemComponent struct {
	Item *item.Item
}

// Type returns the component type identifier.
func (i *ItemComponent) Type() string {
	return "item"
}
