// Package engine provides the event quest component for seasonal event quests.
// EventQuestComponent tracks time-limited quests that are only available during
// specific seasonal events, integrating with the existing quest system.
package engine

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/opd-ai/venture/pkg/procgen/quest"
	"github.com/sirupsen/logrus"
)

// EventQuestType defines the category of event quest.
type EventQuestType string

const (
	// EventQuestCollection requires gathering event-themed items.
	EventQuestCollection EventQuestType = "collection"
	// EventQuestExploration requires discovering event locations.
	EventQuestExploration EventQuestType = "exploration"
	// EventQuestBoss requires defeating an event-specific boss.
	EventQuestBoss EventQuestType = "boss"
)

// EventQuestDefinition defines an event-specific quest template.
type EventQuestDefinition struct {
	// ID is the unique identifier for this quest definition
	ID string `json:"id"`
	// EventID links this quest to a specific seasonal event
	EventID string `json:"event_id"`
	// QuestType categorizes the quest (collection, exploration, boss)
	QuestType EventQuestType `json:"quest_type"`
	// Name is the display name of the quest
	Name string `json:"name"`
	// Description is the quest description
	Description string `json:"description"`
	// Objectives are what the player must accomplish
	Objectives []quest.Objective `json:"objectives"`
	// Reward is what the player receives upon completion
	Reward quest.Reward `json:"reward"`
	// RequiredLevel is minimum level to accept the quest
	RequiredLevel int `json:"required_level"`
	// Seed is used for deterministic generation
	Seed int64 `json:"seed"`
}

// EventQuestInstance represents a running event quest for a player.
type EventQuestInstance struct {
	// Definition contains the quest configuration
	Definition EventQuestDefinition `json:"definition"`
	// Status is the current quest status
	Status quest.QuestStatus `json:"status"`
	// AcceptedAt is when the player accepted the quest
	AcceptedAt time.Time `json:"accepted_at"`
	// ExpiresAt is when the quest expires (when the event ends)
	ExpiresAt time.Time `json:"expires_at"`
	// Progress tracks objective completion (index -> current value)
	Progress map[int]int `json:"progress"`
}

// EventQuestComponent tracks event-specific quests for a player entity.
// It manages quest availability, progress, and expiration tied to seasonal events.
type EventQuestComponent struct {
	// ActiveQuests contains quests currently in progress
	ActiveQuests []EventQuestInstance `json:"active_quests"`
	// CompletedQuests contains quests finished during this event
	CompletedQuests []EventQuestInstance `json:"completed_quests"`
	// AvailableQuests contains quests the player can accept
	AvailableQuests []EventQuestDefinition `json:"available_quests"`
	// ExpiredQuests contains quests that expired without completion
	ExpiredQuests []EventQuestInstance `json:"expired_quests"`
	// MaxActiveEventQuests limits concurrent event quests
	MaxActiveEventQuests int `json:"max_active_event_quests"`
	// LastGenerationEventID tracks the last event quests were generated for
	LastGenerationEventID string `json:"last_generation_event_id"`
}

// NewEventQuestComponent creates a new event quest component.
func NewEventQuestComponent(maxActive int) *EventQuestComponent {
	logrus.WithFields(logrus.Fields{
		"component_type":          "event_quest",
		"max_active_event_quests": maxActive,
	}).Debug("Creating event quest component")

	if maxActive <= 0 {
		maxActive = 3
	}

	return &EventQuestComponent{
		ActiveQuests:         make([]EventQuestInstance, 0),
		CompletedQuests:      make([]EventQuestInstance, 0),
		AvailableQuests:      make([]EventQuestDefinition, 0),
		ExpiredQuests:        make([]EventQuestInstance, 0),
		MaxActiveEventQuests: maxActive,
	}
}

// Type returns the component type identifier.
func (c *EventQuestComponent) Type() string {
	return "event_quest"
}

// CanAcceptQuest checks if a new event quest can be accepted.
func (c *EventQuestComponent) CanAcceptQuest() bool {
	return len(c.ActiveQuests) < c.MaxActiveEventQuests
}

