package ux

import (
	"fmt"
	"time"
)

// AllJourneys returns all defined user journeys.
func AllJourneys() []JourneyDefinition {
	return []JourneyDefinition{
		newPlayerJourney(),
		crafterJourney(),
		socialJourney(),
		explorerJourney(),
		traderJourney(),
		builderJourney(),
		raiderJourney(),
		pvperJourney(),
		questerJourney(),
		companionOwnerJourney(),
		vehicleUserJourney(),
		storytellerJourney(),
		prestigePlayerJourney(),
		guildLeaderJourney(),
		modderJourney(),
		crossServerTravelerJourney(),
		legendaryQuesterJourney(),
		housingDecoratorJourney(),
		siegeParticipantJourney(),
		economyTycoonJourney(),
	}
}

func newPlayerJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyNewPlayer,
		Name:             "New Player Onboarding",
		Description:      "Create character → tutorial → first quest → level 3",
		ExpectedDuration: 30 * time.Minute,
		RequiredFeatures: []string{"character_creation", "tutorial", "quest_system", "progression"},
		Steps: []JourneyStep{
			{Name: "Create Character", Description: "Choose class and customize appearance", Action: createCharacter},
			{Name: "Complete Tutorial", Description: "Learn basic controls and mechanics", Action: completeTutorial},
			{Name: "Accept First Quest", Description: "Talk to starter NPC and accept quest", Action: acceptFirstQuest},
			{Name: "Complete Quest", Description: "Defeat enemies or gather items", Action: completeQuest},
			{Name: "Turn In Quest", Description: "Return to NPC and claim reward", Action: turnInQuest},
			{Name: "Reach Level 3", Description: "Gain enough XP to level up twice", Action: reachLevel3},
		},
	}
}

func crafterJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyCrafter,
		Name:             "Crafting Workflow",
		Description:      "Gather materials → find recipe → craft item → equip",
		ExpectedDuration: 15 * time.Minute,
		RequiredFeatures: []string{"gathering", "recipes", "crafting", "inventory"},
		Steps: []JourneyStep{
			{Name: "Gather Materials", Description: "Collect wood, ore, or herbs", Action: gatherMaterials},
			{Name: "Find Recipe", Description: "Discover or purchase crafting recipe", Action: findRecipe},
			{Name: "Access Crafting Station", Description: "Locate and interact with station", Action: accessCraftingStation},
			{Name: "Craft Item", Description: "Use materials to create item", Action: craftItem},
			{Name: "Equip Item", Description: "Add crafted item to equipment slots", Action: equipItem},
		},
	}
}

func socialJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneySocial,
		Name:             "Social Interaction",
		Description:      "Join guild → participate in guild event → earn reward",
		ExpectedDuration: 20 * time.Minute,
		RequiredFeatures: []string{"guilds", "guild_events", "chat"},
		Steps: []JourneyStep{
			{Name: "Find Guild", Description: "Browse or search for guild", Action: findGuild},
			{Name: "Join Guild", Description: "Submit application or accept invite", Action: joinGuild},
			{Name: "Participate in Event", Description: "Join guild dungeon run or activity", Action: participateInEvent},
			{Name: "Earn Reward", Description: "Receive guild contribution points or items", Action: earnReward},
		},
	}
}

func explorerJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyExplorer,
		Name:             "Dungeon Exploration",
		Description:      "Discover dungeon → complete dungeon → collect loot",
		ExpectedDuration: 30 * time.Minute,
		RequiredFeatures: []string{"terrain_generation", "combat", "loot"},
		Steps: []JourneyStep{
			{Name: "Discover Dungeon", Description: "Find dungeon entrance on map", Action: discoverDungeon},
			{Name: "Enter Dungeon", Description: "Interact with entrance portal", Action: enterDungeon},
			{Name: "Clear Rooms", Description: "Defeat enemies in dungeon rooms", Action: clearRooms},
			{Name: "Defeat Boss", Description: "Beat the dungeon boss encounter", Action: defeatBoss},
			{Name: "Collect Loot", Description: "Open treasure chests and claim rewards", Action: collectLoot},
		},
	}
}

func traderJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyTrader,
		Name:             "Marketplace Trading",
		Description:      "List item → sell on marketplace → receive gold",
		ExpectedDuration: 10 * time.Minute,
		RequiredFeatures: []string{"marketplace", "trading"},
		Steps: []JourneyStep{
			{Name: "Access Marketplace", Description: "Open marketplace interface", Action: accessMarketplace},
			{Name: "List Item", Description: "Set price and list item for sale", Action: listItem},
			{Name: "Wait for Purchase", Description: "Simulated buyer purchases item", Action: waitForPurchase},
			{Name: "Receive Gold", Description: "Gold added to inventory", Action: receiveGold},
		},
	}
}

func builderJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyBuilder,
		Name:             "Housing & Building",
		Description:      "Purchase house → place furniture → invite friends",
		ExpectedDuration: 25 * time.Minute,
		RequiredFeatures: []string{"housing", "furniture", "permissions"},
		Steps: []JourneyStep{
			{Name: "Purchase House", Description: "Buy housing claim from NPC", Action: purchaseHouse},
			{Name: "Enter House", Description: "Access personal housing instance", Action: enterHouse},
			{Name: "Place Furniture", Description: "Add furniture items to house", Action: placeFurniture},
			{Name: "Set Permissions", Description: "Allow friends to visit", Action: setPermissions},
			{Name: "Invite Friends", Description: "Send invite to other players", Action: inviteFriends},
		},
	}
}

func raiderJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyRaider,
		Name:             "Raid Group Play",
		Description:      "Join raid group → defeat boss → distribute loot",
		ExpectedDuration: 60 * time.Minute,
		RequiredFeatures: []string{"group_system", "raid_bosses", "loot_distribution"},
		Steps: []JourneyStep{
			{Name: "Find Raid Group", Description: "Join or create raid party", Action: findRaidGroup},
			{Name: "Enter Raid", Description: "Teleport to raid instance", Action: enterRaid},
			{Name: "Complete Encounters", Description: "Defeat multiple bosses", Action: completeEncounters},
			{Name: "Distribute Loot", Description: "Roll for or bid on items", Action: distributeLoot},
		},
	}
}

func pvperJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyPvPer,
		Name:             "PvP Combat",
		Description:      "Challenge player → duel → earn reputation",
		ExpectedDuration: 10 * time.Minute,
		RequiredFeatures: []string{"pvp", "dueling", "reputation"},
		Steps: []JourneyStep{
			{Name: "Challenge Player", Description: "Send duel request to another player", Action: challengePlayer},
			{Name: "Fight Duel", Description: "Engage in 1v1 combat", Action: fightDuel},
			{Name: "Win or Lose", Description: "Duel concludes with victory or defeat", Action: winOrLose},
			{Name: "Earn Reputation", Description: "Gain PvP ranking points", Action: earnReputation},
		},
	}
}

func questerJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyQuester,
		Name:             "Quest Completion",
		Description:      "Accept quest → complete objectives → turn in",
		ExpectedDuration: 20 * time.Minute,
		RequiredFeatures: []string{"quest_system", "objectives"},
		Steps: []JourneyStep{
			{Name: "Accept Quest", Description: "Receive quest from NPC", Action: acceptQuest},
			{Name: "Complete Objectives", Description: "Kill enemies or collect items", Action: completeObjectives},
			{Name: "Return to NPC", Description: "Travel back to quest giver", Action: returnToNPC},
			{Name: "Turn In Quest", Description: "Claim quest rewards", Action: turnInQuest},
		},
	}
}

func companionOwnerJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyCompanionOwner,
		Name:             "Companion Management",
		Description:      "Tame companion → train skills → use in combat",
		ExpectedDuration: 30 * time.Minute,
		RequiredFeatures: []string{"companions", "companion_skills"},
		Steps: []JourneyStep{
			{Name: "Tame Companion", Description: "Use taming ability on creature", Action: tameCompanion},
			{Name: "Train Skills", Description: "Teach companion new abilities", Action: trainSkills},
			{Name: "Equip Companion", Description: "Set companion as active follower", Action: equipCompanion},
			{Name: "Use in Combat", Description: "Companion assists in battle", Action: useInCombat},
		},
	}
}

func vehicleUserJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyVehicleUser,
		Name:             "Vehicle Usage",
		Description:      "Acquire mount → upgrade → use in travel",
		ExpectedDuration: 15 * time.Minute,
		RequiredFeatures: []string{"vehicles", "vehicle_upgrades"},
		Steps: []JourneyStep{
			{Name: "Acquire Vehicle", Description: "Purchase or find mount", Action: acquireVehicle},
			{Name: "Upgrade Vehicle", Description: "Enhance speed or storage", Action: upgradeVehicle},
			{Name: "Mount Vehicle", Description: "Enter mounted state", Action: mountVehicle},
			{Name: "Fast Travel", Description: "Use vehicle for faster movement", Action: fastTravel},
		},
	}
}

func storytellerJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyStoryteller,
		Name:             "Story Discovery",
		Description:      "Discover lore → complete story arc → unlock epilogue",
		ExpectedDuration: 40 * time.Minute,
		RequiredFeatures: []string{"narrative", "lore_books"},
		Steps: []JourneyStep{
			{Name: "Discover Lore", Description: "Find lore books or fragments", Action: discoverLore},
			{Name: "Complete Story Arc", Description: "Finish all chapters", Action: completeStoryArc},
			{Name: "Unlock Epilogue", Description: "Reveal final story content", Action: unlockEpilogue},
		},
	}
}

func prestigePlayerJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyPrestigePlayer,
		Name:             "Prestige Progression",
		Description:      "Reach max level → unlock prestige → earn paragon points",
		ExpectedDuration: 2 * time.Hour,
		RequiredFeatures: []string{"progression", "prestige_system"},
		Steps: []JourneyStep{
			{Name: "Reach Max Level", Description: "Hit level cap", Action: reachMaxLevel},
			{Name: "Unlock Prestige", Description: "Activate prestige mode", Action: unlockPrestige},
			{Name: "Earn Paragon Points", Description: "Gain prestige currency", Action: earnParagonPoints},
		},
	}
}

func guildLeaderJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyGuildLeader,
		Name:             "Guild Leadership",
		Description:      "Create guild → recruit members → declare war",
		ExpectedDuration: 45 * time.Minute,
		RequiredFeatures: []string{"guilds", "guild_warfare"},
		Steps: []JourneyStep{
			{Name: "Create Guild", Description: "Found new guild with charter", Action: createGuild},
			{Name: "Recruit Members", Description: "Invite players to join", Action: recruitMembers},
			{Name: "Declare War", Description: "Start conflict with rival guild", Action: declareWar},
		},
	}
}

func modderJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyModder,
		Name:             "Mod Installation",
		Description:      "Install mod → configure → observe effects",
		ExpectedDuration: 10 * time.Minute,
		RequiredFeatures: []string{"mod_system"},
		Steps: []JourneyStep{
			{Name: "Install Mod", Description: "Load mod package", Action: installMod},
			{Name: "Configure Mod", Description: "Adjust mod settings", Action: configureMod},
			{Name: "Observe Effects", Description: "Verify mod functionality", Action: observeEffects},
		},
	}
}

func crossServerTravelerJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyCrossServerTraveler,
		Name:             "Cross-Server Travel",
		Description:      "Enter portal → transfer → explore new server",
		ExpectedDuration: 5 * time.Minute,
		RequiredFeatures: []string{"portals", "federation"},
		Steps: []JourneyStep{
			{Name: "Find Portal", Description: "Locate cross-server portal", Action: findPortal},
			{Name: "Enter Portal", Description: "Initiate server transfer", Action: enterPortal},
			{Name: "Transfer Complete", Description: "Arrive on new server", Action: transferComplete},
			{Name: "Explore New Server", Description: "Discover new content", Action: exploreNewServer},
		},
	}
}

func legendaryQuesterJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyLegendaryQuester,
		Name:             "Legendary Quest",
		Description:      "Start legendary quest → complete all steps → claim reward",
		ExpectedDuration: 10 * time.Hour,
		RequiredFeatures: []string{"legendary_quests"},
		Steps: []JourneyStep{
			{Name: "Start Quest", Description: "Accept legendary quest chain", Action: startLegendaryQuest},
			{Name: "Complete Steps", Description: "Finish all quest stages", Action: completeLegendarySteps},
			{Name: "Claim Reward", Description: "Receive legendary item", Action: claimLegendaryReward},
		},
	}
}

func housingDecoratorJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyHousingDecorator,
		Name:             "Housing Decoration",
		Description:      "Buy furniture → place decorations → showcase to friends",
		ExpectedDuration: 20 * time.Minute,
		RequiredFeatures: []string{"housing", "furniture"},
		Steps: []JourneyStep{
			{Name: "Buy Furniture", Description: "Purchase decoration items", Action: buyFurniture},
			{Name: "Place Decorations", Description: "Arrange furniture in house", Action: placeDecorations},
			{Name: "Showcase to Friends", Description: "Invite friends to view", Action: showcaseToFriends},
		},
	}
}

func siegeParticipantJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneySiegeParticipant,
		Name:             "Territory Siege",
		Description:      "Join siege → attack/defend → claim territory",
		ExpectedDuration: 3 * time.Hour,
		RequiredFeatures: []string{"sieges", "territory"},
		Steps: []JourneyStep{
			{Name: "Join Siege", Description: "Enlist in siege battle", Action: joinSiege},
			{Name: "Attack or Defend", Description: "Participate in battle", Action: attackOrDefend},
			{Name: "Claim Territory", Description: "Victory results in territory control", Action: claimTerritory},
		},
	}
}

func economyTycoonJourney() JourneyDefinition {
	return JourneyDefinition{
		Type:             JourneyEconomyTycoon,
		Name:             "Economy Trading",
		Description:      "Buy low on Server A → sell high on Server B → profit",
		ExpectedDuration: 30 * time.Minute,
		RequiredFeatures: []string{"marketplace", "federation"},
		Steps: []JourneyStep{
			{Name: "Buy Low", Description: "Purchase items at low price", Action: buyLow},
			{Name: "Transfer Servers", Description: "Move to server with higher prices", Action: transferServers},
			{Name: "Sell High", Description: "List items at premium price", Action: sellHigh},
			{Name: "Profit", Description: "Calculate net gain", Action: calculateProfit},
		},
	}
}

// Step action implementations (simulation-based)

func createCharacter(ctx *JourneyContext) error {
	ctx.Data["character_created"] = true
	return nil
}

func completeTutorial(ctx *JourneyContext) error {
	created, ok := ctx.Data["character_created"]
	if !ok || !created.(bool) {
		return fmt.Errorf("character must be created first")
	}
	ctx.Data["tutorial_complete"] = true
	return nil
}

func acceptFirstQuest(ctx *JourneyContext) error {
	ctx.Data["quest_accepted"] = true
	return nil
}

func completeQuest(ctx *JourneyContext) error {
	accepted, ok := ctx.Data["quest_accepted"]
	if !ok || !accepted.(bool) {
		return fmt.Errorf("quest must be accepted first")
	}
	ctx.Data["quest_complete"] = true
	return nil
}

func turnInQuest(ctx *JourneyContext) error {
	complete, ok := ctx.Data["quest_complete"]
	if !ok || !complete.(bool) {
		return fmt.Errorf("quest must be completed first")
	}
	ctx.Data["quest_turned_in"] = true
	return nil
}

func reachLevel3(ctx *JourneyContext) error {
	ctx.Data["level"] = 3
	return nil
}

func gatherMaterials(ctx *JourneyContext) error {
	ctx.Data["materials"] = 10
	return nil
}

func findRecipe(ctx *JourneyContext) error {
	ctx.Data["recipe_found"] = true
	return nil
}

func accessCraftingStation(ctx *JourneyContext) error {
	ctx.Data["at_station"] = true
	return nil
}

func craftItem(ctx *JourneyContext) error {
	materials, ok := ctx.Data["materials"]
	if !ok || materials.(int) < 5 {
		return fmt.Errorf("insufficient materials")
	}
	ctx.Data["item_crafted"] = true
	return nil
}

