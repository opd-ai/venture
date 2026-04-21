package story

import (
	"fmt"
	"math/rand"
	"sort"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/procgen"
)

// HistoricalEvent represents a single event in the world's history
type HistoricalEvent struct {
	Timestamp   int64  // Years before present (negative = future, positive = past)
	Title       string // Event name
	Description string // Detailed description
	EventType   EventType
	Importance  float64 // 0.0-1.0, how significant the event was
	Faction     string  // Which faction was involved (if any)
	Location    string  // Where it happened
}

// String returns the string representation of EventType
func (e EventType) String() string {
	switch e {
	case EventFoundation:
		return "Foundation"
	case EventWar:
		return "War"
	case EventDiscovery:
		return "Discovery"
	case EventCatastrophe:
		return "Catastrophe"
	case EventRenaissance:
		return "Renaissance"
	case EventCollapse:
		return "Collapse"
	case EventContact:
		return "Contact"
	case EventRitual:
		return "Ritual"
	default:
		return "Unknown"
	}
}

// Timeline represents the historical timeline of the game world
type Timeline struct {
	WorldSeed   int64             // Seed for timeline generation
	Genre       string            // World genre
	Events      []HistoricalEvent // All events in chronological order
	Eras        []Era             // Major historical eras
	StartYear   int64             // Earliest event (most ancient)
	CurrentYear int64             // Present day (always 0)
	Consistency float64           // How well events connect (0.0-1.0)
}

// Era represents a major period in history
type Era struct {
	Name            string // Era name
	StartYear       int64  // When era began
	EndYear         int64  // When era ended
	Description     string // What defined this era
	DominantFaction string // Primary power during this era
}

// TimelineGenerator creates consistent historical timelines
type TimelineGenerator struct{}

// NewTimelineGenerator creates a new timeline generator
func NewTimelineGenerator() *TimelineGenerator {
	return &TimelineGenerator{}
}

// Generate creates a historical timeline for the world
func (g *TimelineGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if params.Difficulty < 0 || params.Difficulty > 1.0 {
		log.WithFields(log.Fields{
			"seed":       seed,
			"difficulty": params.Difficulty,
		}).Error("invalid difficulty parameter for timeline generation")
		return nil, fmt.Errorf("difficulty must be between 0 and 1, got %.2f", params.Difficulty)
	}

	log.WithFields(log.Fields{
		"seed":  seed,
		"genre": params.GenreID,
		"depth": params.Depth,
	}).Debug("generating timeline")

	rng := rand.New(rand.NewSource(seed))

	// Determine timeline span (100-1000 years based on depth)
	timeSpan := 100 + params.Depth*50
	if timeSpan > 1000 {
		timeSpan = 1000
	}

	startYear := int64(timeSpan)

	// Generate eras (2-5 based on time span)
	numEras := 2 + int(float64(timeSpan)/300)
	if numEras > 5 {
		numEras = 5
	}

	eras := g.generateEras(rng, params.GenreID, numEras, startYear, 0)

	// Generate events (5-15 per era)
	eventsPerEra := 5 + int(params.Difficulty*10)
	if eventsPerEra > 15 {
		eventsPerEra = 15
	}

	events := make([]HistoricalEvent, 0, numEras*eventsPerEra)
	for _, era := range eras {
		eraEvents := g.generateEventsForEra(rng, params.GenreID, era, eventsPerEra)
		events = append(events, eraEvents...)
	}

	// Sort events chronologically (most ancient first)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp > events[j].Timestamp
	})

	timeline := &Timeline{
		WorldSeed:   seed,
		Genre:       params.GenreID,
		Events:      events,
		Eras:        eras,
		StartYear:   startYear,
		CurrentYear: 0,
		Consistency: g.calculateConsistency(events, eras),
	}

	return timeline, nil
}

