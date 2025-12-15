// Package engine provides resource gathering components for the ECS.
// This file implements components for harvestable resource nodes and player gathering state.
// Phase 95: Resource Gathering System

package engine

import (
	"encoding/json"
	"sync"
)

// ResourceType defines the type of harvestable resource.
type ResourceType string

const (
	// ResourceTypeOre represents mineable ore deposits.
	ResourceTypeOre ResourceType = "ore"
	// ResourceTypeWood represents harvestable trees.
	ResourceTypeWood ResourceType = "wood"
	// ResourceTypeHerb represents gatherable plants.
	ResourceTypeHerb ResourceType = "herb"
	// ResourceTypeGem represents rare gem deposits.
	ResourceTypeGem ResourceType = "gem"
	// ResourceTypeFiber represents fiber-producing plants.
	ResourceTypeFiber ResourceType = "fiber"
	// ResourceTypeEssence represents magical essence nodes.
	ResourceTypeEssence ResourceType = "essence"
)

// AllResourceTypes returns all valid resource types.
func AllResourceTypes() []ResourceType {
	return []ResourceType{
		ResourceTypeOre,
		ResourceTypeWood,
		ResourceTypeHerb,
		ResourceTypeGem,
		ResourceTypeFiber,
		ResourceTypeEssence,
	}
}

// ToolType defines the tool required for gathering.
type ToolType string

const (
	// ToolTypePickaxe for mining ore and gems.
	ToolTypePickaxe ToolType = "pickaxe"
	// ToolTypeAxe for chopping wood.
	ToolTypeAxe ToolType = "axe"
	// ToolTypeSickle for harvesting herbs and fiber.
	ToolTypeSickle ToolType = "sickle"
	// ToolTypeStaff for gathering essence.
	ToolTypeStaff ToolType = "staff"
	// ToolTypeNone for hand-gathering.
	ToolTypeNone ToolType = "none"
)

// RequiredToolForResource returns the tool needed for a resource type.
func RequiredToolForResource(rt ResourceType) ToolType {
	switch rt {
	case ResourceTypeOre:
		return ToolTypePickaxe
	case ResourceTypeWood:
		return ToolTypeAxe
	case ResourceTypeHerb:
		return ToolTypeSickle
	case ResourceTypeGem:
		return ToolTypePickaxe
	case ResourceTypeFiber:
		return ToolTypeSickle
	case ResourceTypeEssence:
		return ToolTypeStaff
	default:
		return ToolTypeNone
	}
}

// ResourceNodeComponent represents a harvestable resource node in the world.
// It tracks resource availability, respawn timing, and gathering requirements.
type ResourceNodeComponent struct {
	mu sync.RWMutex

	// ResourceType is the type of resource (ore, wood, herb, etc.)
	ResourceType ResourceType `json:"resource_type"`

	// Quantity is the remaining number of harvests available.
	Quantity int `json:"quantity"`

	// MaxQuantity is the maximum harvests when fully respawned.
	MaxQuantity int `json:"max_quantity"`

	// RespawnTime is the seconds until respawn begins after depletion.
	RespawnTime float64 `json:"respawn_time"`

	// RespawnTimer tracks countdown to next respawn tick.
	RespawnTimer float64 `json:"respawn_timer"`

	// RequiredTool is the tool type needed to harvest.
	RequiredTool ToolType `json:"required_tool"`

	// MinSkillLevel is the minimum gathering skill required.
	MinSkillLevel int `json:"min_skill_level"`

	// BiomeType is the biome where this node naturally spawns.
	BiomeType string `json:"biome_type"`

	// YieldMin is the minimum yield per harvest.
	YieldMin int `json:"yield_min"`

	// YieldMax is the maximum yield per harvest.
	YieldMax int `json:"yield_max"`

	// IsDepleted indicates the node is exhausted.
	IsDepleted bool `json:"is_depleted"`
}

