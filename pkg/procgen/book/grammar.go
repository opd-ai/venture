package book

import (
	"fmt"
	"strings"
)

// Grammar represents a text generation grammar with rules and expansions.
// Originally from: content.go
type Grammar struct {
	Rules map[string][]string
	rng   interface{ Intn(int) int }
}

// NewGrammar creates a new grammar for text generation.
// Originally from: content.go
func NewGrammar(rng interface{ Intn(int) int }) *Grammar {
	return &Grammar{
		Rules: make(map[string][]string),
		rng:   rng,
	}
}

// AddRule adds an expansion rule to the grammar.
// Originally from: content.go
func (g *Grammar) AddRule(symbol string, expansions []string) {
	g.Rules[symbol] = expansions
}

// Expand recursively expands a rule to generate text.
// Originally from: content.go
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

// Grammar loading functions
// Code relocated from: content.go, grammar_lore.go, grammar_recipe.go

// loadSkillGrammar loads grammar rules for skill books.
// Originally from: content.go
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
// loadLoreGrammar loads grammar rules for lore books.
func (g *Generator) loadLoreGrammar(grammar *Grammar, genre string) {
	switch genre {
	case "fantasy":
		grammar.AddRule("lore_paragraph", []string{
			"In the ancient days, #fantasy_event# The kingdoms of old #fantasy_consequence# and the world was forever changed.",
			"The #fantasy_artifact# was created by #fantasy_creator# Legend says that #fantasy_legend# but few believe this today.",
			"During the #fantasy_era# the people #fantasy_action# This led to #fantasy_result# which shaped our world.",
			"The chronicles tell of #fantasy_hero# who #fantasy_deed# Their legacy endures in #fantasy_legacy#",
		})
		grammar.AddRule("fantasy_event", []string{
			"dragons still walked among mortals",
			"magic flowed freely through the land",
			"the gods themselves walked the earth",
			"the first kingdoms were founded",
		})
		grammar.AddRule("fantasy_consequence", []string{
			"rose and fell like the tides",
			"warred endlessly for power",
			"formed alliances that would last millennia",
			"discovered the secrets of ancient magic",
		})
		grammar.AddRule("fantasy_artifact", []string{
			"Sword of Eternal Flame",
			"Crown of Starlight",
			"Staff of the Archmage",
			"Amulet of Protection",
		})
		grammar.AddRule("fantasy_creator", []string{
			"the first mages of the Silver Tower",
			"dwarven smiths in the deep mountains",
			"elven artificers of the ancient forest",
			"a forgotten god whose name is lost",
		})
		grammar.AddRule("fantasy_legend", []string{
			"it holds the power to reshape reality itself",
			"only the pure of heart can wield it",
			"it was forged in dragon fire",
			"a great price was paid for its creation",
		})
		grammar.AddRule("fantasy_era", []string{
			"Age of Dragons",
			"Time of the First Kings",
			"Era of Magic",
			"Dark Years",
		})
		grammar.AddRule("fantasy_action", []string{
			"fought against the darkness",
			"built great monuments to their gods",
			"discovered lost knowledge",
			"formed the first guilds",
		})
		grammar.AddRule("fantasy_result", []string{
			"the founding of modern civilization",
			"the sealing of the dark realms",
			"the establishment of magical law",
			"the golden age of peace",
		})
		grammar.AddRule("fantasy_hero", []string{
			"King Aldric the Brave",
			"Archmage Theron Starweaver",
			"Saint Elara the Pure",
			"General Morgath Ironhand",
		})
		grammar.AddRule("fantasy_deed", []string{
			"defeated the Dragon Lords",
			"sealed the portal to the abyss",
			"united the warring kingdoms",
			"discovered the lost city",
		})
		grammar.AddRule("fantasy_legacy", []string{
			"the laws that govern us today",
			"the traditions we still honor",
			"monuments that stand across the land",
			"songs sung by every child",
		})

	case "sci-fi":
		grammar.AddRule("lore_paragraph", []string{
			"Historical record indicates that #scifi_event# Colonial expansion #scifi_expansion# leading to significant technological advancement.",
			"The #scifi_technology# was developed in year #scifi_year# Research led by #scifi_scientist# resulted in #scifi_breakthrough#",
			"First Contact Protocol #scifi_protocol# established contact with #scifi_alien# Diplomatic relations #scifi_relations# over the following decades.",
			"The AI Singularity of #scifi_year# fundamentally altered human civilization. #scifi_ai_impact# Society adapted through #scifi_adaptation#",
		})
		grammar.AddRule("scifi_event", []string{
			"faster-than-light travel was achieved in 2247",
			"the Mars Colony was established in 2156",
			"artificial general intelligence emerged in 2198",
			"the first extrasolar colony was founded in 2289",
		})
		grammar.AddRule("scifi_expansion", []string{
			"proceeded rapidly throughout the solar system",
			"faced numerous technical challenges",
			"required new governance structures",
			"transformed human society fundamentally",
		})
		grammar.AddRule("scifi_technology", []string{
			"quantum entanglement communicator",
			"antimatter drive system",
			"neural interface protocol",
			"molecular fabricator",
		})
		grammar.AddRule("scifi_year", []string{
			"2187", "2203", "2245", "2276", "2298",
		})
		grammar.AddRule("scifi_scientist", []string{
			"Dr. Chen of the Luna Institute",
			"Professor Nakamura's research team",
			"the Titan Research Facility",
			"Director Rodriguez at Mars Base",
		})
		grammar.AddRule("scifi_breakthrough", []string{
			"revolutionary advances in energy production",
			"the elimination of resource scarcity",
			"humanity's first steps toward post-scarcity",
			"new understanding of physics",
		})
		grammar.AddRule("scifi_protocol", []string{
			"Alpha-7", "Omega-3", "Delta-9", "Gamma-5",
		})
		grammar.AddRule("scifi_alien", []string{
			"the silicon-based Crystalline Intelligence",
			"the aquatic Deepholders of Kepler-442b",
			"the hive-mind Collective",
			"an advanced AI civilization",
		})
		grammar.AddRule("scifi_relations", []string{
			"evolved from tentative to collaborative",
			"remained strained but functional",
			"flourished through cultural exchange",
			"required careful diplomatic navigation",
		})
		grammar.AddRule("scifi_ai_impact", []string{
			"Labor became voluntary rather than necessary.",
			"Human creativity reached unprecedented heights.",
			"New questions arose about consciousness and rights.",
			"Society reorganized around post-work paradigms.",
		})
		grammar.AddRule("scifi_adaptation", []string{
			"universal basic resources and education",
			"new forms of democratic participation",
			"the establishment of AI-human councils",
			"fundamental restructuring of economics",
		})

	case "horror":
		grammar.AddRule("lore_paragraph", []string{
			"The texts speak of #horror_entity# that #horror_action# Witnesses described #horror_description# before madness took them.",
			"In #horror_year# something terrible occurred. #horror_incident# The survivors never spoke of it again.",
			"Ancient warnings tell us that #horror_warning# But we did not listen. #horror_consequence# Now it is too late.",
			"The cult of #horror_cult# believed that #horror_belief# Their rituals #horror_ritual# and reality began to break.",
		})
		grammar.AddRule("horror_entity", []string{
			"a nameless thing in the darkness",
			"the Whispering Presence",
			"something that should not exist",
			"the thing that hungers",
		})
		grammar.AddRule("horror_action", []string{
			"dwells between the spaces of our world",
			"feeds on human consciousness",
			"corrupts all it touches",
			"waits beyond the veil of sanity",
		})
		grammar.AddRule("horror_description", []string{
			"geometries that hurt to perceive",
			"sounds that echoed in dimensions we cannot fathom",
			"a presence that defied all natural law",
			"forms that shifted when observed directly",
		})
		grammar.AddRule("horror_year", []string{
			"1873", "1924", "1967", "1989",
		})
		grammar.AddRule("horror_incident", []string{
			"The entire town vanished overnight.",
			"Those who entered never returned unchanged.",
			"Reality itself seemed to fracture.",
			"The screaming lasted for three days.",
		})
		grammar.AddRule("horror_warning", []string{
			"some doors must never be opened",
			"certain names must never be spoken",
			"the old places should remain undisturbed",
			"there are truths humanity was not meant to know",
		})
		grammar.AddRule("horror_consequence", []string{
			"The darkness spreads.",
			"Madness has taken root.",
			"They are coming through.",
			"The end has begun.",
		})
		grammar.AddRule("horror_cult", []string{
			"the Obsidian Circle",
			"the Whispering Order",
			"the Crimson Covenant",
			"the Shadow Sect",
		})
		grammar.AddRule("horror_belief", []string{
			"death was merely a doorway",
			"gods slept beneath the earth",
			"reality was a thin membrane",
			"sacrifice would bring power",
		})
		grammar.AddRule("horror_ritual", []string{
			"tore holes in the fabric of space",
			"summoned things best left forgotten",
			"opened portals to nightmare realms",
			"called to entities beyond comprehension",
		})

	case "cyberpunk":
		grammar.AddRule("lore_paragraph", []string{
			"Net history shows that #cyber_corp# launched #cyber_product# in #cyber_year# Market response #cyber_market# reshaping the digital landscape.",
			"The #cyber_event# changed everything. #cyber_before# but after, #cyber_after# Streets were never the same.",
			"Corporate warfare between #cyber_corp# and #cyber_rival# resulted in #cyber_result# Thousands of users #cyber_impact#",
			"Underground archives reveal that #cyber_secret# The truth was #cyber_truth# but the corps buried it deep.",
		})
		grammar.AddRule("cyber_corp", []string{
			"OmniCorp", "NeuralTech", "SynthSec", "DataDyne", "MegaCorp",
		})
		grammar.AddRule("cyber_product", []string{
			"the first mass-market neural interface",
			"revolutionary biometric authentication",
			"full-sensory VR immersion technology",
			"quantum encryption protocols",
		})
		grammar.AddRule("cyber_year", []string{
			"2045", "2051", "2058", "2063", "2069",
		})
		grammar.AddRule("cyber_market", []string{
			"exceeded all projections",
			"triggered regulatory intervention",
			"created new underground economies",
			"fundamentally altered human interaction",
		})
		grammar.AddRule("cyber_event", []string{
			"Great Net Crash of '58",
			"Corporate Wars",
			"Neural Interface Plague",
			"AI Liberation Movement",
		})
		grammar.AddRule("cyber_before", []string{
			"The net was open and free",
			"Privacy still meant something",
			"People controlled their own data",
			"Democracy had a chance",
		})
		grammar.AddRule("cyber_after", []string{
			"corporations owned everything",
			"anonymity became a crime",
			"freedom was a memory",
			"resistance went underground",
		})
		grammar.AddRule("cyber_rival", []string{
			"Global Security Solutions", "TechCorp International", "Neo Systems", "Digital Dominion",
		})
		grammar.AddRule("cyber_result", []string{
			"the collapse of the old net infrastructure",
			"massive civilian casualties",
			"the birth of the dark net",
			"permanent changes to global power structures",
		})
		grammar.AddRule("cyber_impact", []string{
			"lost their neural interfaces permanently",
			"were caught in the digital crossfire",
			"had their identities stolen or erased",
			"became refugees in the virtual world",
		})
		grammar.AddRule("cyber_secret", []string{
			"the first AI wasn't created - it evolved",
			"user data was weaponized from day one",
			"the neural interface had hidden backdoors",
			"corporations experimented on unwilling subjects",
		})
		grammar.AddRule("cyber_truth", []string{
			"leaked by a whistleblower who disappeared",
			"hidden in classified government archives",
			"discovered by hackers years later",
			"suppressed through legal warfare",
		})

	case "post-apocalyptic":
		grammar.AddRule("lore_paragraph", []string{
			"Before the Fall, #prewar_world# People lived #prewar_life# without knowing #prewar_ignorance#",
			"The bombs fell on #apocalypse_date# Within #apocalypse_duration# the old world was gone. #apocalypse_result#",
			"Survivors tell of #survival_story# In those early days, #survival_challenge# Only the strongest #survival_outcome#",
			"The wasteland teaches harsh lessons. #wasteland_lesson# Communities that forget this #wasteland_fate#",
		})
		grammar.AddRule("prewar_world", []string{
			"civilization reached its peak",
			"technology advanced beyond wisdom",
			"humanity believed itself invincible",
			"the world was interconnected and thriving",
		})
		grammar.AddRule("prewar_life", []string{
			"in comfort and abundance",
			"without fear of starvation",
			"connected by global networks",
			"with medicine and safety",
		})
		grammar.AddRule("prewar_ignorance", []string{
			"how fragile it all was",
			"that the end was coming",
			"what would be lost forever",
			"how good they had it",
		})
		grammar.AddRule("apocalypse_date", []string{
			"a Tuesday in October",
			"the first day of winter",
			"the anniversary of some forgotten holiday",
			"a morning like any other",
		})
		grammar.AddRule("apocalypse_duration", []string{
			"hours",
			"a single day",
			"three terrible days",
			"a week of fire",
		})
		grammar.AddRule("apocalypse_result", []string{
			"Ash blocked out the sun for years.",
			"The radiation made whole regions uninhabitable.",
			"Billions died in the initial blast.",
			"Civilization collapsed overnight.",
		})
		grammar.AddRule("survival_story", []string{
			"the first terrible winter",
			"when the water ran out",
			"the fight for the last food stores",
			"when the dead walked",
		})
		grammar.AddRule("survival_challenge", []string{
			"humanity showed its true nature",
			"moral choices became impossible",
			"survival meant doing terrible things",
			"communities fractured and reformed",
		})
		grammar.AddRule("survival_outcome", []string{
			"made it through",
			"adapted to the new world",
			"learned to scavenge and survive",
			"formed the first settlements",
		})
		grammar.AddRule("wasteland_lesson", []string{
			"Trust is earned, never given freely.",
			"The old rules no longer apply.",
			"Everything is a resource, including people.",
			"Survival trumps morality every time.",
		})
		grammar.AddRule("wasteland_fate", []string{
			"don't survive long",
			"become prey for others",
			"fade into the wastes",
			"are swept away by raiders",
		})

	default:
		grammar.AddRule("lore_paragraph", []string{
			"Long ago, important events shaped our world. #generic_event# This led to significant changes.",
			"Historical records indicate that #generic_people# accomplished great things. #generic_result#",
			"Ancient wisdom teaches us that #generic_wisdom# We must remember these lessons.",
		})
		grammar.AddRule("generic_event", []string{
			"Heroes rose to face great challenges.",
			"Civilizations rose and fell.",
			"Knowledge was passed down through generations.",
		})
		grammar.AddRule("generic_people", []string{
			"our ancestors", "the ancient ones", "those who came before",
		})
		grammar.AddRule("generic_result", []string{
			"Their legacy endures to this day.",
			"We benefit from their wisdom now.",
			"The world was changed forever.",
		})
		grammar.AddRule("generic_wisdom", []string{
			"knowledge is power",
			"unity brings strength",
			"preparation prevents disaster",
		})
	}
}

