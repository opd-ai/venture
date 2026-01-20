package book

import (
	"strings"
)

// Code relocated from: content.go (Grammar struct and methods moved to grammar.go)

// generateSkillBookContent generates content for skill books.
func (g *Generator) generateSkillBookContent(genre string, difficulty float64, depth int, custom map[string]interface{}) []string {
	skillName, ok := custom["skill_name"].(string)
	if !ok || skillName == "" {
		skillName = "this skill"
	}

	// Create grammar for skill book
	grammar := NewGrammar(g.rng)
	g.loadSkillGrammar(grammar, genre, skillName)

	// Generate 5-7 pages based on depth (increased for more words)
	pageCount := 5 + (depth / 6)
	if pageCount < 5 {
		pageCount = 5
	}
	if pageCount > 7 {
		pageCount = 7
	}

	pages := make([]string, pageCount)
	for i := 0; i < pageCount; i++ {
		// Generate paragraphs for this page (4-6 paragraphs, increased for word count)
		paragraphCount := 4 + g.rng.Intn(3)
		paragraphs := make([]string, paragraphCount)

		for j := 0; j < paragraphCount; j++ {
			paragraphs[j] = grammar.Expand("#skill_paragraph#")
		}

		pages[i] = strings.Join(paragraphs, "\n\n")
	}

	return pages
}

// generateLoreContent generates content for lore books.
func (g *Generator) generateLoreContent(genre string, difficulty float64, depth int) []string {
	// Create grammar for lore
	grammar := NewGrammar(g.rng)
	g.loadLoreGrammar(grammar, genre)

	// Generate 5-7 pages (increased from 4-6)
	pageCount := 5 + g.rng.Intn(3)
	pages := make([]string, pageCount)

	for i := 0; i < pageCount; i++ {
		// Generate 4-6 paragraphs per page (increased from 3-5)
		paragraphCount := 4 + g.rng.Intn(3)
		paragraphs := make([]string, paragraphCount)

		for j := 0; j < paragraphCount; j++ {
			paragraphs[j] = grammar.Expand("#lore_paragraph#")
		}

		pages[i] = strings.Join(paragraphs, "\n\n")
	}

	return pages
}

// generateQuestContent generates content for quest journals.
func (g *Generator) generateQuestContent(genre string, difficulty float64, custom map[string]interface{}) []string {
	// Create grammar for quest journal
	grammar := NewGrammar(g.rng)
	g.loadQuestGrammar(grammar, genre)

	// Quest journals have 5-7 pages (increased for word count)
	pageCount := 5 + g.rng.Intn(3)
	pages := make([]string, pageCount)

	for i := 0; i < pageCount; i++ {
		// Generate 4-6 paragraphs per page (increased for word count)
		paragraphCount := 4 + g.rng.Intn(3)
		paragraphs := make([]string, paragraphCount)

		for j := 0; j < paragraphCount; j++ {
			paragraphs[j] = grammar.Expand("#quest_entry#")
		}

		pages[i] = strings.Join(paragraphs, "\n\n")
	}

	return pages
}

// generateRecipeContent generates content for recipe books.
func (g *Generator) generateRecipeContent(genre string, difficulty float64, custom map[string]interface{}) []string {
	// Create grammar for recipe
	grammar := NewGrammar(g.rng)
	g.loadRecipeGrammar(grammar, genre)

	// Recipe books have 4-6 pages (increased for word count)
	pageCount := 4 + g.rng.Intn(3)
	pages := make([]string, pageCount)

	for i := 0; i < pageCount; i++ {
		if i == 0 {
			// First page: introduction and requirements (more content)
			pages[i] = grammar.Expand("#recipe_intro#") + "\n\n" +
				grammar.Expand("#recipe_requirements#") + "\n\n" +
				grammar.Expand("#recipe_requirements#") + "\n\n" +
				grammar.Expand("#recipe_intro#")
		} else if i == pageCount-1 {
			// Last page: final steps and notes (more content)
			pages[i] = grammar.Expand("#recipe_steps#") + "\n\n" +
				grammar.Expand("#recipe_steps#") + "\n\n" +
				grammar.Expand("#recipe_notes#") + "\n\n" +
				grammar.Expand("#recipe_notes#")
		} else {
			// Middle pages: procedures (more steps)
			paragraphs := make([]string, 4)
			paragraphs[0] = grammar.Expand("#recipe_steps#")
			paragraphs[1] = grammar.Expand("#recipe_steps#")
			paragraphs[2] = grammar.Expand("#recipe_steps#")
			paragraphs[3] = grammar.Expand("#recipe_notes#")
			pages[i] = strings.Join(paragraphs, "\n\n")
		}
	}

	return pages
}

// generateHistoricalContent generates content for historical texts.
func (g *Generator) generateHistoricalContent(genre string, difficulty float64, depth int, custom map[string]interface{}) []string {
	location, ok := custom["location"].(string)
	if !ok || location == "" {
		location = "this place"
	}

	// Create grammar for history
	grammar := NewGrammar(g.rng)
	g.loadHistoryGrammar(grammar, genre, location)

	// Historical texts have 5-8 pages (increased from 4-7)
	pageCount := 5 + g.rng.Intn(4)
	pages := make([]string, pageCount)

	for i := 0; i < pageCount; i++ {
		// Generate 4-6 paragraphs per page (increased from 3-4)
		paragraphCount := 4 + g.rng.Intn(3)
		paragraphs := make([]string, paragraphCount)

		for j := 0; j < paragraphCount; j++ {
			paragraphs[j] = grammar.Expand("#history_paragraph#")
		}

		pages[i] = strings.Join(paragraphs, "\n\n")
	}

	return pages
}
