package engine

import (
	"testing"
)

func TestNewBookReadingSystem(t *testing.T) {
	world := NewWorld()
	system := NewBookReadingSystem(world)

	if system == nil {
		t.Fatal("NewBookReadingSystem returned nil")
	}
	if system.world != world {
		t.Error("system world not set correctly")
	}
	if system.seriesBonuses == nil {
		t.Error("seriesBonuses map not initialized")
	}
}

func TestReadBook(t *testing.T) {
	world := NewWorld()
	system := NewBookReadingSystem(world)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&ExperienceComponent{
		Level:      1,
		CurrentXP:  0,
		RequiredXP: 100,
	})

	// Create book entity
	bookEntity := world.CreateEntity()
	book := &BookComponent{
		Title:    "Test Skill Book",
		Author:   "Test Author",
		BookType: BookTypeSkill,
		Content:  []string{"Page 1", "Page 2"},
		IsRead:   false,
		SkillBonus: map[string]float64{
			"combat": 0.5,
		},
	}
	bookEntity.AddComponent(book)

	// Process pending entity additions
	world.Update(0)

	// Read the book
	err := system.ReadBook(player.ID, bookEntity.ID)
	if err != nil {
		t.Fatalf("ReadBook failed: %v", err)
	}

	// Verify book is marked as read
	if !book.IsRead {
		t.Error("book not marked as read")
	}

	// Verify player has library component
	libComp, ok := player.GetComponent("library")
	if !ok {
		t.Fatal("player should have library component")
	}
	library, ok := libComp.(*LibraryComponent)
	if !ok {
		t.Fatal("invalid library component")
	}

	// Verify book is in library
	if len(library.Books) != 1 {
		t.Errorf("expected 1 book in library, got %d", len(library.Books))
	}
	if library.Books[0] != bookEntity.ID {
		t.Error("book ID not added to library")
	}

	// Verify XP was gained (0.5 * 100 = 50 XP)
	expComp, ok := player.GetComponent("experience")
	if !ok {
		t.Fatal("player should have experience component")
	}
	experience, ok := expComp.(*ExperienceComponent)
	if !ok {
		t.Fatal("invalid experience component")
	}
	if experience.CurrentXP != 50 {
		t.Errorf("expected 50 XP, got %d", experience.CurrentXP)
	}
}

func TestReadBookAlreadyRead(t *testing.T) {
	world := NewWorld()
	system := NewBookReadingSystem(world)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&ExperienceComponent{
		Level:      1,
		CurrentXP:  0,
		RequiredXP: 100,
	})

	// Create book entity (already read)
	bookEntity := world.CreateEntity()
	book := &BookComponent{
		Title:      "Test Book",
		Author:     "Test Author",
		BookType:   BookTypeLore,
		Content:    []string{"Page 1"},
		IsRead:     true,
		SkillBonus: make(map[string]float64),
	}
	bookEntity.AddComponent(book)

	// Process pending entity additions
	world.Update(0)

	// Try to read already read book
	err := system.ReadBook(player.ID, bookEntity.ID)
	if err != nil {
		t.Errorf("ReadBook should not error on already read book: %v", err)
	}

	// XP should still be 0 (no bonus applied)
	expComp, _ := player.GetComponent("experience")
	experience, _ := expComp.(*ExperienceComponent)
	if experience.CurrentXP != 0 {
		t.Error("XP should not change when reading already read book")
	}
}

func TestReadBookRecipeUnlock(t *testing.T) {
	world := NewWorld()
	system := NewBookReadingSystem(world)

	// Create player entity
	player := world.CreateEntity()

	// Create recipe book
	bookEntity := world.CreateEntity()
	book := &BookComponent{
		Title:      "Crafting Guide",
		Author:     "Master Crafter",
		BookType:   BookTypeRecipe,
		Content:    []string{"Recipe instructions"},
		IsRead:     false,
		RecipeID:   "iron_sword",
		SkillBonus: make(map[string]float64),
	}
	bookEntity.AddComponent(book)

	// Process pending entity additions
	world.Update(0)

	// Read the book
	err := system.ReadBook(player.ID, bookEntity.ID)
	if err != nil {
		t.Fatalf("ReadBook failed: %v", err)
	}

	// Verify recipe was unlocked
	knowledgeComp, ok := player.GetComponent("recipe_knowledge")
	if !ok {
		t.Fatal("player should have recipe knowledge component")
	}
	knowledge, ok := knowledgeComp.(*RecipeKnowledgeComponent)
	if !ok {
		t.Fatal("invalid recipe knowledge component")
	}

	if !knowledge.KnowsRecipe("iron_sword") {
		t.Error("recipe iron_sword should be known")
	}
}