// Validate checks timeline quality
func (g *TimelineGenerator) Validate(result interface{}) error {
	timeline, ok := result.(*Timeline)
	if !ok {
		return fmt.Errorf("result is not a *Timeline")
	}

	if len(timeline.Eras) < 2 {
		return fmt.Errorf("too few eras: %d, minimum 2", len(timeline.Eras))
	}

	if len(timeline.Eras) > 5 {
		return fmt.Errorf("too many eras: %d, maximum 5", len(timeline.Eras))
	}

	if len(timeline.Events) < 10 {
		return fmt.Errorf("too few events: %d, minimum 10", len(timeline.Events))
	}

	if timeline.StartYear <= timeline.CurrentYear {
		return fmt.Errorf("start year (%d) must be greater than current year (%d)", timeline.StartYear, timeline.CurrentYear)
	}

	if timeline.Consistency < 0.5 {
		return fmt.Errorf("timeline consistency too low: %.2f, minimum 0.5", timeline.Consistency)
	}

	// Validate events are chronological
	for i := 0; i < len(timeline.Events)-1; i++ {
		if timeline.Events[i].Timestamp < timeline.Events[i+1].Timestamp {
			return fmt.Errorf("events not in chronological order at index %d", i)
		}
	}

	return nil
}

// generateEras creates major historical periods
func (g *TimelineGenerator) generateEras(rng *rand.Rand, genreID string, numEras int, startYear, endYear int64) []Era {
	eras := make([]Era, numEras)
	eraSpan := (startYear - endYear) / int64(numEras)

	eraTemplates := g.getEraTemplates(genreID)

	for i := 0; i < numEras; i++ {
		template := eraTemplates[i%len(eraTemplates)]

		eraStart := startYear - int64(i)*eraSpan
		eraEnd := eraStart - eraSpan
		if i == numEras-1 {
			eraEnd = endYear // Last era ends at present
		}

		eras[i] = Era{
			Name:            g.generateEraName(rng, genreID, template, i),
			StartYear:       eraStart,
			EndYear:         eraEnd,
			Description:     template,
			DominantFaction: g.generateFactionName(rng, genreID),
		}
	}

	return eras
}

// getEraTemplates returns genre-specific era types
func (g *TimelineGenerator) getEraTemplates(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{
			"Age of Creation",
			"Era of Magic",
			"Time of Dragons",
			"Age of Heroes",
			"Dark Ages",
		}
	case "scifi":
		return []string{
			"Pre-Contact Era",
			"First Expansion",
			"Golden Age of Technology",
			"AI Singularity",
			"Post-Scarcity Period",
		}
	case "horror":
		return []string{
			"Before the Corruption",
			"First Awakening",
			"Age of Madness",
			"The Long Night",
			"Era of Despair",
		}
	case "cyberpunk":
		return []string{
			"Corporate Formation",
			"Network Revolution",
			"Megacity Era",
			"Digital Divide",
			"Neo-Feudalism",
		}
	case "postapoc":
		return []string{
			"Before the Fall",
			"The Cataclysm",
			"Survival Period",
			"Rebuilding Age",
			"New Order",
		}
	default:
		return []string{
			"Ancient Times",
			"Classical Period",
			"Medieval Era",
			"Renaissance",
			"Modern Age",
		}
	}
}

// generateEraName creates a name for an era
func (g *TimelineGenerator) generateEraName(rng *rand.Rand, genreID, template string, index int) string {
	return template
}

// generateEventsForEra creates events within an era
func (g *TimelineGenerator) generateEventsForEra(rng *rand.Rand, genreID string, era Era, numEvents int) []HistoricalEvent {
	events := make([]HistoricalEvent, numEvents)
	eraSpan := era.StartYear - era.EndYear

	for i := 0; i < numEvents; i++ {
		// Distribute events throughout era
		progress := float64(i) / float64(numEvents)
		timestamp := era.StartYear - int64(progress*float64(eraSpan))

		// Select event type based on era template and progress
		eventType := g.selectEventType(rng, era.Description, progress)

		events[i] = HistoricalEvent{
			Timestamp:   timestamp,
			Title:       g.generateEventTitle(rng, genreID, eventType, era),
			Description: g.generateEventDescription(rng, genreID, eventType, era),
			EventType:   eventType,
			Importance:  rng.Float64(), // Random importance 0-1
			Faction:     g.selectEventFaction(rng, era.DominantFaction),
			Location:    g.generateLocationName(rng, genreID),
		}
	}

	return events
}

