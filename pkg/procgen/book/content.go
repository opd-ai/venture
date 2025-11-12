package book

import (
	"fmt"
	"strings"
)

// Grammar represents a text generation grammar with rules and expansions.
type Grammar struct {
	Rules map[string][]string
	rng   interface{ Intn(int) int }
}

// NewGrammar creates a new grammar for text generation.
func NewGrammar(rng interface{ Intn(int) int }) *Grammar {
	return &Grammar{
		Rules: make(map[string][]string),
		rng:   rng,
	}
}

// AddRule adds an expansion rule to the grammar.
func (g *Grammar) AddRule(symbol string, expansions []string) {
	g.Rules[symbol] = expansions
}

// Expand recursively expands a rule to generate text.
func (g *Grammar) Expand(symbol string) string {
	// Check if it's a rule reference (surrounded by #)
	if !strings.HasPrefix(symbol, "#") || !strings.HasSuffix(symbol, "#") {
		return symbol
	}

	// Remove # markers
	ruleName := strings.Trim(symbol, "#")

	// Get expansions for this rule
	expansions, ok := g.Rules[ruleName]
	if !ok || len(expansions) == 0 {
		return symbol // Return original if no rule found
	}

	// Pick random expansion
	expansion := expansions[g.rng.Intn(len(expansions))]

	// Recursively expand any embedded rules
	result := strings.Builder{}
	current := strings.Builder{}
	inRule := false

	for _, ch := range expansion {
		if ch == '#' {
			if inRule {
				// End of rule - expand it
				result.WriteString(g.Expand("#" + current.String() + "#"))
				current.Reset()
				inRule = false
			} else {
				// Start of rule
				inRule = true
			}
		} else {
			if inRule {
				current.WriteRune(ch)
			} else {
				result.WriteRune(ch)
			}
		}
	}

	return result.String()
}

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

