package prestige

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"sync"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// Sentinel errors for prestige operations.
var (
	ErrPlayerNotFound         = errors.New("player not found")
	ErrNoParagonPoints        = errors.New("no paragon points available")
	ErrInvalidStat            = errors.New("invalid stat")
	ErrAccountNotFound        = errors.New("account not found")
	ErrUnknownParagonCategory = errors.New("unknown paragon category")
)

// Manager handles prestige system operations.
// Uses GameClock for LastUpdated metadata timestamps on PlayerPrestige and
// AccountPrestige records to maintain determinism. Timestamps are for
// audit/debugging purposes only and do not affect gameplay logic.
type Manager struct {
	mu       sync.RWMutex
	players  map[string]*PlayerPrestige
	accounts map[string]*AccountPrestige
	// classToAccount maps playerID to accountID for lookups
	classToAccount map[string]string
	logger         *logrus.Entry
	clock          engine.GameClock
}

// NewManager creates a new prestige manager with a RealTimeClock.
func NewManager() *Manager {
	return NewManagerWithLogger(nil)
}

// NewManagerWithLogger creates a new prestige manager with a logger and RealTimeClock.
func NewManagerWithLogger(logger *logrus.Logger) *Manager {
	return NewManagerWithClock(logger, engine.NewRealTimeClock())
}

// NewManagerWithClock creates a new prestige manager with a logger and custom clock.
// This constructor enables deterministic testing with SimulationClock.
func NewManagerWithClock(logger *logrus.Logger, clock engine.GameClock) *Manager {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"component": "prestige_manager",
		})
	} else {
		logEntry = logrus.WithFields(logrus.Fields{
			"component": "prestige_manager",
		})
	}

	logEntry.Debug("prestige manager created")

	return &Manager{
		players:        make(map[string]*PlayerPrestige),
		accounts:       make(map[string]*AccountPrestige),
		classToAccount: make(map[string]string),
		logger:         logEntry,
		clock:          clock,
	}
}

// CreatePlayer initializes prestige for a new player.
func (m *Manager) CreatePlayer(playerID, className, accountID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.WithFields(logrus.Fields{
		"playerID":  playerID,
		"className": className,
		"accountID": accountID,
	}).Debug("creating player prestige")

	m.players[playerID] = &PlayerPrestige{
		PlayerID:           playerID,
		ClassName:          className,
		PrestigeLevel:      0,
		CurrentXP:          0,
		TotalXP:            0,
		ParagonPoints:      0,
		ParagonAllocations: make(map[ParagonStat]int),
		UnlockedAbilities:  []int{},
		LastUpdated:        m.clock.Now(),
	}

	m.classToAccount[playerID] = accountID

	// Ensure account exists
	if _, exists := m.accounts[accountID]; !exists {
		m.accounts[accountID] = &AccountPrestige{
			AccountID:        accountID,
			Prestige100Count: 0,
			XPBonus:          0.0,
			CharacterIDs:     []string{},
			LastUpdated:      m.clock.Now(),
		}
	}

	// Add character to account
	account := m.accounts[accountID]
	found := false
	for _, cid := range account.CharacterIDs {
		if cid == playerID {
			found = true
			break
		}
	}
	if !found {
		account.CharacterIDs = append(account.CharacterIDs, playerID)
	}
}

// AddPrestigeXP adds XP and returns levels gained.
func (m *Manager) AddPrestigeXP(playerID, className string, xp int) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.players[playerID]
	if !exists {
		return 0
	}

	player.CurrentXP += xp
	player.TotalXP += xp
	player.LastUpdated = m.clock.Now()

	levelsGained := 0
	for {
		required := m.calculateXPRequired(player.PrestigeLevel + 1)
		if player.CurrentXP < required {
			break
		}

		player.CurrentXP -= required
		player.PrestigeLevel++
		levelsGained++

		// Check for prestige 100 milestone for account bonus
		if player.PrestigeLevel == 100 {
			m.updateAccountBonus(playerID, 1)
		}
	}

	return levelsGained
}

// AddParagonPoints adds paragon points to a player.
func (m *Manager) AddParagonPoints(playerID string, points int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.players[playerID]
	if !exists {
		return
	}

	player.ParagonPoints += points
	player.LastUpdated = m.clock.Now()
}

