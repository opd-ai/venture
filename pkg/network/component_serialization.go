// Package network provides component serialization for networking.
// This file implements serialization and deserialization of ECS components
// for efficient network transmission.
package network

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"math"

	"github.com/sirupsen/logrus"
)

// ComponentSerializer provides methods for serializing ECS components to/from bytes.
type ComponentSerializer struct{}

// NewComponentSerializer creates a new component serializer.
func NewComponentSerializer() *ComponentSerializer {
	return &ComponentSerializer{}
}

// SerializePosition serializes a position component (X, Y as float64).
func (s *ComponentSerializer) SerializePosition(x, y float64) []byte {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[0:8], math.Float64bits(x))
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(y))
	return buf
}

// DeserializePosition deserializes a position component.
func (s *ComponentSerializer) DeserializePosition(data []byte) (x, y float64, err error) {
	if len(data) != 16 {
		logrus.WithFields(logrus.Fields{
			"component_type": "position",
			"data_length":    len(data),
			"expected":       16,
		}).Warn("invalid position data length")
		return 0, 0, fmt.Errorf("invalid position data length: %d (expected 16)", len(data))
	}
	x = math.Float64frombits(binary.LittleEndian.Uint64(data[0:8]))
	y = math.Float64frombits(binary.LittleEndian.Uint64(data[8:16]))
	return x, y, nil
}

// SerializeVelocity serializes a velocity component (VX, VY as float64).
func (s *ComponentSerializer) SerializeVelocity(vx, vy float64) []byte {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[0:8], math.Float64bits(vx))
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(vy))
	return buf
}

// DeserializeVelocity deserializes a velocity component.
func (s *ComponentSerializer) DeserializeVelocity(data []byte) (vx, vy float64, err error) {
	if len(data) != 16 {
		logrus.WithFields(logrus.Fields{
			"component_type": "velocity",
			"data_length":    len(data),
			"expected":       16,
		}).Warn("invalid velocity data length")
		return 0, 0, fmt.Errorf("invalid velocity data length: %d (expected 16)", len(data))
	}
	vx = math.Float64frombits(binary.LittleEndian.Uint64(data[0:8]))
	vy = math.Float64frombits(binary.LittleEndian.Uint64(data[8:16]))
	return vx, vy, nil
}

// SerializeHealth serializes a health component (Current, Max as float64).
func (s *ComponentSerializer) SerializeHealth(current, max float64) []byte {
	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[0:8], math.Float64bits(current))
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(max))
	return buf
}

// DeserializeHealth deserializes a health component.
func (s *ComponentSerializer) DeserializeHealth(data []byte) (current, max float64, err error) {
	if len(data) != 16 {
		logrus.WithFields(logrus.Fields{
			"component_type": "health",
			"data_length":    len(data),
			"expected":       16,
		}).Warn("invalid health data length")
		return 0, 0, fmt.Errorf("invalid health data length: %d (expected 16)", len(data))
	}
	current = math.Float64frombits(binary.LittleEndian.Uint64(data[0:8]))
	max = math.Float64frombits(binary.LittleEndian.Uint64(data[8:16]))
	return current, max, nil
}

// SerializeStats serializes basic stats (Attack, Defense, MagicPower as float64).
func (s *ComponentSerializer) SerializeStats(attack, defense, magicPower float64) []byte {
	buf := make([]byte, 24)
	binary.LittleEndian.PutUint64(buf[0:8], math.Float64bits(attack))
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(defense))
	binary.LittleEndian.PutUint64(buf[16:24], math.Float64bits(magicPower))
	return buf
}

