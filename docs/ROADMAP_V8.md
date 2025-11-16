# Development Roadmap - Version 8.0: Player Housing & Guild Systems

## Current Status

**Status:** PLANNING - Not Yet Started  
**Prerequisites:** V7.0 completion  
**Projected Start:** Post-2028  

This is a future planning document. No V8.0 features have been implemented yet.

## Overview

**Version:** 8.0 (Previous: 7.0 - Planning)  
**Focus:** Comprehensive feature completion addressing all deferred items from V3-V7  
**Timeline:** 10-14 months development window  
**Projected Completion:** Q4 2029

**Major Themes:**
1. **Persistent Housing & Guilds:** Cross-server player structures and social organization
2. **Advanced Physics & Simulation:** Refined vehicle physics, fluid dynamics, environmental interactions
3. **Enhanced Social Systems:** Persistent trust/reputation, chat history, image storage
4. **Federation Extensions:** WebRTC browser-to-browser, mobile federation, server mods
5. **Deep Gameplay Systems:** Companion AI depth, procedural storytelling, class customization
6. **Content Tools:** Blueprint sharing, mod framework, procedural content editor

## Deliverables

**Housing & Guild Systems (V6.0 Future Items):**
- Procedurally generated player housing with cross-server persistence
- Multi-server guild system with shared territories and resources
- Guild halls and cross-server structures
- Territory control and guild warfare mechanics

**Advanced Physics & Simulation (V4.0 Future Items):**
- Enhanced vehicle physics (suspension, weight transfer, terrain deformation)
- Fluid dynamics (water flow, swimming, flooding)
- Advanced companion AI with skill learning and emergent behavior
- Environmental physics (destructible buildings, realistic falling objects)

**Social System Enhancements (V5.0 Future Items):**
- Persistent trust scores and reputation system
- Chat history with delta compression and synchronization
- Persistent image storage and gallery system
- Advanced trade reputation with cross-server tracking

**Federation Extensions (V6.0 Future Items):**
- WebRTC-based federation (browser-to-browser servers, no dedicated server required)
- Mobile device federation (phones/tablets as federated servers)
- Server mod framework (custom rules, zero-asset constraint maintained)
- P2P relay network for NAT traversal

**Gameplay Depth (V4.0 Future Items):**
- Deep companion AI with skill progression and personality evolution
- Complex procedural storytelling with branching narratives
- Advanced class customization (multi-classing, prestige classes, talent trees)
- Expanded mini-game system with tournaments and leaderboards

## Targets

**Performance:** 60 FPS minimum, <500MB memory total  
**Quality:** ≥65% test coverage, backward compatible with v7.0  
**Persistence:** <150MB total per player (housing 50MB + trust/reputation 20MB + chat history 30MB + images 50MB)  
**Network:** <50KB/s total federation overhead (housing 15KB/s + guilds 10KB/s + trust sync 5KB/s + WebRTC signaling 10KB/s + mobile 10KB/s)  
**Physics:** Maintain determinism for core gameplay, allow non-deterministic environmental effects  
**Storage:** <1GB server-side per 100 active players (persistent data)

## Architecture

**Housing System:** `pkg/world/housing/` (manager, builder, persistence) → Building placement → Interior generation → Furniture spawning  
**Guild System:** `pkg/network/federation/guild/` (registry, permissions, resources) → Cross-server sync → Territory management  
**Building Generator:** `pkg/procgen/building/` (templates, materials, rooms) → Procedural architecture → Style variation  
**Furniture Generator:** `pkg/procgen/furniture/` (items, decorations, functional objects) → Genre-specific styles

**Components:**
```go
// pkg/world/housing/components.go
type HousingComponent struct {
    OwnerID       string        // Player or Guild ID
    BuildingType  BuildingType  // House, GuildHall, Workshop, Storage
    Location      ChunkCoords   // World position
    Dimensions    Vector3       // Width, Height, Depth
    Style         ArchStyle     // Genre-based architecture
    Rooms         []*Room       // Interior layout
    Furniture     []*Furniture  // Placed objects
    Upgrades      []Upgrade     // Structural improvements
    Permissions   PermissionSet // Access control
}

type GuildComponent struct {
    GuildID       string
    Name          string          // Procedurally generated
    Emblem        *GuildEmblem    // Procedural symbol
    MemberIDs     []string        // Player IDs
    Ranks         []GuildRank     // Hierarchy
    Resources     ResourcePool    // Shared materials
    Territories   []TerritoryID   // Controlled zones
    GuildHalls    []HousingID     // Structures
    Reputation    map[string]float64 // Server-based rep
    Treasury      int             // Shared gold
}

type BuildingMaterialComponent struct {
    MaterialType  MaterialType  // Wood, Stone, Metal, Crystal, Energy
    Durability    float64
    Color         color.RGBA
    Texture       TexturePattern
    WeatherProof  bool
}
```

**Systems:**
- `HousingSystem`: Construction, placement validation, persistence
- `GuildManagementSystem`: Member management, resource distribution, territory control
- `BuildingGeneratorSystem`: Procedural architecture, room layout, material selection
- `FurnitureSystem`: Object placement, functional interactions (crafting stations, storage)
- `TerritorySystem`: Ownership tracking, conflict resolution, benefits calculation

---

## Phase 49: Housing Foundation

**Focus:** Player housing core systems and procedural building generation  
**Status:** PLANNING

### 49.1: Housing Core Infrastructure

**Deliverables:**
- [ ] Create `pkg/world/housing/` (manager, builder, persistence)
- [ ] HousingComponent with ownership, permissions, dimensions
- [ ] Plot placement system with collision detection and terrain validation
- [ ] Building size tiers: Small (8×8 tiles), Medium (16×16), Large (24×24), Estate (32×32)
- [ ] Save/load for housing structures (JSON + gzip compression)
- [ ] CLI flag: `-enable-housing` (default: true in v8.0)

