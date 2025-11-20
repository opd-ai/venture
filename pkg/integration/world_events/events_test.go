package world_events

import (
	"testing"
	"time"
)

func TestGenerateFactionResponse(t *testing.T) {
	tests := []struct {
		name      string
		seed      int64
		factionID string
		action    string
		severity  Severity
	}{
		{
			name:      "minor action",
			seed:      12345,
			factionID: "faction_a",
			action:    "guild_war",
			severity:  SeverityMinor,
		},
		{
			name:      "major action",
			seed:      67890,
			factionID: "faction_b",
			action:    "market_crash",
			severity:  SeverityMajor,
		},
		{
			name:      "critical action",
			seed:      11111,
			factionID: "faction_c",
			action:    "invasion",
			severity:  SeverityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := GenerateFactionResponse(tt.seed, tt.factionID, tt.action, tt.severity)

			if response.FactionID != tt.factionID {
				t.Errorf("expected faction %s, got %s", tt.factionID, response.FactionID)
			}

			if response.ResponseType == "" {
				t.Error("expected non-empty response type")
			}

			if response.Message == "" {
				t.Error("expected non-empty message")
			}

			response2 := GenerateFactionResponse(tt.seed, tt.factionID, tt.action, tt.severity)
			if response.ResponseType != response2.ResponseType {
				t.Error("expected deterministic response type")
			}
			if response.ReputationChange != response2.ReputationChange {
				t.Error("expected deterministic reputation change")
			}
		})
	}
}

func TestGenerateEconomicEvent(t *testing.T) {
	tests := []struct {
		name     string
		seed     int64
		eventID  string
		itemType string
		severity Severity
	}{
		{
			name:     "minor economic shift",
			seed:     12345,
			eventID:  "event_1",
			itemType: "iron_ore",
			severity: SeverityMinor,
		},
		{
			name:     "major market crash",
			seed:     67890,
			eventID:  "event_2",
			itemType: "gold",
			severity: SeverityMajor,
		},
		{
			name:     "critical shortage",
			seed:     11111,
			eventID:  "event_3",
			itemType: "food",
			severity: SeverityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := GenerateEconomicEvent(tt.seed, tt.eventID, tt.itemType, tt.severity)

			if event.EventID != tt.eventID {
				t.Errorf("expected event ID %s, got %s", tt.eventID, event.EventID)
			}

			if event.ItemType != tt.itemType {
				t.Errorf("expected item type %s, got %s", tt.itemType, event.ItemType)
			}

			if event.PriceModifier == 0 {
				t.Error("expected non-zero price modifier")
			}

			if event.Duration == 0 {
				t.Error("expected non-zero duration")
			}

			if len(event.AffectedZones) == 0 {
				t.Error("expected affected zones")
			}

			event2 := GenerateEconomicEvent(tt.seed, tt.eventID, tt.itemType, tt.severity)
			if event.PriceModifier != event2.PriceModifier {
				t.Error("expected deterministic price modifier")
			}
		})
	}
}

func TestGenerateWeatherDisaster(t *testing.T) {
	tests := []struct {
		name     string
		seed     int64
		centerX  float64
		centerY  float64
		severity Severity
	}{
		{
			name:     "minor storm",
			seed:     12345,
			centerX:  100.0,
			centerY:  100.0,
			severity: SeverityMinor,
		},
		{
			name:     "major hurricane",
			seed:     67890,
			centerX:  200.0,
			centerY:  200.0,
			severity: SeverityMajor,
		},
		{
			name:     "catastrophic disaster",
			seed:     11111,
			centerX:  300.0,
			centerY:  300.0,
			severity: SeverityCritical,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disaster := GenerateWeatherDisaster(tt.seed, tt.centerX, tt.centerY, tt.severity)

			if disaster.DisasterType == "" {
				t.Error("expected non-empty disaster type")
			}

			if disaster.Intensity <= 0 {
				t.Error("expected positive intensity")
			}

			if disaster.Radius <= 0 {
				t.Error("expected positive radius")
			}

			if disaster.CenterX != tt.centerX {
				t.Errorf("expected centerX %f, got %f", tt.centerX, disaster.CenterX)
			}

			if disaster.CenterY != tt.centerY {
				t.Errorf("expected centerY %f, got %f", tt.centerY, disaster.CenterY)
			}

			if disaster.Damage <= 0 {
				t.Error("expected positive damage")
			}

			disaster2 := GenerateWeatherDisaster(tt.seed, tt.centerX, tt.centerY, tt.severity)
			if disaster.DisasterType != disaster2.DisasterType {
				t.Error("expected deterministic disaster type")
			}
		})
	}
}

