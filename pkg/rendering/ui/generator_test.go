package ui

import (
	"image"
	"testing"
)

func TestElementType_String(t *testing.T) {
	tests := []struct {
		name     string
		eType    ElementType
		expected string
	}{
		{"Button", ElementButton, "button"},
		{"Panel", ElementPanel, "panel"},
		{"HealthBar", ElementHealthBar, "healthbar"},
		{"Label", ElementLabel, "label"},
		{"Icon", ElementIcon, "icon"},
		{"Frame", ElementFrame, "frame"},
		{"Unknown", ElementType(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.eType.String()
			if got != tt.expected {
				t.Errorf("ElementType.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestElementState_String(t *testing.T) {
	tests := []struct {
		name     string
		state    ElementState
		expected string
	}{
		{"Normal", StateNormal, "normal"},
		{"Hover", StateHover, "hover"},
		{"Pressed", StatePressed, "pressed"},
		{"Disabled", StateDisabled, "disabled"},
		{"Unknown", ElementState(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.state.String()
			if got != tt.expected {
				t.Errorf("ElementState.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestBorderStyle_String(t *testing.T) {
	tests := []struct {
		name     string
		style    BorderStyle
		expected string
	}{
		{"Solid", BorderSolid, "solid"},
		{"Double", BorderDouble, "double"},
		{"Ornate", BorderOrnate, "ornate"},
		{"Glow", BorderGlow, "glow"},
		{"Unknown", BorderStyle(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.style.String()
			if got != tt.expected {
				t.Errorf("BorderStyle.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Type != ElementButton {
		t.Errorf("DefaultConfig Type = %v, want %v", config.Type, ElementButton)
	}
	if config.Width != 100 {
		t.Errorf("DefaultConfig Width = %v, want 100", config.Width)
	}
	if config.Height != 30 {
		t.Errorf("DefaultConfig Height = %v, want 30", config.Height)
	}
	if config.GenreID != "fantasy" {
		t.Errorf("DefaultConfig GenreID = %v, want fantasy", config.GenreID)
	}
	if config.Value != 1.0 {
		t.Errorf("DefaultConfig Value = %v, want 1.0", config.Value)
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
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "fantasy",
				Value:   0.5,
			},
			wantErr: false,
		},
		{
			name: "Zero width",
			config: Config{
				Type:    ElementButton,
				Width:   0,
				Height:  30,
				GenreID: "fantasy",
				Value:   0.5,
			},
			wantErr: true,
		},
		{
			name: "Zero height",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  0,
				GenreID: "fantasy",
				Value:   0.5,
			},
			wantErr: true,
		},
		{
			name: "Empty GenreID",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "",
				Value:   0.5,
			},
			wantErr: true,
		},
		{
			name: "Value too low",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Value:   -0.1,
			},
			wantErr: true,
		},
		{
			name: "Value too high",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Value:   1.1,
			},
			wantErr: true,
		},
		{
			name: "Value at bounds - 0",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Value:   0.0,
			},
			wantErr: false,
		},
		{
			name: "Value at bounds - 1",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Value:   1.0,
			},
			wantErr: false,
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
			name: "Generate button",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "fantasy",
				Seed:    12345,
				Text:    "Click Me",
				State:   StateNormal,
			},
			wantErr: false,
		},
		{
			name: "Generate button - hover state",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "fantasy",
				Seed:    12345,
				State:   StateHover,
			},
			wantErr: false,
		},
		{
			name: "Generate button - pressed state",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "scifi",
				Seed:    12345,
				State:   StatePressed,
			},
			wantErr: false,
		},
		{
			name: "Generate button - disabled state",
			config: Config{
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: "fantasy",
				Seed:    12345,
				State:   StateDisabled,
			},
			wantErr: false,
		},
		{
			name: "Generate panel",
			config: Config{
				Type:    ElementPanel,
				Width:   200,
				Height:  150,
				GenreID: "fantasy",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "Generate health bar - full",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Seed:    12345,
				Value:   1.0,
			},
			wantErr: false,
		},
		{
			name: "Generate health bar - half",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Seed:    12345,
				Value:   0.5,
			},
			wantErr: false,
		},
		{
			name: "Generate health bar - low",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Seed:    12345,
				Value:   0.2,
			},
			wantErr: false,
		},
		{
			name: "Generate health bar - empty",
			config: Config{
				Type:    ElementHealthBar,
				Width:   100,
				Height:  20,
				GenreID: "fantasy",
				Seed:    12345,
				Value:   0.0,
			},
			wantErr: false,
		},
		{
			name: "Generate label",
			config: Config{
				Type:    ElementLabel,
				Width:   80,
				Height:  20,
				GenreID: "fantasy",
				Seed:    12345,
				Text:    "Score: 100",
			},
			wantErr: false,
		},
		{
			name: "Generate icon",
			config: Config{
				Type:    ElementIcon,
				Width:   32,
				Height:  32,
				GenreID: "fantasy",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "Generate icon - scifi",
			config: Config{
				Type:    ElementIcon,
				Width:   32,
				Height:  32,
				GenreID: "scifi",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "Generate frame",
			config: Config{
				Type:    ElementFrame,
				Width:   300,
				Height:  200,
				GenreID: "fantasy",
				Seed:    12345,
			},
			wantErr: false,
		},
		{
			name: "Invalid config - zero width",
			config: Config{
				Type:    ElementButton,
				Width:   0,
				Height:  30,
				GenreID: "fantasy",
				Seed:    12345,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			img, err := gen.Generate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Generator.Generate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if img == nil {
					t.Error("Generator.Generate() returned nil image")
					return
				}
				bounds := img.Bounds()
				if bounds.Dx() != tt.config.Width {
					t.Errorf("Generated image width = %v, want %v", bounds.Dx(), tt.config.Width)
				}
				if bounds.Dy() != tt.config.Height {
					t.Errorf("Generated image height = %v, want %v", bounds.Dy(), tt.config.Height)
				}
			}
		})
	}
}

func TestGenerator_Determinism(t *testing.T) {
	gen := NewGenerator()

	config := Config{
		Type:    ElementButton,
		Width:   100,
		Height:  30,
		GenreID: "fantasy",
		Seed:    12345,
		State:   StateNormal,
	}

	// Generate the same element twice
	img1, err1 := gen.Generate(config)
	img2, err2 := gen.Generate(config)

	if err1 != nil || err2 != nil {
		t.Fatalf("Error generating UI elements: %v, %v", err1, err2)
	}

	// Compare pixel by pixel
	bounds := img1.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			r1, g1, b1, a1 := c1.RGBA()
			r2, g2, b2, a2 := c2.RGBA()
			if r1 != r2 || g1 != g2 || b1 != b2 || a1 != a2 {
				t.Errorf("Pixel at (%d, %d) differs: %v vs %v", x, y, c1, c2)
				return
			}
		}
	}
}