func TestReadBookErrors(t *testing.T) {
	world := NewWorld()
	system := NewBookReadingSystem(world)

	tests := []struct {
		name     string
		setupFn  func() (playerID, bookID uint64)
		wantErr  bool
		errMatch string
	}{
		{
			name: "player not found",
			setupFn: func() (playerID, bookID uint64) {
				return 9999, 0
			},
			wantErr:  true,
			errMatch: "not found",
		},
		{
			name: "book not found",
			setupFn: func() (playerID, bookID uint64) {
				player := world.CreateEntity()
				return player.ID, 9999
			},
			wantErr:  true,
			errMatch: "not found",
		},
		{
			name: "entity is not a book",
			setupFn: func() (playerID, bookID uint64) {
				player := world.CreateEntity()
				notBook := world.CreateEntity()
				// Don't add book component
				return player.ID, notBook.ID
			},
			wantErr:  true,
			errMatch: "not a book",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playerID, bookID := tt.setupFn()
			err := system.ReadBook(playerID, bookID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSeriesCompletion(t *testing.T) {
	world := NewWorld()
	system := NewBookReadingSystem(world)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&ExperienceComponent{
		Level:      1,
		CurrentXP:  0,
		RequiredXP: 100,
	})

	// Process pending entity additions
	world.Update(0)

	// Create a series of books
	bookTitles := []string{
		"Adventures in Wonderland - Volume 1",
		"Adventures in Wonderland - Volume 2",
		"Adventures in Wonderland - Volume 3",
	}

	for _, title := range bookTitles {
		bookEntity := world.CreateEntity()
		book := &BookComponent{
			Title:      title,
			Author:     "Lewis Carroll",
			BookType:   BookTypeLore,
			Content:    []string{"Content"},
			IsRead:     false,
			SkillBonus: make(map[string]float64),
		}
		bookEntity.AddComponent(book)

		// Process pending entity additions
		world.Update(0)

		// Read the book
		err := system.ReadBook(player.ID, bookEntity.ID)
		if err != nil {
			t.Fatalf("ReadBook failed: %v", err)
		}
	}

	// Verify series completion
	libComp, ok := player.GetComponent("library")
	if !ok {
		t.Fatal("player should have library component")
	}
	library, ok := libComp.(*LibraryComponent)
	if !ok {
		t.Fatal("invalid library component")
	}

	seriesName := "Adventures in Wonderland"
	if !library.Completions[seriesName] {
		t.Error("series should be marked as complete")
	}

	// Verify bonus XP was awarded (100 XP per book * 3 books = 300 XP)
	expComp, _ := player.GetComponent("experience")
	experience, _ := expComp.(*ExperienceComponent)
	if experience.CurrentXP != 300 {
		t.Errorf("expected 300 XP from series completion, got %d", experience.CurrentXP)
	}
}

func TestExtractSeriesName(t *testing.T) {
	tests := []struct {
		title string
		want  string
	}{
		{"The Lord of the Rings - Volume 1", "The Lord of the Rings"},
		{"Harry Potter: Part 2", "Harry Potter"},
		{"Foundation - Part 3", "Foundation"},
		{"Standalone Book", ""},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			got := extractSeriesName(tt.title)
			if got != tt.want {
				t.Errorf("extractSeriesName(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestGetReadBooksCount(t *testing.T) {
	world := NewWorld()
	system := NewBookReadingSystem(world)

	// Create player entity
	player := world.CreateEntity()
	player.AddComponent(&LibraryComponent{
		Books:       []uint64{1, 2, 3},
		Completions: make(map[string]bool),
	})

	// Process pending entity additions
	world.Update(0)

	count := system.GetReadBooksCount(player.ID)
	if count != 3 {
		t.Errorf("expected 3 books, got %d", count)
	}

	// Test with non-existent player
	count = system.GetReadBooksCount(9999)
	if count != 0 {
		t.Errorf("expected 0 books for non-existent player, got %d", count)
	}
}

func TestGetCompletedSeries(t *testing.T) {
	world := NewWorld()
	system := NewBookReadingSystem(world)

	// Create player entity
	player := world.CreateEntity()
	completions := map[string]bool{
		"Series A": true,
		"Series B": true,
	}
	player.AddComponent(&LibraryComponent{
		Books:       []uint64{},
		Completions: completions,
	})

	// Process pending entity additions
	world.Update(0)

	series := system.GetCompletedSeries(player.ID)
	if len(series) != 2 {
		t.Errorf("expected 2 completed series, got %d", len(series))
	}

	// Verify both series are present
	foundA, foundB := false, false
	for _, s := range series {
		if s == "Series A" {
			foundA = true
		}
		if s == "Series B" {
			foundB = true
		}
	}
	if !foundA || !foundB {
		t.Error("not all series found in result")
	}
}
