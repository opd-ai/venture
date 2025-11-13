package minigame

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestGenerator(t *testing.T) {
	gen := NewGenerator()
	seed := int64(12345)

	tests := []struct {
		name       string
		params     procgen.GenerationParams
		wantErr    bool
		checkState bool
	}{
		{
			name: "valid fantasy card game",
			params: procgen.GenerationParams{
				Difficulty: 0.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: false,
		},
		{
			name: "valid scifi hacking game",
			params: procgen.GenerationParams{
				Difficulty: 0.7,
				Depth:      8,
				GenreID:    "scifi",
			},
			wantErr: false,
		},
		{
			name: "invalid difficulty too low",
			params: procgen.GenerationParams{
				Difficulty: -0.1,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid difficulty too high",
			params: procgen.GenerationParams{
				Difficulty: 1.5,
				Depth:      5,
				GenreID:    "fantasy",
			},
			wantErr: true,
		},
		{
			name: "easy difficulty",
			params: procgen.GenerationParams{
				Difficulty: 0.2,
				Depth:      1,
				GenreID:    "fantasy",
			},
			wantErr: false,
		},
		{
			name: "hard difficulty",
			params: procgen.GenerationParams{
				Difficulty: 0.9,
				Depth:      10,
				GenreID:    "horror",
			},
			wantErr: false,
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

			game, ok := result.(*MiniGame)
			if !ok {
				t.Errorf("Generate() returned wrong type, want *MiniGame")
				return
			}

			// Validate basic fields
			if game.Name == "" {
				t.Errorf("Game name is empty")
			}

			if game.Difficulty != tt.params.Difficulty {
				t.Errorf("Difficulty = %.2f, want %.2f", game.Difficulty, tt.params.Difficulty)
			}

			if game.TimeLimit <= 0 {
				t.Errorf("TimeLimit = %.2f, want > 0", game.TimeLimit)
			}

			if game.Rules == "" {
				t.Errorf("Rules are empty")
			}

			// Validate state is not nil
			if game.State == nil {
				t.Errorf("State is nil")
			}
		})
	}
}

func TestValidate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		game    *MiniGame
		wantErr bool
	}{
		{
			name: "valid game",
			game: &MiniGame{
				Type:       GameTypeCard,
				Name:       "Test Game",
				Difficulty: 0.5,
				TimeLimit:  300.0,
				Rules:      "Test rules",
				State:      &CardGameState{},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			game: &MiniGame{
				Type:       GameTypeCard,
				Name:       "",
				Difficulty: 0.5,
				TimeLimit:  300.0,
				Rules:      "Test rules",
				State:      &CardGameState{},
			},
			wantErr: true,
		},
		{
			name: "zero time limit",
			game: &MiniGame{
				Type:       GameTypeCard,
				Name:       "Test Game",
				Difficulty: 0.5,
				TimeLimit:  0,
				Rules:      "Test rules",
				State:      &CardGameState{},
			},
			wantErr: true,
		},
		{
			name: "negative time limit",
			game: &MiniGame{
				Type:       GameTypeCard,
				Name:       "Test Game",
				Difficulty: 0.5,
				TimeLimit:  -10.0,
				Rules:      "Test rules",
				State:      &CardGameState{},
			},
			wantErr: true,
		},
		{
			name: "empty rules",
			game: &MiniGame{
				Type:       GameTypeCard,
				Name:       "Test Game",
				Difficulty: 0.5,
				TimeLimit:  300.0,
				Rules:      "",
				State:      &CardGameState{},
			},
			wantErr: true,
		},
		{
			name: "difficulty out of range",
			game: &MiniGame{
				Type:       GameTypeCard,
				Name:       "Test Game",
				Difficulty: 1.5,
				TimeLimit:  300.0,
				Rules:      "Test rules",
				State:      &CardGameState{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.game)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDeterminism(t *testing.T) {
	gen := NewGenerator()
	seed := int64(99999)
	params := procgen.GenerationParams{
		Difficulty: 0.6,
		Depth:      7,
		GenreID:    "fantasy",
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

	game1 := result1.(*MiniGame)
	game2 := result2.(*MiniGame)

	// Verify determinism
	if game1.Type != game2.Type {
		t.Errorf("Game types differ: %v vs %v", game1.Type, game2.Type)
	}

	if game1.Name != game2.Name {
		t.Errorf("Game names differ: %v vs %v", game1.Name, game2.Name)
	}

	if game1.TimeLimit != game2.TimeLimit {
		t.Errorf("Time limits differ: %v vs %v", game1.TimeLimit, game2.TimeLimit)
	}

	if game1.Rules != game2.Rules {
		t.Errorf("Rules differ: %v vs %v", game1.Rules, game2.Rules)
	}
}

func TestAllGameTypes(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	gameTypes := []GameType{
		GameTypeCard,
		GameTypeDice,
		GameTypePuzzle,
		GameTypeMemory,
		GameTypeLockPicking,
		GameTypeHacking,
		GameTypeRitual,
	}

	for _, gameType := range gameTypes {
		t.Run(gameType.String(), func(t *testing.T) {
			// Use different seeds to try to get different game types
			// This is probabilistic, but with enough attempts should hit all types
			var game *MiniGame
			for i := 0; i < 100; i++ {
				seed := int64(i * 1000)
				result, err := gen.Generate(seed, params)
				if err != nil {
					t.Fatalf("Generate failed: %v", err)
				}

				g := result.(*MiniGame)
				if g.Type == gameType {
					game = g
					break
				}
			}

			if game == nil {
				t.Skip("Could not generate specific game type in 100 attempts")
			}

			// Validate the game
			if err := gen.Validate(game); err != nil {
				t.Errorf("Validation failed: %v", err)
			}

			// Check state type matches game type
			switch game.Type {
			case GameTypeCard:
				if _, ok := game.State.(*CardGameState); !ok {
					t.Errorf("Wrong state type for card game")
				}
			case GameTypeDice:
				if _, ok := game.State.(*DiceGameState); !ok {
					t.Errorf("Wrong state type for dice game")
				}
			case GameTypePuzzle:
				if _, ok := game.State.(*PuzzleGameState); !ok {
					t.Errorf("Wrong state type for puzzle game")
				}
			case GameTypeMemory:
				if _, ok := game.State.(*MemoryGameState); !ok {
					t.Errorf("Wrong state type for memory game")
				}
			case GameTypeLockPicking:
				if _, ok := game.State.(*LockPickingGameState); !ok {
					t.Errorf("Wrong state type for lock-picking game")
				}
			case GameTypeHacking:
				if _, ok := game.State.(*HackingGameState); !ok {
					t.Errorf("Wrong state type for hacking game")
				}
			case GameTypeRitual:
				if _, ok := game.State.(*RitualGameState); !ok {
					t.Errorf("Wrong state type for ritual game")
				}
			}
		})
	}
}

func TestGenreSpecificGames(t *testing.T) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			params.GenreID = genre
			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate failed for genre %s: %v", genre, err)
			}

			game := result.(*MiniGame)
			if game.GenreID != genre {
				t.Errorf("GenreID = %s, want %s", game.GenreID, genre)
			}

			// Validate game is valid for this genre
			if err := gen.Validate(game); err != nil {
				t.Errorf("Validation failed for genre %s: %v", genre, err)
			}
		})
	}
}

func BenchmarkGenerate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(int64(i), params)
		if err != nil {
			b.Fatalf("Generate failed: %v", err)
		}
	}
}

func BenchmarkValidate(b *testing.B) {
	gen := NewGenerator()
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      5,
		GenreID:    "fantasy",
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		b.Fatalf("Generate failed: %v", err)
	}

	game := result.(*MiniGame)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := gen.Validate(game)
		if err != nil {
			b.Fatalf("Validate failed: %v", err)
		}
	}
}
