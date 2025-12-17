// Package engine provides stub types for the CarryOver system.
// These types support carry-over functionality for cosmetics, skills, and achievements.
//
// Phase 112: Carry-Over System Support Types
package engine

import (
	"sync"
)

// CosmeticComponent tracks unlocked cosmetic items for an entity.
// Cosmetics are visual customizations that persist across NG+ cycles.
type CosmeticComponent struct {
	mu sync.RWMutex

	// UnlockedCosmetics contains IDs of all unlocked cosmetic items
	UnlockedCosmetics []string

	// ActiveCosmetics contains currently equipped cosmetic IDs per slot
	ActiveCosmetics map[string]string

	// FavoriteCosmetics contains IDs marked as favorites
	FavoriteCosmetics []string
}

// Type returns the component type identifier.
func (c *CosmeticComponent) Type() string {
	return "cosmetic"
}

// NewCosmeticComponent creates a new cosmetic component.
func NewCosmeticComponent() *CosmeticComponent {
	return &CosmeticComponent{
		UnlockedCosmetics: []string{},
		ActiveCosmetics:   make(map[string]string),
		FavoriteCosmetics: []string{},
	}
}

// UnlockCosmetic adds a cosmetic ID to the unlocked list.
func (c *CosmeticComponent) UnlockCosmetic(cosmeticID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, id := range c.UnlockedCosmetics {
		if id == cosmeticID {
			return false // Already unlocked
		}
	}
	c.UnlockedCosmetics = append(c.UnlockedCosmetics, cosmeticID)
	return true
}

// HasCosmetic checks if a cosmetic is unlocked.
func (c *CosmeticComponent) HasCosmetic(cosmeticID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, id := range c.UnlockedCosmetics {
		if id == cosmeticID {
			return true
		}
	}
	return false
}

// GetUnlockedCosmetics returns a copy of all unlocked cosmetics.
func (c *CosmeticComponent) GetUnlockedCosmetics() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]string, len(c.UnlockedCosmetics))
	copy(result, c.UnlockedCosmetics)
	return result
}

// EquipCosmetic sets a cosmetic as active in a slot.
func (c *CosmeticComponent) EquipCosmetic(slot, cosmeticID string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if unlocked
	found := false
	for _, id := range c.UnlockedCosmetics {
		if id == cosmeticID {
			found = true
			break
		}
	}
	if !found {
		return false
	}

	c.ActiveCosmetics[slot] = cosmeticID
	return true
}

// GetActiveCosmetic returns the active cosmetic for a slot.
func (c *CosmeticComponent) GetActiveCosmetic(slot string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ActiveCosmetics[slot]
}

// SkillBookComponent tracks learned skills for an entity.
// Used for carry-over skill selection.
type SkillBookComponent struct {
	mu sync.RWMutex

	// LearnedSkills maps skill ID to skill data
	LearnedSkills map[string]*LearnedSkill

	// SkillPoints available for learning new skills
	SkillPoints int

	// TotalSkillsLearned count
	TotalSkillsLearned int
}

// Type returns the component type identifier.
func (s *SkillBookComponent) Type() string {
	return "skill_book"
}

// NewSkillBookComponent creates a new skill book component.
func NewSkillBookComponent() *SkillBookComponent {
	return &SkillBookComponent{
		LearnedSkills: make(map[string]*LearnedSkill),
		SkillPoints:   0,
	}
}

// LearnedSkill represents a skill the entity has learned.
type LearnedSkill struct {
	Name        string
	Type        string // offensive, support, passive, utility
	Level       int
	MaxLevel    int
	Experience  int
	CooldownSec float64
	ManaCost    int
}

// LearnSkillByID adds a skill to the skill book.
func (s *SkillBookComponent) LearnSkillByID(skillID string, skill *LearnedSkill) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.LearnedSkills[skillID]; exists {
		return false
	}

	s.LearnedSkills[skillID] = skill
	s.TotalSkillsLearned++
	return true
}

// HasSkill checks if a skill is learned.
func (s *SkillBookComponent) HasSkill(skillID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.LearnedSkills[skillID]
	return exists
}

// GetSkill returns a learned skill by ID.
func (s *SkillBookComponent) GetSkill(skillID string) *LearnedSkill {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.LearnedSkills[skillID]
}

// GetAllSkillIDs returns a list of all learned skill IDs.
func (s *SkillBookComponent) GetAllSkillIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.LearnedSkills))
	for id := range s.LearnedSkills {
		ids = append(ids, id)
	}
	return ids
}

// SpellComponent tracks known spells for an entity.
// Used for carry-over spell selection.
type SpellComponent struct {
	mu sync.RWMutex

	// KnownSpells maps spell ID to spell data
	KnownSpells map[string]*KnownSpell

	// MaxSpellSlots is the maximum active spells
	MaxSpellSlots int

	// ActiveSpells are currently equipped spells
	ActiveSpells []string
}

// Type returns the component type identifier.
func (s *SpellComponent) Type() string {
	return "spell"
}

// NewSpellComponent creates a new spell component.
func NewSpellComponent() *SpellComponent {
	return &SpellComponent{
		KnownSpells:   make(map[string]*KnownSpell),
		MaxSpellSlots: 4,
		ActiveSpells:  []string{},
	}
}

// KnownSpell represents a spell the entity has learned.
type KnownSpell struct {
	Name       string
	Type       string // fire, ice, lightning, holy, dark
	SpellLevel int
	ManaCost   int
	CastTime   float64
	Cooldown   float64
	Damage     int
	Healing    int
}

// LearnSpell adds a spell to known spells.
func (s *SpellComponent) LearnSpell(spellID string, spell *KnownSpell) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.KnownSpells[spellID]; exists {
		return false
	}

	s.KnownSpells[spellID] = spell
	return true
}

// HasSpell checks if a spell is known.
func (s *SpellComponent) HasSpell(spellID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.KnownSpells[spellID]
	return exists
}

// GetSpell returns a known spell by ID.
func (s *SpellComponent) GetSpell(spellID string) *KnownSpell {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.KnownSpells[spellID]
}

// GetAllSpellIDs returns a list of all known spell IDs.
func (s *SpellComponent) GetAllSpellIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.KnownSpells))
	for id := range s.KnownSpells {
		ids = append(ids, id)
	}
	return ids
}

// Unlock adds an achievement ID to the unlocked list.
// This is a convenience method for carry-over that uses string IDs.
// It adds a placeholder achievement to track the unlock.
func (a *AchievementComponent) Unlock(achievementID string) bool {
	// Check if already unlocked
	for _, ach := range a.Achievements {
		if ach.Description == achievementID {
			return false
		}
	}
	// Add placeholder achievement using Description to store the string ID
	a.Achievements = append(a.Achievements, Achievement{
		Type:        0,             // Placeholder type
		UnlockedAt:  0,             // Will be set properly elsewhere
		Description: achievementID, // Store string ID in description
	})
	return true
}

// IsUnlocked checks if an achievement is unlocked by string ID.
func (a *AchievementComponent) IsUnlocked(achievementID string) bool {
	for _, ach := range a.Achievements {
		if ach.Description == achievementID {
			return true
		}
	}
	return false
}

// Unlocked returns a list of unlocked achievement IDs (for carry-over).
func (a *AchievementComponent) GetUnlockedIDs() []string {
	result := make([]string, len(a.Achievements))
	for i, ach := range a.Achievements {
		result[i] = ach.Type.String()
	}
	return result
}
