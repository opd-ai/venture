package world_events

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateFactionResponse creates a faction's reaction to player actions.
// The triggerAction parameter influences the type of response generated:
// hostile actions (e.g., "attack", "kill", "steal") increase negative responses,
// while diplomatic actions (e.g., "trade", "help", "negotiate") increase positive responses.
func GenerateFactionResponse(seed int64, factionID, triggerAction string, severity Severity) *FactionResponse {
	rng := rand.New(rand.NewSource(seed))

	// Determine base reaction based on trigger action
	actionModifier := 0.0
	isHostileAction := false
	isDiplomaticAction := false

	// Classify the trigger action
	switch triggerAction {
	case "attack", "kill", "steal", "destroy", "raid", "invasion", "guild_war":
		isHostileAction = true
		actionModifier = 0.2 // Increases negative response
	case "trade", "help", "negotiate", "gift", "alliance", "peace":
		isDiplomaticAction = true
		actionModifier = -0.2 // Decreases negative response
	}

	repChange := -0.05 * float64(severity)
	hostilityChange := 0.1 * float64(severity)
	tradeBonus := 0.0

	// Apply action modifier to reputation and hostility
	if isHostileAction {
		repChange -= actionModifier * float64(severity)
		hostilityChange += actionModifier * float64(severity)
	} else if isDiplomaticAction {
		repChange -= actionModifier * float64(severity) // Subtracting negative = adding
		hostilityChange += actionModifier * float64(severity)
	}

	if rng.Float64() < 0.2 {
		repChange = -repChange
		hostilityChange = -hostilityChange
		tradeBonus = 0.1 * float64(severity)
	}

	responseTypes := []string{
		"disapproval",
		"concern",
		"warning",
		"condemnation",
		"support",
		"neutrality",
	}

	responseType := responseTypes[rng.Intn(len(responseTypes))]
	if repChange < 0 {
		responseType = responseTypes[rng.Intn(4)]
	} else {
		responseType = responseTypes[4+rng.Intn(2)]
	}

	messages := map[string][]string{
		"disapproval": {
			"The faction expresses disapproval of recent actions.",
			"Faction leaders issue statements of discontent.",
		},
		"concern": {
			"The faction voices concern over recent developments.",
			"Faction representatives request clarification on recent events.",
		},
		"warning": {
			"The faction issues a formal warning.",
			"Faction leadership threatens consequences if actions continue.",
		},
		"condemnation": {
			"The faction formally condemns recent actions.",
			"Faction leaders denounce recent behavior in strongest terms.",
		},
		"support": {
			"The faction expresses support for recent actions.",
			"Faction leaders publicly endorse recent decisions.",
		},
		"neutrality": {
			"The faction maintains a neutral stance.",
			"Faction representatives decline to comment.",
		},
	}

	message := messages[responseType][rng.Intn(len(messages[responseType]))]

	return &FactionResponse{
		FactionID:        factionID,
		ResponseType:     responseType,
		ReputationChange: repChange,
		HostilityChange:  hostilityChange,
		TradeBonus:       tradeBonus,
		Message:          message,
	}
}

// GenerateEconomicEvent creates a market or economic disruption.
func GenerateEconomicEvent(seed int64, eventID, itemType string, severity Severity) *EconomicEvent {
	rng := rand.New(rand.NewSource(seed))

	priceModifier := 0.1 + (0.1 * float64(severity))
	if rng.Float64() < 0.5 {
		priceModifier = -priceModifier
	}

	supplyChange := -100 * int(severity)
	if priceModifier < 0 {
		supplyChange = -supplyChange
	}

	duration := 12 * time.Hour * time.Duration(severity)

	zoneCount := 1 + rng.Intn(int(severity))
	zones := make([]string, zoneCount)
	for i := 0; i < zoneCount; i++ {
		zones[i] = fmt.Sprintf("zone_%d", rng.Intn(100))
	}

	return &EconomicEvent{
		EventID:       eventID,
		ItemType:      itemType,
		PriceModifier: priceModifier,
		SupplyChange:  supplyChange,
		Duration:      duration,
		AffectedZones: zones,
	}
}