// NewResourceNodeComponent creates a new resource node with default values.
func NewResourceNodeComponent(resourceType ResourceType, biomeType string) *ResourceNodeComponent {
	return &ResourceNodeComponent{
		ResourceType:  resourceType,
		Quantity:      3,
		MaxQuantity:   3,
		RespawnTime:   300.0, // 5 minutes default
		RespawnTimer:  0,
		RequiredTool:  RequiredToolForResource(resourceType),
		MinSkillLevel: 1,
		BiomeType:     biomeType,
		YieldMin:      1,
		YieldMax:      3,
		IsDepleted:    false,
	}
}

// Type returns the component type identifier.
func (r *ResourceNodeComponent) Type() string {
	return "resource_node"
}

// CanHarvest checks if the node can be harvested with given skill level.
func (r *ResourceNodeComponent) CanHarvest(skillLevel int) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return !r.IsDepleted && r.Quantity > 0 && skillLevel >= r.MinSkillLevel
}

// Harvest attempts to harvest from the node, returns true if successful.
func (r *ResourceNodeComponent) Harvest() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.IsDepleted || r.Quantity <= 0 {
		return false
	}

	r.Quantity--
	if r.Quantity <= 0 {
		r.IsDepleted = true
		r.RespawnTimer = r.RespawnTime
	}
	return true
}

// UpdateRespawn processes respawn timer, returns true if node respawned.
func (r *ResourceNodeComponent) UpdateRespawn(deltaTime float64) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.IsDepleted {
		return false
	}

	r.RespawnTimer -= deltaTime
	if r.RespawnTimer <= 0 {
		r.Quantity = r.MaxQuantity
		r.IsDepleted = false
		r.RespawnTimer = 0
		return true
	}
	return false
}

// GetQuantity returns the current quantity safely.
func (r *ResourceNodeComponent) GetQuantity() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Quantity
}

// GetRespawnProgress returns respawn progress from 0.0 to 1.0.
func (r *ResourceNodeComponent) GetRespawnProgress() float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if !r.IsDepleted || r.RespawnTime <= 0 {
		return 1.0
	}
	progress := 1.0 - (r.RespawnTimer / r.RespawnTime)
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

// Serialize converts the component to JSON bytes.
func (r *ResourceNodeComponent) Serialize() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return json.Marshal(r)
}

// Deserialize loads the component from JSON bytes.
func (r *ResourceNodeComponent) Deserialize(data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return json.Unmarshal(data, r)
}

// GatheringComponent tracks a player's gathering state and skills.
type GatheringComponent struct {
	mu sync.RWMutex

	// GatheringSkill is the player's gathering skill level (1-100).
	GatheringSkill int `json:"gathering_skill"`

	// GatheringXP tracks experience toward next skill level.
	GatheringXP int `json:"gathering_xp"`

	// ToolBonuses maps tool types to yield multipliers.
	ToolBonuses map[ToolType]float64 `json:"tool_bonuses"`

	// EquippedTool is the currently equipped gathering tool.
	EquippedTool ToolType `json:"equipped_tool"`

	// IsGathering indicates actively gathering a resource.
	IsGathering bool `json:"is_gathering"`

	// GatherProgress tracks gathering completion (0.0-1.0).
	GatherProgress float64 `json:"gather_progress"`

	// TargetNodeID is the entity ID of the target resource node.
	TargetNodeID uint64 `json:"target_node_id"`

	// GatherSpeed is base gather speed multiplier.
	GatherSpeed float64 `json:"gather_speed"`

	// TotalHarvested tracks lifetime harvest count per type.
	TotalHarvested map[ResourceType]int `json:"total_harvested"`
}

// NewGatheringComponent creates a new gathering component with defaults.
func NewGatheringComponent() *GatheringComponent {
	return &GatheringComponent{
		GatheringSkill: 1,
		GatheringXP:    0,
		ToolBonuses:    make(map[ToolType]float64),
		EquippedTool:   ToolTypeNone,
		IsGathering:    false,
		GatherProgress: 0,
		TargetNodeID:   0,
		GatherSpeed:    1.0,
		TotalHarvested: make(map[ResourceType]int),
	}
}

