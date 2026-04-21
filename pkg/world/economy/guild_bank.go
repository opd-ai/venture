package economy

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// GuildBankManager manages guild bank vaults with cross-server treasury sync.
type GuildBankManager struct {
	vaults       map[string]*GuildVault // guildID -> vault
	timeProvider TimeProvider
	mu           sync.RWMutex
}

// NewGuildBankManager creates a new guild bank manager.
func NewGuildBankManager() *GuildBankManager {
	return NewGuildBankManagerWithTime(DefaultTimeProvider())
}

// NewGuildBankManagerWithTime creates a new guild bank manager with a custom time provider.
func NewGuildBankManagerWithTime(tp TimeProvider) *GuildBankManager {
	return &GuildBankManager{
		vaults:       make(map[string]*GuildVault),
		timeProvider: tp,
	}
}

// GuildVault represents a guild's shared storage and treasury.
type GuildVault struct {
	GuildID           string
	Items             map[string]*VaultItem // itemID -> item
	GoldBalance       int
	InterestRate      float64 // 0.001-0.01 (0.1%-1.0% daily)
	LastInterestTime  time.Time
	WithdrawalLimits  map[string]int              // rankID -> daily gold limit
	MemberWithdrawals map[string]*DailyWithdrawal // memberID -> today's withdrawals
	AuditLog          []*AuditEntry
	MaxAuditEntries   int
	CreatedAt         time.Time
	LastSyncTime      time.Time
}

// VaultItem represents an item stored in the guild vault.
type VaultItem struct {
	ItemID        string
	ItemName      string
	ItemType      string
	Quantity      int
	MaxStackSize  int
	ContributedBy string // Last member who deposited
	LastModified  time.Time
}

// DailyWithdrawal tracks a member's gold withdrawals for the current day.
type DailyWithdrawal struct {
	MemberID       string
	Date           time.Time
	TotalWithdrawn int
}

// AuditEntry records a transaction in the guild vault.
type AuditEntry struct {
	EntryID       string
	Timestamp     time.Time
	MemberID      string
	MemberName    string
	ActionType    AuditAction
	ItemID        string // Empty for gold transactions
	ItemName      string
	Quantity      int
	GoldAmount    int
	BalanceBefore int
	BalanceAfter  int
	Notes         string
}

// AuditAction defines the type of vault action.
type AuditAction int

const (
	// AuditDeposit records a deposit transaction.
	AuditDeposit AuditAction = iota

	// AuditWithdraw records a withdrawal transaction.
	AuditWithdraw

	// AuditInterest records interest earned.
	AuditInterest
)

// String returns the name of the audit action.
func (aa AuditAction) String() string {
	switch aa {
	case AuditDeposit:
		return "Deposit"
	case AuditWithdraw:
		return "Withdraw"
	case AuditInterest:
		return "Interest"
	default:
		return "Unknown"
	}
}

// CreateVault creates a new guild vault.
func (m *GuildBankManager) CreateVault(guildID string, interestRate float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.vaults[guildID]; exists {
		return fmt.Errorf("vault already exists for guild %s", guildID)
	}

	// Validate interest rate (0.1% to 1.0% daily)
	if interestRate < 0.001 || interestRate > 0.01 {
		return fmt.Errorf("interest rate must be between 0.001 and 0.01, got %f", interestRate)
	}

	now := m.timeProvider.Now()
	m.vaults[guildID] = &GuildVault{
		GuildID:           guildID,
		Items:             make(map[string]*VaultItem),
		GoldBalance:       0,
		InterestRate:      interestRate,
		LastInterestTime:  now,
		WithdrawalLimits:  make(map[string]int),
		MemberWithdrawals: make(map[string]*DailyWithdrawal),
		AuditLog:          make([]*AuditEntry, 0),
		MaxAuditEntries:   1000, // 30 days at ~33 transactions/day
		CreatedAt:         now,
		LastSyncTime:      now,
	}

	return nil
}

