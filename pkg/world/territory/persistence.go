package territory

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// Serialize encodes Territory to JSON bytes for save/load persistence.
func (t *Territory) Serialize() ([]byte, error) {
	return json.Marshal(t)
}

// Deserialize restores Territory state from JSON bytes.
func (t *Territory) Deserialize(data []byte) error {
	return json.Unmarshal(data, t)
}

// Serialize encodes DefensiveStructure to JSON bytes.
func (d *DefensiveStructure) Serialize() ([]byte, error) {
	return json.Marshal(d)
}

// Deserialize restores DefensiveStructure state from JSON bytes.
func (d *DefensiveStructure) Deserialize(data []byte) error {
	return json.Unmarshal(data, d)
}

// Serialize encodes WarDeclaration to JSON bytes.
func (w *WarDeclaration) Serialize() ([]byte, error) {
	return json.Marshal(w)
}

// Deserialize restores WarDeclaration state from JSON bytes.
func (w *WarDeclaration) Deserialize(data []byte) error {
	return json.Unmarshal(data, w)
}

// Serialize encodes Siege to JSON bytes.
func (s *Siege) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize restores Siege state from JSON bytes.
func (s *Siege) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

// managerSnapshot is used for JSON marshaling of all manager state.
type managerSnapshot struct {
	Territories map[string]*Territory      `json:"territories"`
	Wars        map[string]*WarDeclaration `json:"wars"`
	GuildWars   map[string][]string        `json:"guild_wars"`
}

// Save persists all manager state to a JSON file at the given path.
func (m *Manager) Save(path string) error {
	m.mu.RLock()
	snapshot := managerSnapshot{
		Territories: m.territories,
		Wars:        m.wars,
		GuildWars:   m.guildWars,
	}
	m.mu.RUnlock()

	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("territory manager serialize: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("territory manager create dir: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("territory manager write: %w", err)
	}

	log.WithFields(log.Fields{
		"path":        path,
		"territories": len(snapshot.Territories),
		"wars":        len(snapshot.Wars),
		"system_name": "territory",
	}).Info("territory state saved")

	return nil
}

// Load restores all manager state from a JSON file at the given path.
func (m *Manager) Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("territory manager read: %w", err)
	}

	var snapshot managerSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return fmt.Errorf("territory manager deserialize: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if snapshot.Territories != nil {
		m.territories = snapshot.Territories
	}
	if snapshot.Wars != nil {
		m.wars = snapshot.Wars
	}
	if snapshot.GuildWars != nil {
		m.guildWars = snapshot.GuildWars
	}

	log.WithFields(log.Fields{
		"path":        path,
		"territories": len(m.territories),
		"wars":        len(m.wars),
		"system_name": "territory",
	}).Info("territory state loaded")

	return nil
}
