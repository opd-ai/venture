// Package engine provides merchant gambling functionality for Phase 27.3.
// Allows players to bet gold on mini-games at merchant locations.
package engine

// MerchantGamblingComponent adds gambling mini-game support to merchant entities.
// Merchants can offer dice games, card games, etc. with betting mechanics.
//
// Phase 27.3: Mini-Game Integration - Merchant Gambling
type MerchantGamblingComponent struct {
	// OffersGambling indicates if this merchant has gambling games
	OffersGambling bool

	// GamesOffered lists which mini-game types are available for gambling
	GamesOffered []MiniGameType

	// MinimumBet is the lowest gold amount accepted for betting
	MinimumBet int

	// MaximumBet is the highest gold amount accepted for betting
	MaximumBet int

	// HouseEdge is the merchant's advantage (0.0-1.0)
	// Higher values mean better odds for the merchant
	HouseEdge float64

	// WinMultiplier determines payout for winning (e.g., 2.0 = double your bet)
	WinMultiplier float64

	// CurrentBet tracks active bet amount (0 if no active bet)
	CurrentBet int

	// CurrentGameType tracks which game is being bet on
	CurrentGameType MiniGameType

	// BettingPlayerID is the entity ID of player currently betting
	BettingPlayerID uint64
}

// Type returns the component type identifier "merchantGambling".
func (m MerchantGamblingComponent) Type() string {
	return "merchantGambling"
}

// NewMerchantGamblingComponent creates a gambling component with default values.
// Common for tavern merchants in safe zones.
func NewMerchantGamblingComponent() *MerchantGamblingComponent {
	return &MerchantGamblingComponent{
		OffersGambling:  true,
		GamesOffered:    []MiniGameType{MiniGameCard, MiniGameDice},
		MinimumBet:      10,
		MaximumBet:      100,
		HouseEdge:       0.1, // 10% house advantage
		WinMultiplier:   1.8, // Win pays 1.8x bet (net profit 0.8x)
		CurrentBet:      0,
		BettingPlayerID: 0,
	}
}

// CanPlaceBet returns true if player can bet the specified amount.
func (m *MerchantGamblingComponent) CanPlaceBet(betAmount int) bool {
	if !m.OffersGambling {
		return false
	}
	if m.CurrentBet > 0 {
		return false // Already has active bet
	}
	if betAmount < m.MinimumBet {
		return false
	}
	if betAmount > m.MaximumBet {
		return false
	}
	return true
}

// PlaceBet starts a gambling game with the specified bet amount.
// Returns true if bet was accepted.
func (m *MerchantGamblingComponent) PlaceBet(playerID uint64, betAmount int, gameType MiniGameType) bool {
	if !m.CanPlaceBet(betAmount) {
		return false
	}

	// Verify game type is offered
	gameOffered := false
	for _, offered := range m.GamesOffered {
		if offered == gameType {
			gameOffered = true
			break
		}
	}
	if !gameOffered {
		return false
	}

	m.CurrentBet = betAmount
	m.CurrentGameType = gameType
	m.BettingPlayerID = playerID
	return true
}

// CalculateWinnings determines payout based on bet and game result.
// playerWon indicates if player succeeded at the mini-game.
// Returns gold amount to award (includes original bet if won).
func (m *MerchantGamblingComponent) CalculateWinnings(playerWon bool) int {
	if playerWon {
		// Player wins: return original bet + winnings
		winnings := float64(m.CurrentBet) * m.WinMultiplier
		return m.CurrentBet + int(winnings)
	}
	// Player loses: merchant keeps the bet
	return 0
}

// ClearBet resets gambling state after game completes.
func (m *MerchantGamblingComponent) ClearBet() {
	m.CurrentBet = 0
	m.CurrentGameType = 0
	m.BettingPlayerID = 0
}

// HasActiveBet returns true if there's currently a bet in progress.
func (m *MerchantGamblingComponent) HasActiveBet() bool {
	return m.CurrentBet > 0
}
