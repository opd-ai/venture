// Package engine provides bounty contract system for cross-server quests.
package engine

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// ObjectiveType defines bounty objective categories.
type ObjectiveType int

const (
	ObjectiveKill ObjectiveType = iota
	ObjectiveDeliver
	ObjectiveEscort
	ObjectiveExplore
	ObjectiveCraft
)

// String returns human-readable objective type.
func (o ObjectiveType) String() string {
	switch o {
	case ObjectiveKill:
		return "Kill"
	case ObjectiveDeliver:
		return "Deliver"
	case ObjectiveEscort:
		return "Escort"
	case ObjectiveExplore:
		return "Explore"
	case ObjectiveCraft:
		return "Craft"
	default:
		return "Unknown"
	}
}

// BountyContract represents a cross-server quest.
type BountyContract struct {
	ID           string
	IssuerServer string
	TargetServer string
	Objective    ObjectiveType
	Description  string
	Reward       int
	ExpiresAt    int64
	AcceptedBy   string
	CompletedAt  int64
	Difficulty   int
}

// BountyComponent marks an entity as having bounty contracts.
type BountyComponent struct {
	AvailableBounties []*BountyContract
	AcceptedBounties  []*BountyContract
	CompletedBounties []*BountyContract
}

// Type returns the component type identifier.
func (b BountyComponent) Type() string {
	return "bounty"
}

// BountySystem manages bounty contracts and completion tracking.
type BountySystem struct {
	world          *World
	logger         *logrus.Logger
	activeBounties map[string]*BountyContract
	completionRate float64
	nextBountyID   int
}

// NewBountySystem creates a new bounty system.
func NewBountySystem(world *World, logger *logrus.Logger) *BountySystem {
	return &BountySystem{
		world:          world,
		logger:         logger,
		activeBounties: make(map[string]*BountyContract),
		completionRate: 0.0,
		nextBountyID:   1,
	}
}

// Update processes bounty expiration and completion tracking.
func (bs *BountySystem) Update(deltaTime float64) {
	now := time.Now().Unix()

	for id, bounty := range bs.activeBounties {
		if bounty.ExpiresAt < now && bounty.CompletedAt == 0 {
			delete(bs.activeBounties, id)
			if bs.logger != nil {
				bs.logger.WithFields(logrus.Fields{
					"bountyID": id,
					"issuer":   bounty.IssuerServer,
					"target":   bounty.TargetServer,
				}).Info("bounty expired")
			}
		}
	}

	bs.updateCompletionRate()
}

// CreateBounty creates a new bounty contract.
func (bs *BountySystem) CreateBounty(issuer, target string, objective ObjectiveType, description string, reward, difficulty int, duration int64) *BountyContract {
	id := fmt.Sprintf("bounty_%d", bs.nextBountyID)
	bs.nextBountyID++

	bounty := &BountyContract{
		ID:           id,
		IssuerServer: issuer,
		TargetServer: target,
		Objective:    objective,
		Description:  description,
		Reward:       reward,
		ExpiresAt:    time.Now().Unix() + duration,
		AcceptedBy:   "",
		CompletedAt:  0,
		Difficulty:   difficulty,
	}

	bs.activeBounties[id] = bounty

	if bs.logger != nil {
		bs.logger.WithFields(logrus.Fields{
			"bountyID":   id,
			"issuer":     issuer,
			"target":     target,
			"objective":  objective.String(),
			"reward":     reward,
			"difficulty": difficulty,
		}).Info("bounty created")
	}

	return bounty
}

// AcceptBounty assigns a bounty to a player.
func (bs *BountySystem) AcceptBounty(bountyID, playerID string) error {
	bounty, exists := bs.activeBounties[bountyID]
	if !exists {
		return fmt.Errorf("bounty not found: %s", bountyID)
	}

	if bounty.AcceptedBy != "" {
		return fmt.Errorf("bounty already accepted by: %s", bounty.AcceptedBy)
	}

	now := time.Now().Unix()
	if bounty.ExpiresAt < now {
		return fmt.Errorf("bounty expired: %s", bountyID)
	}

	bounty.AcceptedBy = playerID

	if bs.logger != nil {
		bs.logger.WithFields(logrus.Fields{
			"bountyID": bountyID,
			"playerID": playerID,
		}).Info("bounty accepted")
	}

	return nil
}

