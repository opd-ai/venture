// Package entity provides entity generation templates for all genres.
// This file defines EntityTemplate struct and genre-specific template functions.
// Originally from: types.go
package entity

// EntityTemplate defines a template for generating entities.
type EntityTemplate struct {
	BaseType     EntityType
	BaseSize     EntitySize
	NamePrefixes []string
	NameSuffixes []string
	Tags         []string
	HealthRange  [2]int // min, max
	DamageRange  [2]int
	DefenseRange [2]int
	SpeedRange   [2]float64
}

// GetFantasyTemplates returns entity templates for fantasy genre.
func GetFantasyTemplates() []EntityTemplate {
	return []EntityTemplate{
		{
			BaseType:     TypeMinion,
			BaseSize:     SizeSmall,
			NamePrefixes: []string{"Goblin", "Kobold", "Imp", "Sprite"},
			NameSuffixes: []string{"Scout", "Warrior", "Shaman", "Raider"},
			Tags:         []string{"weak", "fast", "group"},
			HealthRange:  [2]int{10, 30},
			DamageRange:  [2]int{2, 8},
			DefenseRange: [2]int{0, 3},
			SpeedRange:   [2]float64{1.2, 1.5},
		},
		{
			BaseType:     TypeMonster,
			BaseSize:     SizeMedium,
			NamePrefixes: []string{"Orc", "Skeleton", "Zombie", "Ghoul"},
			NameSuffixes: []string{"Warrior", "Brute", "Hunter", "Berserker"},
			Tags:         []string{"medium", "balanced"},
			HealthRange:  [2]int{40, 80},
			DamageRange:  [2]int{8, 15},
			DefenseRange: [2]int{3, 8},
			SpeedRange:   [2]float64{0.8, 1.0},
		},
		{
			BaseType:     TypeMonster,
			BaseSize:     SizeLarge,
			NamePrefixes: []string{"Ogre", "Troll", "Minotaur", "Golem"},
			NameSuffixes: []string{"Crusher", "Smasher", "Guardian", "Destroyer"},
			Tags:         []string{"tough", "slow", "powerful"},
			HealthRange:  [2]int{100, 200},
			DamageRange:  [2]int{15, 30},
			DefenseRange: [2]int{8, 15},
			SpeedRange:   [2]float64{0.5, 0.7},
		},
		{
			BaseType:     TypeBoss,
			BaseSize:     SizeHuge,
			NamePrefixes: []string{"Ancient", "Elder", "Lord", "King"},
			NameSuffixes: []string{"Dragon", "Demon", "Lich", "Wyrm"},
			Tags:         []string{"boss", "elite", "legendary"},
			HealthRange:  [2]int{300, 500},
			DamageRange:  [2]int{30, 60},
			DefenseRange: [2]int{15, 30},
			SpeedRange:   [2]float64{0.6, 0.9},
		},
		{
			BaseType:     TypeNPC,
			BaseSize:     SizeMedium,
			NamePrefixes: []string{"Merchant", "Guard", "Priest", "Wizard"},
			NameSuffixes: []string{"Smith", "Elder", "Scholar", "Keeper"},
			Tags:         []string{"friendly", "trader", "quest"},
			HealthRange:  [2]int{50, 100},
			DamageRange:  [2]int{5, 10},
			DefenseRange: [2]int{5, 10},
			SpeedRange:   [2]float64{1.0, 1.0},
		},
	}
}

