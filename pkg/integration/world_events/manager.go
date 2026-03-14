package world_events

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/procgen"
)

// EventManager manages world event generation and chains.
type EventManager struct {
	seed          int64
	rng           *rand.Rand
	config        EventManagerConfig
	activeEvents  map[string]*WorldEvent
	eventChains   map[string]*EventChain
	mu            sync.RWMutex
	eventCounter  int
	lastEventTime time.Time
}

// NewEventManager creates a new event manager with the given seed.
func NewEventManager(seed int64) *EventManager {
	return &EventManager{
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		config:        NewDefaultEventManagerConfig(),
		activeEvents:  make(map[string]*WorldEvent),
		eventChains:   make(map[string]*EventChain),
		lastEventTime: now(),
	}
}

// NewEventManagerWithConfig creates a new event manager with custom configuration.
func NewEventManagerWithConfig(seed int64, config EventManagerConfig) *EventManager {
	return &EventManager{
		seed:          seed,
		rng:           rand.New(rand.NewSource(seed)),
		config:        config,
		activeEvents:  make(map[string]*WorldEvent),
		eventChains:   make(map[string]*EventChain),
		lastEventTime: now(),
	}
}

// GenerateEvent creates a new world event from trigger parameters.
func (m *EventManager) GenerateEvent(trigger TriggerType, params TriggerParams) (*WorldEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("invalid trigger params: %w", err)
	}

	if len(m.activeEvents) >= m.config.MaxActiveEvents {
		return nil, fmt.Errorf("max active events (%d) reached", m.config.MaxActiveEvents)
	}

	m.eventCounter++
	eventID := fmt.Sprintf("event_%s_%d", params.ServerID, m.eventCounter)

	var eventType EventType
	switch trigger {
	case TriggerGuildWar, TriggerGuildPeace:
		eventType = EventGuildWarfare
	case TriggerPoliticalShift:
		eventType = EventFactionResponse
	case TriggerTradeVolume:
		eventType = EventEconomic
	case TriggerWeatherChange:
		eventType = EventWeatherDisaster
	case TriggerPlayerChoice:
		eventType = EventChained
	default:
		return nil, fmt.Errorf("unknown trigger type: %s", trigger)
	}

	event := &WorldEvent{
		ID:          eventID,
		Type:        eventType,
		Trigger:     trigger,
		Severity:    params.Severity,
		Location:    params.Location,
		ServerID:    params.ServerID,
		StartTime:   now().Add(m.getResponseDelay()),
		Duration:    m.getDuration(params.Severity),
		Impacts:     m.generateImpacts(eventType, params),
		ChainEvents: []string{},
		Permanent:   false,
		CenterX:     params.CenterX,
		CenterY:     params.CenterY,
	}

	event.Title, event.Description = m.generateText(event, params)

	if m.rng.Float64() < m.config.ChainProbability {
		chainID := fmt.Sprintf("chain_%s", eventID)
		chain := &EventChain{
			ChainID:       chainID,
			Events:        []string{eventID},
			CurrentIndex:  0,
			BranchChoices: make(map[string][]string),
		}
		m.eventChains[chainID] = chain
		event.ChainEvents = m.generateChainEvents(eventID, params)
	}

	m.activeEvents[eventID] = event
	m.lastEventTime = now()

	return event, nil
}

// GetEvent retrieves an event by ID.
func (m *EventManager) GetEvent(eventID string) (*WorldEvent, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	event, ok := m.activeEvents[eventID]
	return event, ok
}

// GetActiveEvents returns all currently active events.
func (m *EventManager) GetActiveEvents(currentTime time.Time) []*WorldEvent {
	m.mu.RLock()
	defer m.mu.RUnlock()

	active := make([]*WorldEvent, 0, len(m.activeEvents))
	for _, event := range m.activeEvents {
		if event.IsActive(currentTime) {
			active = append(active, event)
		}
	}
	return active
}

// GetEventChain retrieves an event chain by event ID.
func (m *EventManager) GetEventChain(eventID string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, chain := range m.eventChains {
		for _, id := range chain.Events {
			if id == eventID {
				return chain.Events
			}
		}
	}
	return []string{eventID}
}

// CleanupExpiredEvents removes events that are no longer active.
func (m *EventManager) CleanupExpiredEvents(currentTime time.Time) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	removed := 0
	for id, event := range m.activeEvents {
		if !event.IsActive(currentTime) && !event.Permanent {
			delete(m.activeEvents, id)
			removed++
		}
	}
	return removed
}

