package territory

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// Manager manages guild territories, warfare, and defensive structures.
type Manager struct {
	mu            sync.RWMutex
	territories   map[string]*Territory
	wars          map[string]*WarDeclaration
	guildWars     map[string][]string
	captureRadius float64
	timeProvider  TimeProvider
}

// NewManager creates a new territory manager with the default time provider.
func NewManager() *Manager {
	return NewManagerWithTimeProvider(DefaultTimeProvider())
}

// NewManagerWithTimeProvider creates a new territory manager with a custom time provider.
// This enables deterministic timestamps for testing and reproducible state.
func NewManagerWithTimeProvider(tp TimeProvider) *Manager {
	return &Manager{
		territories:   make(map[string]*Territory),
		wars:          make(map[string]*WarDeclaration),
		guildWars:     make(map[string][]string),
		captureRadius: 50.0,
		timeProvider:  tp,
	}
}

// CreateTerritory creates a new territory zone at the specified chunk coordinates.
func (m *Manager) CreateTerritory(id string, coords TerritoryCoords) (*Territory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.territories[id]; exists {
		log.WithFields(log.Fields{
			"territory_id": id,
		}).Debug("territory already exists")
		return nil, fmt.Errorf("territory already exists: %s", id)
	}

	territory := &Territory{
		ID:              id,
		Coords:          coords,
		OwnerGuildID:    "",
		Status:          StatusNeutral,
		CaptureProgress: 0.0,
		CapturingGuild:  "",
		LastUpdate:      m.timeProvider.Now(),
		Structures:      make([]*DefensiveStructure, 0),
		ResourceBonus:   BaseResourceBonus,
		XPBonus:         BaseXPBonus,
	}

	m.territories[id] = territory
	return territory, nil
}

// GetTerritory retrieves a territory by ID.
// WARNING: Returned territory is a pointer to internal state. Do not mutate directly.
// Use AssignOwner, UpdateCaptureProgress, BuildDefensiveStructure, etc. to modify territories.
func (m *Manager) GetTerritory(id string) (*Territory, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	territory, exists := m.territories[id]
	if !exists {
		log.WithFields(log.Fields{
			"territory_id": id,
		}).Debug("territory not found")
		return nil, fmt.Errorf("territory not found: %s", id)
	}
	return territory, nil
}

// AssignOwner assigns a guild as the owner of a territory.
func (m *Manager) AssignOwner(territoryID, guildID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	territory, exists := m.territories[territoryID]
	if !exists {
		log.WithFields(log.Fields{
			"territory_id": territoryID,
			"guild_id":     guildID,
		}).Debug("territory not found for ownership assignment")
		return fmt.Errorf("territory not found: %s", territoryID)
	}

	territory.OwnerGuildID = guildID
	territory.Status = StatusOwned
	territory.CaptureProgress = 0.0
	territory.CapturingGuild = ""
	territory.LastUpdate = m.timeProvider.Now()

	return nil
}

// UpdateCaptureProgress updates the capture progress for a territory.
func (m *Manager) UpdateCaptureProgress(territoryID string, attackers, defenders int, attackingGuild string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	territory, exists := m.territories[territoryID]
	if !exists {
		log.WithFields(log.Fields{
			"territory_id":    territoryID,
			"attacking_guild": attackingGuild,
		}).Debug("territory not found for capture progress update")
		return fmt.Errorf("territory not found: %s", territoryID)
	}

	if err := validateAttackingGuild(attackingGuild, attackers); err != nil {
		log.WithFields(log.Fields{
			"territory_id":    territoryID,
			"attacking_guild": attackingGuild,
			"attackers":       attackers,
		}).Debug("invalid attacking guild for capture progress")
		return err
	}

	now := m.timeProvider.Now()
	elapsed := now.Sub(territory.LastUpdate).Seconds()

	if attackers > 0 && attackers > defenders {
		m.applyAttackerProgress(territory, attackingGuild, defenders, elapsed)
	} else if defenders > 0 && defenders >= attackers {
		m.applyDefenderProgress(territory, defenders, elapsed)
	}

	territory.LastUpdate = now
	return nil
}

// validateAttackingGuild ensures attackers have a valid guild specified.
func validateAttackingGuild(attackingGuild string, attackers int) error {
	if attackingGuild == "" && attackers > 0 {
		return fmt.Errorf("attacking guild must be specified when attackers > 0")
	}
	return nil
}

// applyAttackerProgress updates territory progress when attackers outnumber defenders.
func (m *Manager) applyAttackerProgress(territory *Territory, attackingGuild string, defenders int, elapsed float64) {
	captureTime := float64(BaseCaptureTime + (defenders * DefenderTimeBonus))
	progressPerSecond := 1.0 / captureTime
	territory.CaptureProgress += progressPerSecond * elapsed
	territory.CapturingGuild = attackingGuild

	if territory.Status == StatusOwned && territory.OwnerGuildID != attackingGuild {
		territory.Status = StatusContested
	}

	if territory.CaptureProgress >= 1.0 {
		m.completeCapture(territory, attackingGuild)
	}
}

