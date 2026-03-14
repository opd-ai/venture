package class

import "encoding/json"

// Serialize encodes a ClassPreset to JSON bytes for persistence.
func (c *ClassPreset) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize decodes JSON bytes into the ClassPreset.
func (c *ClassPreset) Deserialize(data []byte) error {
	return json.Unmarshal(data, c)
}
