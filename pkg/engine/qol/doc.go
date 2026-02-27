// Package qol provides Quality of Life (QoL) features for enhanced player convenience.
//
// This package implements various convenience features to improve the player experience:
//
// Auto-Loot System:
//   - Companions automatically collect nearby items
//   - Configurable loot radius (5-10 tiles)
//   - Companion-specific collection preferences
//   - Smart filtering based on item quality
//
// Smart Crafting:
//   - Queue multiple recipes for sequential crafting
//   - Material availability checking
//   - Auto-consume materials from inventory
//   - Progress tracking and notifications
//
// Guild Invitations:
//   - Offline member acceptance system
//   - 7-day invitation expiry
//   - Notification on next login
//   - Automatic guild roster updates
//
// Mount Whistle:
//   - Summon nearby vehicles to player
//   - Distance-based arrival time (<5 seconds)
//   - Automatic pathfinding around obstacles
//   - Vehicle type filtering
//
// Storage Sorting:
//   - Auto-organize inventory by type/rarity
//   - Multiple sort criteria (type, rarity, name, value)
//   - Container management (chests, guild banks)
//   - Sort templates and presets
//
// Recipe Tracking:
//   - Show missing materials for crafts
//   - Highlight available recipes
//   - Material source hints
//   - Shopping list generation
//
// Example usage:
//
//	// Auto-loot system
//	autoLoot := qol.NewAutoLootManager()
//	autoLoot.SetRadius(companionID, 8.0) // 8 tiles
//	items := autoLoot.CollectNearby(companionID, playerPos)
//
//	// Smart crafting queue
//	craftQueue := qol.NewCraftQueue()
//	craftQueue.AddRecipe("iron_sword", 5)
//	craftQueue.Process(playerInventory)
//
//	// Storage sorting
//	sorter := qol.NewStorageSorter()
//	sorter.Sort(inventory, qol.SortByRarity)
//
// Thread Safety:
//
// All managers use sync.RWMutex for concurrent access protection.
// Safe to use from multiple goroutines (e.g., client, server, background tasks).
//
// Performance:
//
//   - Auto-loot: <1ms per collection cycle
//   - Craft queue: <5ms per recipe processing
//   - Storage sort: <10ms for 100 items
//   - Mount summon: <100ms pathfinding
//
// ECS Integration:
//
// The QoL package integrates with the ECS architecture through two mechanisms:
//
//  1. QoLSystemWrapper (pkg/engine/qol_system_wrapper.go):
//     - Implements engine.System interface for ECS world integration
//     - Registered in cmd/client/handlers.go:1422 via World.AddSystem()
//     - Update() runs every frame for periodic maintenance:
//     * Cleans up expired guild invitations every 5 minutes
//     * Other features (auto-loot, craft queue, mount whistle) are
//     triggered by their respective systems (companion AI, crafting, vehicles)
//     - System update order: QoL cleanup runs after gameplay systems
//
//  2. QoLComponent (types.go):
//     - Pure data component implementing Component interface
//     - Attached to player entities in cmd/client/handlers.go:2841
//     - Persisted via Serialize()/Deserialize() for save/load
//     - Stores player-specific QoL preferences (auto-loot radius, queue size, etc.)
//
// System Interaction Flow:
//
//   - Companion AI System → queries Manager.AutoLoot() to determine collection behavior
//   - Crafting System → queries Manager.CraftQueue() for queued recipes
//   - Vehicle System → queries Manager.MountWhistle() for summon requests
//   - Inventory System → queries Manager.RecipeTracker() for material tracking
//   - Guild System → queries Manager.GuildInvites() for offline invitations
//   - UI Systems → access Manager methods directly for player-initiated actions
//
// Integration:
//
//   - V4 Companions: Auto-loot collection behavior
//   - V8 Crafting: Smart queue system
//   - V8 Guilds: Offline invitation acceptance
//   - V4 Vehicles: Mount whistle summoning
package qol
