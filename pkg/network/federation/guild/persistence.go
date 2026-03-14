// persistence.go handles guild federation state persistence including
// serialization and deserialization of cross-server guild data.
package guild

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
)

// Guild state persistence with compression.
//
// This file handles serialization and deserialization of guild data using
// gzip-compressed JSON format. Includes decompression bomb protection via
// size limits to prevent memory exhaustion attacks.
//
// Code relocated from: manager.go

// Code relocated from: manager.go
// Guild save/load persistence operations isolated for clarity.

// Save serializes all guilds to gzip-compressed JSON
// Originally defined in: manager.go
func (m *Manager) Save() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data, err := json.Marshal(m.guilds)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal guilds: %w", err)
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("failed to compress: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip: %w", err)
	}

	return buf.Bytes(), nil
}

// Load deserializes guilds from gzip-compressed JSON.
// Returns ErrGuildDataSizeExceeded if the decompressed data exceeds MaxGuildDataSize.
// Originally defined in: manager.go
func (m *Manager) Load(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decompress: %w", err)
	}
	defer gz.Close()

	// Use LimitReader to protect against decompression bombs.
	// Read up to MaxGuildDataSize + 1 to detect if limit is exceeded.
	limitedReader := io.LimitReader(gz, MaxGuildDataSize+1)

	var buf bytes.Buffer
	n, err := buf.ReadFrom(limitedReader)
	if err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}

	// If we read exactly MaxGuildDataSize+1, the data exceeded the limit
	if n > MaxGuildDataSize {
		return ErrGuildDataSizeExceeded
	}

	guilds := make(map[string]*Guild)
	if err := json.Unmarshal(buf.Bytes(), &guilds); err != nil {
		return fmt.Errorf("failed to unmarshal: %w", err)
	}

	m.guilds = guilds
	return nil
}
