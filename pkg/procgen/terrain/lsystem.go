// Package terrain provides procedural terrain generation using various algorithms including L-systems.
package terrain

import (
	"math/rand"
	"strings"
)

// Symbol represents a single symbol in an L-system string.
type Symbol rune

// L-system symbols for dungeon generation
const (
	SymbolStart     Symbol = 'S' // Starting room (entrance)
	SymbolEnd       Symbol = 'E' // Ending room (boss/exit)
	SymbolCombat    Symbol = 'C' // Combat room
	SymbolTreasure  Symbol = 'T' // Treasure room
	SymbolPuzzle    Symbol = 'P' // Puzzle room
	SymbolCorridor  Symbol = '-' // Corridor/hallway
	SymbolBranch    Symbol = '+' // Branch point
	SymbolShop      Symbol = '$' // Shop/merchant room
	SymbolRest      Symbol = 'R' // Rest/safe room
	SymbolSecret    Symbol = '?' // Secret room
	SymbolEmpty     Symbol = '.' // Empty/optional room
)

// ProductionRule defines a rewrite rule for L-system generation.
type ProductionRule struct {
	// From is the symbol to replace
	From Symbol
	// To is the string of symbols to replace it with
	To string
	// Weight is the probability weight (for stochastic L-systems)
	Weight float64
}

// LSystemConfig defines parameters for L-system generation.
type LSystemConfig struct {
	// Axiom is the starting string
	Axiom string
	// Rules are the production rules
	Rules []ProductionRule
	// Iterations is the number of generation iterations
	Iterations int
	// Seed is the random seed for stochastic rules
	Seed int64
	// MinRoomCount is the minimum number of rooms to generate
	MinRoomCount int
	// MaxRoomCount is the maximum number of rooms to generate
	MaxRoomCount int
}

// LSystemGenerator generates dungeon layouts using L-systems.
type LSystemGenerator struct {
	config LSystemConfig
	rng    *rand.Rand
}

// NewLSystemGenerator creates a new L-system generator with the given configuration.
func NewLSystemGenerator(config LSystemConfig) *LSystemGenerator {
	return &LSystemGenerator{
		config: config,
		rng:    rand.New(rand.NewSource(config.Seed)),
	}
}

// Generate produces a dungeon layout string using L-system rewriting.
func (g *LSystemGenerator) Generate() string {
	current := g.config.Axiom

	for i := 0; i < g.config.Iterations; i++ {
		// Check room count BEFORE iterating
		roomCount := g.countRooms(current)
		if roomCount >= g.config.MaxRoomCount {
			break
		}

		current = g.iterate(current)

		// Check again after iteration
		roomCount = g.countRooms(current)
		if roomCount >= g.config.MaxRoomCount {
			break
		}
	}

	// Ensure we have at least the minimum number of rooms
	roomCount := g.countRooms(current)
	if roomCount < g.config.MinRoomCount {
		// Add more iterations if needed
		for roomCount < g.config.MinRoomCount && len(current) < 1000 {
			current = g.iterate(current)
			roomCount = g.countRooms(current)
			
			// Stop if we exceed max rooms
			if roomCount >= g.config.MaxRoomCount {
				break
			}
		}
	}

	return current
}

// iterate performs one iteration of L-system rewriting.
func (g *LSystemGenerator) iterate(input string) string {
	var result strings.Builder
	result.Grow(len(input) * 2) // Pre-allocate for efficiency

	for _, char := range input {
		symbol := Symbol(char)
		replacement := g.applyRules(symbol)
		result.WriteString(replacement)
	}

	return result.String()
}

