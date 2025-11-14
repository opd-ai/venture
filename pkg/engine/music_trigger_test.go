package engine

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/audio/music"
)

func TestNewMusicTriggerComponent(t *testing.T) {
	comp := NewMusicTriggerComponent()
	if comp == nil {
		t.Fatal("NewMusicTriggerComponent() returned nil")
	}

	if comp.CurrentContext.Location != "exploration" {
		t.Errorf("Initial location = %s, want exploration", comp.CurrentContext.Location)
	}

	if comp.CombatActive {
		t.Error("CombatActive should be false initially")
	}

	if comp.BossNearby {
		t.Error("BossNearby should be false initially")
	}

	if comp.ReputationTier != "neutral" {
		t.Errorf("ReputationTier = %s, want neutral", comp.ReputationTier)
	}
}

func TestMusicTriggerComponent_Type(t *testing.T) {
	comp := NewMusicTriggerComponent()
	if comp.Type() != "music_trigger" {
		t.Errorf("Type() = %s, want music_trigger", comp.Type())
	}
}

func TestMusicTriggerComponent_TriggerCombat(t *testing.T) {
	comp := NewMusicTriggerComponent()

	// Start combat
	comp.TriggerCombat(true)

	if !comp.CombatActive {
		t.Error("CombatActive should be true after TriggerCombat(true)")
	}

	if !comp.CurrentContext.Combat {
		t.Error("CurrentContext.Combat should be true")
	}

	if comp.CurrentContext.Danger < 0.5 {
		t.Errorf("Danger = %f, want >= 0.5 during combat", comp.CurrentContext.Danger)
	}

	// End combat
	comp.TriggerCombat(false)

	if comp.CombatActive {
		t.Error("CombatActive should be false after TriggerCombat(false)")
	}

	if comp.CurrentContext.Combat {
		t.Error("CurrentContext.Combat should be false")
	}
}

func TestMusicTriggerComponent_TriggerBoss(t *testing.T) {
	comp := NewMusicTriggerComponent()

	// Boss appears
	comp.TriggerBoss(true)

	if !comp.BossNearby {
		t.Error("BossNearby should be true after TriggerBoss(true)")
	}

	if !comp.CurrentContext.BossNearby {
		t.Error("CurrentContext.BossNearby should be true")
	}

	if comp.CurrentContext.Danger != 1.0 {
		t.Errorf("Danger = %f, want 1.0 during boss fight", comp.CurrentContext.Danger)
	}

	if !comp.CombatActive {
		t.Error("CombatActive should be true during boss fight")
	}

	// Boss defeated
	comp.TriggerBoss(false)

	if comp.BossNearby {
		t.Error("BossNearby should be false after TriggerBoss(false)")
	}
}

func TestMusicTriggerComponent_TriggerQuestCompletion(t *testing.T) {
	comp := NewMusicTriggerComponent()

	comp.TriggerQuestCompletion()

	if comp.PendingContext == nil {
		t.Fatal("PendingContext should not be nil after quest completion")
	}

	if comp.PendingContext.Location != "victory" {
		t.Errorf("PendingContext.Location = %s, want victory", comp.PendingContext.Location)
	}

	if comp.TransitionTime <= 0 {
		t.Errorf("TransitionTime = %f, want > 0", comp.TransitionTime)
	}
}

func TestMusicTriggerComponent_TriggerExploration(t *testing.T) {
	comp := NewMusicTriggerComponent()

	initialMilestones := comp.ExplorationMilestones

	comp.TriggerExploration(true)

	if comp.ExplorationMilestones != initialMilestones+1 {
		t.Errorf("ExplorationMilestones = %d, want %d", comp.ExplorationMilestones, initialMilestones+1)
	}

	if comp.CurrentContext.Location != "exploration" {
		t.Errorf("CurrentContext.Location = %s, want exploration", comp.CurrentContext.Location)
	}
}

func TestMusicTriggerComponent_TriggerReputationChange(t *testing.T) {
	comp := NewMusicTriggerComponent()

	tests := []struct {
		tier      string
		maxDanger float64
	}{
		{"hated", 0.5},
		{"hostile", 0.4},
		{"unfriendly", 0.3},
		{"neutral", 0.1},
		{"friendly", 0.05},
		{"honored", 0.0},
		{"revered", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.tier, func(t *testing.T) {
			comp.TriggerReputationChange(tt.tier)

			if comp.ReputationTier != tt.tier {
				t.Errorf("ReputationTier = %s, want %s", comp.ReputationTier, tt.tier)
			}

			if comp.CurrentContext.Danger > tt.maxDanger {
				t.Errorf("Danger = %f, want <= %f for tier %s", comp.CurrentContext.Danger, tt.maxDanger, tt.tier)
			}
		})
	}
}

func TestMusicTriggerComponent_UpdatePendingTransition(t *testing.T) {
	comp := NewMusicTriggerComponent()

	// Set pending context
	comp.TriggerQuestCompletion()

	if comp.PendingContext == nil {
		t.Fatal("PendingContext should be set")
	}

	initialTransitionTime := comp.TransitionTime

	// Update with small delta
	comp.UpdatePendingTransition(0.1)

	if comp.TransitionTime >= initialTransitionTime {
		t.Error("TransitionTime should decrease after update")
	}

	if comp.CurrentContext.Location != "victory" {
		t.Error("CurrentContext should be victory while pending")
	}

	// Update past transition time
	comp.UpdatePendingTransition(10.0)

	if comp.PendingContext != nil {
		t.Error("PendingContext should be nil after transition completes")
	}
}

