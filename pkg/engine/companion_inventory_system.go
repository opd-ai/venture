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
	allEntities := s.world.GetEntitiesWith("item", "position")

	for _, entity := range allEntities {
		if inv.IsFull() {
			break
		}

		if s.tryPickupItem(companion, entity, pos, inv) {
			s.logItemFetched(companion.ID, entity)
		}
	}
}

// tryPickupItem attempts to pick up an item entity if it's within range.
func (s *CompanionInventorySystem) tryPickupItem(companion, itemEntity *Entity, companionPos *PositionComponent, inv *CompanionInventoryComponent) bool {
	itemComp, itemPos, ok := s.getItemComponents(itemEntity)
	if !ok {
		return false
	}

	if !s.isItemInRange(companionPos, itemPos, inv.FetchRadius) {
		return false
	}

	return s.addItemToInventory(itemEntity, itemComp, inv)
}

// getItemComponents retrieves item and position components from an entity.
func (s *CompanionInventorySystem) getItemComponents(entity *Entity) (*ItemComponent, *PositionComponent, bool) {
	itemCompRaw, ok := entity.GetComponent("item")
	if !ok {
		return nil, nil, false
	}
	itemComp := itemCompRaw.(*ItemComponent)

	itemPosRaw, ok := entity.GetComponent("position")
	if !ok {
		return nil, nil, false
	}
	itemPos := itemPosRaw.(*PositionComponent)

	return itemComp, itemPos, true
}

// isItemInRange checks if an item is within the companion's fetch radius.
func (s *CompanionInventorySystem) isItemInRange(companionPos, itemPos *PositionComponent, fetchRadius float64) bool {
	distance := s.distance(companionPos, itemPos)
	return distance <= fetchRadius
}

// addItemToInventory adds an item to the companion's inventory and removes it from the world.
func (s *CompanionInventorySystem) addItemToInventory(itemEntity *Entity, itemComp *ItemComponent, inv *CompanionInventoryComponent) bool {
	if !inv.CanAddItem(itemComp.Item) {
		return false
	}

	if !inv.AddItem(itemComp.Item) {
		return false
	}

	s.world.RemoveEntity(itemEntity.ID)
	return true
}

// logItemFetched logs successful item pickup by companion.
func (s *CompanionInventorySystem) logItemFetched(companionID uint64, itemEntity *Entity) {
	if s.logger == nil {
		return
	}

	itemCompRaw, ok := itemEntity.GetComponent("item")
	if !ok {
		return
	}
	itemComp := itemCompRaw.(*ItemComponent)

	s.logger.WithFields(logrus.Fields{
		"companion": companionID,
		"item":      itemComp.Item.Name,
	}).Debug("companion fetched item")
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
