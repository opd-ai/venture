package world_events

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeProvider(t *testing.T) {
	fixedTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
	defer ResetTimeProvider()

	config := EventManagerConfig{
		MaxActiveEvents:      50,
		EventFrequency:       2.0,
		ChainProbability:     0.0,
		CrossServerPropDelay: 30 * time.Second,
		ResponseTimeMin:      0 * time.Millisecond,
		ResponseTimeMax:      0 * time.Millisecond,
	}

	manager := NewEventManagerWithConfig(12345, config)

	params := TriggerParams{
		TriggerType: TriggerGuildWar,
		Severity:    SeverityMajor,
		Location:    "zone_1",
		ServerID:    "server_1",
		GuildID:     "guild_a",
	}
	event, err := manager.GenerateEvent(TriggerGuildWar, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !event.StartTime.Equal(fixedTime) {
		t.Errorf("expected StartTime %v, got %v", fixedTime, event.StartTime)
	}

	// Verify determinism: same seed + same TimeProvider = same result
	manager2 := NewEventManagerWithConfig(12345, config)
	event2, err := manager2.GenerateEvent(TriggerGuildWar, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !event2.StartTime.Equal(event.StartTime) {
		t.Errorf("expected deterministic StartTime, got %v vs %v", event.StartTime, event2.StartTime)
	}
}

func TestRealTimeProvider(t *testing.T) {
	rtp := RealTimeProvider{}
	before := time.Now()
	ts := rtp.Now()
	after := time.Now()

	if ts.Before(before) || ts.After(after) {
		t.Errorf("RealTimeProvider.Now() = %v, expected between %v and %v", ts, before, after)
	}
}

func TestFixedTimeProvider(t *testing.T) {
	fixedTime := time.Date(2026, 6, 15, 10, 30, 0, 0, time.UTC)
	ftp := FixedTimeProvider{FixedTime: fixedTime}

	if !ftp.Now().Equal(fixedTime) {
		t.Errorf("FixedTimeProvider.Now() = %v, expected %v", ftp.Now(), fixedTime)
	}

	// Calling multiple times returns same value
	if !ftp.Now().Equal(ftp.Now()) {
		t.Error("FixedTimeProvider should return consistent values")
	}
}

func TestEventDeterminismWithFixedTime(t *testing.T) {
	fixedTime := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	SetTimeProvider(FixedTimeProvider{FixedTime: fixedTime})
	defer ResetTimeProvider()

	config := EventManagerConfig{
		MaxActiveEvents:      50,
		EventFrequency:       2.0,
		ChainProbability:     1.0, // Force chains
		CrossServerPropDelay: 30 * time.Second,
		ResponseTimeMin:      0 * time.Millisecond,
		ResponseTimeMax:      0 * time.Millisecond,
	}

	params := TriggerParams{
		TriggerType: TriggerGuildWar,
		Severity:    SeverityMajor,
		Location:    "zone_1",
		ServerID:    "server_1",
		GuildID:     "guild_a",
	}

	// Generate events with two managers using the same seed
	m1 := NewEventManagerWithConfig(99999, config)
	e1, _ := m1.GenerateEvent(TriggerGuildWar, params)

	m2 := NewEventManagerWithConfig(99999, config)
	e2, _ := m2.GenerateEvent(TriggerGuildWar, params)

	if !e1.StartTime.Equal(e2.StartTime) {
		t.Errorf("StartTime not deterministic: %v vs %v", e1.StartTime, e2.StartTime)
	}

	if len(e1.ChainEvents) != len(e2.ChainEvents) {
		t.Errorf("chain length not deterministic: %d vs %d", len(e1.ChainEvents), len(e2.ChainEvents))
	}

	// Verify chain event start times are deterministic
	for i, chainID := range e1.ChainEvents {
		ce1, _ := m1.GetEvent(chainID)
		ce2, _ := m2.GetEvent(e2.ChainEvents[i])
		if ce1 != nil && ce2 != nil && !ce1.StartTime.Equal(ce2.StartTime) {
			t.Errorf("chain event %d StartTime not deterministic: %v vs %v", i, ce1.StartTime, ce2.StartTime)
		}
	}
}

func TestEventManagerCreation(t *testing.T) {
	tests := []struct {
		name   string
		seed   int64
		config EventManagerConfig
	}{
		{
			name:   "default config",
			seed:   12345,
			config: NewDefaultEventManagerConfig(),
		},
		{
			name: "custom config",
			seed: 67890,
			config: EventManagerConfig{
				MaxActiveEvents:      100,
				EventFrequency:       3.0,
				ChainProbability:     0.5,
				CrossServerPropDelay: 60 * time.Second,
				ResponseTimeMin:      30 * time.Second,
				ResponseTimeMax:      10 * time.Minute,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewEventManagerWithConfig(tt.seed, tt.config)
			if manager == nil {
				t.Fatal("expected non-nil manager")
			}
			if manager.seed != tt.seed {
				t.Errorf("expected seed %d, got %d", tt.seed, manager.seed)
			}
			if manager.config.MaxActiveEvents != tt.config.MaxActiveEvents {
				t.Errorf("expected MaxActiveEvents %d, got %d", tt.config.MaxActiveEvents, manager.config.MaxActiveEvents)
			}
		})
	}
}

func TestGenerateEvent(t *testing.T) {
	tests := []struct {
		name        string
		trigger     TriggerType
		params      TriggerParams
		wantErr     bool
		wantType    EventType
		wantChained bool
	}{
		{
			name:    "guild war event",
			trigger: TriggerGuildWar,
			params: TriggerParams{
				TriggerType: TriggerGuildWar,
				Severity:    SeverityMajor,
				Location:    "zone_1",
				ServerID:    "server_1",
				GuildID:     "guild_a",
			},
			wantErr:  false,
			wantType: EventGuildWarfare,
		},
		{
			name:    "economic event",
			trigger: TriggerTradeVolume,
			params: TriggerParams{
				TriggerType: TriggerTradeVolume,
				Severity:    SeverityModerate,
				Location:    "market_square",
				ServerID:    "server_2",
				ItemType:    "iron_ore",
			},
			wantErr:  false,
			wantType: EventEconomic,
		},
		{
			name:    "weather disaster",
			trigger: TriggerWeatherChange,
			params: TriggerParams{
				TriggerType: TriggerWeatherChange,
				Severity:    SeverityCritical,
				Location:    "coastal_region",
				ServerID:    "server_3",
			},
			wantErr:  false,
			wantType: EventWeatherDisaster,
		},
		{
			name:    "missing location",
			trigger: TriggerGuildWar,
			params: TriggerParams{
				TriggerType: TriggerGuildWar,
				Severity:    SeverityMinor,
				ServerID:    "server_1",
			},
			wantErr: true,
		},
		{
			name:    "invalid severity",
			trigger: TriggerGuildWar,
			params: TriggerParams{
				TriggerType: TriggerGuildWar,
				Severity:    5,
				Location:    "zone_1",
				ServerID:    "server_1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewEventManager(12345)
			event, err := manager.GenerateEvent(tt.trigger, tt.params)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if event.Type != tt.wantType {
				t.Errorf("expected type %s, got %s", tt.wantType, event.Type)
			}

			if event.Severity != tt.params.Severity {
				t.Errorf("expected severity %d, got %d", tt.params.Severity, event.Severity)
			}

			if event.Location != tt.params.Location {
				t.Errorf("expected location %s, got %s", tt.params.Location, event.Location)
			}

			if len(event.Impacts) == 0 {
				t.Error("expected impacts, got none")
			}
		})
	}
}

func TestEventChainGeneration(t *testing.T) {
	manager := NewEventManagerWithConfig(12345, EventManagerConfig{
		MaxActiveEvents:      50,
		EventFrequency:       2.0,
		ChainProbability:     1.0,
		CrossServerPropDelay: 30 * time.Second,
		ResponseTimeMin:      1 * time.Minute,
		ResponseTimeMax:      5 * time.Minute,
	})

	params := TriggerParams{
		TriggerType: TriggerGuildWar,
		Severity:    SeverityMajor,
		Location:    "zone_1",
		ServerID:    "server_1",
		GuildID:     "guild_a",
	}

	event, err := manager.GenerateEvent(TriggerGuildWar, params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(event.ChainEvents) == 0 {
		t.Error("expected chain events with probability 1.0")
	}

	chain := manager.GetEventChain(event.ID)
	if len(chain) < 2 {
		t.Errorf("expected chain length >= 2, got %d", len(chain))
	}
}

func TestGetActiveEvents(t *testing.T) {
	config := EventManagerConfig{
		MaxActiveEvents:      50,
		EventFrequency:       2.0,
		ChainProbability:     0.0,
		CrossServerPropDelay: 30 * time.Second,
		ResponseTimeMin:      0 * time.Millisecond,
		ResponseTimeMax:      0 * time.Millisecond,
	}
	manager := NewEventManagerWithConfig(12345, config)

	params1 := TriggerParams{
		TriggerType: TriggerGuildWar,
		Severity:    SeverityMinor,
		Location:    "zone_1",
		ServerID:    "server_1",
	}
	event1, _ := manager.GenerateEvent(TriggerGuildWar, params1)

	params2 := TriggerParams{
		TriggerType: TriggerTradeVolume,
		Severity:    SeverityModerate,
		Location:    "zone_2",
		ServerID:    "server_1",
		ItemType:    "gold",
	}
	event2, _ := manager.GenerateEvent(TriggerTradeVolume, params2)

	currentTime := time.Now().Add(5 * time.Second)
	activeEvents := manager.GetActiveEvents(currentTime)
	if len(activeEvents) != 2 {
		t.Errorf("expected 2 active events, got %d", len(activeEvents))
	}

	_ = event1
	_ = event2
}

func TestCleanupExpiredEvents(t *testing.T) {
	manager := NewEventManager(12345)

	params := TriggerParams{
		TriggerType: TriggerGuildWar,
		Severity:    SeverityMinor,
		Location:    "zone_1",
		ServerID:    "server_1",
	}
	event, _ := manager.GenerateEvent(TriggerGuildWar, params)

	event.StartTime = time.Now().Add(-2 * time.Hour)
	event.Duration = 1 * time.Hour

	currentTime := time.Now()
	removed := manager.CleanupExpiredEvents(currentTime)

	if removed != 1 {
		t.Errorf("expected 1 removed event, got %d", removed)
	}

	activeEvents := manager.GetActiveEvents(currentTime)
	if len(activeEvents) != 0 {
		t.Errorf("expected 0 active events after cleanup, got %d", len(activeEvents))
	}
}

func TestMaxActiveEvents(t *testing.T) {
	config := EventManagerConfig{
		MaxActiveEvents:      2,
		EventFrequency:       2.0,
		ChainProbability:     0.0,
		CrossServerPropDelay: 30 * time.Second,
		ResponseTimeMin:      1 * time.Minute,
		ResponseTimeMax:      5 * time.Minute,
	}
	manager := NewEventManagerWithConfig(12345, config)

	params := TriggerParams{
		TriggerType: TriggerGuildWar,
		Severity:    SeverityMinor,
		Location:    "zone_1",
		ServerID:    "server_1",
	}

	_, err := manager.GenerateEvent(TriggerGuildWar, params)
	if err != nil {
		t.Fatalf("first event failed: %v", err)
	}

	_, err = manager.GenerateEvent(TriggerGuildWar, params)
	if err != nil {
		t.Fatalf("second event failed: %v", err)
	}

	_, err = manager.GenerateEvent(TriggerGuildWar, params)
	if err == nil {
		t.Error("expected error when exceeding max events, got nil")
	}
}

func TestEventIsActive(t *testing.T) {
	event := &WorldEvent{
		StartTime: time.Now().Add(-1 * time.Hour),
		Duration:  2 * time.Hour,
		Permanent: false,
	}

	if !event.IsActive(time.Now()) {
		t.Error("event should be active")
	}

	event.StartTime = time.Now().Add(-3 * time.Hour)
	if event.IsActive(time.Now()) {
		t.Error("event should not be active")
	}

	event.Permanent = true
	if !event.IsActive(time.Now()) {
		t.Error("permanent event should always be active")
	}
}

func TestTriggerParamsValidation(t *testing.T) {
	tests := []struct {
		name    string
		params  TriggerParams
		wantErr bool
	}{
		{
			name: "valid params",
			params: TriggerParams{
				TriggerType: TriggerGuildWar,
				Severity:    SeverityMajor,
				Location:    "zone_1",
				ServerID:    "server_1",
			},
			wantErr: false,
		},
		{
			name: "missing trigger type",
			params: TriggerParams{
				Severity: SeverityMajor,
				Location: "zone_1",
				ServerID: "server_1",
			},
			wantErr: true,
		},
		{
			name: "invalid severity",
			params: TriggerParams{
				TriggerType: TriggerGuildWar,
				Severity:    0,
				Location:    "zone_1",
				ServerID:    "server_1",
			},
			wantErr: true,
		},
		{
			name: "missing location",
			params: TriggerParams{
				TriggerType: TriggerGuildWar,
				Severity:    SeverityMajor,
				ServerID:    "server_1",
			},
			wantErr: true,
		},
		{
			name: "missing server id",
			params: TriggerParams{
				TriggerType: TriggerGuildWar,
				Severity:    SeverityMajor,
				Location:    "zone_1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.params.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetEventStats(t *testing.T) {
	config := EventManagerConfig{
		MaxActiveEvents:      50,
		EventFrequency:       2.0,
		ChainProbability:     0.0,
		CrossServerPropDelay: 30 * time.Second,
		ResponseTimeMin:      0 * time.Millisecond,
		ResponseTimeMax:      0 * time.Millisecond,
	}
	manager := NewEventManagerWithConfig(12345, config)

	params1 := TriggerParams{
		TriggerType: TriggerGuildWar,
		Severity:    SeverityMinor,
		Location:    "zone_1",
		ServerID:    "server_1",
	}
	manager.GenerateEvent(TriggerGuildWar, params1)

	params2 := TriggerParams{
		TriggerType: TriggerTradeVolume,
		Severity:    SeverityModerate,
		Location:    "zone_2",
		ServerID:    "server_1",
		ItemType:    "gold",
	}
	manager.GenerateEvent(TriggerTradeVolume, params2)

	stats := manager.GetStats()

	activeEvents, ok := stats["active_events"].(int)
	if !ok || activeEvents != 2 {
		t.Errorf("expected 2 active events, got %v", stats["active_events"])
	}

	typeCounts, ok := stats["type_counts"].(map[EventType]int)
	if !ok {
		t.Fatal("expected type_counts to be map[EventType]int")
	}

	if typeCounts[EventGuildWarfare] != 1 {
		t.Errorf("expected 1 guild warfare event, got %d", typeCounts[EventGuildWarfare])
	}

	if typeCounts[EventEconomic] != 1 {
		t.Errorf("expected 1 economic event, got %d", typeCounts[EventEconomic])
	}
}

func TestEventUpdate(t *testing.T) {
	manager := NewEventManagerWithConfig(12345, EventManagerConfig{
		MaxActiveEvents:      50,
		EventFrequency:       2.0,
		ChainProbability:     1.0,
		CrossServerPropDelay: 30 * time.Second,
		ResponseTimeMin:      1 * time.Millisecond,
		ResponseTimeMax:      2 * time.Millisecond,
	})

	params := TriggerParams{
		TriggerType: TriggerGuildWar,
		Severity:    SeverityMinor,
		Location:    "zone_1",
		ServerID:    "server_1",
	}
	event, _ := manager.GenerateEvent(TriggerGuildWar, params)

	event.Duration = 1 * time.Millisecond
	time.Sleep(10 * time.Millisecond)

	manager.Update(0.016)

	chain := manager.GetEventChain(event.ID)
	if len(chain) < 2 {
		t.Error("expected event chain progression")
	}
}

func TestEventChainedImpacts(t *testing.T) {
	// Verify that EventChained events now generate proper impacts
	manager := NewEventManagerWithConfig(12345, EventManagerConfig{
		MaxActiveEvents:      50,
		EventFrequency:       2.0,
		ChainProbability:     1.0, // Force chain generation
		CrossServerPropDelay: 30 * time.Second,
		ResponseTimeMin:      0 * time.Millisecond,
		ResponseTimeMax:      0 * time.Millisecond,
	})

	params := TriggerParams{
		TriggerType: TriggerPlayerChoice,
		Severity:    SeverityMajor,
		Location:    "zone_1",
		ServerID:    "server_1",
		PlayerID:    "player_123",
	}
	event, err := manager.GenerateEvent(TriggerPlayerChoice, params)
	if err != nil {
		t.Fatalf("failed to generate event: %v", err)
	}

	// TriggerPlayerChoice generates EventChained type
	if event.Type != EventChained {
		t.Errorf("expected EventChained type, got %s", event.Type)
	}

	// EventChained should now have impacts
	if len(event.Impacts) == 0 {
		t.Error("expected EventChained to have impacts, got none")
	}

	// Verify impacts are properly generated
	hasReputationImpact := false
	hasSpawnRateImpact := false
	for _, impact := range event.Impacts {
		if impact.Type == ImpactNPCReputation {
			hasReputationImpact = true
			if impact.Target != "player_123" {
				t.Errorf("expected reputation impact target 'player_123', got '%s'", impact.Target)
			}
		}
		if impact.Type == ImpactSpawnRate {
			hasSpawnRateImpact = true
			if impact.Target != "zone_1" {
				t.Errorf("expected spawn rate impact target 'zone_1', got '%s'", impact.Target)
			}
		}
	}

	if !hasReputationImpact {
		t.Error("expected reputation impact for chained event")
	}
	if !hasSpawnRateImpact {
		t.Error("expected spawn rate impact for chained event")
	}
}

func TestEventWithCoordinatesViaManager(t *testing.T) {
	manager := NewEventManagerWithConfig(12345, EventManagerConfig{
		MaxActiveEvents:      50,
		EventFrequency:       2.0,
		ChainProbability:     0.0,
		CrossServerPropDelay: 30 * time.Second,
		ResponseTimeMin:      0 * time.Millisecond,
		ResponseTimeMax:      0 * time.Millisecond,
	})

	params := TriggerParams{
		TriggerType: TriggerWeatherChange,
		Severity:    SeverityMajor,
		Location:    "zone_1",
		ServerID:    "server_1",
		CenterX:     150.5,
		CenterY:     275.3,
	}
	event, err := manager.GenerateEvent(TriggerWeatherChange, params)
	if err != nil {
		t.Fatalf("failed to generate event: %v", err)
	}

	// Verify coordinates are properly stored in the event
	if event.CenterX != 150.5 {
		t.Errorf("expected CenterX 150.5, got %f", event.CenterX)
	}
	if event.CenterY != 275.3 {
		t.Errorf("expected CenterY 275.3, got %f", event.CenterY)
	}

	// Verify GetAffectedArea returns the coordinates
	centerX, centerY, _ := GetAffectedArea(event)
	if centerX != 150.5 {
		t.Errorf("expected centerX 150.5 from GetAffectedArea, got %f", centerX)
	}
	if centerY != 275.3 {
		t.Errorf("expected centerY 275.3 from GetAffectedArea, got %f", centerY)
	}
}

func BenchmarkGenerateEvent(b *testing.B) {
	manager := NewEventManager(12345)
	params := TriggerParams{
		TriggerType: TriggerGuildWar,
		Severity:    SeverityMajor,
		Location:    "zone_1",
		ServerID:    "server_1",
		GuildID:     "guild_a",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		params.Location = fmt.Sprintf("zone_%d", i%100)
		_, _ = manager.GenerateEvent(TriggerGuildWar, params)
		if i%100 == 0 {
			manager.CleanupExpiredEvents(time.Now())
		}
	}
}

func BenchmarkGetActiveEvents(b *testing.B) {
	manager := NewEventManager(12345)
	params := TriggerParams{
		TriggerType: TriggerGuildWar,
		Severity:    SeverityMajor,
		Location:    "zone_1",
		ServerID:    "server_1",
	}

	for i := 0; i < 50; i++ {
		params.Location = fmt.Sprintf("zone_%d", i)
		manager.GenerateEvent(TriggerGuildWar, params)
	}

	currentTime := time.Now()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = manager.GetActiveEvents(currentTime)
	}
}

func BenchmarkEventIsActive(b *testing.B) {
	event := &WorldEvent{
		StartTime: time.Now().Add(-1 * time.Hour),
		Duration:  2 * time.Hour,
		Permanent: false,
	}
	currentTime := time.Now()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = event.IsActive(currentTime)
	}
}
