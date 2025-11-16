// Package world provides server ranking and meta-game systems.
package world

import (
	"sort"
	"sync"
)

// ServerRank represents a server's ranking data.
type ServerRank struct {
	ServerID            string
	Population          int
	EconomicPower       int64
	MilitaryStrength    int
	DiplomaticInfluence int
	TotalScore          float64
}

// LeaderboardType defines ranking categories.
type LeaderboardType int

const (
	LeaderboardPopulation LeaderboardType = iota
	LeaderboardEconomic
	LeaderboardMilitary
	LeaderboardDiplomatic
	LeaderboardOverall
)

// String returns human-readable leaderboard type.
func (l LeaderboardType) String() string {
	switch l {
	case LeaderboardPopulation:
		return "Population"
	case LeaderboardEconomic:
		return "Economic Power"
	case LeaderboardMilitary:
		return "Military Strength"
	case LeaderboardDiplomatic:
		return "Diplomatic Influence"
	case LeaderboardOverall:
		return "Overall Score"
	default:
		return "Unknown"
	}
}

// RankingManager manages server rankings and leaderboards.
type RankingManager struct {
	mu      sync.RWMutex
	servers map[string]*ServerRank
}

// NewRankingManager creates a new ranking manager.
func NewRankingManager() *RankingManager {
	return &RankingManager{
		servers: make(map[string]*ServerRank),
	}
}

// RegisterServer adds or updates a server's ranking data.
func (rm *RankingManager) RegisterServer(serverID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.servers[serverID]; !exists {
		rm.servers[serverID] = &ServerRank{
			ServerID:            serverID,
			Population:          0,
			EconomicPower:       0,
			MilitaryStrength:    0,
			DiplomaticInfluence: 0,
			TotalScore:          0.0,
		}
	}
}

// UpdatePopulation updates a server's active player count.
func (rm *RankingManager) UpdatePopulation(serverID string, population int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rank, exists := rm.servers[serverID]; exists {
		rank.Population = population
		rm.recalculateScore(rank)
	}
}

// UpdateEconomicPower updates a server's total trade volume.
func (rm *RankingManager) UpdateEconomicPower(serverID string, power int64) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rank, exists := rm.servers[serverID]; exists {
		rank.EconomicPower = power
		rm.recalculateScore(rank)
	}
}

// UpdateMilitaryStrength updates a server's controlled territory count.
func (rm *RankingManager) UpdateMilitaryStrength(serverID string, strength int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rank, exists := rm.servers[serverID]; exists {
		rank.MilitaryStrength = strength
		rm.recalculateScore(rank)
	}
}

// UpdateDiplomaticInfluence updates a server's alliance count.
func (rm *RankingManager) UpdateDiplomaticInfluence(serverID string, influence int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rank, exists := rm.servers[serverID]; exists {
		rank.DiplomaticInfluence = influence
		rm.recalculateScore(rank)
	}
}

// recalculateScore computes the overall score for a server.
func (rm *RankingManager) recalculateScore(rank *ServerRank) {
	popScore := float64(rank.Population) * 1.0
	economicScore := float64(rank.EconomicPower) * 0.0001
	militaryScore := float64(rank.MilitaryStrength) * 10.0
	diplomaticScore := float64(rank.DiplomaticInfluence) * 5.0

	rank.TotalScore = popScore + economicScore + militaryScore + diplomaticScore
}

// GetLeaderboard returns ranked servers for a specific category.
func (rm *RankingManager) GetLeaderboard(leaderboardType LeaderboardType, limit int) []*ServerRank {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	ranks := make([]*ServerRank, 0, len(rm.servers))
	for _, rank := range rm.servers {
		rankCopy := *rank
		ranks = append(ranks, &rankCopy)
	}

	switch leaderboardType {
	case LeaderboardPopulation:
		sort.Slice(ranks, func(i, j int) bool {
			return ranks[i].Population > ranks[j].Population
		})
	case LeaderboardEconomic:
		sort.Slice(ranks, func(i, j int) bool {
			return ranks[i].EconomicPower > ranks[j].EconomicPower
		})
	case LeaderboardMilitary:
		sort.Slice(ranks, func(i, j int) bool {
			return ranks[i].MilitaryStrength > ranks[j].MilitaryStrength
		})
	case LeaderboardDiplomatic:
		sort.Slice(ranks, func(i, j int) bool {
			return ranks[i].DiplomaticInfluence > ranks[j].DiplomaticInfluence
		})
	case LeaderboardOverall:
		sort.Slice(ranks, func(i, j int) bool {
			return ranks[i].TotalScore > ranks[j].TotalScore
		})
	}

	if limit > 0 && limit < len(ranks) {
		return ranks[:limit]
	}
	return ranks
}

// GetServerRank returns ranking data for a specific server.
func (rm *RankingManager) GetServerRank(serverID string) (*ServerRank, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	rank, exists := rm.servers[serverID]
	if !exists {
		return nil, false
	}

	rankCopy := *rank
	return &rankCopy, true
}

// GetServerCount returns the number of registered servers.
func (rm *RankingManager) GetServerCount() int {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	return len(rm.servers)
}

// GetTopServer returns the highest-ranked server for a category.
func (rm *RankingManager) GetTopServer(leaderboardType LeaderboardType) (*ServerRank, bool) {
	leaderboard := rm.GetLeaderboard(leaderboardType, 1)
	if len(leaderboard) == 0 {
		return nil, false
	}
	return leaderboard[0], true
}

// GetServerPosition returns a server's position in a leaderboard (1-indexed).
func (rm *RankingManager) GetServerPosition(serverID string, leaderboardType LeaderboardType) int {
	leaderboard := rm.GetLeaderboard(leaderboardType, 0)
	for i, rank := range leaderboard {
		if rank.ServerID == serverID {
			return i + 1
		}
	}
	return -1
}

// RemoveServer removes a server from rankings.
func (rm *RankingManager) RemoveServer(serverID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	delete(rm.servers, serverID)
}
