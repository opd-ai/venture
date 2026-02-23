//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen/narrative"
	"github.com/sirupsen/logrus"
)

// testTimeProvider implements TimeProvider for deterministic testing.
type testTimeProvider struct {
	fixedTime time.Time
}

func (t *testTimeProvider) Now() time.Time {
	return t.fixedTime
}

// TestSetupNarrativeComponentNew tests creation of new narrative component.
func TestSetupNarrativeComponentNew(t *testing.T) {
	entity := engine.NewEntity(1)
	arc := &narrative.StoryArc{
		Title:        "Test Story",
		MainConflict: "Save the world",
		Antagonist:   "Dark Lord",
		PlotPoints: []narrative.PlotPoint{
			{Act: 1, Description: "Plot 1"},
			{Act: 2, Description: "Plot 2"},
		},
	}

	comp := setupNarrativeComponent(entity, arc)

	if comp == nil {
		t.Fatal("setupNarrativeComponent returned nil")
	}
	if comp.MainObjective != arc.MainConflict {
		t.Errorf("MainObjective = %q, want %q", comp.MainObjective, arc.MainConflict)
	}
	if comp.CurrentAct != engine.ActSetup {
		t.Errorf("CurrentAct = %v, want %v", comp.CurrentAct, engine.ActSetup)
	}
	if comp.StoryProgress != 0.0 {
		t.Errorf("StoryProgress = %f, want 0.0", comp.StoryProgress)
	}
	if len(comp.ActiveThreads) != 0 {
		t.Errorf("ActiveThreads len = %d, want 0", len(comp.ActiveThreads))
	}
}

// TestSetupNarrativeComponentExisting tests updating existing narrative component.
func TestSetupNarrativeComponentExisting(t *testing.T) {
	entity := engine.NewEntity(1)
	existingComp := &engine.NarrativeComponent{
		CurrentAct:      engine.ActConfrontation,
		MainObjective:   "Old objective",
		StoryProgress:   0.5,
		ActiveThreads:   []string{"thread1"},
		ResolvedThreads: []string{"resolved1"},
		WorldStateFlags: make(map[string]bool),
		Relationships:   make(map[string]float64),
		EventHistory:    make([]engine.NarrativeEvent, 0),
		PlayerDecisions: make([]engine.PlayerDecision, 0),
	}
	entity.AddComponent(existingComp)

	arc := &narrative.StoryArc{
		Title:        "New Story",
		MainConflict: "New objective",
	}

	comp := setupNarrativeComponent(entity, arc)

	// Should update main objective but preserve other state
	if comp.MainObjective != "New objective" {
		t.Errorf("MainObjective = %q, want %q", comp.MainObjective, "New objective")
	}
	// Should preserve existing progress
	if comp.StoryProgress != 0.5 {
		t.Errorf("StoryProgress = %f, want 0.5", comp.StoryProgress)
	}
	// Should preserve existing threads
	if len(comp.ActiveThreads) != 1 {
		t.Errorf("ActiveThreads len = %d, want 1", len(comp.ActiveThreads))
	}
}

// TestAddInitialNarrativeEvent tests adding the initial story event.
func TestAddInitialNarrativeEvent(t *testing.T) {
	fixedTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tp := &testTimeProvider{fixedTime: fixedTime}

	narrativeComp := &engine.NarrativeComponent{
		CurrentAct:      engine.ActSetup,
		EventHistory:    make([]engine.NarrativeEvent, 0),
		MainObjective:   "Test objective",
		StoryProgress:   0.0,
		ActiveThreads:   make([]string, 0),
		ResolvedThreads: make([]string, 0),
		WorldStateFlags: make(map[string]bool),
		Relationships:   make(map[string]float64),
		PlayerDecisions: make([]engine.PlayerDecision, 0),
	}

	arc := &narrative.StoryArc{
		Title: "Epic Adventure",
	}

	addInitialNarrativeEvent(narrativeComp, arc, tp)

	if len(narrativeComp.EventHistory) != 1 {
		t.Fatalf("EventHistory len = %d, want 1", len(narrativeComp.EventHistory))
	}

	event := narrativeComp.EventHistory[0]
	if event.Type != engine.EventDiscovery {
		t.Errorf("event Type = %v, want %v", event.Type, engine.EventDiscovery)
	}
	// Note: Timestamp is overridden by NarrativeComponent.AddEvent() with time.Now()
	// This is existing behavior in the engine package - verify only that a timestamp exists
	if event.Timestamp.IsZero() {
		t.Error("event Timestamp should not be zero")
	}
	// Note: Act is also overridden by AddEvent() to use CurrentAct of the component
	if event.Act != engine.ActSetup {
		t.Errorf("event Act = %v, want %v", event.Act, engine.ActSetup)
	}
	expectedDesc := "The story begins: Epic Adventure"
	if event.Description != expectedDesc {
		t.Errorf("event Description = %q, want %q", event.Description, expectedDesc)
	}
}