// Type returns the component type identifier.
func (g *GatheringComponent) Type() string {
	return "gathering"
}

// StartGathering begins gathering from a target node.
func (g *GatheringComponent) StartGathering(nodeID uint64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.IsGathering = true
	g.GatherProgress = 0
	g.TargetNodeID = nodeID
}

// StopGathering cancels the current gathering action.
func (g *GatheringComponent) StopGathering() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.IsGathering = false
	g.GatherProgress = 0
	g.TargetNodeID = 0
}

// GetTargetNodeID returns the target node ID safely.
func (g *GatheringComponent) GetTargetNodeID() uint64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.TargetNodeID
}

// UpdateProgress advances gathering progress, returns true if complete.
func (g *GatheringComponent) UpdateProgress(deltaTime, baseTime float64) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if !g.IsGathering {
		return false
	}

	// Calculate progress per second based on skill and speed
	progressRate := g.GatherSpeed * (1.0 + float64(g.GatheringSkill)*0.01)
	if baseTime > 0 {
		progressRate = progressRate / baseTime
	}

	g.GatherProgress += progressRate * deltaTime
	if g.GatherProgress >= 1.0 {
		g.GatherProgress = 1.0
		return true
	}
	return false
}

// CompleteGathering finishes gathering and resets state.
func (g *GatheringComponent) CompleteGathering(resourceType ResourceType) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.IsGathering = false
	g.GatherProgress = 0
	g.TargetNodeID = 0
	g.TotalHarvested[resourceType]++
}

// AddXP adds gathering experience and handles level ups.
func (g *GatheringComponent) AddXP(amount int) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.GatheringXP += amount
	xpNeeded := g.xpForNextLevel()

	if g.GatheringXP >= xpNeeded && g.GatheringSkill < 100 {
		g.GatheringXP -= xpNeeded
		g.GatheringSkill++
		return true // Leveled up
	}
	return false
}

// xpForNextLevel calculates XP needed for next level.
func (g *GatheringComponent) xpForNextLevel() int {
	return 100 + (g.GatheringSkill * 50)
}

// GetSkillLevel returns current skill level safely.
func (g *GatheringComponent) GetSkillLevel() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.GatheringSkill
}

// GetToolBonus returns the bonus multiplier for a tool type.
func (g *GatheringComponent) GetToolBonus(tool ToolType) float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if bonus, ok := g.ToolBonuses[tool]; ok {
		return bonus
	}
	return 1.0
}

// SetToolBonus sets the bonus multiplier for a tool type.
func (g *GatheringComponent) SetToolBonus(tool ToolType, bonus float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.ToolBonuses[tool] = bonus
}

// EquipTool sets the currently equipped tool.
func (g *GatheringComponent) EquipTool(tool ToolType) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.EquippedTool = tool
}

// GetEquippedTool returns the currently equipped tool.
func (g *GatheringComponent) GetEquippedTool() ToolType {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.EquippedTool
}

// HasCorrectTool checks if equipped tool matches required tool.
func (g *GatheringComponent) HasCorrectTool(required ToolType) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if required == ToolTypeNone {
		return true
	}
	return g.EquippedTool == required
}

// IsCurrentlyGathering returns gathering state safely.
func (g *GatheringComponent) IsCurrentlyGathering() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.IsGathering
}

// GetProgress returns current gathering progress.
func (g *GatheringComponent) GetProgress() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.GatherProgress
}

// GetTotalHarvested returns lifetime harvest count for a type.
func (g *GatheringComponent) GetTotalHarvested(rt ResourceType) int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.TotalHarvested[rt]
}

// Serialize converts the component to JSON bytes.
func (g *GatheringComponent) Serialize() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return json.Marshal(g)
}

// Deserialize loads the component from JSON bytes.
func (g *GatheringComponent) Deserialize(data []byte) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return json.Unmarshal(data, g)
}