**Components:**
```go
// pkg/world/housing/manager.go
type HousingManager struct {
    plots         map[PlotID]*HousingPlot
    ownerIndex    map[string][]PlotID  // PlayerID -> PlotIDs
    serverIndex   map[string][]PlotID  // ServerID -> PlotIDs
    persistence   *HousingPersistence
    maxPlotsPerPlayer int  // Default: 3
}

type HousingPlot struct {
    PlotID      string
    OwnerID     string
    ServerID    string
    Location    ChunkCoords
    Building    *Building
    CreatedAt   int64
    LastModified int64
}

type Building struct {
    BuildingType BuildingType
    Style        ArchStyle
    Materials    []BuildingMaterial
    Dimensions   Vector3
    Rooms        []*Room
    Entrance     Vector2  // Door position
    Foundation   FoundationType
}
```

**Metrics:** 
- Plot allocation: <50ms per placement
- Collision detection: <10ms for 1000 existing plots
- Save: <100ms per housing structure
- Load: <200ms per housing structure
- Memory: <5MB per 100 housing plots

**Success Criteria:**
- [x] Players can claim building plots in designated housing zones
- [x] Buildings persist across server restarts
- [x] Placement validation prevents overlap with terrain/entities/other plots
- [x] Support for 1000+ concurrent housing plots per server
- [x] Cross-server housing support via V6.0 federation

### 49.2: Procedural Building Generation

**Deliverables:**
- [ ] Create `pkg/procgen/building/` (templates, generator, validator)
- [ ] 5 building types: House, Workshop, Storage, Tower, Manor
- [ ] 5 architectural styles per genre (25 total style templates)
- [ ] Procedural floor plans with 1-8 rooms
- [ ] Roof generation (flat, gabled, hipped, pyramidal, domed)
- [ ] Window and door placement based on architectural rules
- [ ] Material palette generation (wood, stone, brick, metal, crystal, energy)

**Generation Algorithm:**
```go
// pkg/procgen/building/generator.go
type BuildingGenerator struct {
    styleTemplates map[GenreID][]ArchStyle
    roomTemplates  []RoomTemplate
}

func (g *BuildingGenerator) Generate(seed int64, params GenerationParams) (*Building, error) {
    rng := rand.New(rand.NewSource(seed))
    
    // 1. Select architectural style based on genre
    style := g.selectStyle(params.GenreID, rng)
    
    // 2. Generate floor plan (room layout)
    rooms := g.generateFloorPlan(params.BuildingType, params.Depth, rng)
    
    // 3. Assign materials based on style and depth
    materials := g.selectMaterials(style, params.Depth, rng)
    
    // 4. Generate roof structure
    roof := g.generateRoof(style, params.BuildingType, rng)
    
    // 5. Place windows and doors
    openings := g.placeOpenings(rooms, style, rng)
    
    // 6. Add decorative elements (optional, based on rarity)
    decorations := g.generateDecorations(style, params.Difficulty, rng)
    
    return &Building{
        Style:       style,
        Materials:   materials,
        Rooms:       rooms,
        Roof:        roof,
        Openings:    openings,
        Decorations: decorations,
    }, nil
}
```

**Style Examples:**
- Fantasy: Medieval timber frame, stone castles, thatched cottages, wizard towers, grand manors
- Sci-Fi: Modular prefab, energy domes, crystalline structures, hovering platforms, tech spires
- Horror: Decrepit mansions, twisted wood, broken windows, shadowy interiors, haunted estates
- Cyberpunk: Neon-lit apartments, metal containers, holographic walls, rooftop shanties, corp towers
- Post-Apocalyptic: Scrap shelters, reinforced bunkers, reclaimed ruins, fortified compounds, wasteland huts

**Metrics:**
- Generation time: <100ms per building
- Floor plan validation: 100% navigable (all rooms accessible)
- Material consistency: ≥90% style adherence
- Determinism: Same seed = identical building

**Success Criteria:**
- [x] 5 building types generate distinct layouts
- [x] Genre recognition: 85%+ player identification of style
- [x] Room accessibility: 100% rooms have valid pathways
- [x] Visual variety: <5% duplicate buildings with random seeds

### 49.3: Interior Generation & Room Layouts

**Deliverables:**
- [ ] Room type system: Bedroom, Kitchen, Workshop, Storage, Library, Laboratory, Armory, Hall
- [ ] Door and corridor generation for multi-room buildings
- [ ] Wall decoration generation (paintings, windows, shelves)
- [ ] Floor and ceiling texture patterns
- [ ] Lighting fixtures (chandeliers, torches, lamps, crystals)
- [ ] Functional room requirements (workshop needs crafting station, kitchen needs storage)

**Room Templates:**
```go
// pkg/procgen/building/room_template.go
type RoomTemplate struct {
    RoomType      RoomType
    MinSize       Vector2  // Minimum dimensions
    MaxSize       Vector2  // Maximum dimensions
    RequiredItems []ItemType  // Must-have furniture
    OptionalItems []ItemType  // Nice-to-have furniture
    WallDecor     []DecorType // Paintings, shelves, etc.
    FloorPattern  PatternType
    LightingType  LightType
}

// Example: Workshop room
var WorkshopTemplate = RoomTemplate{
    RoomType:      RoomWorkshop,
    MinSize:       Vector2{4, 4},
    MaxSize:       Vector2{8, 8},
    RequiredItems: []ItemType{CraftingStation, Storage, Workbench},
    OptionalItems: []ItemType{Furnace, Anvil, ToolRack, Blueprint},
    WallDecor:     []DecorType{ToolBoard, Blueprint, Clock},
    FloorPattern:  PatternStone,
    LightingType:  LightWorkshop,  // Bright, even illumination
}
```

**Metrics:**
- Room generation: <20ms per room
- Furniture placement: 100% non-overlapping
- Door placement: Valid connections between all rooms
- Lighting coverage: ≥80% room area illuminated

**Success Criteria:**
- [x] 8 room types with unique functionality
- [x] Multi-room buildings have connected layouts
- [x] Functional rooms spawn required furniture automatically
- [x] Lighting adapts to room size and purpose

---

## Phase 50: Furniture & Decoration System

**Focus:** Procedural furniture generation and placement  
**Status:** PLANNING

### 50.1: Furniture Generation