// selectEventType chooses appropriate event type for era/progress
func (g *TimelineGenerator) selectEventType(rng *rand.Rand, eraTemplate string, progress float64) EventType {
	// Early in era: foundations and discoveries
	if progress < 0.3 {
		types := []EventType{EventFoundation, EventDiscovery, EventContact}
		return types[rng.Intn(len(types))]
	}

	// Middle of era: wars and renaissances
	if progress < 0.7 {
		types := []EventType{EventWar, EventRenaissance, EventRitual}
		return types[rng.Intn(len(types))]
	}

	// End of era: catastrophes and collapses
	types := []EventType{EventCatastrophe, EventCollapse, EventWar}
	return types[rng.Intn(len(types))]
}

// generateEventTitle creates a title for a historical event
func (g *TimelineGenerator) generateEventTitle(rng *rand.Rand, genreID string, eventType EventType, era Era) string {
	switch eventType {
	case EventFoundation:
		subjects := g.getFoundationSubjects(genreID)
		return fmt.Sprintf("Founding of %s", subjects[rng.Intn(len(subjects))])

	case EventWar:
		conflicts := g.getWarNames(genreID)
		return conflicts[rng.Intn(len(conflicts))]

	case EventDiscovery:
		discoveries := g.getDiscoveries(genreID)
		return fmt.Sprintf("Discovery of %s", discoveries[rng.Intn(len(discoveries))])

	case EventCatastrophe:
		disasters := g.getCatastrophes(genreID)
		return disasters[rng.Intn(len(disasters))]

	case EventRenaissance:
		advances := g.getRenaissances(genreID)
		return advances[rng.Intn(len(advances))]

	case EventCollapse:
		falls := g.getCollapses(genreID)
		return falls[rng.Intn(len(falls))]

	case EventContact:
		contacts := g.getContactEvents(genreID)
		return contacts[rng.Intn(len(contacts))]

	case EventRitual:
		rituals := g.getRituals(genreID)
		return rituals[rng.Intn(len(rituals))]

	default:
		return "Unknown Event"
	}
}

// Genre-specific event name generators
func (g *TimelineGenerator) getFoundationSubjects(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"the Kingdom", "the Mage Council", "the Dragon Alliance", "the Sacred Order"}
	case "scifi":
		return []string{"the Colony", "the Federation", "the Orbital Station", "the Research Facility"}
	case "horror":
		return []string{"the Cult", "the Dark Temple", "the Cursed Monastery", "the Blood Circle"}
	case "cyberpunk":
		return []string{"the Corporation", "the Underground Network", "the Mega-City", "the Data Collective"}
	case "postapoc":
		return []string{"the Settlement", "the Trading Post", "the Survivor Enclave", "the Bunker City"}
	default:
		return []string{"the City", "the Empire", "the Alliance", "the Federation"}
	}
}

func (g *TimelineGenerator) getWarNames(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"The Dragon War", "War of Mages", "The Dark Crusade", "Battle for the Crown"}
	case "scifi":
		return []string{"The AI Conflict", "First Contact War", "The Colony Rebellion", "Corporate Wars"}
	case "horror":
		return []string{"The Purge", "Night of Blood", "The Cleansing", "War Against Darkness"}
	case "cyberpunk":
		return []string{"The Network War", "Corporate Takeover", "Data Revolution", "The Great Hack"}
	case "postapoc":
		return []string{"The Resource War", "Wasteland Conflict", "Battle for Water", "Survivor's War"}
	default:
		return []string{"The Great War", "Civil Conflict", "The Rebellion", "Border Wars"}
	}
}

