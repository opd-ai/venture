// Package engine provides the book reading system for Phase 23.2.
// This system handles:
// - Reading books and marking them as read
// - Applying skill bonuses from skill books
// - Tracking book series and awarding completion bonuses
// - Managing book reading interactions
package engine

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// BookReadingSystem manages book reading and progression integration.
type BookReadingSystem struct {
	world  *World
	logger *logrus.Entry

	// Series completion bonuses (series name -> bonus multiplier)
	seriesBonuses map[string]float64
}

// NewBookReadingSystem creates a new book reading system.
func NewBookReadingSystem(world *World) *BookReadingSystem {
	logger := world.GetLogger()
	return &BookReadingSystem{
		world:         world,
		logger:        logger,
		seriesBonuses: make(map[string]float64),
	}
}

// Update processes book reading system updates.
// Currently a no-op as book reading is event-driven via ReadBook().
func (s *BookReadingSystem) Update(entities []*Entity, deltaTime float64) {
	// Book reading is handled via interaction system and ReadBook() method
}

// ReadBook processes reading a book entity by a player entity.
// Returns an error if reading fails.
func (s *BookReadingSystem) ReadBook(playerID, bookEntityID uint64) error {
	player, book, err := s.validateAndGetEntities(playerID, bookEntityID)
	if err != nil {
		return err
	}

	if book.IsRead {
		s.logBookAlreadyRead(playerID, bookEntityID, book.Title)
		return nil
	}

	book.IsRead = true

	library, err := s.getOrCreateLibrary(player)
	if err != nil {
		return err
	}

	s.addBookToLibrary(library, bookEntityID)

	if err := s.applyBookEffects(player, book); err != nil {
		return err
	}

	if err := s.checkSeriesCompletion(player, library, book); err != nil && s.logger != nil {
		s.logger.WithError(err).Warn("failed to check series completion")
	}

	s.logBookReadSuccess(playerID, bookEntityID, book)
	return nil
}

// validateAndGetEntities validates and retrieves player and book entities.
func (s *BookReadingSystem) validateAndGetEntities(playerID, bookEntityID uint64) (*Entity, *BookComponent, error) {
	player, ok := s.world.GetEntity(playerID)
	if !ok {
		return nil, nil, fmt.Errorf("player entity %d not found", playerID)
	}

	bookEntity, ok := s.world.GetEntity(bookEntityID)
	if !ok {
		return nil, nil, fmt.Errorf("book entity %d not found", bookEntityID)
	}

	bookComp, ok := bookEntity.GetComponent("book")
	if !ok {
		return nil, nil, fmt.Errorf("entity %d is not a book", bookEntityID)
	}

	book, ok := bookComp.(*BookComponent)
	if !ok {
		return nil, nil, fmt.Errorf("invalid book component")
	}

	return player, book, nil
}

// logBookAlreadyRead logs when a book has already been read.
func (s *BookReadingSystem) logBookAlreadyRead(playerID, bookEntityID uint64, title string) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID": playerID,
			"bookID":   bookEntityID,
			"title":    title,
		}).Debug("book already read")
	}
}

// getOrCreateLibrary gets or creates a library component for the player.
func (s *BookReadingSystem) getOrCreateLibrary(player *Entity) (*LibraryComponent, error) {
	if libComp, ok := player.GetComponent("library"); ok {
		library, ok := libComp.(*LibraryComponent)
		if !ok {
			return nil, fmt.Errorf("invalid library component")
		}
		return library, nil
	}

	library := &LibraryComponent{
		Books:       make([]uint64, 0),
		Completions: make(map[string]bool),
	}
	player.AddComponent(library)
	return library, nil
}

// addBookToLibrary adds a book to the library if not already present.
func (s *BookReadingSystem) addBookToLibrary(library *LibraryComponent, bookEntityID uint64) {
	for _, id := range library.Books {
		if id == bookEntityID {
			return
		}
	}
	library.Books = append(library.Books, bookEntityID)
}

// applyBookEffects applies book-specific effects based on book type.
func (s *BookReadingSystem) applyBookEffects(player *Entity, book *BookComponent) error {
	switch book.BookType {
	case BookTypeSkill:
		if err := s.applySkillBonuses(player, book); err != nil {
			return fmt.Errorf("failed to apply skill bonuses: %w", err)
		}
	case BookTypeRecipe:
		if err := s.unlockRecipe(player, book); err != nil {
			return fmt.Errorf("failed to unlock recipe: %w", err)
		}
	}
	return nil
}

// logBookReadSuccess logs successful book reading.
func (s *BookReadingSystem) logBookReadSuccess(playerID, bookEntityID uint64, book *BookComponent) {
	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID": playerID,
			"bookID":   bookEntityID,
			"title":    book.Title,
			"type":     book.BookType,
		}).Info("book read successfully")
	}
}