**Deliverables:**
- [ ] Create `pkg/procgen/furniture/` (generator, templates, materials)
- [ ] 30+ furniture types across 8 categories:
  - Seating: Chair, Bench, Throne, Stool, Couch
  - Tables: Dining, Work, Coffee, Nightstand, Desk
  - Storage: Chest, Cabinet, Wardrobe, Shelf, Barrel
  - Beds: Single, Double, Bunk, Royal, Sleeping Bag
  - Crafting: Workbench, Anvil, Alchemy Station, Enchanting Table
  - Lighting: Chandelier, Torch, Lamp, Candle, Crystal
  - Decoration: Painting, Sculpture, Plant, Rug, Tapestry
  - Functional: Fireplace, Door, Window, Stairs, Ladder
- [ ] Material variation: Wood (oak, pine, ebony), Metal (iron, steel, gold), Stone, Crystal, Fabric
- [ ] Rarity tiers affecting visual detail and functionality
- [ ] Genre-specific style variations

**Generation System:**
```go
// pkg/procgen/furniture/generator.go
type FurnitureGenerator struct {
    templates map[FurnitureType]*FurnitureTemplate
}

func (g *FurnitureGenerator) Generate(seed int64, params GenerationParams) (*Furniture, error) {
    rng := rand.New(rand.NewSource(seed))
    
    template := g.templates[params.FurnitureType]
    
    // 1. Select base material
    material := g.selectMaterial(params.GenreID, params.Depth, rng)
    
    // 2. Generate dimensions (within template constraints)
    dimensions := g.scaleDimensions(template.BaseSize, params.Difficulty, rng)
    
    // 3. Generate color palette
    colors := g.generateColorScheme(material, params.GenreID, rng)
    
    // 4. Add decorative elements based on rarity
    decorations := g.generateDecorations(template.DecorSlots, params.Difficulty, rng)
    
    // 5. Determine functional properties
    functionality := g.assignFunctionality(template.FunctionType, params.Depth)
    
    return &Furniture{
        Type:         params.FurnitureType,
        Material:     material,
        Dimensions:   dimensions,
        Colors:       colors,
        Decorations:  decorations,
        Functionality: functionality,
    }, nil
}
```

**Metrics:**
- Generation time: <10ms per furniture item
- Visual variety: >95% unique items with random seeds
- Material consistency: Genre-appropriate 90%+ of time
- Functional validation: 100% functional items spawn correctly

**Success Criteria:**
- [x] 30+ furniture types with distinct visuals
- [x] Genre recognition: 80%+ style identification
- [x] Rarity affects visual detail (common: simple, legendary: ornate)
- [x] Functional furniture integrates with existing systems (crafting, storage)

### 50.2: Furniture Placement & Interaction

**Deliverables:**
- [ ] Placement validation system (collision detection, room boundaries)
- [ ] Snap-to-wall logic for wall-mounted items
- [ ] Furniture rotation (4 or 8 directions)
- [ ] Stacking rules (items on tables, shelves)
- [ ] Interaction system: Use, Move, Remove, Upgrade
- [ ] FurnitureComponent integration with ECS
- [ ] Storage furniture inventory integration

**Component:**
```go
// pkg/engine/furniture_component.go
type FurnitureComponent struct {
    FurnitureType  FurnitureType
    OwnerID        string        // Player or Guild ID
    PlotID         string        // Housing plot
    RoomID         string        // Room within building
    Position       Vector2       // Tile position
    Rotation       int           // 0-7 (8 directions)
    Material       MaterialType
    Rarity         Rarity
    Durability     float64
    Functionality  FunctionType  // Storage, Crafting, Seating, Lighting, Decoration
    Inventory      *Inventory    // For storage furniture
    UpgradeSlots   []UpgradeSlot
}
```

**Interaction Types:**
- **Use:** Sit on chair, sleep in bed, craft at workbench, read book from shelf
- **Move:** Drag furniture to new position (if owner/permissions allow)
- **Remove:** Delete furniture, returns to inventory (if owner/permissions allow)
- **Upgrade:** Apply enhancement (reinforced, enchanted, gilded)
- **Store/Retrieve:** Access storage furniture inventory

**Metrics:**
- Placement validation: <5ms per item
- Interaction response: <10ms
- Storage access: <50ms (reuse existing inventory system)
- Move/rotate: <20ms with collision re-check

**Success Criteria:**
- [x] Furniture placement respects collision boundaries
- [x] Wall items snap to walls automatically
- [x] Players can interact with functional furniture
- [x] Storage furniture works with existing inventory UI
- [x] Furniture persists position/rotation across server restarts

### 50.3: Decoration & Customization

**Deliverables:**
- [ ] Wall decoration system (paintings, tapestries, trophies, weapons)
- [ ] Floor covering (rugs, carpets, tiles)
- [ ] Lighting customization (color, intensity, fixture type)
- [ ] Procedural painting generation (abstract patterns, landscapes, portraits)
- [ ] Trophy system (display killed boss, rare item, achievement)
- [ ] Color theme selector (player picks palette, updates all furniture)

**Procedural Painting:**
```go
// pkg/procgen/furniture/painting.go
type PaintingGenerator struct {
    styles []PaintingStyle
}

func (g *PaintingGenerator) Generate(seed int64, params GenerationParams) (*Painting, error) {
    rng := rand.New(rand.NewSource(seed))
    
    style := g.styles[rng.Intn(len(g.styles))]
    
    switch style {
    case StyleAbstract:
        return g.generateAbstract(rng, params.GenreID)
    case StyleLandscape:
        return g.generateLandscape(rng, params.GenreID)
    case StylePortrait:
        return g.generatePortrait(rng, params.GenreID)
    case StyleGeometric:
        return g.generateGeometric(rng, params.GenreID)
    }
}
```

**Metrics:**
- Painting generation: <50ms per painting
- Trophy rendering: <20ms per trophy item
- Theme update: <500ms for entire building
- Visual coherence: 85%+ theme consistency

**Success Criteria:**
- [x] Paintings are visually distinct and genre-appropriate
- [x] Trophies display correct item/entity representation
- [x] Color themes apply to all compatible furniture
- [x] Decorations persist and sync across multiplayer

---

## Phase 51: Guild System Foundation

**Focus:** Cross-server guild creation, membership, and basic management  
**Status:** PLANNING

### 51.1: Guild Creation & Management

