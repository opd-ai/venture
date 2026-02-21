// Package engine provides the TagComponent for entity classification.
// Tags are used to identify entity types and enable targeted behaviors.
package engine

// TagComponent stores string tags for entity classification.
// Used by AI systems to identify interactable objects and entity types.
type TagComponent struct {
	// Tags is a list of string identifiers for this entity.
	// Common tags: "lever", "trap", "door", "hazard", "interactable",
	// "enemy", "ally", "npc", "boss", "merchant"
	Tags []string
}

// Type returns the component type identifier.
func (t *TagComponent) Type() string {
	return "tag"
}

// NewTagComponent creates a new tag component with the given tags.
func NewTagComponent(tags ...string) *TagComponent {
	return &TagComponent{
		Tags: tags,
	}
}

// HasTag checks if the component has a specific tag.
func (t *TagComponent) HasTag(tag string) bool {
	for _, existingTag := range t.Tags {
		if existingTag == tag {
			return true
		}
	}
	return false
}

// AddTag adds a tag if it doesn't already exist.
func (t *TagComponent) AddTag(tag string) {
	if !t.HasTag(tag) {
		t.Tags = append(t.Tags, tag)
	}
}

// RemoveTag removes a tag if it exists.
func (t *TagComponent) RemoveTag(tag string) {
	for i, existingTag := range t.Tags {
		if existingTag == tag {
			t.Tags = append(t.Tags[:i], t.Tags[i+1:]...)
			return
		}
	}
}