func (g *TimelineGenerator) getDiscoveries(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"Ancient Magic", "Dragon Fire", "The Sacred Texts", "Immortality Spell"}
	case "scifi":
		return []string{"Faster-Than-Light Travel", "Alien Technology", "AI Consciousness", "Terraforming"}
	case "horror":
		return []string{"The Book of Shadows", "Gateway to Beyond", "Forbidden Knowledge", "Dark Ritual"}
	case "cyberpunk":
		return []string{"Neural Interface", "Digital Consciousness", "Quantum Encryption", "Nanotech"}
	case "postapoc":
		return []string{"Clean Water Source", "Functional Technology", "Radiation Cure", "Safe Zone"}
	default:
		return []string{"New Technology", "Lost Knowledge", "Ancient Artifact", "Hidden Truth"}
	}
}

func (g *TimelineGenerator) getCatastrophes(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"The Great Plague", "Dragon Apocalypse", "Magical Cataclysm", "The Dark Summoning"}
	case "scifi":
		return []string{"Colony Failure", "AI Uprising", "Reactor Meltdown", "Asteroid Impact"}
	case "horror":
		return []string{"The Awakening", "Outbreak of Madness", "The Possession", "Portal Opening"}
	case "cyberpunk":
		return []string{"Grid Collapse", "Corporate Massacre", "System Virus", "Mass Data Loss"}
	case "postapoc":
		return []string{"Nuclear Holocaust", "Biological Plague", "Environmental Collapse", "Solar Flare"}
	default:
		return []string{"Natural Disaster", "Economic Collapse", "Epidemic", "Invasion"}
	}
}

func (g *TimelineGenerator) getRenaissances(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"Age of Enlightenment", "Magical Renaissance", "Golden Age of Heroes", "Cultural Flowering"}
	case "scifi":
		return []string{"Technological Leap", "Scientific Revolution", "Golden Age of Exploration", "Innovation Era"}
	case "horror":
		return []string{"Brief Peace", "Period of Recovery", "Temporary Sanctuary", "False Dawn"}
	case "cyberpunk":
		return []string{"Digital Renaissance", "Network Boom", "Tech Revolution", "Cultural Awakening"}
	case "postapoc":
		return []string{"Rebuilding Era", "New Growth", "Hope Returns", "Reconstruction Period"}
	default:
		return []string{"Renaissance", "Cultural Revolution", "Age of Progress", "Enlightenment"}
	}
}

func (g *TimelineGenerator) getCollapses(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"Fall of the Kingdom", "End of the Dragon Age", "Collapse of Magic", "Empire's Ruin"}
	case "scifi":
		return []string{"Colony Abandonment", "Federation Dissolution", "Tech Collapse", "Station Evacuation"}
	case "horror":
		return []string{"Final Descent", "Lost to Darkness", "End of Hope", "Complete Corruption"}
	case "cyberpunk":
		return []string{"Corporate Fall", "Network Shutdown", "System Failure", "City Abandonment"}
	case "postapoc":
		return []string{"Settlement Lost", "Enclave Overrun", "Resource Depletion", "Final Exodus"}
	default:
		return []string{"Civilization Fall", "Empire Collapse", "Social Breakdown", "System Failure"}
	}
}

func (g *TimelineGenerator) getContactEvents(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"First Dragon Sighting", "Meeting with Elves", "Demon Encounter", "Angel Visitation"}
	case "scifi":
		return []string{"First Contact", "Alien Discovery", "Unknown Signal", "Interstellar Meeting"}
	case "horror":
		return []string{"First Sighting", "Contact with Beyond", "Entity Appearance", "Otherworldly Encounter"}
	case "cyberpunk":
		return []string{"AI First Communication", "Network Connection", "Corporate Contact", "Underground Alliance"}
	case "postapoc":
		return []string{"Survivor Contact", "Trader Discovery", "Settlement Found", "Radio Contact"}
	default:
		return []string{"First Meeting", "New Alliance", "Discovery", "Initial Contact"}
	}
}

