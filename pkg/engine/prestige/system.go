package prestige

import (
	"github.com/sirupsen/logrus"
)

// Entity represents a minimal entity interface for the prestige system.
// This avoids circular dependencies with pkg/engine while allowing
// the system to work with any entity implementation that satisfies this interface.
type Entity interface {
	GetID() string
	HasComponent(componentType string) bool
	GetComponent(componentType string) interface{}
	// AddComponent accepts any component with a Type() string method
	// This matches pkg/engine.Entity's signature
	AddComponent(component interface{ Type() string })
	RemoveComponent(componentType string)
}

// System wraps the prestige Manager for ECS integration.
// It handles prestige progression, paragon points, and prestige abilities
// for entities that reach max level (level 50).
type System struct {
	manager *Manager
	logger  *logrus.Entry
}

// NewSystem creates a new prestige system.
func NewSystem() *System {
	return NewSystemWithLogger(nil)
}

// NewSystemWithLogger creates a new prestige system with a logger.
func NewSystemWithLogger(logger *logrus.Logger) *System {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "prestige",
		})
	} else {
		logEntry = logrus.WithFields(logrus.Fields{
			"system": "prestige",
		})
	}

	logEntry.Debug("prestige system created")

	return &System{
		manager: NewManagerWithLogger(logger),
		logger:  logEntry,
	}
}

// Update processes prestige system logic for entities.
// It checks for prestige ability unlocks and updates visual tiers.
func (s *System) Update(entities []Entity, deltaTime float64) {
	for _, entity := range entities {
		if !entity.HasComponent("prestige") {
			continue
		}

		comp := entity.GetComponent("prestige")
		if comp == nil {
			continue
		}

		prestigeComp, ok := comp.(*PrestigeComponent)
		if !ok {
			continue
		}

		playerID := prestigeComp.PlayerID
		if playerID == "" {
			continue
		}

		// Check for ability unlocks at current prestige level
		ability := s.manager.CheckAbilityUnlock(playerID)
		if ability != nil {
			s.logger.WithFields(logrus.Fields{
				"playerID":    playerID,
				"ability":     ability.Name,
				"unlockLevel": ability.UnlockLevel,
			}).Info("prestige ability unlocked")

			// Add ability to active abilities
			prestigeComp.ActiveAbilities = append(prestigeComp.ActiveAbilities, ability.Name)
		}

		// Update visual tier based on current prestige level
		currentLevel := s.manager.GetPrestigeLevel(playerID)
		prestigeComp.PrestigeLevel = currentLevel
		prestigeComp.VisualTier = s.manager.GetVisualTier(currentLevel)
	}
}

// GetManager returns the underlying prestige manager for direct access.
func (s *System) GetManager() *Manager {
	return s.manager
}

// InitializePlayer creates prestige tracking for a player entity.
func (s *System) InitializePlayer(entity Entity, playerID, className, accountID string) {
	s.manager.CreatePlayer(playerID, className, accountID)

	// Add prestige component to entity
	comp := &PrestigeComponent{
		PlayerID:        playerID,
		PrestigeLevel:   0,
		VisualTier:      VisualNone,
		ActiveAbilities: []string{},
	}
	entity.AddComponent(comp)

	s.logger.WithFields(logrus.Fields{
		"playerID":  playerID,
		"className": className,
		"accountID": accountID,
	}).Debug("initialized prestige for player")
}

// AwardPrestigeXP adds prestige XP to a player and handles level-ups.
// Returns the number of prestige levels gained.
func (s *System) AwardPrestigeXP(playerID, className string, xp int) int {
	levelsGained := s.manager.AddPrestigeXP(playerID, className, xp)

	if levelsGained > 0 {
		// Award paragon points (1 per level)
		s.manager.AddParagonPoints(playerID, levelsGained)

		currentLevel := s.manager.GetPrestigeLevel(playerID)
		s.logger.WithFields(logrus.Fields{
			"playerID":     playerID,
			"levelsGained": levelsGained,
			"newLevel":     currentLevel,
			"xpAwarded":    xp,
		}).Info("prestige level up")
	}

	return levelsGained
}

// AllocateParagonPoint allocates a paragon point to a stat for a player.
func (s *System) AllocateParagonPoint(playerID string, stat ParagonStat) error {
	err := s.manager.AllocateParagonPoint(playerID, stat)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID": playerID,
			"stat":     stat.String(),
			"error":    err.Error(),
		}).Warn("failed to allocate paragon point")
		return err
	}

	s.logger.WithFields(logrus.Fields{
		"playerID": playerID,
		"stat":     stat.String(),
	}).Debug("allocated paragon point")

	return nil
}

// GetStatBonus returns the multiplicative stat bonus from allocated paragon points.
func (s *System) GetStatBonus(playerID string, stat ParagonStat) float64 {
	return s.manager.GetStatBonus(playerID, stat)
}

// GetAccountXPBonus returns the account-wide XP bonus.
func (s *System) GetAccountXPBonus(accountID string) float64 {
	return s.manager.GetAccountXPBonus(accountID)
}

// RespecParagonPoints resets all paragon point allocations.
// Returns the gold cost for the respec.
func (s *System) RespecParagonPoints(playerID string) (int, error) {
	cost, err := s.manager.RespecParagonPoints(playerID)
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"playerID": playerID,
			"error":    err.Error(),
		}).Warn("failed to respec paragon points")
		return 0, err
	}

	s.logger.WithFields(logrus.Fields{
		"playerID": playerID,
		"cost":     cost,
	}).Info("respec paragon points")

	return cost, nil
}

// Save persists the prestige data.
func (s *System) Save() ([]byte, error) {
	return s.manager.Save()
}

// Load restores prestige data.
func (s *System) Load(data []byte) error {
	return s.manager.Load(data)
}