func equipItem(ctx *JourneyContext) error {
	crafted, ok := ctx.Data["item_crafted"]
	if !ok || !crafted.(bool) {
		return fmt.Errorf("no item to equip")
	}
	ctx.Data["item_equipped"] = true
	return nil
}

func findGuild(ctx *JourneyContext) error {
	ctx.Data["guild_found"] = true
	return nil
}

func joinGuild(ctx *JourneyContext) error {
	ctx.Data["in_guild"] = true
	return nil
}

func participateInEvent(ctx *JourneyContext) error {
	ctx.Data["event_participated"] = true
	return nil
}

func earnReward(ctx *JourneyContext) error {
	ctx.Data["reward_earned"] = true
	return nil
}

func discoverDungeon(ctx *JourneyContext) error {
	ctx.Data["dungeon_discovered"] = true
	return nil
}

func enterDungeon(ctx *JourneyContext) error {
	ctx.Data["in_dungeon"] = true
	return nil
}

func clearRooms(ctx *JourneyContext) error {
	ctx.Data["rooms_cleared"] = 3
	return nil
}

func defeatBoss(ctx *JourneyContext) error {
	ctx.Data["boss_defeated"] = true
	return nil
}

func collectLoot(ctx *JourneyContext) error {
	ctx.Data["loot_collected"] = true
	return nil
}

func accessMarketplace(ctx *JourneyContext) error {
	ctx.Data["marketplace_open"] = true
	return nil
}

func listItem(ctx *JourneyContext) error {
	ctx.Data["item_listed"] = true
	return nil
}

func waitForPurchase(ctx *JourneyContext) error {
	ctx.Data["item_sold"] = true
	return nil
}

func receiveGold(ctx *JourneyContext) error {
	ctx.Data["gold"] = 100
	return nil
}

func purchaseHouse(ctx *JourneyContext) error {
	ctx.Data["house_owned"] = true
	return nil
}

func enterHouse(ctx *JourneyContext) error {
	ctx.Data["in_house"] = true
	return nil
}

func placeFurniture(ctx *JourneyContext) error {
	ctx.Data["furniture_placed"] = 5
	return nil
}

func setPermissions(ctx *JourneyContext) error {
	ctx.Data["permissions_set"] = true
	return nil
}

func inviteFriends(ctx *JourneyContext) error {
	ctx.Data["friends_invited"] = true
	return nil
}

func findRaidGroup(ctx *JourneyContext) error {
	ctx.Data["in_raid_group"] = true
	return nil
}

func enterRaid(ctx *JourneyContext) error {
	ctx.Data["in_raid"] = true
	return nil
}

func completeEncounters(ctx *JourneyContext) error {
	ctx.Data["encounters_complete"] = true
	return nil
}

func distributeLoot(ctx *JourneyContext) error {
	ctx.Data["loot_distributed"] = true
	return nil
}

func challengePlayer(ctx *JourneyContext) error {
	ctx.Data["duel_challenged"] = true
	return nil
}

func fightDuel(ctx *JourneyContext) error {
	ctx.Data["duel_fought"] = true
	return nil
}

func winOrLose(ctx *JourneyContext) error {
	ctx.Data["duel_result"] = "win"
	return nil
}

func earnReputation(ctx *JourneyContext) error {
	ctx.Data["reputation"] = 10
	return nil
}

func acceptQuest(ctx *JourneyContext) error {
	ctx.Data["quest_accepted"] = true
	return nil
}

func completeObjectives(ctx *JourneyContext) error {
	ctx.Data["objectives_complete"] = true
	ctx.Data["quest_complete"] = true // Set for turnInQuest dependency
	return nil
}

func returnToNPC(ctx *JourneyContext) error {
	ctx.Data["at_npc"] = true
	return nil
}

func tameCompanion(ctx *JourneyContext) error {
	ctx.Data["companion_tamed"] = true
	return nil
}

func trainSkills(ctx *JourneyContext) error {
	ctx.Data["skills_trained"] = 3
	return nil
}

func equipCompanion(ctx *JourneyContext) error {
	ctx.Data["companion_equipped"] = true
	return nil
}

func useInCombat(ctx *JourneyContext) error {
	ctx.Data["companion_used"] = true
	return nil
}

