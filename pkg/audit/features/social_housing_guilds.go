package features

import "time"

// RegisterSocialFeatures adds all social system features to the registry
func RegisterSocialFeatures(r *FeatureRegistry) {
	// Chat (V5.0)
	r.Register(&Feature{
		ID:                   "chat.send",
		Name:                 "Send Chat Message",
		Category:             CategorySocial,
		Description:          "Type and send chat messages",
		Accessible:           true,
		AccessibilityTime:    2 * time.Minute,
		AccessibilityPath:    "Press Enter, type message, press Enter again",
		HasTutorial:          true,
		TutorialLocation:     "First multiplayer session",
		TutorialCompleteness: 1.0,
		IntegratedSystems:    []string{"input", "network", "ui"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "chat.receive",
		Name:                 "Receive Chat Messages",
		Category:             CategorySocial,
		Description:          "Receive and view chat messages",
		Accessible:           true,
		AccessibilityTime:    3 * time.Minute,
		AccessibilityPath:    "Join multiplayer, observe chat window",
		HasTutorial:          true,
		TutorialLocation:     "Chat window tooltip",
		TutorialCompleteness: 1.0,
		IntegratedSystems:    []string{"network", "ui", "notifications"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "chat.channels",
		Name:                 "Chat Channels",
		Category:             CategorySocial,
		Description:          "Switch between global, local, party, whisper",
		Accessible:           true,
		AccessibilityTime:    5 * time.Minute,
		AccessibilityPath:    "Type /global, /local, /party, or /whisper",
		HasTutorial:          true,
		TutorialLocation:     "Chat commands help",
		TutorialCompleteness: 0.9,
		IntegratedSystems:    []string{"commands", "routing", "ui"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	// Expressions (V4.0)
	r.Register(&Feature{
		ID:                   "expressions.use",
		Name:                 "Use Expressions",
		Category:             CategorySocial,
		Description:          "Perform emotes (wave, dance, etc.)",
		Accessible:           true,
		AccessibilityTime:    5 * time.Minute,
		AccessibilityPath:    "Press Shift+1 through Shift+=",
		HasTutorial:          true,
		TutorialLocation:     "Expression hotkey tutorial",
		TutorialCompleteness: 0.8,
		IntegratedSystems:    []string{"animation", "input", "visual"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "expressions.combos",
		Name:                 "Expression Combos",
		Category:             CategorySocial,
		Description:          "Synchronized group expressions",
		Accessible:           true,
		AccessibilityTime:    10 * time.Minute,
		AccessibilityPath:    "Use same expression with nearby player",
		HasTutorial:          true,
		TutorialLocation:     "Expression combo notification",
		TutorialCompleteness: 0.7,
		IntegratedSystems:    []string{"sync", "detection", "achievements"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	// Mini-Games (V4.0)
	r.Register(&Feature{
		ID:                   "minigames.play",
		Name:                 "Play Mini-Games",
		Category:             CategoryAdvanced,
		Description:          "Play tavern games (cards, dice, puzzles)",
		Accessible:           true,
		AccessibilityTime:    25 * time.Minute,
		AccessibilityPath:    "Find tavern, interact with game station",
		HasTutorial:          true,
		TutorialLocation:     "Mini-game UI tutorial",
		TutorialCompleteness: 0.8,
		IntegratedSystems:    []string{"interaction", "rewards", "ui"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	// Reputation (V4.0)
	r.Register(&Feature{
		ID:                   "reputation.gain",
		Name:                 "Gain Reputation",
		Category:             CategorySocial,
		Description:          "Build reputation with factions",
		Accessible:           true,
		AccessibilityTime:    30 * time.Minute,
		AccessibilityPath:    "Complete faction quests",
		HasTutorial:          true,
		TutorialLocation:     "Reputation UI tooltip",
		TutorialCompleteness: 0.8,
		IntegratedSystems:    []string{"factions", "quests", "rewards"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "reputation.effects",
		Name:                 "Reputation Effects",
		Category:             CategorySocial,
		Description:          "Benefits/penalties from reputation",
		Accessible:           true,
		AccessibilityTime:    40 * time.Minute,
		AccessibilityPath:    "Observe NPC prices, quest availability",
		HasTutorial:          true,
		TutorialLocation:     "Merchant dialog, quest giver dialog",
		TutorialCompleteness: 0.7,
		IntegratedSystems:    []string{"pricing", "access", "dialog"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})
}

// RegisterHousingFeatures adds all housing system features to the registry
func RegisterHousingFeatures(r *FeatureRegistry) {
	// Housing (V8.0)
	r.Register(&Feature{
		ID:                   "housing.claim",
		Name:                 "Claim Housing Plot",
		Category:             CategoryHousing,
		Description:          "Purchase and claim housing plot",
		Accessible:           true,
		AccessibilityTime:    60 * time.Minute,
		AccessibilityPath:    "Earn gold, find housing area, purchase plot",
		HasTutorial:          true,
		TutorialLocation:     "Housing system tutorial",
		TutorialCompleteness: 0.9,
		IntegratedSystems:    []string{"economy", "world", "permissions"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "housing.build",
		Name:                 "Build Structure",
		Category:             CategoryHousing,
		Description:          "Construct buildings on plot",
		Accessible:           true,
		AccessibilityTime:    70 * time.Minute,
		AccessibilityPath:    "Claim plot, open build menu, select blueprint",
		HasTutorial:          true,
		TutorialLocation:     "Building UI tutorial",
		TutorialCompleteness: 0.9,
		IntegratedSystems:    []string{"construction", "materials", "blueprints"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "housing.furniture",
		Name:                 "Place Furniture",
		Category:             CategoryHousing,
		Description:          "Furnish interior with items",
		Accessible:           true,
		AccessibilityTime:    80 * time.Minute,
		AccessibilityPath:    "Enter building, open furniture menu",
		HasTutorial:          true,
		TutorialLocation:     "Furniture placement UI",
		TutorialCompleteness: 0.8,
		IntegratedSystems:    []string{"placement", "storage", "decoration"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "housing.storage",
		Name:                 "Housing Storage",
		Category:             CategoryHousing,
		Description:          "Store items in house chests",
		Accessible:           true,
		AccessibilityTime:    85 * time.Minute,
		AccessibilityPath:    "Place chest, interact to store items",
		HasTutorial:          true,
		TutorialLocation:     "Storage chest UI",
		TutorialCompleteness: 0.9,
		IntegratedSystems:    []string{"inventory", "persistence", "access"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "housing.permissions",
		Name:                 "Housing Permissions",
		Category:             CategoryHousing,
		Description:          "Grant access to friends/guildmates",
		Accessible:           true,
		AccessibilityTime:    90 * time.Minute,
		AccessibilityPath:    "Open permissions UI, add players",
		HasTutorial:          true,
		TutorialLocation:     "Permissions UI help text",
		TutorialCompleteness: 0.7,
		IntegratedSystems:    []string{"access-control", "social", "ui"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})
}

// RegisterGuildFeatures adds all guild system features to the registry
func RegisterGuildFeatures(r *FeatureRegistry) {
	// Guilds (V8.0)
	r.Register(&Feature{
		ID:                   "guilds.create",
		Name:                 "Create Guild",
		Category:             CategoryGuilds,
		Description:          "Found a new guild",
		Accessible:           true,
		AccessibilityTime:    120 * time.Minute,
		AccessibilityPath:    "Reach level 20, earn gold, visit guild hall",
		HasTutorial:          true,
		TutorialLocation:     "Guild creation UI",
		TutorialCompleteness: 0.9,
		IntegratedSystems:    []string{"economy", "social", "management"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "guilds.join",
		Name:                 "Join Guild",
		Category:             CategoryGuilds,
		Description:          "Join existing guild",
		Accessible:           true,
		AccessibilityTime:    30 * time.Minute,
		AccessibilityPath:    "Receive invite, accept in UI",
		HasTutorial:          true,
		TutorialLocation:     "Guild invite notification",
		TutorialCompleteness: 1.0,
		IntegratedSystems:    []string{"invites", "roster", "chat"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "guilds.resources",
		Name:                 "Guild Resources",
		Category:             CategoryGuilds,
		Description:          "Contribute and access guild resources",
		Accessible:           true,
		AccessibilityTime:    140 * time.Minute,
		AccessibilityPath:    "Join guild, donate items/gold",
		HasTutorial:          true,
		TutorialLocation:     "Guild treasury UI",
		TutorialCompleteness: 0.8,
		IntegratedSystems:    []string{"economy", "storage", "permissions"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "guilds.territory",
		Name:                 "Guild Territory",
		Category:             CategoryGuilds,
		Description:          "Claim and control territory",
		Accessible:           true,
		AccessibilityTime:    180 * time.Minute,
		AccessibilityPath:    "Guild leader initiates claim",
		HasTutorial:          true,
		TutorialLocation:     "Territory UI tutorial",
		TutorialCompleteness: 0.8,
		IntegratedSystems:    []string{"conquest", "warfare", "defense"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "guilds.warfare",
		Name:                 "Guild Warfare",
		Category:             CategoryGuilds,
		Description:          "Participate in guild vs guild combat",
		Accessible:           true,
		AccessibilityTime:    200 * time.Minute,
		AccessibilityPath:    "Guild declares war, join siege",
		HasTutorial:          true,
		TutorialLocation:     "Siege UI tutorial",
		TutorialCompleteness: 0.7,
		IntegratedSystems:    []string{"combat", "objectives", "rewards"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})
}
