package raids

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestGenerator_Generate(t *testing.T) {
	tests := []struct {
		name       string
		seed       int64
		params     procgen.GenerationParams
		wantBosses int
		wantErr    bool
	}{
		{
			name: "normal tier raid",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      10,
				GenreID:    "fantasy",
				Custom: map[string]interface{}{
					"tier":       TierNormal,
					"group_id":   "group-1",
					"group_size": 5,
				},
			},
			wantBosses: 3,
			wantErr:    false,
		},
		{
			name: "heroic tier raid",
			seed: 67890,
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      15,
				GenreID:    "scifi",
				Custom: map[string]interface{}{
					"tier":       TierHeroic,
					"group_id":   "group-2",
					"group_size": 6,
				},
			},
			wantBosses: 4,
			wantErr:    false,
		},
		{
			name: "mythic tier raid",
			seed: 11111,
			params: procgen.GenerationParams{
				Difficulty: 0.9,
				Depth:      20,
				GenreID:    "horror",
				Custom: map[string]interface{}{
					"tier":       TierMythic,
					"group_id":   "group-3",
					"group_size": 8,
				},
			},
			wantBosses: 5,
			wantErr:    false,
		},
		{
			name: "invalid difficulty",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 1.5,
				Depth:      10,
				GenreID:    "fantasy",
			},
			wantBosses: 0,
			wantErr:    true,
		},
		{
			name: "invalid depth",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      0,
				GenreID:    "fantasy",
			},
			wantBosses: 0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := NewGenerator(999)
			result, err := gen.Generate(tt.seed, tt.params)

			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			raid, ok := result.(*RaidDungeon)
			if !ok {
				t.Fatalf("Generate() returned %T, want *RaidDungeon", result)
			}

			if len(raid.Bosses) != tt.wantBosses {
				t.Errorf("Generate() boss count = %d, want %d", len(raid.Bosses), tt.wantBosses)
			}

			if raid.Terrain == nil {
				t.Error("Generate() terrain is nil")
			}

			if len(raid.Rooms) < 4 {
				t.Errorf("Generate() room count = %d, want >= 4", len(raid.Rooms))
			}
		})
	}
}

func TestGenerator_Validate(t *testing.T) {
	gen := NewGenerator(999)

	t.Run("valid raid", func(t *testing.T) {
		params := procgen.GenerationParams{
			Difficulty: 0.5,
			Depth:      10,
			GenreID:    "fantasy",
			Custom: map[string]interface{}{
				"tier":     TierNormal,
				"group_id": "test-group",
			},
		}

		result, err := gen.Generate(12345, params)
		if err != nil {
			t.Fatalf("Generate() failed: %v", err)
		}

		if err := gen.Validate(result); err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		err := gen.Validate("not a raid")
		if err == nil {
			t.Error("Validate() expected error for invalid type")
		}
	})
}

func TestGenerator_Determinism(t *testing.T) {
	gen1 := NewGenerator(999)
	gen2 := NewGenerator(999)

	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      12,
		GenreID:    "cyberpunk",
		Custom: map[string]interface{}{
			"tier":       TierHeroic,
			"group_id":   "determinism-test",
			"group_size": 7,
		},
	}

	seed := int64(54321)

	result1, err1 := gen1.Generate(seed, params)
	if err1 != nil {
		t.Fatalf("Generate() first call failed: %v", err1)
	}

	result2, err2 := gen2.Generate(seed, params)
	if err2 != nil {
		t.Fatalf("Generate() second call failed: %v", err2)
	}

	raid1 := result1.(*RaidDungeon)
	raid2 := result2.(*RaidDungeon)

	if raid1.Name != raid2.Name {
		t.Errorf("Raid names differ: %q vs %q", raid1.Name, raid2.Name)
	}

	if len(raid1.Bosses) != len(raid2.Bosses) {
		t.Errorf("Boss counts differ: %d vs %d", len(raid1.Bosses), len(raid2.Bosses))
	}

	for i := 0; i < len(raid1.Bosses) && i < len(raid2.Bosses); i++ {
		b1 := raid1.Bosses[i]
		b2 := raid2.Bosses[i]

		if len(b1.Mechanics) != len(b2.Mechanics) {
			t.Errorf("Boss %d mechanic counts differ: %d vs %d", i, len(b1.Mechanics), len(b2.Mechanics))
		}

		if b1.Entity.Stats.Health != b2.Entity.Stats.Health {
			t.Errorf("Boss %d health differs: %d vs %d", i, b1.Entity.Stats.Health, b2.Entity.Stats.Health)
		}
	}
}

func TestRaidTier_Methods(t *testing.T) {
	tests := []struct {
		tier        RaidTier
		wantString  string
		wantDiff    float64
		wantMinPlrs int
		wantMaxPlrs int
	}{
		{TierNormal, "Normal", 2.0, 5, 8},
		{TierHeroic, "Heroic", 4.0, 6, 9},
		{TierMythic, "Mythic", 6.0, 7, 10},
		{TierLegendary, "Legendary", 8.0, 8, 10},
		{TierNightmare, "Nightmare", 10.0, 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.wantString, func(t *testing.T) {
			if got := tt.tier.String(); got != tt.wantString {
				t.Errorf("String() = %q, want %q", got, tt.wantString)
			}
			if got := tt.tier.DifficultyMultiplier(); got != tt.wantDiff {
				t.Errorf("DifficultyMultiplier() = %.1f, want %.1f", got, tt.wantDiff)
			}
			if got := tt.tier.MinPlayers(); got != tt.wantMinPlrs {
				t.Errorf("MinPlayers() = %d, want %d", got, tt.wantMinPlrs)
			}
			if got := tt.tier.MaxPlayers(); got != tt.wantMaxPlrs {
				t.Errorf("MaxPlayers() = %d, want %d", got, tt.wantMaxPlrs)
			}
		})
	}
}

func TestMechanicType_String(t *testing.T) {
	tests := []struct {
		mechanic MechanicType
		want     string
	}{
		{MechanicSummon, "Summon"},
		{MechanicGroundEffect, "GroundEffect"},
		{MechanicDebuff, "Debuff"},
		{MechanicBuff, "Buff"},
		{MechanicChanneled, "Channeled"},
		{MechanicInstant, "Instant"},
		{MechanicPeriodic, "Periodic"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mechanic.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRoomType_String(t *testing.T) {
	tests := []struct {
		room RoomType
		want string
	}{
		{RoomEntrance, "Entrance"},
		{RoomBoss, "Boss"},
		{RoomTrash, "Trash"},
		{RoomTreasure, "Treasure"},
		{RoomPuzzle, "Puzzle"},
		{RoomRest, "Rest"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.room.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestHashString(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"", 0},
		{"a", 97},
		{"test", 3556498},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := hashString(tt.input); got != tt.want {
				t.Errorf("hashString(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func BenchmarkGenerator_Generate(b *testing.B) {
	gen := NewGenerator(999)
	params := procgen.GenerationParams{
		Difficulty: 0.7,
		Depth:      15,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"tier":       TierHeroic,
			"group_id":   "bench-group",
			"group_size": 8,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i), params)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerator_Validate(b *testing.B) {
	gen := NewGenerator(999)
	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      12,
		GenreID:    "scifi",
		Custom: map[string]interface{}{
			"tier":     TierMythic,
			"group_id": "bench-validate",
		},
	}

	raid, err := gen.Generate(12345, params)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := gen.Validate(raid); err != nil {
			b.Fatal(err)
		}
	}
}