// AllocateParagonPoint allocates a paragon point to a stat.
func (m *Manager) AllocateParagonPoint(playerID string, stat ParagonStat) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.players[playerID]
	if !exists {
		return fmt.Errorf("%w: %s", ErrPlayerNotFound, playerID)
	}

	if player.ParagonPoints <= 0 {
		return fmt.Errorf("%w: player %s", ErrNoParagonPoints, playerID)
	}

	if stat < StatHealth || stat > StatCritical {
		return fmt.Errorf("%w: %d for player %s", ErrInvalidStat, stat, playerID)
	}

	player.ParagonPoints--
	player.ParagonAllocations[stat]++
	player.LastUpdated = m.clock.Now()

	return nil
}

// RespecParagonPoints resets all allocations and returns points (costs gold).
func (m *Manager) RespecParagonPoints(playerID string) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.players[playerID]
	if !exists {
		return 0, fmt.Errorf("%w: %s", ErrPlayerNotFound, playerID)
	}

	totalPoints := 0
	for _, points := range player.ParagonAllocations {
		totalPoints += points
	}

	if totalPoints == 0 {
		return 0, fmt.Errorf("no points allocated to respec")
	}

	// Clear allocations
	player.ParagonAllocations = make(map[ParagonStat]int)
	player.ParagonPoints += totalPoints
	player.LastUpdated = m.clock.Now()

	cost := totalPoints * RespecCostPerPoint
	return cost, nil
}

// GetStatBonus calculates the multiplicative bonus for a stat.
func (m *Manager) GetStatBonus(playerID string, stat ParagonStat) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	player, exists := m.players[playerID]
	if !exists {
		return 0.0
	}

	points := player.ParagonAllocations[stat]
	return float64(points) * ParagonPointBonus
}

// GetPrestigeLevel returns the player's current prestige level.
func (m *Manager) GetPrestigeLevel(playerID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	player, exists := m.players[playerID]
	if !exists {
		return 0
	}

	return player.PrestigeLevel
}

// GetVisualTier returns the visual effect tier for a prestige level.
func (m *Manager) GetVisualTier(prestigeLevel int) VisualTier {
	if prestigeLevel < 10 {
		return VisualNone
	} else if prestigeLevel < 25 {
		return VisualSubtle
	} else if prestigeLevel < 50 {
		return VisualModerate
	} else if prestigeLevel < 100 {
		return VisualIntense
	}
	return VisualRadiant
}

// GetPrestigeAbility returns the prestige ability for a class at a level.
func (m *Manager) GetPrestigeAbility(className string, level int) *PrestigeAbility {
	// Generate deterministic ability based on class and level
	abilities := m.generateAbilitiesForClass(className)

	var ability *PrestigeAbility
	switch level {
	case 10:
		ability = &abilities[0]
	case 25:
		ability = &abilities[1]
	case 50:
		ability = &abilities[2]
	case 100:
		ability = &abilities[3]
	default:
		return nil
	}

	return ability
}

// CheckAbilityUnlock checks if player unlocked an ability at current prestige level.
func (m *Manager) CheckAbilityUnlock(playerID string) *PrestigeAbility {
	m.mu.Lock()
	defer m.mu.Unlock()

	player, exists := m.players[playerID]
	if !exists {
		return nil
	}

	level := player.PrestigeLevel
	milestones := []int{10, 25, 50, 100}

	for _, milestone := range milestones {
		if level == milestone {
			// Check if already unlocked
			alreadyUnlocked := false
			for _, unlocked := range player.UnlockedAbilities {
				if unlocked == milestone {
					alreadyUnlocked = true
					break
				}
			}

			if !alreadyUnlocked {
				player.UnlockedAbilities = append(player.UnlockedAbilities, milestone)
				player.LastUpdated = m.clock.Now()
				return m.GetPrestigeAbility(player.ClassName, milestone)
			}
		}
	}

	return nil
}

// GetAccountXPBonus returns the account-wide XP bonus.
func (m *Manager) GetAccountXPBonus(accountID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	account, exists := m.accounts[accountID]
	if !exists {
		return 0.0
	}

	return account.XPBonus
}

// updateAccountBonus updates account bonus when character hits prestige 100.
func (m *Manager) updateAccountBonus(playerID string, delta int) {
	accountID, exists := m.classToAccount[playerID]
	if !exists {
		return
	}

	account, exists := m.accounts[accountID]
	if !exists {
		return
	}

	account.Prestige100Count += delta
	// XP bonus stacks multiplicatively: 1 char = 5%, 2 chars = 10.25%, etc.
	account.XPBonus = math.Pow(1.0+AccountXPBonus, float64(account.Prestige100Count)) - 1.0
	account.LastUpdated = m.clock.Now()
}

