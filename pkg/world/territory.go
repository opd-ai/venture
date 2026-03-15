// Package world provides world state and territory control.
package world

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// BorderZone represents a contested area between two servers.
type BorderZone struct {
	ZoneID           string
	ServerA          string
	ServerB          string
	ControlPoints    []*ControlPoint
	OwnerFaction     string
	ContestedAt      int64
	ResourceBonus    float64
	LastCaptureCheck int64
}

// ControlPoint represents a capturable location within a border zone.
type ControlPoint struct {
	X                float64
	Y                float64
	CaptureProgress  float64
	CapturingFaction string
	LastUpdateTime   int64
	DefenderCount    int
	AttackerCount    int
}

// TerritoryManager manages border zones and control points.
type TerritoryManager struct {
	mu                sync.RWMutex
	zones             map[string]*BorderZone
	captureRadius     float64
	captureTimeBase   int64
	captureTimePerDef int64
	resourceBonus     float64
}

// NewTerritoryManager creates a new territory manager.
func NewTerritoryManager() *TerritoryManager {
	return &TerritoryManager{
		zones:             make(map[string]*BorderZone),
		captureRadius:     50.0,
		captureTimeBase:   60,
		captureTimePerDef: 30,
		resourceBonus:     0.10,
	}
}

// CreateBorderZone creates a new contested zone between two servers.
func (tm *TerritoryManager) CreateBorderZone(zoneID, serverA, serverB string, controlPointCount int) *BorderZone {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	zone := &BorderZone{
		ZoneID:           zoneID,
		ServerA:          serverA,
		ServerB:          serverB,
		ControlPoints:    make([]*ControlPoint, 0, controlPointCount),
		OwnerFaction:     "",
		ContestedAt:      time.Now().Unix(),
		ResourceBonus:    tm.resourceBonus,
		LastCaptureCheck: time.Now().Unix(),
	}

	for i := 0; i < controlPointCount; i++ {
		cp := &ControlPoint{
			X:                float64(i*100 + 100),
			Y:                float64(i*50 + 50),
			CaptureProgress:  0.0,
			CapturingFaction: "",
			LastUpdateTime:   time.Now().Unix(),
			DefenderCount:    0,
			AttackerCount:    0,
		}
		zone.ControlPoints = append(zone.ControlPoints, cp)
	}

	tm.zones[zoneID] = zone
	return zone
}

// getZone retrieves a border zone by ID without acquiring a lock.
// Callers must hold at least tm.mu.RLock() before calling this method.
func (tm *TerritoryManager) getZone(zoneID string) (*BorderZone, error) {
	zone, exists := tm.zones[zoneID]
	if !exists {
		return nil, fmt.Errorf("zone not found: %s", zoneID)
	}
	return zone, nil
}

// GetZone retrieves a border zone by ID.
func (tm *TerritoryManager) GetZone(zoneID string) (*BorderZone, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	return tm.getZone(zoneID)
}

// UpdateControlPoint updates capture progress for a control point.
func (tm *TerritoryManager) UpdateControlPoint(zoneID string, cpIndex, attackers, defenders int, faction string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	zone, err := tm.getZone(zoneID)
	if err != nil {
		return err
	}

	if cpIndex < 0 || cpIndex >= len(zone.ControlPoints) {
		return fmt.Errorf("invalid control point index: %d", cpIndex)
	}

	if attackers < 0 || defenders < 0 {
		return fmt.Errorf("attackers (%d) and defenders (%d) must be non-negative", attackers, defenders)
	}

	cp := zone.ControlPoints[cpIndex]
	now := time.Now().Unix()
	elapsed := now - cp.LastUpdateTime

	cp.AttackerCount = attackers
	cp.DefenderCount = defenders

	if attackers > 0 && defenders == 0 {
		captureTime := tm.captureTimeBase
		progressPerSecond := 1.0 / float64(captureTime)
		cp.CaptureProgress += progressPerSecond * float64(elapsed)
		cp.CapturingFaction = faction

		if cp.CaptureProgress >= 1.0 {
			cp.CaptureProgress = 1.0
			zone.OwnerFaction = faction
			zone.ContestedAt = now
		}
	} else if defenders > 0 {
		decayTime := tm.captureTimeBase + (tm.captureTimePerDef * int64(defenders))
		decayPerSecond := 1.0 / float64(decayTime)
		cp.CaptureProgress -= decayPerSecond * float64(elapsed)

		if cp.CaptureProgress < 0.0 {
			cp.CaptureProgress = 0.0
			cp.CapturingFaction = ""
		}
	}

	cp.LastUpdateTime = now
	zone.LastCaptureCheck = now
	return nil
}

// IsPlayerInCaptureRange checks if a player is within capture range of a control point.
func (tm *TerritoryManager) IsPlayerInCaptureRange(playerX, playerY, cpX, cpY float64) bool {
	dx := playerX - cpX
	dy := playerY - cpY
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance <= tm.captureRadius
}

// GetControlledZones returns zones controlled by a faction.
func (tm *TerritoryManager) GetControlledZones(faction string) []*BorderZone {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	controlled := make([]*BorderZone, 0)
	for _, zone := range tm.zones {
		if zone.OwnerFaction == faction {
			controlled = append(controlled, zone)
		}
	}
	return controlled
}

// GetContestedZones returns zones that are currently being contested.
func (tm *TerritoryManager) GetContestedZones() []*BorderZone {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	contested := make([]*BorderZone, 0)
	for _, zone := range tm.zones {
		for _, cp := range zone.ControlPoints {
			if cp.CaptureProgress > 0.0 && cp.CaptureProgress < 1.0 {
				contested = append(contested, zone)
				break
			}
		}
	}
	return contested
}

// GetResourceBonus returns the resource spawn bonus for a faction in controlled zones.
func (tm *TerritoryManager) GetResourceBonus(faction string) float64 {
	tm.mu.RLock()
	controlledCount := 0
	for _, zone := range tm.zones {
		if zone.OwnerFaction == faction {
			controlledCount++
		}
	}
	tm.mu.RUnlock()
	return float64(controlledCount) * tm.resourceBonus
}

// ResetControlPoint resets a control point to neutral state.
func (tm *TerritoryManager) ResetControlPoint(zoneID string, cpIndex int) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	zone, err := tm.getZone(zoneID)
	if err != nil {
		return err
	}

	if cpIndex < 0 || cpIndex >= len(zone.ControlPoints) {
		return fmt.Errorf("invalid control point index: %d", cpIndex)
	}

	cp := zone.ControlPoints[cpIndex]
	cp.CaptureProgress = 0.0
	cp.CapturingFaction = ""
	cp.LastUpdateTime = time.Now().Unix()
	cp.DefenderCount = 0
	cp.AttackerCount = 0

	return nil
}

// GetAllZones returns all border zones.
func (tm *TerritoryManager) GetAllZones() []*BorderZone {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	zones := make([]*BorderZone, 0, len(tm.zones))
	for _, zone := range tm.zones {
		zones = append(zones, zone)
	}
	return zones
}
