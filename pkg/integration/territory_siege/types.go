package territory_siege

import (
	"time"
)

// SiegePhase represents the current phase of a siege.
type SiegePhase int

const (
	// PhasePreparation is the initial phase where defenders can call reinforcements.
	// Duration: 1 hour
	PhasePreparation SiegePhase = iota

	// PhaseAssault is the combat phase where fighting occurs.
	// Duration: 2 hours
	PhaseAssault

	// PhaseResolution is the final phase where victory conditions are checked.
	// Duration: immediate
	PhaseResolution
)

// String returns the name of the siege phase.
func (sp SiegePhase) String() string {
	switch sp {
	case PhasePreparation:
		return "Preparation"
	case PhaseAssault:
		return "Assault"
	case PhaseResolution:
		return "Resolution"
	default:
		return "Unknown"
	}
}

// StructureType represents the type of defensive structure.
type StructureType int

const (
	// StructureWall blocks movement, high HP (1000-5000)
	StructureWall StructureType = iota

	// StructureTower provides defensive fire, medium HP (500-1500)
	StructureTower

	// StructureGate is an entry point, destroyable, medium HP (800-2000)
	StructureGate

	// StructureBarracks spawns defender NPCs, low HP (300-800)
	StructureBarracks

	// StructureKeep is the guild hall defense, very high HP (10000-20000)
	StructureKeep
)

// String returns the name of the structure type.
func (st StructureType) String() string {
	switch st {
	case StructureWall:
		return "Wall"
	case StructureTower:
		return "Tower"
	case StructureGate:
		return "Gate"
	case StructureBarracks:
		return "Barracks"
	case StructureKeep:
		return "Keep"
	default:
		return "Unknown"
	}
}

// DefensiveStructure represents a defensive structure in a territory.
type DefensiveStructure struct {
	StructureID  string
	Type         StructureType
	X            float64
	Y            float64
	MaxHP        int
	CurrentHP    int
	LastDamageAt int64 // Unix timestamp
	IsDestroyed  bool
}

// Siege represents an active territory siege.
type Siege struct {
	SiegeID             string
	AttackerGuildID     string
	DefenderGuildID     string
	ZoneID              string
	CurrentPhase        SiegePhase
	PhaseStartTime      int64    // Unix timestamp
	PreparationDuration int64    // Seconds (default: 3600 = 1 hour)
	AssaultDuration     int64    // Seconds (default: 7200 = 2 hours)
	ReinforcementGuilds []string // Allied guilds joining defense
	DefensiveStructures []*DefensiveStructure
	AttackerPlayerCount int
	DefenderPlayerCount int
	Victor              string // GuildID of winner, empty if ongoing
	TreasuryLoot        int    // Gold looted from defender treasury
	LastUpdate          int64  // Unix timestamp
}

// VictoryCondition represents the condition under which a siege was won.
type VictoryCondition int

const (
	// VictoryConditionNone means the siege is still ongoing.
	VictoryConditionNone VictoryCondition = iota

	// VictoryConditionAllPointsCaptured means attackers captured all control points.
	VictoryConditionAllPointsCaptured

	// VictoryConditionGuildHallDestroyed means attackers destroyed the keep.
	VictoryConditionGuildHallDestroyed

	// VictoryConditionTimeExpired means defenders held until assault phase ended.
	VictoryConditionTimeExpired

	// VictoryConditionAttackersEliminated means all attackers were defeated.
	VictoryConditionAttackersEliminated
)

// String returns the name of the victory condition.
func (vc VictoryCondition) String() string {
	switch vc {
	case VictoryConditionNone:
		return "Ongoing"
	case VictoryConditionAllPointsCaptured:
		return "All Control Points Captured"
	case VictoryConditionGuildHallDestroyed:
		return "Guild Hall Destroyed"
	case VictoryConditionTimeExpired:
		return "Time Expired"
	case VictoryConditionAttackersEliminated:
		return "Attackers Eliminated"
	default:
		return "Unknown"
	}
}

// SiegeResult contains the outcome of a completed siege.
type SiegeResult struct {
	VictorGuildID         string
	VictoryCondition      VictoryCondition
	DurationSeconds       int64
	TreasuryLoot          int
	CapturedControlPoints int
	DestroyedStructures   int
	RewardMultiplier      float64 // 1.0-3.0 based on performance
}

// IsStructureDestroyed returns true if the structure is destroyed.
func (ds *DefensiveStructure) IsStructureDestroyed() bool {
	return ds.IsDestroyed || ds.CurrentHP <= 0
}

// TakeDamage applies damage to the structure and updates state.
func (ds *DefensiveStructure) TakeDamage(amount int) {
	if ds.IsDestroyed {
		return
	}

	ds.CurrentHP -= amount
	if ds.CurrentHP <= 0 {
		ds.CurrentHP = 0
		ds.IsDestroyed = true
	}
	ds.LastDamageAt = time.Now().Unix()
}

// GetElapsedTime returns the time elapsed since the current phase started.
func (s *Siege) GetElapsedTime() int64 {
	return time.Now().Unix() - s.PhaseStartTime
}

// GetRemainingTime returns the time remaining in the current phase (negative if expired).
func (s *Siege) GetRemainingTime() int64 {
	elapsed := s.GetElapsedTime()
	switch s.CurrentPhase {
	case PhasePreparation:
		return s.PreparationDuration - elapsed
	case PhaseAssault:
		return s.AssaultDuration - elapsed
	default:
		return 0
	}
}

// ShouldAdvancePhase returns true if the current phase should end.
func (s *Siege) ShouldAdvancePhase() bool {
	// Resolution phase doesn't advance
	if s.CurrentPhase == PhaseResolution {
		return false
	}
	return s.GetRemainingTime() <= 0
}

// CountDestroyedStructures returns the number of destroyed defensive structures.
func (s *Siege) CountDestroyedStructures() int {
	count := 0
	for _, structure := range s.DefensiveStructures {
		if structure.IsStructureDestroyed() {
			count++
		}
	}
	return count
}

// GetDestructionPercentage returns the percentage of destroyed structures (0.0-1.0).
func (s *Siege) GetDestructionPercentage() float64 {
	if len(s.DefensiveStructures) == 0 {
		return 0.0
	}
	destroyed := s.CountDestroyedStructures()
	return float64(destroyed) / float64(len(s.DefensiveStructures))
}

// SiegeParticipantComponent tracks an entity's participation in a territory siege.
type SiegeParticipantComponent struct {
	SiegeID      string
	IsAttacker   bool  // true if attacker, false if defender
	IsActive     bool  // true if siege is ongoing
	LastSeenTime int64 // Unix timestamp for tracking presence
}

// Type returns the component type identifier.
func (c *SiegeParticipantComponent) Type() string {
	return "siege_participant"
}