// AcceptQuest accepts an available event quest.
func (c *EventQuestComponent) AcceptQuest(questID string, expiresAt time.Time) bool {
	if !c.CanAcceptQuest() {
		logrus.WithFields(logrus.Fields{
			"component_type": "event_quest",
			"quest_id":       questID,
			"reason":         "max_active_reached",
		}).Debug("Cannot accept event quest")
		return false
	}

	// Find the quest in available quests
	var questDef *EventQuestDefinition
	var questIdx int
	for i, q := range c.AvailableQuests {
		if q.ID == questID {
			questDef = &c.AvailableQuests[i]
			questIdx = i
			break
		}
	}

	if questDef == nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "event_quest",
			"quest_id":       questID,
			"reason":         "not_found",
		}).Debug("Cannot accept event quest")
		return false
	}

	// Create the quest instance
	instance := EventQuestInstance{
		Definition: *questDef,
		Status:     quest.StatusActive,
		AcceptedAt: time.Now(),
		ExpiresAt:  expiresAt,
		Progress:   make(map[int]int),
	}

	// Initialize progress for all objectives
	for i := range questDef.Objectives {
		instance.Progress[i] = 0
	}

	// Add to active and remove from available
	c.ActiveQuests = append(c.ActiveQuests, instance)
	c.AvailableQuests = append(c.AvailableQuests[:questIdx], c.AvailableQuests[questIdx+1:]...)

	logrus.WithFields(logrus.Fields{
		"component_type": "event_quest",
		"quest_id":       questID,
		"event_id":       questDef.EventID,
		"expires_at":     expiresAt.Format(time.RFC3339),
	}).Info("Event quest accepted")

	return true
}

// UpdateProgress updates the progress of a quest objective.
func (c *EventQuestComponent) UpdateProgress(questID string, objectiveIndex, progress int) {
	for i := range c.ActiveQuests {
		if c.ActiveQuests[i].Definition.ID == questID {
			if objectiveIndex >= 0 && objectiveIndex < len(c.ActiveQuests[i].Definition.Objectives) {
				c.ActiveQuests[i].Progress[objectiveIndex] = progress

				logrus.WithFields(logrus.Fields{
					"component_type":  "event_quest",
					"quest_id":        questID,
					"objective_index": objectiveIndex,
					"progress":        progress,
				}).Debug("Event quest progress updated")
			}
			return
		}
	}
}

// IncrementProgress increments the progress of a quest objective.
func (c *EventQuestComponent) IncrementProgress(questID string, objectiveIndex, amount int) {
	for i := range c.ActiveQuests {
		if c.ActiveQuests[i].Definition.ID == questID {
			if objectiveIndex >= 0 && objectiveIndex < len(c.ActiveQuests[i].Definition.Objectives) {
				c.ActiveQuests[i].Progress[objectiveIndex] += amount
				req := c.ActiveQuests[i].Definition.Objectives[objectiveIndex].Required
				if c.ActiveQuests[i].Progress[objectiveIndex] > req {
					c.ActiveQuests[i].Progress[objectiveIndex] = req
				}
			}
			return
		}
	}
}

// IsQuestComplete checks if all objectives of a quest are complete.
func (c *EventQuestComponent) IsQuestComplete(questID string) bool {
	for _, q := range c.ActiveQuests {
		if q.Definition.ID == questID {
			for i, obj := range q.Definition.Objectives {
				if q.Progress[i] < obj.Required {
					return false
				}
			}
			return true
		}
	}
	return false
}

// CompleteQuest marks a quest as completed and moves it to completed list.
func (c *EventQuestComponent) CompleteQuest(questID string) bool {
	for i, q := range c.ActiveQuests {
		if q.Definition.ID == questID {
			q.Status = quest.StatusComplete

			// Remove from active
			c.ActiveQuests = append(c.ActiveQuests[:i], c.ActiveQuests[i+1:]...)

			// Add to completed
			c.CompletedQuests = append(c.CompletedQuests, q)

			logrus.WithFields(logrus.Fields{
				"component_type": "event_quest",
				"quest_id":       questID,
				"event_id":       q.Definition.EventID,
			}).Info("Event quest completed")

			return true
		}
	}
	return false
}

// ExpireQuest marks a quest as expired and moves it to expired list.
func (c *EventQuestComponent) ExpireQuest(questID string) bool {
	for i, q := range c.ActiveQuests {
		if q.Definition.ID == questID {
			q.Status = quest.StatusFailed

			// Remove from active
			c.ActiveQuests = append(c.ActiveQuests[:i], c.ActiveQuests[i+1:]...)

			// Add to expired
			c.ExpiredQuests = append(c.ExpiredQuests, q)

			logrus.WithFields(logrus.Fields{
				"component_type": "event_quest",
				"quest_id":       questID,
				"event_id":       q.Definition.EventID,
			}).Info("Event quest expired")

			return true
		}
	}
	return false
}

// GetActiveQuest returns an active event quest by ID.
func (c *EventQuestComponent) GetActiveQuest(questID string) *EventQuestInstance {
	for i := range c.ActiveQuests {
		if c.ActiveQuests[i].Definition.ID == questID {
			return &c.ActiveQuests[i]
		}
	}
	return nil
}