// loadSkillGrammar loads grammar rules for skill books.
func (g *Generator) loadSkillGrammar(grammar *Grammar, genre, skillName string) {
	switch genre {
	case "fantasy":
		grammar.AddRule("skill_paragraph", []string{
			fmt.Sprintf("To master %s, one must first understand the fundamentals. #skill_technique# #skill_advice#", skillName),
			fmt.Sprintf("The ancient masters of %s knew that #skill_wisdom# Through diligent practice, you too can achieve mastery.", skillName),
			fmt.Sprintf("When practicing %s, remember to #skill_tip# This will greatly improve your abilities.", skillName),
			fmt.Sprintf("Advanced practitioners of %s should focus on #skill_advanced# This separates masters from novices.", skillName),
		})
		grammar.AddRule("skill_technique", []string{
			"Begin with proper stance and form.",
			"Focus your energy and maintain balance.",
			"The key lies in controlled movements.",
			"Precision matters more than raw power.",
		})
		grammar.AddRule("skill_advice", []string{
			"Practice daily to build muscle memory.",
			"Study the techniques of those who came before.",
			"Never underestimate the importance of preparation.",
			"True mastery requires patience and dedication.",
		})
		grammar.AddRule("skill_wisdom", []string{
			"power comes from harmony with the world around us.",
			"true strength is found within, not without.",
			"the mind and body must work as one.",
			"mastery is a journey, not a destination.",
		})
		grammar.AddRule("skill_tip", []string{
			"start slowly and gradually increase intensity",
			"maintain focus on your breathing",
			"visualize success before each attempt",
			"learn from each mistake rather than dwelling on it",
		})
		grammar.AddRule("skill_advanced", []string{
			"combining multiple techniques fluidly",
			"adapting to unpredictable situations",
			"developing your own personal style",
			"teaching others what you have learned",
		})

	case "sci-fi":
		grammar.AddRule("skill_paragraph", []string{
			fmt.Sprintf("Protocol for %s training module. #skill_instruction# #skill_technical#", skillName),
			fmt.Sprintf("System optimization for %s requires #skill_system# Refer to technical specifications for details.", skillName),
			fmt.Sprintf("%s enhancement protocol: #skill_enhancement# Performance metrics will be logged.", skillName),
			fmt.Sprintf("Advanced %s procedures involve #skill_procedure# Failure to comply may result in system errors.", skillName),
		})
		grammar.AddRule("skill_instruction", []string{
			"Initialize neural interface connection.",
			"Calibrate sensor arrays before proceeding.",
			"Run diagnostic tests to establish baseline.",
			"Verify system compatibility requirements.",
		})
		grammar.AddRule("skill_technical", []string{
			"Recommended training cycle: 500 iterations minimum.",
			"Expected improvement rate: 15% per session.",
			"Monitor biometric feedback during exercises.",
			"Adjust parameters based on performance data.",
		})
		grammar.AddRule("skill_system", []string{
			"proper hardware integration",
			"firmware version 3.2 or higher",
			"adequate power supply (minimum 500W)",
			"network latency below 50ms",
		})
		grammar.AddRule("skill_enhancement", []string{
			"Begin with Level 1 augmentation protocols",
			"Gradually increase neural load capacity",
			"Implement adaptive feedback systems",
			"Synchronize with central processing unit",
		})
		grammar.AddRule("skill_procedure", []string{
			"multi-threaded execution patterns",
			"parallel processing optimization",
			"real-time data stream analysis",
			"predictive algorithm implementation",
		})

	case "horror":
		grammar.AddRule("skill_paragraph", []string{
			fmt.Sprintf("The cursed art of %s demands a terrible price. #skill_warning# #skill_dark#", skillName),
			fmt.Sprintf("Those who seek %s must walk a dark path. #skill_forbidden# The shadows grow longer with each practice.", skillName),
			fmt.Sprintf("I have learned %s, though it cost me dearly. #skill_confession# May God have mercy on my soul.", skillName),
			fmt.Sprintf("The technique of %s was taught to me by something that should not exist. #skill_horror# I can never forget what I saw.", skillName),
		})
		grammar.AddRule("skill_warning", []string{
			"Do not attempt this after sundown.",
			"Never practice this technique alone.",
			"Some doors, once opened, cannot be closed.",
			"The voices in your head are not your own.",
		})
		grammar.AddRule("skill_dark", []string{
			"Blood must be spilled for each success.",
			"The pain will become unbearable, but you must endure.",
			"Each use brings you closer to damnation.",
			"Your reflection will betray the truth of what you've become.",
		})
		grammar.AddRule("skill_forbidden", []string{
			"They told me to stop, but I couldn't",
			"The whispers promised power beyond imagining",
			"I see things in the darkness now, watching me",
			"My hands remember movements I never learned",
		})
		grammar.AddRule("skill_confession", []string{
			"I can no longer sleep without the nightmares",
			"The marks on my skin won't fade",
			"Sometimes I forget which thoughts are mine",
			"I hear it calling to me, even now",
		})
		grammar.AddRule("skill_horror", []string{
			"Its eyes were empty, yet they saw everything",
			"The screaming stopped, but I still hear it",
			"Reality bent and twisted in ways that defy description",
			"Time lost all meaning in that place",
		})

	case "cyberpunk":
		grammar.AddRule("skill_paragraph", []string{
			fmt.Sprintf("Street guide to %s. #skill_street# #skill_hack#", skillName),
			fmt.Sprintf("Corporate training for %s is expensive. Here's the underground version. #skill_underground# Stay sharp, choom.", skillName),
			fmt.Sprintf("Jacking into %s requires the right chrome. #skill_chrome# Don't cheap out on the wetware.", skillName),
			fmt.Sprintf("Learned %s the hard way on the streets. #skill_lesson# This knowledge ain't free, but I'm sharing anyway.", skillName),
		})
		grammar.AddRule("skill_street", []string{
			"First rule: trust nobody.",
			"Corps want to keep this tech locked down.",
			"Black market mods work better anyway.",
			"Keep your ICE updated or get flatlined.",
		})
		grammar.AddRule("skill_hack", []string{
			"Bypass authentication with spoofed credentials.",
			"Side-channel attacks work on old systems.",
			"Physical access trumps all security.",
			"Social engineering is your best tool.",
		})
		grammar.AddRule("skill_underground", []string{
			"Find a ripperdoc you can trust",
			"Test everything before you plug it in",
			"The net remembers everything - be careful",
			"Never run hot without backup protocols",
		})
		grammar.AddRule("skill_chrome", []string{
			"Military-grade neural interfaces (hard to get)",
			"Reflex boosters (illegal in most districts)",
			"Synthetic muscle fibers (watch the rejection rate)",
			"Optical enhancements (don't go too cheap)",
		})
		grammar.AddRule("skill_lesson", []string{
			"Lost a friend who didn't follow these rules",
			"Spent three months in a corporate black site learning this",
			"Stole this from a corp database - cost me an arm (literally)",
			"Almost got zeroed before I figured this out",
		})

	case "post-apocalyptic":
		grammar.AddRule("skill_paragraph", []string{
			fmt.Sprintf("Survival guide for %s in the wasteland. #skill_survival# #skill_wasteland#", skillName),
			fmt.Sprintf("Before the bombs, they had schools for %s. Now we learn by doing. #skill_learn# Or we don't survive.", skillName),
			fmt.Sprintf("Old world knowledge of %s is rare. #skill_knowledge# Guard these words carefully.", skillName),
			fmt.Sprintf("My grandfather taught me %s before he passed. #skill_passed# This is how we keep humanity alive.", skillName),
		})
		grammar.AddRule("skill_survival", []string{
			"Water comes first, always.",
			"Never travel alone if you can help it.",
			"The old maps are mostly useless now.",
			"Trust your instincts - they'll keep you alive.",
		})
		grammar.AddRule("skill_wasteland", []string{
			"Radiation suits are worth their weight in bullets.",
			"The mutants are smarter than you think.",
			"Stay away from the dead zones.",
			"Scavenge carefully - traps are everywhere.",
		})
		grammar.AddRule("skill_learn", []string{
			"Books are precious, but experience is better",
			"Watch the old-timers and remember everything",
			"Mistakes in the wasteland are usually fatal",
			"Share knowledge freely - we're all in this together",
		})
		grammar.AddRule("skill_knowledge", []string{
			"Most pre-war tech is broken beyond repair",
			"Some things are better left forgotten",
			"The old ways still work, if you know how",
			"Adaptation is the only true survival skill",
		})
		grammar.AddRule("skill_passed", []string{
			"He learned it from the before-times",
			"These techniques kept our family alive",
			"Every generation must pass this on",
			"Without this knowledge, we're just animals",
		})

	default:
		grammar.AddRule("skill_paragraph", []string{
			fmt.Sprintf("To improve your %s, practice regularly. #skill_basic#", skillName),
			fmt.Sprintf("The fundamentals of %s are simple to learn but difficult to master. #skill_basic#", skillName),
			fmt.Sprintf("Advanced %s techniques require dedication. #skill_basic#", skillName),
		})
		grammar.AddRule("skill_basic", []string{
			"Focus on the basics before attempting advanced moves.",
			"Consistent practice yields the best results.",
			"Learn from experienced practitioners.",
			"Don't rush the learning process.",
		})
	}
}
