// Package engine provides the build template component for saving and loading
// complete character builds. A build template captures attributes, talents, and skills
// as a snapshot that can be restored later or shared as a preset archetype.
package engine

import (
	"encoding/json"
	"fmt"
	"time"
)

// BuildArchetype represents a predefined build style.
type BuildArchetype int

const (
	// BuildArchetypeCustom is a user-created build.
	BuildArchetypeCustom BuildArchetype = iota
	// BuildArchetypeTank focuses on defense, health, and survivability.
	BuildArchetypeTank
	// BuildArchetypeDPS focuses on damage output and critical hits.
	BuildArchetypeDPS
	// BuildArchetypeSupport focuses on healing, buffs, and utility.
	BuildArchetypeSupport
	// BuildArchetypeHybrid balances multiple roles.
	BuildArchetypeHybrid
	// BuildArchetypeBattlemage combines physical and magical damage.
	BuildArchetypeBattlemage
	// BuildArchetypeAssassin focuses on burst damage and evasion.
	BuildArchetypeAssassin
	// BuildArchetypePaladin combines tank and support capabilities.
	BuildArchetypePaladin
)

// String returns the display name for a build archetype.
func (a BuildArchetype) String() string {
	switch a {
	case BuildArchetypeCustom:
		return "Custom"
	case BuildArchetypeTank:
		return "Tank"
	case BuildArchetypeDPS:
		return "DPS"
	case BuildArchetypeSupport:
		return "Support"
	case BuildArchetypeHybrid:
		return "Hybrid"
	case BuildArchetypeBattlemage:
		return "Battlemage"
	case BuildArchetypeAssassin:
		return "Assassin"
	case BuildArchetypePaladin:
		return "Paladin"
	default:
		return "Unknown"
	}
}

// Description returns a brief description of the archetype playstyle.
func (a BuildArchetype) Description() string {
	switch a {
	case BuildArchetypeCustom:
		return "A personalized build created by the player."
	case BuildArchetypeTank:
		return "High health and defense, draws enemy attention."
	case BuildArchetypeDPS:
		return "Maximum damage output, glass cannon approach."
	case BuildArchetypeSupport:
		return "Healing, buffs, and team utility focus."
	case BuildArchetypeHybrid:
		return "Balanced approach across multiple roles."
	case BuildArchetypeBattlemage:
		return "Combines physical attacks with magic spells."
	case BuildArchetypeAssassin:
		return "High burst damage with mobility and stealth."
	case BuildArchetypePaladin:
		return "Tanky frontliner with healing abilities."
	default:
		return "Unknown archetype."
	}
}

// BuildTemplate stores a complete character build configuration.
type BuildTemplate struct {
	// ID is the unique identifier for this template.
	ID string `json:"id"`
	// Name is the display name for the build.
	Name string `json:"name"`
	// Description provides details about the build strategy.
	Description string `json:"description"`
	// Archetype categorizes the build playstyle.
	Archetype BuildArchetype `json:"archetype"`
	// Class is the primary character class for this build.
	Class CharacterClass `json:"class"`
	// Specialization is the primary class specialization.
	Specialization SpecializationType `json:"specialization"`
	// SecondaryClass is the dual-class selection (optional).
	SecondaryClass *CharacterClass `json:"secondary_class,omitempty"`
	// SecondarySpec is the secondary class specialization (optional).
	SecondarySpec SpecializationType `json:"secondary_spec,omitempty"`

	// Attributes stores the allocated attribute points.
	// Map key is CoreAttribute as int, value is allocated points.
	Attributes map[int]int `json:"attributes"`

	// Talents stores the allocated talent points.
	// Map key is talent ID, value is ranks.
	Talents map[string]int `json:"talents"`

	// Skills stores the learned skill levels.
	// Map key is skill ID, value is skill level.
	Skills map[string]int `json:"skills"`

	// SkillTreeID identifies which skill tree this build uses.
	SkillTreeID string `json:"skill_tree_id"`

	// RequiredLevel is the minimum level to fully apply this build.
	RequiredLevel int `json:"required_level"`

	// CreatedAt is when the template was created.
	CreatedAt int64 `json:"created_at"`
	// UpdatedAt is when the template was last modified.
	UpdatedAt int64 `json:"updated_at"`

	// IsPreset indicates if this is a system-provided template.
	IsPreset bool `json:"is_preset"`
	// GenreID indicates genre-specific templates.
	GenreID string `json:"genre_id,omitempty"`
}