// DeserializeStats deserializes basic stats.
func (s *ComponentSerializer) DeserializeStats(data []byte) (attack, defense, magicPower float64, err error) {
	if len(data) != 24 {
		logrus.WithFields(logrus.Fields{
			"component_type": "stats",
			"data_length":    len(data),
			"expected":       24,
		}).Warn("invalid stats data length")
		return 0, 0, 0, fmt.Errorf("invalid stats data length: %d (expected 24)", len(data))
	}
	attack = math.Float64frombits(binary.LittleEndian.Uint64(data[0:8]))
	defense = math.Float64frombits(binary.LittleEndian.Uint64(data[8:16]))
	magicPower = math.Float64frombits(binary.LittleEndian.Uint64(data[16:24]))
	return attack, defense, magicPower, nil
}

// SerializeTeam serializes a team component (TeamID as uint64).
func (s *ComponentSerializer) SerializeTeam(teamID uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, teamID)
	return buf
}

// DeserializeTeam deserializes a team component.
func (s *ComponentSerializer) DeserializeTeam(data []byte) (teamID uint64, err error) {
	if len(data) != 8 {
		return 0, fmt.Errorf("invalid team data length: %d (expected 8)", len(data))
	}
	teamID = binary.LittleEndian.Uint64(data)
	return teamID, nil
}

// SerializeLevel serializes a level component (Level, XP as uint32).
func (s *ComponentSerializer) SerializeLevel(level, xp uint32) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], level)
	binary.LittleEndian.PutUint32(buf[4:8], xp)
	return buf
}

// DeserializeLevel deserializes a level component.
func (s *ComponentSerializer) DeserializeLevel(data []byte) (level, xp uint32, err error) {
	if len(data) != 8 {
		return 0, 0, fmt.Errorf("invalid level data length: %d (expected 8)", len(data))
	}
	level = binary.LittleEndian.Uint32(data[0:4])
	xp = binary.LittleEndian.Uint32(data[4:8])
	return level, xp, nil
}

// SerializeInput serializes movement input (DX, DY as int8).
func (s *ComponentSerializer) SerializeInput(dx, dy int8) []byte {
	return []byte{byte(dx), byte(dy)}
}

// DeserializeInput deserializes movement input.
func (s *ComponentSerializer) DeserializeInput(data []byte) (dx, dy int8, err error) {
	if len(data) != 2 {
		return 0, 0, fmt.Errorf("invalid input data length: %d (expected 2)", len(data))
	}
	dx = int8(data[0])
	dy = int8(data[1])
	return dx, dy, nil
}

// SerializeAttack serializes attack command (TargetID as uint64).
func (s *ComponentSerializer) SerializeAttack(targetID uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, targetID)
	return buf
}

// DeserializeAttack deserializes attack command.
func (s *ComponentSerializer) DeserializeAttack(data []byte) (targetID uint64, err error) {
	if len(data) != 8 {
		return 0, fmt.Errorf("invalid attack data length: %d (expected 8)", len(data))
	}
	targetID = binary.LittleEndian.Uint64(data)
	return targetID, nil
}

// SerializeItem serializes item usage (ItemID as uint64).
func (s *ComponentSerializer) SerializeItem(itemID uint64) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, itemID)
	return buf
}

// DeserializeItem deserializes item usage.
func (s *ComponentSerializer) DeserializeItem(data []byte) (itemID uint64, err error) {
	if len(data) != 8 {
		return 0, fmt.Errorf("invalid item data length: %d (expected 8)", len(data))
	}
	itemID = binary.LittleEndian.Uint64(data)
	return itemID, nil
}

// SerializeExpression serializes an expression component.
// Format: 1 byte for ExpressionType, 8 bytes for ExpressionTime, 8 bytes for Cooldown.
// Total: 17 bytes (well under 50 byte budget).
func (s *ComponentSerializer) SerializeExpression(expressionType uint8, expressionTime, cooldown float64) []byte {
	buf := make([]byte, 17)
	buf[0] = expressionType
	binary.LittleEndian.PutUint64(buf[1:9], math.Float64bits(expressionTime))
	binary.LittleEndian.PutUint64(buf[9:17], math.Float64bits(cooldown))
	return buf
}

