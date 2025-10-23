// Package engine provides quest tracking components.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/quest"
)

// QuestStatus represents the current state of a quest.
type QuestStatus int

const (
	QuestStatusActive QuestStatus = iota
	QuestStatusCompleted
	QuestStatusFailed
)

// TrackedQuest wraps a generated quest with runtime tracking data.
type TrackedQuest struct {
	Quest     *quest.Quest
	Status    QuestStatus
	StartTime int64 // Unix timestamp when quest was accepted
	EndTime   int64 // Unix timestamp when quest was completed/failed
}

// QuestTrackerComponent manages an entity's active and completed quests.
type QuestTrackerComponent struct {
	// ActiveQuests are quests currently in progress
	ActiveQuests []*TrackedQuest

	// CompletedQuests are quests that have been finished
	CompletedQuests []*TrackedQuest

	// FailedQuests are quests that were failed
	FailedQuests []*TrackedQuest

	// MaxActiveQuests is the maximum number of concurrent active quests
	MaxActiveQuests int
}

// Type returns the component type identifier.
func (q *QuestTrackerComponent) Type() string {
	return "questtracker"
}

// NewQuestTrackerComponent creates a new quest tracker.
func NewQuestTrackerComponent(maxActive int) *QuestTrackerComponent {
	return &QuestTrackerComponent{
		ActiveQuests:    make([]*TrackedQuest, 0),
		CompletedQuests: make([]*TrackedQuest, 0),
		FailedQuests:    make([]*TrackedQuest, 0),
		MaxActiveQuests: maxActive,
	}
}

// CanAcceptQuest checks if a new quest can be accepted.
func (q *QuestTrackerComponent) CanAcceptQuest() bool {
	return len(q.ActiveQuests) < q.MaxActiveQuests
}

// AcceptQuest adds a quest to the active list.
func (q *QuestTrackerComponent) AcceptQuest(qst *quest.Quest, startTime int64) bool {
	if !q.CanAcceptQuest() {
		return false
	}

	// Make a copy of the quest to avoid modifying the original
	questCopy := *qst
	questCopy.Status = quest.StatusActive

	tracked := &TrackedQuest{
		Quest:     &questCopy,
		Status:    QuestStatusActive,
		StartTime: startTime,
	}

	q.ActiveQuests = append(q.ActiveQuests, tracked)
	return true
}

// UpdateProgress updates the progress of a quest objective.
func (q *QuestTrackerComponent) UpdateProgress(questID string, objectiveIndex int, progress int) {
	for _, tracked := range q.ActiveQuests {
		if tracked.Quest.ID == questID {
			if objectiveIndex >= 0 && objectiveIndex < len(tracked.Quest.Objectives) {
				tracked.Quest.Objectives[objectiveIndex].Current = progress
			}
			return
		}
	}
}

// IncrementProgress increments the progress of a quest objective.
func (q *QuestTrackerComponent) IncrementProgress(questID string, objectiveIndex int, amount int) {
	for _, tracked := range q.ActiveQuests {
		if tracked.Quest.ID == questID {
			if objectiveIndex >= 0 && objectiveIndex < len(tracked.Quest.Objectives) {
				tracked.Quest.Objectives[objectiveIndex].Current += amount
				if tracked.Quest.Objectives[objectiveIndex].Current > tracked.Quest.Objectives[objectiveIndex].Required {
					tracked.Quest.Objectives[objectiveIndex].Current = tracked.Quest.Objectives[objectiveIndex].Required
				}
			}
			return
		}
	}
}

// IsObjectiveComplete checks if an objective has reached its target.
func (q *QuestTrackerComponent) IsObjectiveComplete(questID string, objectiveIndex int) bool {
	for _, tracked := range q.ActiveQuests {
		if tracked.Quest.ID == questID {
			if objectiveIndex >= 0 && objectiveIndex < len(tracked.Quest.Objectives) {
				return tracked.Quest.Objectives[objectiveIndex].IsComplete()
			}
		}
	}
	return false
}

// IsQuestComplete checks if all objectives of a quest are complete.
func (q *QuestTrackerComponent) IsQuestComplete(questID string) bool {
	for _, tracked := range q.ActiveQuests {
		if tracked.Quest.ID == questID {
			return tracked.Quest.IsComplete()
		}
	}
	return false
}

// CompleteQuest marks a quest as completed and moves it to completed list.
func (q *QuestTrackerComponent) CompleteQuest(questID string, endTime int64) bool {
	for i, tracked := range q.ActiveQuests {
		if tracked.Quest.ID == questID {
			tracked.Status = QuestStatusCompleted
			tracked.EndTime = endTime

			// Remove from active
			q.ActiveQuests = append(q.ActiveQuests[:i], q.ActiveQuests[i+1:]...)

			// Add to completed
			q.CompletedQuests = append(q.CompletedQuests, tracked)
			return true
		}
	}
	return false
}

// FailQuest marks a quest as failed and moves it to failed list.
func (q *QuestTrackerComponent) FailQuest(questID string, endTime int64) bool {
	for i, tracked := range q.ActiveQuests {
		if tracked.Quest.ID == questID {
			tracked.Status = QuestStatusFailed
			tracked.EndTime = endTime

			// Remove from active
			q.ActiveQuests = append(q.ActiveQuests[:i], q.ActiveQuests[i+1:]...)

			// Add to failed
			q.FailedQuests = append(q.FailedQuests, tracked)
			return true
		}
	}
	return false
}

// GetActiveQuest returns an active quest by ID.
func (q *QuestTrackerComponent) GetActiveQuest(questID string) *TrackedQuest {
	for _, tracked := range q.ActiveQuests {
		if tracked.Quest.ID == questID {
			return tracked
		}
	}
	return nil
}

// GetQuestProgress returns the progress of a specific objective.
func (q *QuestTrackerComponent) GetQuestProgress(questID string, objectiveIndex int) int {
	for _, tracked := range q.ActiveQuests {
		if tracked.Quest.ID == questID {
			if objectiveIndex >= 0 && objectiveIndex < len(tracked.Quest.Objectives) {
				return tracked.Quest.Objectives[objectiveIndex].Current
			}
		}
	}
	return 0
}

// AbandonQuest removes a quest from active quests without completing it.
func (q *QuestTrackerComponent) AbandonQuest(questID string) bool {
	for i, tracked := range q.ActiveQuests {
		if tracked.Quest.ID == questID {
			q.ActiveQuests = append(q.ActiveQuests[:i], q.ActiveQuests[i+1:]...)
			return true
		}
	}
	return false
}