// TestAddInitialNarrativeEventDescriptionFormat verifies the event description format.
func TestAddInitialNarrativeEventDescriptionFormat(t *testing.T) {
	tests := []struct {
		name     string
		arcTitle string
		wantDesc string
	}{
		{"simple title", "Epic Quest", "The story begins: Epic Quest"},
		{"empty title", "", "The story begins: "},
		{"special chars", "The Hero's Journey", "The story begins: The Hero's Journey"},
		{"long title", "A Very Long Title That Goes On And On", "The story begins: A Very Long Title That Goes On And On"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp := &testTimeProvider{fixedTime: time.Now()}
			comp := &engine.NarrativeComponent{
				CurrentAct:      engine.ActSetup,
				EventHistory:    make([]engine.NarrativeEvent, 0),
				MainObjective:   "",
				StoryProgress:   0.0,
				ActiveThreads:   make([]string, 0),
				ResolvedThreads: make([]string, 0),
				WorldStateFlags: make(map[string]bool),
				Relationships:   make(map[string]float64),
				PlayerDecisions: make([]engine.PlayerDecision, 0),
			}
			arc := &narrative.StoryArc{Title: tt.arcTitle}

			addInitialNarrativeEvent(comp, arc, tp)

			if comp.EventHistory[0].Description != tt.wantDesc {
				t.Errorf("Description = %q, want %q", comp.EventHistory[0].Description, tt.wantDesc)
			}
		})
	}
}

// TestAddPlotPointsAsThreads tests adding Act 1 plot points as threads.
func TestAddPlotPointsAsThreads(t *testing.T) {
	narrativeComp := &engine.NarrativeComponent{
		ActiveThreads: make([]string, 0),
	}

	arc := &narrative.StoryArc{
		PlotPoints: []narrative.PlotPoint{
			{Act: 1, Description: "First plot point"},
			{Act: 1, Description: "Second plot point"},
			{Act: 2, Description: "Midpoint (should not be added)"},
			{Act: 3, Description: "Climax (should not be added)"},
		},
	}

	addPlotPointsAsThreads(narrativeComp, arc)

	if len(narrativeComp.ActiveThreads) != 2 {
		t.Errorf("ActiveThreads len = %d, want 2", len(narrativeComp.ActiveThreads))
	}

	// Verify only Act 1 plot points were added
	for _, thread := range narrativeComp.ActiveThreads {
		if thread != "First plot point" && thread != "Second plot point" {
			t.Errorf("unexpected thread: %q", thread)
		}
	}
}

// TestAddPlotPointsAsThreadsEmpty tests with no plot points.
func TestAddPlotPointsAsThreadsEmpty(t *testing.T) {
	narrativeComp := &engine.NarrativeComponent{
		ActiveThreads: make([]string, 0),
	}

	arc := &narrative.StoryArc{
		PlotPoints: []narrative.PlotPoint{},
	}

	addPlotPointsAsThreads(narrativeComp, arc)

	if len(narrativeComp.ActiveThreads) != 0 {
		t.Errorf("ActiveThreads len = %d, want 0", len(narrativeComp.ActiveThreads))
	}
}

// TestAddPlotPointsAsThreadsNoAct1 tests with no Act 1 plot points.
func TestAddPlotPointsAsThreadsNoAct1(t *testing.T) {
	narrativeComp := &engine.NarrativeComponent{
		ActiveThreads: make([]string, 0),
	}

	arc := &narrative.StoryArc{
		PlotPoints: []narrative.PlotPoint{
			{Act: 2, Description: "Midpoint"},
			{Act: 3, Description: "Climax"},
		},
	}

	addPlotPointsAsThreads(narrativeComp, arc)

	if len(narrativeComp.ActiveThreads) != 0 {
		t.Errorf("ActiveThreads len = %d, want 0 (no Act 1 points)", len(narrativeComp.ActiveThreads))
	}
}

