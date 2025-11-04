package procgen

import "testing"

func TestSeedGenerator(t *testing.T) {
	sg := NewSeedGenerator(12345)

	// Test deterministic seed generation
	seed1 := sg.GetSeed("terrain", 0)
	seed2 := sg.GetSeed("terrain", 0)

	if seed1 != seed2 {
		t.Error("Expected deterministic seed generation")
	}

	// Test different categories produce different seeds
	terrainSeed := sg.GetSeed("terrain", 0)
	entitySeed := sg.GetSeed("entity", 0)

	if terrainSeed == entitySeed {
		t.Error("Expected different seeds for different categories")
	}

	// Test different indices produce different seeds
	index0 := sg.GetSeed("terrain", 0)
	index1 := sg.GetSeed("terrain", 1)

	if index0 == index1 {
		t.Error("Expected different seeds for different indices")
	}
}

func TestGenerationParams(t *testing.T) {
	params := GenerationParams{
		Difficulty: 0.5,
		Depth:      10,
		GenreID:    "fantasy",
		Custom:     make(map[string]interface{}),
	}

	params.Custom["test"] = "value"

	if params.Custom["test"] != "value" {
		t.Error("Custom parameters not working")
	}
}

func TestValidateDepth(t *testing.T) {
	tests := []struct {
		name    string
		depth   int
		wantErr bool
	}{
		{"valid depth 0", 0, false},
		{"valid depth 5", 5, false},
		{"valid depth 100", 100, false},
		{"negative depth", -1, true},
		{"negative depth -10", -10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDepth(tt.depth)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDepth(%d) error = %v, wantErr %v", tt.depth, err, tt.wantErr)
			}
		})
	}
}

func TestValidateDifficulty(t *testing.T) {
	tests := []struct {
		name       string
		difficulty float64
		wantErr    bool
	}{
		{"valid difficulty 0", 0.0, false},
		{"valid difficulty 0.5", 0.5, false},
		{"valid difficulty 1", 1.0, false},
		{"negative difficulty", -0.1, true},
		{"too high difficulty", 1.1, true},
		{"negative difficulty -1", -1.0, true},
		{"too high difficulty 2", 2.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDifficulty(tt.difficulty)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDifficulty(%f) error = %v, wantErr %v", tt.difficulty, err, tt.wantErr)
			}
		})
	}
}

func TestValidateParams(t *testing.T) {
	tests := []struct {
		name    string
		params  GenerationParams
		wantErr bool
	}{
		{
			name: "valid params",
			params: GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: false,
		},
		{
			name: "invalid difficulty negative",
			params: GenerationParams{
				Difficulty: -0.1,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty too high",
			params: GenerationParams{
				Difficulty: 1.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid depth negative",
			params: GenerationParams{
				Difficulty: 0.5,
				Depth:      -1,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "empty genre ID",
			params: GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "",
			},
			wantErr: true,
		},
		{
			name: "all edge cases valid",
			params: GenerationParams{
				Difficulty: 0.0,
				Depth:      0,
				GenreID:    "scifi",
			},
			wantErr: false,
		},
		{
			name: "all edge cases max valid",
			params: GenerationParams{
				Difficulty: 1.0,
				Depth:      1000,
				GenreID:    "horror",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateParams(tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateParams() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

