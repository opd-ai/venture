package dialog

import (
	"fmt"
	"math/rand"
)

// PersonalityType defines common NPC personality archetypes.
type PersonalityType string

const (
	// PersonalityHelpful represents friendly, cooperative NPCs.
	PersonalityHelpful PersonalityType = "helpful"

	// PersonalityMerchant represents traders and shopkeepers.
	PersonalityMerchant PersonalityType = "merchant"

	// PersonalityHostile represents aggressive or unfriendly NPCs.
	PersonalityHostile PersonalityType = "hostile"

	// PersonalityMysterious represents enigmatic or cryptic NPCs.
	PersonalityMysterious PersonalityType = "mysterious"

	// PersonalityScholarly represents wise or knowledgeable NPCs.
	PersonalityScholarly PersonalityType = "scholarly"

	// PersonalityWarrior represents brave, combat-focused NPCs.
	PersonalityWarrior PersonalityType = "warrior"

	// PersonalityTimid represents fearful or cautious NPCs.
	PersonalityTimid PersonalityType = "timid"

	// PersonalityArrogant represents prideful or condescending NPCs.
	PersonalityArrogant PersonalityType = "arrogant"
)

// Personality defines NPC character traits that influence dialog generation.
type Personality struct {
	// Type is the personality archetype.
	Type PersonalityType

	// Friendliness affects greeting warmth and helpfulness (0.0-1.0).
	// 0.0 = hostile, 0.5 = neutral, 1.0 = extremely friendly
	Friendliness float64

	// Verbosity affects response length (0.0-1.0).
	// 0.0 = terse/brief, 0.5 = normal, 1.0 = very wordy
	Verbosity float64

	// Formality affects language style (0.0-1.0).
	// 0.0 = casual/slang, 0.5 = normal, 1.0 = very formal
	Formality float64

	// Humor affects joke frequency and lightheartedness (0.0-1.0).
	// 0.0 = serious, 0.5 = occasional humor, 1.0 = constant jokes
	Humor float64

	// Knowledge affects technical/lore detail level (0.0-1.0).
	// 0.0 = simple explanations, 1.0 = complex details
	Knowledge float64
}

// NewPersonality creates a personality with specified type and default traits.
func NewPersonality(ptype PersonalityType) *Personality {
	p := &Personality{
		Type:         ptype,
		Friendliness: 0.5,
		Verbosity:    0.5,
		Formality:    0.5,
		Humor:        0.5,
		Knowledge:    0.5,
	}

	// Adjust traits based on archetype
	switch ptype {
	case PersonalityHelpful:
		p.Friendliness = 0.8
		p.Verbosity = 0.6
		p.Formality = 0.4
		p.Humor = 0.6

	case PersonalityMerchant:
		p.Friendliness = 0.7
		p.Verbosity = 0.7
		p.Formality = 0.6
		p.Humor = 0.4

	case PersonalityHostile:
		p.Friendliness = 0.2
		p.Verbosity = 0.3
		p.Formality = 0.3
		p.Humor = 0.1

	case PersonalityMysterious:
		p.Friendliness = 0.4
		p.Verbosity = 0.4
		p.Formality = 0.7
		p.Humor = 0.2
		p.Knowledge = 0.8

	case PersonalityScholarly:
		p.Friendliness = 0.6
		p.Verbosity = 0.8
		p.Formality = 0.8
		p.Humor = 0.3
		p.Knowledge = 0.9

	case PersonalityWarrior:
		p.Friendliness = 0.5
		p.Verbosity = 0.4
		p.Formality = 0.5
		p.Humor = 0.5
		p.Knowledge = 0.4

	case PersonalityTimid:
		p.Friendliness = 0.6
		p.Verbosity = 0.3
		p.Formality = 0.6
		p.Humor = 0.2

	case PersonalityArrogant:
		p.Friendliness = 0.3
		p.Verbosity = 0.6
		p.Formality = 0.7
		p.Humor = 0.2
		p.Knowledge = 0.7
	}

	return p
}

