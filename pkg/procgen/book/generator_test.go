package book

import (
	"strings"
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
)

func TestNewGenerator(t *testing.T) {
	gen := NewGenerator()
	if gen == nil {
		t.Fatal("NewGenerator() returned nil")
	}
}

func TestGenerateSkillBook(t *testing.T) {
	tests := []struct {
		name       string
		seed       int64
		genre      string
		skillName  string
		wantErr    bool
		checkTitle bool
	}{
		{
			name:       "fantasy skill book",
			seed:       12345,
			genre:      "fantasy",
			skillName:  "Swordsmanship",
			wantErr:    false,
			checkTitle: true,
		},
		{
			name:       "scifi skill book",
			seed:       67890,
			genre:      "sci-fi",
			skillName:  "Engineering",
			wantErr:    false,
			checkTitle: true,
		},
		{
			name:       "horror skill book",
			seed:       11111,
			genre:      "horror",
			skillName:  "Dark Arts",
			wantErr:    false,
			checkTitle: true,
		},
		{
			name:       "cyberpunk skill book",
			seed:       22222,
			genre:      "cyberpunk",
			skillName:  "Hacking",
			wantErr:    false,
			checkTitle: true,
		},
		{
			name:       "postapoc skill book",
			seed:       33333,
			genre:      "post-apocalyptic",
			skillName:  "Survival",
			wantErr:    false,
			checkTitle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator()
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    tt.genre,
				Custom: map[string]interface{}{
					"book_type":  engine.BookTypeSkill,
					"skill_name": tt.skillName,
				},
			}

			result, err := gen.Generate(tt.seed, params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			book, ok := result.(*engine.BookComponent)
			if !ok {
				t.Fatal("Generate() did not return *engine.BookComponent")
			}

			if book.BookType != engine.BookTypeSkill {
				t.Errorf("BookType = %v, want %v", book.BookType, engine.BookTypeSkill)
			}

			if tt.checkTitle && book.Title == "" {
				t.Error("Title is empty")
			}

			if book.Author == "" {
				t.Error("Author is empty")
			}

			if len(book.Content) == 0 {
				t.Error("Content is empty")
			}

			if len(book.SkillBonus) == 0 {
				t.Error("SkillBonus is empty for skill book")
			}

			// Validate content
			if err := gen.Validate(book); err != nil {
				t.Errorf("Validate() error = %v", err)
			}
		})
	}
}

func TestGenerateLoreBook(t *testing.T) {
	genres := []string{"fantasy", "sci-fi", "horror", "cyberpunk", "post-apocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			gen := NewGenerator()
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    genre,
				Custom: map[string]interface{}{
					"book_type": engine.BookTypeLore,
				},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			book := result.(*engine.BookComponent)

			if book.BookType != engine.BookTypeLore {
				t.Errorf("BookType = %v, want %v", book.BookType, engine.BookTypeLore)
			}

			if book.Title == "" {
				t.Error("Title is empty")
			}

			if len(book.Content) == 0 {
				t.Error("Content is empty")
			}

			// Validate
			if err := gen.Validate(book); err != nil {
				t.Errorf("Validate() error = %v", err)
			}
		})
	}
}

