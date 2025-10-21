package particles

import (
	"testing"
)

func TestParticleType_String(t *testing.T) {
	tests := []struct {
		name     string
		pType    ParticleType
		expected string
	}{
		{"Spark", ParticleSpark, "spark"},
		{"Smoke", ParticleSmoke, "smoke"},
		{"Magic", ParticleMagic, "magic"},
		{"Flame", ParticleFlame, "flame"},
		{"Blood", ParticleBlood, "blood"},
		{"Dust", ParticleDust, "dust"},
		{"Unknown", ParticleType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.pType.String()
			if got != tt.expected {
				t.Errorf("ParticleType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Type != ParticleSpark {
		t.Errorf("DefaultConfig Type = %v, want %v", config.Type, ParticleSpark)
	}
	if config.Count != 20 {
		t.Errorf("DefaultConfig Count = %v, want 20", config.Count)
	}
	if config.GenreID != "fantasy" {
		t.Errorf("DefaultConfig GenreID = %v, want fantasy", config.GenreID)
	}
	if config.Duration != 1.0 {
		t.Errorf("DefaultConfig Duration = %v, want 1.0", config.Duration)
	}
	if config.Custom == nil {
		t.Error("DefaultConfig Custom should not be nil")
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Valid config",
			config: Config{
				Type:     ParticleSpark,
				Count:    50,
				GenreID:  "fantasy",
				Duration: 1.0,
				MinSize:  1.0,
				MaxSize:  3.0,
			},
			wantErr: false,
		},
		{
			name: "Zero count",
			config: Config{
				Type:     ParticleSpark,
				Count:    0,
				GenreID:  "fantasy",
				Duration: 1.0,
				MinSize:  1.0,
				MaxSize:  3.0,
			},
			wantErr: true,
		},
		{
			name: "Negative count",
			config: Config{
				Type:     ParticleSpark,
				Count:    -5,
				GenreID:  "fantasy",
				Duration: 1.0,
				MinSize:  1.0,
				MaxSize:  3.0,
			},
			wantErr: true,
		},
		{
			name: "Count too large",
			config: Config{
				Type:     ParticleSpark,
				Count:    20000,
				GenreID:  "fantasy",
				Duration: 1.0,
				MinSize:  1.0,
				MaxSize:  3.0,
			},
			wantErr: true,
		},
		{
			name: "Empty GenreID",
			config: Config{
				Type:     ParticleSpark,
				Count:    50,
				GenreID:  "",
				Duration: 1.0,
				MinSize:  1.0,
				MaxSize:  3.0,
			},
			wantErr: true,
		},
		{
			name: "Zero duration",
			config: Config{
				Type:     ParticleSpark,
				Count:    50,
				GenreID:  "fantasy",
				Duration: 0,
				MinSize:  1.0,
				MaxSize:  3.0,
			},
			wantErr: true,
		},
		{
			name: "Negative duration",
			config: Config{
				Type:     ParticleSpark,
				Count:    50,
				GenreID:  "fantasy",
				Duration: -1.0,
				MinSize:  1.0,
				MaxSize:  3.0,
			},
			wantErr: true,
		},
		{
			name: "Zero minSize",
			config: Config{
				Type:     ParticleSpark,
				Count:    50,
				GenreID:  "fantasy",
				Duration: 1.0,
				MinSize:  0,
				MaxSize:  3.0,
			},
			wantErr: true,
		},
		{
			name: "MaxSize less than MinSize",
			config: Config{
				Type:     ParticleSpark,
				Count:    50,
				GenreID:  "fantasy",
				Duration: 1.0,
				MinSize:  5.0,
				MaxSize:  3.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Generate spark particles",
			config: Config{
				Type:     ParticleSpark,
				Count:    30,
				GenreID:  "fantasy",
				Seed:     12345,
				Duration: 1.0,
				SpreadX:  10.0,
				SpreadY:  10.0,
				MinSize:  1.0,
				MaxSize:  3.0,
			},
			wantErr: false,
		},
		{
			name: "Generate smoke particles",
			config: Config{
				Type:     ParticleSmoke,
				Count:    40,
				GenreID:  "fantasy",
				Seed:     12345,
				Duration: 2.0,
				SpreadX:  5.0,
				SpreadY:  8.0,
				MinSize:  2.0,
				MaxSize:  5.0,
			},
			wantErr: false,
		},
		{
			name: "Generate magic particles",
			config: Config{
				Type:     ParticleMagic,
				Count:    50,
				GenreID:  "fantasy",
				Seed:     12345,
				Duration: 1.5,
				SpreadX:  8.0,
				SpreadY:  8.0,
				MinSize:  1.5,
				MaxSize:  4.0,
			},
			wantErr: false,
		},
		{
			name: "Generate flame particles",
			config: Config{
				Type:     ParticleFlame,
				Count:    60,
				GenreID:  "fantasy",
				Seed:     12345,
				Duration: 0.8,
				SpreadX:  6.0,
				SpreadY:  12.0,
				Gravity:  5.0,
				MinSize:  1.0,
				MaxSize:  3.0,
			},
			wantErr: false,
		},
		{
			name: "Generate blood particles",
			config: Config{
				Type:     ParticleBlood,
				Count:    25,
				GenreID:  "horror",
				Seed:     12345,
				Duration: 0.5,
				SpreadX:  10.0,
				SpreadY:  10.0,
				Gravity:  10.0,
				MinSize:  1.0,
				MaxSize:  2.5,
			},
			wantErr: false,
		},
		{
			name: "Generate dust particles",
			config: Config{
				Type:     ParticleDust,
				Count:    100,
				GenreID:  "postapoc",
				Seed:     12345,
				Duration: 3.0,
				SpreadX:  3.0,
				SpreadY:  3.0,
				MinSize:  0.5,
				MaxSize:  1.5,
			},
			wantErr: false,
		},
		{
			name: "Invalid config - zero count",
			config: Config{
				Type:     ParticleSpark,
				Count:    0,
				GenreID:  "fantasy",
				Seed:     12345,
				Duration: 1.0,
				MinSize:  1.0,
				MaxSize:  3.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			system, err := gen.Generate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if system == nil {
					t.Error("Generator.Generate() returned nil system")
					return
				}
				if len(system.Particles) != tt.config.Count {
					t.Errorf("Generated %d particles, want %d", len(system.Particles), tt.config.Count)
				}
				if system.Type != tt.config.Type {
					t.Errorf("System type = %v, want %v", system.Type, tt.config.Type)
				}
			}
		})
	}
}

func TestGenerator_Determinism(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:     ParticleSpark,
		Count:    50,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 1.0,
		SpreadX:  10.0,
		SpreadY:  10.0,
		MinSize:  1.0,
		MaxSize:  3.0,
	}

	// Generate the same particle system twice
	system1, err1 := gen.Generate(config)
	system2, err2 := gen.Generate(config)

	if err1 != nil || err2 != nil {
		t.Fatalf("Error generating particle systems: %v, %v", err1, err2)
	}

	// Compare particles
	if len(system1.Particles) != len(system2.Particles) {
		t.Fatalf("Particle counts differ: %d vs %d", len(system1.Particles), len(system2.Particles))
	}

	for i := range system1.Particles {
		p1 := system1.Particles[i]
		p2 := system2.Particles[i]

		if p1.X != p2.X || p1.Y != p2.Y {
			t.Errorf("Particle %d position differs: (%f, %f) vs (%f, %f)", i, p1.X, p1.Y, p2.X, p2.Y)
		}
		if p1.VX != p2.VX || p1.VY != p2.VY {
			t.Errorf("Particle %d velocity differs: (%f, %f) vs (%f, %f)", i, p1.VX, p1.VY, p2.VX, p2.VY)
		}
		if p1.Size != p2.Size {
			t.Errorf("Particle %d size differs: %f vs %f", i, p1.Size, p2.Size)
		}
		if p1.Life != p2.Life {
			t.Errorf("Particle %d life differs: %f vs %f", i, p1.Life, p2.Life)
		}
	}
}

func TestGenerator_DifferentSeeds(t *testing.T) {
	gen := NewGenerator()

	config1 := Config{
		Type:     ParticleSpark,
		Count:    50,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 1.0,
		SpreadX:  10.0,
		SpreadY:  10.0,
		MinSize:  1.0,
		MaxSize:  3.0,
	}

	config2 := Config{
		Type:     ParticleSpark,
		Count:    50,
		GenreID:  "fantasy",
		Seed:     54321,
		Duration: 1.0,
		SpreadX:  10.0,
		SpreadY:  10.0,
		MinSize:  1.0,
		MaxSize:  3.0,
	}

	system1, err1 := gen.Generate(config1)
	system2, err2 := gen.Generate(config2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Error generating particle systems: %v, %v", err1, err2)
	}

	// Systems should be different
	different := false
	for i := range system1.Particles {
		p1 := system1.Particles[i]
		p2 := system2.Particles[i]
		if p1.X != p2.X || p1.Y != p2.Y || p1.VX != p2.VX || p1.VY != p2.VY {
			different = true
			break
		}
	}

	if !different {
		t.Error("Particle systems generated with different seeds should be different")
	}
}

func TestParticleSystem_Update(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:     ParticleSpark,
		Count:    10,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 1.0,
		SpreadX:  10.0,
		SpreadY:  10.0,
		Gravity:  5.0,
		MinSize:  1.0,
		MaxSize:  3.0,
	}

	system, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Failed to generate particle system: %v", err)
	}

	// Store initial state
	initialParticle := system.Particles[0]

	// Update system
	deltaTime := 0.1
	system.Update(deltaTime)

	// Check that elapsed time increased
	if system.ElapsedTime != deltaTime {
		t.Errorf("ElapsedTime = %f, want %f", system.ElapsedTime, deltaTime)
	}

	// Check that particle life decreased
	if system.Particles[0].Life >= initialParticle.Life {
		t.Error("Particle life should decrease after update")
	}

	// Check that position changed
	if system.Particles[0].X == initialParticle.X && system.Particles[0].Y == initialParticle.Y {
		t.Error("Particle position should change after update")
	}
}