// GetActiveQuestsForEvent returns all active quests for a specific event.
func (c *EventQuestComponent) GetActiveQuestsForEvent(eventID string) []EventQuestInstance {
	result := make([]EventQuestInstance, 0)
	for _, q := range c.ActiveQuests {
		if q.Definition.EventID == eventID {
			result = append(result, q)
		}
	}
	return result
}

// GetAvailableQuestsForEvent returns all available quests for a specific event.
func (c *EventQuestComponent) GetAvailableQuestsForEvent(eventID string) []EventQuestDefinition {
	result := make([]EventQuestDefinition, 0)
	for _, q := range c.AvailableQuests {
		if q.EventID == eventID {
			result = append(result, q)
		}
	}
	return result
}

// ClearEventQuests removes all quests associated with a specific event.
// Used when an event ends to clean up state.
func (c *EventQuestComponent) ClearEventQuests(eventID string) {
	// Expire active quests for this event
	for i := len(c.ActiveQuests) - 1; i >= 0; i-- {
		if c.ActiveQuests[i].Definition.EventID == eventID {
			c.ActiveQuests[i].Status = quest.StatusFailed
			c.ExpiredQuests = append(c.ExpiredQuests, c.ActiveQuests[i])
			c.ActiveQuests = append(c.ActiveQuests[:i], c.ActiveQuests[i+1:]...)
		}
	}

	// Remove available quests for this event
	for i := len(c.AvailableQuests) - 1; i >= 0; i-- {
		if c.AvailableQuests[i].EventID == eventID {
			c.AvailableQuests = append(c.AvailableQuests[:i], c.AvailableQuests[i+1:]...)
		}
	}

	logrus.WithFields(logrus.Fields{
		"component_type": "event_quest",
		"event_id":       eventID,
	}).Info("Cleared event quests for ended event")
}

// Serialize encodes the component to bytes for persistence.
func (c *EventQuestComponent) Serialize() ([]byte, error) {
	logrus.WithFields(logrus.Fields{
		"component_type": "event_quest",
		"active_quests":  len(c.ActiveQuests),
		"available":      len(c.AvailableQuests),
		"completed":      len(c.CompletedQuests),
	}).Debug("Serializing event quest component")

	data, err := json.Marshal(c)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "event_quest",
			"error":          err.Error(),
		}).Error("Failed to serialize event quest component")
		return nil, err
	}
	return data, nil
}

// Deserialize decodes the component from bytes.
func (c *EventQuestComponent) Deserialize(data []byte) error {
	logrus.WithFields(logrus.Fields{
		"component_type": "event_quest",
		"bytes":          len(data),
	}).Debug("Deserializing event quest component")

	if err := json.Unmarshal(data, c); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "event_quest",
			"error":          err.Error(),
		}).Error("Failed to deserialize event quest component")
		return err
	}
	return nil
}

// GenerateEventQuests generates quests for a seasonal event.
// Returns 3 quests per event (one of each type).
func GenerateEventQuests(event EventInstance, seed int64) []EventQuestDefinition {
	rng := rand.New(rand.NewSource(seed))

	quests := make([]EventQuestDefinition, 0, 3)

	// Generate one quest of each type
	quests = append(quests, generateCollectionQuest(rng, event))
	quests = append(quests, generateExplorationQuest(rng, event))
	quests = append(quests, generateBossQuest(rng, event))

	logrus.WithFields(logrus.Fields{
		"component_type": "event_quest",
		"event_id":       event.Definition.ID,
		"event_name":     event.Definition.Name,
		"quests_count":   len(quests),
		"seed":           seed,
	}).Debug("Generated event quests")

	return quests
}

// generateCollectionQuest creates a collection-type event quest.
func generateCollectionQuest(rng *rand.Rand, event EventInstance) EventQuestDefinition {
	items := getThemeItems(event.Definition.Theme)
	itemName := items[rng.Intn(len(items))]
	required := 5 + rng.Intn(10) // 5-14 items

	return EventQuestDefinition{
		ID:        event.Definition.ID + "_collection",
		EventID:   event.Definition.ID,
		QuestType: EventQuestCollection,
		Name:      "Gather " + itemName,
		Description: generateCollectionDescription(
			event.Definition.Name, itemName, required,
		),
		Objectives: []quest.Objective{
			{
				Description: "Collect " + itemName,
				Target:      itemName,
				Required:    required,
				Current:     0,
			},
		},
		Reward: quest.Reward{
			XP:   100 + rng.Intn(100),
			Gold: 50 + rng.Intn(50),
		},
		RequiredLevel: 1,
		Seed:          rng.Int63(),
	}
}

