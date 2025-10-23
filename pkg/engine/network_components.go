// Package engine provides network-related components for multiplayer support.
package engine

// NetworkComponent marks an entity as network-synchronized.
// This component is used to associate entities with player IDs and control
// whether the entity's state should be synchronized over the network.
type NetworkComponent struct {
	// PlayerID is the network player ID this entity belongs to (0 for NPCs/items)
	PlayerID uint64

	// Synced indicates whether this entity should be synchronized over the network
	Synced bool

	// LastUpdateSeq tracks the last sequence number this entity was updated with
	LastUpdateSeq uint32
}

// Type returns the component type identifier.
func (n *NetworkComponent) Type() string {
	return "network"
}
