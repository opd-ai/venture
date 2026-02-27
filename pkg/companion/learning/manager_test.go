package learning

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestNewManager(t *testing.T) {
	manager := NewManager()
	if manager == nil {
		t.Fatal("NewManager returned nil")
	}
	if manager.companions == nil {
		t.Error("companions map not initialized")
	}
}

func TestNewManagerWithOptions(t *testing.T) {
	tests := []struct {
		name         string
		timeProvider TimeProvider
		logger       *logrus.Logger
	}{
		{
			name:         "nil logger uses default",
			timeProvider: DefaultTimeProvider(),
			logger:       nil,
		},
		{
			name:         "custom logger",
			timeProvider: DefaultTimeProvider(),
			logger:       logrus.New(),
		},
		{
			name:         "custom time provider and logger",
			timeProvider: &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)},
			logger:       logrus.New(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManagerWithOptions(tt.timeProvider, tt.logger)
			if manager == nil {
				t.Fatal("NewManagerWithOptions returned nil")
			}
			if manager.companions == nil {
				t.Error("companions map not initialized")
			}
			if manager.logger == nil {
				t.Error("logger not initialized")
			}
			if tt.logger != nil && manager.logger != tt.logger {
				t.Error("custom logger not set")
			}
		})
	}
}

func TestManagerWithCustomLogger(t *testing.T) {
	// Create a custom logger that writes to a buffer
	var buf bytes.Buffer
	customLogger := logrus.New()
	customLogger.SetOutput(&buf)
	customLogger.SetLevel(logrus.DebugLevel)

	manager := NewManagerWithOptions(DefaultTimeProvider(), customLogger)
	manager.AddCompanion("test_companion", 1.0)

	// Verify logs were written to our custom logger
	output := buf.String()
	if len(output) == 0 {
		t.Error("expected log output from custom logger")
	}
}

func TestAddCompanion(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test_companion", 1.5)

	if comp == nil {
		t.Fatal("AddCompanion returned nil")
	}
	if comp.CompanionID != "test_companion" {
		t.Errorf("expected ID 'test_companion', got '%s'", comp.CompanionID)
	}
	if comp.LearningRate != 1.5 {
		t.Errorf("expected learning rate 1.5, got %f", comp.LearningRate)
	}
	if comp.SkillTree == nil {
		t.Error("SkillTree not initialized")
	}
	if comp.Personality == nil {
		t.Error("Personality not initialized")
	}
	if comp.Memory == nil {
		t.Error("Memory not initialized")
	}
}

func TestAddCompanionDefaultLearningRate(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test_companion", 0)

	if comp.LearningRate != 1.0 {
		t.Errorf("expected default learning rate 1.0, got %f", comp.LearningRate)
	}
}

func TestGetCompanion(t *testing.T) {
	manager := NewManager()
	manager.AddCompanion("test_companion", 1.0)

	comp, ok := manager.GetCompanion("test_companion")
	if !ok {
		t.Fatal("GetCompanion returned false")
	}
	if comp.CompanionID != "test_companion" {
		t.Errorf("expected ID 'test_companion', got '%s'", comp.CompanionID)
	}

	_, ok = manager.GetCompanion("nonexistent")
	if ok {
		t.Error("GetCompanion returned true for nonexistent companion")
	}
}

func TestRemoveCompanion(t *testing.T) {
	manager := NewManager()
	manager.AddCompanion("test_companion", 1.0)
	manager.RemoveCompanion("test_companion")

	_, ok := manager.GetCompanion("test_companion")
	if ok {
		t.Error("companion still exists after removal")
	}
}

func TestSkillProgression(t *testing.T) {
	sp := NewSkillProgression()

	if sp.Skills == nil {
		t.Fatal("Skills map not initialized")
	}
	if sp.SkillTree == nil {
		t.Fatal("SkillTree map not initialized")
	}
	if len(sp.Skills) == 0 {
		t.Error("Skills map is empty after initialization")
	}
}

func TestAddExperience(t *testing.T) {
	sp := NewSkillProgression()

	err := sp.AddExperience("Basic Attack", 50.0, 1.0)
	if err != nil {
		t.Fatalf("AddExperience failed: %v", err)
	}

	skill := sp.Skills["Basic Attack"]
	if skill.Experience != 50.0 {
		t.Errorf("expected XP 50.0, got %f", skill.Experience)
	}
}