func acquireVehicle(ctx *JourneyContext) error {
	ctx.Data["vehicle_owned"] = true
	return nil
}

func upgradeVehicle(ctx *JourneyContext) error {
	ctx.Data["vehicle_upgraded"] = true
	return nil
}

func mountVehicle(ctx *JourneyContext) error {
	ctx.Data["vehicle_mounted"] = true
	return nil
}

func fastTravel(ctx *JourneyContext) error {
	ctx.Data["distance_traveled"] = 1000
	return nil
}

func discoverLore(ctx *JourneyContext) error {
	ctx.Data["lore_discovered"] = 5
	return nil
}

func completeStoryArc(ctx *JourneyContext) error {
	ctx.Data["story_complete"] = true
	return nil
}

func unlockEpilogue(ctx *JourneyContext) error {
	ctx.Data["epilogue_unlocked"] = true
	return nil
}

func reachMaxLevel(ctx *JourneyContext) error {
	ctx.Data["level"] = 50
	return nil
}

func unlockPrestige(ctx *JourneyContext) error {
	ctx.Data["prestige_unlocked"] = true
	return nil
}

func earnParagonPoints(ctx *JourneyContext) error {
	ctx.Data["paragon_points"] = 10
	return nil
}

func createGuild(ctx *JourneyContext) error {
	ctx.Data["guild_created"] = true
	return nil
}

func recruitMembers(ctx *JourneyContext) error {
	ctx.Data["members_recruited"] = 5
	return nil
}

func declareWar(ctx *JourneyContext) error {
	ctx.Data["war_declared"] = true
	return nil
}

func installMod(ctx *JourneyContext) error {
	ctx.Data["mod_installed"] = true
	return nil
}

func configureMod(ctx *JourneyContext) error {
	ctx.Data["mod_configured"] = true
	return nil
}

func observeEffects(ctx *JourneyContext) error {
	ctx.Data["effects_observed"] = true
	return nil
}

func findPortal(ctx *JourneyContext) error {
	ctx.Data["portal_found"] = true
	return nil
}

func enterPortal(ctx *JourneyContext) error {
	ctx.Data["in_portal"] = true
	return nil
}

func transferComplete(ctx *JourneyContext) error {
	ctx.Data["server_transferred"] = true
	return nil
}

func exploreNewServer(ctx *JourneyContext) error {
	ctx.Data["explored"] = true
	return nil
}

func startLegendaryQuest(ctx *JourneyContext) error {
	ctx.Data["legendary_started"] = true
	return nil
}

func completeLegendarySteps(ctx *JourneyContext) error {
	ctx.Data["legendary_steps"] = 10
	return nil
}

func claimLegendaryReward(ctx *JourneyContext) error {
	ctx.Data["legendary_reward"] = true
	return nil
}

func buyFurniture(ctx *JourneyContext) error {
	ctx.Data["furniture_bought"] = 10
	return nil
}

func placeDecorations(ctx *JourneyContext) error {
	ctx.Data["decorations_placed"] = 10
	return nil
}

func showcaseToFriends(ctx *JourneyContext) error {
	ctx.Data["showcased"] = true
	return nil
}

func joinSiege(ctx *JourneyContext) error {
	ctx.Data["in_siege"] = true
	return nil
}

func attackOrDefend(ctx *JourneyContext) error {
	ctx.Data["siege_participated"] = true
	return nil
}

func claimTerritory(ctx *JourneyContext) error {
	ctx.Data["territory_claimed"] = true
	return nil
}

func buyLow(ctx *JourneyContext) error {
	ctx.Data["items_bought"] = 10
	ctx.Data["purchase_price"] = 50
	return nil
}

func transferServers(ctx *JourneyContext) error {
	ctx.Data["server_switched"] = true
	return nil
}

func sellHigh(ctx *JourneyContext) error {
	ctx.Data["items_sold"] = 10
	ctx.Data["sale_price"] = 100
	return nil
}

func calculateProfit(ctx *JourneyContext) error {
	profit := ctx.Data["sale_price"].(int) - ctx.Data["purchase_price"].(int)
	ctx.Data["profit"] = profit
	return nil
}