// DeserializeExpression deserializes an expression component.
func (s *ComponentSerializer) DeserializeExpression(data []byte) (expressionType uint8, expressionTime, cooldown float64, err error) {
	if len(data) != 17 {
		return 0, 0, 0, fmt.Errorf("invalid expression data length: %d (expected 17)", len(data))
	}
	expressionType = data[0]
	expressionTime = math.Float64frombits(binary.LittleEndian.Uint64(data[1:9]))
	cooldown = math.Float64frombits(binary.LittleEndian.Uint64(data[9:17]))
	return expressionType, expressionTime, cooldown, nil
}

// SerializeVehicle serializes a vehicle component (V4.0).
// Format: 1 byte VehicleType, 8 bytes Speed, 8 bytes MaxSpeed, 8 bytes Durability, 8 bytes FuelAmount, 1 byte Occupied.
// Total: 34 bytes.
func (s *ComponentSerializer) SerializeVehicle(vehicleType uint8, speed, maxSpeed, durability, fuelAmount float64, occupied bool) []byte {
	buf := make([]byte, 34)
	buf[0] = vehicleType
	binary.LittleEndian.PutUint64(buf[1:9], math.Float64bits(speed))
	binary.LittleEndian.PutUint64(buf[9:17], math.Float64bits(maxSpeed))
	binary.LittleEndian.PutUint64(buf[17:25], math.Float64bits(durability))
	binary.LittleEndian.PutUint64(buf[25:33], math.Float64bits(fuelAmount))
	if occupied {
		buf[33] = 1
	} else {
		buf[33] = 0
	}
	return buf
}

// DeserializeVehicle deserializes a vehicle component (V4.0).
func (s *ComponentSerializer) DeserializeVehicle(data []byte) (vehicleType uint8, speed, maxSpeed, durability, fuelAmount float64, occupied bool, err error) {
	if len(data) != 34 {
		return 0, 0, 0, 0, 0, false, fmt.Errorf("invalid vehicle data length: %d (expected 34)", len(data))
	}
	vehicleType = data[0]
	speed = math.Float64frombits(binary.LittleEndian.Uint64(data[1:9]))
	maxSpeed = math.Float64frombits(binary.LittleEndian.Uint64(data[9:17]))
	durability = math.Float64frombits(binary.LittleEndian.Uint64(data[17:25]))
	fuelAmount = math.Float64frombits(binary.LittleEndian.Uint64(data[25:33]))
	occupied = data[33] == 1
	return vehicleType, speed, maxSpeed, durability, fuelAmount, occupied, nil
}

// SerializeCompanion serializes a companion component (V4.0).
// Format: 8 bytes OwnerID, 1 byte CompanionType, 8 bytes Loyalty, 4 bytes Level, 1 byte Behavior.
// Total: 22 bytes.
func (s *ComponentSerializer) SerializeCompanion(ownerID uint64, companionType uint8, loyalty float64, level uint32, behavior uint8) []byte {
	buf := make([]byte, 22)
	binary.LittleEndian.PutUint64(buf[0:8], ownerID)
	buf[8] = companionType
	binary.LittleEndian.PutUint64(buf[9:17], math.Float64bits(loyalty))
	binary.LittleEndian.PutUint32(buf[17:21], level)
	buf[21] = behavior
	return buf
}

// DeserializeCompanion deserializes a companion component (V4.0).
func (s *ComponentSerializer) DeserializeCompanion(data []byte) (ownerID uint64, companionType uint8, loyalty float64, level uint32, behavior uint8, err error) {
	if len(data) != 22 {
		return 0, 0, 0, 0, 0, fmt.Errorf("invalid companion data length: %d (expected 22)", len(data))
	}
	ownerID = binary.LittleEndian.Uint64(data[0:8])
	companionType = data[8]
	loyalty = math.Float64frombits(binary.LittleEndian.Uint64(data[9:17]))
	level = binary.LittleEndian.Uint32(data[17:21])
	behavior = data[21]
	return ownerID, companionType, loyalty, level, behavior, nil
}