**Deliverables:**
- [ ] Create `pkg/network/federation/guild/` (manager, registry, permissions)
- [ ] Guild creation with procedurally generated names and emblems
- [ ] Member invitation and approval system
- [ ] Guild ranks with customizable permissions (5 default ranks, up to 10 custom)
- [ ] Guild roster management (promote, demote, kick)
- [ ] Guild message board (text announcements, no chat persistence)
- [ ] Cross-server guild registry via V6.0 federation

**Components:**
```go
// pkg/network/federation/guild/manager.go
type GuildManager struct {
    guilds        map[GuildID]*Guild
    memberIndex   map[PlayerID]GuildID
    serverRegistry map[ServerID][]GuildID
    federation    *FederationClient
}

type Guild struct {
    GuildID       string
    Name          string
    Emblem        *GuildEmblem
    FounderID     string
    CreatedAt     int64
    HomeServerID  string
    Members       map[PlayerID]*GuildMember
    Ranks         []GuildRank
    Permissions   PermissionMatrix
    Treasury      int
    Reputation    map[ServerID]float64
    MessageBoard  []*GuildMessage
}

type GuildMember struct {
    PlayerID      string
    JoinedAt      int64
    RankID        int
    Contributions int  // Guild points
    LastActive    int64
}

type GuildRank struct {
    RankID        int
    Name          string
    Permissions   PermissionSet
    MinContrib    int  // Required contributions
}

type PermissionSet struct {
    CanInvite     bool
    CanKick       bool
    CanPromote    bool
    CanWithdraw   bool
    CanBuild      bool
    CanEditHall   bool
    CanManageTerritory bool
}
```

**Default Ranks:**
1. **Guild Master:** Full permissions
2. **Officer:** Invite, promote (to Member), manage resources
3. **Veteran:** Build, edit guild hall, withdraw limited gold
4. **Member:** Access guild facilities, contribute resources
5. **Recruit:** View only, cannot withdraw

**Metrics:**
- Guild creation: <100ms
- Member add/remove: <50ms
- Permission check: <1ms
- Cross-server sync: <500ms per guild update
- Memory: <50KB per guild (100 members)

**Success Criteria:**
- [x] Players can create guilds with procedural names/emblems
- [x] Guild membership persists across sessions
- [x] Permission system prevents unauthorized actions
- [x] Guilds function across federated servers
- [x] Support for 1000+ guilds per server cluster

### 51.2: Guild Emblems & Identity

**Deliverables:**
- [ ] Procedural guild emblem generation
- [ ] 10 base shapes (shield, circle, triangle, hexagon, star, etc.)
- [ ] 20 symbol types (animal, weapon, element, celestial, abstract)
- [ ] Color scheme generation (2-4 colors, complementary/triadic)
- [ ] Emblem display on guild halls, member nameplates, banners
- [ ] Manual emblem editor (optional, selects from procedural options)

**Emblem Generation:**
```go
// pkg/procgen/guild/emblem.go
type EmblemGenerator struct {
    shapes  []ShapeTemplate
    symbols []SymbolTemplate
}

func (g *EmblemGenerator) Generate(seed int64, params GenerationParams) (*GuildEmblem, error) {
    rng := rand.New(rand.NewSource(seed))
    
    // 1. Select base shape
    shape := g.shapes[rng.Intn(len(g.shapes))]
    
    // 2. Select symbol (animal, weapon, element, etc.)
    symbol := g.symbols[rng.Intn(len(g.symbols))]
    
    // 3. Generate color scheme
    colors := g.generateColorScheme(params.GenreID, rng)
    
    // 4. Add optional border/flourishes based on guild prestige
    decorations := g.generateDecorations(params.Depth, rng)
    
    return &GuildEmblem{
        Shape:       shape,
        Symbol:      symbol,
        Colors:      colors,
        Decorations: decorations,
    }, nil
}
```

**Metrics:**
- Emblem generation: <20ms
- Visual uniqueness: >98% distinct emblems (10×20×100 color schemes = 20,000 combinations)
- Rendering: <5ms per emblem
- Editor updates: <50ms

**Success Criteria:**
- [x] Emblems are visually distinct between guilds
- [x] Genre-appropriate color schemes and symbols
- [x] Emblems display on guild halls and member UI
- [x] Manual editor allows customization from procedural options

### 51.3: Guild Resource System

**Deliverables:**
- [ ] Guild treasury (shared gold pool)
- [ ] Resource contribution system (members donate items/gold)
- [ ] Withdrawal permissions (rank-based limits)
- [ ] Contribution tracking (leaderboard, rewards)
- [ ] Guild bank (shared storage, 100-500 slots based on upgrades)
- [ ] Resource allocation for guild projects (construction, upgrades)

**Resource Management:**
```go
// pkg/network/federation/guild/resources.go
type GuildResourceManager struct {
    guild         *Guild
    treasury      int
    bank          *Inventory  // Shared storage
    contributions map[PlayerID]*ContributionRecord
}

type ContributionRecord struct {
    PlayerID      string
    TotalGold     int
    TotalItems    int
    ItemValue     int  // Calculated based on rarity
    LastContrib   int64
    Rank          int  // Leaderboard position
}

func (m *GuildResourceManager) Contribute(playerID string, gold int, items []ItemID) error {
    // 1. Validate player is guild member
    // 2. Add gold to treasury
    // 3. Add items to guild bank
    // 4. Update contribution record
    // 5. Check for rank rewards (achievements)
}

func (m *GuildResourceManager) Withdraw(playerID string, gold int, items []ItemID) error {
    // 1. Check player rank permissions
    // 2. Validate withdrawal limits (e.g., max 1000 gold/day for Member rank)
    // 3. Deduct from treasury/bank
    // 4. Log transaction (audit trail)
}
```

**Metrics:**
- Contribution processing: <50ms
- Withdrawal validation: <10ms
- Bank access: <100ms (reuse inventory system)
- Leaderboard update: <20ms

**Success Criteria:**
- [x] Members can contribute gold and items to guild
- [x] Rank-based withdrawal limits enforced
- [x] Contribution leaderboard displays top contributors
- [x] Guild bank persists across server restarts
- [x] Audit log tracks all transactions

---

## Phase 52: Guild Halls & Shared Structures