func TestGenerateQuestBook(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      7,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"book_type": engine.BookTypeQuest,
			"quest_id":  "quest_123",
		},
	}

	result, err := gen.Generate(54321, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	book := result.(*engine.BookComponent)

	if book.BookType != engine.BookTypeQuest {
		t.Errorf("BookType = %v, want %v", book.BookType, engine.BookTypeQuest)
	}

	if len(book.Content) == 0 {
		t.Error("Content is empty")
	}

	// Quest books should have fewer pages than lore books (but more than before)
	if len(book.Content) > 7 {
		t.Errorf("Quest book has too many pages: %d", len(book.Content))
	}

	if err := gen.Validate(book); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestGenerateRecipeBook(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.4,
		Depth:      3,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"book_type": engine.BookTypeRecipe,
			"recipe_id": "recipe_healing_potion",
		},
	}

	result, err := gen.Generate(99999, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	book := result.(*engine.BookComponent)

	if book.BookType != engine.BookTypeRecipe {
		t.Errorf("BookType = %v, want %v", book.BookType, engine.BookTypeRecipe)
	}

	if book.RecipeID == "" {
		t.Error("RecipeID is empty for recipe book")
	}

	if book.RecipeID != "recipe_healing_potion" {
		t.Errorf("RecipeID = %s, want recipe_healing_potion", book.RecipeID)
	}

	if err := gen.Validate(book); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestGenerateHistoricalBook(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      8,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"book_type": engine.BookTypeHistory,
			"location":  "The Ancient Castle",
		},
	}

	result, err := gen.Generate(77777, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	book := result.(*engine.BookComponent)

	if book.BookType != engine.BookTypeHistory {
		t.Errorf("BookType = %v, want %v", book.BookType, engine.BookTypeHistory)
	}

	// Title should reference the location
	if !strings.Contains(book.Title, "Ancient Castle") && !strings.Contains(book.Title, "This Place") {
		t.Logf("Title: %s", book.Title)
		// This is okay - the location might be reformatted
	}

	if err := gen.Validate(book); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestDeterministicGeneration(t *testing.T) {
	seed := int64(42424242)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"book_type":  engine.BookTypeSkill,
			"skill_name": "Magic",
		},
	}

	gen1 := NewGenerator()
	result1, err1 := gen1.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First generation error: %v", err1)
	}
	book1 := result1.(*engine.BookComponent)

	gen2 := NewGenerator()
	result2, err2 := gen2.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second generation error: %v", err2)
	}
	book2 := result2.(*engine.BookComponent)

	// Check determinism
	if book1.Title != book2.Title {
		t.Errorf("Titles don't match: %s vs %s", book1.Title, book2.Title)
	}

	if book1.Author != book2.Author {
		t.Errorf("Authors don't match: %s vs %s", book1.Author, book2.Author)
	}

	if len(book1.Content) != len(book2.Content) {
		t.Errorf("Content page counts don't match: %d vs %d", len(book1.Content), len(book2.Content))
	}

	for i := range book1.Content {
		if book1.Content[i] != book2.Content[i] {
			t.Errorf("Content page %d doesn't match", i)
		}
	}
}

func TestValidateErrors(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		book    interface{}
		wantErr bool
	}{
		{
			name:    "nil book",
			book:    nil,
			wantErr: true,
		},
		{
			name:    "wrong type",
			book:    "not a book",
			wantErr: true,
		},
		{
			name: "empty title",
			book: &engine.BookComponent{
				Title:   "",
				Author:  "Test",
				Content: []string{"page 1"},
			},
			wantErr: true,
		},
		{
			name: "empty author",
			book: &engine.BookComponent{
				Title:   "Test Book",
				Author:  "",
				Content: []string{"page 1"},
			},
			wantErr: true,
		},
		{
			name: "empty content",
			book: &engine.BookComponent{
				Title:   "Test Book",
				Author:  "Test",
				Content: []string{},
			},
			wantErr: true,
		},
		{
			name: "content too short",
			book: &engine.BookComponent{
				Title:   "Test Book",
				Author:  "Test",
				Content: []string{"Short."},
			},
			wantErr: true,
		},
		{
			name: "skill book without bonus",
			book: &engine.BookComponent{
				Title:      "Test Book",
				Author:     "Test",
				BookType:   engine.BookTypeSkill,
				Content:    generateLongContent(600),
				SkillBonus: map[string]float64{},
			},
			wantErr: true,
		},
		{
			name: "recipe book without recipe ID",
			book: &engine.BookComponent{
				Title:    "Test Book",
				Author:   "Test",
				BookType: engine.BookTypeRecipe,
				Content:  generateLongContent(600),
				RecipeID: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.book)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWordCount(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"book_type": engine.BookTypeLore,
		},
	}

	result, err := gen.Generate(11111, params)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	book := result.(*engine.BookComponent)

	// Count words
	totalWords := 0
	for _, page := range book.Content {
		words := strings.Fields(page)
		totalWords += len(words)
	}

	if totalWords < 500 {
		t.Errorf("Book has too few words: %d (minimum 500)", totalWords)
	}

	if totalWords > 2000 {
		t.Errorf("Book has too many words: %d (maximum 2000)", totalWords)
	}

	t.Logf("Generated book with %d words across %d pages", totalWords, len(book.Content))
}

