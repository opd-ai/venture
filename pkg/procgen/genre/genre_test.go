package genre

import (
	"testing"
)

func TestGenre_Validate(t *testing.T) {
	tests := []struct {
		name    string
		genre   *Genre
		wantErr bool
	}{
		{
			name: "valid genre",
			genre: &Genre{
				ID:          "test",
				Name:        "Test Genre",
				Description: "A test genre",
				Themes:      []string{"testing"},
			},
			wantErr: false,
		},
		{
			name: "missing ID",
			genre: &Genre{
				Name:        "Test Genre",
				Description: "A test genre",
				Themes:      []string{"testing"},
			},
			wantErr: true,
		},
		{
			name: "missing name",
			genre: &Genre{
				ID:          "test",
				Description: "A test genre",
				Themes:      []string{"testing"},
			},
			wantErr: true,
		},
		{
			name: "missing description",
			genre: &Genre{
				ID:     "test",
				Name:   "Test Genre",
				Themes: []string{"testing"},
			},
			wantErr: true,
		},
		{
			name: "missing themes",
			genre: &Genre{
				ID:          "test",
				Name:        "Test Genre",
				Description: "A test genre",
				Themes:      []string{},
			},
			wantErr: true,
		},
		{
			name: "nil themes",
			genre: &Genre{
				ID:          "test",
				Name:        "Test Genre",
				Description: "A test genre",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.genre.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Genre.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenre_ColorPalette(t *testing.T) {
	genre := &Genre{
		PrimaryColor:   "#FF0000",
		SecondaryColor: "#00FF00",
		AccentColor:    "#0000FF",
	}
	
	palette := genre.ColorPalette()
	if len(palette) != 3 {
		t.Errorf("Expected 3 colors, got %d", len(palette))
	}
	if palette[0] != "#FF0000" {
		t.Errorf("Expected primary color #FF0000, got %s", palette[0])
	}
	if palette[1] != "#00FF00" {
		t.Errorf("Expected secondary color #00FF00, got %s", palette[1])
	}
	if palette[2] != "#0000FF" {
		t.Errorf("Expected accent color #0000FF, got %s", palette[2])
	}
}

func TestGenre_HasTheme(t *testing.T) {
	genre := &Genre{
		Themes: []string{"magic", "dragons", "knights"},
	}
	
	if !genre.HasTheme("magic") {
		t.Error("Expected genre to have 'magic' theme")
	}
	if !genre.HasTheme("dragons") {
		t.Error("Expected genre to have 'dragons' theme")
	}
	if genre.HasTheme("robots") {
		t.Error("Expected genre not to have 'robots' theme")
	}
}

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()
	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}
	if registry.Count() != 0 {
		t.Errorf("Expected empty registry, got %d genres", registry.Count())
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	
	genre := &Genre{
		ID:          "test",
		Name:        "Test",
		Description: "Test genre",
		Themes:      []string{"testing"},
	}
	
	// Register valid genre
	err := registry.Register(genre)
	if err != nil {
		t.Errorf("Failed to register valid genre: %v", err)
	}
	
	// Try to register duplicate
	err = registry.Register(genre)
	if err == nil {
		t.Error("Expected error when registering duplicate genre")
	}
	
	// Try to register invalid genre
	invalidGenre := &Genre{
		ID: "invalid",
		// Missing required fields
	}
	err = registry.Register(invalidGenre)
	if err == nil {
		t.Error("Expected error when registering invalid genre")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	
	genre := &Genre{
		ID:          "test",
		Name:        "Test",
		Description: "Test genre",
		Themes:      []string{"testing"},
	}
	
	registry.Register(genre)
	
	// Get existing genre
	retrieved, err := registry.Get("test")
	if err != nil {
		t.Errorf("Failed to get genre: %v", err)
	}
	if retrieved.ID != genre.ID {
		t.Errorf("Expected genre ID %s, got %s", genre.ID, retrieved.ID)
	}
	
	// Get non-existent genre
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting non-existent genre")
	}
}

func TestRegistry_Has(t *testing.T) {
	registry := NewRegistry()
	
	genre := &Genre{
		ID:          "test",
		Name:        "Test",
		Description: "Test genre",
		Themes:      []string{"testing"},
	}
	
	registry.Register(genre)
	
	if !registry.Has("test") {
		t.Error("Expected registry to have 'test' genre")
	}
	if registry.Has("nonexistent") {
		t.Error("Expected registry not to have 'nonexistent' genre")
	}
}

func TestRegistry_All(t *testing.T) {
	registry := NewRegistry()
	
	genre1 := &Genre{
		ID:          "test1",
		Name:        "Test 1",
		Description: "Test genre 1",
		Themes:      []string{"testing"},
	}
	genre2 := &Genre{
		ID:          "test2",
		Name:        "Test 2",
		Description: "Test genre 2",
		Themes:      []string{"testing"},
	}
	
	registry.Register(genre1)
	registry.Register(genre2)
	
	all := registry.All()
	if len(all) != 2 {
		t.Errorf("Expected 2 genres, got %d", len(all))
	}
}

func TestRegistry_IDs(t *testing.T) {
	registry := NewRegistry()
	
	genre1 := &Genre{
		ID:          "test1",
		Name:        "Test 1",
		Description: "Test genre 1",
		Themes:      []string{"testing"},
	}
	genre2 := &Genre{
		ID:          "test2",
		Name:        "Test 2",
		Description: "Test genre 2",
		Themes:      []string{"testing"},
	}
	
	registry.Register(genre1)
	registry.Register(genre2)
	
	ids := registry.IDs()
	if len(ids) != 2 {
		t.Errorf("Expected 2 IDs, got %d", len(ids))
	}
	
	// Check that both IDs are present
	idMap := make(map[string]bool)
	for _, id := range ids {
		idMap[id] = true
	}
	if !idMap["test1"] || !idMap["test2"] {
		t.Error("Expected both test1 and test2 in IDs")
	}
}

func TestRegistry_Count(t *testing.T) {
	registry := NewRegistry()
	
	if registry.Count() != 0 {
		t.Errorf("Expected 0 genres, got %d", registry.Count())
	}
	
	genre := &Genre{
		ID:          "test",
		Name:        "Test",
		Description: "Test genre",
		Themes:      []string{"testing"},
	}
	registry.Register(genre)
	
	if registry.Count() != 1 {
		t.Errorf("Expected 1 genre, got %d", registry.Count())
	}
}

func TestDefaultRegistry(t *testing.T) {
	registry := DefaultRegistry()
	
	if registry == nil {
		t.Fatal("DefaultRegistry() returned nil")
	}
	
	// Check that predefined genres are registered
	if !registry.Has("fantasy") {
		t.Error("Expected fantasy genre in default registry")
	}
	if !registry.Has("scifi") {
		t.Error("Expected scifi genre in default registry")
	}
	if !registry.Has("horror") {
		t.Error("Expected horror genre in default registry")
	}
	if !registry.Has("cyberpunk") {
		t.Error("Expected cyberpunk genre in default registry")
	}
	if !registry.Has("postapoc") {
		t.Error("Expected postapoc genre in default registry")
	}
}

func TestPredefinedGenres(t *testing.T) {
	genres := PredefinedGenres()
	
	if len(genres) != 5 {
		t.Errorf("Expected 5 predefined genres, got %d", len(genres))
	}
	
	// Check that all genres are valid
	for _, genre := range genres {
		if err := genre.Validate(); err != nil {
			t.Errorf("Predefined genre %s is invalid: %v", genre.ID, err)
		}
	}
}

func TestFantasyGenre(t *testing.T) {
	genre := FantasyGenre()
	
	if genre.ID != "fantasy" {
		t.Errorf("Expected ID 'fantasy', got '%s'", genre.ID)
	}
	if genre.Name != "Fantasy" {
		t.Errorf("Expected name 'Fantasy', got '%s'", genre.Name)
	}
	if err := genre.Validate(); err != nil {
		t.Errorf("Fantasy genre validation failed: %v", err)
	}
	if !genre.HasTheme("magic") {
		t.Error("Expected fantasy genre to have 'magic' theme")
	}
}

func TestSciFiGenre(t *testing.T) {
	genre := SciFiGenre()
	
	if genre.ID != "scifi" {
		t.Errorf("Expected ID 'scifi', got '%s'", genre.ID)
	}
	if genre.Name != "Sci-Fi" {
		t.Errorf("Expected name 'Sci-Fi', got '%s'", genre.Name)
	}
	if err := genre.Validate(); err != nil {
		t.Errorf("Sci-Fi genre validation failed: %v", err)
	}
	if !genre.HasTheme("technology") {
		t.Error("Expected scifi genre to have 'technology' theme")
	}
}

func TestHorrorGenre(t *testing.T) {
	genre := HorrorGenre()
	
	if genre.ID != "horror" {
		t.Errorf("Expected ID 'horror', got '%s'", genre.ID)
	}
	if genre.Name != "Horror" {
		t.Errorf("Expected name 'Horror', got '%s'", genre.Name)
	}
	if err := genre.Validate(); err != nil {
		t.Errorf("Horror genre validation failed: %v", err)
	}
	if !genre.HasTheme("dark") {
		t.Error("Expected horror genre to have 'dark' theme")
	}
}

func TestCyberpunkGenre(t *testing.T) {
	genre := CyberpunkGenre()
	
	if genre.ID != "cyberpunk" {
		t.Errorf("Expected ID 'cyberpunk', got '%s'", genre.ID)
	}
	if genre.Name != "Cyberpunk" {
		t.Errorf("Expected name 'Cyberpunk', got '%s'", genre.Name)
	}
	if err := genre.Validate(); err != nil {
		t.Errorf("Cyberpunk genre validation failed: %v", err)
	}
	if !genre.HasTheme("cybernetic") {
		t.Error("Expected cyberpunk genre to have 'cybernetic' theme")
	}
}

func TestPostApocalypticGenre(t *testing.T) {
	genre := PostApocalypticGenre()
	
	if genre.ID != "postapoc" {
		t.Errorf("Expected ID 'postapoc', got '%s'", genre.ID)
	}
	if genre.Name != "Post-Apocalyptic" {
		t.Errorf("Expected name 'Post-Apocalyptic', got '%s'", genre.Name)
	}
	if err := genre.Validate(); err != nil {
		t.Errorf("Post-Apocalyptic genre validation failed: %v", err)
	}
	if !genre.HasTheme("wasteland") {
		t.Error("Expected postapoc genre to have 'wasteland' theme")
	}
}

func TestGenre_ColorPaletteLength(t *testing.T) {
	genres := PredefinedGenres()
	for _, genre := range genres {
		palette := genre.ColorPalette()
		if len(palette) != 3 {
			t.Errorf("Genre %s: expected 3 colors in palette, got %d", genre.ID, len(palette))
		}
	}
}

func TestRegistry_GetOrDefault(t *testing.T) {
	registry := DefaultRegistry()
	
	// Test getting existing genre
	fantasy, err := registry.Get("fantasy")
	if err != nil {
		t.Errorf("Failed to get fantasy genre: %v", err)
	}
	if fantasy.ID != "fantasy" {
		t.Errorf("Expected fantasy genre, got %s", fantasy.ID)
	}
	
	// Test fallback behavior for non-existent genre
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent genre")
	}
}