func TestAddExperienceLevelUp(t *testing.T) {
	sp := NewSkillProgression()

	err := sp.AddExperience("Basic Attack", 150.0, 1.0)
	if err != nil {
		t.Fatalf("AddExperience failed: %v", err)
	}

	skill := sp.Skills["Basic Attack"]
	if skill.Level != 1 {
		t.Errorf("expected level 1, got %d", skill.Level)
	}
	if sp.AvailablePoints != 1 {
		t.Errorf("expected 1 skill point, got %d", sp.AvailablePoints)
	}
}

func TestAddExperienceWithLearningRate(t *testing.T) {
	sp := NewSkillProgression()

	err := sp.AddExperience("Basic Attack", 50.0, 2.0)
	if err != nil {
		t.Fatalf("AddExperience failed: %v", err)
	}

	skill := sp.Skills["Basic Attack"]
	// 50 * 2.0 = 100 XP, which exactly levels up from 0→1, leaving 0 XP
	if skill.Level != 1 {
		t.Errorf("expected level 1, got %d", skill.Level)
	}
	if skill.Experience != 0.0 {
		t.Errorf("expected 0 XP after level-up, got %f", skill.Experience)
	}
}

func TestAddExperienceMaxLevel(t *testing.T) {
	sp := NewSkillProgression()
	skill := sp.Skills["Basic Attack"]
	skill.Level = skill.MaxLevel

	err := sp.AddExperience("Basic Attack", 100.0, 1.0)
	if err != nil {
		t.Fatalf("AddExperience failed: %v", err)
	}

	if skill.Experience != 0.0 {
		t.Errorf("expected no XP gain at max level, got %f", skill.Experience)
	}
}

func TestAddExperienceInvalidSkill(t *testing.T) {
	sp := NewSkillProgression()

	err := sp.AddExperience("Invalid Skill", 50.0, 1.0)
	if err == nil {
		t.Error("expected error for invalid skill, got nil")
	}
}

func TestCanLearnSkill(t *testing.T) {
	sp := NewSkillProgression()
	sp.AvailablePoints = 5

	canLearn, err := sp.CanLearnSkill("Basic Attack")
	if err != nil {
		t.Fatalf("CanLearnSkill failed: %v", err)
	}
	if !canLearn {
		t.Error("expected to be able to learn Basic Attack")
	}
}

func TestCanLearnSkillInsufficientPoints(t *testing.T) {
	sp := NewSkillProgression()
	sp.AvailablePoints = 0

	canLearn, err := sp.CanLearnSkill("Basic Attack")
	if canLearn {
		t.Error("expected cannot learn with 0 points")
	}
	if err == nil {
		t.Error("expected error for insufficient points")
	}
}

func TestCanLearnSkillPrerequisiteNotMet(t *testing.T) {
	sp := NewSkillProgression()
	sp.AvailablePoints = 5

	canLearn, err := sp.CanLearnSkill("Power Strike")
	if canLearn {
		t.Error("expected cannot learn without prerequisite")
	}
	if err == nil {
		t.Error("expected error for prerequisite not met")
	}
}

func TestLearnSkill(t *testing.T) {
	sp := NewSkillProgression()
	sp.AvailablePoints = 5

	err := sp.LearnSkill("Basic Attack")
	if err != nil {
		t.Fatalf("LearnSkill failed: %v", err)
	}

	if sp.AvailablePoints != 4 {
		t.Errorf("expected 4 points remaining, got %d", sp.AvailablePoints)
	}
}

func TestPersonalityEvolution(t *testing.T) {
	pe := NewPersonalityEvolution()

	if pe.Traits == nil {
		t.Fatal("Traits map not initialized")
	}
	if len(pe.Traits) != 10 {
		t.Errorf("expected 10 traits, got %d", len(pe.Traits))
	}

	for trait, value := range pe.Traits {
		if value != 0.5 {
			t.Errorf("expected neutral value 0.5 for %s, got %f", trait.String(), value)
		}
	}
}

func TestAdjustTrait(t *testing.T) {
	pe := NewPersonalityEvolution()

	pe.AdjustTrait(TraitBrave, 0.2, "test reason")

	if pe.Traits[TraitBrave] != 0.7 {
		t.Errorf("expected 0.7, got %f", pe.Traits[TraitBrave])
	}

	if len(pe.Changes) != 1 {
		t.Errorf("expected 1 change record, got %d", len(pe.Changes))
	}
}