// Update processes event chains and spawns follow-up events.
func (m *EventManager) Update(deltaTime float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentTime := now()
	for chainID, chain := range m.eventChains {
		if chain.CurrentIndex >= len(chain.Events) {
			continue
		}

		currentEventID := chain.Events[chain.CurrentIndex]
		currentEvent, ok := m.activeEvents[currentEventID]
		if !ok {
			continue
		}

		if !currentEvent.IsActive(currentTime) {
			chain.CurrentIndex++
			if chain.CurrentIndex < len(chain.Events) {
				nextEventID := chain.Events[chain.CurrentIndex]
				if nextEvent, exists := m.activeEvents[nextEventID]; exists {
					nextEvent.StartTime = currentTime
				}
			} else {
				delete(m.eventChains, chainID)
			}
		}
	}
}

// GetStats returns event manager statistics.
func (m *EventManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	typeCounts := make(map[EventType]int)
	for _, event := range m.activeEvents {
		typeCounts[event.Type]++
	}

	return map[string]interface{}{
		"active_events":   len(m.activeEvents),
		"event_chains":    len(m.eventChains),
		"type_counts":     typeCounts,
		"total_generated": m.eventCounter,
	}
}

func (m *EventManager) getResponseDelay() time.Duration {
	min := m.config.ResponseTimeMin.Milliseconds()
	max := m.config.ResponseTimeMax.Milliseconds()

	if min >= max {
		return m.config.ResponseTimeMin
	}

	delay := min + m.rng.Int63n(max-min)
	return time.Duration(delay) * time.Millisecond
}

func (m *EventManager) getDuration(severity Severity) time.Duration {
	baseDuration := 30 * time.Minute
	return baseDuration * time.Duration(severity)
}

func (m *EventManager) generateImpacts(eventType EventType, params TriggerParams) []Impact {
	impacts := make([]Impact, 0, 3)
	severity := float64(params.Severity)

	switch eventType {
	case EventGuildWarfare:
		impacts = append(impacts, Impact{
			Type:     ImpactNPCReputation,
			Target:   params.GuildID,
			Modifier: -0.1 * severity,
			Duration: 24 * time.Hour * time.Duration(severity),
		})
		impacts = append(impacts, Impact{
			Type:     ImpactSpawnRate,
			Target:   params.Location,
			Modifier: 1.0 + (0.2 * severity),
			Duration: 2 * time.Hour * time.Duration(severity),
		})

	case EventFactionResponse:
		impacts = append(impacts, Impact{
			Type:     ImpactNPCReputation,
			Target:   params.FactionID,
			Modifier: -0.05 * severity,
			Duration: 48 * time.Hour,
		})

	case EventEconomic:
		priceChange := 0.1 + (0.1 * severity)
		if m.rng.Float64() < 0.5 {
			priceChange = -priceChange
		}
		impacts = append(impacts, Impact{
			Type:     ImpactPriceChange,
			Target:   params.ItemType,
			Modifier: priceChange,
			Duration: 12 * time.Hour * time.Duration(severity),
		})

	case EventWeatherDisaster:
		impacts = append(impacts, Impact{
			Type:     ImpactWeather,
			Target:   params.Location,
			Modifier: 0.5 + (0.5 * severity),
			Duration: 1 * time.Hour * time.Duration(severity),
		})
		impacts = append(impacts, Impact{
			Type:     ImpactTerrain,
			Target:   params.Location,
			Modifier: -0.1 * severity,
			Duration: 0,
		})

	case EventChained:
		// Chained events inherit characteristics from their parent event
		// and add narrative/reputation impacts
		impacts = append(impacts, Impact{
			Type:     ImpactNPCReputation,
			Target:   params.PlayerID,
			Modifier: -0.02 * severity,
			Duration: 6 * time.Hour * time.Duration(severity),
		})
		// Chained events can also affect spawn rates in the area
		if params.Location != "" {
			impacts = append(impacts, Impact{
				Type:     ImpactSpawnRate,
				Target:   params.Location,
				Modifier: 1.0 + (0.1 * severity),
				Duration: 1 * time.Hour * time.Duration(severity),
			})
		}
	}

	return impacts
}