// applySkillBonuses applies skill bonuses from a skill book to the player.
func (s *BookReadingSystem) applySkillBonuses(player *Entity, book *BookComponent) error {
	// Get player's experience component
	expComp, ok := player.GetComponent("experience")
	if !ok {
		return fmt.Errorf("player has no experience component")
	}
	experience, ok := expComp.(*ExperienceComponent)
	if !ok {
		return fmt.Errorf("invalid experience component")
	}

	// Apply each skill bonus as XP
	for skillName, bonus := range book.SkillBonus {
		// Convert bonus to XP (bonus * 100 for meaningful XP gain)
		xpGain := int(bonus * 100)
		if xpGain > 0 {
			experience.AddXP(xpGain)

			if s.logger != nil {
				s.logger.WithFields(logrus.Fields{
					"playerID":  player.ID,
					"skillName": skillName,
					"bonus":     bonus,
					"xpGain":    xpGain,
					"newXP":     experience.CurrentXP,
				}).Info("skill bonus applied from book")
			}
		}
	}

	return nil
}

// unlockRecipe unlocks a crafting recipe for the player.
func (s *BookReadingSystem) unlockRecipe(player *Entity, book *BookComponent) error {
	if book.RecipeID == "" {
		return fmt.Errorf("book has no recipe ID")
	}

	// Get or create recipe knowledge component
	var knowledge *RecipeKnowledgeComponent
	if knowledgeComp, ok := player.GetComponent("recipe_knowledge"); ok {
		knowledge, ok = knowledgeComp.(*RecipeKnowledgeComponent)
		if !ok {
			return fmt.Errorf("invalid recipe knowledge component")
		}
	} else {
		// Create new recipe knowledge component
		knowledge = NewRecipeKnowledgeComponent(0) // 0 = unlimited slots
		player.AddComponent(knowledge)
	}

	// Check if recipe already known
	if knowledge.KnowsRecipe(book.RecipeID) {
		if s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"playerID": player.ID,
				"recipeID": book.RecipeID,
			}).Debug("recipe already known")
		}
		return nil
	}

	// Create a basic recipe from the book's recipe ID
	// The actual recipe details would come from a recipe generator
	recipe := &Recipe{
		ID:   book.RecipeID,
		Name: book.Title,
	}

	// Add recipe to known recipes
	knowledge.KnownRecipes[book.RecipeID] = recipe

	if s.logger != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID": player.ID,
			"recipeID": book.RecipeID,
		}).Info("recipe unlocked from book")
	}

	return nil
}

// checkSeriesCompletion checks if the player has completed any book series.
func (s *BookReadingSystem) checkSeriesCompletion(player *Entity, library *LibraryComponent, newBook *BookComponent) error {
	// Extract series name from book title (format: "Title - Volume N")
	// For simplicity, we'll consider books with similar prefixes as part of a series
	seriesName := extractSeriesName(newBook.Title)
	if seriesName == "" {
		return nil // Not part of a series
	}

	// Count books in this series
	seriesBooks := 0
	for _, bookID := range library.Books {
		bookEntity, ok := s.world.GetEntity(bookID)
		if !ok {
			continue
		}
		bookComp, ok := bookEntity.GetComponent("book")
		if !ok {
			continue
		}
		book, ok := bookComp.(*BookComponent)
		if !ok {
			continue
		}
		if extractSeriesName(book.Title) == seriesName {
			seriesBooks++
		}
	}

	// Check if series is complete (requires 3+ books for completion bonus)
	if seriesBooks >= 3 {
		// Check if already marked as complete
		if !library.Completions[seriesName] {
			// Mark as complete
			library.Completions[seriesName] = true

			// Apply completion bonus (10% additional XP for future skill books in series)
			s.seriesBonuses[seriesName] = 0.1

			// Award immediate XP bonus
			if expComp, ok := player.GetComponent("experience"); ok {
				if experience, ok := expComp.(*ExperienceComponent); ok {
					bonusXP := 100 * seriesBooks // 100 XP per book in series
					experience.AddXP(bonusXP)

					if s.logger != nil {
						s.logger.WithFields(logrus.Fields{
							"playerID":   player.ID,
							"seriesName": seriesName,
							"booksCount": seriesBooks,
							"bonusXP":    bonusXP,
						}).Info("book series completed!")
					}
				}
			}
		}
	}

	return nil
}

// extractSeriesName extracts the series name from a book title.
// Assumes series format: "Series Name - Volume N" or "Series Name: Part N"
func extractSeriesName(title string) string {
	// Look for common separators
	for _, sep := range []string{" - Volume ", " - Part ", ": Volume ", ": Part "} {
		if idx := findString(title, sep); idx != -1 {
			return title[:idx]
		}
	}
	return ""
}

// findString is a helper to find substring index (equivalent to strings.Index but inline)
func findString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// GetReadBooksCount returns the number of books the player has read.
func (s *BookReadingSystem) GetReadBooksCount(playerID uint64) int {
	player, ok := s.world.GetEntity(playerID)
	if !ok {
		return 0
	}

	libComp, ok := player.GetComponent("library")
	if !ok {
		return 0
	}
	library, ok := libComp.(*LibraryComponent)
	if !ok {
		return 0
	}

	return len(library.Books)
}

// GetCompletedSeries returns a list of completed book series for the player.
func (s *BookReadingSystem) GetCompletedSeries(playerID uint64) []string {
	player, ok := s.world.GetEntity(playerID)
	if !ok {
		return nil
	}

	libComp, ok := player.GetComponent("library")
	if !ok {
		return nil
	}
	library, ok := libComp.(*LibraryComponent)
	if !ok {
		return nil
	}

	series := make([]string, 0, len(library.Completions))
	for name := range library.Completions {
		series = append(series, name)
	}
	return series
}
