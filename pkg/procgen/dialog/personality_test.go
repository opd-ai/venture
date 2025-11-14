package dialog

import (
	"math/rand"
	"strings"
	"testing"
)

// TestNewPersonality verifies personality creation for all types.
func TestNewPersonality(t *testing.T) {
	types := []PersonalityType{
		PersonalityHelpful,
		PersonalityMerchant,
		PersonalityHostile,
		PersonalityMysterious,
		PersonalityScholarly,
		PersonalityWarrior,
		PersonalityTimid,
		PersonalityArrogant,
	}

	for _, ptype := range types {
		t.Run(string(ptype), func(t *testing.T) {
			p := NewPersonality(ptype)

			if p == nil {
				t.Fatal("NewPersonality returned nil")
			}

			if p.Type != ptype {
				t.Errorf("Type = %v, want %v", p.Type, ptype)
			}

			// Verify traits are within valid range [0.0, 1.0]
			checkRange := func(name string, value float64) {
				if value < 0.0 || value > 1.0 {
					t.Errorf("%s = %.2f, want [0.0, 1.0]", name, value)
				}
			}

			checkRange("Friendliness", p.Friendliness)
			checkRange("Verbosity", p.Verbosity)
			checkRange("Formality", p.Formality)
			checkRange("Humor", p.Humor)
			checkRange("Knowledge", p.Knowledge)
		})
	}
}

// TestNewPersonalityDefaults verifies default trait values.
func TestNewPersonalityDefaults(t *testing.T) {
	tests := []struct {
		ptype      PersonalityType
		expectHigh string // Trait that should be high (>0.6)
		expectLow  string // Trait that should be low (<0.4)
	}{
		{PersonalityHelpful, "Friendliness", ""},
		{PersonalityHostile, "", "Friendliness"},
		{PersonalityScholarly, "Knowledge", ""},
		{PersonalityTimid, "", "Humor"},
	}

	for _, tt := range tests {
		t.Run(string(tt.ptype), func(t *testing.T) {
			p := NewPersonality(tt.ptype)

			getTrait := func(name string) float64 {
				switch name {
				case "Friendliness":
					return p.Friendliness
				case "Verbosity":
					return p.Verbosity
				case "Formality":
					return p.Formality
				case "Humor":
					return p.Humor
				case "Knowledge":
					return p.Knowledge
				default:
					return 0.5
				}
			}

			if tt.expectHigh != "" {
				value := getTrait(tt.expectHigh)
				if value <= 0.6 {
					t.Errorf("%s for %s = %.2f, want > 0.6", tt.expectHigh, tt.ptype, value)
				}
			}

			if tt.expectLow != "" {
				value := getTrait(tt.expectLow)
				if value >= 0.4 {
					t.Errorf("%s for %s = %.2f, want < 0.4", tt.expectLow, tt.ptype, value)
				}
			}
		})
	}
}

// TestGenerateRandomPersonality verifies random personality generation.
func TestGenerateRandomPersonality(t *testing.T) {
	rng := rand.New(rand.NewSource(12345))

	// Generate multiple personalities
	personalities := make([]*Personality, 10)
	for i := 0; i < 10; i++ {
		personalities[i] = GenerateRandomPersonality(rng)

		if personalities[i] == nil {
			t.Fatalf("iteration %d: GenerateRandomPersonality returned nil", i)
		}

		// Verify traits are in valid range
		checkRange := func(name string, value float64) {
			if value < 0.0 || value > 1.0 {
				t.Errorf("iteration %d: %s = %.2f, want [0.0, 1.0]", i, name, value)
			}
		}

		p := personalities[i]
		checkRange("Friendliness", p.Friendliness)
		checkRange("Verbosity", p.Verbosity)
		checkRange("Formality", p.Formality)
		checkRange("Humor", p.Humor)
		checkRange("Knowledge", p.Knowledge)
	}

	// Verify some variation in types
	types := make(map[PersonalityType]bool)
	for _, p := range personalities {
		types[p.Type] = true
	}

	if len(types) < 3 {
		t.Errorf("generated %d unique personality types out of 10, want >= 3", len(types))
	}
}