// GetVault retrieves a guild vault (read-only copy).
func (m *GuildBankManager) GetVault(guildID string) (*GuildVault, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vault, exists := m.vaults[guildID]
	if !exists {
		return nil, fmt.Errorf("vault not found for guild %s", guildID)
	}

	// Return a copy to prevent external modification
	vaultCopy := *vault
	vaultCopy.Items = make(map[string]*VaultItem, len(vault.Items))
	for k, v := range vault.Items {
		itemCopy := *v
		vaultCopy.Items[k] = &itemCopy
	}
	vaultCopy.AuditLog = make([]*AuditEntry, len(vault.AuditLog))
	copy(vaultCopy.AuditLog, vault.AuditLog)

	return &vaultCopy, nil
}

// DepositGold deposits gold into the guild vault.
func (m *GuildBankManager) DepositGold(guildID, memberID, memberName string, amount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vault, exists := m.vaults[guildID]
	if !exists {
		return fmt.Errorf("vault not found for guild %s", guildID)
	}

	if amount <= 0 {
		return fmt.Errorf("deposit amount must be positive, got %d", amount)
	}

	balanceBefore := vault.GoldBalance
	vault.GoldBalance += amount

	// Add audit entry
	now := m.timeProvider.Now()
	m.addAuditEntry(vault, &AuditEntry{
		EntryID:       fmt.Sprintf("%s-%d", memberID, now.UnixNano()),
		Timestamp:     now,
		MemberID:      memberID,
		MemberName:    memberName,
		ActionType:    AuditDeposit,
		GoldAmount:    amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  vault.GoldBalance,
		Notes:         "Gold deposit",
	})

	log.WithFields(log.Fields{
		"guildID":       guildID,
		"memberID":      memberID,
		"amount":        amount,
		"balanceBefore": balanceBefore,
		"balanceAfter":  vault.GoldBalance,
	}).Debug("gold deposited into guild vault")

	return nil
}

// WithdrawGold withdraws gold from the guild vault with rank-based limits.
func (m *GuildBankManager) WithdrawGold(guildID, memberID, memberName, rankID string, amount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vault, exists := m.vaults[guildID]
	if !exists {
		return fmt.Errorf("vault not found for guild %s", guildID)
	}

	if amount <= 0 {
		return fmt.Errorf("withdrawal amount must be positive, got %d", amount)
	}

	if vault.GoldBalance < amount {
		return fmt.Errorf("insufficient funds: vault has %d gold, requested %d", vault.GoldBalance, amount)
	}

	// Check withdrawal limit
	limit, hasLimit := vault.WithdrawalLimits[rankID]
	if hasLimit {
		// Check today's withdrawals
		today := m.timeProvider.Now().Truncate(24 * time.Hour)
		withdrawal, exists := vault.MemberWithdrawals[memberID]

		if !exists || !withdrawal.Date.Equal(today) {
			// First withdrawal today
			vault.MemberWithdrawals[memberID] = &DailyWithdrawal{
				MemberID:       memberID,
				Date:           today,
				TotalWithdrawn: amount,
			}
		} else {
			// Check if limit would be exceeded
			if withdrawal.TotalWithdrawn+amount > limit {
				return fmt.Errorf("daily withdrawal limit exceeded: %d/%d gold used, requested %d",
					withdrawal.TotalWithdrawn, limit, amount)
			}
			withdrawal.TotalWithdrawn += amount
		}
	}

	balanceBefore := vault.GoldBalance
	vault.GoldBalance -= amount

	// Add audit entry
	withdrawNow := m.timeProvider.Now()
	m.addAuditEntry(vault, &AuditEntry{
		EntryID:       fmt.Sprintf("%s-%d", memberID, withdrawNow.UnixNano()),
		Timestamp:     withdrawNow,
		MemberID:      memberID,
		MemberName:    memberName,
		ActionType:    AuditWithdraw,
		GoldAmount:    amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  vault.GoldBalance,
		Notes:         "Gold withdrawal",
	})

	log.WithFields(log.Fields{
		"guildID":       guildID,
		"memberID":      memberID,
		"amount":        amount,
		"balanceBefore": balanceBefore,
		"balanceAfter":  vault.GoldBalance,
	}).Debug("gold withdrawn from guild vault")

	return nil
}