func TestPropagateEventCrossServer(t *testing.T) {
	originalEvent := &WorldEvent{
		ID:          "event_1",
		Type:        EventGuildWarfare,
		Trigger:     TriggerGuildWar,
		Severity:    SeverityMajor,
		Title:       "Guild War",
		Description: "Major conflict between guilds",
		Location:    "zone_1",
		ServerID:    "server_1",
		StartTime:   time.Now(),
		Duration:    2 * time.Hour,
		Impacts: []Impact{
			{Type: ImpactNPCReputation, Target: "guild_a", Modifier: -0.3},
			{Type: ImpactSpawnRate, Target: "zone_1", Modifier: 1.6},
		},
	}

	targetServers := []string{"server_2", "server_3", "server_4"}
	delay := 30 * time.Second

	propagated := PropagateEventCrossServer(originalEvent, targetServers, delay)

	if len(propagated) != len(targetServers) {
		t.Errorf("expected %d propagated events, got %d", len(targetServers), len(propagated))
	}

	for i, event := range propagated {
		if event.ServerID != targetServers[i] {
			t.Errorf("expected server ID %s, got %s", targetServers[i], event.ServerID)
		}

		if event.Type != EventCrossServer {
			t.Errorf("expected type EventCrossServer, got %s", event.Type)
		}

		if !event.StartTime.After(originalEvent.StartTime) {
			t.Error("expected delayed start time")
		}

		if len(event.Impacts) != len(originalEvent.Impacts) {
			t.Errorf("expected %d impacts, got %d", len(originalEvent.Impacts), len(event.Impacts))
		}

		for j, impact := range event.Impacts {
			originalImpact := originalEvent.Impacts[j]
			if impact.Modifier != originalImpact.Modifier*0.5 {
				t.Errorf("expected reduced modifier %f, got %f", originalImpact.Modifier*0.5, impact.Modifier)
			}
		}
	}
}

func TestCalculateEventFrequency(t *testing.T) {
	tests := []struct {
		name          string
		baseFrequency float64
		activityLevel float64
		want          float64
	}{
		{
			name:          "zero activity",
			baseFrequency: 2.0,
			activityLevel: 0.0,
			want:          2.0,
		},
		{
			name:          "half activity",
			baseFrequency: 2.0,
			activityLevel: 0.5,
			want:          3.0,
		},
		{
			name:          "max activity",
			baseFrequency: 2.0,
			activityLevel: 1.0,
			want:          4.0,
		},
		{
			name:          "negative activity clamped",
			baseFrequency: 2.0,
			activityLevel: -0.5,
			want:          2.0,
		},
		{
			name:          "over max activity clamped",
			baseFrequency: 2.0,
			activityLevel: 1.5,
			want:          4.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateEventFrequency(tt.baseFrequency, tt.activityLevel)
			if got != tt.want {
				t.Errorf("expected frequency %f, got %f", tt.want, got)
			}
		})
	}
}

