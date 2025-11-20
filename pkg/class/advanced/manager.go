package advanced

import (
	"fmt"
	"sync"
)

// Manager handles multi-classing, prestige classes, and talent trees
type Manager struct {
	mu          sync.RWMutex
	players     map[string]*AdvancedClassComponent
	talentTrees map[ClassID]*TalentTree
	synergies   []SynergyBonus
	respecCost  RespecCost
}

// NewManager creates a new advanced class manager
func NewManager() *Manager {
	m := &Manager{
		players:     make(map[string]*AdvancedClassComponent),
		talentTrees: make(map[ClassID]*TalentTree),
		synergies:   buildSynergies(),
		respecCost: RespecCost{
			BaseGold:  1000,
			PerRespec: 500,
			MaxCost:   10000,
		},
	}

	m.initializeTalentTrees()
	return m
}

// SetPrimaryClass sets a player's primary class
func (m *Manager) SetPrimaryClass(playerID string, classID ClassID) error {
	_, err := GetClassDefinition(classID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.players[playerID]
	if !exists {
		player = &AdvancedClassComponent{
			TalentPoints: TalentAllocation{
				Talents: make(map[TalentID]int),
			},
		}
		m.players[playerID] = player
	}

	player.PrimaryClass = classID
	return nil
}

// SetSecondaryClass sets a player's secondary class for multi-classing
func (m *Manager) SetSecondaryClass(playerID string, classID ClassID) error {
	_, err := GetClassDefinition(classID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.players[playerID]
	if !exists {
		return fmt.Errorf("player not found: %s", playerID)
	}

	if player.PrimaryClass == classID {
		return fmt.Errorf("secondary class cannot be same as primary")
	}

	player.SecondaryClass = classID
	return nil
}

// SetPrestigeClass assigns a prestige class if requirements are met
func (m *Manager) SetPrestigeClass(playerID string, prestigeID PrestigeClassID) error {
	def, err := GetPrestigeClassDefinition(prestigeID)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.players[playerID]
	if !exists {
		return fmt.Errorf("player not found: %s", playerID)
	}

	if err := m.checkPrestigeRequirements(player, def.Requirements); err != nil {
		return fmt.Errorf("prestige requirements not met: %w", err)
	}

	player.PrestigeClass = prestigeID
	return nil
}

// checkPrestigeRequirements verifies if a player meets prestige class requirements
func (m *Manager) checkPrestigeRequirements(player *AdvancedClassComponent, req PrestigeRequirements) error {
	if player.Level < req.MinLevel {
		return fmt.Errorf("level %d required, have %d", req.MinLevel, player.Level)
	}

	if len(req.RequiredPrimary) > 0 {
		found := false
		for _, reqClass := range req.RequiredPrimary {
			if player.PrimaryClass == reqClass {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("requires one of primary classes: %v", req.RequiredPrimary)
		}
	}

	if len(req.RequiredSecondary) > 0 && player.SecondaryClass != "" {
		found := false
		for _, reqClass := range req.RequiredSecondary {
			if player.SecondaryClass == reqClass {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("requires one of secondary classes: %v", req.RequiredSecondary)
		}
	}

	return nil
}

// AllocateTalent adds a point to a talent
func (m *Manager) AllocateTalent(playerID string, talentID TalentID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.players[playerID]
	if !exists {
		return fmt.Errorf("player not found: %s", playerID)
	}

	tree := m.talentTrees[player.PrimaryClass]
	if tree == nil {
		return fmt.Errorf("no talent tree for class: %s", player.PrimaryClass)
	}

	talent := m.findTalent(tree, talentID)
	if talent == nil {
		return fmt.Errorf("talent not found: %s", talentID)
	}

	currentRank := player.TalentPoints.Talents[talentID]
	if currentRank >= talent.MaxRank {
		return fmt.Errorf("talent already at max rank: %d/%d", currentRank, talent.MaxRank)
	}

	for _, prereq := range talent.Prerequisites {
		if player.TalentPoints.Talents[prereq] == 0 {
			return fmt.Errorf("prerequisite not met: %s", prereq)
		}
	}

	availablePoints := player.TalentPoints.PointsTotal - player.TalentPoints.PointsSpent
	if availablePoints <= 0 {
		return fmt.Errorf("no talent points available")
	}

	player.TalentPoints.Talents[talentID]++
	player.TalentPoints.PointsSpent++

	return nil
}

// findTalent searches for a talent in a talent tree
func (m *Manager) findTalent(tree *TalentTree, id TalentID) *TalentDefinition {
	for i := range tree.Offensive {
		if tree.Offensive[i].ID == id {
			return &tree.Offensive[i]
		}
	}
	for i := range tree.Defensive {
		if tree.Defensive[i].ID == id {
			return &tree.Defensive[i]
		}
	}
	for i := range tree.Utility {
		if tree.Utility[i].ID == id {
			return &tree.Utility[i]
		}
	}
	return nil
}

// RespecTalents resets all talent points for a gold cost
func (m *Manager) RespecTalents(playerID string, gold int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.players[playerID]
	if !exists {
		return fmt.Errorf("player not found: %s", playerID)
	}

	cost := m.calculateRespecCost(player.RespecCount)
	if gold < cost {
		return fmt.Errorf("insufficient gold: need %d, have %d", cost, gold)
	}

	player.TalentPoints.Talents = make(map[TalentID]int)
	player.TalentPoints.PointsSpent = 0
	player.RespecCount++

	return nil
}

// calculateRespecCost calculates the gold cost for respec based on respec count
func (m *Manager) calculateRespecCost(respecCount int) int {
	cost := m.respecCost.BaseGold + (respecCount * m.respecCost.PerRespec)
	if cost > m.respecCost.MaxCost {
		cost = m.respecCost.MaxCost
	}
	return cost
}

// GetRespecCost returns the gold cost for a player's next respec
func (m *Manager) GetRespecCost(playerID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	player, exists := m.players[playerID]
	if !exists {
		return m.respecCost.BaseGold
	}

	return m.calculateRespecCost(player.RespecCount)
}

// CalculateTotalStats computes all stat bonuses from classes and talents
func (m *Manager) CalculateTotalStats(playerID string) (StatBonuses, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	player, exists := m.players[playerID]
	if !exists {
		return StatBonuses{}, fmt.Errorf("player not found: %s", playerID)
	}

	total := StatBonuses{}

	if player.PrimaryClass != "" {
		primaryDef, _ := GetClassDefinition(player.PrimaryClass)
		total = total.Add(primaryDef.BaseStats)
	}

	if player.SecondaryClass != "" {
		secondaryDef, _ := GetClassDefinition(player.SecondaryClass)
		total = total.Add(secondaryDef.BaseStats.Scale(0.5))
	}

	if player.PrestigeClass != "" {
		prestigeDef, _ := GetPrestigeClassDefinition(player.PrestigeClass)
		total = total.Add(prestigeDef.BaseStats)
	}

	synergy := m.getSynergyBonus(player.PrimaryClass, player.SecondaryClass)
	if synergy != nil {
		total = total.Add(synergy.Bonuses)
	}

	tree := m.talentTrees[player.PrimaryClass]
	if tree != nil {
		talentBonuses := m.calculateTalentBonuses(player, tree)
		total = total.Add(talentBonuses)
	}

	return total, nil
}

// calculateTalentBonuses sums all bonuses from allocated talents
func (m *Manager) calculateTalentBonuses(player *AdvancedClassComponent, tree *TalentTree) StatBonuses {
	total := StatBonuses{}

	for talentID, rank := range player.TalentPoints.Talents {
		talent := m.findTalent(tree, talentID)
		if talent != nil {
			scaledBonus := talent.Bonuses.Scale(float64(rank))
			total = total.Add(scaledBonus)
		}
	}

	return total
}

// getSynergyBonus finds a synergy bonus for a class combination
func (m *Manager) getSynergyBonus(primary, secondary ClassID) *SynergyBonus {
	if secondary == "" {
		return nil
	}

	for i := range m.synergies {
		s := &m.synergies[i]
		if (s.Primary == primary && s.Secondary == secondary) ||
			(s.Primary == secondary && s.Secondary == primary) {
			return s
		}
	}

	return nil
}

// GetPlayerClass returns a player's class configuration
func (m *Manager) GetPlayerClass(playerID string) (*AdvancedClassComponent, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	player, exists := m.players[playerID]
	if !exists {
		return nil, fmt.Errorf("player not found: %s", playerID)
	}

	copy := *player
	copy.TalentPoints.Talents = make(map[TalentID]int, len(player.TalentPoints.Talents))
	for k, v := range player.TalentPoints.Talents {
		copy.TalentPoints.Talents[k] = v
	}

	return &copy, nil
}

// SetLevel sets a player's level and awards talent points
func (m *Manager) SetLevel(playerID string, level int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.players[playerID]
	if !exists {
		player = &AdvancedClassComponent{
			TalentPoints: TalentAllocation{
				Talents: make(map[TalentID]int),
			},
		}
		m.players[playerID] = player
	}

	if level < player.Level {
		return fmt.Errorf("cannot decrease level")
	}

	oldLevel := player.Level
	player.Level = level

	pointsGained := (level - oldLevel)
	player.TalentPoints.PointsTotal += pointsGained

	return nil
}

// GetTalentTree returns the talent tree for a class
func (m *Manager) GetTalentTree(classID ClassID) (*TalentTree, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tree, exists := m.talentTrees[classID]
	if !exists {
		return nil, fmt.Errorf("no talent tree for class: %s", classID)
	}

	return tree, nil
}

// GetAllSynergies returns all class synergy bonuses
func (m *Manager) GetAllSynergies() []SynergyBonus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	copy := make([]SynergyBonus, len(m.synergies))
	for i := range m.synergies {
		copy[i] = m.synergies[i]
	}
	return copy
}
