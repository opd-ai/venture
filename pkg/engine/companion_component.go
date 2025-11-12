package engine

// CompanionType represents different types of companions
type CompanionType int

const (
	CompanionTypePet CompanionType = iota
	CompanionTypeSummon
	CompanionTypeHireling
	CompanionTypeElemental
	CompanionTypeUndead
	CompanionTypeRobot
	CompanionTypeSpirit
	CompanionTypeInsect
)

// BehaviorMode defines companion AI behavior
type BehaviorMode int

const (
	BehaviorAggressive BehaviorMode = iota
	BehaviorDefensive
	BehaviorPassive
)

// CommandType represents commands that can be given to companions
type CommandType int

const (
	CommandFollow CommandType = iota
	CommandStay
	CommandAttack
	CommandDefend
	CommandGather
	CommandScout
)

// CompanionComponent tracks companion state
type CompanionComponent struct {
	OwnerID       uint64
	CompanionType CompanionType
	Loyalty       float64 // 0-100, affects obedience
	Experience    float64
	Level         int
	Behavior      BehaviorMode
	Commands      []CommandType
}

// Type returns the component type
func (c CompanionComponent) Type() string {
	return "companion"
}

// CompanionStatsComponent holds companion-specific stats
type CompanionStatsComponent struct {
	Attack  float64
	Defense float64
	Speed   float64
	HP      float64
	MaxHP   float64
}

// Type returns the component type
func (c CompanionStatsComponent) Type() string {
	return "companionstats"
}