// completeCapture finalizes territory capture by the attacking guild.
func (m *Manager) completeCapture(territory *Territory, attackingGuild string) {
	territory.CaptureProgress = 1.0
	territory.OwnerGuildID = attackingGuild
	territory.Status = StatusOwned
	territory.CapturingGuild = ""
}

// applyDefenderProgress reduces capture progress when defenders hold the line.
func (m *Manager) applyDefenderProgress(territory *Territory, defenders int, elapsed float64) {
	decayTime := float64(BaseCaptureTime + (defenders * DefenderTimeBonus))
	decayPerSecond := 1.0 / decayTime
	territory.CaptureProgress -= decayPerSecond * elapsed

	if territory.CaptureProgress < 0.0 {
		m.resetCaptureProgress(territory)
	}
}

// resetCaptureProgress clears capture state and updates territory status.
func (m *Manager) resetCaptureProgress(territory *Territory) {
	territory.CaptureProgress = 0.0
	territory.CapturingGuild = ""
	if territory.OwnerGuildID != "" {
		territory.Status = StatusOwned
	} else {
		territory.Status = StatusNeutral
	}
}

// BuildDefensiveStructure constructs a defensive structure in a territory.
func (m *Manager) BuildDefensiveStructure(territoryID string, structureType StructureType, x, y float64) (*DefensiveStructure, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	territory, exists := m.territories[territoryID]
	if !exists {
		log.WithFields(log.Fields{
			"territory_id":   territoryID,
			"structure_type": structureType.String(),
		}).Debug("territory not found for structure building")
		return nil, fmt.Errorf("territory not found: %s", territoryID)
	}

	if territory.OwnerGuildID == "" {
		log.WithFields(log.Fields{
			"territory_id":   territoryID,
			"structure_type": structureType.String(),
		}).Debug("cannot build in unowned territory")
		return nil, fmt.Errorf("cannot build in unowned territory: %s", territoryID)
	}

	var maxHP, damage float64
	var level int

	switch structureType {
	case StructureTypeWall:
		maxHP = WallBaseHP
		damage = 0.0
		level = 1
	case StructureTypeTower:
		maxHP = TowerBaseHP
		damage = TowerDamage
		level = 1
	case StructureTypeGuard:
		maxHP = GuardBaseHP
		damage = 0.0
		level = GuardLevel
	default:
		log.WithFields(log.Fields{
			"territory_id":   territoryID,
			"structure_type": int(structureType),
		}).Debug("unknown structure type")
		return nil, fmt.Errorf("unknown structure type: %d", structureType)
	}

	structure := &DefensiveStructure{
		ID:            uuid.New().String(),
		Type:          structureType,
		X:             x,
		Y:             y,
		HP:            maxHP,
		MaxHP:         maxHP,
		Damage:        damage,
		Level:         level,
		ConstructedAt: m.timeProvider.Now(),
	}

	territory.Structures = append(territory.Structures, structure)
	return structure, nil
}

// DamageStructure applies damage to a defensive structure.
func (m *Manager) DamageStructure(territoryID, structureID string, damage float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	territory, exists := m.territories[territoryID]
	if !exists {
		log.WithFields(log.Fields{
			"territory_id": territoryID,
			"structure_id": structureID,
		}).Debug("territory not found for structure damage")
		return fmt.Errorf("territory not found: %s", territoryID)
	}

	for i, structure := range territory.Structures {
		if structure.ID == structureID {
			structure.HP -= damage
			if structure.HP <= 0.0 {
				territory.Structures = append(territory.Structures[:i], territory.Structures[i+1:]...)
			}
			return nil
		}
	}

	log.WithFields(log.Fields{
		"territory_id": territoryID,
		"structure_id": structureID,
	}).Debug("structure not found")
	return fmt.Errorf("structure not found: %s", structureID)
}

