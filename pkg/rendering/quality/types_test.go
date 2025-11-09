package quality

import (
	"strings"
	"testing"
)

func TestQualityLevel_String(t *testing.T) {
	tests := []struct {
		level QualityLevel
		want  string
	}{
		{QualityLow, "Low"},
		{QualityMedium, "Medium"},
		{QualityHigh, "High"},
		{QualityLevel(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.level.String()
			if got != tt.want {
				t.Errorf("QualityLevel.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLowQualityConfig(t *testing.T) {
	config := LowQualityConfig()

	if config.Level != QualityLow {
		t.Errorf("Level = %v, want %v", config.Level, QualityLow)
	}

	// Verify low quality disables expensive features
	if config.EnablePostProcessing {
		t.Error("EnablePostProcessing should be false for low quality")
	}
	if config.EnableBloom {
		t.Error("EnableBloom should be false for low quality")
	}
	if config.EnableAmbientOcclusion {
		t.Error("EnableAmbientOcclusion should be false for low quality")
	}
	if config.EnableSoftShadows {
		t.Error("EnableSoftShadows should be false for low quality")
	}
	if config.EnableParticlePhysics {
		t.Error("EnableParticlePhysics should be false for low quality")
	}

	// Verify low quality reduces particle count significantly
	if config.ParticleCountMultiplier >= 0.5 {
		t.Errorf("ParticleCountMultiplier = %f, should be < 0.5 for low quality", config.ParticleCountMultiplier)
	}

	// Verify low quality uses minimal sprite detail
	if config.SpriteDetailLevel >= 0.5 {
		t.Errorf("SpriteDetailLevel = %f, should be < 0.5 for low quality", config.SpriteDetailLevel)
	}

	// Verify optimization features are enabled
	if !config.ViewportCulling {
		t.Error("ViewportCulling should be enabled for low quality")
	}
	if !config.BatchRendering {
		t.Error("BatchRendering should be enabled for low quality")
	}
	if !config.ObjectPooling {
		t.Error("ObjectPooling should be enabled for low quality")
	}
}

func TestMediumQualityConfig(t *testing.T) {
	config := MediumQualityConfig()

	if config.Level != QualityMedium {
		t.Errorf("Level = %v, want %v", config.Level, QualityMedium)
	}

	// Verify medium quality enables key features
	if !config.EnableColorGrading {
		t.Error("EnableColorGrading should be true for medium quality")
	}
	if !config.EnableVignette {
		t.Error("EnableVignette should be true for medium quality")
	}
	if !config.EnableSoftShadows {
		t.Error("EnableSoftShadows should be true for medium quality")
	}

	// Verify medium quality uses moderate settings
	if config.ParticleCountMultiplier < 0.5 || config.ParticleCountMultiplier > 0.7 {
		t.Errorf("ParticleCountMultiplier = %f, want in range [0.5, 0.7]", config.ParticleCountMultiplier)
	}
	if config.SpriteDetailLevel < 0.6 || config.SpriteDetailLevel > 0.8 {
		t.Errorf("SpriteDetailLevel = %f, want in range [0.6, 0.8]", config.SpriteDetailLevel)
	}
}

func TestHighQualityConfig(t *testing.T) {
	config := HighQualityConfig()

	if config.Level != QualityHigh {
		t.Errorf("Level = %v, want %v", config.Level, QualityHigh)
	}

	// Verify high quality enables all features
	if !config.EnablePostProcessing {
		t.Error("EnablePostProcessing should be true for high quality")
	}
	if !config.EnableBloom {
		t.Error("EnableBloom should be true for high quality")
	}
	if !config.EnableAmbientOcclusion {
		t.Error("EnableAmbientOcclusion should be true for high quality")
	}
	if !config.EnableSoftShadows {
		t.Error("EnableSoftShadows should be true for high quality")
	}
	if !config.EnableParticlePhysics {
		t.Error("EnableParticlePhysics should be true for high quality")
	}

	// Verify high quality uses maximum settings
	if config.ParticleCountMultiplier != 1.0 {
		t.Errorf("ParticleCountMultiplier = %f, want 1.0", config.ParticleCountMultiplier)
	}
	if config.SpriteDetailLevel != 1.0 {
		t.Errorf("SpriteDetailLevel = %f, want 1.0", config.SpriteDetailLevel)
	}
	if config.DecorationDensity != 1.0 {
		t.Errorf("DecorationDensity = %f, want 1.0", config.DecorationDensity)
	}

	// Verify high quality uses best shadow quality
	if config.ShadowSampleCount < 5 {
		t.Errorf("ShadowSampleCount = %d, want >= 5", config.ShadowSampleCount)
	}
}

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid low quality",
			config:  LowQualityConfig(),
			wantErr: false,
		},
		{
			name:    "valid medium quality",
			config:  MediumQualityConfig(),
			wantErr: false,
		},
		{
			name:    "valid high quality",
			config:  HighQualityConfig(),
			wantErr: false,
		},
		{
			name: "invalid sprite detail - too low",
			config: Config{
				SpriteDetailLevel: -0.1,
			},
			wantErr: true,
			errMsg:  "SpriteDetailLevel",
		},
		{
			name: "invalid sprite detail - too high",
			config: Config{
				SpriteDetailLevel: 1.5,
			},
			wantErr: true,
			errMsg:  "SpriteDetailLevel",
		},
		{
			name: "invalid AA quality - too low",
			config: Config{
				SpriteDetailLevel:   0.5,
				AntiAliasingQuality: -1,
			},
			wantErr: true,
			errMsg:  "AntiAliasingQuality",
		},
		{
			name: "invalid AA quality - too high",
			config: Config{
				SpriteDetailLevel:   0.5,
				AntiAliasingQuality: 4,
			},
			wantErr: true,
			errMsg:  "AntiAliasingQuality",
		},
		{
			name: "invalid tile layer count - too low",
			config: Config{
				SpriteDetailLevel:   0.5,
				AntiAliasingQuality: 1,
				TileLayerCount:      0,
			},
			wantErr: true,
			errMsg:  "TileLayerCount",
		},
		{
			name: "invalid tile layer count - too high",
			config: Config{
				SpriteDetailLevel:   0.5,
				AntiAliasingQuality: 1,
				TileLayerCount:      4,
			},
			wantErr: true,
			errMsg:  "TileLayerCount",
		},
		{
			name: "invalid particle multiplier - negative",
			config: Config{
				SpriteDetailLevel:       0.5,
				AntiAliasingQuality:     1,
				TileLayerCount:          2,
				ParticleCountMultiplier: -0.1,
			},
			wantErr: true,
			errMsg:  "ParticleCountMultiplier",
		},
		{
			name: "invalid decoration density - too high",
			config: Config{
				SpriteDetailLevel:       0.5,
				AntiAliasingQuality:     1,
				TileLayerCount:          2,
				ParticleCountMultiplier: 0.5,
				DecorationDensity:       1.5,
			},
			wantErr: true,
			errMsg:  "DecorationDensity",
		},
		{
			name: "invalid shadow sample count - too low",
			config: Config{
				SpriteDetailLevel:       0.5,
				AntiAliasingQuality:     1,
				TileLayerCount:          2,
				ParticleCountMultiplier: 0.5,
				DecorationDensity:       0.5,
				ShadowSampleCount:       0,
			},
			wantErr: true,
			errMsg:  "ShadowSampleCount",
		},
		{
			name: "invalid max particles - negative",
			config: Config{
				SpriteDetailLevel:       0.5,
				AntiAliasingQuality:     1,
				TileLayerCount:          2,
				ParticleCountMultiplier: 0.5,
				DecorationDensity:       0.5,
				ShadowSampleCount:       3,
				MaxParticles:            -100,
			},
			wantErr: true,
			errMsg:  "MaxParticles",
		},
		{
			name: "invalid cache size - negative",
			config: Config{
				SpriteDetailLevel:       0.5,
				AntiAliasingQuality:     1,
				TileLayerCount:          2,
				ParticleCountMultiplier: 0.5,
				DecorationDensity:       0.5,
				ShadowSampleCount:       3,
				MaxParticles:            1000,
				CacheSizeMB:             -50,
			},
			wantErr: true,
			errMsg:  "CacheSizeMB",
		},
		{
			name: "invalid particle LOD distance - negative",
			config: Config{
				SpriteDetailLevel:       0.5,
				AntiAliasingQuality:     1,
				TileLayerCount:          2,
				ParticleCountMultiplier: 0.5,
				DecorationDensity:       0.5,
				ShadowSampleCount:       3,
				MaxParticles:            1000,
				CacheSizeMB:             50,
				ParticleLODDistance:     -100.0,
			},
			wantErr: true,
			errMsg:  "ParticleLODDistance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Config.Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil {
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

func TestConfig_ApplyLevel(t *testing.T) {
	tests := []struct {
		level      QualityLevel
		wantLevel  QualityLevel
		checkField func(Config) bool
		fieldName  string
	}{
		{
			level:     QualityLow,
			wantLevel: QualityLow,
			checkField: func(c Config) bool {
				return c.ParticleCountMultiplier < 0.3
			},
			fieldName: "ParticleCountMultiplier should be < 0.3",
		},
		{
			level:     QualityMedium,
			wantLevel: QualityMedium,
			checkField: func(c Config) bool {
				return c.EnableVignette
			},
			fieldName: "EnableVignette should be true",
		},
		{
			level:     QualityHigh,
			wantLevel: QualityHigh,
			checkField: func(c Config) bool {
				return c.EnableBloom
			},
			fieldName: "EnableBloom should be true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			config := DefaultConfig()
			config.ApplyLevel(tt.level)

			if config.Level != tt.wantLevel {
				t.Errorf("ApplyLevel(%v) Level = %v, want %v", tt.level, config.Level, tt.wantLevel)
			}

			if !tt.checkField(config) {
				t.Errorf("ApplyLevel(%v) %s", tt.level, tt.fieldName)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Default should be medium quality
	if config.Level != QualityMedium {
		t.Errorf("DefaultConfig() Level = %v, want %v", config.Level, QualityMedium)
	}

	// Should be valid
	if err := config.Validate(); err != nil {
		t.Errorf("DefaultConfig() invalid: %v", err)
	}
}

func TestQualityProgression(t *testing.T) {
	low := LowQualityConfig()
	medium := MediumQualityConfig()
	high := HighQualityConfig()

	// Verify progressive increase in quality settings
	if !(low.ParticleCountMultiplier < medium.ParticleCountMultiplier &&
		medium.ParticleCountMultiplier < high.ParticleCountMultiplier) {
		t.Error("ParticleCountMultiplier should increase with quality level")
	}

	if !(low.SpriteDetailLevel < medium.SpriteDetailLevel &&
		medium.SpriteDetailLevel < high.SpriteDetailLevel) {
		t.Error("SpriteDetailLevel should increase with quality level")
	}

	if !(low.DecorationDensity < medium.DecorationDensity &&
		medium.DecorationDensity <= high.DecorationDensity) {
		t.Error("DecorationDensity should increase with quality level")
	}

	if !(low.ShadowSampleCount < medium.ShadowSampleCount &&
		medium.ShadowSampleCount < high.ShadowSampleCount) {
		t.Error("ShadowSampleCount should increase with quality level")
	}

	if !(low.MaxParticles < medium.MaxParticles &&
		medium.MaxParticles < high.MaxParticles) {
		t.Error("MaxParticles should increase with quality level")
	}
}

func TestPerformanceOptimizations(t *testing.T) {
	// All quality levels should have core optimizations enabled
	configs := []Config{
		LowQualityConfig(),
		MediumQualityConfig(),
		HighQualityConfig(),
	}

	for i, config := range configs {
		level := QualityLevel(i)
		t.Run(level.String(), func(t *testing.T) {
			if !config.ViewportCulling {
				t.Error("ViewportCulling should always be enabled")
			}
			if !config.BatchRendering {
				t.Error("BatchRendering should always be enabled")
			}
			if !config.ObjectPooling {
				t.Error("ObjectPooling should always be enabled")
			}
		})
	}
}

// Benchmarks

func BenchmarkConfig_Validate(b *testing.B) {
	config := HighQualityConfig()
	for i := 0; i < b.N; i++ {
		_ = config.Validate()
	}
}

func BenchmarkConfig_ApplyLevel(b *testing.B) {
	config := DefaultConfig()
	for i := 0; i < b.N; i++ {
		config.ApplyLevel(QualityHigh)
	}
}
