package raids

import "encoding/json"

// Serialize encodes RaidInstance to JSON bytes for save/load persistence.
func (r *RaidInstance) Serialize() ([]byte, error) {
	return json.Marshal(r)
}

// Deserialize restores RaidInstance state from JSON bytes.
func (r *RaidInstance) Deserialize(data []byte) error {
	return json.Unmarshal(data, r)
}

// Serialize encodes PlayerLockout to JSON bytes for save/load persistence.
func (p *PlayerLockout) Serialize() ([]byte, error) {
	return json.Marshal(p)
}

// Deserialize restores PlayerLockout state from JSON bytes.
func (p *PlayerLockout) Deserialize(data []byte) error {
	return json.Unmarshal(data, p)
}
