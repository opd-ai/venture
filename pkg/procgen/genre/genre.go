// Package genre provides genre type definitions.
// This file defines the core Genre struct and its methods.
// Code relocated from: types.go
package genre

import "fmt"

// Genre represents a game genre with associated metadata and theming.
type Genre struct {
	// ID is the unique identifier for this genre (lowercase, no spaces)
	ID string

	// Name is the human-readable name of the genre
	Name string

	// Description provides a brief description of the genre
	Description string

	// Themes are keywords that describe the genre's aesthetic and content
	Themes []string

	// PrimaryColor is the main color associated with this genre (RGB hex)
	PrimaryColor string

	// SecondaryColor is an accent color for this genre (RGB hex)
	SecondaryColor string

	// AccentColor is another accent color for variety (RGB hex)
	AccentColor string

	// EntityPrefix is the prefix used for entity names in this genre
	EntityPrefix string

	// ItemPrefix is the prefix used for item names in this genre
	ItemPrefix string

	// LocationPrefix is the prefix used for location names in this genre
	LocationPrefix string
}

// ColorPalette returns the genre's color palette as a slice of hex colors.
func (g *Genre) ColorPalette() []string {
	return []string{g.PrimaryColor, g.SecondaryColor, g.AccentColor}
}

// HasTheme checks if the genre contains a specific theme keyword.
func (g *Genre) HasTheme(theme string) bool {
	for _, t := range g.Themes {
		if t == theme {
			return true
		}
	}
	return false
}

// Validate checks if the genre definition is valid.
func (g *Genre) Validate() error {
	if g.ID == "" {
		return fmt.Errorf("genre ID cannot be empty")
	}
	if g.Name == "" {
		return fmt.Errorf("genre name cannot be empty")
	}
	if g.Description == "" {
		return fmt.Errorf("genre description cannot be empty")
	}
	if len(g.Themes) == 0 {
		return fmt.Errorf("genre must have at least one theme")
	}
	// Color validation is optional - some genres might not define colors
	return nil
}