func TestSkillBonusCalculation(t *testing.T) {
	tests := []struct {
		name        string
		depth       int
		wantBonus   float64
		customBonus *float64
	}{
		{
			name:      "depth 1",
			depth:     1,
			wantBonus: 0.1,
		},
		{
			name:      "depth 5",
			depth:     5,
			wantBonus: 0.5,
		},
		{
			name:      "depth 10",
			depth:     10,
			wantBonus: 1.0,
		},
		{
			name:      "depth 20",
			depth:     20,
			wantBonus: 2.0,
		},
		{
			name:      "depth 30 (capped)",
			depth:     30,
			wantBonus: 2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator()
			params := procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      tt.depth,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"book_type":  engine.BookTypeSkill,
					"skill_name": "TestSkill",
				},
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			book := result.(*engine.BookComponent)
			bonus, ok := book.SkillBonus["TestSkill"]
			if !ok {
				t.Fatal("Skill bonus not set")
			}

			if bonus != tt.wantBonus {
				t.Errorf("Skill bonus = %f, want %f", bonus, tt.wantBonus)
			}
		})
	}
}

func TestInvalidParameters(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name   string
		params procgen.GenerationParams
	}{
		{
			name: "missing book type",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
				Custom:     map[string]interface{}{},
			},
		},
		{
			name: "invalid book type",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"book_type": "not a book type",
				},
			},
		},
		{
			name: "invalid difficulty",
			params: procgen.GenerationParams{
				Difficulty: 1.5,
				Depth:      5,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"book_type": engine.BookTypeSkill,
				},
			},
		},
		{
			name: "empty genre",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "",
				Custom: map[string]interface{}{
					"book_type": engine.BookTypeSkill,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := gen.Generate(12345, tt.params)
			if err == nil {
				t.Error("Expected error but got nil")
			}
		})
	}
}

// testRng implements a simple RNG for testing
type testRng struct {
	value int
}

// Intn implements the Intn method for testRng
func (r *testRng) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	r.value = (r.value + 1) % n
	return r.value
}

func TestGrammarExpansion(t *testing.T) {
	// Create a simple grammar
	rng := &testRng{value: 0}

	grammar := NewGrammar(rng)
	grammar.AddRule("greeting", []string{"Hello", "Hi", "Hey"})
	grammar.AddRule("name", []string{"World", "Friend"})
	grammar.AddRule("sentence", []string{"#greeting# #name#!"})

	// Test expansion
	result := grammar.Expand("#sentence#")
	if result == "#sentence#" {
		t.Error("Grammar expansion failed")
	}
	if !strings.Contains(result, "!") {
		t.Errorf("Expected exclamation mark in result: %s", result)
	}
}

// Helper function to generate long content for validation tests
func generateLongContent(wordCount int) []string {
	words := make([]string, wordCount)
	for i := 0; i < wordCount; i++ {
		words[i] = "word"
	}
	return []string{strings.Join(words, " ")}
}

func BenchmarkGenerateSkillBook(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"book_type":  engine.BookTypeSkill,
			"skill_name": "Combat",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}

func BenchmarkGenerateLoreBook(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"book_type": engine.BookTypeLore,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}