// DeclareWar creates a formal war declaration between two guilds.
func (m *Manager) DeclareWar(attackerGuild, defenderGuild string) (*WarDeclaration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if attackerGuild == defenderGuild {
		log.WithFields(log.Fields{
			"guild_id": attackerGuild,
		}).Debug("guild cannot declare war on itself")
		return nil, fmt.Errorf("guild cannot declare war on itself")
	}

	for _, war := range m.wars {
		if war.Active &&
			((war.AttackerGuild == attackerGuild && war.DefenderGuild == defenderGuild) ||
				(war.AttackerGuild == defenderGuild && war.DefenderGuild == attackerGuild)) {
			log.WithFields(log.Fields{
				"attacker_guild": attackerGuild,
				"defender_guild": defenderGuild,
			}).Debug("war already exists between these guilds")
			return nil, fmt.Errorf("war already exists between these guilds")
		}
	}

	warID := uuid.New().String()
	now := m.timeProvider.Now()
	war := &WarDeclaration{
		ID:            warID,
		AttackerGuild: attackerGuild,
		DefenderGuild: defenderGuild,
		DeclaredAt:    now,
		EndsAt:        now.Add(WarDurationDays * 24 * 3600 * 1e9), // days in nanoseconds
		Active:        true,
		Cost:          WarDeclarationCost,
	}

	m.wars[warID] = war

	if m.guildWars[attackerGuild] == nil {
		m.guildWars[attackerGuild] = make([]string, 0)
	}
	if m.guildWars[defenderGuild] == nil {
		m.guildWars[defenderGuild] = make([]string, 0)
	}

	m.guildWars[attackerGuild] = append(m.guildWars[attackerGuild], warID)
	m.guildWars[defenderGuild] = append(m.guildWars[defenderGuild], warID)

	return war, nil
}

// EndWar ends an active war declaration.
func (m *Manager) EndWar(warID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	war, exists := m.wars[warID]
	if !exists {
		log.WithFields(log.Fields{
			"war_id": warID,
		}).Debug("war not found")
		return fmt.Errorf("war not found: %s", warID)
	}

	war.Active = false
	war.EndsAt = m.timeProvider.Now()

	return nil
}

// IsAtWar checks if two guilds are currently at war.
func (m *Manager) IsAtWar(guildA, guildB string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, war := range m.wars {
		if war.Active &&
			((war.AttackerGuild == guildA && war.DefenderGuild == guildB) ||
				(war.AttackerGuild == guildB && war.DefenderGuild == guildA)) {
			return true
		}
	}

	return false
}

// GetGuildTerritories returns all territories owned by a guild.
// WARNING: Returned territories are pointers to internal state. Do not mutate directly.
// Use AssignOwner, UpdateCaptureProgress, BuildDefensiveStructure, etc. to modify territories.
func (m *Manager) GetGuildTerritories(guildID string) []*Territory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	territories := make([]*Territory, 0)
	for _, territory := range m.territories {
		if territory.OwnerGuildID == guildID {
			territories = append(territories, territory)
		}
	}
	return territories
}

// GetResourceBonus returns the total resource spawn bonus for a guild based on controlled territories.
func (m *Manager) GetResourceBonus(guildID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, territory := range m.territories {
		if territory.OwnerGuildID == guildID && territory.Status == StatusOwned {
			count++
		}
	}
	return float64(count) * BaseResourceBonus
}

// GetXPBonus returns the total XP gain bonus for a guild based on controlled territories.
func (m *Manager) GetXPBonus(guildID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, territory := range m.territories {
		if territory.OwnerGuildID == guildID && territory.Status == StatusOwned {
			count++
		}
	}
	return float64(count) * BaseXPBonus
}

// GetContestedTerritories returns all territories that are currently being contested.
// WARNING: Returned territories are pointers to internal state. Do not mutate directly.
func (m *Manager) GetContestedTerritories() []*Territory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	contested := make([]*Territory, 0)
	for _, territory := range m.territories {
		if territory.Status == StatusContested {
			contested = append(contested, territory)
		}
	}
	return contested
}

// GetAllTerritories returns all territories.
// WARNING: Returned territories are pointers to internal state. Do not mutate directly.
func (m *Manager) GetAllTerritories() []*Territory {
	m.mu.RLock()
	defer m.mu.RUnlock()

	territories := make([]*Territory, 0, len(m.territories))
	for _, territory := range m.territories {
		territories = append(territories, territory)
	}
	return territories
}

// GetActiveWars returns all active war declarations.
// WARNING: Returned wars are pointers to internal state. Do not mutate directly.
func (m *Manager) GetActiveWars() []*WarDeclaration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	wars := make([]*WarDeclaration, 0)
	for _, war := range m.wars {
		if war.Active {
			wars = append(wars, war)
		}
	}
	return wars
}

// GetGuildWars returns all wars (active and inactive) involving a guild.
// WARNING: Returned wars are pointers to internal state. Do not mutate directly.
func (m *Manager) GetGuildWars(guildID string) []*WarDeclaration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	warIDs, exists := m.guildWars[guildID]
	if !exists {
		return make([]*WarDeclaration, 0)
	}

	wars := make([]*WarDeclaration, 0, len(warIDs))
	for _, warID := range warIDs {
		if war, exists := m.wars[warID]; exists {
			wars = append(wars, war)
		}
	}
	return wars
}
