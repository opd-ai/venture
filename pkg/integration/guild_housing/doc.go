// Package guild_housing integrates V8 guild systems with V8 housing to create shared guild spaces.
//
// # Overview
//
// This package provides communal guild housing features including:
//   - Rank-based access permissions for guild houses
//   - Shared crafting stations allowing simultaneous use
//   - Guild storage rooms with deposit/withdraw logging
//   - Meeting halls with chat radius bonuses
//   - Guild-funded house upgrades
//
// # Features
//
// Shared Access Permissions:
//   - View: Can enter guild house and view furniture
//   - Use: Can interact with crafting stations
//   - Manage: Can place/remove furniture
//   - Admin: Full control including permissions
//
// Communal Crafting:
//   - 10+ members can craft simultaneously
//   - Queue system prevents resource conflicts
//   - Bonus multipliers stack with house station quality
//
// Guild Storage:
//   - 1000+ slot capacity for large guilds
//   - Deposit/withdraw transaction logging
//   - Rank-based withdrawal limits
//   - Automatic sorting and filtering
//
// Meeting Halls:
//   - Visual gathering space for guild members
//   - +50% chat radius for in-hall communication
//   - Guild announcement displays
//   - Event scheduling integration
//
// Guild Upgrades:
//   - Pooled resource improvements benefit all members
//   - 10k-100k gold per tier (4 tiers total)
//   - Enhanced station bonuses, storage capacity, chat radius
//
// # Usage Example
//
//	// Create guild housing manager
//	manager := guild_housing.NewManager()
//
//	// Create guild house with permissions
//	guildHouse := manager.CreateGuildHouse("guild-001", "player-001", housing.PlotSize{Width: 24, Height: 24})
//
//	// Set rank-based permissions
//	manager.SetPermission("guild-001", guild.RankMember, guild_housing.PermissionUse)
//
//	// Add communal crafting station
//	stationID := manager.AddCraftingStation("guild-001", housing_crafting.TypeForge, housing_crafting.QualityAdvanced)
//
//	// Add guild storage
//	storageID := manager.AddGuildStorage("guild-001", 500) // 500 slots
//
//	// Deposit items to storage
//	manager.DepositItem("guild-001", "player-001", "item-001", 10)
//
//	// Withdraw with rank check
//	withdrawn, err := manager.WithdrawItem("guild-001", "player-002", "item-001", 5)
//
//	// Upgrade guild house tier
//	manager.UpgradeHouse("guild-001", 50000) // 50k gold
//
// # Integration
//
// This package integrates with:
//   - pkg/world/housing: Base housing system, plot management
//   - pkg/network/federation/guild: Guild membership and ranks
//   - pkg/integration/housing_crafting: Crafting station bonuses
//   - pkg/procgen/furniture: Furniture generation and placement
//
// # Performance
//
// Target metrics:
//   - Permission check: <1ms
//   - Concurrent crafting: 10+ members simultaneously
//   - Storage operations: <10ms for 1000-slot searches
//   - Upgrade application: <100ms
//
// All operations are thread-safe using sync.RWMutex.
package guild_housing
