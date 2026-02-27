package audit

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestGetBaselinePrefix(t *testing.T) {
	tests := []struct {
		name       string
		generator  string
		wantPrefix string
		wantEmpty  bool
	}{
		{
			name:       "Entity exists",
			generator:  "Entity",
			wantPrefix: "f0302eb430a7d0cd",
			wantEmpty:  false,
		},
		{
			name:       "Item exists",
			generator:  "Item",
			wantPrefix: "2b36ce659bf7c7b6",
			wantEmpty:  false,
		},
		{
			name:       "Unknown generator",
			generator:  "UnknownGenerator",
			wantPrefix: "",
			wantEmpty:  true,
		},
		{
			name:       "Empty string",
			generator:  "",
			wantPrefix: "",
			wantEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetBaselinePrefix(tt.generator)
			if tt.wantEmpty && got != "" {
				t.Errorf("GetBaselinePrefix(%q) = %q, want empty string", tt.generator, got)
			}
			if !tt.wantEmpty && got != tt.wantPrefix {
				t.Errorf("GetBaselinePrefix(%q) = %q, want %q", tt.generator, got, tt.wantPrefix)
			}
		})
	}
}

func TestHashMatchesBaseline(t *testing.T) {
	// Create a test hash that matches Entity baseline
	entityHash := [32]byte{}
	entityPrefix, _ := hex.DecodeString("f0302eb430a7d0cd")
	copy(entityHash[:8], entityPrefix)

	// Create a test hash that doesn't match
	wrongHash := [32]byte{}
	wrongPrefix, _ := hex.DecodeString("0000000000000000")
	copy(wrongHash[:8], wrongPrefix)

	tests := []struct {
		name      string
		generator string
		hash      [32]byte
		want      bool
	}{
		{
			name:      "matching hash",
			generator: "Entity",
			hash:      entityHash,
			want:      true,
		},
		{
			name:      "non-matching hash",
			generator: "Entity",
			hash:      wrongHash,
			want:      false,
		},
		{
			name:      "unknown generator",
			generator: "UnknownGenerator",
			hash:      entityHash,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashMatchesBaseline(tt.generator, tt.hash)
			if got != tt.want {
				t.Errorf("HashMatchesBaseline(%q, hash) = %v, want %v", tt.generator, got, tt.want)
			}
		})
	}
}

func TestBaselineHashesFileOperations(t *testing.T) {
	// Create temp directory for test
	tmpDir, err := os.MkdirTemp("", "baseline_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test saving baseline hashes
	t.Run("SaveAndLoad", func(t *testing.T) {
		hashes := &BaselineHashes{
			Version: "1.0.0",
			Seed:    99999,
			Generators: map[string]string{
				"TestGenerator": "abcdef1234567890",
			},
		}

		// Save
		err := SaveBaselineHashes(tmpDir, hashes)
		if err != nil {
			t.Fatalf("SaveBaselineHashes() error = %v", err)
		}

		// Verify file exists
		path := filepath.Join(tmpDir, BaselineHashFile)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatal("Baseline file was not created")
		}

		// Load
		loaded, err := LoadBaselineHashes(tmpDir)
		if err != nil {
			t.Fatalf("LoadBaselineHashes() error = %v", err)
		}

		// Verify content
		if loaded.Version != hashes.Version {
			t.Errorf("Version = %q, want %q", loaded.Version, hashes.Version)
		}
		if loaded.Seed != hashes.Seed {
			t.Errorf("Seed = %d, want %d", loaded.Seed, hashes.Seed)
		}
		if loaded.Generators["TestGenerator"] != hashes.Generators["TestGenerator"] {
			t.Errorf("Generator hash mismatch")
		}
	})

	t.Run("LoadNonExistent", func(t *testing.T) {
		_, err := LoadBaselineHashes("/nonexistent/path")
		if err == nil {
			t.Error("LoadBaselineHashes() should return error for non-existent path")
		}
	})
}

func TestBaselineConstants(t *testing.T) {
	// Verify baseline constants are set correctly
	if BaselineVersion != "1.0.0" {
		t.Errorf("BaselineVersion = %q, want %q", BaselineVersion, "1.0.0")
	}

	if BaselineSeed != 99999 {
		t.Errorf("BaselineSeed = %d, want %d", BaselineSeed, 99999)
	}

	if BaselineHashFile != "baseline_hashes.json" {
		t.Errorf("BaselineHashFile = %q, want %q", BaselineHashFile, "baseline_hashes.json")
	}
}

func TestBaselineHashPrefixesComplete(t *testing.T) {
	// Verify all expected generators have baseline entries
	expectedGenerators := []string{
		"Book",
		"Building",
		"Companion",
		"Entity",
		"Furniture",
		"Item",
		"Legendary",
		"Magic",
		"Quest",
		"Recipe",
		"Skills",
		"Station",
		"Terrain",
		"Vehicle",
	}

	for _, gen := range expectedGenerators {
		prefix := GetBaselinePrefix(gen)
		if prefix == "" {
			t.Errorf("Missing baseline prefix for %s", gen)
		}
		// Verify prefix is 16 hex characters (8 bytes)
		if len(prefix) != 16 {
			t.Errorf("%s prefix length = %d, want 16 hex chars", gen, len(prefix))
		}
		// Verify it's valid hex
		if _, err := hex.DecodeString(prefix); err != nil {
			t.Errorf("%s prefix is not valid hex: %v", gen, err)
		}
	}
}
