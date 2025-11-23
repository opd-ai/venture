package raids

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/entity"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// Generator generates procedural raid dungeons.
type Generator struct {
	baseSeed    int64
	terrainGen  *terrain.BSPGenerator
	entityGen   *entity.EntityGenerator
	bossNameGen *BossNameGenerator
	mechanicGen *MechanicGenerator
}

// NewGenerator creates a new raid generator with the given base seed.
func NewGenerator(baseSeed int64) *Generator {
	return &Generator{
		baseSeed:    baseSeed,
		terrainGen:  terrain.NewBSPGenerator(),
		entityGen:   entity.NewEntityGenerator(),
		bossNameGen: NewBossNameGenerator(),
		mechanicGen: NewMechanicGenerator(),
	}
}

// Generate implements the procgen.Generator interface.
func (g *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if err := validateGenerationParams(params); err != nil {
		return nil, err
	}

	tier, groupID, groupSize := extractRaidParameters(params)

	instanceSeed := seed + int64(hashString(groupID))
	rng := rand.New(rand.NewSource(instanceSeed))

	raid, err := g.generateRaid(rng, tier, params, instanceSeed, groupSize)
	if err != nil {
		return nil, fmt.Errorf("generate raid: %w", err)
	}

	return raid, nil
}

// validateGenerationParams validates the generation parameters.
func validateGenerationParams(params procgen.GenerationParams) error {
	if params.Difficulty < 0.0 || params.Difficulty > 1.0 {
		return fmt.Errorf("difficulty must be 0.0-1.0, got %.2f", params.Difficulty)
	}
	if params.Depth < 1 {
		return fmt.Errorf("depth must be >= 1, got %d", params.Depth)
	}
	return nil
}

// extractRaidParameters extracts tier, groupID, and groupSize from params.
func extractRaidParameters(params procgen.GenerationParams) (RaidTier, string, int) {
	tier := extractTier(params)
	groupID := extractGroupID(params)
	groupSize := extractGroupSize(params, tier)
	return tier, groupID, groupSize
}

// extractTier extracts tier from custom params.
func extractTier(params procgen.GenerationParams) RaidTier {
	tier := TierNormal
	if tierVal, ok := params.Custom["tier"]; ok {
		if t, ok := tierVal.(RaidTier); ok {
			tier = t
		} else if tInt, ok := tierVal.(int); ok {
			tier = RaidTier(tInt)
		}
	}
	return tier
}

// extractGroupID extracts group ID from custom params.
func extractGroupID(params procgen.GenerationParams) string {
	groupID := "default"
	if gid, ok := params.Custom["group_id"]; ok {
		if s, ok := gid.(string); ok {
			groupID = s
		}
	}
	return groupID
}

// extractGroupSize extracts group size from custom params.
func extractGroupSize(params procgen.GenerationParams, tier RaidTier) int {
	groupSize := tier.MinPlayers()
	if gs, ok := params.Custom["group_size"]; ok {
		if size, ok := gs.(int); ok {
			groupSize = size
		}
	}
	return groupSize
}

// Validate implements the procgen.Generator interface.
func (g *Generator) Validate(result interface{}) error {
	raid, ok := result.(*RaidDungeon)
	if !ok {
		return fmt.Errorf("result is not *RaidDungeon, got %T", result)
	}

	if err := validateRaidBosses(raid); err != nil {
		return err
	}

	if err := validateRaidTerrain(raid); err != nil {
		return err
	}

	if err := validateRaidRooms(raid); err != nil {
		return err
	}

	return nil
}

// validateRaidBosses checks boss count and mechanics.
func validateRaidBosses(raid *RaidDungeon) error {
	if len(raid.Bosses) < 3 || len(raid.Bosses) > 5 {
		return fmt.Errorf("raid must have 3-5 bosses, got %d", len(raid.Bosses))
	}

	for i, boss := range raid.Bosses {
		if len(boss.Mechanics) == 0 {
			return fmt.Errorf("boss %d has no mechanics", i)
		}
		if len(boss.Phases) == 0 {
			return fmt.Errorf("boss %d has no phases", i)
		}
	}

	return nil
}

// validateRaidTerrain checks terrain dimensions.
func validateRaidTerrain(raid *RaidDungeon) error {
	if raid.Terrain == nil {
		return fmt.Errorf("raid has no terrain")
	}

	if raid.Terrain.Width < 50 || raid.Terrain.Height < 50 {
		return fmt.Errorf("raid terrain too small: %dx%d", raid.Terrain.Width, raid.Terrain.Height)
	}

	return nil
}

// validateRaidRooms checks room count and entrance presence.
func validateRaidRooms(raid *RaidDungeon) error {
	if len(raid.Rooms) < 4 {
		return fmt.Errorf("raid must have at least 4 rooms, got %d", len(raid.Rooms))
	}

	hasEntrance := false
	for _, room := range raid.Rooms {
		if room.Type == RoomEntrance {
			hasEntrance = true
			break
		}
	}

	if !hasEntrance {
		return fmt.Errorf("raid has no entrance room")
	}

	return nil
}