func (g *TimelineGenerator) getRituals(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"The Grand Summoning", "Crown Ceremony", "Dragon Binding", "Sacred Rite"}
	case "scifi":
		return []string{"Launch Ceremony", "AI Activation", "First Flight", "Station Dedication"}
	case "horror":
		return []string{"Blood Ritual", "Dark Ceremony", "Summoning Circle", "Unholy Rite"}
	case "cyberpunk":
		return []string{"System Initialization", "Network Baptism", "Corporate Merger", "Data Ceremony"}
	case "postapoc":
		return []string{"Remembrance Day", "Founding Ceremony", "Hope Festival", "Survivor's Oath"}
	default:
		return []string{"Grand Ceremony", "Sacred Rite", "Royal Coronation", "Religious Festival"}
	}
}

// generateEventDescription creates a detailed description
func (g *TimelineGenerator) generateEventDescription(rng *rand.Rand, genreID string, eventType EventType, era Era) string {
	templates := []string{
		"During the %s, this event marked a turning point in history.",
		"A significant moment in the %s that changed everything.",
		"One of the defining events of the %s.",
		"This occurrence during the %s reshaped the world.",
	}

	template := templates[rng.Intn(len(templates))]
	return fmt.Sprintf(template, era.Name, era.Name, era.Name, era.Name)
}

// selectEventFaction chooses which faction was involved
func (g *TimelineGenerator) selectEventFaction(rng *rand.Rand, dominantFaction string) string {
	// 70% chance dominant faction, 30% chance other
	if rng.Float64() < 0.7 {
		return dominantFaction
	}

	others := []string{"Rebels", "Outsiders", "Unknown", "Opposition", "Rivals"}
	return others[rng.Intn(len(others))]
}

// generateFactionName creates a faction name
func (g *TimelineGenerator) generateFactionName(rng *rand.Rand, genreID string) string {
	prefixes := g.getFactionPrefixes(genreID)
	suffixes := g.getFactionSuffixes(genreID)

	prefix := prefixes[rng.Intn(len(prefixes))]
	suffix := suffixes[rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s %s", prefix, suffix)
}

func (g *TimelineGenerator) getFactionPrefixes(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"Ancient", "Royal", "Dark", "Sacred", "Silver", "Golden"}
	case "scifi":
		return []string{"United", "Galactic", "Stellar", "Orbital", "Colonial", "Federal"}
	case "horror":
		return []string{"Cursed", "Blood", "Shadow", "Twisted", "Damned", "Forsaken"}
	case "cyberpunk":
		return []string{"Mega", "Cyber", "Neo", "Digital", "Corporate", "Underground"}
	case "postapoc":
		return []string{"Wasteland", "Survivor", "Raider", "Bunker", "Scavenger", "Nomad"}
	default:
		return []string{"First", "Grand", "High", "Imperial", "Free", "United"}
	}
}

func (g *TimelineGenerator) getFactionSuffixes(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"Kingdom", "Order", "Council", "Alliance", "Empire", "Brotherhood"}
	case "scifi":
		return []string{"Federation", "Coalition", "Consortium", "Alliance", "Collective", "Union"}
	case "horror":
		return []string{"Cult", "Circle", "Coven", "Sect", "Order", "Brotherhood"}
	case "cyberpunk":
		return []string{"Corporation", "Syndicate", "Network", "Collective", "Coalition", "Cartel"}
	case "postapoc":
		return []string{"Tribe", "Gang", "Enclave", "Settlement", "Clan", "Band"}
	default:
		return []string{"Nation", "State", "Republic", "Kingdom", "Alliance", "Coalition"}
	}
}

// generateLocationName creates a location name
func (g *TimelineGenerator) generateLocationName(rng *rand.Rand, genreID string) string {
	prefixes := []string{"New", "Old", "Lost", "Hidden", "Ancient", "Forgotten"}
	locations := g.getLocationNames(genreID)

	prefix := prefixes[rng.Intn(len(prefixes))]
	location := locations[rng.Intn(len(locations))]

	return fmt.Sprintf("%s %s", prefix, location)
}

