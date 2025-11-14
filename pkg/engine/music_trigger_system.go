package engine

import (
	"time"

	"github.com/opd-ai/venture/pkg/audio"
)

// MusicTriggerSystem manages music context changes based on gameplay events.
// This system monitors game state and triggers adaptive music responses.
type MusicTriggerSystem struct {
	world           *World
	musicManager    audio.AdaptiveMusicSystem
	eventQueue      []MusicTriggerEvent
	updateInterval  float64
	timeSinceUpdate float64
}

// NewMusicTriggerSystem creates a new music trigger system.
func NewMusicTriggerSystem(world *World, musicManager audio.AdaptiveMusicSystem) *MusicTriggerSystem {
	return &MusicTriggerSystem{
		world:          world,
		musicManager:   musicManager,
		eventQueue:     make([]MusicTriggerEvent, 0),
		updateInterval: 0.5, // Update music context every 0.5 seconds
	}
}

// Update processes music triggers and updates the music system.
func (mts *MusicTriggerSystem) Update(deltaTime float64) {
	if mts.world == nil || mts.musicManager == nil {
		return
	}

	// Process queued events immediately and track if any were processed
	hadEvents := len(mts.eventQueue) > 0
	mts.processEventQueue()

	mts.timeSinceUpdate += deltaTime

	// Process at regular intervals to avoid excessive updates,
	// but force an update if we just processed events
	shouldUpdate := mts.timeSinceUpdate >= mts.updateInterval || hadEvents
	
	if !shouldUpdate {
		// Still update music manager even if not checking entities
		mts.musicManager.Update(deltaTime)
		return
	}

	mts.timeSinceUpdate = 0.0

	// Get all entities with music trigger components
	entities := mts.world.GetEntitiesWith("music_trigger")

	for _, entity := range entities {
		comp, ok := entity.GetComponent("music_trigger")
		if !ok {
			continue
		}

		triggerComp, ok := comp.(*MusicTriggerComponent)
		if !ok {
			continue
		}

		// Update pending transitions
		triggerComp.UpdatePendingTransition(deltaTime)

		// Apply current context to music manager
		ctx := triggerComp.GetMusicContext()
		err := mts.musicManager.SetContext(ctx)
		if err != nil {
			// Log error but continue
			continue
		}
	}

	// Update music manager
	mts.musicManager.Update(deltaTime)
}

// QueueEvent adds a music trigger event to the queue.
func (mts *MusicTriggerSystem) QueueEvent(event MusicTriggerEvent) {
	mts.eventQueue = append(mts.eventQueue, event)
}

// processEventQueue handles queued music trigger events.
func (mts *MusicTriggerSystem) processEventQueue() {
	for _, event := range mts.eventQueue {
		mts.handleEvent(event)
	}
	mts.eventQueue = mts.eventQueue[:0] // Clear queue
}

// handleEvent processes a single music trigger event.
func (mts *MusicTriggerSystem) handleEvent(event MusicTriggerEvent) {
	entity, ok := mts.world.GetEntity(event.EntityID)
	if !ok {
		return
	}

	comp, ok := entity.GetComponent("music_trigger")
	if !ok {
		return
	}

	triggerComp, ok := comp.(*MusicTriggerComponent)
	if !ok {
		return
	}

	switch event.Type {
	case TriggerCombatStart:
		triggerComp.TriggerCombat(true)

	case TriggerCombatEnd:
		triggerComp.TriggerCombat(false)

	case TriggerBossAppear:
		triggerComp.TriggerBoss(true)

	case TriggerBossDefeated:
		triggerComp.TriggerBoss(false)
		triggerComp.TriggerQuestCompletion() // Victory music

	case TriggerQuestComplete:
		triggerComp.TriggerQuestCompletion()

	case TriggerExplorationMilestone:
		newArea := true
		if val, ok := event.Data["new_area"]; ok {
			if b, ok := val.(bool); ok {
				newArea = b
			}
		}
		triggerComp.TriggerExploration(newArea)

	case TriggerReputationChange:
		if tier, ok := event.Data["tier"].(string); ok {
			triggerComp.TriggerReputationChange(tier)
		}
	}
}

// OnCombatStart triggers combat music.
func (mts *MusicTriggerSystem) OnCombatStart(entityID uint64) {
	mts.QueueEvent(MusicTriggerEvent{
		Type:      TriggerCombatStart,
		EntityID:  entityID,
		Timestamp: time.Now(),
	})
}

// OnCombatEnd triggers post-combat music.
func (mts *MusicTriggerSystem) OnCombatEnd(entityID uint64) {
	mts.QueueEvent(MusicTriggerEvent{
		Type:      TriggerCombatEnd,
		EntityID:  entityID,
		Timestamp: time.Now(),
	})
}

// OnBossAppear triggers boss battle music.
func (mts *MusicTriggerSystem) OnBossAppear(entityID uint64) {
	mts.QueueEvent(MusicTriggerEvent{
		Type:      TriggerBossAppear,
		EntityID:  entityID,
		Timestamp: time.Now(),
	})
}

// OnBossDefeated triggers victory music for boss defeat.
func (mts *MusicTriggerSystem) OnBossDefeated(entityID uint64) {
	mts.QueueEvent(MusicTriggerEvent{
		Type:      TriggerBossDefeated,
		EntityID:  entityID,
		Timestamp: time.Now(),
	})
}

// OnQuestComplete triggers quest completion music.
func (mts *MusicTriggerSystem) OnQuestComplete(entityID uint64) {
	mts.QueueEvent(MusicTriggerEvent{
		Type:      TriggerQuestComplete,
		EntityID:  entityID,
		Timestamp: time.Now(),
	})
}

// OnExplorationMilestone triggers exploration music update.
func (mts *MusicTriggerSystem) OnExplorationMilestone(entityID uint64, newArea bool) {
	mts.QueueEvent(MusicTriggerEvent{
		Type:      TriggerExplorationMilestone,
		EntityID:  entityID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"new_area": newArea,
		},
	})
}

// OnReputationChange triggers music based on reputation tier.
func (mts *MusicTriggerSystem) OnReputationChange(entityID uint64, tier string) {
	mts.QueueEvent(MusicTriggerEvent{
		Type:      TriggerReputationChange,
		EntityID:  entityID,
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"tier": tier,
		},
	})
}

// SetMusicManager updates the music manager reference.
func (mts *MusicTriggerSystem) SetMusicManager(manager audio.AdaptiveMusicSystem) {
	mts.musicManager = manager
}

// GetEventQueueLength returns the number of pending events.
func (mts *MusicTriggerSystem) GetEventQueueLength() int {
	return len(mts.eventQueue)
}
