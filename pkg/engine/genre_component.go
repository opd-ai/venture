// Package engine provides genre tracking for entities.
// This file implements GenreComponent which stores the genre/theme for an entity,
// affecting visual appearance, item generation, and thematic elements.
package engine

// GenreComponent tracks the genre/theme of an entity.
// This affects procedural generation for visuals, items, and flavor text.
// Genres include: fantasy, scifi, horror, cyberpunk, post_apocalyptic.
type GenreComponent struct {
	// GenreID is the primary genre identifier
	GenreID string

	// SecondaryGenres allows genre blending (e.g., dark fantasy = fantasy + horror)
	SecondaryGenres []string

	// BlendRatio controls how much secondary genres influence generation (0.0-1.0)
	BlendRatio float64
}

// Type implements Component interface.
func (g *GenreComponent) Type() string {
	return "genre"
}

// NewGenreComponent creates a new genre component with the specified genre.
func NewGenreComponent(genreID string) *GenreComponent {
	return &GenreComponent{
		GenreID:         genreID,
		SecondaryGenres: nil,
		BlendRatio:      0.0,
	}
}

// NewBlendedGenreComponent creates a genre component with genre blending.
// blendRatio controls how much secondary genres influence (0.0 = none, 1.0 = equal).
func NewBlendedGenreComponent(primaryGenre string, secondaryGenres []string, blendRatio float64) *GenreComponent {
	if blendRatio < 0 {
		blendRatio = 0
	}
	if blendRatio > 1 {
		blendRatio = 1
	}

	return &GenreComponent{
		GenreID:         primaryGenre,
		SecondaryGenres: secondaryGenres,
		BlendRatio:      blendRatio,
	}
}

// GetPrimaryGenre returns the primary genre ID.
func (g *GenreComponent) GetPrimaryGenre() string {
	return g.GenreID
}

// GetAllGenres returns all genres (primary + secondary).
func (g *GenreComponent) GetAllGenres() []string {
	genres := []string{g.GenreID}
	genres = append(genres, g.SecondaryGenres...)
	return genres
}

// HasSecondaryGenres returns true if genre blending is active.
func (g *GenreComponent) HasSecondaryGenres() bool {
	return len(g.SecondaryGenres) > 0 && g.BlendRatio > 0
}

// IsBlended returns true if this uses genre blending.
func (g *GenreComponent) IsBlended() bool {
	return g.HasSecondaryGenres()
}

// SetGenre changes the primary genre.
func (g *GenreComponent) SetGenre(genreID string) {
	g.GenreID = genreID
}

// AddSecondaryGenre adds a secondary genre for blending.
func (g *GenreComponent) AddSecondaryGenre(genreID string) {
	// Check if already present
	for _, existingGenre := range g.SecondaryGenres {
		if existingGenre == genreID {
			return
		}
	}
	g.SecondaryGenres = append(g.SecondaryGenres, genreID)
}

// RemoveSecondaryGenre removes a secondary genre.
func (g *GenreComponent) RemoveSecondaryGenre(genreID string) {
	newSecondary := make([]string, 0, len(g.SecondaryGenres))
	for _, existingGenre := range g.SecondaryGenres {
		if existingGenre != genreID {
			newSecondary = append(newSecondary, existingGenre)
		}
	}
	g.SecondaryGenres = newSecondary
}

// SetBlendRatio sets the genre blend ratio (0.0-1.0).
func (g *GenreComponent) SetBlendRatio(ratio float64) {
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	g.BlendRatio = ratio
}

// GetBlendRatio returns the current genre blend ratio.
func (g *GenreComponent) GetBlendRatio() float64 {
	return g.BlendRatio
}
