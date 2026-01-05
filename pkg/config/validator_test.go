package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidator_ValidatePort(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{"valid port 8080", "8080", false},
		{"valid port 1024", "1024", false},
		{"valid port 65535", "65535", false},
		{"valid port 3000", "3000", false},
		{"invalid - too low", "1023", true},
		{"invalid - requires root", "80", true},
		{"invalid - too high", "65536", true},
		{"invalid - not a number", "abc", true},
		{"invalid - empty", "", true},
		{"invalid - negative", "-1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePort(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePort(%q) error = %v, wantErr %v", tt.port, err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateMaxPlayers(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name       string
		maxPlayers int
		wantErr    bool
	}{
		{"valid - 1 player", 1, false},
		{"valid - 4 players", 4, false},
		{"valid - 50 players", 50, false},
		{"valid - 100 players", 100, false},
		{"invalid - 0 players", 0, true},
		{"invalid - negative", -1, true},
		{"invalid - too many", 101, true},
		{"invalid - way too many", 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMaxPlayers(tt.maxPlayers)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMaxPlayers(%d) error = %v, wantErr %v", tt.maxPlayers, err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateTickRate(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name     string
		tickRate int
		wantErr  bool
	}{
		{"valid - 1 Hz", 1, false},
		{"valid - 20 Hz", 20, false},
		{"valid - 30 Hz", 30, false},
		{"valid - 60 Hz", 60, false},
		{"invalid - 0 Hz", 0, true},
		{"invalid - negative", -1, true},
		{"invalid - too high", 61, true},
		{"invalid - way too high", 1000, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTickRate(tt.tickRate)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateTickRate(%d) error = %v, wantErr %v", tt.tickRate, err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateGenre(t *testing.T) {
	validator := NewValidator()

	tests := []struct {
		name    string
		genre   string
		wantErr bool
	}{
		{"valid - fantasy", "fantasy", false},
		{"valid - scifi", "scifi", false},
		{"valid - horror", "horror", false},
		{"valid - cyberpunk", "cyberpunk", false},
		{"valid - postapocalyptic", "postapocalyptic", false},
		{"invalid - unknown genre", "medieval", true},
		{"invalid - empty", "", true},
		{"invalid - typo", "fantazy", true},
		{"invalid - case mismatch", "Fantasy", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateGenre(tt.genre)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateGenre(%q) error = %v, wantErr %v", tt.genre, err, tt.wantErr)
			}
		})
	}
}

func TestValidator_ValidateDirectory(t *testing.T) {
	validator := NewValidator()

	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	existingDir := filepath.Join(tmpDir, "existing")
	if err := os.MkdirAll(existingDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Create a file (not a directory) for testing
	notADir := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(notADir, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		create  bool
		wantErr bool
	}{
		{"valid - existing directory", existingDir, false, false},
		{"valid - create new directory", filepath.Join(tmpDir, "newdir"), true, false},
		{"invalid - nonexistent without create", filepath.Join(tmpDir, "missing"), false, true},
		{"invalid - not a directory", notADir, false, true},
		{"invalid - empty path", "", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateDirectory(tt.path, tt.create)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateDirectory(%q, %v) error = %v, wantErr %v", tt.path, tt.create, err, tt.wantErr)
			}
		})
	}
}

func TestValidator_GetAvailableGenres(t *testing.T) {
	validator := NewValidator()
	genres := validator.GetAvailableGenres()

	// Should have exactly 5 genres
	if len(genres) != 5 {
		t.Errorf("GetAvailableGenres() returned %d genres, want 5", len(genres))
	}

	// Check that expected genres are present
	expectedGenres := map[string]bool{
		"fantasy":         false,
		"scifi":           false,
		"horror":          false,
		"cyberpunk":       false,
		"postapocalyptic": false,
	}

	for _, genre := range genres {
		if _, exists := expectedGenres[genre]; exists {
			expectedGenres[genre] = true
		}
	}

	for genre, found := range expectedGenres {
		if !found {
			t.Errorf("Expected genre %q not found in available genres", genre)
		}
	}
}

func TestValidator_ValidateAll(t *testing.T) {
	validator := NewValidator()
	tmpDir := t.TempDir()

	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid complete config",
			config: &Config{
				Port:               "8080",
				MaxPlayers:         4,
				ValidateMaxPlayers: true,
				TickRate:           20,
				ValidateTickRate:   true,
				Genre:              "fantasy",
				SaveDir:            tmpDir,
				CreateDirs:         false,
			},
			wantErr: false,
		},
		{
			name: "valid minimal config",
			config: &Config{
				Genre: "scifi",
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: &Config{
				Port:  "80",
				Genre: "fantasy",
			},
			wantErr: true,
		},
		{
			name: "invalid max players",
			config: &Config{
				MaxPlayers:         0,
				ValidateMaxPlayers: true,
			},
			wantErr: true,
		},
		{
			name: "invalid tick rate",
			config: &Config{
				TickRate:         100,
				ValidateTickRate: true,
			},
			wantErr: true,
		},
		{
			name: "invalid genre",
			config: &Config{
				Genre: "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid save directory",
			config: &Config{
				SaveDir:    "/nonexistent/invalid/path",
				CreateDirs: false,
			},
			wantErr: true,
		},
		{
			name: "create missing directory",
			config: &Config{
				SaveDir:    filepath.Join(tmpDir, "newsave"),
				CreateDirs: true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateAll(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewValidator verifies that validator is initialized correctly.
func TestNewValidator(t *testing.T) {
	validator := NewValidator()

	if validator == nil {
		t.Fatal("NewValidator() returned nil")
	}

	if validator.validGenres == nil {
		t.Fatal("NewValidator() did not initialize validGenres map")
	}

	if len(validator.validGenres) != 5 {
		t.Errorf("NewValidator() initialized with %d genres, want 5", len(validator.validGenres))
	}
}
