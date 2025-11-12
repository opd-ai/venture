package book

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
