package guild_housing

// upgrades.go defines guild house upgrade tiers, costs, and bonuses.
// This includes tier definitions and bonus calculation logic.
//
// Code relocated from: types.go

// UpgradeTier represents guild house upgrade levels.
type UpgradeTier int

const (
	TierBasic    UpgradeTier = iota // Base tier, no bonuses
	TierStandard                    // 10k gold, +20% bonuses
	TierAdvanced                    // 50k gold, +50% bonuses
	TierMaster                      // 100k gold, +100% bonuses
)

// String returns human-readable tier name.
func (t UpgradeTier) String() string {
	switch t {
	case TierBasic:
		return "Basic"
	case TierStandard:
		return "Standard"
	case TierAdvanced:
		return "Advanced"
	case TierMaster:
		return "Master"
	default:
		return "Unknown"
	}
}

// Cost returns upgrade cost in gold.
func (t UpgradeTier) Cost() int {
	switch t {
	case TierBasic:
		return 0
	case TierStandard:
		return 10000
	case TierAdvanced:
		return 50000
	case TierMaster:
		return 100000
	default:
		return 0
	}
}

// BonusMultiplier returns upgrade bonus multiplier.
func (t UpgradeTier) BonusMultiplier() float64 {
	switch t {
	case TierBasic:
		return 1.0
	case TierStandard:
		return 1.2
	case TierAdvanced:
		return 1.5
	case TierMaster:
		return 2.0
	default:
		return 1.0
	}
}