// DepositItem deposits an item into the guild vault.
func (m *GuildBankManager) DepositItem(guildID, memberID, memberName, itemID, itemName, itemType string, quantity, maxStackSize int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vault, exists := m.vaults[guildID]
	if !exists {
		return fmt.Errorf("vault not found for guild %s", guildID)
	}

	if quantity <= 0 {
		return fmt.Errorf("deposit quantity must be positive, got %d", quantity)
	}

	// Check vault capacity (5000 unique items)
	if len(vault.Items) >= 5000 {
		if _, exists := vault.Items[itemID]; !exists {
			return fmt.Errorf("vault capacity exceeded: 5000 items")
		}
	}

	now := m.timeProvider.Now()
	if item, exists := vault.Items[itemID]; exists {
		item.Quantity += quantity
		item.ContributedBy = memberID
		item.LastModified = now
	} else {
		vault.Items[itemID] = &VaultItem{
			ItemID:        itemID,
			ItemName:      itemName,
			ItemType:      itemType,
			Quantity:      quantity,
			MaxStackSize:  maxStackSize,
			ContributedBy: memberID,
			LastModified:  now,
		}
	}

	// Add audit entry
	m.addAuditEntry(vault, &AuditEntry{
		EntryID:       fmt.Sprintf("%s-%d", memberID, now.UnixNano()),
		Timestamp:     now,
		MemberID:      memberID,
		MemberName:    memberName,
		ActionType:    AuditDeposit,
		ItemID:        itemID,
		ItemName:      itemName,
		Quantity:      quantity,
		BalanceBefore: vault.GoldBalance,
		BalanceAfter:  vault.GoldBalance,
		Notes:         "Item deposit",
	})

	return nil
}

// WithdrawItem withdraws an item from the guild vault.
func (m *GuildBankManager) WithdrawItem(guildID, memberID, memberName, itemID string, quantity int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vault, exists := m.vaults[guildID]
	if !exists {
		return fmt.Errorf("vault not found for guild %s", guildID)
	}

	if quantity <= 0 {
		return fmt.Errorf("withdrawal quantity must be positive, got %d", quantity)
	}

	item, exists := vault.Items[itemID]
	if !exists {
		return fmt.Errorf("item %s not found in vault", itemID)
	}

	if item.Quantity < quantity {
		return fmt.Errorf("insufficient quantity: vault has %d, requested %d", item.Quantity, quantity)
	}

	item.Quantity -= quantity
	item.LastModified = m.timeProvider.Now()

	// Remove item if quantity reaches zero
	if item.Quantity == 0 {
		delete(vault.Items, itemID)
	}

	// Add audit entry
	withdrawItemNow := m.timeProvider.Now()
	m.addAuditEntry(vault, &AuditEntry{
		EntryID:       fmt.Sprintf("%s-%d", memberID, withdrawItemNow.UnixNano()),
		Timestamp:     withdrawItemNow,
		MemberID:      memberID,
		MemberName:    memberName,
		ActionType:    AuditWithdraw,
		ItemID:        itemID,
		ItemName:      item.ItemName,
		Quantity:      quantity,
		BalanceBefore: vault.GoldBalance,
		BalanceAfter:  vault.GoldBalance,
		Notes:         "Item withdrawal",
	})

	return nil
}

// SetWithdrawalLimit sets the daily gold withdrawal limit for a rank.
func (m *GuildBankManager) SetWithdrawalLimit(guildID, rankID string, limit int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vault, exists := m.vaults[guildID]
	if !exists {
		return fmt.Errorf("vault not found for guild %s", guildID)
	}

	if limit < 0 {
		return fmt.Errorf("withdrawal limit must be non-negative, got %d", limit)
	}

	vault.WithdrawalLimits[rankID] = limit
	return nil
}

// CalculateInterest calculates and applies daily interest to the vault.
func (m *GuildBankManager) CalculateInterest(guildID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vault, exists := m.vaults[guildID]
	if !exists {
		return fmt.Errorf("vault not found for guild %s", guildID)
	}

	now := m.timeProvider.Now()
	daysSinceLastInterest := now.Sub(vault.LastInterestTime).Hours() / 24.0

	// Only apply interest if at least 1 day has passed
	if daysSinceLastInterest < 1.0 {
		return nil
	}

	// Calculate compound interest for full days
	days := int(daysSinceLastInterest)
	balanceBefore := vault.GoldBalance

	for i := 0; i < days; i++ {
		interest := int(float64(vault.GoldBalance) * vault.InterestRate)
		if interest > 0 {
			vault.GoldBalance += interest

			// Add audit entry for interest
			m.addAuditEntry(vault, &AuditEntry{
				EntryID:       fmt.Sprintf("interest-%d", now.UnixNano()+int64(i)),
				Timestamp:     now,
				MemberID:      "system",
				MemberName:    "Guild Bank",
				ActionType:    AuditInterest,
				GoldAmount:    interest,
				BalanceBefore: balanceBefore + (i * interest),
				BalanceAfter:  vault.GoldBalance,
				Notes:         fmt.Sprintf("Daily interest (%.1f%%)", vault.InterestRate*100),
			})
		}
	}

	vault.LastInterestTime = now.Truncate(24 * time.Hour)
	return nil
}