func TestAdjustTraitClamping(t *testing.T) {
	pe := NewPersonalityEvolution()

	pe.AdjustTrait(TraitBrave, 1.0, "max test")
	if pe.Traits[TraitBrave] != 1.0 {
		t.Errorf("expected clamped to 1.0, got %f", pe.Traits[TraitBrave])
	}

	pe.AdjustTrait(TraitBrave, -2.0, "min test")
	if pe.Traits[TraitBrave] != 0.0 {
		t.Errorf("expected clamped to 0.0, got %f", pe.Traits[TraitBrave])
	}
}

func TestGetDominantTrait(t *testing.T) {
	pe := NewPersonalityEvolution()
	pe.Traits[TraitBrave] = 0.9
	pe.Traits[TraitCautious] = 0.1

	dominant := pe.GetDominantTrait()
	if dominant != TraitBrave {
		t.Errorf("expected TraitBrave, got %s", dominant.String())
	}
}

func TestEventMemory(t *testing.T) {
	em := NewEventMemory(1000)

	if em.Events == nil {
		t.Fatal("Events slice not initialized")
	}
	if em.MaxEvents != 1000 {
		t.Errorf("expected MaxEvents 1000, got %d", em.MaxEvents)
	}
}

func TestAddEvent(t *testing.T) {
	em := NewEventMemory(1000)

	mockTime := &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)}
	event := MemorableEvent{
		Type:        EventCombat,
		Description: "test combat",
		Timestamp:   mockTime.Now(),
		Importance:  0.5,
	}

	em.AddEvent(event)

	if em.TotalEvents != 1 {
		t.Errorf("expected TotalEvents 1, got %d", em.TotalEvents)
	}
	if len(em.Events) != 1 {
		t.Errorf("expected 1 event, got %d", len(em.Events))
	}
}

func TestAddEventLRUEviction(t *testing.T) {
	em := NewEventMemory(10)

	mockTime := &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)}
	for i := 0; i < 15; i++ {
		event := MemorableEvent{
			Type:        EventCombat,
			Description: "test event",
			Timestamp:   mockTime.Now(),
			Importance:  0.5,
		}
		em.AddEvent(event)
	}

	if em.TotalEvents != 15 {
		t.Errorf("expected TotalEvents 15, got %d", em.TotalEvents)
	}
	if len(em.Events) != 10 {
		t.Errorf("expected 10 events (LRU), got %d", len(em.Events))
	}
}

func TestGetRecentEvents(t *testing.T) {
	em := NewEventMemory(1000)

	mockTime := &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)}
	for i := 0; i < 10; i++ {
		event := MemorableEvent{
			Type:        EventCombat,
			Description: "test event",
			Timestamp:   mockTime.Now(),
			Importance:  0.5,
		}
		em.AddEvent(event)
	}

	recent := em.GetRecentEvents(5)
	if len(recent) != 5 {
		t.Errorf("expected 5 recent events, got %d", len(recent))
	}
}

func TestGetEventsByType(t *testing.T) {
	em := NewEventMemory(1000)

	mockTime := &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)}
	em.AddEvent(MemorableEvent{Type: EventCombat, Timestamp: mockTime.Now()})
	em.AddEvent(MemorableEvent{Type: EventDialog, Timestamp: mockTime.Now()})
	em.AddEvent(MemorableEvent{Type: EventCombat, Timestamp: mockTime.Now()})

	combatEvents := em.GetEventsByType(EventCombat)
	if len(combatEvents) != 2 {
		t.Errorf("expected 2 combat events, got %d", len(combatEvents))
	}
}

func TestProcessCombatAction(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	ProcessCombatAction(comp, true, true)

	if comp.SkillTree.TotalXP == 0 {
		t.Error("expected XP gain from combat action")
	}
	if len(comp.Memory.Events) == 0 {
		t.Error("expected combat event in memory")
	}
}

func TestProcessSocialInteraction(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	ProcessSocialInteraction(comp, "player123", true)

	if comp.SkillTree.TotalXP == 0 {
		t.Error("expected XP gain from social interaction")
	}
	if len(comp.Memory.Events) == 0 {
		t.Error("expected social event in memory")
	}
}

func TestProcessExploration(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	ProcessExploration(comp, true)

	if comp.SkillTree.TotalXP == 0 {
		t.Error("expected XP gain from exploration")
	}
	if len(comp.Memory.Events) == 0 {
		t.Error("expected exploration event in memory")
	}
}

func TestGeneratePersonalityDescription(t *testing.T) {
	pe := NewPersonalityEvolution()
	pe.Traits[TraitBrave] = 0.9

	desc := GeneratePersonalityDescription(pe)
	if len(desc) == 0 {
		t.Error("expected non-empty personality description")
	}
}

