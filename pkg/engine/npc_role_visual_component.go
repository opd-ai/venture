// Package engine provides the NpcRoleVisualComponent which stores a humanoid
// entity's visual role (mage, warrior, merchant, etc.) for role-aware sprite
// template selection.
package engine

// NpcRoleVisualComponent holds the visual archetype used to select a
// role-specific aerial anatomy template during sprite generation. The role
// string maps directly to sprites.VisualRole values.
type NpcRoleVisualComponent struct {
	// Role is the visual archetype: "mage", "warrior", "knight", "rogue",
	// "merchant", "ranger", "priest", or "" for generic humanoid.
	Role string
}

// Type returns the component type identifier.
func (n *NpcRoleVisualComponent) Type() string {
	return "npc_role_visual"
}

// NewNpcRoleVisualComponent creates a new NPC role visual component.
func NewNpcRoleVisualComponent(role string) *NpcRoleVisualComponent {
	return &NpcRoleVisualComponent{Role: role}
}
