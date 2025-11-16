package federation

import (
	"sync"
	"time"
)

// ServerInfo represents metadata about a federated server
type ServerInfo struct {
	ServerID    string     // Public key fingerprint
	ServerName  string     // Human-readable name
	Address     string     // IP:port
	Version     string     // Protocol version
	Features    []string   // Supported features
	TrustLevel  TrustLevel // Trust relationship
	LastSeen    time.Time  // Last heartbeat timestamp
	PlayerCount int        // Active players on this server
	IsOnline    bool       // Connection status
	Latency     int64      // Round-trip latency in milliseconds
	Reputation  float64    // Server reputation score (0.0-1.0)
}

// FederationState tracks the state of the federation network
type FederationState struct {
	mu               sync.RWMutex
	ConnectedServers map[string]*ServerInfo // ServerID -> ServerInfo
	PlayerCounts     map[string]int         // ServerID -> player count
	MarketPrices     map[string]float64     // ItemID -> price
	lastHeartbeat    time.Time
	lastMarketSync   time.Time
}

// NewFederationState creates a new federation state tracker
func NewFederationState() *FederationState {
	return &FederationState{
		ConnectedServers: make(map[string]*ServerInfo),
		PlayerCounts:     make(map[string]int),
		MarketPrices:     make(map[string]float64),
		lastHeartbeat:    time.Now(),
		lastMarketSync:   time.Now(),
	}
}

// AddServer registers a new connected server
func (fs *FederationState) AddServer(info *ServerInfo) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	info.LastSeen = time.Now()
	info.IsOnline = true
	fs.ConnectedServers[info.ServerID] = info
	fs.PlayerCounts[info.ServerID] = info.PlayerCount
}

// RemoveServer unregisters a disconnected server
func (fs *FederationState) RemoveServer(serverID string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	delete(fs.ConnectedServers, serverID)
	delete(fs.PlayerCounts, serverID)
}

// UpdateServer updates metadata for an existing server
func (fs *FederationState) UpdateServer(serverID string, updates func(*ServerInfo)) bool {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	info, exists := fs.ConnectedServers[serverID]
	if !exists {
		return false
	}

	updates(info)
	info.LastSeen = time.Now()
	fs.PlayerCounts[serverID] = info.PlayerCount
	return true
}

// GetServer retrieves server info by ID (thread-safe copy)
func (fs *FederationState) GetServer(serverID string) (*ServerInfo, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	info, exists := fs.ConnectedServers[serverID]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent concurrent modification
	copy := *info
	return &copy, true
}

// GetAllServers returns a copy of all connected servers
func (fs *FederationState) GetAllServers() []*ServerInfo {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	servers := make([]*ServerInfo, 0, len(fs.ConnectedServers))
	for _, info := range fs.ConnectedServers {
		copy := *info
		servers = append(servers, &copy)
	}
	return servers
}

// GetTotalPlayers returns the total player count across all servers
func (fs *FederationState) GetTotalPlayers() int {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	total := 0
	for _, count := range fs.PlayerCounts {
		total += count
	}
	return total
}

// UpdateMarketPrice updates the price for an item
func (fs *FederationState) UpdateMarketPrice(itemID string, price float64) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.MarketPrices[itemID] = price
	fs.lastMarketSync = time.Now()
}

// GetMarketPrice retrieves the current price for an item
func (fs *FederationState) GetMarketPrice(itemID string) (float64, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	price, exists := fs.MarketPrices[itemID]
	return price, exists
}

// GetAllMarketPrices returns a copy of all market prices
func (fs *FederationState) GetAllMarketPrices() map[string]float64 {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	prices := make(map[string]float64, len(fs.MarketPrices))
	for itemID, price := range fs.MarketPrices {
		prices[itemID] = price
	}
	return prices
}

// CheckStaleServers marks servers as offline if they haven't sent heartbeat in 30 seconds
func (fs *FederationState) CheckStaleServers(timeout time.Duration) []string {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	now := time.Now()
	staleServers := []string{}

	for serverID, info := range fs.ConnectedServers {
		if now.Sub(info.LastSeen) > timeout {
			info.IsOnline = false
			staleServers = append(staleServers, serverID)
		}
	}

	return staleServers
}

// GetLastHeartbeat returns the timestamp of the last heartbeat sent
func (fs *FederationState) GetLastHeartbeat() time.Time {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.lastHeartbeat
}

// UpdateHeartbeat updates the last heartbeat timestamp
func (fs *FederationState) UpdateHeartbeat() {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	fs.lastHeartbeat = time.Now()
}

// GetLastMarketSync returns the timestamp of the last market sync
func (fs *FederationState) GetLastMarketSync() time.Time {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.lastMarketSync
}

// ServerCount returns the number of connected servers
func (fs *FederationState) ServerCount() int {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return len(fs.ConnectedServers)
}

// OnlineServerCount returns the number of online servers
func (fs *FederationState) OnlineServerCount() int {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	count := 0
	for _, info := range fs.ConnectedServers {
		if info.IsOnline {
			count++
		}
	}
	return count
}