// generateRaid creates a complete raid dungeon.
func (g *Generator) generateRaid(rng *rand.Rand, tier RaidTier, params procgen.GenerationParams, seed int64, groupSize int) (*RaidDungeon, error) {
	// Generate terrain layout
	terrainParams := procgen.GenerationParams{
		Difficulty: params.Difficulty * tier.DifficultyMultiplier(),
		Depth:      params.Depth,
		GenreID:    params.GenreID,
		Custom: map[string]interface{}{
			"width":      80 + (int(tier) * 10), // Larger for higher tiers
			"height":     60 + (int(tier) * 10),
			"room_count": 10 + (int(tier) * 2), // More rooms for higher tiers
			"room_size":  "large",
		},
	}

	terrainResult, err := g.terrainGen.Generate(seed, terrainParams)
	if err != nil {
		return nil, fmt.Errorf("generate terrain: %w", err)
	}
	raidTerrain := terrainResult.(*terrain.Terrain)

	// Generate rooms based on terrain
	rooms := g.generateRooms(rng, raidTerrain, tier)

	// Determine boss count (3-5 based on tier)
	bossCount := 3 + int(tier) // Normal: 3, Nightmare: 7, capped at 5
	if bossCount > 5 {
		bossCount = 5
	}

	// Generate bosses
	bosses := make([]*RaidBoss, 0, bossCount)
	bossRooms := filterRoomsByType(rooms, RoomBoss)

	for i := 0; i < bossCount && i < len(bossRooms); i++ {
		boss, err := g.generateBoss(rng, tier, params, i, bossRooms[i], groupSize)
		if err != nil {
			return nil, fmt.Errorf("generate boss %d: %w", i, err)
		}
		bosses = append(bosses, boss)
	}

	// Generate raid name
	raidName := g.generateRaidName(rng, params.GenreID, tier)

	raid := &RaidDungeon{
		ID:          fmt.Sprintf("raid-%d", seed),
		Name:        raidName,
		Description: fmt.Sprintf("A %s tier raid dungeon", tier.String()),
		Tier:        tier,
		Terrain:     raidTerrain,
		Bosses:      bosses,
		Rooms:       rooms,
		CreatedAt:   time.Now(),
		Seed:        seed,
	}

	return raid, nil
}

// generateRooms creates raid rooms from terrain.
func (g *Generator) generateRooms(rng *rand.Rand, t *terrain.Terrain, tier RaidTier) []*RaidRoom {
	rooms := make([]*RaidRoom, 0)

	// Always create entrance room (top-left quadrant)
	entrance := &RaidRoom{
		ID:          "room-entrance",
		Type:        RoomEntrance,
		X:           5,
		Y:           5,
		W:           10,
		H:           10,
		Connections: []string{},
	}
	rooms = append(rooms, entrance)

	// Create boss rooms (3-5 based on tier)
	bossCount := 3 + int(tier)
	if bossCount > 5 {
		bossCount = 5
	}

	for i := 0; i < bossCount; i++ {
		room := &RaidRoom{
			ID:     fmt.Sprintf("room-boss-%d", i),
			Type:   RoomBoss,
			X:      20 + (i * 15),
			Y:      20 + (rng.Intn(20) - 10),
			W:      12 + rng.Intn(8),
			H:      12 + rng.Intn(8),
			BossID: fmt.Sprintf("boss-%d", i),
		}
		rooms = append(rooms, room)
	}

	// Add trash rooms between bosses
	trashCount := 2 + int(tier)
	for i := 0; i < trashCount; i++ {
		room := &RaidRoom{
			ID:   fmt.Sprintf("room-trash-%d", i),
			Type: RoomTrash,
			X:    10 + (i * 12),
			Y:    10 + rng.Intn(20),
			W:    8 + rng.Intn(4),
			H:    8 + rng.Intn(4),
		}
		rooms = append(rooms, room)
	}

	// Add treasure room after final boss
	treasure := &RaidRoom{
		ID:   "room-treasure",
		Type: RoomTreasure,
		X:    t.Width - 20,
		Y:    t.Height/2 - 5,
		W:    15,
		H:    15,
	}
	rooms = append(rooms, treasure)

	// Add rest area midway
	rest := &RaidRoom{
		ID:   "room-rest",
		Type: RoomRest,
		X:    t.Width / 2,
		Y:    5,
		W:    10,
		H:    10,
	}
	rooms = append(rooms, rest)

	return rooms
}