func TestShouldSpawnEvent(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name          string
		lastEventTime time.Time
		frequency     float64
		want          bool
	}{
		{
			name:          "spawn after interval",
			lastEventTime: now.Add(-2 * time.Hour),
			frequency:     2.0,
			want:          true,
		},
		{
			name:          "too soon",
			lastEventTime: now.Add(-10 * time.Second),
			frequency:     2.0,
			want:          false,
		},
		{
			name:          "high frequency spawn",
			lastEventTime: now.Add(-10 * time.Minute),
			frequency:     10.0,
			want:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldSpawnEvent(tt.lastEventTime, tt.frequency)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestGetAffectedArea(t *testing.T) {
	tests := []struct {
		name        string
		event       *WorldEvent
		minRadius   float64
		expectBonus bool
	}{
		{
			name: "minor event",
			event: &WorldEvent{
				Severity: SeverityMinor,
				Impacts:  []Impact{},
			},
			minRadius:   100.0,
			expectBonus: false,
		},
		{
			name: "weather event with bonus",
			event: &WorldEvent{
				Severity: SeverityMajor,
				Impacts: []Impact{
					{Type: ImpactWeather},
				},
			},
			minRadius:   300.0,
			expectBonus: true,
		},
		{
			name: "terrain event with bonus",
			event: &WorldEvent{
				Severity: SeverityModerate,
				Impacts: []Impact{
					{Type: ImpactTerrain},
				},
			},
			minRadius:   200.0,
			expectBonus: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, radius := GetAffectedArea(tt.event)

			if radius < tt.minRadius {
				t.Errorf("expected radius >= %f, got %f", tt.minRadius, radius)
			}

			expectedRadius := 100.0 * float64(tt.event.Severity)
			if tt.expectBonus {
				expectedRadius *= 1.5
			}

			if radius != expectedRadius {
				t.Errorf("expected radius %f, got %f", expectedRadius, radius)
			}
		})
	}
}

func TestMergeEventImpacts(t *testing.T) {
	events := []*WorldEvent{
		{
			Impacts: []Impact{
				{Type: ImpactPriceChange, Target: "iron_ore", Modifier: 0.1, Duration: 1 * time.Hour},
				{Type: ImpactNPCReputation, Target: "faction_a", Modifier: -0.2, Duration: 2 * time.Hour},
			},
		},
		{
			Impacts: []Impact{
				{Type: ImpactPriceChange, Target: "iron_ore", Modifier: 0.15, Duration: 2 * time.Hour},
				{Type: ImpactSpawnRate, Target: "zone_1", Modifier: 1.2, Duration: 1 * time.Hour},
			},
		},
	}

	merged := MergeEventImpacts(events)

	if len(merged) != 3 {
		t.Errorf("expected 3 merged impacts, got %d", len(merged))
	}

	foundPrice := false
	for _, impact := range merged {
		if impact.Type == ImpactPriceChange && impact.Target == "iron_ore" {
			foundPrice = true
			expected := 0.1 + 0.15
			if impact.Modifier != expected {
				t.Errorf("expected merged modifier %f, got %f", expected, impact.Modifier)
			}
			if impact.Duration != 2*time.Hour {
				t.Errorf("expected max duration 2h, got %v", impact.Duration)
			}
		}
	}

	if !foundPrice {
		t.Error("expected merged price change impact")
	}
}

func BenchmarkGenerateFactionResponse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateFactionResponse(12345, "faction_a", "guild_war", SeverityMajor)
	}
}

func BenchmarkGenerateEconomicEvent(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateEconomicEvent(12345, "event_1", "iron_ore", SeverityMajor)
	}
}

func BenchmarkGenerateWeatherDisaster(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GenerateWeatherDisaster(12345, 100.0, 100.0, SeverityMajor)
	}
}

func BenchmarkPropagateEventCrossServer(b *testing.B) {
	event := &WorldEvent{
		ID:        "event_1",
		Type:      EventGuildWarfare,
		Trigger:   TriggerGuildWar,
		Severity:  SeverityMajor,
		ServerID:  "server_1",
		StartTime: time.Now(),
		Duration:  2 * time.Hour,
		Impacts: []Impact{
			{Type: ImpactNPCReputation, Target: "guild_a", Modifier: -0.3},
		},
	}
	targetServers := []string{"server_2", "server_3", "server_4"}
	delay := 30 * time.Second

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = PropagateEventCrossServer(event, targetServers, delay)
	}
}

func BenchmarkMergeEventImpacts(b *testing.B) {
	events := []*WorldEvent{
		{
			Impacts: []Impact{
				{Type: ImpactPriceChange, Target: "iron_ore", Modifier: 0.1},
				{Type: ImpactNPCReputation, Target: "faction_a", Modifier: -0.2},
			},
		},
		{
			Impacts: []Impact{
				{Type: ImpactPriceChange, Target: "iron_ore", Modifier: 0.15},
				{Type: ImpactSpawnRate, Target: "zone_1", Modifier: 1.2},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = MergeEventImpacts(events)
	}
}