// generateExplorationQuest creates an exploration-type event quest.
func generateExplorationQuest(rng *rand.Rand, event EventInstance) EventQuestDefinition {
	locations := getThemeLocations(event.Definition.Theme)
	locationName := locations[rng.Intn(len(locations))]

	return EventQuestDefinition{
		ID:        event.Definition.ID + "_exploration",
		EventID:   event.Definition.ID,
		QuestType: EventQuestExploration,
		Name:      "Discover " + locationName,
		Description: generateExplorationDescription(
			event.Definition.Name, locationName,
		),
		Objectives: []quest.Objective{
			{
				Description: "Find " + locationName,
				Target:      locationName,
				Required:    1,
				Current:     0,
			},
		},
		Reward: quest.Reward{
			XP:   150 + rng.Intn(100),
			Gold: 75 + rng.Intn(50),
		},
		RequiredLevel: 1,
		Seed:          rng.Int63(),
	}
}

// generateBossQuest creates a boss-type event quest.
func generateBossQuest(rng *rand.Rand, event EventInstance) EventQuestDefinition {
	bosses := getThemeBosses(event.Definition.Theme)
	bossName := bosses[rng.Intn(len(bosses))]

	return EventQuestDefinition{
		ID:        event.Definition.ID + "_boss",
		EventID:   event.Definition.ID,
		QuestType: EventQuestBoss,
		Name:      "Defeat " + bossName,
		Description: generateBossDescription(
			event.Definition.Name, bossName,
		),
		Objectives: []quest.Objective{
			{
				Description: "Slay " + bossName,
				Target:      bossName,
				Required:    1,
				Current:     0,
			},
		},
		Reward: quest.Reward{
			XP:          300 + rng.Intn(200),
			Gold:        150 + rng.Intn(100),
			SkillPoints: 1,
		},
		RequiredLevel: 5,
		Seed:          rng.Int63(),
	}
}

// getThemeItems returns collectible items themed for the event.
func getThemeItems(theme EventTheme) []string {
	switch theme {
	case EventThemeSpring:
		return []string{"Spring Blossom", "Fresh Dewdrop", "Rainbow Feather", "Sprouting Seed"}
	case EventThemeSummer:
		return []string{"Sunfire Crystal", "Golden Honey", "Radiant Shell", "Solar Essence"}
	case EventThemeAutumn:
		return []string{"Harvest Apple", "Amber Leaf", "Twilight Berry", "Cornucopia Seed"}
	case EventThemeWinter:
		return []string{"Frost Shard", "Starlight Crystal", "Evergreen Pinecone", "Hearthstone"}
	default:
		return []string{"Mysterious Token", "Strange Artifact"}
	}
}

// getThemeLocations returns exploration locations themed for the event.
func getThemeLocations(theme EventTheme) []string {
	switch theme {
	case EventThemeSpring:
		return []string{"the Awakening Grove", "the Blossom Meadow", "the Rainbow Bridge"}
	case EventThemeSummer:
		return []string{"the Solstice Shrine", "the Sun Temple", "the Golden Beach"}
	case EventThemeAutumn:
		return []string{"the Harvest Altar", "the Twilight Forest", "the Amber Valley"}
	case EventThemeWinter:
		return []string{"the Frost Palace", "the Starlight Cavern", "the Hearthfire Hall"}
	default:
		return []string{"the Hidden Sanctum", "the Ancient Ruins"}
	}
}

// getThemeBosses returns boss enemies themed for the event.
func getThemeBosses(theme EventTheme) []string {
	switch theme {
	case EventThemeSpring:
		return []string{"the Bloom Guardian", "the Spring Wraith", "Florius the Verdant"}
	case EventThemeSummer:
		return []string{"the Sun Elemental", "the Heat Phantom", "Solaris the Radiant"}
	case EventThemeAutumn:
		return []string{"the Harvest Golem", "the Twilight Specter", "Croptus the Eternal"}
	case EventThemeWinter:
		return []string{"the Frost Giant", "the Blizzard Dragon", "Glacius the Frozen"}
	default:
		return []string{"the Event Champion", "the Seasonal Guardian"}
	}
}

// generateCollectionDescription creates a description for collection quests.
func generateCollectionDescription(eventName, itemName string, required int) string {
	return "During the " + eventName + ", special " + itemName + " can be found throughout the land. " +
		"Collect " + string(rune('0'+required/10)) + string(rune('0'+required%10)) + " of them to earn festival rewards."
}

// generateExplorationDescription creates a description for exploration quests.
func generateExplorationDescription(eventName, locationName string) string {
	return "Legend says that " + locationName + " only appears during the " + eventName + ". " +
		"Find this mystical place to prove your adventuring prowess."
}

// generateBossDescription creates a description for boss quests.
func generateBossDescription(eventName, bossName string) string {
	return "The " + eventName + " has awakened " + bossName + " from slumber. " +
		"Only the bravest heroes can defeat this powerful foe and save the celebration."
}