// loadQuestGrammar loads grammar rules for quest journals.
func (g *Generator) loadQuestGrammar(grammar *Grammar, genre string) {
	switch genre {
	case "fantasy":
		grammar.AddRule("quest_entry", []string{
			"Day #quest_day#: #quest_progress# The path ahead remains #quest_difficulty#",
			"Found #quest_clue# which suggests #quest_hint# Must investigate further.",
			"Encountered #quest_obstacle# Today. #quest_reaction# Will need to #quest_plan#",
		})
		grammar.AddRule("quest_day", []string{
			"3", "7", "12", "15", "21",
		})
		grammar.AddRule("quest_progress", []string{
			"Made good progress toward the objective",
			"Suffered a setback but pressed onward",
			"Discovered crucial information",
			"The journey continues despite challenges",
		})
		grammar.AddRule("quest_difficulty", []string{
			"treacherous", "unclear", "fraught with danger", "promising",
		})
		grammar.AddRule("quest_clue", []string{
			"an ancient map", "a cryptic message", "evidence of passage", "a hidden symbol",
		})
		grammar.AddRule("quest_hint", []string{
			"the artifact is nearby",
			"someone was here recently",
			"danger lies ahead",
			"we're on the right track",
		})
		grammar.AddRule("quest_obstacle", []string{
			"hostile creatures", "a locked door", "treacherous terrain", "rival adventurers",
		})
		grammar.AddRule("quest_reaction", []string{
			"It was difficult but manageable.",
			"Nearly cost us everything.",
			"Learned valuable lessons.",
			"Barely escaped with our lives.",
		})
		grammar.AddRule("quest_plan", []string{
			"find another route",
			"prepare better equipment",
			"seek local knowledge",
			"recruit assistance",
		})

	case "sci-fi":
		grammar.AddRule("quest_entry", []string{
			"Mission log, day #quest_day#: #quest_status# Parameters #quest_parameters#",
			"Sensor readings indicate #quest_data# Analysis suggests #quest_analysis#",
			"Encountered #quest_anomaly# Protocol requires #quest_protocol#",
		})
		grammar.AddRule("quest_day", []string{
			"05", "12", "18", "23", "31",
		})
		grammar.AddRule("quest_status", []string{
			"Mission proceeding as planned.",
			"Unexpected complications have arisen.",
			"Significant progress achieved.",
			"Requiring mission parameter adjustment.",
		})
		grammar.AddRule("quest_parameters", []string{
			"within acceptable ranges",
			"require recalibration",
			"optimal for objective completion",
			"deviating from baseline",
		})
		grammar.AddRule("quest_data", []string{
			"unusual energy signatures",
			"structural anomalies",
			"unidentified objects",
			"atmospheric disturbances",
		})
		grammar.AddRule("quest_analysis", []string{
			"artificial origin",
			"natural phenomenon",
			"potential danger",
			"mission-critical information",
		})
		grammar.AddRule("quest_anomaly", []string{
			"system malfunction",
			"unknown entities",
			"spatial distortion",
			"communication interference",
		})
		grammar.AddRule("quest_protocol", []string{
			"immediate investigation",
			"defensive measures",
			"data collection",
			"command notification",
		})

	default:
		grammar.AddRule("quest_entry", []string{
			"Progress update: #quest_generic# Continuing forward.",
			"Today: #quest_event# This changes things.",
			"Note: #quest_observation# Important to remember.",
		})
		grammar.AddRule("quest_generic", []string{
			"Made steady progress.",
			"Encountered challenges.",
			"Discovered something important.",
		})
		grammar.AddRule("quest_event", []string{
			"Found what we were looking for.",
			"Had to change our plans.",
			"Learned something unexpected.",
		})
		grammar.AddRule("quest_observation", []string{
			"The situation is more complex than expected.",
			"Resources are limited.",
			"Time is running out.",
		})
	}
}