**Focus:** Guild-owned buildings with cross-server presence  
**Status:** PLANNING

### 52.1: Guild Hall Construction

**Deliverables:**
- [ ] Guild hall building type (larger than player houses: 32×32 to 64×64 tiles)
- [ ] Multi-floor support (1-5 floors based on guild level)
- [ ] Shared construction system (members contribute materials)
- [ ] Construction phases: Foundation → Walls → Roof → Interior → Furnishing
- [ ] Cross-server guild hall synchronization
- [ ] Guild hall upgrade system (expand size, add floors, improve materials)

**Guild Hall Types:**
- **Guildhall:** Standard meeting hall with treasury room, message board
- **Fortress:** Defensive structure with walls, towers, armory
- **Manor:** Luxury residence with private quarters for officers
- **Workshop:** Crafting-focused with multiple stations
- **Archive:** Library and research center with book storage

**Construction System:**
```go
// pkg/world/housing/guild_construction.go
type GuildConstructionProject struct {
    ProjectID     string
    GuildID       string
    BuildingType  BuildingType
    ServerID      string
    Location      ChunkCoords
    Phase         ConstructionPhase
    Required      ResourceRequirement
    Contributed   ResourceContribution
    Progress      float64  // 0.0-1.0
    StartedAt     int64
    EstimatedCompletion int64
}

type ConstructionPhase int
const (
    PhaseFoundation ConstructionPhase = iota
    PhaseWalls
    PhaseRoof
    PhaseInterior
    PhaseFurnishing
    PhaseComplete
)

type ResourceRequirement struct {
    Wood      int
    Stone     int
    Metal     int
    Crystal   int
    Gold      int
}

func (p *GuildConstructionProject) AddContribution(playerID string, resources ResourceContribution) error {
    // 1. Validate player is guild member
    // 2. Add resources to project
    // 3. Update progress percentage
    // 4. Check if phase complete → advance to next phase
    // 5. Notify guild of milestone (UI popup)
}
```

**Metrics:**
- Construction validation: <50ms
- Progress update: <20ms
- Phase transition: <100ms
- Cross-server sync: <1s for guild hall state
- Memory: <10MB per active construction project

**Success Criteria:**
- [x] Guilds can initiate construction projects
- [x] Members contribute resources to advance construction
- [x] Construction phases visually update (partial walls → complete structure)
- [x] Completed guild halls persist across servers
- [x] Multi-floor guild halls function correctly

### 52.2: Guild Hall Features & Rooms

**Deliverables:**
- [ ] Dedicated room types: Treasury, Armory, Library, Hall, Meeting Room, Throne Room
- [ ] Guild-specific furniture (throne, banner stands, trophy displays)
- [ ] Meeting room with seating for 10-50 members
- [ ] Treasury vault with enhanced security (requires officer+ to access)
- [ ] Armory with weapon/armor racks (guild equipment loan system)
- [ ] Library with procedurally generated guild history books

**Special Rooms:**
```go
// pkg/procgen/building/guild_rooms.go
type GuildRoomTemplate struct {
    RoomType      GuildRoomType
    MinSize       Vector2
    MaxSize       Vector2
    RequiredItems []FurnitureType
    Functionality GuildFunction
}

var TreasuryTemplate = GuildRoomTemplate{
    RoomType:      RoomTreasury,
    MinSize:       Vector2{8, 8},
    MaxSize:       Vector2{16, 16},
    RequiredItems: []FurnitureType{Vault, GoldPile, Chest, Guard},
    Functionality: FunctionVaultAccess,
}

var ThroneRoomTemplate = GuildRoomTemplate{
    RoomType:      RoomThrone,
    MinSize:       Vector2{12, 12},
    MaxSize:       Vector2{24, 24},
    RequiredItems: []FurnitureType{Throne, Carpet, Banners, Pillars},
    Functionality: FunctionCeremonial,
}
```

**Metrics:**
- Room generation: <100ms per room
- Furniture placement: <200ms for fully furnished room
- Access control: <5ms per permission check
- Visual rendering: 60 FPS with 100+ furniture items

**Success Criteria:**
- [x] 6 guild-specific room types with unique functionality
- [x] Treasury room restricts access based on rank
- [x] Meeting rooms accommodate large gatherings
- [x] Guild history books generate procedural lore
- [x] Equipment loan system tracks borrowed items

### 52.3: Cross-Server Guild Hall Sync

**Deliverables:**
- [ ] Guild hall state synchronization across federated servers
- [ ] Furniture placement replication
- [ ] Permission matrix synchronization
- [ ] Construction progress synchronization
- [ ] Conflict resolution (concurrent edits from different servers)
- [ ] Guild hall visitation system (players travel to guild halls on other servers)

**Synchronization Protocol:**
```go
// pkg/network/federation/guild/sync.go
type GuildHallSyncManager struct {
    federation    *FederationClient
    localGuilds   map[GuildID]*GuildHall
    syncInterval  time.Duration  // Default: 60 seconds
}

func (m *GuildHallSyncManager) SyncGuildHall(guildID GuildID) error {
    // 1. Serialize guild hall state (building, furniture, permissions)
    state := m.serializeGuildHall(guildID)
    
    // 2. Compute state hash for change detection
    hash := computeHash(state)
    
    // 3. Send sync request to all servers hosting guild members
    for _, serverID := range m.getGuildServers(guildID) {
        m.sendSyncRequest(serverID, guildID, hash, state)
    }
    
    // 4. Apply received updates from other servers
    m.mergeIncomingUpdates(guildID)
}

func (m *GuildHallSyncManager) MergeIncomingUpdates(guildID GuildID) error {
    // Conflict resolution: Last-write-wins with server timestamp
    // For concurrent furniture placement: Both items placed if no collision
    // For conflicting edits: Higher-rank player's edit wins
}
```

**Metrics:**
- Sync interval: 60 seconds (configurable)
- State serialization: <200ms per guild hall
- Conflict detection: <50ms
- Merge operation: <100ms
- Network overhead: <10KB per guild hall per sync

**Success Criteria:**
- [x] Guild halls sync across all servers with guild members
- [x] Furniture changes replicate within 60 seconds
- [x] Concurrent edits resolve without data loss
- [x] Players on different servers see consistent guild hall state
- [x] Network bandwidth <10KB/s per guild for sync traffic