// SerializeMount serializes a mount component (V4.0).
// Format: 8 bytes MountedEntityID, 8 bytes MountTime, 8 bytes Stamina.
// Total: 24 bytes.
func (s *ComponentSerializer) SerializeMount(mountedEntityID uint64, mountTime, stamina float64) []byte {
	buf := make([]byte, 24)
	binary.LittleEndian.PutUint64(buf[0:8], mountedEntityID)
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(mountTime))
	binary.LittleEndian.PutUint64(buf[16:24], math.Float64bits(stamina))
	return buf
}

// DeserializeMount deserializes a mount component (V4.0).
func (s *ComponentSerializer) DeserializeMount(data []byte) (mountedEntityID uint64, mountTime, stamina float64, err error) {
	if len(data) != 24 {
		return 0, 0, 0, fmt.Errorf("invalid mount data length: %d (expected 24)", len(data))
	}
	mountedEntityID = binary.LittleEndian.Uint64(data[0:8])
	mountTime = math.Float64frombits(binary.LittleEndian.Uint64(data[8:16]))
	stamina = math.Float64frombits(binary.LittleEndian.Uint64(data[16:24]))
	return mountedEntityID, mountTime, stamina, nil
}

// SerializeMiniGame serializes a mini-game component (V4.0).
// Format: 1 byte GameType, 1 byte State, 4 bytes Score, 4 bytes HighScore, 8 bytes TimeRemaining.
// Total: 18 bytes.
func (s *ComponentSerializer) SerializeMiniGame(gameType, state uint8, score, highScore uint32, timeRemaining float64) []byte {
	buf := make([]byte, 18)
	buf[0] = gameType
	buf[1] = state
	binary.LittleEndian.PutUint32(buf[2:6], score)
	binary.LittleEndian.PutUint32(buf[6:10], highScore)
	binary.LittleEndian.PutUint64(buf[10:18], math.Float64bits(timeRemaining))
	return buf
}

// DeserializeMiniGame deserializes a mini-game component (V4.0).
func (s *ComponentSerializer) DeserializeMiniGame(data []byte) (gameType, state uint8, score, highScore uint32, timeRemaining float64, err error) {
	if len(data) != 18 {
		return 0, 0, 0, 0, 0, fmt.Errorf("invalid mini-game data length: %d (expected 18)", len(data))
	}
	gameType = data[0]
	state = data[1]
	score = binary.LittleEndian.Uint32(data[2:6])
	highScore = binary.LittleEndian.Uint32(data[6:10])
	timeRemaining = math.Float64frombits(binary.LittleEndian.Uint64(data[10:18]))
	return gameType, state, score, highScore, timeRemaining, nil
}

// SerializeAchievement serializes an achievement component (V4.0).
// Format: 4 bytes UnlockedCount, 4 bytes TotalPoints.
// Total: 8 bytes.
func (s *ComponentSerializer) SerializeAchievement(unlockedCount, totalPoints uint32) []byte {
	buf := make([]byte, 8)
	binary.LittleEndian.PutUint32(buf[0:4], unlockedCount)
	binary.LittleEndian.PutUint32(buf[4:8], totalPoints)
	return buf
}

// DeserializeAchievement deserializes an achievement component (V4.0).
func (s *ComponentSerializer) DeserializeAchievement(data []byte) (unlockedCount, totalPoints uint32, err error) {
	if len(data) != 8 {
		return 0, 0, fmt.Errorf("invalid achievement data length: %d (expected 8)", len(data))
	}
	unlockedCount = binary.LittleEndian.Uint32(data[0:4])
	totalPoints = binary.LittleEndian.Uint32(data[4:8])
	return unlockedCount, totalPoints, nil
}

// TerritoryData represents serializable territory data for network replication.
type TerritoryData struct {
	ID              string
	ChunkX          int
	ChunkZ          int
	OwnerGuildID    string
	Status          int // TerritoryStatus as int
	CaptureProgress float64
	CapturingGuild  string
	LastUpdateUnix  int64 // Unix timestamp
	ResourceBonus   float64
	XPBonus         float64
}