func (g *TimelineGenerator) getLocationNames(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"Keep", "Tower", "Forest", "Mountain", "Valley", "Castle"}
	case "scifi":
		return []string{"Station", "Colony", "Sector", "Outpost", "Facility", "Platform"}
	case "horror":
		return []string{"Manor", "Asylum", "Cemetery", "Mansion", "Catacombs", "Ruins"}
	case "cyberpunk":
		return []string{"District", "Level", "Sector", "Block", "Zone", "Quarter"}
	case "postapoc":
		return []string{"Ruins", "Wasteland", "Bunker", "Settlement", "Deadzone", "Shelter"}
	default:
		return []string{"City", "Town", "Village", "Port", "Capital", "Fortress"}
	}
}

// calculateConsistency measures how well events fit together
func (g *TimelineGenerator) calculateConsistency(events []HistoricalEvent, eras []Era) float64 {
	if len(events) == 0 || len(eras) == 0 {
		return 0.5
	}

	if !g.allErasHaveEvents(events, eras) {
		return 0.5
	}

	foundationCount, collapseCount := g.countCriticalEvents(events)
	if foundationCount == 0 || collapseCount == 0 {
		return 0.6
	}

	return g.calculateScore(foundationCount, collapseCount, len(events))
}

// allErasHaveEvents verifies that all eras have at least one event.
func (g *TimelineGenerator) allErasHaveEvents(events []HistoricalEvent, eras []Era) bool {
	eventsPerEra := g.mapEventsToEras(events, eras)
	return len(eventsPerEra) >= len(eras)
}

// mapEventsToEras distributes events across their respective eras.
func (g *TimelineGenerator) mapEventsToEras(events []HistoricalEvent, eras []Era) map[string]int {
	eventsPerEra := make(map[string]int)
	for _, event := range events {
		for _, era := range eras {
			if event.Timestamp >= era.EndYear && event.Timestamp <= era.StartYear {
				eventsPerEra[era.Name]++
				break
			}
		}
	}
	return eventsPerEra
}

// countCriticalEvents counts foundation and collapse events for consistency check.
func (g *TimelineGenerator) countCriticalEvents(events []HistoricalEvent) (int, int) {
	foundationCount := 0
	collapseCount := 0
	for _, event := range events {
		if event.EventType == EventFoundation {
			foundationCount++
		} else if event.EventType == EventCollapse {
			collapseCount++
		}
	}
	return foundationCount, collapseCount
}

// calculateScore computes the final consistency score based on critical events.
func (g *TimelineGenerator) calculateScore(foundationCount, collapseCount, totalEvents int) float64 {
	criticalRatio := float64(foundationCount+collapseCount) / float64(totalEvents)
	return 0.7 + criticalRatio*0.3
}

// GetEventsInPeriod returns all events within a time range
func (t *Timeline) GetEventsInPeriod(startYear, endYear int64) []HistoricalEvent {
	result := make([]HistoricalEvent, 0)

	for _, event := range t.Events {
		if event.Timestamp >= endYear && event.Timestamp <= startYear {
			result = append(result, event)
		}
	}

	return result
}

// GetCurrentEra returns the era containing the current year
func (t *Timeline) GetCurrentEra() *Era {
	for i := range t.Eras {
		if t.CurrentYear >= t.Eras[i].EndYear && t.CurrentYear <= t.Eras[i].StartYear {
			return &t.Eras[i]
		}
	}
	return nil
}

// GetEventsByType returns all events of a specific type
func (t *Timeline) GetEventsByType(eventType EventType) []HistoricalEvent {
	result := make([]HistoricalEvent, 0)

	for _, event := range t.Events {
		if event.EventType == eventType {
			result = append(result, event)
		}
	}

	return result
}