// applyRules applies production rules to a symbol.
func (g *LSystemGenerator) applyRules(symbol Symbol) string {
	// Find all applicable rules
	var applicableRules []ProductionRule
	totalWeight := 0.0

	for _, rule := range g.config.Rules {
		if rule.From == symbol {
			applicableRules = append(applicableRules, rule)
			totalWeight += rule.Weight
		}
	}

	// No rules found - return the symbol unchanged
	if len(applicableRules) == 0 {
		return string(symbol)
	}

	// Single rule - apply it
	if len(applicableRules) == 1 {
		return applicableRules[0].To
	}

	// Multiple rules - choose stochastically based on weights
	choice := g.rng.Float64() * totalWeight
	cumulative := 0.0

	for _, rule := range applicableRules {
		cumulative += rule.Weight
		if choice <= cumulative {
			return rule.To
		}
	}

	// Fallback (shouldn't happen)
	return applicableRules[len(applicableRules)-1].To
}

// countRooms counts the number of room symbols in the string.
func (g *LSystemGenerator) countRooms(s string) int {
	count := 0
	for _, char := range s {
		symbol := Symbol(char)
		if g.isRoomSymbol(symbol) {
			count++
		}
	}
	return count
}

// isRoomSymbol returns true if the symbol represents a room.
func (g *LSystemGenerator) isRoomSymbol(symbol Symbol) bool {
	switch symbol {
	case SymbolStart, SymbolEnd, SymbolCombat, SymbolTreasure,
		SymbolPuzzle, SymbolShop, SymbolRest, SymbolSecret:
		return true
	default:
		return false
	}
}

// GetFantasyConfig returns an L-system configuration for fantasy dungeons.
func GetFantasyConfig(seed int64) LSystemConfig {
	return LSystemConfig{
		Axiom:        "S",
		Iterations:   3,
		Seed:         seed,
		MinRoomCount: 8,
		MaxRoomCount: 20,
		Rules: []ProductionRule{
			// Start expands to corridor with combat or puzzle
			{From: SymbolStart, To: "S-C", Weight: 0.5},
			{From: SymbolStart, To: "S-P", Weight: 0.3},
			{From: SymbolStart, To: "S-+", Weight: 0.2},
			// Combat leads to more combat or treasure
			{From: SymbolCombat, To: "C-C", Weight: 0.4},
			{From: SymbolCombat, To: "C-T", Weight: 0.3},
			{From: SymbolCombat, To: "C-E", Weight: 0.2},
			{From: SymbolCombat, To: "C", Weight: 0.1}, // Terminal
			// Puzzle leads to treasure or secret
			{From: SymbolPuzzle, To: "P-T", Weight: 0.5},
			{From: SymbolPuzzle, To: "P-?", Weight: 0.3},
			{From: SymbolPuzzle, To: "P", Weight: 0.2}, // Terminal
			// Branch creates multiple paths
			{From: SymbolBranch, To: "+C+P", Weight: 0.6},
			{From: SymbolBranch, To: "+C+R", Weight: 0.3},
			{From: SymbolBranch, To: "+", Weight: 0.1}, // Terminal
			// Treasure is usually terminal
			{From: SymbolTreasure, To: "T", Weight: 1.0},
			// Shop is usually terminal
			{From: SymbolShop, To: "$", Weight: 1.0},
			// Rest leads to more combat
			{From: SymbolRest, To: "R-C", Weight: 0.7},
			{From: SymbolRest, To: "R", Weight: 0.3}, // Terminal
			// Secret is always terminal
			{From: SymbolSecret, To: "?", Weight: 1.0},
			// End is always terminal
			{From: SymbolEnd, To: "E", Weight: 1.0},
		},
	}
}