// TestAddPlotPointsAsThreadsPreservesExisting tests that existing threads are preserved.
func TestAddPlotPointsAsThreadsPreservesExisting(t *testing.T) {
	narrativeComp := &engine.NarrativeComponent{
		ActiveThreads: []string{"existing thread 1", "existing thread 2"},
	}

	arc := &narrative.StoryArc{
		PlotPoints: []narrative.PlotPoint{
			{Act: 1, Description: "New thread"},
		},
	}

	addPlotPointsAsThreads(narrativeComp, arc)

	if len(narrativeComp.ActiveThreads) != 3 {
		t.Errorf("ActiveThreads len = %d, want 3", len(narrativeComp.ActiveThreads))
	}

	// Check existing threads are preserved
	hasExisting1 := false
	hasExisting2 := false
	hasNew := false
	for _, thread := range narrativeComp.ActiveThreads {
		switch thread {
		case "existing thread 1":
			hasExisting1 = true
		case "existing thread 2":
			hasExisting2 = true
		case "New thread":
			hasNew = true
		}
	}

	if !hasExisting1 || !hasExisting2 || !hasNew {
		t.Errorf("threads not preserved correctly: existing1=%v, existing2=%v, new=%v",
			hasExisting1, hasExisting2, hasNew)
	}
}

// TestLogNarrativeArcSuccess tests logging function does not panic.
func TestLogNarrativeArcSuccess(t *testing.T) {
	// Test that logging doesn't panic with various arc configurations
	tests := []struct {
		name string
		arc  *narrative.StoryArc
	}{
		{
			name: "full arc",
			arc: &narrative.StoryArc{
				Title:      "Full Story",
				Antagonist: "Villain",
				Ally:       "Sidekick",
				PlotPoints: []narrative.PlotPoint{
					{Act: 1, Description: "Point 1"},
				},
				PossibleEndings: []string{"Good ending", "Bad ending"},
				Difficulty:      0.5,
				Seed:            12345,
			},
		},
		{
			name: "minimal arc",
			arc: &narrative.StoryArc{
				Title: "Minimal",
			},
		},
		{
			name: "empty strings",
			arc: &narrative.StoryArc{
				Title:      "",
				Antagonist: "",
				Ally:       "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := logrus.New()
			logger.SetLevel(logrus.WarnLevel) // Suppress output
			entry := logger.WithField("test", true)

			// Should not panic
			logNarrativeArcSuccess(tt.arc, entry)
		})
	}
}

// TestLogNarrativeArcSuccessVerbose tests verbose logging mode.
func TestLogNarrativeArcSuccessVerbose(t *testing.T) {
	// Save and restore verbose flag
	oldVerbose := *verbose
	defer func() { *verbose = oldVerbose }()

	*verbose = true

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	entry := logger.WithField("test", true)

	arc := &narrative.StoryArc{
		Title:           "Verbose Test",
		Antagonist:      "Test Villain",
		Ally:            "Test Ally",
		PlotPoints:      make([]narrative.PlotPoint, 3),
		PossibleEndings: []string{"End 1", "End 2"},
		Difficulty:      0.7,
		Seed:            98765,
	}

	// Should not panic and should log with verbose mode enabled
	logNarrativeArcSuccess(arc, entry)
}

// BenchmarkSetupNarrativeComponent benchmarks narrative component setup.
func BenchmarkSetupNarrativeComponent(b *testing.B) {
	arc := &narrative.StoryArc{
		Title:        "Benchmark Story",
		MainConflict: "Performance testing",
		PlotPoints:   make([]narrative.PlotPoint, 10),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		entity := engine.NewEntity(uint64(i))
		setupNarrativeComponent(entity, arc)
	}
}

// BenchmarkAddPlotPointsAsThreads benchmarks plot point thread addition.
func BenchmarkAddPlotPointsAsThreads(b *testing.B) {
	arc := &narrative.StoryArc{
		PlotPoints: []narrative.PlotPoint{
			{Act: 1, Description: "Plot 1"},
			{Act: 1, Description: "Plot 2"},
			{Act: 1, Description: "Plot 3"},
			{Act: 2, Description: "Plot 4"},
			{Act: 3, Description: "Plot 5"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		comp := &engine.NarrativeComponent{
			ActiveThreads: make([]string, 0),
		}
		addPlotPointsAsThreads(comp, arc)
	}
}
