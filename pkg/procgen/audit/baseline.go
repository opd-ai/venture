package audit

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
)

// BaselineVersion is the version these baseline hashes were generated from.
const BaselineVersion = "1.0.0"

// BaselineSeed is the seed used to generate baseline hashes.
const BaselineSeed int64 = 99999

// baselineHashPrefixes contains the first 8 bytes of SHA256 hashes for each generator.
// Full hashes are stored in baseline_hashes.json for complete validation.
// These prefixes are used for quick validation without file I/O.
var baselineHashPrefixes = map[string]string{
	"Book":      "7e632693e8468f7d",
	"Building":  "00e9ff14a9fe39ff",
	"Companion": "4dda1e3a6cd24740",
	"Entity":    "f0302eb430a7d0cd",
	"Furniture": "25ee7a0f0e519996",
	"Item":      "2b36ce659bf7c7b6",
	"Legendary": "bc3b12fd01179b64",
	"Magic":     "67956a60c3646731",
	"Quest":     "c77dac4cf03d38fb",
	"Recipe":    "547f0b59015c7510",
	"Skills":    "cef103c9c0f578e7",
	"Station":   "9f3cbfe6094f8491",
	"Terrain":   "8ddd8234ec8a966f",
	"Vehicle":   "202dac42c53e9d2a",
}

// GetBaselinePrefix returns the baseline hash prefix for a generator.
// Returns empty string if generator is not in baseline.
func GetBaselinePrefix(generatorName string) string {
	return baselineHashPrefixes[generatorName]
}

// HashMatchesBaseline checks if a hash matches the baseline prefix.
// Returns true if the first 8 bytes match, false otherwise.
func HashMatchesBaseline(generatorName string, hash [32]byte) bool {
	baseline, exists := baselineHashPrefixes[generatorName]
	if !exists {
		return false
	}
	hashPrefix := hex.EncodeToString(hash[:8])
	return hashPrefix == baseline
}

// BaselineHashFile is the filename for full baseline hashes.
const BaselineHashFile = "baseline_hashes.json"

// BaselineHashes represents the full baseline data stored in JSON.
type BaselineHashes struct {
	Version    string            `json:"version"`
	Seed       int64             `json:"seed"`
	Generators map[string]string `json:"generators"`
}

// LoadBaselineHashes loads full baseline hashes from the JSON file.
// Returns nil if file doesn't exist or can't be parsed.
func LoadBaselineHashes(baseDir string) (*BaselineHashes, error) {
	path := filepath.Join(baseDir, BaselineHashFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var hashes BaselineHashes
	if err := json.Unmarshal(data, &hashes); err != nil {
		return nil, err
	}
	return &hashes, nil
}

// SaveBaselineHashes saves full baseline hashes to the JSON file.
func SaveBaselineHashes(baseDir string, hashes *BaselineHashes) error {
	path := filepath.Join(baseDir, BaselineHashFile)
	data, err := json.MarshalIndent(hashes, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
