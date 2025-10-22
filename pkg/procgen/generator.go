package procgen

// GenerationParams contains parameters that control content generation.
type GenerationParams struct {
	// Difficulty affects the challenge level of generated content (0.0-1.0)
	Difficulty float64

	// Depth represents how far into the game this content appears
	Depth int

	// Genre influences the style and theme of generated content
	GenreID string

	// Additional custom parameters
	Custom map[string]interface{}
}

// Generator is the base interface for all procedural generation systems.
// Generators produce deterministic output based on a seed value.
type Generator interface {
	// Generate creates content based on the seed and parameters
	Generate(seed int64, params GenerationParams) (interface{}, error)

	// Validate checks if the generated content is valid
	Validate(result interface{}) error
}

// SeedGenerator manages seed generation for different aspects of the game.
type SeedGenerator struct {
	baseSeed int64
}

// NewSeedGenerator creates a new seed generator with a base seed.
func NewSeedGenerator(baseSeed int64) *SeedGenerator {
	return &SeedGenerator{baseSeed: baseSeed}
}

// GetSeed generates a deterministic seed for a specific purpose.
// This allows different aspects of the game to have independent but deterministic seeds.
func (sg *SeedGenerator) GetSeed(category string, index int) int64 {
	// Simple hash combination - can be improved with better hash function
	hash := sg.baseSeed
	for _, c := range category {
		hash = hash*31 + int64(c)
	}
	hash = hash*31 + int64(index)
	return hash
}