---

## Phase 53: Territory Control & Guild Warfare

**Focus:** Guild territory ownership, contested zones, and competitive mechanics  
**Status:** PLANNING

### 53.1: Territory System

**Deliverables:**
- [ ] Territory zones on world map (5×5 chunk territories)
- [ ] Territory ownership assignment to guilds
- [ ] Territory benefits: +10% resource spawn, +5% XP gain, access to special NPCs
- [ ] Territory maintenance cost (gold/resources per week)
- [ ] Territory capture mechanics (claim unclaimed, challenge owned)
- [ ] Cross-server territory registry

**Territory Zones:**
```go
// pkg/world/territory.go
type Territory struct {
    TerritoryID   string
    Bounds        Rect          // World coordinates
    OwnerGuildID  string
    ClaimedAt     int64
    LastMaintenance int64
    ControlPoints []*ControlPoint
    Benefits      TerritoryBenefits
    MaintenanceCost ResourceRequirement
}

type ControlPoint struct {
    Position      Vector2
    CaptureProgress float64  // 0.0-1.0
    CapturingGuild string
    Defenders     []PlayerID
}

type TerritoryBenefits struct {
    ResourceBonus float64  // +10% spawn rate
    XPBonus       float64  // +5% XP gain
    NPCAccess     []NPCType // Special merchants
}
```

**Capture Mechanics:**
1. **Unclaimed Territory:** Guild members stand on control point for 5 minutes
2. **Owned Territory:** Challenge initiated, 24-hour preparation period, capture window
3. **Capture Window:** 2-hour period where control points can be captured
4. **Victory:** Control majority (3/5) control points at end of window

**Metrics:**
- Territory load: <50ms for 100 territories
- Capture progress update: <10ms per tick
- Benefit calculation: <5ms per player entering territory
- Maintenance check: <100ms per territory (weekly batch process)

**Success Criteria:**
- [x] Guilds can claim unclaimed territories
- [x] Territory ownership provides tangible benefits
- [x] Maintenance cost prevents territory hoarding
- [x] Capture mechanics balance offense and defense
- [x] Cross-server territories function via federation

### 53.2: Guild Warfare & Conflict

**Deliverables:**
- [ ] Guild war declaration system (mutual consent or territory challenge)
- [ ] War objectives: Capture territory, raid guild hall, deplete treasury
- [ ] War duration: 1-7 days (configurable by mutual agreement)
- [ ] Participation rewards (contribution points, achievements)
- [ ] Guild alliances (allied guilds can assist in wars)
- [ ] Peace treaties (end war early, negotiate terms)

**War System:**
```go
// pkg/network/federation/guild/warfare.go
type GuildWar struct {
    WarID         string
    Attacker      GuildID
    Defender      GuildID
    DeclaredAt    int64
    StartTime     int64  // 24 hours after declaration
    EndTime       int64
    Objectives    []WarObjective
    Participants  map[PlayerID]*WarParticipation
    Allies        map[GuildID][]GuildID
    Result        WarResult
}

type WarObjective struct {
    ObjectiveType ObjectiveType
    TargetID      string  // TerritoryID, BuildingID, etc.
    Progress      float64
    Completed     bool
    Reward        int  // Contribution points
}

type WarParticipation struct {
    PlayerID      string
    GuildID       GuildID
    Kills         int
    Deaths        int
    Objectives    int  // Completed objectives
    Contribution  int  // Points earned
}
```

**Objectives:**
- **Capture Territory:** Control majority of control points
- **Raid Guild Hall:** Reach treasury vault, steal percentage of gold
- **Deplete Treasury:** Reduce enemy guild treasury to 0 (via raids)
- **Kill Count:** Eliminate 100 enemy guild members
- **Hold Position:** Defend territory for 24 hours

**Metrics:**
- War declaration: <100ms
- Objective update: <50ms
- Participation tracking: <10ms per combat event
- Alliance coordination: <1s for multi-server communication

**Success Criteria:**
- [x] Guilds can declare war with mutual consent or territory challenge
- [x] War objectives provide clear victory conditions
- [x] Participation rewards incentivize member involvement
- [x] Alliances enable multi-guild cooperation
- [x] Peace treaties allow early war conclusion

### 53.3: Territory Defense & Fortifications

**Deliverables:**
- [ ] Defensive structures: Walls, towers, gates, traps
- [ ] NPC guards (hired with guild gold, procedurally generated stats)
- [ ] Territory patrol system (guards move along defined paths)
- [ ] Alarm system (notify guild members of attack)
- [ ] Fortification upgrades (stone walls → metal walls → energy barriers)
- [ ] Siege equipment (catapults, rams, trebuchets for attackers)

**Defensive Structures:**
```go
// pkg/world/territory/fortifications.go
type Fortification struct {
    FortID        string
    TerritoryID   string
    StructureType FortificationType
    Position      Vector2
    Health        float64
    MaxHealth     float64
    UpgradeLevel  int
    MaintenanceCost int
}

type FortificationType int
const (
    FortWall FortificationType = iota
    FortTower
    FortGate
    FortTrap
    FortBarricade
)

type TerritoryGuard struct {
    GuardID       string
    TerritoryID   string
    GuardType     EntityType  // Archer, Swordsman, Mage, etc.
    Stats         EntityStats
    PatrolPath    []Vector2
    HireCost      int
    MaintCost     int  // Per week
}
```

**Metrics:**
- Structure placement: <50ms
- Guard AI update: <5ms per guard (reuse existing AI systems)
- Alarm notification: <1s to all online guild members
- Siege damage: <10ms per hit

**Success Criteria:**
- [x] Guilds can build defensive structures in owned territories
- [x] NPC guards patrol and engage attackers
- [x] Alarm system notifies members of territory attacks
- [x] Fortifications provide meaningful defense bonuses
- [x] Siege equipment enables attackers to breach defenses

---

## Phase 54: Housing & Guild Polish

**Focus:** Quality of life, visual enhancements, and system integration  
**Status:** PLANNING

### 54.1: Housing Quality of Life

