// Package network provides component serialization for networking.
// This file implements serialization and deserialization of ECS components
// for efficient network transmission.
package network

import (
	"encoding/binary"
	"fmt"
	"math"
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
