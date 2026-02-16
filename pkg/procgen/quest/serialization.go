// Package quest provides serialization support for quest types.
// This file implements JSON-based Serialize/Deserialize methods for
// Quest, Objective, and Reward types to support save/load persistence.
package quest

import "encoding/json"

// Serialize encodes the Quest as JSON bytes for persistence.
func (q *Quest) Serialize() ([]byte, error) {
	return json.Marshal(q)
}

// Deserialize decodes JSON bytes into the Quest, restoring saved state.
func (q *Quest) Deserialize(data []byte) error {
	return json.Unmarshal(data, q)
}

// Serialize encodes the Objective as JSON bytes for persistence.
func (o *Objective) Serialize() ([]byte, error) {
	return json.Marshal(o)
}

// Deserialize decodes JSON bytes into the Objective, restoring saved state.
func (o *Objective) Deserialize(data []byte) error {
	return json.Unmarshal(data, o)
}

// Serialize encodes the Reward as JSON bytes for persistence.
func (r *Reward) Serialize() ([]byte, error) {
	return json.Marshal(r)
}

// Deserialize decodes JSON bytes into the Reward, restoring saved state.
func (r *Reward) Deserialize(data []byte) error {
	return json.Unmarshal(data, r)
}
