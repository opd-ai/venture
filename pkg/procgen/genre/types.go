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

// DefaultRegistry returns a registry pre-populated with standard genres.
func DefaultRegistry() *Registry {
	registry := NewRegistry()
	
	// Register all predefined genres
	for _, g := range PredefinedGenres() {
		// Ignore errors for predefined genres - they should always be valid
		_ = registry.Register(g)
	}
	
	return registry
}

// PredefinedGenres returns a slice of all predefined genre definitions.
func PredefinedGenres() []*Genre {
	return []*Genre{
		FantasyGenre(),
		SciFiGenre(),
		HorrorGenre(),
		CyberpunkGenre(),
		PostApocalypticGenre(),
	}
}

// FantasyGenre returns the Fantasy genre definition.
func FantasyGenre() *Genre {
	return &Genre{
		ID:             "fantasy",
		Name:           "Fantasy",
		Description:    "Traditional medieval fantasy with magic, dragons, and ancient mysteries",
		Themes:         []string{"medieval", "magic", "dragons", "knights", "wizards", "dungeons"},
		PrimaryColor:   "#8B4513", // Saddle Brown
		SecondaryColor: "#DAA520", // Goldenrod
		AccentColor:    "#4169E1", // Royal Blue
		EntityPrefix:   "Ancient",
		ItemPrefix:     "Enchanted",
		LocationPrefix: "The",
	}
}

// SciFiGenre returns the Science Fiction genre definition.
func SciFiGenre() *Genre {
	return &Genre{
		ID:             "scifi",
		Name:           "Sci-Fi",
		Description:    "Science fiction with advanced technology, space exploration, and alien encounters",
		Themes:         []string{"technology", "space", "aliens", "robots", "lasers", "future"},
		PrimaryColor:   "#00CED1", // Dark Turquoise
		SecondaryColor: "#7B68EE", // Medium Slate Blue
		AccentColor:    "#00FF00", // Lime
		EntityPrefix:   "Prototype",
		ItemPrefix:     "Advanced",
		LocationPrefix: "Station",
	}
}

// HorrorGenre returns the Horror genre definition.
func HorrorGenre() *Genre {
	return &Genre{
		ID:             "horror",
		Name:           "Horror",
		Description:    "Dark, atmospheric horror with supernatural threats and psychological terror",
		Themes:         []string{"dark", "supernatural", "undead", "cursed", "twisted", "nightmare"},
		PrimaryColor:   "#8B0000", // Dark Red
		SecondaryColor: "#2F4F4F", // Dark Slate Gray
		AccentColor:    "#9370DB", // Medium Purple
		EntityPrefix:   "Cursed",
		ItemPrefix:     "Twisted",
		LocationPrefix: "The Haunted",
	}
}

// CyberpunkGenre returns the Cyberpunk genre definition.
func CyberpunkGenre() *Genre {
	return &Genre{
		ID:             "cyberpunk",
		Name:           "Cyberpunk",
		Description:    "High-tech dystopian future with cybernetic enhancements and corporate dominance",
		Themes:         []string{"cybernetic", "neon", "corporate", "hacker", "augmented", "dystopian"},
		PrimaryColor:   "#FF1493", // Deep Pink
		SecondaryColor: "#00FFFF", // Cyan
		AccentColor:    "#FFD700", // Gold
		EntityPrefix:   "Augmented",
		ItemPrefix:     "Cyber",
		LocationPrefix: "Neo",
	}
}

// PostApocalypticGenre returns the Post-Apocalyptic genre definition.
func PostApocalypticGenre() *Genre {
	return &Genre{
		ID:             "postapoc",
		Name:           "Post-Apocalyptic",
		Description:    "Wasteland survival in a world devastated by catastrophe",
		Themes:         []string{"wasteland", "survival", "scavenged", "mutated", "ruined", "barren"},
		PrimaryColor:   "#CD853F", // Peru
		SecondaryColor: "#696969", // Dim Gray
		AccentColor:    "#FF6347", // Tomato
		EntityPrefix:   "Mutated",
		ItemPrefix:     "Salvaged",
		LocationPrefix: "Ruins of",
	}
}