func TestAdaptBehaviorToCombatStyle(t *testing.T) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)

	// Add some combat events
	mockTime := &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)}
	for i := 0; i < 10; i++ {
		comp.Memory.AddEvent(MemorableEvent{
			Type:      EventCombat,
			Timestamp: mockTime.Now(),
		})
	}

	initialBrave := comp.Personality.Traits[TraitBrave]

	AdaptBehaviorToCombatStyle(comp, 12345)

	// Traits should have changed
	finalBrave := comp.Personality.Traits[TraitBrave]
	if initialBrave == finalBrave && comp.Personality.Traits[TraitCautious] == 0.5 {
		t.Error("expected personality traits to change after adaptation")
	}
}

// Benchmarks

func BenchmarkAddExperience(b *testing.B) {
	sp := NewSkillProgression()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = sp.AddExperience("Basic Attack", 10.0, 1.0)
	}
}

func BenchmarkAdjustTrait(b *testing.B) {
	pe := NewPersonalityEvolution()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		pe.AdjustTrait(TraitBrave, 0.01, "benchmark test")
	}
}

func BenchmarkAddEvent(b *testing.B) {
	em := NewEventMemory(1000)
	mockTime := &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)}
	event := MemorableEvent{
		Type:        EventCombat,
		Description: "benchmark event",
		Timestamp:   mockTime.Now(),
		Importance:  0.5,
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		em.AddEvent(event)
	}
}

func BenchmarkProcessCombatAction(b *testing.B) {
	manager := NewManager()
	comp := manager.AddCompanion("test", 1.0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ProcessCombatAction(comp, true, true)
	}
}

// TestProcessCombatAction_NilComponent verifies nil safety.
func TestProcessCombatAction_NilComponent(t *testing.T) {
	// Should not panic
	ProcessCombatAction(nil, true, true)
	ProcessCombatAction(nil, false, false)
}

// TestProcessSocialInteraction_NilComponent verifies nil safety.
func TestProcessSocialInteraction_NilComponent(t *testing.T) {
	ProcessSocialInteraction(nil, "player1", true)
	ProcessSocialInteraction(nil, "player1", false)
}

// TestProcessExploration_NilComponent verifies nil safety.
func TestProcessExploration_NilComponent(t *testing.T) {
	ProcessExploration(nil, true)
	ProcessExploration(nil, false)
}

// TestAdaptBehaviorToCombatStyle_NilComponent verifies nil safety.
func TestAdaptBehaviorToCombatStyle_NilComponent(t *testing.T) {
	AdaptBehaviorToCombatStyle(nil, 42)
}

// TestGeneratePersonalityDescription_NilPersonality verifies nil safety.
func TestGeneratePersonalityDescription_NilPersonality(t *testing.T) {
	result := GeneratePersonalityDescription(nil)
	if result != "No personality data" {
		t.Errorf("expected 'No personality data', got %q", result)
	}
}

// TestConcurrentUpdateAndModify verifies Update is safe during concurrent Add/Remove.
func TestConcurrentUpdateAndModify(t *testing.T) {
	mockTime := &MockTimeProvider{CurrentTime: time.Unix(1000000, 0)}
	manager := NewManagerWithOptions(mockTime, nil)
	system := &CompanionLearningSystem{
		manager:        manager,
		updateInterval: time.Millisecond,
		timeProvider:   mockTime,
		lastUpdate:     time.Time{},
	}

	// Pre-populate some companions
	for i := 0; i < 5; i++ {
		comp := manager.AddCompanion(string(rune('A'+i)), 1.0)
		_ = comp.SkillTree.AddExperience("Basic Attack", 50.0, 1.0)
		comp.LastSkillUse["Basic Attack"] = mockTime.Now().Add(-48 * time.Hour)
	}

	done := make(chan struct{})

	// Concurrent Update calls
	go func() {
		defer func() { done <- struct{}{} }()
		for i := 0; i < 100; i++ {
			time.Sleep(time.Millisecond)
			system.Update(0.016)
		}
	}()

	// Concurrent Add/Remove of different companions (not modifying pre-existing ones)
	go func() {
		defer func() { done <- struct{}{} }()
		for i := 0; i < 100; i++ {
			id := fmt.Sprintf("dynamic-%d", i)
			manager.AddCompanion(id, 1.0)
			if i%3 == 0 {
				manager.RemoveCompanion(id)
			}
		}
	}()

	<-done
	<-done
}
