//go:build headless
// +build headless

// Package sprites provides headless stubs for sprite generation.
// This file is used for server builds that don't require actual sprite rendering.
package sprites

// Generator is a stub for headless builds.
type Generator struct{}

// NewGenerator creates a new headless sprite generator stub.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateDirectionalSprites is a stub that returns empty directional sprite data.
// In headless mode, sprites aren't actually generated since there's no rendering.
// Returns nil to indicate no sprite images are available.
func (g *Generator) GenerateDirectionalSprites(config Config) (map[int]interface{}, error) {
	// Return empty map - headless server doesn't need actual sprite images
	return make(map[int]interface{}), nil
}