// GetSciFiConfig returns an L-system configuration for sci-fi dungeons.
func GetSciFiConfig(seed int64) LSystemConfig {
	return LSystemConfig{
		Axiom:        "S",
		Iterations:   3,
		Seed:         seed,
		MinRoomCount: 10,
		MaxRoomCount: 25,
		Rules: []ProductionRule{
			// Start expands with branching corridors (space station feel)
			{From: SymbolStart, To: "S-+", Weight: 0.6},
			{From: SymbolStart, To: "S-C", Weight: 0.4},
			// Combat with higher branching (modular station design)
			{From: SymbolCombat, To: "C-+", Weight: 0.5},
			{From: SymbolCombat, To: "C-C", Weight: 0.3},
			{From: SymbolCombat, To: "C-E", Weight: 0.15},
			{From: SymbolCombat, To: "C", Weight: 0.05},
			// Puzzle leads to research/treasure
			{From: SymbolPuzzle, To: "P-T", Weight: 0.6},
			{From: SymbolPuzzle, To: "P-?", Weight: 0.3},
			{From: SymbolPuzzle, To: "P", Weight: 0.1},
			// Branch with more options (interconnected)
			{From: SymbolBranch, To: "+C+P+$", Weight: 0.4},
			{From: SymbolBranch, To: "+C+C", Weight: 0.4},
			{From: SymbolBranch, To: "+", Weight: 0.2},
			// Other terminals
			{From: SymbolTreasure, To: "T", Weight: 1.0},
			{From: SymbolShop, To: "$", Weight: 1.0},
			{From: SymbolRest, To: "R-C", Weight: 0.8},
			{From: SymbolRest, To: "R", Weight: 0.2},
			{From: SymbolSecret, To: "?", Weight: 1.0},
			{From: SymbolEnd, To: "E", Weight: 1.0},
		},
	}
}

// GetHorrorConfig returns an L-system configuration for horror dungeons.
func GetHorrorConfig(seed int64) LSystemConfig {
	return LSystemConfig{
		Axiom:        "S",
		Iterations:   4,
		Seed:         seed,
		MinRoomCount: 6,
		MaxRoomCount: 15,
		Rules: []ProductionRule{
			// Start with linear progression (oppressive atmosphere)
			{From: SymbolStart, To: "S-C", Weight: 0.7},
			{From: SymbolStart, To: "S-P", Weight: 0.3},
			// Combat leads mostly to more combat (relentless)
			{From: SymbolCombat, To: "C-C", Weight: 0.6},
			{From: SymbolCombat, To: "C-?", Weight: 0.2},
			{From: SymbolCombat, To: "C-E", Weight: 0.15},
			{From: SymbolCombat, To: "C", Weight: 0.05},
			// Puzzle with secrets (hidden lore)
			{From: SymbolPuzzle, To: "P-?", Weight: 0.6},
			{From: SymbolPuzzle, To: "P-C", Weight: 0.3},
			{From: SymbolPuzzle, To: "P", Weight: 0.1},
			// Branch is rare (linear feel)
			{From: SymbolBranch, To: "+C+?", Weight: 0.7},
			{From: SymbolBranch, To: "+", Weight: 0.3},
			// Minimal safe rooms
			{From: SymbolRest, To: "R-C", Weight: 0.9},
			{From: SymbolRest, To: "R", Weight: 0.1},
			// Other terminals
			{From: SymbolTreasure, To: "T", Weight: 1.0},
			{From: SymbolShop, To: "$", Weight: 1.0},
			{From: SymbolSecret, To: "?", Weight: 1.0},
			{From: SymbolEnd, To: "E", Weight: 1.0},
		},
	}
}

