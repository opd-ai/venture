package terrain

import (
	"bytes"
	"testing"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// TestLoggingDoesNotAffectDeterminism verifies that enabling logging
// does not change the generated output for any terrain generator.
func TestLoggingDoesNotAffectDeterminism(t *testing.T) {
	tests := []struct {
		name        string
		createNoLog func() procgen.Generator
		createLog   func() procgen.Generator
	}{
		{
			name:        "BSP",
			createNoLog: func() procgen.Generator { return NewBSPGenerator() },
			createLog:   func() procgen.Generator { return NewBSPGeneratorWithLogger(createTestLogger()) },
		},
		{
			name:        "Cellular",
			createNoLog: func() procgen.Generator { return NewCellularGenerator() },
			createLog:   func() procgen.Generator { return NewCellularGeneratorWithLogger(createTestLogger()) },
		},
		{
			name:        "Maze",
			createNoLog: func() procgen.Generator { return NewMazeGenerator() },
			createLog:   func() procgen.Generator { return NewMazeGeneratorWithLogger(createTestLogger()) },
		},
		{
			name:        "Forest",
			createNoLog: func() procgen.Generator { return NewForestGenerator() },
			createLog:   func() procgen.Generator { return NewForestGeneratorWithLogger(createTestLogger()) },
		},
		{
			name:        "City",
			createNoLog: func() procgen.Generator { return NewCityGenerator() },
			createLog:   func() procgen.Generator { return NewCityGeneratorWithLogger(createTestLogger()) },
		},
	}

	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	seed := int64(12345)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate without logging
			genNoLog := tt.createNoLog()
			resultNoLog, err := genNoLog.Generate(seed, params)
			if err != nil {
				t.Fatalf("Generate (no log) failed: %v", err)
			}
			terrainNoLog := resultNoLog.(*Terrain)

			// Generate with logging
			genLog := tt.createLog()
			resultLog, err := genLog.Generate(seed, params)
			if err != nil {
				t.Fatalf("Generate (with log) failed: %v", err)
			}
			terrainLog := resultLog.(*Terrain)

			// Compare dimensions
			if terrainNoLog.Width != terrainLog.Width {
				t.Errorf("Width mismatch: %d (no log) vs %d (with log)", terrainNoLog.Width, terrainLog.Width)
			}
			if terrainNoLog.Height != terrainLog.Height {
				t.Errorf("Height mismatch: %d (no log) vs %d (with log)", terrainNoLog.Height, terrainLog.Height)
			}

			// Compare tiles - they should be identical
			if len(terrainNoLog.Tiles) != len(terrainLog.Tiles) {
				t.Fatalf("Tile count mismatch: %d vs %d", len(terrainNoLog.Tiles), len(terrainLog.Tiles))
			}

			for y := 0; y < terrainNoLog.Height; y++ {
				for x := 0; x < terrainNoLog.Width; x++ {
					tileNoLog := terrainNoLog.GetTile(x, y)
					tileLog := terrainLog.GetTile(x, y)
					if tileNoLog != tileLog {
						t.Errorf("Tile mismatch at (%d,%d): %v (no log) vs %v (with log)", x, y, tileNoLog, tileLog)
						return // Only report first mismatch
					}
				}
			}
		})
	}
}

// TestLoggingOutputStructure verifies that logging produces expected fields.
func TestLoggingOutputStructure(t *testing.T) {
	// Create a logger that writes to a buffer
	logger := logging.NewLogger(logging.Config{
		Level:       logging.DebugLevel,
		Format:      logging.TextFormat,
		AddCaller:   false,
		EnableColor: false,
	})

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	gen := NewCellularGeneratorWithLogger(logger)
	params := procgen.GenerationParams{
		Difficulty: 0.5,
		Depth:      1,
		GenreID:    "fantasy",
		Custom: map[string]interface{}{
			"width":  40,
			"height": 30,
		},
	}

	_, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	output := buf.String()

	// Verify that expected log messages appear
	expectedPhrases := []string{
		"starting cellular automata terrain generation",
		"cellular automata terrain generation complete",
		"seed=12345",
		"genreID=fantasy",
		"width=40",
		"height=30",
	}

	for _, phrase := range expectedPhrases {
		if !bytes.Contains([]byte(output), []byte(phrase)) {
			t.Errorf("Log output missing expected phrase: %q", phrase)
		}
	}
}

// createTestLogger creates a logger for testing that writes to a discard buffer.
func createTestLogger() *logrus.Logger {
	logger := logging.NewLogger(logging.Config{
		Level:       logging.DebugLevel,
		Format:      logging.TextFormat,
		AddCaller:   false,
		EnableColor: false,
	})
	// Discard output for tests
	var buf bytes.Buffer
	logger.SetOutput(&buf)
	return logger
}