func TestTriggerType_String(t *testing.T) {
	tests := []struct {
		trigger TriggerType
		want    string
	}{
		{TriggerCombatStart, "combat_start"},
		{TriggerCombatEnd, "combat_end"},
		{TriggerBossAppear, "boss_appear"},
		{TriggerBossDefeated, "boss_defeated"},
		{TriggerQuestComplete, "quest_complete"},
		{TriggerExplorationMilestone, "exploration_milestone"},
		{TriggerReputationChange, "reputation_change"},
		{TriggerType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.trigger.String()
			if got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestNewMusicTriggerSystem(t *testing.T) {
	world := NewWorld()
	manager := music.NewAdaptiveMusicManager(44100, 12345)
	manager.Initialize("fantasy", 60)

	system := NewMusicTriggerSystem(world, manager)

	if system == nil {
		t.Fatal("NewMusicTriggerSystem() returned nil")
	}

	if system.world != world {
		t.Error("System world not set correctly")
	}

	if system.musicManager != manager {
		t.Error("System musicManager not set correctly")
	}
}

func TestMusicTriggerSystem_QueueEvent(t *testing.T) {
	world := NewWorld()
	manager := music.NewAdaptiveMusicManager(44100, 12345)
	manager.Initialize("fantasy", 60)
	system := NewMusicTriggerSystem(world, manager)

	event := MusicTriggerEvent{
		Type:      TriggerCombatStart,
		EntityID:  1,
		Timestamp: time.Now(),
	}

	initialLen := system.GetEventQueueLength()

	system.QueueEvent(event)

	if system.GetEventQueueLength() != initialLen+1 {
		t.Errorf("Event queue length = %d, want %d", system.GetEventQueueLength(), initialLen+1)
	}
}

func TestMusicTriggerSystem_OnCombatStart(t *testing.T) {
	world := NewWorld()
	manager := music.NewAdaptiveMusicManager(44100, 12345)
	manager.Initialize("fantasy", 60)
	system := NewMusicTriggerSystem(world, manager)

	// Create entity with music trigger component
	entity := world.CreateEntity()
	comp := NewMusicTriggerComponent()
	entity.AddComponent(comp)

	// Process pending entity additions
	world.Update(0.0)

	// Verify initial state
	if comp.CombatActive {
		t.Fatal("CombatActive should be false initially")
	}

	// Queue event
	system.OnCombatStart(entity.ID)

	if system.GetEventQueueLength() != 1 {
		t.Errorf("Event queue length = %d, want 1", system.GetEventQueueLength())
	}

	// Process events (processes immediately now)
	system.Update(0.001)

	// Check queue was cleared
	if system.GetEventQueueLength() != 0 {
		t.Errorf("Event queue should be empty after Update(), got %d", system.GetEventQueueLength())
	}

	// Check that combat was triggered
	if !comp.CombatActive {
		t.Error("CombatActive should be true after processing combat start event")
	}
}

func TestMusicTriggerSystem_OnBossAppear(t *testing.T) {
	world := NewWorld()
	manager := music.NewAdaptiveMusicManager(44100, 12345)
	manager.Initialize("fantasy", 60)
	system := NewMusicTriggerSystem(world, manager)

	entity := world.CreateEntity()
	comp := NewMusicTriggerComponent()
	entity.AddComponent(comp)

	// Process pending entity additions
	world.Update(0.0)

	system.OnBossAppear(entity.ID)
	system.Update(0.001)

	if !comp.BossNearby {
		t.Error("BossNearby should be true after processing boss appear event")
	}

	if comp.CurrentContext.Danger != 1.0 {
		t.Errorf("Danger = %f, want 1.0", comp.CurrentContext.Danger)
	}
}

func TestMusicTriggerSystem_OnQuestComplete(t *testing.T) {
	world := NewWorld()
	manager := music.NewAdaptiveMusicManager(44100, 12345)
	manager.Initialize("fantasy", 60)
	system := NewMusicTriggerSystem(world, manager)

	entity := world.CreateEntity()
	comp := NewMusicTriggerComponent()
	entity.AddComponent(comp)

	// Process pending entity additions
	world.Update(0.0)

	system.OnQuestComplete(entity.ID)
	system.Update(0.001)

	if comp.PendingContext == nil {
		t.Error("PendingContext should be set after quest completion")
	}

	if comp.PendingContext != nil && comp.PendingContext.Location != "victory" {
		t.Errorf("PendingContext.Location = %s, want victory", comp.PendingContext.Location)
	}
}

func TestMusicTriggerSystem_SetMusicManager(t *testing.T) {
	world := NewWorld()
	manager1 := music.NewAdaptiveMusicManager(44100, 12345)
	system := NewMusicTriggerSystem(world, manager1)

	manager2 := music.NewAdaptiveMusicManager(44100, 54321)
	system.SetMusicManager(manager2)

	if system.musicManager != manager2 {
		t.Error("SetMusicManager did not update manager reference")
	}
}