// GetCyberpunkConfig returns an L-system configuration for cyberpunk dungeons.
func GetCyberpunkConfig(seed int64) LSystemConfig {
	return LSystemConfig{
		Axiom:        "S",
		Iterations:   3,
		Seed:         seed,
		MinRoomCount: 12,
		MaxRoomCount: 25,
		Rules: []ProductionRule{
			// Start with branching (corporate tower, multiple paths)
			{From: SymbolStart, To: "S-+", Weight: 0.5},
			{From: SymbolStart, To: "S-C", Weight: 0.3},
			{From: SymbolStart, To: "S-$", Weight: 0.2}, // Shops common (black market)
			// Combat with shops nearby (arms dealers)
			{From: SymbolCombat, To: "C-$", Weight: 0.4},
			{From: SymbolCombat, To: "C-+", Weight: 0.3},
			{From: SymbolCombat, To: "C-E", Weight: 0.2},
			{From: SymbolCombat, To: "C", Weight: 0.1},
			// Puzzle (hacking terminals)
			{From: SymbolPuzzle, To: "P-T", Weight: 0.5},
			{From: SymbolPuzzle, To: "P-$", Weight: 0.3},
			{From: SymbolPuzzle, To: "P", Weight: 0.2},
			// Branch with many options
			{From: SymbolBranch, To: "+C+$+P", Weight: 0.5},
			{From: SymbolBranch, To: "+C+C", Weight: 0.3},
			{From: SymbolBranch, To: "+", Weight: 0.2},
			// Other terminals
			{From: SymbolTreasure, To: "T", Weight: 1.0},
			{From: SymbolShop, To: "$", Weight: 1.0},
			{From: SymbolRest, To: "R-C", Weight: 0.7},
			{From: SymbolRest, To: "R", Weight: 0.3},
			{From: SymbolSecret, To: "?", Weight: 1.0},
			{From: SymbolEnd, To: "E", Weight: 1.0},
		},
	}
}

// GetPostApocalypticConfig returns an L-system configuration for post-apocalyptic dungeons.
func GetPostApocalypticConfig(seed int64) LSystemConfig {
	return LSystemConfig{
		Axiom:        "S",
		Iterations:   3,
		Seed:         seed,
		MinRoomCount: 8,
		MaxRoomCount: 18,
		Rules: []ProductionRule{
			// Start with scavenging feel
			{From: SymbolStart, To: "S-C", Weight: 0.5},
			{From: SymbolStart, To: "S-T", Weight: 0.3}, // Loot common
			{From: SymbolStart, To: "S-R", Weight: 0.2}, // Safe rooms (bunkers)
			// Combat with loot
			{From: SymbolCombat, To: "C-T", Weight: 0.5},
			{From: SymbolCombat, To: "C-C", Weight: 0.3},
			{From: SymbolCombat, To: "C-E", Weight: 0.15},
			{From: SymbolCombat, To: "C", Weight: 0.05},
			// Puzzle leads to treasure (locked bunkers)
			{From: SymbolPuzzle, To: "P-T", Weight: 0.7},
			{From: SymbolPuzzle, To: "P-?", Weight: 0.2},
			{From: SymbolPuzzle, To: "P", Weight: 0.1},
			// Branch with resource focus
			{From: SymbolBranch, To: "+T+C", Weight: 0.6},
			{From: SymbolBranch, To: "+R+$", Weight: 0.3},
			{From: SymbolBranch, To: "+", Weight: 0.1},
			// Rest areas (survivor camps)
			{From: SymbolRest, To: "R-$", Weight: 0.6},
			{From: SymbolRest, To: "R-C", Weight: 0.3},
			{From: SymbolRest, To: "R", Weight: 0.1},
			// Other terminals
			{From: SymbolTreasure, To: "T", Weight: 1.0},
			{From: SymbolShop, To: "$", Weight: 1.0},
			{From: SymbolSecret, To: "?", Weight: 1.0},
			{From: SymbolEnd, To: "E", Weight: 1.0},
		},
	}
}

// GetConfigForGenre returns the appropriate L-system configuration for a genre.
func GetConfigForGenre(genre string, seed int64) LSystemConfig {
	switch genre {
	case "fantasy":
		return GetFantasyConfig(seed)
	case "sci-fi", "scifi":
		return GetSciFiConfig(seed)
	case "horror":
		return GetHorrorConfig(seed)
	case "cyberpunk":
		return GetCyberpunkConfig(seed)
	case "post-apocalyptic", "postapocalyptic":
		return GetPostApocalypticConfig(seed)
	default:
		// Default to fantasy
		return GetFantasyConfig(seed)
	}
}
