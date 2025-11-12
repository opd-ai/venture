package engine

// SkillInheritanceComponent tracks abilities that companions learn from their owner.
// Companions gradually learn player skills based on observation and loyalty.
type SkillInheritanceComponent struct {
	// LearnedSkills maps skill IDs to learning progress (0.0-1.0)
	LearnedSkills map[string]float64

	// MaxSkills is the maximum number of skills companion can learn
	MaxSkills int

	// LearningRate affects how fast companion learns (0.0-1.0)
	LearningRate float64

	// RequiredLoyalty is minimum loyalty needed to learn skills
	RequiredLoyalty float64

	// ActiveSkills are skills the companion can currently use
	ActiveSkills []string
}

// Type returns the component type identifier.
func (s *SkillInheritanceComponent) Type() string {
	return "skillinheritance"
}

// NewSkillInheritanceComponent creates a new skill inheritance component.
func NewSkillInheritanceComponent(maxSkills int, learningRate float64) *SkillInheritanceComponent {
	return &SkillInheritanceComponent{
		LearnedSkills:   make(map[string]float64),
		MaxSkills:       maxSkills,
		LearningRate:    learningRate,
		RequiredLoyalty: 50.0, // Default: need 50+ loyalty to learn
		ActiveSkills:    make([]string, 0),
	}
}

// CanLearnSkill checks if companion can learn a new skill.
func (s *SkillInheritanceComponent) CanLearnSkill(skillID string) bool {
	// Check if skill is already learned
	if _, exists := s.LearnedSkills[skillID]; exists {
		return true // Can continue learning
	}

	// Check if max skills reached
	return len(s.LearnedSkills) < s.MaxSkills
}

// GetSkillProgress returns the learning progress for a skill (0.0-1.0).
func (s *SkillInheritanceComponent) GetSkillProgress(skillID string) float64 {
	if progress, exists := s.LearnedSkills[skillID]; exists {
		return progress
	}
	return 0.0
}

// AddSkillProgress increases learning progress for a skill.
// Returns true if skill was fully learned (reached 1.0).
func (s *SkillInheritanceComponent) AddSkillProgress(skillID string, amount float64) bool {
	if !s.CanLearnSkill(skillID) {
		return false
	}

	current := s.GetSkillProgress(skillID)
	current += amount

	// Clamp to 1.0
	if current > 1.0 {
		current = 1.0
	}

	s.LearnedSkills[skillID] = current

	// Activate skill if fully learned and not already active
	if current >= 1.0 {
		s.activateSkill(skillID)
		return true
	}

	return false
}

// activateSkill adds a skill to active skills if not already present.
func (s *SkillInheritanceComponent) activateSkill(skillID string) {
	for _, active := range s.ActiveSkills {
		if active == skillID {
			return // Already active
		}
	}
	s.ActiveSkills = append(s.ActiveSkills, skillID)
}

// IsSkillActive checks if a skill can be used by the companion.
func (s *SkillInheritanceComponent) IsSkillActive(skillID string) bool {
	for _, active := range s.ActiveSkills {
		if active == skillID {
			return true
		}
	}
	return false
}

// GetLearnedSkillCount returns the number of skills being learned or learned.
func (s *SkillInheritanceComponent) GetLearnedSkillCount() int {
	return len(s.LearnedSkills)
}

// GetActiveSkillCount returns the number of skills fully learned and usable.
func (s *SkillInheritanceComponent) GetActiveSkillCount() int {
	return len(s.ActiveSkills)
}
