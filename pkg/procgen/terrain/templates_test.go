package terrain

import (
	"testing"
)

func TestNewTemplateLibrary(t *testing.T) {
	seed := int64(12345)
	lib := NewTemplateLibrary(seed)

	if lib == nil {
		t.Fatal("NewTemplateLibrary() returned nil")
	}

	if lib.templates == nil {
		t.Error("Templates map not initialized")
	}

	if lib.rng == nil {
		t.Error("RNG not initialized")
	}

	// Verify genres are initialized
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}
	for _, genre := range genres {
		if _, ok := lib.templates[genre]; !ok {
			t.Errorf("Genre %s not initialized", genre)
		}
	}
}

func TestTemplateLibrary_GetTemplate_Fantasy(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	tests := []struct {
		name     string
		roomType RoomType
	}{
		{"start", RoomSpawn},
		{"combat", RoomCombat},
		{"treasure", RoomTreasure},
		{"puzzle", RoomPuzzle},
		{"boss", RoomBoss},
		{"shop", RoomShop},
		{"rest", RoomRest},
		{"secret", RoomSecret},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := lib.GetTemplate("fantasy", tt.roomType)

			if template == nil {
				t.Fatal("GetTemplate() returned nil")
			}

			if template.Genre != "fantasy" {
				t.Errorf("Genre = %s, want fantasy", template.Genre)
			}

			if template.RoomType != tt.roomType {
				t.Errorf("RoomType = %v, want %v", template.RoomType, tt.roomType)
			}

			if template.Name == "" {
				t.Error("Template has no name")
			}

			if template.Description == "" {
				t.Error("Template has no description")
			}

			if template.TileTheme == "" {
				t.Error("Template has no tile theme")
			}

			if template.Lighting == "" {
				t.Error("Template has no lighting")
			}
		})
	}
}

func TestTemplateLibrary_GetTemplate_SciFi(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	template := lib.GetTemplate("sci-fi", RoomSpawn)

	if template == nil {
		t.Fatal("GetTemplate() returned nil")
	}

	if template.Genre != "sci-fi" {
		t.Errorf("Genre = %s, want sci-fi", template.Genre)
	}

	if template.Name == "" {
		t.Error("Sci-fi template has no name")
	}
}

func TestTemplateLibrary_GetTemplate_Horror(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	template := lib.GetTemplate("horror", RoomCombat)

	if template == nil {
		t.Fatal("GetTemplate() returned nil")
	}

	if template.Genre != "horror" {
		t.Errorf("Genre = %s, want horror", template.Genre)
	}

	if template.Name == "" {
		t.Error("Horror template has no name")
	}
}

func TestTemplateLibrary_GetTemplate_Cyberpunk(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	template := lib.GetTemplate("cyberpunk", RoomShop)

	if template == nil {
		t.Fatal("GetTemplate() returned nil")
	}

	if template.Genre != "cyberpunk" {
		t.Errorf("Genre = %s, want cyberpunk", template.Genre)
	}

	if template.Name == "" {
		t.Error("Cyberpunk template has no name")
	}
}

func TestTemplateLibrary_GetTemplate_PostApocalyptic(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	template := lib.GetTemplate("post-apocalyptic", RoomTreasure)

	if template == nil {
		t.Fatal("GetTemplate() returned nil")
	}

	if template.Genre != "post-apocalyptic" {
		t.Errorf("Genre = %s, want post-apocalyptic", template.Genre)
	}

	if template.Name == "" {
		t.Error("Post-apocalyptic template has no name")
	}
}

func TestTemplateLibrary_GetTemplate_UnknownGenre(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	// Should fallback to fantasy
	template := lib.GetTemplate("unknown_genre", RoomSpawn)

	if template == nil {
		t.Fatal("GetTemplate() returned nil for unknown genre")
	}

	// Should fallback to fantasy
	if template.Genre != "fantasy" && template.Genre != "unknown_genre" {
		t.Error("Unknown genre didn't fallback properly")
	}
}

func TestTemplateLibrary_GetTemplate_UnknownRoomType(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	// Should return default template
	template := lib.GetTemplate("fantasy", RoomType(999))

	if template == nil {
		t.Fatal("GetTemplate() returned nil for unknown room type")
	}

	if template.Name == "" {
		t.Error("Default template has no name")
	}
}

func TestTemplateLibrary_Randomness(t *testing.T) {
	// Test that different seeds produce different template selections
	lib1 := NewTemplateLibrary(12345)
	lib2 := NewTemplateLibrary(54321)

	// Get multiple templates from each
	templates1 := make([]string, 10)
	templates2 := make([]string, 10)

	for i := 0; i < 10; i++ {
		templates1[i] = lib1.GetTemplate("fantasy", RoomCombat).Name
		templates2[i] = lib2.GetTemplate("fantasy", RoomCombat).Name
	}

	// At least some should be different (very unlikely all are same with different seeds)
	allSame := true
	for i := range templates1 {
		if templates1[i] != templates2[i] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Different seeds produced identical template selections")
	}
}

func TestTemplateLibrary_Determinism(t *testing.T) {
	// Test that same seed produces same template selections
	seed := int64(12345)
	lib1 := NewTemplateLibrary(seed)
	lib2 := NewTemplateLibrary(seed)

	for i := 0; i < 10; i++ {
		template1 := lib1.GetTemplate("fantasy", RoomCombat)
		template2 := lib2.GetTemplate("fantasy", RoomCombat)

		if template1.Name != template2.Name {
			t.Errorf("Same seed produced different templates: %s vs %s", template1.Name, template2.Name)
		}
	}
}