func TestParticleSystem_IsAlive(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:     ParticleSpark,
		Count:    5,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 0.1, // Short duration
		SpreadX:  10.0,
		SpreadY:  10.0,
		MinSize:  1.0,
		MaxSize:  3.0,
	}

	system, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Failed to generate particle system: %v", err)
	}

	// Initially should be alive
	if !system.IsAlive() {
		t.Error("Newly generated particle system should be alive")
	}

	// Update until all particles are dead
	for i := 0; i < 100; i++ {
		system.Update(0.01)
	}

	// Should not be alive anymore
	if system.IsAlive() {
		t.Error("Particle system should be dead after sufficient updates")
	}
}

func TestParticleSystem_GetAliveParticles(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:     ParticleSpark,
		Count:    10,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 1.0,
		SpreadX:  10.0,
		SpreadY:  10.0,
		MinSize:  1.0,
		MaxSize:  3.0,
	}

	system, err := gen.Generate(config)
	if err != nil {
		t.Fatalf("Failed to generate particle system: %v", err)
	}

	// Initially all particles should be alive
	alive := system.GetAliveParticles()
	if len(alive) != config.Count {
		t.Errorf("Got %d alive particles, want %d", len(alive), config.Count)
	}

	// Kill some particles manually
	system.Particles[0].Life = 0
	system.Particles[1].Life = -1
	system.Particles[2].Life = 0

	alive = system.GetAliveParticles()
	if len(alive) != config.Count-3 {
		t.Errorf("Got %d alive particles, want %d", len(alive), config.Count-3)
	}
}