// GetSciFiTemplates returns entity templates for sci-fi genre.
func GetSciFiTemplates() []EntityTemplate {
	return []EntityTemplate{
		{
			BaseType:     TypeMinion,
			BaseSize:     SizeSmall,
			NamePrefixes: []string{"Scout", "Drone", "Bot", "Probe"},
			NameSuffixes: []string{"MK-I", "Alpha", "Beta", "Unit"},
			Tags:         []string{"robotic", "fast", "scout"},
			HealthRange:  [2]int{15, 35},
			DamageRange:  [2]int{3, 10},
			DefenseRange: [2]int{1, 4},
			SpeedRange:   [2]float64{1.3, 1.6},
		},
		{
			BaseType:     TypeMonster,
			BaseSize:     SizeMedium,
			NamePrefixes: []string{"Combat", "Security", "War", "Battle"},
			NameSuffixes: []string{"Android", "Cyborg", "Mech", "Trooper"},
			Tags:         []string{"armored", "tactical"},
			HealthRange:  [2]int{50, 90},
			DamageRange:  [2]int{10, 18},
			DefenseRange: [2]int{5, 10},
			SpeedRange:   [2]float64{0.9, 1.1},
		},
		{
			BaseType:     TypeBoss,
			BaseSize:     SizeHuge,
			NamePrefixes: []string{"Titan", "Colossus", "Omega", "Prime"},
			NameSuffixes: []string{"Mech", "Destroyer", "Sentinel", "Core"},
			Tags:         []string{"boss", "mechanical", "heavy"},
			HealthRange:  [2]int{350, 550},
			DamageRange:  [2]int{35, 65},
			DefenseRange: [2]int{20, 35},
			SpeedRange:   [2]float64{0.5, 0.8},
		},
		{
			BaseType:     TypeNPC,
			BaseSize:     SizeMedium,
			NamePrefixes: []string{"Station", "Orbital", "Tech", "Supply"},
			NameSuffixes: []string{"Trader", "Merchant", "Vendor", "Dealer"},
			Tags:         []string{"friendly", "trader", "merchant"},
			HealthRange:  [2]int{50, 100},
			DamageRange:  [2]int{5, 10},
			DefenseRange: [2]int{5, 10},
			SpeedRange:   [2]float64{1.0, 1.0},
		},
	}
}

// GetHorrorTemplates returns entity templates for horror genre.
// GAP-005 REPAIR: Added horror genre templates for variety.
func GetHorrorTemplates() []EntityTemplate {
	return []EntityTemplate{
		{
			BaseType:     TypeMinion,
			BaseSize:     SizeSmall,
			NamePrefixes: []string{"Creeping", "Twisted", "Cursed", "Vile"},
			NameSuffixes: []string{"Wraith", "Shadow", "Corpse", "Thing"},
			Tags:         []string{"undead", "horrifying", "fast"},
			HealthRange:  [2]int{20, 40},
			DamageRange:  [2]int{5, 12},
			DefenseRange: [2]int{1, 5},
			SpeedRange:   [2]float64{1.2, 1.5},
		},
		{
			BaseType:     TypeMonster,
			BaseSize:     SizeMedium,
			NamePrefixes: []string{"Rotten", "Shambling", "Ghastly", "Bloated"},
			NameSuffixes: []string{"Zombie", "Ghoul", "Revenant", "Abomination"},
			Tags:         []string{"undead", "resilient"},
			HealthRange:  [2]int{60, 100},
			DamageRange:  [2]int{8, 16},
			DefenseRange: [2]int{3, 8},
			SpeedRange:   [2]float64{0.7, 0.9},
		},
		{
			BaseType:     TypeBoss,
			BaseSize:     SizeLarge,
			NamePrefixes: []string{"Ancient", "Nightmare", "Eldritch", "Dread"},
			NameSuffixes: []string{"Horror", "Terror", "Lord", "Entity"},
			Tags:         []string{"boss", "horrifying", "powerful"},
			HealthRange:  [2]int{400, 600},
			DamageRange:  [2]int{40, 70},
			DefenseRange: [2]int{15, 30},
			SpeedRange:   [2]float64{0.6, 0.9},
		},
		{
			BaseType:     TypeNPC,
			BaseSize:     SizeMedium,
			NamePrefixes: []string{"Wandering", "Lost", "Cursed", "Mad"},
			NameSuffixes: []string{"Trader", "Survivor", "Merchant", "Keeper"},
			Tags:         []string{"friendly", "trader", "merchant"},
			HealthRange:  [2]int{50, 100},
			DamageRange:  [2]int{5, 10},
			DefenseRange: [2]int{5, 10},
			SpeedRange:   [2]float64{1.0, 1.0},
		},
	}
}

