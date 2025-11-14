package story

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestTimelineGenerate(t *testing.T) {
	gen := NewTimelineGenerator()
	seed := int64(99999)

	tests := []struct {
		name     string
		params   procgen.GenerationParams
		wantErr  bool
		minEras  int
		maxEras  int
		minEvents int
	}{
		{
			name: "basic generation",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr:   false,
			minEras:   2,
			maxEras:   5,
			minEvents: 10,
		},
		{
			name: "high depth",
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      15,
				GenreID:    "scifi",
			},
			wantErr:   false,
			minEras:   2,
			maxEras:   5,
			minEvents: 10,
		},
		{
			name: "low depth",
			params: procgen.GenerationParams{
				Difficulty: 0.3,
				Depth:      2,
				GenreID:    "horror",
			},
			wantErr:   false,
			minEras:   2,
			maxEras:   5,
			minEvents: 10,
		},
		{
			name: "invalid difficulty low",
			params: procgen.GenerationParams{
				Difficulty: -1.0,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty high",
			params: procgen.GenerationParams{
				Difficulty: 1.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := gen.Generate(seed, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			timeline, ok := result.(*Timeline)
			if !ok {
				t.Errorf("Generate() returned wrong type")
				return
			}

			// Check era count
			if len(timeline.Eras) < tt.minEras || len(timeline.Eras) > tt.maxEras {
				t.Errorf("Era count = %d, want between %d and %d", len(timeline.Eras), tt.minEras, tt.maxEras)
			}

			// Check event count
			if len(timeline.Events) < tt.minEvents {
				t.Errorf("Event count = %d, want at least %d", len(timeline.Events), tt.minEvents)
			}

			// Check consistency
			if timeline.Consistency < 0.5 || timeline.Consistency > 1.0 {
				t.Errorf("Consistency = %.2f, want between 0.5 and 1.0", timeline.Consistency)
			}

			// Check year range
			if timeline.StartYear <= timeline.CurrentYear {
				t.Errorf("StartYear (%d) should be > CurrentYear (%d)", timeline.StartYear, timeline.CurrentYear)
			}

			// Check events are chronological (most ancient first)
			for i := 0; i < len(timeline.Events)-1; i++ {
				if timeline.Events[i].Timestamp < timeline.Events[i+1].Timestamp {
					t.Errorf("Events not chronological at index %d: %d < %d", i, timeline.Events[i].Timestamp, timeline.Events[i+1].Timestamp)
				}
			}
		})
	}
}

func TestTimelineValidate(t *testing.T) {
	gen := NewTimelineGenerator()

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "valid timeline",
			input: &Timeline{
				Eras: []Era{
					{Name: "Era 1", StartYear: 1000, EndYear: 500},
					{Name: "Era 2", StartYear: 500, EndYear: 0},
				},
				Events: []HistoricalEvent{
					{Timestamp: 900, Title: "Event 1", EventType: EventFoundation},
					{Timestamp: 800, Title: "Event 2", EventType: EventWar},
					{Timestamp: 700, Title: "Event 3", EventType: EventDiscovery},
					{Timestamp: 600, Title: "Event 4", EventType: EventCatastrophe},
					{Timestamp: 500, Title: "Event 5", EventType: EventRenaissance},
					{Timestamp: 400, Title: "Event 6", EventType: EventCollapse},
					{Timestamp: 300, Title: "Event 7", EventType: EventContact},
					{Timestamp: 200, Title: "Event 8", EventType: EventRitual},
					{Timestamp: 100, Title: "Event 9", EventType: EventFoundation},
					{Timestamp: 50, Title: "Event 10", EventType: EventWar},
				},
				StartYear:   1000,
				CurrentYear: 0,
				Consistency: 0.7,
			},
			wantErr: false,
		},
		{
			name:    "wrong type",
			input:   "not a timeline",
			wantErr: true,
		},
		{
			name: "too few eras",
			input: &Timeline{
				Eras:        []Era{{Name: "Era 1"}},
				Events:      make([]HistoricalEvent, 10),
				StartYear:   1000,
				CurrentYear: 0,
				Consistency: 0.7,
			},
			wantErr: true,
		},
		{
			name: "too many eras",
			input: &Timeline{
				Eras:        make([]Era, 6),
				Events:      make([]HistoricalEvent, 10),
				StartYear:   1000,
				CurrentYear: 0,
				Consistency: 0.7,
			},
			wantErr: true,
		},
		{
			name: "too few events",
			input: &Timeline{
				Eras:        []Era{{Name: "Era 1"}, {Name: "Era 2"}},
				Events:      make([]HistoricalEvent, 5),
				StartYear:   1000,
				CurrentYear: 0,
				Consistency: 0.7,
			},
			wantErr: true,
		},
		{
			name: "invalid year range",
			input: &Timeline{
				Eras:        []Era{{Name: "Era 1"}, {Name: "Era 2"}},
				Events:      make([]HistoricalEvent, 10),
				StartYear:   0,
				CurrentYear: 1000,
				Consistency: 0.7,
			},
			wantErr: true,
		},
		{
			name: "low consistency",
			input: &Timeline{
				Eras:        []Era{{Name: "Era 1"}, {Name: "Era 2"}},
				Events:      make([]HistoricalEvent, 10),
				StartYear:   1000,
				CurrentYear: 0,
				Consistency: 0.3,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTimelineDeterminism(t *testing.T) {
	gen := NewTimelineGenerator()
	seed := int64(22222)
	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      8,
		GenreID:    "cyberpunk",
	}

	// Generate twice with same seed
	result1, err1 := gen.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("First generation failed: %v", err1)
	}

	result2, err2 := gen.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Second generation failed: %v", err2)
	}

	timeline1 := result1.(*Timeline)
	timeline2 := result2.(*Timeline)

	// Check same era count
	if len(timeline1.Eras) != len(timeline2.Eras) {
		t.Errorf("Different era counts: %d vs %d", len(timeline1.Eras), len(timeline2.Eras))
	}

	// Check same event count
	if len(timeline1.Events) != len(timeline2.Events) {
		t.Errorf("Different event counts: %d vs %d", len(timeline1.Events), len(timeline2.Events))
	}

	// Check same start year
	if timeline1.StartYear != timeline2.StartYear {
		t.Errorf("Different start years: %d vs %d", timeline1.StartYear, timeline2.StartYear)
	}

	// Check same consistency
	if timeline1.Consistency != timeline2.Consistency {
		t.Errorf("Different consistency: %.4f vs %.4f", timeline1.Consistency, timeline2.Consistency)
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		eventType EventType
		want      string
	}{
		{EventFoundation, "Foundation"},
		{EventWar, "War"},
		{EventDiscovery, "Discovery"},
		{EventCatastrophe, "Catastrophe"},
		{EventRenaissance, "Renaissance"},
		{EventCollapse, "Collapse"},
		{EventContact, "Contact"},
		{EventRitual, "Ritual"},
		{EventType(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.eventType.String()
			if got != tt.want {
				t.Errorf("EventType.String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestGetEventsInPeriod(t *testing.T) {
	timeline := &Timeline{
		Events: []HistoricalEvent{
			{Timestamp: 1000, Title: "Ancient"},
			{Timestamp: 800, Title: "Old"},
			{Timestamp: 500, Title: "Medieval"},
			{Timestamp: 300, Title: "Recent"},
			{Timestamp: 100, Title: "Modern"},
		},
	}

	tests := []struct {
		name      string
		startYear int64
		endYear   int64
		wantCount int
	}{
		{
			name:      "all events",
			startYear: 1000,
			endYear:   0,
			wantCount: 5,
		},
		{
			name:      "middle period",
			startYear: 800,
			endYear:   300,
			wantCount: 3, // 800, 500, 300
		},
		{
			name:      "recent only",
			startYear: 200,
			endYear:   0,
			wantCount: 1, // 100
		},
		{
			name:      "no events",
			startYear: 50,
			endYear:   0,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := timeline.GetEventsInPeriod(tt.startYear, tt.endYear)
			if len(events) != tt.wantCount {
				t.Errorf("GetEventsInPeriod() returned %d events, want %d", len(events), tt.wantCount)
			}
		})
	}
}

func TestGetCurrentEra(t *testing.T) {
	timeline := &Timeline{
		Eras: []Era{
			{Name: "Ancient", StartYear: 1000, EndYear: 500},
			{Name: "Medieval", StartYear: 500, EndYear: 0},
		},
		CurrentYear: 0,
	}

	era := timeline.GetCurrentEra()
	if era == nil {
		t.Fatal("GetCurrentEra() returned nil")
	}
	if era.Name != "Medieval" {
		t.Errorf("GetCurrentEra() = %s, want Medieval", era.Name)
	}

	// Test with different current year
	timeline.CurrentYear = 750
	era = timeline.GetCurrentEra()
	if era == nil {
		t.Fatal("GetCurrentEra() returned nil for year 750")
	}
	if era.Name != "Ancient" {
		t.Errorf("GetCurrentEra() = %s, want Ancient", era.Name)
	}

	// Test with no matching era
	timeline.CurrentYear = -100
	era = timeline.GetCurrentEra()
	if era != nil {
		t.Errorf("GetCurrentEra() should return nil for future year, got %v", era)
	}
}

func TestGetEventsByType(t *testing.T) {
	timeline := &Timeline{
		Events: []HistoricalEvent{
			{EventType: EventFoundation, Title: "Found 1"},
			{EventType: EventWar, Title: "War 1"},
			{EventType: EventFoundation, Title: "Found 2"},
			{EventType: EventDiscovery, Title: "Disc 1"},
			{EventType: EventWar, Title: "War 2"},
		},
	}

	tests := []struct {
		name      string
		eventType EventType
		wantCount int
	}{
		{
			name:      "foundations",
			eventType: EventFoundation,
			wantCount: 2,
		},
		{
			name:      "wars",
			eventType: EventWar,
			wantCount: 2,
		},
		{
			name:      "discoveries",
			eventType: EventDiscovery,
			wantCount: 1,
		},
		{
			name:      "catastrophes (none)",
			eventType: EventCatastrophe,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := timeline.GetEventsByType(tt.eventType)
			if len(events) != tt.wantCount {
				t.Errorf("GetEventsByType() returned %d events, want %d", len(events), tt.wantCount)
			}
		})
	}
}

func BenchmarkTimelineGenerate(b *testing.B) {
	gen := NewTimelineGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = gen.Generate(int64(i), params)
	}
}

func BenchmarkGetEventsInPeriod(b *testing.B) {
	timeline := &Timeline{
		Events: make([]HistoricalEvent, 100),
	}
	for i := 0; i < 100; i++ {
		timeline.Events[i].Timestamp = int64(1000 - i*10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = timeline.GetEventsInPeriod(800, 200)
	}
}