// CompleteBounty marks a bounty as completed.
func (bs *BountySystem) CompleteBounty(bountyID, playerID string) error {
	bounty, exists := bs.activeBounties[bountyID]
	if !exists {
		return fmt.Errorf("bounty not found: %s", bountyID)
	}

	if bounty.AcceptedBy != playerID {
		return fmt.Errorf("bounty not accepted by player: %s", playerID)
	}

	if bounty.CompletedAt != 0 {
		return fmt.Errorf("bounty already completed: %s", bountyID)
	}

	bounty.CompletedAt = time.Now().Unix()

	if bs.logger != nil {
		bs.logger.WithFields(logrus.Fields{
			"bountyID": bountyID,
			"playerID": playerID,
			"reward":   bounty.Reward,
		}).Info("bounty completed")
	}

	return nil
}

// GetBounty retrieves a bounty by ID.
func (bs *BountySystem) GetBounty(bountyID string) (*BountyContract, error) {
	bounty, exists := bs.activeBounties[bountyID]
	if !exists {
		return nil, fmt.Errorf("bounty not found: %s", bountyID)
	}
	return bounty, nil
}

// GetAvailableBounties returns all non-accepted, non-expired bounties.
func (bs *BountySystem) GetAvailableBounties() []*BountyContract {
	available := make([]*BountyContract, 0)
	now := time.Now().Unix()

	for _, bounty := range bs.activeBounties {
		if bounty.AcceptedBy == "" && bounty.ExpiresAt >= now {
			available = append(available, bounty)
		}
	}

	return available
}

// GetBountiesByServer returns bounties for a specific server.
func (bs *BountySystem) GetBountiesByServer(server string) []*BountyContract {
	bounties := make([]*BountyContract, 0)
	now := time.Now().Unix()

	for _, bounty := range bs.activeBounties {
		if (bounty.IssuerServer == server || bounty.TargetServer == server) && bounty.ExpiresAt >= now {
			bounties = append(bounties, bounty)
		}
	}

	return bounties
}

// GetBountiesByDifficulty returns bounties of a specific difficulty level.
func (bs *BountySystem) GetBountiesByDifficulty(difficulty int) []*BountyContract {
	bounties := make([]*BountyContract, 0)
	now := time.Now().Unix()

	for _, bounty := range bs.activeBounties {
		if bounty.Difficulty == difficulty && bounty.ExpiresAt >= now && bounty.AcceptedBy == "" {
			bounties = append(bounties, bounty)
		}
	}

	return bounties
}

// GetCompletionRate returns the percentage of accepted bounties that are completed.
func (bs *BountySystem) GetCompletionRate() float64 {
	return bs.completionRate
}

// updateCompletionRate recalculates the bounty completion rate.
func (bs *BountySystem) updateCompletionRate() {
	accepted := 0
	completed := 0

	for _, bounty := range bs.activeBounties {
		if bounty.AcceptedBy != "" {
			accepted++
			if bounty.CompletedAt != 0 {
				completed++
			}
		}
	}

	if accepted > 0 {
		bs.completionRate = float64(completed) / float64(accepted)
	} else {
		bs.completionRate = 0.0
	}
}

// GetActiveBountyCount returns the number of active bounties.
func (bs *BountySystem) GetActiveBountyCount() int {
	return len(bs.activeBounties)
}

// CancelBounty removes a bounty from the active list.
func (bs *BountySystem) CancelBounty(bountyID string) error {
	bounty, exists := bs.activeBounties[bountyID]
	if !exists {
		return fmt.Errorf("bounty not found: %s", bountyID)
	}

	delete(bs.activeBounties, bountyID)

	if bs.logger != nil {
		bs.logger.WithFields(logrus.Fields{
			"bountyID": bountyID,
			"issuer":   bounty.IssuerServer,
		}).Info("bounty cancelled")
	}

	return nil
}