// GetCyberpunkTemplates returns entity templates for cyberpunk genre.
// GAP-005 REPAIR: Added cyberpunk genre templates for variety.
func GetCyberpunkTemplates() []EntityTemplate {
	return []EntityTemplate{
		{
			BaseType:     TypeMinion,
			BaseSize:     SizeSmall,
			NamePrefixes: []string{"Street", "Corpo", "Gang", "Hack"},
			NameSuffixes: []string{"Runner", "Goon", "Agent", "Merc"},
			Tags:         []string{"augmented", "fast", "human"},
			HealthRange:  [2]int{25, 45},
			DamageRange:  [2]int{6, 14},
			DefenseRange: [2]int{2, 6},
			SpeedRange:   [2]float64{1.1, 1.4},
		},
		{
			BaseType:     TypeMonster,
			BaseSize:     SizeMedium,
			NamePrefixes: []string{"Cyber", "Enhanced", "Corp", "Military"},
			NameSuffixes: []string{"Enforcer", "Assassin", "Operative", "Soldier"},
			Tags:         []string{"augmented", "tactical", "human"},
			HealthRange:  [2]int{55, 95},
			DamageRange:  [2]int{12, 20},
			DefenseRange: [2]int{6, 12},
			SpeedRange:   [2]float64{1.0, 1.2},
		},
		{
			BaseType:     TypeBoss,
			BaseSize:     SizeLarge,
			NamePrefixes: []string{"Corporate", "Syndicate", "Elite", "Mega"},
			NameSuffixes: []string{"Boss", "Executive", "Commander", "Director"},
			Tags:         []string{"boss", "augmented", "powerful"},
			HealthRange:  [2]int{380, 580},
			DamageRange:  [2]int{38, 68},
			DefenseRange: [2]int{18, 32},
			SpeedRange:   [2]float64{0.7, 1.0},
		},
		{
			BaseType:     TypeNPC,
			BaseSize:     SizeMedium,
			NamePrefixes: []string{"Black Market", "Data", "Tech", "Neural"},
			NameSuffixes: []string{"Broker", "Dealer", "Fence", "Trader"},
			Tags:         []string{"friendly", "trader", "merchant"},
			HealthRange:  [2]int{50, 100},
			DamageRange:  [2]int{5, 10},
			DefenseRange: [2]int{5, 10},
			SpeedRange:   [2]float64{1.0, 1.0},
		},
	}
}

// GetPostApocTemplates returns entity templates for post-apocalyptic genre.
// GAP-005 REPAIR: Added post-apocalyptic genre templates for variety.
func GetPostApocTemplates() []EntityTemplate {
	return []EntityTemplate{
		{
			BaseType:     TypeMinion,
			BaseSize:     SizeSmall,
			NamePrefixes: []string{"Feral", "Rabid", "Mutated", "Irradiated"},
			NameSuffixes: []string{"Scavenger", "Rat", "Dog", "Crawler"},
			Tags:         []string{"mutant", "fast", "wild"},
			HealthRange:  [2]int{18, 38},
			DamageRange:  [2]int{4, 11},
			DefenseRange: [2]int{1, 4},
			SpeedRange:   [2]float64{1.3, 1.6},
		},
		{
			BaseType:     TypeMonster,
			BaseSize:     SizeMedium,
			NamePrefixes: []string{"Wasteland", "Raider", "Mutant", "Savage"},
			NameSuffixes: []string{"Marauder", "Brute", "Berserker", "Hunter"},
			Tags:         []string{"mutant", "aggressive", "human"},
			HealthRange:  [2]int{65, 105},
			DamageRange:  [2]int{11, 19},
			DefenseRange: [2]int{4, 9},
			SpeedRange:   [2]float64{0.8, 1.0},
		},
		{
			BaseType:     TypeBoss,
			BaseSize:     SizeHuge,
			NamePrefixes: []string{"Radiation", "Apex", "Warlord", "Mutant"},
			NameSuffixes: []string{"Beast", "King", "Overlord", "Titan"},
			Tags:         []string{"boss", "mutant", "massive"},
			HealthRange:  [2]int{420, 620},
			DamageRange:  [2]int{42, 72},
			DefenseRange: [2]int{16, 28},
			SpeedRange:   [2]float64{0.5, 0.7},
		},
		{
			BaseType:     TypeNPC,
			BaseSize:     SizeMedium,
			NamePrefixes: []string{"Wasteland", "Scrap", "Salvage", "Nomad"},
			NameSuffixes: []string{"Trader", "Merchant", "Peddler", "Dealer"},
			Tags:         []string{"friendly", "trader", "merchant"},
			HealthRange:  [2]int{50, 100},
			DamageRange:  [2]int{5, 10},
			DefenseRange: [2]int{5, 10},
			SpeedRange:   [2]float64{1.0, 1.0},
		},
	}
}