// TotalAttributePoints returns the total attribute points in this build.
func (t *BuildTemplate) TotalAttributePoints() int {
	total := 0
	for _, pts := range t.Attributes {
		total += pts
	}
	return total
}

// TotalTalentPoints returns the total talent points in this build.
func (t *BuildTemplate) TotalTalentPoints() int {
	total := 0
	for _, ranks := range t.Talents {
		total += ranks
	}
	return total
}

// TotalSkillPoints returns the total skill points in this build.
func (t *BuildTemplate) TotalSkillPoints() int {
	total := 0
	for _, level := range t.Skills {
		total += level
	}
	return total
}

// Validate checks if the template data is internally consistent.
func (t *BuildTemplate) Validate() error {
	if t.ID == "" {
		return fmt.Errorf("build template ID is required")
	}
	if t.Name == "" {
		return fmt.Errorf("build template name is required")
	}
	if t.RequiredLevel < 1 {
		return fmt.Errorf("required level must be at least 1")
	}
	return nil
}

// Clone creates a deep copy of the template.
func (t *BuildTemplate) Clone() *BuildTemplate {
	clone := &BuildTemplate{
		ID:             t.ID,
		Name:           t.Name,
		Description:    t.Description,
		Archetype:      t.Archetype,
		Class:          t.Class,
		Specialization: t.Specialization,
		SecondarySpec:  t.SecondarySpec,
		SkillTreeID:    t.SkillTreeID,
		RequiredLevel:  t.RequiredLevel,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
		IsPreset:       t.IsPreset,
		GenreID:        t.GenreID,
		Attributes:     make(map[int]int),
		Talents:        make(map[string]int),
		Skills:         make(map[string]int),
	}
	if t.SecondaryClass != nil {
		sc := *t.SecondaryClass
		clone.SecondaryClass = &sc
	}
	for k, v := range t.Attributes {
		clone.Attributes[k] = v
	}
	for k, v := range t.Talents {
		clone.Talents[k] = v
	}
	for k, v := range t.Skills {
		clone.Skills[k] = v
	}
	return clone
}

// BuildTemplateComponent stores an entity's saved build templates.
// Pure data component - all logic lives in BuildTemplateSystem.
type BuildTemplateComponent struct {
	// Templates is the list of saved build templates.
	Templates []*BuildTemplate `json:"templates"`
	// MaxTemplates is the maximum number of templates allowed.
	MaxTemplates int `json:"max_templates"`
	// ActiveTemplateID is the ID of the currently applied template (empty if none).
	ActiveTemplateID string `json:"active_template_id"`

	// PendingApply holds the index of a template waiting to be applied.
	// -1 means no pending apply.
	PendingApply int `json:"pending_apply"`

	// ApplyCooldown is the time in seconds between template applications.
	ApplyCooldown float64 `json:"apply_cooldown"`
	// LastApplyTime is when the last template was applied.
	LastApplyTime float64 `json:"last_apply_time"`

	// PresetTemplates holds system-provided archetype templates.
	PresetTemplates []*BuildTemplate `json:"-"`

	// Dirty indicates if a recalculation is needed.
	Dirty bool `json:"-"`
}

// Type returns the component type identifier.
func (c *BuildTemplateComponent) Type() string {
	return "build_template"
}

// NewBuildTemplateComponent creates a new component with default values.
func NewBuildTemplateComponent() *BuildTemplateComponent {
	return &BuildTemplateComponent{
		Templates:       make([]*BuildTemplate, 0, 5),
		MaxTemplates:    10,
		PendingApply:    -1,
		ApplyCooldown:   30.0, // 30 seconds between full respecs
		PresetTemplates: make([]*BuildTemplate, 0),
	}
}