// TestApplyToGenerator verifies personality affects generation parameters.
func TestApplyToGenerator(t *testing.T) {
	tests := []struct {
		name         string
		personality  *Personality
		initialMax   int
		initialMin   int
		initialTemp  float64
		expectChange string // "max", "min", "temp"
	}{
		{
			name:         "high verbosity increases words",
			personality:  &Personality{Type: PersonalityScholarly, Verbosity: 0.9},
			initialMax:   30,
			initialMin:   10,
			initialTemp:  0.7,
			expectChange: "max",
		},
		{
			name:         "low verbosity decreases words",
			personality:  &Personality{Type: PersonalityTimid, Verbosity: 0.2},
			initialMax:   30,
			initialMin:   10,
			initialTemp:  0.7,
			expectChange: "max",
		},
		{
			name:         "high friendliness increases temperature",
			personality:  &Personality{Type: PersonalityHelpful, Friendliness: 0.9},
			initialMax:   30,
			initialMin:   10,
			initialTemp:  0.7,
			expectChange: "temp",
		},
		{
			name:         "high knowledge increases words",
			personality:  &Personality{Type: PersonalityScholarly, Knowledge: 0.9},
			initialMax:   30,
			initialMin:   10,
			initialTemp:  0.7,
			expectChange: "max",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &GenerateParams{
				MaxWords:    tt.initialMax,
				MinWords:    tt.initialMin,
				Temperature: tt.initialTemp,
			}

			tt.personality.ApplyToGenerator(params)

			// Verify parameters were modified
			switch tt.expectChange {
			case "max":
				if params.MaxWords == tt.initialMax {
					t.Error("MaxWords unchanged")
				}
			case "min":
				if params.MinWords == tt.initialMin {
					t.Error("MinWords unchanged")
				}
			case "temp":
				if params.Temperature == tt.initialTemp {
					t.Error("Temperature unchanged")
				}
			}

			// Verify parameters remain valid
			if params.MaxWords < 10 || params.MaxWords > 100 {
				t.Errorf("MaxWords = %d, want [10, 100]", params.MaxWords)
			}
			if params.MinWords < 5 || params.MinWords > 50 {
				t.Errorf("MinWords = %d, want [5, 50]", params.MinWords)
			}
			if params.Temperature < 0.0 || params.Temperature > 1.0 {
				t.Errorf("Temperature = %.2f, want [0.0, 1.0]", params.Temperature)
			}
			if params.MinWords > params.MaxWords {
				t.Errorf("MinWords (%d) > MaxWords (%d)", params.MinWords, params.MaxWords)
			}
		})
	}
}

// TestGetGreeting verifies personality-based greetings.
func TestGetGreeting(t *testing.T) {
	personalities := []PersonalityType{
		PersonalityHelpful,
		PersonalityMerchant,
		PersonalityHostile,
		PersonalityMysterious,
		PersonalityScholarly,
		PersonalityWarrior,
		PersonalityTimid,
		PersonalityArrogant,
	}

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapocalyptic"}

	for _, ptype := range personalities {
		for _, genre := range genres {
			t.Run(string(ptype)+"_"+genre, func(t *testing.T) {
				p := NewPersonality(ptype)
				greeting := p.GetGreeting(genre)

				if greeting == "" {
					t.Error("GetGreeting returned empty string")
				}

				// Greeting should be reasonable length
				if len(greeting) < 3 {
					t.Errorf("greeting too short: %q", greeting)
				}
				if len(greeting) > 200 {
					t.Errorf("greeting too long (%d chars): %q", len(greeting), greeting)
				}
			})
		}
	}
}

// TestGetGreetingUnknownGenre verifies fallback for unknown genre.
func TestGetGreetingUnknownGenre(t *testing.T) {
	p := NewPersonality(PersonalityHelpful)
	greeting := p.GetGreeting("unknown-genre")

	if greeting == "" {
		t.Error("GetGreeting with unknown genre returned empty string")
	}

	// Should get a fallback greeting
	if greeting != "Hello." {
		// Might be a valid greeting from personality/genre combo
		if len(greeting) < 3 {
			t.Errorf("fallback greeting too short: %q", greeting)
		}
	}
}

// TestPersonalityString verifies String() method.
func TestPersonalityString(t *testing.T) {
	p := NewPersonality(PersonalityHelpful)
	str := p.String()

	if str == "" {
		t.Error("String() returned empty string")
	}

	if !strings.Contains(str, "helpful") {
		t.Error("String() should contain personality type")
	}

	if !strings.Contains(str, "Personality{") {
		t.Error("String() should start with Personality{")
	}
}

// TestClamp verifies clamp utility function.
func TestClamp(t *testing.T) {
	tests := []struct {
		value float64
		min   float64
		max   float64
		want  float64
	}{
		{0.5, 0.0, 1.0, 0.5},    // Within range
		{-0.5, 0.0, 1.0, 0.0},   // Below min
		{1.5, 0.0, 1.0, 1.0},    // Above max
		{0.0, 0.0, 1.0, 0.0},    // At min
		{1.0, 0.0, 1.0, 1.0},    // At max
		{5.0, 10.0, 20.0, 10.0}, // Below min
	}

	for _, tt := range tests {
		got := clamp(tt.value, tt.min, tt.max)
		if got != tt.want {
			t.Errorf("clamp(%.2f, %.2f, %.2f) = %.2f, want %.2f",
				tt.value, tt.min, tt.max, got, tt.want)
		}
	}
}

// TestMaxMin verifies max/min utility functions.
func TestMaxMin(t *testing.T) {
	if max(5, 10) != 10 {
		t.Error("max(5, 10) should be 10")
	}
	if max(10, 5) != 10 {
		t.Error("max(10, 5) should be 10")
	}
	if max(5, 5) != 5 {
		t.Error("max(5, 5) should be 5")
	}

	if min(5, 10) != 5 {
		t.Error("min(5, 10) should be 5")
	}
	if min(10, 5) != 5 {
		t.Error("min(10, 5) should be 5")
	}
	if min(5, 5) != 5 {
		t.Error("min(5, 5) should be 5")
	}
}
