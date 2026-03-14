package story

import "encoding/json"

// Serialize encodes a StorySequence to JSON bytes.
func (s *StorySequence) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize decodes JSON bytes into the StorySequence.
func (s *StorySequence) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

// Serialize encodes an ArchaeologicalSite to JSON bytes.
func (s *ArchaeologicalSite) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize decodes JSON bytes into the ArchaeologicalSite.
func (s *ArchaeologicalSite) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

// Serialize encodes a BranchingNarrative to JSON bytes.
func (s *BranchingNarrative) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize decodes JSON bytes into the BranchingNarrative.
func (s *BranchingNarrative) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

// Serialize encodes a Timeline to JSON bytes.
func (s *Timeline) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize decodes JSON bytes into the Timeline.
func (s *Timeline) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}

// Serialize encodes a CrossDungeonStory to JSON bytes.
func (s *CrossDungeonStory) Serialize() ([]byte, error) {
	return json.Marshal(s)
}

// Deserialize decodes JSON bytes into the CrossDungeonStory.
func (s *CrossDungeonStory) Deserialize(data []byte) error {
	return json.Unmarshal(data, s)
}