func TestGenerator_DifferentSeeds(t *testing.T) {
	gen := NewGenerator()

	config1 := Config{
		Type:    ElementButton,
		Width:   100,
		Height:  30,
		GenreID: "fantasy",
		Seed:    12345,
		State:   StateNormal,
	}

	config2 := Config{
		Type:    ElementButton,
		Width:   100,
		Height:  30,
		GenreID: "fantasy",
		Seed:    54321,
		State:   StateNormal,
	}

	img1, err1 := gen.Generate(config1)
	img2, err2 := gen.Generate(config2)

	if err1 != nil || err2 != nil {
		t.Fatalf("Error generating UI elements: %v, %v", err1, err2)
	}

	// Images should be different
	bounds := img1.Bounds()
	different := false
	for y := bounds.Min.Y; y < bounds.Max.Y && !different; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c1 := img1.At(x, y)
			c2 := img2.At(x, y)
			if c1 != c2 {
				different = true
				break
			}
		}
	}

	if !different {
		t.Error("UI elements generated with different seeds should be different")
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
			name:    "Valid image",
			result:  image.NewRGBA(image.Rect(0, 0, 100, 30)),
			wantErr: false,
		},
		{
			name:    "Nil image",
			result:  (*image.RGBA)(nil),
			wantErr: true,
		},
		{
			name:    "Wrong type",
			result:  "not an image",
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
				Type:    ElementButton,
				Width:   100,
				Height:  30,
				GenreID: genre,
				Seed:    12345,
				State:   StateNormal,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Errorf("Failed to generate UI element for genre %s: %v", genre, err)
			}
			if img == nil {
				t.Errorf("Generated nil image for genre %s", genre)
			}
		})
	}
}

func TestGenerator_AllElementTypes(t *testing.T) {
	gen := NewGenerator()
	elementTypes := []ElementType{
		ElementButton, ElementPanel, ElementHealthBar,
		ElementLabel, ElementIcon, ElementFrame,
	}

	for _, eType := range elementTypes {
		t.Run(eType.String(), func(t *testing.T) {
			config := Config{
				Type:    eType,
				Width:   100,
				Height:  50,
				GenreID: "fantasy",
				Seed:    12345,
				Value:   0.5,
			}

			img, err := gen.Generate(config)
			if err != nil {
				t.Errorf("Failed to generate %s: %v", eType, err)
			}
			if img == nil {
				t.Errorf("Generated nil image for %s", eType)
			}
		})
	}
}

func BenchmarkGenerator_GenerateButton(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:    ElementButton,
		Width:   100,
		Height:  30,
		GenreID: "fantasy",
		Seed:    12345,
		State:   StateNormal,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerator_GenerateHealthBar(b *testing.B) {
	gen := NewGenerator()
	config := Config{
		Type:    ElementHealthBar,
		Width:   100,
		Height:  20,
		GenreID: "fantasy",
		Seed:    12345,
		Value:   0.75,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := gen.Generate(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}