// GenerateWeatherDisaster creates a severe weather event.
func GenerateWeatherDisaster(seed int64, centerX, centerY float64, severity Severity) *WeatherDisaster {
	rng := rand.New(rand.NewSource(seed))

	disasterTypes := []string{
		"hurricane",
		"blizzard",
		"drought",
		"tornado",
		"flood",
		"heatwave",
	}

	disasterType := disasterTypes[rng.Intn(len(disasterTypes))]
	intensity := 0.5 + (0.125 * float64(severity))
	radius := 50.0 + (25.0 * float64(severity))
	duration := 1 * time.Hour * time.Duration(severity)
	damage := 10.0 * float64(severity)

	return &WeatherDisaster{
		DisasterType: disasterType,
		Intensity:    intensity,
		Radius:       radius,
		CenterX:      centerX,
		CenterY:      centerY,
		Duration:     duration,
		Damage:       damage,
	}
}

// PropagateEventCrossServer simulates event propagation to federated servers.
func PropagateEventCrossServer(event *WorldEvent, targetServers []string, delay time.Duration) []*WorldEvent {
	propagatedEvents := make([]*WorldEvent, len(targetServers))

	for i, serverID := range targetServers {
		propagatedEvent := &WorldEvent{
			ID:          fmt.Sprintf("%s_propagated_%s", event.ID, serverID),
			Type:        EventCrossServer,
			Trigger:     event.Trigger,
			Severity:    event.Severity,
			Title:       fmt.Sprintf("[Remote] %s", event.Title),
			Description: fmt.Sprintf("Event from %s: %s", event.ServerID, event.Description),
			Location:    event.Location,
			ServerID:    serverID,
			StartTime:   event.StartTime.Add(delay),
			Duration:    event.Duration,
			Impacts:     make([]Impact, len(event.Impacts)),
			ChainEvents: []string{},
			Permanent:   false,
		}

		for j, impact := range event.Impacts {
			propagatedEvent.Impacts[j] = Impact{
				Type:     impact.Type,
				Target:   impact.Target,
				Modifier: impact.Modifier * 0.5,
				Duration: impact.Duration,
			}
		}

		propagatedEvents[i] = propagatedEvent
	}

	return propagatedEvents
}

// CalculateEventFrequency determines how often events should spawn based on activity.
func CalculateEventFrequency(baseFrequency, activityLevel float64) float64 {
	if activityLevel < 0 {
		activityLevel = 0
	}
	if activityLevel > 1 {
		activityLevel = 1
	}

	return baseFrequency * (1.0 + activityLevel)
}

// ShouldSpawnEvent checks if an event should spawn based on time and frequency.
func ShouldSpawnEvent(lastEventTime time.Time, frequency float64) bool {
	elapsed := time.Since(lastEventTime)
	expectedInterval := time.Duration(60.0/frequency) * time.Minute

	return elapsed >= expectedInterval
}

// GetAffectedArea calculates the geographical area affected by an event.
// Returns the event's center coordinates and the radius of effect.
func GetAffectedArea(event *WorldEvent) (centerX, centerY, radius float64) {
	radius = 100.0 * float64(event.Severity)

	for _, impact := range event.Impacts {
		if impact.Type == ImpactWeather || impact.Type == ImpactTerrain {
			radius *= 1.5
			break
		}
	}

	return event.CenterX, event.CenterY, radius
}

// MergeEventImpacts combines impacts from multiple concurrent events.
func MergeEventImpacts(events []*WorldEvent) []Impact {
	impactMap := make(map[string]*Impact)

	for _, event := range events {
		for _, impact := range event.Impacts {
			key := fmt.Sprintf("%s_%s", impact.Type, impact.Target)

			if existing, ok := impactMap[key]; ok {
				existing.Modifier += impact.Modifier

				if impact.Duration > existing.Duration {
					existing.Duration = impact.Duration
				}
			} else {
				impactCopy := impact
				impactMap[key] = &impactCopy
			}
		}
	}

	merged := make([]Impact, 0, len(impactMap))
	for _, impact := range impactMap {
		merged = append(merged, *impact)
	}

	return merged
}