**Deliverables:**
- [ ] Blueprint system (save/load furniture layouts)
- [ ] Quick-build templates (pre-designed room layouts)
- [ ] Furniture search and filter in placement UI
- [ ] Bulk furniture operations (move multiple items, delete room contents)
- [ ] Guest access system (invite players to your house)
- [ ] Housing showcase mode (display house on server leaderboard)
- [ ] Rent system (pay gold/week to NPC landlords for premium plots)

**Blueprint System:**
```go
// pkg/world/housing/blueprint.go
type Blueprint struct {
    BlueprintID   string
    Name          string
    CreatorID     string
    RoomType      RoomType
    Dimensions    Vector2
    Furniture     []PlacedFurniture
    CreatedAt     int64
    Uses          int  // Popularity counter
}

type PlacedFurniture struct {
    FurnitureType FurnitureType
    Position      Vector2
    Rotation      int
    Material      MaterialType
    Rarity        Rarity
}

func (b *Blueprint) Apply(room *Room) error {
    // 1. Validate dimensions match
    // 2. Clear existing furniture (optional)
    // 3. Spawn furniture from blueprint
    // 4. Increment use counter
}
```

**Quick-Build Templates:**
- **Bedroom:** Bed, nightstand, wardrobe, rug, lamp
- **Kitchen:** Table, chairs, cabinets, stove, storage
- **Workshop:** Workbench, tool rack, storage, lighting
- **Library:** Bookshelves, reading chair, desk, lamp
- **Armory:** Weapon racks, armor stands, chest, display cases

**Metrics:**
- Blueprint save: <100ms
- Blueprint apply: <500ms (depends on furniture count)
- Template instantiation: <200ms
- Search performance: <50ms for 1000 items

**Success Criteria:**
- [x] Players can save and share blueprints
- [x] Quick-build templates accelerate decoration
- [x] Bulk operations save time for large buildings
- [x] Guest system enables house tours
- [x] Showcase mode displays top-rated houses

### 54.2: Visual Enhancements

**Deliverables:**
- [ ] Enhanced building exterior rendering (window glow at night, smoke from chimneys)
- [ ] Interior lighting with shadows from furniture
- [ ] Weather effects on buildings (rain on roof, snow accumulation)
- [ ] Seasonal decoration support (holiday themes)
- [ ] Building damage states (cracks, broken windows for abandoned houses)
- [ ] Procedural landscaping (gardens, paths, fences)

**Visual Features:**
```go
// pkg/rendering/housing/visuals.go
type BuildingVisualEffects struct {
    WindowGlow    bool
    ChimneySmoke  bool
    RoofWeather   WeatherEffect
    DamageState   float64  // 0.0 = pristine, 1.0 = ruined
    Landscaping   *Landscape
}

type Landscape struct {
    Gardens       []Garden
    Paths         []Path
    Fences        []Fence
    Trees         []Tree
}
```

**Metrics:**
- Exterior rendering: 60 FPS with 100 buildings visible
- Interior shadows: <5ms per room
- Weather effects: <2ms per building
- Damage rendering: <1ms per building

**Success Criteria:**
- [x] Buildings have realistic visual details
- [x] Interior lighting creates atmosphere
- [x] Weather effects enhance immersion
- [x] Landscaping improves exterior aesthetics
- [x] All visual enhancements maintain 60 FPS

### 54.3: System Integration & Testing

**Deliverables:**
- [ ] Integration with V6.0 federation (housing/guild sync)
- [ ] Integration with existing crafting system (furniture crafting)
- [ ] Integration with quest system (building quests, guild quests)
- [ ] Integration with trading system (furniture marketplace)
- [ ] Comprehensive test suite (unit, integration, acceptance tests)
- [ ] Performance benchmarks (1000 houses, 100 guilds, 10,000 furniture items)
- [ ] Documentation updates (user guide, API reference, architecture docs)

**Integration Points:**
- **Federation:** Housing/guild state sync across servers
- **Crafting:** Furniture recipes, building materials
- **Quests:** "Build your first house", "Join a guild", "Contribute 1000 gold to guild"
- **Trading:** Furniture marketplace, blueprint selling

**Test Coverage Targets:**
- `pkg/world/housing/`: ≥75% (higher for persistence)
- `pkg/network/federation/guild/`: ≥70%
- `pkg/procgen/building/`: ≥80%
- `pkg/procgen/furniture/`: ≥75%
- Overall: Maintain ≥65% average

**Performance Benchmarks:**
```go
// pkg/world/housing/housing_benchmark_test.go
func BenchmarkHousingLoad(b *testing.B) {
    // Load 1000 houses from disk
}

func BenchmarkFurniturePlacement(b *testing.B) {
    // Place 10,000 furniture items with collision detection
}

func BenchmarkGuildSync(b *testing.B) {
    // Sync 100 guilds across 5 servers
}
```

**Metrics:**
- Housing load: <5s for 1000 houses
- Furniture placement: <100ms for 100 items
- Guild sync: <10s for 100 guilds across 5 servers
- Memory: <200MB for all housing/guild data

**Success Criteria:**
- [x] All systems integrate without conflicts
- [x] Test coverage meets targets (≥65% overall)
- [x] Performance benchmarks pass
- [x] Documentation complete and accurate
- [x] No critical bugs in housing/guild systems

---

## Success Criteria

**Functionality:**
- [x] Players can build and customize houses with procedural generation
- [x] Guilds can be created, managed, and span multiple federated servers
- [x] Guild halls provide shared spaces for members
- [x] Territory system enables competitive guild gameplay
- [x] All features 100% procedural (zero external assets)

**Performance:**
- [x] 60 FPS minimum with 100 buildings visible
- [x] <500MB total memory (housing + guild + existing systems)
- [x] <100ms house save/load
- [x] <1s guild sync across servers
- [x] <25KB/s network overhead for housing/guild federation

**Quality:**
- [x] ≥65% test coverage per package (target ≥75% for housing/guild core)
- [x] All tests pass with race detection
- [x] Deterministic generation (same seed = identical building)
- [x] Cross-platform builds (desktop, web, mobile)
- [x] Backward compatible with v7.0 saves