func TestGenerator_Validate(t *testing.T) {
	gen := NewGenerator()

	tests := []struct {
		name    string
		result  interface{}
		wantErr bool
	}{
		{
			name: "Valid particle system",
			result: &ParticleSystem{
				Particles: []Particle{
					{Size: 2.0, Life: 0.5, InitialLife: 1.0},
				},
			},
			wantErr: false,
		},
		{
			name:    "Nil system",
			result:  (*ParticleSystem)(nil),
			wantErr: true,
		},
		{
			name:    "Wrong type",
			result:  "not a particle system",
			wantErr: true,
		},
		{
			name: "No particles",
			result: &ParticleSystem{
				Particles: []Particle{},
			},
			wantErr: true,
		},
		{
			name: "Invalid particle size",
			result: &ParticleSystem{
				Particles: []Particle{
					{Size: 0, Life: 0.5, InitialLife: 1.0},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid particle initial life",
			result: &ParticleSystem{
				Particles: []Particle{
					{Size: 2.0, Life: 0.5, InitialLife: 0},
				},
			},
			wantErr: true,
		},
		{
			name: "Invalid particle life",
			result: &ParticleSystem{
				Particles: []Particle{
					{Size: 2.0, Life: 1.5, InitialLife: 1.0},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.Validate(tt.result)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerator_AllGenres(t *testing.T) {
	gen := NewGenerator()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			config := Config{
				Type:     ParticleMagic,
				Count:    20,
				GenreID:  genre,
				Seed:     12345,
				Duration: 1.0,
				SpreadX:  10.0,
				SpreadY:  10.0,
				MinSize:  1.0,
				MaxSize:  3.0,
			}

			system, err := gen.Generate(config)
			if err != nil {
				t.Errorf("Failed to generate particles for genre %s: %v", genre, err)
			}
			if system == nil {
				t.Errorf("Generated nil system for genre %s", genre)
			}
		})
	}
}

func BenchmarkGenerator_Generate(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:     ParticleSpark,
		Count:    100,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 1.0,
		SpreadX:  10.0,
		SpreadY:  10.0,
		MinSize:  1.0,
		MaxSize:  3.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkParticleSystem_Update(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:     ParticleSpark,
		Count:    1000,
		GenreID:  "fantasy",
		Seed:     12345,
		Duration: 1.0,
		SpreadX:  10.0,
		SpreadY:  10.0,
		Gravity:  5.0,
		MinSize:  1.0,
		MaxSize:  3.0,
	}

	system, err := gen.Generate(config)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(0.016) // ~60 FPS
	}
}