// loadRecipeGrammar loads grammar rules for recipe books.
func (g *Generator) loadRecipeGrammar(grammar *Grammar, genre string) {
	switch genre {
	case "fantasy":
		grammar.AddRule("recipe_intro", []string{
			"This tome contains the secrets of #recipe_craft# passed down through generations of masters.",
			"Within these pages lie the instructions for creating #recipe_item# of exceptional quality.",
			"Learn the ancient art of #recipe_craft# as practiced by the greatest craftsmen of our age.",
		})
		grammar.AddRule("recipe_requirements", []string{
			"Required materials: #recipe_materials# Gather these components before beginning. The quality of materials directly affects the final result.",
			"You will need: #recipe_materials# All components must be fresh and properly prepared. Substitutions are not recommended.",
			"Essential ingredients: #recipe_materials# Handle each component with care, for they are imbued with latent power.",
		})
		grammar.AddRule("recipe_steps", []string{
			"Begin by #recipe_action# Pay close attention to the temperature and timing. #recipe_warning#",
			"Next, #recipe_action# This step requires patience and precision. #recipe_tip#",
			"Carefully #recipe_action# The outcome depends on proper technique here. #recipe_detail#",
		})
		grammar.AddRule("recipe_notes", []string{
			"Master craftsmen recommend #recipe_advice# This will greatly improve your results. With practice, you will develop your own techniques.",
			"Note: #recipe_warning# Many apprentices make this mistake. Learn from their errors and yours will be fewer.",
			"Advanced practitioners suggest #recipe_tip# This knowledge comes from years of experience in the craft.",
		})
		grammar.AddRule("recipe_craft", []string{
			"potion brewing", "weapon smithing", "armor crafting", "enchanting", "alchemy",
		})
		grammar.AddRule("recipe_item", []string{
			"healing elixirs", "enchanted blades", "protective amulets", "powerful potions",
		})
		grammar.AddRule("recipe_materials", []string{
			"dragon scales, moonstone dust, and purified water",
			"iron ore, charcoal, and leather strips",
			"herbs from the sacred grove, crystal vials, and spring water",
			"silver wire, gemstones, and blessed oils",
		})
		grammar.AddRule("recipe_action", []string{
			"heating the mixture to precisely 300 degrees",
			"grinding the components into fine powder",
			"infusing the base with magical energy",
			"combining ingredients in the correct sequence",
		})
		grammar.AddRule("recipe_warning", []string{
			"Overheating will ruin the entire batch.",
			"Never skip this step or the result will be unstable.",
			"Rushing this process leads to failure.",
		})
		grammar.AddRule("recipe_tip", []string{
			"Adding a pinch of salt enhances the effect.",
			"Clockwise stirring produces better results.",
			"Morning dew works better than regular water.",
		})
		grammar.AddRule("recipe_detail", []string{
			"The mixture should turn golden when ready.",
			"You will hear a soft chiming sound at completion.",
			"The aroma indicates proper preparation.",
		})
		grammar.AddRule("recipe_advice", []string{
			"practicing the motions before working with materials",
			"keeping your workspace clean and organized",
			"starting with simpler recipes to build skill",
		})

	case "sci-fi":
		grammar.AddRule("recipe_intro", []string{
			"Technical specifications for #recipe_tech# manufacturing. Clearance Level: #recipe_clearance#",
			"Assembly protocol for #recipe_device# requires precise adherence to specifications.",
			"Fabrication guide for #recipe_product# Unauthorized reproduction is prohibited.",
		})
		grammar.AddRule("recipe_requirements", []string{
			"Component list: #recipe_components# All parts must meet ISO-9001 standards. Verify serial numbers before assembly.",
			"Required equipment: #recipe_tools# Ensure proper calibration before beginning. Tolerance: +/- 0.001mm.",
			"Materials needed: #recipe_materials# Source only from approved vendors. Substitutions void warranty.",
		})
		grammar.AddRule("recipe_steps", []string{
			"Step #recipe_step#: #recipe_procedure# Verify output before proceeding. Error threshold: <0.1%.",
			"Procedure #recipe_step#: #recipe_technical# Monitor sensor readings continuously. Abort if parameters exceed safe limits.",
			"Phase #recipe_step#: #recipe_assembly# Quality check required at this stage. Document all measurements.",
		})
		grammar.AddRule("recipe_notes", []string{
			"WARNING: #recipe_safety# Failure to follow safety protocols may result in equipment damage or injury.",
			"Technical note: #recipe_optimization# Performance gains of 15-20% have been observed.",
			"Troubleshooting: #recipe_debug# Refer to error codes in Appendix C for detailed diagnostics.",
		})
		grammar.AddRule("recipe_tech", []string{
			"neural interface modules",
			"quantum processors",
			"antimatter containment units",
			"graviton emitters",
		})
		grammar.AddRule("recipe_clearance", []string{
			"Alpha-3", "Beta-7", "Gamma-2", "Delta-5",
		})
		grammar.AddRule("recipe_device", []string{
			"portable fusion reactor",
			"holographic display matrix",
			"emergency life support system",
			"subspace communication array",
		})
		grammar.AddRule("recipe_product", []string{
			"cybernetic enhancement modules",
			"portable force field generators",
			"molecular assemblers",
			"quantum entanglement communicators",
		})
		grammar.AddRule("recipe_components", []string{
			"Processor Unit (PN: QC-7742), Memory Banks (512 Petabyte), Power Coupling (Type-C)",
			"Optical Fibers (Grade-A), Superconducting Coils, Coolant System (Model RS-9)",
			"Nano-assemblers (Gen-4), Carbon Nanotube Framework, Energy Matrix (Class-7)",
		})
		grammar.AddRule("recipe_tools", []string{
			"molecular welder, precision laser cutter, electron microscope",
			"clean room facility (Class 1000), automated assembly station",
			"3D molecular printer, vacuum chamber, spectrometer",
		})
		grammar.AddRule("recipe_materials", []string{
			"titanium-steel alloy (99.9% purity), optical-grade silicon wafers, superconducting ceramic composites",
			"carbon nanotube sheets, graphene layers, rare earth elements",
			"artificial diamond substrates, metallic hydrogen cells, exotic matter samples",
		})
		grammar.AddRule("recipe_step", []string{
			"1", "2", "3", "4", "5",
		})
		grammar.AddRule("recipe_procedure", []string{
			"Initialize fabrication sequence",
			"Apply molecular bonding agent",
			"Activate quantum alignment protocol",
		})
		grammar.AddRule("recipe_technical", []string{
			"Maintain temperature at 4 Kelvin",
			"Apply electromagnetic field at 50 Tesla",
			"Inject nanobots for assembly",
		})
		grammar.AddRule("recipe_assembly", []string{
			"Connect primary power conduits",
			"Seal containment chamber",
			"Install control interface",
		})
		grammar.AddRule("recipe_safety", []string{
			"High voltage present - do not open while powered",
			"Radiation hazard - use proper shielding",
			"Extreme temperatures - allow cooldown period",
		})
		grammar.AddRule("recipe_optimization", []string{
			"Pre-cooling components improves bond strength",
			"Parallel processing reduces assembly time",
			"Using higher-grade materials extends lifespan",
		})
		grammar.AddRule("recipe_debug", []string{
			"Error code E-401 indicates calibration drift",
			"Intermittent signals suggest loose connections",
			"Power fluctuations may require capacitor replacement",
		})

	case "horror":
		grammar.AddRule("recipe_intro", []string{
			"These forbidden instructions detail the creation of #recipe_cursed# May God forgive what you are about to do.",
			"The ritual for crafting #recipe_unholy# should never be attempted. I write this only so others may understand what I have done.",
			"Dark knowledge: how to make #recipe_forbidden# The price is higher than you know.",
		})
		grammar.AddRule("recipe_requirements", []string{
			"You will need: #recipe_dark_materials# Obtain these under cover of darkness. Some are best left unspoken.",
			"Required components: #recipe_profane# Each one damns you further. There is no turning back after this.",
			"Gather: #recipe_cursed_items# The sources are... unnatural. Do not ask where I found them.",
		})
		grammar.AddRule("recipe_steps", []string{
			"When the moon is dark, #recipe_ritual# #recipe_horror# I can still hear the sounds.",
			"At midnight, #recipe_dark_action# #recipe_dread# My hands shake even now, remembering.",
			"During the witching hour, #recipe_profane_act# #recipe_terror# Some things cannot be unseen.",
		})
		grammar.AddRule("recipe_notes", []string{
			"WARNING: #recipe_warning# I learned this too late. The voices never stop.",
			"Do not, under any circumstances, #recipe_prohibition# Three have died ignoring this. I will be the fourth.",
			"A note of caution: #recipe_madness# They watch through the cracks in reality now.",
		})
		grammar.AddRule("recipe_cursed", []string{
			"items that hunger",
			"weapons that remember",
			"armor that feeds",
			"potions that whisper",
		})
		grammar.AddRule("recipe_unholy", []string{
			"the Blade of Endless Night",
			"the Amulet of Screaming Souls",
			"the Elixir of Undeath",
			"the Crown of Madness",
		})
		grammar.AddRule("recipe_forbidden", []string{
			"things that should not exist",
			"doors to places we cannot comprehend",
			"vessels for entities beyond naming",
			"keys to realms of eternal suffering",
		})
		grammar.AddRule("recipe_dark_materials", []string{
			"grave dust from a murderer's tomb, blood freely given, a name that must not be spoken",
			"bones that walked again, tears of the innocent, iron forged in human suffering",
			"ashes of the unburied, water that never knew sunlight, hair from the condemned",
		})
		grammar.AddRule("recipe_profane", []string{
			"flesh that still moves, eyes that still see, a soul fragment",
			"something that died screaming, proof of betrayal, the last breath",
			"a binding written in pain, symbols that burn to look upon, essence of nightmare",
		})
		grammar.AddRule("recipe_cursed_items", []string{
			"items from murder scenes, relics of atrocity, fragments of broken minds",
			"stolen grave goods, cursed heirlooms, objects of obsession",
			"things touched by tragedy, remnants of ritual sacrifice, vessels of despair",
		})
		grammar.AddRule("recipe_ritual", []string{
			"speak the words that should remain unspoken",
			"perform the gestures learned from dreams",
			"call to things that wait in darkness",
		})
		grammar.AddRule("recipe_horror", []string{
			"Something answered.",
			"Reality cracked.",
			"They came through.",
		})
		grammar.AddRule("recipe_dark_action", []string{
			"combine the components in the old way",
			"trace the symbols in your own blood",
			"open yourself to the void",
		})
		grammar.AddRule("recipe_dread", []string{
			"I felt it watching me.",
			"Time moved strangely.",
			"Part of me stayed there.",
		})
		grammar.AddRule("recipe_profane_act", []string{
			"complete the final binding",
			"seal the pact in flesh",
			"offer the sacrifice",
		})
		grammar.AddRule("recipe_terror", []string{
			"I am changed forever.",
			"It knows my true name.",
			"I see them in mirrors now.",
		})
		grammar.AddRule("recipe_warning", []string{
			"The creation will turn on its maker eventually",
			"Using this brings them closer to our world",
			"Each use feeds something hungry in the dark",
		})
		grammar.AddRule("recipe_prohibition", []string{
			"use this during a full moon",
			"create more than one at a time",
			"speak its name aloud",
		})
		grammar.AddRule("recipe_madness", []string{
			"The process changes you in ways you cannot see",
			"Your dreams will never be your own again",
			"Something follows those who craft these things",
		})

	default:
		grammar.AddRule("recipe_intro", []string{
			"This guide explains how to craft useful items. Follow instructions carefully.",
			"Instructions for creating quality products. Precision is important.",
		})
		grammar.AddRule("recipe_requirements", []string{
			"Materials needed: basic components. Gather everything before starting.",
			"Required items: standard supplies. Quality materials produce better results.",
		})
		grammar.AddRule("recipe_steps", []string{
			"Step: Combine materials properly. Follow the sequence exactly.",
			"Next: Process components carefully. Take your time with each stage.",
		})
		grammar.AddRule("recipe_notes", []string{
			"Tips: Practice makes perfect. Don't be discouraged by initial failures.",
			"Remember: Safety first. Work in appropriate conditions.",
		})
	}
}

