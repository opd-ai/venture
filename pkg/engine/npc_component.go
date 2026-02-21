package engine

// NPCComponent marks an entity as a non-player character.
// Used by systems that need to distinguish NPCs from player entities
// (e.g., HumanoidTextureSystem assigns textures to NPC entities).
type NPCComponent struct {
	// Role describes the NPC's functional role (e.g., "merchant", "guard", "quest_giver").
	Role string

	// Name is the NPC's display name.
	Name string
}

// Type returns the component type identifier.
func (n *NPCComponent) Type() string {
	return "npc"
}
