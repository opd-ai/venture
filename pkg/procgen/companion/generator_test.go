package companion

import (
	"testing"
	
	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
)

func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator()
	
	tests := []struct {
		name       string
		seed       int64
		params     procgen.GenerationParams
		wantErr    bool
	}{
		{
			name: "valid generation",
			seed: 12345,
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: false,
		},
		{
			name: "sci-fi genre",
			seed: 67890,
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      10,
				GenreID:    "sci-fi",
			},
			wantErr: false,
		},
		{
			name: "invalid difficulty",
			seed: 99999,
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
			result, err := gen.Generate(tt.seed, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !tt.wantErr {
				companion, ok := result.(*Companion)
				if !ok {
					t.Fatal("Result is not a Companion")
				}
				
				if companion.Attack <= 0 {
					t.Errorf("Invalid attack: %f", companion.Attack)
				}
				
				if companion.MaxHP <= 0 {
					t.Errorf("Invalid MaxHP: %f", companion.MaxHP)
				}
				
				if len(companion.Commands) == 0 {
					t.Error("Companion has no commands")
				}
			}
		})
	}
}

func TestGenerator_Validate(t *testing.T) {
	gen := NewGenerator()
	
	tests := []struct {
		name      string
		companion *Companion
		wantErr   bool
	}{
		{
			name: "valid companion",
			companion: &Companion{
				Name:     "Test",
				Attack:   10.0,
				Defense:  8.0,
				MaxHP:    50.0,
				Loyalty:  50.0,
				Commands: []engine.CommandType{engine.CommandFollow},
			},
			wantErr: false,
		},
		{
			name: "invalid attack",
			companion: &Companion{
				Name:     "Test",
				Attack:   -5.0,
				Defense:  8.0,
				MaxHP:    50.0,
				Loyalty:  50.0,
				Commands: []engine.CommandType{engine.CommandFollow},
			},
			wantErr: true,
		},
		{
			name: "no commands",
			companion: &Companion{
				Name:     "Test",
				Attack:   10.0,
				Defense:  8.0,
				MaxHP:    50.0,
				Loyalty:  50.0,
				Commands: []engine.CommandType{},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.companion)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerator_Determinism(t *testing.T) {
	gen := NewGenerator()
	seed := int64(12345)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}
	
	result1, err1 := gen.Generate(seed, params)
	result2, err2 := gen.Generate(seed, params)
	
	if err1 != nil || err2 != nil {
		t.Fatalf("Generation failed: err1=%v, err2=%v", err1, err2)
	}
	
	comp1 := result1.(*Companion)
	comp2 := result2.(*Companion)
	
	if comp1.Name != comp2.Name {
		t.Errorf("Name differs: %s vs %s", comp1.Name, comp2.Name)
	}
	
	if comp1.Attack != comp2.Attack {
		t.Errorf("Attack differs: %f vs %f", comp1.Attack, comp2.Attack)
	}
	
	if comp1.Type != comp2.Type {
		t.Errorf("Type differs: %v vs %v", comp1.Type, comp2.Type)
	}
}