// SerializeTerritory serializes territory state using gob encoding for complex data.
// Returns serialized bytes or error. Used for territory ownership sync.
func (s *ComponentSerializer) SerializeTerritory(data TerritoryData) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "territory",
			"territory_id":   data.ID,
			"error":          err.Error(),
		}).Error("failed to serialize territory")
		return nil, fmt.Errorf("serialize territory: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"component_type":   "territory",
		"territory_id":     data.ID,
		"owner_guild_id":   data.OwnerGuildID,
		"capture_progress": data.CaptureProgress,
		"size_bytes":       buf.Len(),
	}).Debug("serialized territory")
	return buf.Bytes(), nil
}

// DeserializeTerritory deserializes territory state from gob-encoded bytes.
func (s *ComponentSerializer) DeserializeTerritory(data []byte) (TerritoryData, error) {
	var territory TerritoryData
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&territory); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "territory",
			"data_length":    len(data),
			"error":          err.Error(),
		}).Error("failed to deserialize territory")
		return TerritoryData{}, fmt.Errorf("deserialize territory: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"component_type":   "territory",
		"territory_id":     territory.ID,
		"owner_guild_id":   territory.OwnerGuildID,
		"capture_progress": territory.CaptureProgress,
	}).Debug("deserialized territory")
	return territory, nil
}

// SiegeData represents serializable siege data for network replication.
type SiegeData struct {
	ID                    string
	TerritoryID           string
	AttackerGuildID       string
	DefenderGuildID       string
	Phase                 int   // SiegePhase as int
	StartTimeUnix         int64 // Unix timestamp
	PhaseStartTimeUnix    int64 // Unix timestamp
	EndTimeUnix           int64 // Unix timestamp
	VictoryCondition      int   // VictoryCondition as int
	WinnerGuildID         string
	Attackers             []string // Player IDs
	Defenders             []string // Player IDs
	ControlPointsCaptured int
	TotalControlPoints    int
	GuildHallHP           float64
	GuildHallMaxHP        float64
	DefenderTreasury      int
	LootPercentage        float64
	LootDistributed       bool
}

// SerializeSiege serializes siege state using gob encoding for complex data.
// Returns serialized bytes or error. Used for siege phase and participant sync.
func (s *ComponentSerializer) SerializeSiege(data SiegeData) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "siege",
			"siege_id":       data.ID,
			"error":          err.Error(),
		}).Error("failed to serialize siege")
		return nil, fmt.Errorf("serialize siege: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"component_type":          "siege",
		"siege_id":                data.ID,
		"territory_id":            data.TerritoryID,
		"phase":                   data.Phase,
		"control_points_captured": data.ControlPointsCaptured,
		"attackers_count":         len(data.Attackers),
		"defenders_count":         len(data.Defenders),
		"size_bytes":              buf.Len(),
	}).Debug("serialized siege")
	return buf.Bytes(), nil
}

// DeserializeSiege deserializes siege state from gob-encoded bytes.
func (s *ComponentSerializer) DeserializeSiege(data []byte) (SiegeData, error) {
	var siege SiegeData
	buf := bytes.NewReader(data)
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&siege); err != nil {
		logrus.WithFields(logrus.Fields{
			"component_type": "siege",
			"data_length":    len(data),
			"error":          err.Error(),
		}).Error("failed to deserialize siege")
		return SiegeData{}, fmt.Errorf("deserialize siege: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"component_type":          "siege",
		"siege_id":                siege.ID,
		"territory_id":            siege.TerritoryID,
		"phase":                   siege.Phase,
		"control_points_captured": siege.ControlPointsCaptured,
		"attackers_count":         len(siege.Attackers),
		"defenders_count":         len(siege.Defenders),
	}).Debug("deserialized siege")
	return siege, nil
}