// generateBoss creates a raid boss with mechanics.
func (g *Generator) generateBoss(rng *rand.Rand, tier RaidTier, params procgen.GenerationParams, index int, room *RaidRoom, groupSize int) (*RaidBoss, error) {
	// Scale difficulty for bosses (2x-10x normal mobs)
	bossDifficulty := params.Difficulty * tier.DifficultyMultiplier()

	// Generate base entity with scaled stats
	entityParams := procgen.GenerationParams{
		Difficulty: bossDifficulty,
		Depth:      params.Depth + 5 + index, // Each boss deeper
		GenreID:    params.GenreID,
		Custom: map[string]interface{}{
			"entity_type": "boss",
			"scale":       1.0 + (float64(tier) * 0.2), // Larger for higher tiers
			"count":       1,                           // Request single entity
		},
	}

	entitySeed := g.baseSeed + int64(index)*1000
	entityResult, err := g.entityGen.Generate(entitySeed, entityParams)
	if err != nil {
		return nil, fmt.Errorf("generate entity: %w", err)
	}
	entities := entityResult.([]*entity.Entity)
	if len(entities) == 0 {
		return nil, fmt.Errorf("entity generator returned empty slice")
	}
	bossEntity := entities[0]

	// Scale boss health for group size
	bossEntity.Stats.Health = int(float64(bossEntity.Stats.Health) * (1.0 + float64(groupSize-5)*0.1))
	bossEntity.Stats.MaxHealth = bossEntity.Stats.Health

	// Generate boss mechanics (3-7 based on tier)
	mechanicCount := 3 + int(tier)
	if mechanicCount > 7 {
		mechanicCount = 7
	}

	mechanics := make([]BossMechanic, 0, mechanicCount)
	for i := 0; i < mechanicCount; i++ {
		mechanic := g.mechanicGen.GenerateMechanic(rng, tier, i)
		mechanics = append(mechanics, mechanic)
	}

	// Generate boss phases (3 phases at 75%, 50%, 25% HP)
	phases := []BossPhase{
		{Number: 1, HealthThresh: 1.0, Mechanics: []string{mechanics[0].ID}, AddSpawns: 0},
		{Number: 2, HealthThresh: 0.75, Mechanics: []string{mechanics[0].ID, mechanics[1%len(mechanics)].ID}, AddSpawns: 2 + int(tier)},
		{Number: 3, HealthThresh: 0.50, Mechanics: []string{mechanics[0].ID, mechanics[1%len(mechanics)].ID, mechanics[2%len(mechanics)].ID}, AddSpawns: 3 + int(tier)},
	}
	if mechanicCount >= 4 {
		phases = append(phases, BossPhase{Number: 4, HealthThresh: 0.25, Mechanics: []string{mechanics[0].ID, mechanics[2%len(mechanics)].ID, mechanics[3].ID}, AddSpawns: 5 + int(tier)*2})
	}

	// Generate loot table
	lootTable := g.generateLootTable(rng, tier, index)

	// Position boss in center of room
	posX := float64(room.X + room.W/2)
	posY := float64(room.Y + room.H/2)

	boss := &RaidBoss{
		Entity:    bossEntity,
		RoomID:    room.ID,
		Mechanics: mechanics,
		Phases:    phases,
		Position:  Position{X: posX, Y: posY},
		LootTable: lootTable,
	}

	return boss, nil
}

// generateLootTable creates a loot table for a boss.
func (g *Generator) generateLootTable(rng *rand.Rand, tier RaidTier, bossIndex int) *LootTable {
	// Guaranteed items increase with tier
	guaranteed := 1 + int(tier)/2

	// Possible items (each boss can drop 5-10 different items)
	possibleCount := 5 + rng.Intn(6)
	possibleItems := make([]LootItem, possibleCount)

	rarities := []string{"Uncommon", "Rare", "Epic", "Legendary", "Mythic"}
	tierRarity := int(tier)
	if tierRarity >= len(rarities) {
		tierRarity = len(rarities) - 1
	}

	for i := 0; i < possibleCount; i++ {
		// Higher tier bosses drop better loot
		rarityIdx := tierRarity
		if i == 0 {
			// First item can be even better
			rarityIdx = min(tierRarity+1, len(rarities)-1)
		}

		dropRate := 0.1 + (float64(tier) * 0.05) // 10-30% drop rate
		if i == 0 {
			dropRate = 0.8 // Guaranteed near-drop for best item
		}

		possibleItems[i] = LootItem{
			ItemID:   fmt.Sprintf("raid-item-%s-%d-%d", tier.String(), bossIndex, i),
			Rarity:   rarities[rarityIdx],
			DropRate: dropRate,
		}
	}

	// Currency scales with tier
	currencyBase := 1000 * (1 + int(tier))

	return &LootTable{
		GuaranteedItems: guaranteed,
		PossibleItems:   possibleItems,
		CurrencyMin:     currencyBase,
		CurrencyMax:     currencyBase * 2,
	}
}

// generateRaidName creates a procedural raid name.
func (g *Generator) generateRaidName(rng *rand.Rand, genreID string, tier RaidTier) string {
	return g.bossNameGen.GenerateRaidName(rng, genreID, tier)
}

// filterRoomsByType returns rooms of a specific type.
func filterRoomsByType(rooms []*RaidRoom, roomType RoomType) []*RaidRoom {
	filtered := make([]*RaidRoom, 0)
	for _, room := range rooms {
		if room.Type == roomType {
			filtered = append(filtered, room)
		}
	}
	return filtered
}

// hashString creates a simple hash of a string for seed derivation.
func hashString(s string) int {
	h := 0
	for i := 0; i < len(s); i++ {
		h = h*31 + int(s[i])
	}
	return h
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