func (m *EventManager) generateText(event *WorldEvent, params TriggerParams) (string, string) {
	seedGen := procgen.NewSeedGenerator(m.seed + int64(m.eventCounter))
	textSeed := seedGen.GetSeed("event_text", int(event.Severity))
	textRNG := rand.New(rand.NewSource(textSeed))

	titles := m.getTitlesForType(event.Type, event.Severity)
	descriptions := m.getDescriptionsForType(event.Type, event.Severity)

	title := titles[textRNG.Intn(len(titles))]
	description := descriptions[textRNG.Intn(len(descriptions))]

	return title, description
}

func (m *EventManager) getTitlesForType(eventType EventType, severity Severity) []string {
	switch eventType {
	case EventGuildWarfare:
		if severity >= SeverityMajor {
			return []string{"Total War", "Regional Conflict", "Guild Invasion"}
		}
		return []string{"Skirmish Reported", "Border Dispute", "Minor Conflict"}

	case EventFactionResponse:
		if severity >= SeverityMajor {
			return []string{"Faction Uprising", "Political Revolution", "Mass Protest"}
		}
		return []string{"Diplomatic Tension", "Public Discontent", "Faction Concerns"}

	case EventEconomic:
		if severity >= SeverityMajor {
			return []string{"Market Crash", "Economic Crisis", "Trade Collapse"}
		}
		return []string{"Price Fluctuation", "Supply Issue", "Market Adjustment"}

	case EventWeatherDisaster:
		if severity >= SeverityMajor {
			return []string{"Catastrophic Storm", "Natural Disaster", "Weather Emergency"}
		}
		return []string{"Severe Weather", "Storm Warning", "Weather Alert"}

	default:
		return []string{"World Event"}
	}
}

func (m *EventManager) getDescriptionsForType(eventType EventType, severity Severity) []string {
	switch eventType {
	case EventGuildWarfare:
		return []string{
			"Guild forces clash in the region, causing widespread unrest.",
			"Military operations disrupt trade and civilian activity.",
			"Conflict escalates as guilds mobilize their forces.",
		}

	case EventFactionResponse:
		return []string{
			"NPC factions express displeasure with recent developments.",
			"Political tensions rise as factions take defensive positions.",
			"Factional leadership issues statements condemning recent actions.",
		}

	case EventEconomic:
		return []string{
			"Market forces shift dramatically, affecting regional prices.",
			"Supply chains disrupted, causing economic instability.",
			"Traders report unusual market conditions.",
		}

	case EventWeatherDisaster:
		return []string{
			"Severe weather conditions threaten the region.",
			"Natural forces unleash devastating power on the area.",
			"Weather patterns shift violently, endangering travelers.",
		}

	default:
		return []string{"An event occurs in the world."}
	}
}

func (m *EventManager) generateChainEvents(eventID string, params TriggerParams) []string {
	chainLength := 2 + m.rng.Intn(3)
	chainEvents := make([]string, chainLength)

	for i := 0; i < chainLength; i++ {
		m.eventCounter++
		chainEvents[i] = fmt.Sprintf("event_%s_%d_chain", params.ServerID, m.eventCounter)

		followUpSeverity := params.Severity
		if m.rng.Float64() < 0.3 {
			if followUpSeverity < SeverityCritical {
				followUpSeverity++
			}
		}

		followUpParams := TriggerParams{
			TriggerType: TriggerPlayerChoice,
			Severity:    followUpSeverity,
			Location:    params.Location,
			ServerID:    params.ServerID,
			GuildID:     params.GuildID,
			FactionID:   params.FactionID,
			ItemType:    params.ItemType,
			PlayerID:    params.PlayerID,
			ChoiceID:    eventID,
			CenterX:     params.CenterX,
			CenterY:     params.CenterY,
		}

		followUpEvent := &WorldEvent{
			ID:          chainEvents[i],
			Type:        EventChained,
			Trigger:     TriggerPlayerChoice,
			Severity:    followUpSeverity,
			Location:    params.Location,
			ServerID:    params.ServerID,
			StartTime:   now().Add(time.Duration(i+1) * time.Hour),
			Duration:    m.getDuration(followUpSeverity),
			Impacts:     m.generateImpacts(EventChained, followUpParams),
			ChainEvents: []string{},
			Permanent:   false,
			CenterX:     params.CenterX,
			CenterY:     params.CenterY,
		}

		followUpEvent.Title, followUpEvent.Description = m.generateText(followUpEvent, followUpParams)
		m.activeEvents[chainEvents[i]] = followUpEvent
	}

	if existingChain, ok := m.eventChains[fmt.Sprintf("chain_%s", eventID)]; ok {
		existingChain.Events = append(existingChain.Events, chainEvents...)
	}

	return chainEvents
}