func TestApplyTemplateToRoom(t *testing.T) {
	lib := NewTemplateLibrary(12345)
	template := lib.GetTemplate("fantasy", RoomSpawn)

	room := &RoomNode{
		ID:   0,
		Type: RoomSpawn,
	}

	ApplyTemplateToRoom(room, template)

	if room.Theme != template.Name {
		t.Errorf("Theme = %s, want %s", room.Theme, template.Name)
	}

	if room.Properties == nil {
		t.Fatal("Properties not initialized")
	}

	if room.Properties["description"] != template.Description {
		t.Error("Description not applied to properties")
	}

	if room.Properties["tileTheme"] != template.TileTheme {
		t.Error("TileTheme not applied to properties")
	}

	if room.Properties["lighting"] != template.Lighting {
		t.Error("Lighting not applied to properties")
	}

	if room.Properties["ambience"] != template.Ambience {
		t.Error("Ambience not applied to properties")
	}
}

func TestApplyTemplateToRoom_NilProperties(t *testing.T) {
	lib := NewTemplateLibrary(12345)
	template := lib.GetTemplate("fantasy", RoomSpawn)

	room := &RoomNode{
		ID:         0,
		Type:       RoomSpawn,
		Properties: nil, // Start with nil
	}

	ApplyTemplateToRoom(room, template)

	if room.Properties == nil {
		t.Fatal("Properties not initialized by ApplyTemplateToRoom")
	}

	if room.Properties["description"] != template.Description {
		t.Error("Description not applied when properties was nil")
	}
}

func TestFantasyTemplates_Coverage(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	// Ensure fantasy genre has templates for all major room types
	requiredTypes := []RoomType{
		RoomSpawn,
		RoomCombat,
		RoomTreasure,
		RoomPuzzle,
		RoomBoss,
		RoomShop,
		RoomRest,
		RoomSecret,
	}

	for _, roomType := range requiredTypes {
		template := lib.GetTemplate("fantasy", roomType)
		if template.Name == "Generic Room" {
			t.Errorf("Fantasy genre missing template for %s", roomType.String())
		}
	}
}

func TestFantasyTemplates_MultipleOptions(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	// Test that some room types have multiple template options
	// by getting multiple templates and checking for variety
	startTemplates := make(map[string]bool)

	for i := 0; i < 20; i++ {
		template := lib.GetTemplate("fantasy", RoomSpawn)
		startTemplates[template.Name] = true
	}

	if len(startTemplates) < 2 {
		t.Error("Fantasy start rooms should have multiple template options")
	}
}

func TestSciFiTemplates_Unique(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	template := lib.GetTemplate("sci-fi", RoomSpawn)

	// Verify sci-fi specific elements
	if template.TileTheme != "metal_grating" {
		t.Error("Sci-fi template doesn't have appropriate tile theme")
	}

	// Check for sci-fi specific decorations
	hasSciFiDecor := false
	for _, decor := range template.Decorations {
		if decor == "loading_cranes" || decor == "cargo_containers" || decor == "airlocks" {
			hasSciFiDecor = true
			break
		}
	}

	if !hasSciFiDecor {
		t.Error("Sci-fi template missing genre-specific decorations")
	}
}

func TestHorrorTemplates_Atmosphere(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	template := lib.GetTemplate("horror", RoomCombat)

	// Verify horror atmosphere
	if template.Lighting != "dim" && template.Lighting != "flickering" && template.Lighting != "dark" {
		t.Errorf("Horror template has inappropriate lighting: %s", template.Lighting)
	}

	// Check for horror-specific elements in description
	desc := template.Description
	if len(desc) == 0 {
		t.Error("Horror template has no description")
	}
}

func TestCyberpunkTemplates_Theme(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	template := lib.GetTemplate("cyberpunk", RoomShop)

	// Verify cyberpunk theme elements
	if template.Name != "Black Market Stall" {
		t.Error("Cyberpunk shop doesn't have expected name")
	}

	// Check for cyberpunk-specific decorations
	hasCyberDecor := false
	for _, decor := range template.Decorations {
		if decor == "tech_displays" || decor == "cyber_implants" || decor == "holo_prices" {
			hasCyberDecor = true
			break
		}
	}

	if !hasCyberDecor {
		t.Error("Cyberpunk template missing genre-specific decorations")
	}
}

func TestPostApocalypticTemplates_Scarcity(t *testing.T) {
	lib := NewTemplateLibrary(12345)

	template := lib.GetTemplate("post-apocalyptic", RoomTreasure)

	// Verify post-apocalyptic scarcity theme
	if template.Name != "Supply Cache" {
		t.Error("Post-apocalyptic treasure doesn't reflect scarcity theme")
	}

	// Check for survival-related decorations
	hasSurvivalDecor := false
	for _, decor := range template.Decorations {
		if decor == "supply_containers" || decor == "water_barrels" || decor == "medical_kits" {
			hasSurvivalDecor = true
			break
		}
	}

	if !hasSurvivalDecor {
		t.Error("Post-apocalyptic template missing survival decorations")
	}
}

func BenchmarkTemplateLibrary_GetTemplate(b *testing.B) {
	lib := NewTemplateLibrary(12345)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lib.GetTemplate("fantasy", RoomCombat)
	}
}

func BenchmarkApplyTemplateToRoom(b *testing.B) {
	lib := NewTemplateLibrary(12345)
	template := lib.GetTemplate("fantasy", RoomSpawn)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		room := &RoomNode{ID: i, Type: RoomSpawn}
		ApplyTemplateToRoom(room, template)
	}
}