// HeartbeatMessage represents a periodic server status update
type HeartbeatMessage struct {
	ServerID    string    // Sender server ID
	Timestamp   time.Time // Message timestamp
	PlayerCount int       // Active players
	Latency     int64     // Round-trip latency in milliseconds
	IsOnline    bool      // Server status
}

// MarketSyncMessage represents market price updates
type MarketSyncMessage struct {
	ServerID  string             // Sender server ID
	Timestamp time.Time          // Message timestamp
	Prices    map[string]float64 // ItemID -> price
}

// PoliticalEvent represents a political state change
type PoliticalEvent struct {
	EventID     string    // Unique event identifier
	Type        EventType // Event type
	ServerIDs   []string  // Participating servers
	Timestamp   time.Time // Event occurrence time
	Duration    int64     // Duration in seconds
	Description string    // Human-readable description
}

// EventType represents different political event types
type EventType int

const (
	// EventAlliance represents an alliance between servers
	EventAlliance EventType = iota
	// EventWar represents a war declaration
	EventWar
	// EventTreaty represents a peace treaty
	EventTreaty
	// EventEmbargo represents a trade embargo
	EventEmbargo
	// EventTradePact represents a trade agreement
	EventTradePact
)

// String returns human-readable event type name
func (e EventType) String() string {
	switch e {
	case EventAlliance:
		return "Alliance"
	case EventWar:
		return "War"
	case EventTreaty:
		return "Treaty"
	case EventEmbargo:
		return "Embargo"
	case EventTradePact:
		return "TradePact"
	default:
		return "Unknown"
	}
}

// SyncManager manages periodic synchronization tasks
type SyncManager struct {
	state              *FederationState
	heartbeatInterval  time.Duration
	marketSyncInterval time.Duration
	staleTimeout       time.Duration
	stopChan           chan struct{}
	wg                 sync.WaitGroup
}

// NewSyncManager creates a new sync manager with default intervals
func NewSyncManager(state *FederationState) *SyncManager {
	return &SyncManager{
		state:              state,
		heartbeatInterval:  10 * time.Second, // 10 seconds
		marketSyncInterval: 60 * time.Second, // 60 seconds
		staleTimeout:       30 * time.Second, // 30 seconds
		stopChan:           make(chan struct{}),
	}
}

// SetHeartbeatInterval configures the heartbeat frequency
func (sm *SyncManager) SetHeartbeatInterval(interval time.Duration) {
	sm.heartbeatInterval = interval
}

// SetMarketSyncInterval configures the market sync frequency
func (sm *SyncManager) SetMarketSyncInterval(interval time.Duration) {
	sm.marketSyncInterval = interval
}

// SetStaleTimeout configures the stale server detection timeout
func (sm *SyncManager) SetStaleTimeout(timeout time.Duration) {
	sm.staleTimeout = timeout
}

// Start begins periodic synchronization tasks
func (sm *SyncManager) Start() {
	sm.wg.Add(1)
	go sm.syncLoop()
}

// Stop terminates synchronization tasks
func (sm *SyncManager) Stop() {
	close(sm.stopChan)
	sm.wg.Wait()
}

// syncLoop performs periodic sync tasks
func (sm *SyncManager) syncLoop() {
	defer sm.wg.Done()

	heartbeatTicker := time.NewTicker(sm.heartbeatInterval)
	marketTicker := time.NewTicker(sm.marketSyncInterval)
	staleTicker := time.NewTicker(sm.staleTimeout)

	defer heartbeatTicker.Stop()
	defer marketTicker.Stop()
	defer staleTicker.Stop()

	for {
		select {
		case <-sm.stopChan:
			return
		case <-heartbeatTicker.C:
			sm.state.UpdateHeartbeat()
		case <-marketTicker.C:
			// Market sync timestamp updated when prices are updated
			// This ticker exists for future broadcast implementation
		case <-staleTicker.C:
			sm.state.CheckStaleServers(sm.staleTimeout)
		}
	}
}

// CreateHeartbeat creates a heartbeat message for the local server
func (sm *SyncManager) CreateHeartbeat(serverID string, playerCount int) *HeartbeatMessage {
	return &HeartbeatMessage{
		ServerID:    serverID,
		Timestamp:   time.Now(),
		PlayerCount: playerCount,
		IsOnline:    true,
	}
}

// ProcessHeartbeat handles an incoming heartbeat from a peer server
func (sm *SyncManager) ProcessHeartbeat(msg *HeartbeatMessage) {
	sm.state.UpdateServer(msg.ServerID, func(info *ServerInfo) {
		info.PlayerCount = msg.PlayerCount
		info.IsOnline = msg.IsOnline
		info.Latency = msg.Latency
	})
}

// CreateMarketSync creates a market sync message with current prices
func (sm *SyncManager) CreateMarketSync(serverID string) *MarketSyncMessage {
	return &MarketSyncMessage{
		ServerID:  serverID,
		Timestamp: time.Now(),
		Prices:    sm.state.GetAllMarketPrices(),
	}
}

// ProcessMarketSync handles an incoming market sync from a peer server
func (sm *SyncManager) ProcessMarketSync(msg *MarketSyncMessage) {
	for itemID, price := range msg.Prices {
		sm.state.UpdateMarketPrice(itemID, price)
	}
}
