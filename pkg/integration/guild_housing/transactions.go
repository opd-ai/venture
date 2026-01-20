package guild_housing

import "time"

// transactions.go defines guild storage transaction tracking types.
// This includes deposit/withdraw logging and transaction history.
//
// Code relocated from: types.go

// Transaction represents a storage deposit/withdraw log entry.
type Transaction struct {
	TransactionID string
	PlayerID      string
	ItemID        string
	Quantity      int
	Action        TransactionType
	Timestamp     time.Time
}

// TransactionType represents deposit or withdrawal.
type TransactionType int

const (
	TransactionDeposit  TransactionType = iota // Item deposited
	TransactionWithdraw                        // Item withdrawn
)

// String returns human-readable transaction type.
func (t TransactionType) String() string {
	switch t {
	case TransactionDeposit:
		return "Deposit"
	case TransactionWithdraw:
		return "Withdraw"
	default:
		return "Unknown"
	}
}