**User Experience:**
- [x] Intuitive housing placement and decoration UI
- [x] Guild management accessible to non-technical players
- [x] Territory warfare balanced and engaging
- [x] Visual quality matches v7.0 standards
- [x] Multiplayer synchronization seamless (high-latency support maintained)

---

## Risk Mitigation

**Performance Risks:**
- **Risk:** Rendering 100+ buildings with furniture (10,000+ entities) tanks FPS
- **Mitigation:** Aggressive LOD (simplified models beyond 50 tiles), culling, instanced rendering for identical furniture
- **Fallback:** Reduce max concurrent buildings per view to 50, add quality settings

**Persistence Risks:**
- **Risk:** Housing save files grow to hundreds of MB (storage and load time issues)
- **Mitigation:** Incremental saves (only changed furniture), compression (gzip), lazy loading (load rooms as player enters)
- **Fallback:** Limit furniture per building (500 items max), periodic cleanup of inactive houses

**Synchronization Risks:**
- **Risk:** Cross-server guild hall sync causes conflicts and data loss
- **Mitigation:** Last-write-wins with timestamp, conflict-free replicated data types (CRDTs) for furniture placement
- **Fallback:** Guild halls tied to single "home" server, read-only replicas on other servers

**Gameplay Balance Risks:**
- **Risk:** Territory warfare dominated by largest guilds, small guilds excluded
- **Mitigation:** Territory size limits per guild (max 5 territories), alliance caps (max 3 allied guilds), diminishing returns on large territories
- **Fallback:** Separate territory tiers (small/medium/large), small guilds compete in small tier only

**Technical Complexity:**
- **Risk:** Multi-floor buildings complicate pathfinding and collision
- **Mitigation:** Reuse existing multi-layer terrain from Phase 11, restrict player movement to one floor at a time (stairs trigger floor transition)
- **Fallback:** Limit buildings to single floor initially, add multi-floor in v8.1

---

## Dependencies

**Existing Systems (Must Be Complete):**
- Phase 43-48 (v7.0): Display scaling, 64x64 sprites, anti-aliasing, sub-pixel collision
- Phase 37-42 (v6.0): World persistence, federation, cross-server travel, authentication
- Phase 3: Viewport culling, batch rendering (reused for building rendering)
- Phase 15: Enhanced sprites (basis for furniture rendering)
- Phase 17: Lighting system (interior building lighting)

**New Packages:**
- `pkg/world/housing/`: Housing manager, plot system, persistence
- `pkg/world/territory/`: Territory management, control points, capture mechanics
- `pkg/procgen/building/`: Building generator, floor plans, architecture styles
- `pkg/procgen/furniture/`: Furniture generator, placement validation
- `pkg/network/federation/guild/`: Guild manager, permissions, resource system, sync protocol

**Build Requirements:**
- Go 1.24.5+ (maintained from v7.0)
- Ebiten v2.9.3+
- Minimum 16GB RAM for development (building generation tests)
- 1920x1080 display for visual testing

**Testing:**
- Visual regression suite for buildings and furniture
- Multi-server integration tests (3+ federated servers)
- Load testing: 100 guilds, 1000 houses, 10,000 furniture items
- Race detection for concurrent guild operations
- Save/load validation for all housing data

---

## Timeline & Milestones

**Month 1-2: Housing Foundation (Phase 49)**
- Week 1-2: Housing infrastructure and plot system
- Week 3-4: Building generation and templates
- Week 5-6: Interior generation and room layouts
- Week 7-8: Testing and refinement

**Month 3-4: Furniture System (Phase 50)**
- Week 9-10: Furniture generation and templates
- Week 11-12: Placement and interaction systems
- Week 13-14: Decoration and customization
- Week 15-16: Testing and visual polish

**Month 5-6: Guild Foundation (Phase 51)**
- Week 17-18: Guild creation and management
- Week 19-20: Emblem generation and identity
- Week 21-22: Resource system
- Week 23-24: Testing and multiplayer sync

**Month 7-8: Guild Halls (Phase 52)**
- Week 25-26: Guild hall construction
- Week 27-28: Guild hall features and rooms
- Week 29-30: Cross-server synchronization
- Week 31-32: Testing and conflict resolution

**Month 9-10: Territory System (Phase 53)**
- Week 33-34: Territory ownership and benefits
- Week 35-36: Guild warfare mechanics
- Week 37-38: Defense and fortifications
- Week 39-40: Testing and balance tuning

**Month 11-12: Polish & Integration (Phase 54)**
- Week 41-42: Quality of life features
- Week 43-44: Visual enhancements
- Week 45-46: System integration and testing
- Week 47-48: Documentation, benchmarks, release preparation

**Contingency:** 2-week buffer for unexpected issues or scope adjustments

---

## Completion Checklist

**Features:**
- [ ] Player housing with procedural generation (5 building types)
- [ ] Guild system with cross-server support (1000+ guilds)
- [ ] Guild halls with shared construction (32×32 to 64×64 tiles)
- [ ] Territory control mechanics (5×5 chunk territories)
- [ ] Furniture system (30+ types, 5 rarity tiers)
- [ ] 60 FPS, <500MB memory maintained

**Quality:**
- [ ] ≥65% test coverage (target ≥75% for housing/guild core)
- [ ] All tests pass (unit, integration, acceptance)
- [ ] Cross-platform builds (desktop, web, mobile)
- [ ] v7.0 backward compatibility (save migration)
- [ ] Multiplayer sync (high-latency support maintained)

**Documentation:**
- [ ] Update ARCHITECTURE.md (housing/guild/territory packages)
- [ ] Update TECHNICAL_SPEC.md (persistence format, sync protocol)
- [ ] Update USER_MANUAL.md (housing guide, guild guide)
- [ ] Update API_REFERENCE.md (new components, systems, generators)
- [ ] Create RELEASE_NOTES_V8.0.md

**Release:**
- [ ] Version bump 7.0→8.0 in all files
- [ ] Build all platforms (Linux, macOS, Windows, WebAssembly, iOS, Android)
- [ ] Deploy WebAssembly to GitHub Pages
- [ ] Create release tag v8.0.0
- [ ] Publish release notes and migration guide

---

**Document Status:** Draft  
**Last Updated:** November 16, 2025  
**Next Review:** Upon Phase 48 completion (V7.0 release)