// AddTemplate adds a new template to the component.
// Returns the index of the new template, or -1 if at capacity.
func (c *BuildTemplateComponent) AddTemplate(template *BuildTemplate) int {
	if len(c.Templates) >= c.MaxTemplates {
		return -1
	}
	c.Templates = append(c.Templates, template)
	return len(c.Templates) - 1
}

// GetTemplate returns the template at the given index, or nil if invalid.
func (c *BuildTemplateComponent) GetTemplate(index int) *BuildTemplate {
	if index < 0 || index >= len(c.Templates) {
		return nil
	}
	return c.Templates[index]
}

// GetTemplateByID returns the template with the given ID, or nil if not found.
func (c *BuildTemplateComponent) GetTemplateByID(id string) *BuildTemplate {
	for _, t := range c.Templates {
		if t.ID == id {
			return t
		}
	}
	return nil
}

// RemoveTemplate removes the template at the given index.
// Returns true if successful.
func (c *BuildTemplateComponent) RemoveTemplate(index int) bool {
	if index < 0 || index >= len(c.Templates) {
		return false
	}
	// Don't allow removing preset templates
	if c.Templates[index].IsPreset {
		return false
	}
	c.Templates = append(c.Templates[:index], c.Templates[index+1:]...)
	return true
}

// UpdateTemplate updates the template at the given index.
// Returns true if successful.
func (c *BuildTemplateComponent) UpdateTemplate(index int, template *BuildTemplate) bool {
	if index < 0 || index >= len(c.Templates) {
		return false
	}
	// Don't allow modifying preset templates
	if c.Templates[index].IsPreset {
		return false
	}
	template.UpdatedAt = time.Now().Unix()
	c.Templates[index] = template
	return true
}

// GetTemplateCount returns the number of saved templates.
func (c *BuildTemplateComponent) GetTemplateCount() int {
	return len(c.Templates)
}

// GetAvailableSlots returns the number of available template slots.
func (c *BuildTemplateComponent) GetAvailableSlots() int {
	return c.MaxTemplates - len(c.Templates)
}

// CanApplyTemplate checks if enough time has passed since last apply.
func (c *BuildTemplateComponent) CanApplyTemplate(currentTime float64) bool {
	return currentTime >= c.LastApplyTime+c.ApplyCooldown
}

// GetApplyCooldownRemaining returns the remaining cooldown time.
func (c *BuildTemplateComponent) GetApplyCooldownRemaining(currentTime float64) float64 {
	remaining := (c.LastApplyTime + c.ApplyCooldown) - currentTime
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RequestApply queues a template for application.
// Returns true if the request was accepted.
func (c *BuildTemplateComponent) RequestApply(index int, currentTime float64) bool {
	if index < 0 || index >= len(c.Templates) {
		return false
	}
	if !c.CanApplyTemplate(currentTime) {
		return false
	}
	c.PendingApply = index
	c.Dirty = true
	return true
}

// MarkApplyComplete clears the pending apply and updates timestamps.
func (c *BuildTemplateComponent) MarkApplyComplete(templateID string, currentTime float64) {
	c.PendingApply = -1
	c.ActiveTemplateID = templateID
	c.LastApplyTime = currentTime
	c.Dirty = false
}

// GetTemplateNames returns a list of all template names.
func (c *BuildTemplateComponent) GetTemplateNames() []string {
	names := make([]string, len(c.Templates))
	for i, t := range c.Templates {
		names[i] = t.Name
	}
	return names
}

// Serialize encodes the component for persistence.
func (c *BuildTemplateComponent) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize decodes the component from persistence data.
func (c *BuildTemplateComponent) Deserialize(data []byte) error {
	if err := json.Unmarshal(data, c); err != nil {
		return err
	}
	// Initialize transient fields
	if c.PresetTemplates == nil {
		c.PresetTemplates = make([]*BuildTemplate, 0)
	}
	c.Dirty = true
	return nil
}