// SetLastInterestTime sets the last interest time for a vault (for testing).
func (m *GuildBankManager) SetLastInterestTime(guildID string, t time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vault, exists := m.vaults[guildID]
	if !exists {
		return fmt.Errorf("vault not found for guild %s", guildID)
	}

	vault.LastInterestTime = t
	return nil
}

// SetMaxAuditEntries sets the maximum audit log entries for a vault (for testing).
func (m *GuildBankManager) SetMaxAuditEntries(guildID string, max int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vault, exists := m.vaults[guildID]
	if !exists {
		return fmt.Errorf("vault not found for guild %s", guildID)
	}

	if max <= 0 {
		return fmt.Errorf("max audit entries must be positive, got %d", max)
	}

	vault.MaxAuditEntries = max

	// Trim existing log if necessary
	if len(vault.AuditLog) > max {
		vault.AuditLog = vault.AuditLog[len(vault.AuditLog)-max:]
	}

	return nil
}

// GetAuditLog retrieves the last N audit entries.
func (m *GuildBankManager) GetAuditLog(guildID string, limit int) ([]*AuditEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vault, exists := m.vaults[guildID]
	if !exists {
		return nil, fmt.Errorf("vault not found for guild %s", guildID)
	}

	logSize := len(vault.AuditLog)
	if limit <= 0 || limit > logSize {
		limit = logSize
	}

	// Return most recent entries
	start := logSize - limit
	entries := make([]*AuditEntry, limit)
	copy(entries, vault.AuditLog[start:])

	return entries, nil
}

// addAuditEntry adds an entry to the vault's audit log with LRU eviction.
func (m *GuildBankManager) addAuditEntry(vault *GuildVault, entry *AuditEntry) {
	vault.AuditLog = append(vault.AuditLog, entry)

	// Evict oldest entries if limit exceeded
	if len(vault.AuditLog) > vault.MaxAuditEntries {
		vault.AuditLog = vault.AuditLog[len(vault.AuditLog)-vault.MaxAuditEntries:]
	}
}

// UpdateSyncTime updates the last cross-server sync timestamp.
func (m *GuildBankManager) UpdateSyncTime(guildID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vault, exists := m.vaults[guildID]
	if !exists {
		return fmt.Errorf("vault not found for guild %s", guildID)
	}

	vault.LastSyncTime = m.timeProvider.Now()
	return nil
}

// Save persists guild vaults to disk with gzip compression.
func (m *GuildBankManager) Save(filename string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	gzipWriter := gzip.NewWriter(file)
	encoder := json.NewEncoder(gzipWriter)
	if err := encoder.Encode(m.vaults); err != nil {
		gzipWriter.Close()
		file.Close()
		return fmt.Errorf("failed to encode vaults: %w", err)
	}

	if err := gzipWriter.Close(); err != nil {
		file.Close()
		return fmt.Errorf("failed to flush gzip writer: %w", err)
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("failed to close file: %w", err)
	}

	return nil
}

// Load reads guild vaults from disk.
func (m *GuildBankManager) Load(filename string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Try gzip decompression
	gzipReader, err := gzip.NewReader(file)
	if err == nil {
		defer gzipReader.Close()
		decoder := json.NewDecoder(gzipReader)
		if err := decoder.Decode(&m.vaults); err != nil {
			return fmt.Errorf("failed to decode vaults: %w", err)
		}
		return nil
	}

	// Fallback to uncompressed JSON
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&m.vaults); err != nil {
		return fmt.Errorf("failed to decode vaults: %w", err)
	}

	return nil
}
