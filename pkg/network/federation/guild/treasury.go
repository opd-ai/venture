// treasury.go implements federated guild treasury operations including
// deposits, withdrawals, and transaction history for cross-server guilds.
package guild

import (
	"fmt"
)

// Guild treasury operations.
//
// This file manages the shared guild treasury (gold pool) with deposit and
// withdrawal operations. All transactions are logged with timestamps and
// player attribution for audit trails.
//
// Code relocated from: manager.go

// DepositTreasury adds gold to guild treasury
// Originally defined in: manager.go
func (m *Manager) DepositTreasury(guildID, playerID string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	guild, exists := m.guilds[guildID]
	if !exists {
		return fmt.Errorf("guild not found: %s", guildID)
	}

	now := m.timeProvider.Now()
	guild.Treasury += amount
	guild.Transactions = append(guild.Transactions, TreasuryTransaction{
		PlayerID:  playerID,
		Amount:    amount,
		Timestamp: now,
		Reason:    "deposit",
	})
	guild.UpdatedAt = now
	return nil
}

// WithdrawTreasury removes gold from guild treasury
// Originally defined in: manager.go
func (m *Manager) WithdrawTreasury(guildID, playerID string, amount int) error {
	if amount <= 0 {
		return fmt.Errorf("amount must be positive")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	guild, exists := m.guilds[guildID]
	if !exists {
		return fmt.Errorf("guild not found: %s", guildID)
	}

	if guild.Treasury < amount {
		return fmt.Errorf("insufficient treasury funds")
	}

	now := m.timeProvider.Now()
	guild.Treasury -= amount
	guild.Transactions = append(guild.Transactions, TreasuryTransaction{
		PlayerID:  playerID,
		Amount:    -amount,
		Timestamp: now,
		Reason:    "withdrawal",
	})
	guild.UpdatedAt = now
	return nil
}