// loadHistoryGrammar loads grammar rules for historical texts.
func (g *Generator) loadHistoryGrammar(grammar *Grammar, genre, location string) {
	switch genre {
	case "fantasy":
		grammar.AddRule("history_paragraph", []string{
			fmt.Sprintf("%s was built #history_when# by #history_builder# Its purpose was to #history_purpose#", location),
			fmt.Sprintf("In the annals of %s, #history_event# occurred. #history_consequence# The effects are felt even today.", location),
			fmt.Sprintf("The architecture of %s reveals #history_secret# Scholars have debated #history_debate# for centuries.", location),
			fmt.Sprintf("During the #history_era# %s #history_role# Many famous #history_figures# walked these halls.", location),
		})
		grammar.AddRule("history_when", []string{
			"in the First Age",
			"during the reign of King Aldric",
			"after the Dragon Wars",
			"in ancient times",
		})
		grammar.AddRule("history_builder", []string{
			"the archmages of the Crystal Tower",
			"dwarven master craftsmen",
			"elven architects of renown",
			"a forgotten civilization",
		})
		grammar.AddRule("history_purpose", []string{
			"protect against the darkness",
			"house a great magical artifact",
			"serve as a center of learning",
			"defend the realm from invasion",
		})
		grammar.AddRule("history_event", []string{
			"a great battle",
			"a historic treaty signing",
			"a magical catastrophe",
			"a royal coronation",
		})
		grammar.AddRule("history_consequence", []string{
			"The structure was damaged but stood firm.",
			"Power shifted throughout the realm.",
			"New alliances were forged.",
			"The place gained legendary status.",
		})
		grammar.AddRule("history_secret", []string{
			"hidden chambers and passages",
			"magical wards still active today",
			"symbols of an older civilization",
			"evidence of lost technologies",
		})
		grammar.AddRule("history_debate", []string{
			"its true original purpose",
			"the identity of its founders",
			"the source of its power",
			"the meaning of inscriptions",
		})
		grammar.AddRule("history_era", []string{
			"Golden Age",
			"Time of Troubles",
			"Age of Magic",
			"Dark Years",
		})
		grammar.AddRule("history_role", []string{
			"served as the capital",
			"was a center of magical learning",
			"housed the royal court",
			"protected the borders",
		})
		grammar.AddRule("history_figures", []string{
			"heroes and warriors",
			"mages and scholars",
			"kings and queens",
			"legendary craftsmen",
		})

	case "sci-fi":
		grammar.AddRule("history_paragraph", []string{
			fmt.Sprintf("Station %s was constructed in year %s. Initial purpose: %s Mission parameters have evolved significantly.", location, "#history_year#", "#history_mission#"),
			fmt.Sprintf("Historical records show %s experienced #history_incident# in %s. Response protocols: #history_response#", location, "#history_year#"),
			fmt.Sprintf("Technical analysis of %s reveals #history_tech# Modifications made over #history_timespan# improved functionality by #history_improvement#", location),
			fmt.Sprintf("During the #history_period# %s #history_significance# Personnel logs document #history_activity#", location),
		})
		grammar.AddRule("history_year", []string{
			"2156", "2187", "2203", "2245",
		})
		grammar.AddRule("history_mission", []string{
			"deep space research and observation",
			"colonial support and resupply",
			"military defense and surveillance",
			"scientific experimentation",
		})
		grammar.AddRule("history_incident", []string{
			"critical systems failure",
			"first contact with alien intelligence",
			"hull breach in sector 7",
			"AI malfunction event",
		})
		grammar.AddRule("history_response", []string{
			"Emergency evacuation successful, zero casualties.",
			"Diplomatic protocols initiated, contact ongoing.",
			"Repair crews deployed, systems restored within 72 hours.",
			"Manual override engaged, AI systems purged.",
		})
		grammar.AddRule("history_tech", []string{
			"advanced propulsion systems of unknown origin",
			"quantum computing cores predating current technology",
			"modular construction allowing expansion",
			"self-repairing hull materials",
		})
		grammar.AddRule("history_timespan", []string{
			"50 years",
			"three decades",
			"multiple generations",
			"the past century",
		})
		grammar.AddRule("history_improvement", []string{
			"200%",
			"450%",
			"a factor of 10",
			"several orders of magnitude",
		})
		grammar.AddRule("history_period", []string{
			"Colonial Expansion Era",
			"AI Integration Period",
			"Post-Contact Years",
			"Corporate Wars",
		})
		grammar.AddRule("history_significance", []string{
			"served as a strategic waypoint",
			"housed critical research programs",
			"provided emergency haven",
			"acted as a diplomatic neutral zone",
		})
		grammar.AddRule("history_activity", []string{
			"extensive scientific breakthroughs",
			"first contact negotiations",
			"classified military operations",
			"civilian life over generations",
		})

	default:
		grammar.AddRule("history_paragraph", []string{
			fmt.Sprintf("%s has a rich history. Built long ago, it has served many purposes over the years.", location),
			fmt.Sprintf("Historical records indicate %s was important for various reasons. Many significant events occurred here.", location),
			fmt.Sprintf("The story of %s is one of change and adaptation. Each era left its mark on this place.", location),
		})
	}
}