// GenerateRandomPersonality creates a personality with random traits.
func GenerateRandomPersonality(rng *rand.Rand) *Personality {
	types := []PersonalityType{
		PersonalityHelpful,
		PersonalityMerchant,
		PersonalityHostile,
		PersonalityMysterious,
		PersonalityScholarly,
		PersonalityWarrior,
		PersonalityTimid,
		PersonalityArrogant,
	}

	ptype := types[rng.Intn(len(types))]
	p := NewPersonality(ptype)

	// Add random variation (±0.2)
	p.Friendliness = clamp(p.Friendliness+(rng.Float64()-0.5)*0.4, 0.0, 1.0)
	p.Verbosity = clamp(p.Verbosity+(rng.Float64()-0.5)*0.4, 0.0, 1.0)
	p.Formality = clamp(p.Formality+(rng.Float64()-0.5)*0.4, 0.0, 1.0)
	p.Humor = clamp(p.Humor+(rng.Float64()-0.5)*0.4, 0.0, 1.0)
	p.Knowledge = clamp(p.Knowledge+(rng.Float64()-0.5)*0.4, 0.0, 1.0)

	return p
}

// ApplyToGenerator modifies generation parameters based on personality.
//
// This adjusts temperature, min/max words, and other parameters to match
// the NPC's personality traits. For example:
//   - High verbosity → longer responses
//   - High friendliness → warmer temperature (more variation)
//   - High knowledge → longer, more detailed responses
func (p *Personality) ApplyToGenerator(params *GenerateParams) {
	// Verbosity affects word count
	if p.Verbosity < 0.3 {
		// Terse: reduce max words
		params.MaxWords = int(float64(params.MaxWords) * 0.6)
		params.MinWords = int(float64(params.MinWords) * 0.5)
	} else if p.Verbosity > 0.7 {
		// Wordy: increase max words
		params.MaxWords = int(float64(params.MaxWords) * 1.4)
		params.MinWords = int(float64(params.MinWords) * 1.2)
	}

	// Friendliness affects temperature (friendly = more varied)
	if p.Friendliness > 0.7 {
		params.Temperature *= 1.2
	} else if p.Friendliness < 0.3 {
		params.Temperature *= 0.8
	}

	// Knowledge affects verbosity (knowledgeable NPCs speak more)
	if p.Knowledge > 0.7 {
		params.MaxWords = int(float64(params.MaxWords) * 1.2)
	}

	// Clamp values to reasonable ranges
	params.MaxWords = max(10, min(100, params.MaxWords))
	params.MinWords = max(5, min(50, params.MinWords))
	params.Temperature = clamp(params.Temperature, 0.0, 1.0)

	// Ensure MinWords <= MaxWords
	if params.MinWords > params.MaxWords {
		params.MinWords = params.MaxWords / 2
	}
}

// GetGreeting returns a personality-appropriate greeting using a fixed deterministic
// selection (always the first greeting). Maintained for backward compatibility.
// For randomized greetings, use GetGreetingWithSeed instead.
func (p *Personality) GetGreeting(genreID string) string {
	greetings := p.buildGreetingsMap()

	// Get greeting for this personality and genre
	if genreGreetings, ok := greetings[p.Type]; ok {
		if greetingList, ok := genreGreetings[genreID]; ok && len(greetingList) > 0 {
			// Return first greeting for backward compatibility
			// Use GetGreetingWithSeed for randomized selection
			return greetingList[0]
		}
	}

	// Fallback generic greeting
	return "Hello."
}

// GetGreetingWithSeed returns a deterministically randomized personality-appropriate greeting.
//
// Uses the provided seed to select from available greetings for this personality
// and genre combination. Same seed always produces the same greeting, ensuring
// deterministic behavior required for multiplayer synchronization.
//
// Example:
//
//	p := NewPersonality(PersonalityMerchant)
//	greeting := p.GetGreetingWithSeed("fantasy", npcSeed)
func (p *Personality) GetGreetingWithSeed(genreID string, seed int64) string {
	greetings := p.buildGreetingsMap()

	// Get greeting for this personality and genre
	if genreGreetings, ok := greetings[p.Type]; ok {
		if greetingList, ok := genreGreetings[genreID]; ok && len(greetingList) > 0 {
			rng := rand.New(rand.NewSource(seed))
			return greetingList[rng.Intn(len(greetingList))]
		}
	}

	// Fallback generic greeting
	return "Hello."
}

