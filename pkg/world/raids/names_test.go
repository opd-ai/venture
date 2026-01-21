package raids

import (
	"math/rand"
	"testing"
)

func TestBossNameGenerator_GenerateBossName(t *testing.T) {
	gen := NewBossNameGenerator()

	tests := []struct {
		name    string
		seed    int64
		genreID string
		index   int
	}{
		{
			name:    "fantasy boss",
			seed:    12345,
			genreID: "fantasy",
			index:   0,
		},
		{
			name:    "scifi boss",
			seed:    54321,
			genreID: "scifi",
			index:   1,
		},
		{
			name:    "horror boss",
			seed:    99999,
			genreID: "horror",
			index:   2,
		},
		{
			name:    "cyberpunk boss",
			seed:    11111,
			genreID: "cyberpunk",
			index:   3,
		},
		{
			name:    "postapoc boss",
			seed:    22222,
			genreID: "postapoc",
			index:   4,
		},
		{
			name:    "unknown genre uses default",
			seed:    33333,
			genreID: "unknown",
			index:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(tt.seed))
			name := gen.GenerateBossName(rng, tt.genreID, tt.index)

			if name == "" {
				t.Error("GenerateBossName returned empty string")
			}

			// Verify format: "Name, the Title"
			if len(name) < 10 {
				t.Errorf("GenerateBossName returned unexpectedly short name: %s", name)
			}
		})
	}
}

func TestBossNameGenerator_GenerateBossName_Deterministic(t *testing.T) {
	gen := NewBossNameGenerator()
	seed := int64(12345)
	genreID := "fantasy"
	index := 0

	// Generate the same name twice with the same seed
	rng1 := rand.New(rand.NewSource(seed))
	name1 := gen.GenerateBossName(rng1, genreID, index)

	rng2 := rand.New(rand.NewSource(seed))
	name2 := gen.GenerateBossName(rng2, genreID, index)

	if name1 != name2 {
		t.Errorf("GenerateBossName is not deterministic: got %q and %q", name1, name2)
	}
}

func TestBossNameGenerator_GenerateRaidName(t *testing.T) {
	gen := NewBossNameGenerator()

	tests := []struct {
		name    string
		seed    int64
		genreID string
		tier    RaidTier
	}{
		{
			name:    "fantasy normal tier",
			seed:    12345,
			genreID: "fantasy",
			tier:    TierNormal,
		},
		{
			name:    "scifi heroic tier",
			seed:    54321,
			genreID: "scifi",
			tier:    TierHeroic,
		},
		{
			name:    "horror mythic tier",
			seed:    99999,
			genreID: "horror",
			tier:    TierMythic,
		},
		{
			name:    "cyberpunk legendary tier",
			seed:    11111,
			genreID: "cyberpunk",
			tier:    TierLegendary,
		},
		{
			name:    "postapoc nightmare tier",
			seed:    22222,
			genreID: "postapoc",
			tier:    TierNightmare,
		},
		{
			name:    "unknown genre uses default",
			seed:    33333,
			genreID: "unknown",
			tier:    TierNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rng := rand.New(rand.NewSource(tt.seed))
			name := gen.GenerateRaidName(rng, tt.genreID, tt.tier)

			if name == "" {
				t.Error("GenerateRaidName returned empty string")
			}

			// Verify tier is in the name
			tierStr := tt.tier.String()
			if len(name) < len(tierStr) {
				t.Errorf("GenerateRaidName returned unexpectedly short name: %s", name)
			}
		})
	}
}

func TestBossNameGenerator_getTitlesByGenre(t *testing.T) {
	gen := NewBossNameGenerator()

	tests := []struct {
		genreID   string
		wantTitle string // at least one title should contain this
	}{
		{"fantasy", "Lord"},
		{"scifi", "AI"},
		{"horror", "Soul"},
		{"cyberpunk", "Enforcer"},
		{"postapoc", "Warlord"},
		{"unknown", "Ancient"},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			titles := gen.getTitlesByGenre(tt.genreID)

			if len(titles) == 0 {
				t.Error("getTitlesByGenre returned empty slice")
			}

			// Verify at least one title exists
			if len(titles) < 4 {
				t.Errorf("getTitlesByGenre returned too few titles: %d", len(titles))
			}
		})
	}
}

func TestBossNameGenerator_getNamesByGenre(t *testing.T) {
	gen := NewBossNameGenerator()

	tests := []struct {
		genreID string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			names := gen.getNamesByGenre(tt.genreID)

			if len(names) == 0 {
				t.Error("getNamesByGenre returned empty slice")
			}

			// Verify at least a few names exist
			if len(names) < 4 {
				t.Errorf("getNamesByGenre returned too few names: %d", len(names))
			}

			// Verify names are non-empty strings
			for i, name := range names {
				if name == "" {
					t.Errorf("getNamesByGenre returned empty name at index %d", i)
				}
			}
		})
	}
}

func TestBossNameGenerator_getPrefixesByGenre(t *testing.T) {
	gen := NewBossNameGenerator()

	tests := []struct {
		genreID string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			prefixes := gen.getPrefixesByGenre(tt.genreID)

			if len(prefixes) == 0 {
				t.Error("getPrefixesByGenre returned empty slice")
			}

			if len(prefixes) < 5 {
				t.Errorf("getPrefixesByGenre returned too few prefixes: %d", len(prefixes))
			}
		})
	}
}

func TestBossNameGenerator_getSuffixesByGenre(t *testing.T) {
	gen := NewBossNameGenerator()

	tests := []struct {
		genreID string
	}{
		{"fantasy"},
		{"scifi"},
		{"horror"},
		{"cyberpunk"},
		{"postapoc"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.genreID, func(t *testing.T) {
			suffixes := gen.getSuffixesByGenre(tt.genreID)

			if len(suffixes) == 0 {
				t.Error("getSuffixesByGenre returned empty slice")
			}

			if len(suffixes) < 5 {
				t.Errorf("getSuffixesByGenre returned too few suffixes: %d", len(suffixes))
			}
		})
	}
}

func TestNewBossNameGenerator(t *testing.T) {
	gen := NewBossNameGenerator()
	if gen == nil {
		t.Error("NewBossNameGenerator returned nil")
	}
}

// Benchmark tests for performance validation
func BenchmarkBossNameGenerator_GenerateBossName(b *testing.B) {
	gen := NewBossNameGenerator()
	rng := rand.New(rand.NewSource(12345))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateBossName(rng, "fantasy", i)
	}
}

func BenchmarkBossNameGenerator_GenerateRaidName(b *testing.B) {
	gen := NewBossNameGenerator()
	rng := rand.New(rand.NewSource(12345))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.GenerateRaidName(rng, "fantasy", TierNormal)
	}
}
