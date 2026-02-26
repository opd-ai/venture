// Package genre provides genre registry functionality.
// This file defines the Registry struct for managing genre collections.
// Code relocated from: types.go
package genre

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

// Registry manages a collection of genres.
type Registry struct {
	genres map[string]*Genre
}

// NewRegistry creates a new empty genre registry.
func NewRegistry() *Registry {
	return &Registry{
		genres: make(map[string]*Genre),
	}
}

// Register adds a genre to the registry.
// Returns an error if the genre is invalid or if a genre with the same ID already exists.
func (r *Registry) Register(g *Genre) error {
	if err := g.Validate(); err != nil {
		return fmt.Errorf("invalid genre: %w", err)
	}
	if _, exists := r.genres[g.ID]; exists {
		return fmt.Errorf("genre with ID '%s' already registered", g.ID)
	}
	r.genres[g.ID] = g
	return nil
}

// Get retrieves a genre by its ID.
// Returns an error if the genre is not found.
func (r *Registry) Get(id string) (*Genre, error) {
	g, exists := r.genres[id]
	if !exists {
		return nil, fmt.Errorf("genre '%s' not found", id)
	}
	return g, nil
}

// Has checks if a genre with the given ID exists in the registry.
func (r *Registry) Has(id string) bool {
	_, exists := r.genres[id]
	return exists
}

// All returns a slice of all registered genres.
func (r *Registry) All() []*Genre {
	genres := make([]*Genre, 0, len(r.genres))
	for _, g := range r.genres {
		genres = append(genres, g)
	}
	return genres
}

// IDs returns a slice of all registered genre IDs.
func (r *Registry) IDs() []string {
	ids := make([]string, 0, len(r.genres))
	for id := range r.genres {
		ids = append(ids, id)
	}
	return ids
}

// Count returns the number of registered genres.
func (r *Registry) Count() int {
	return len(r.genres)
}

// DefaultRegistry returns a registry pre-populated with standard genres
// (fantasy, sci-fi, horror, cyberpunk, post-apocalyptic).
//
// Panics if any predefined genre fails validation, which indicates a programmer
// error in the predefined genre definitions. This panic-on-error behavior is
// intentional as the default registry must always be valid for the game to function.
// All predefined genres are guaranteed to pass validation in normal operation.
func DefaultRegistry() *Registry {
	registry := NewRegistry()

	// Register all predefined genres
	for _, g := range PredefinedGenres() {
		if err := registry.Register(g); err != nil {
			log.WithFields(log.Fields{
				"genre_id": g.ID,
				"error":    err.Error(),
			}).Fatal("Failed to register predefined genre")
		}
	}

	return registry
}