// calculateXPRequired calculates XP needed for a prestige level.
func (m *Manager) calculateXPRequired(level int) int {
	if level <= 0 {
		return BasePrestigeXP
	}
	// Exponential curve: BaseXP * (2 ^ (level-1))
	// Level 1 requires BaseXP, Level 2 requires BaseXP*2, Level 3 requires BaseXP*4, etc.
	return int(float64(BasePrestigeXP) * math.Pow(2, float64(level-1)))
}

// generateAbilitiesForClass generates 4 prestige abilities for a class.
func (m *Manager) generateAbilitiesForClass(className string) [4]PrestigeAbility {
	// Simplified deterministic generation based on class name
	return [4]PrestigeAbility{
		{
			Name:        className + "'s Resolve",
			Description: "Reduces damage taken by 25% for 10 seconds",
			UnlockLevel: 10,
			ClassName:   className,
			Cooldown:    120,
			ManaCost:    50,
		},
		{
			Name:        "Legendary " + className + " Strike",
			Description: "Deals 500% weapon damage to target",
			UnlockLevel: 25,
			ClassName:   className,
			Cooldown:    60,
			ManaCost:    75,
		},
		{
			Name:        "Ascended " + className + " Power",
			Description: "Increases all stats by 50% for 20 seconds",
			UnlockLevel: 50,
			ClassName:   className,
			Cooldown:    180,
			ManaCost:    100,
		},
		{
			Name:        "Transcendent " + className + " Form",
			Description: "Become invulnerable for 5 seconds, dealing 1000% damage",
			UnlockLevel: 100,
			ClassName:   className,
			Cooldown:    300,
			ManaCost:    150,
		},
	}
}

// Save persists prestige data to JSON with gzip compression.
func (m *Manager) Save() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := PrestigeData{
		Players:  m.players,
		Accounts: m.accounts,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal prestige data: %w", err)
	}

	var buf bytes.Buffer
	gzipWriter := gzip.NewWriter(&buf)
	if _, err := gzipWriter.Write(jsonData); err != nil {
		return nil, fmt.Errorf("failed to compress prestige data: %w", err)
	}
	if err := gzipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

// Load restores prestige data from JSON with gzip compression.
func (m *Manager) Load(compressedData []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	gzipReader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	jsonData, err := io.ReadAll(gzipReader)
	if err != nil {
		return fmt.Errorf("failed to decompress prestige data: %w", err)
	}

	var data PrestigeData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal prestige data: %w", err)
	}

	m.players = data.Players
	m.accounts = data.Accounts

	// Rebuild classToAccount map
	m.classToAccount = make(map[string]string)
	for accountID, account := range m.accounts {
		for _, playerID := range account.CharacterIDs {
			m.classToAccount[playerID] = accountID
		}
	}

	return nil
}

// GetPlayerPrestige returns a copy of the player's prestige data for UI display.
// Returns nil if player not found.
func (m *Manager) GetPlayerPrestige(playerID string) *PlayerPrestige {
	m.mu.RLock()
	defer m.mu.RUnlock()

	player, exists := m.players[playerID]
	if !exists {
		return nil
	}

	// Return a copy to avoid race conditions
	copy := *player
	copy.ParagonAllocations = make(map[ParagonStat]int)
	for k, v := range player.ParagonAllocations {
		copy.ParagonAllocations[k] = v
	}
	copy.UnlockedAbilities = make([]int, len(player.UnlockedAbilities))
	for i := range player.UnlockedAbilities {
		copy.UnlockedAbilities[i] = player.UnlockedAbilities[i]
	}

	return &copy
}

// GetXPProgress returns current XP and XP required for next level.
func (m *Manager) GetXPProgress(playerID string) (currentXP, requiredXP int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	player, exists := m.players[playerID]
	if !exists {
		return 0, m.calculateXPRequired(1)
	}

	return player.CurrentXP, m.calculateXPRequired(player.PrestigeLevel + 1)
}

// GetTotalAllocatedPoints returns the total paragon points allocated.
func (m *Manager) GetTotalAllocatedPoints(playerID string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	player, exists := m.players[playerID]
	if !exists {
		return 0
	}

	total := 0
	for _, pts := range player.ParagonAllocations {
		total += pts
	}
	return total
}