// buildGreetingsMap returns the full greeting database for all personality types.
func (p *Personality) buildGreetingsMap() map[PersonalityType]map[string][]string {
	greetings := make(map[PersonalityType]map[string][]string)

	greetings[PersonalityHelpful] = map[string][]string{
		"fantasy":   {"Greetings, friend!", "Welcome, traveler!", "Well met!"},
		"scifi":     {"Hello there, citizen.", "Greetings, traveler.", "Welcome to the station."},
		"horror":    {"You... you're still alive?", "Please, help me...", "Don't go in there!"},
		"cyberpunk": {"Hey choom, need something?", "What's the word?", "You looking for work?"},
		"postapoc":  {"Another survivor!", "Glad to see a friendly face.", "You made it through!"},
	}

	greetings[PersonalityMerchant] = map[string][]string{
		"fantasy":   {"Looking to trade?", "I have wares if you have coin.", "What can I get you?"},
		"scifi":     {"Browse my inventory.", "All items guaranteed.", "What do you need?"},
		"horror":    {"Take what you need, quickly.", "I have supplies... for a price.", "Coin still spends here."},
		"cyberpunk": {"Got eddies? Got goods.", "What're you buying?", "Best prices in the sector."},
		"postapoc":  {"What have you got to trade?", "Let's see your scrap.", "Barter or bullets?"},
	}

	greetings[PersonalityHostile] = map[string][]string{
		"fantasy":   {"What do you want?", "Move along.", "This doesn't concern you."},
		"scifi":     {"State your business.", "You're not authorized.", "Access denied."},
		"horror":    {"Get out while you can.", "Leave me alone!", "You shouldn't be here."},
		"cyberpunk": {"What're you looking at?", "Beat it, gonk.", "This is my turf."},
		"postapoc":  {"Keep moving.", "Don't have time for this.", "You lost?"},
	}

	greetings[PersonalityMysterious] = map[string][]string{
		"fantasy":   {"We meet at last.", "The threads of fate converge.", "You seek answers?"},
		"scifi":     {"Interesting timing.", "The data suggested you would come.", "Calculating probabilities..."},
		"horror":    {"They told me you would come.", "Do you hear them too?", "The signs were clear."},
		"cyberpunk": {"I know what you're looking for.", "The matrix whispers your name.", "Encrypted messages, decrypted fate."},
		"postapoc":  {"The old ways predicted this.", "Signs in the wastes led you here.", "I've been waiting."},
	}

	greetings[PersonalityScholarly] = map[string][]string{
		"fantasy":   {"Ah, a student of knowledge!", "Welcome, seeker of wisdom.", "What brings you to my studies?"},
		"scifi":     {"Fascinating data patterns today.", "Research continues apace.", "Science waits for no one."},
		"horror":    {"The texts spoke of this.", "Ancient knowledge preserved.", "I've documented everything."},
		"cyberpunk": {"The algorithms predicted this interaction.", "Processing your inquiry.", "Data streams converge."},
		"postapoc":  {"Pre-war knowledge is precious.", "I preserve what was lost.", "The old books still teach."},
	}

	greetings[PersonalityWarrior] = map[string][]string{
		"fantasy":   {"Hail, warrior!", "Your blade looks sharp.", "Ready for battle?"},
		"scifi":     {"Weapons ready.", "Combat systems online.", "Reporting for duty."},
		"horror":    {"Still fighting?", "We must survive.", "Stand and fight!"},
		"cyberpunk": {"Chrome and steel, choom.", "Lock and load.", "Combat protocol engaged."},
		"postapoc":  {"Armed and ready.", "Survival first.", "Watch your back out there."},
	}

	greetings[PersonalityTimid] = map[string][]string{
		"fantasy":   {"Oh! You startled me.", "P-please don't hurt me.", "Are you... friendly?"},
		"scifi":     {"Identify yourself!", "Don't come closer!", "Who sent you?"},
		"horror":    {"They're... they're everywhere!", "Hide! Quickly!", "We're not safe!"},
		"cyberpunk": {"Don't shoot!", "I'm just a civilian!", "I didn't see anything!"},
		"postapoc":  {"Please, I have nothing!", "Don't take my supplies!", "I'm no threat!"},
	}

	greetings[PersonalityArrogant] = map[string][]string{
		"fantasy":   {"Do you know who I am?", "Make it quick, peasant.", "I suppose I can spare a moment."},
		"scifi":     {"Your clearance level is insufficient.", "Executive priority only.", "This better be important."},
		"horror":    {"I'm above such concerns.", "My position protects me.", "You wouldn't understand."},
		"cyberpunk": {"Corpo life, choom. You wouldn't get it.", "My time is valuable.", "Street trash..."},
		"postapoc":  {"I've survived this long for a reason.", "Know your place.", "I've seen worse than you."},
	}

	return greetings
}

// String returns a human-readable description of the personality.
func (p *Personality) String() string {
	return fmt.Sprintf("Personality{type=%s, friendliness=%.2f, verbosity=%.2f, formality=%.2f, humor=%.2f, knowledge=%.2f}",
		p.Type, p.Friendliness, p.Verbosity, p.Formality, p.Humor, p.Knowledge)
}
