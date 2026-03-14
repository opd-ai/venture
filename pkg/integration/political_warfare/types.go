package political_warfare

import "time"

// WarDeclaration represents a war between two guilds
type WarDeclaration struct {
	AttackerGuildID   string        `json:"attacker_guild_id"`
	DefenderGuildID   string        `json:"defender_guild_id"`
	DeclaredAt        time.Time     `json:"declared_at"`
	PreparationEnds   time.Time     `json:"preparation_ends"`
	PreparationPeriod time.Duration `json:"preparation_period"`
	Active            bool          `json:"active"`
	Ended             bool          `json:"ended"`
	EndedAt           time.Time     `json:"ended_at,omitempty"`
	Victor            string        `json:"victor,omitempty"` // Guild ID of winner
	VictoryType       VictoryType   `json:"victory_type,omitempty"`
}

// VictoryType represents how a war was won
type VictoryType string

const (
	VictoryTypeMilitary   VictoryType = "military"   // Captured all control points
	VictoryTypeDiplomatic VictoryType = "diplomatic" // Negotiated surrender
	VictoryTypeDefault    VictoryType = "default"    // Enemy guild disbanded or abandoned
)

// PeaceTreaty represents a peace agreement between guilds
type PeaceTreaty struct {
	GuildID1     string        `json:"guild_id_1"`
	GuildID2     string        `json:"guild_id_2"`
	SignedAt     time.Time     `json:"signed_at"`
	ExpiresAt    time.Time     `json:"expires_at"`
	CooldownEnds time.Time     `json:"cooldown_ends"`
	Duration     time.Duration `json:"duration"`
	Active       bool          `json:"active"`
}

// TradeEmbargo represents a trade restriction between guilds
type TradeEmbargo struct {
	ImposingGuildID string    `json:"imposing_guild_id"`
	TargetGuildID   string    `json:"target_guild_id"`
	PriceIncrease   float64   `json:"price_increase"` // Percentage markup (0.5 = 50%, 0.9 = 90%)
	ImposedAt       time.Time `json:"imposed_at"`
	ExpiresAt       time.Time `json:"expires_at,omitempty"`
	Active          bool      `json:"active"`
}

// AllianceCall represents a request for siege reinforcements
type AllianceCall struct {
	CallingGuildID   string             `json:"calling_guild_id"`
	TargetGuildID    string             `json:"target_guild_id"` // Guild being attacked
	CalledAt         time.Time          `json:"called_at"`
	RespondingAllies []AllianceResponse `json:"responding_allies"`
	Completed        bool               `json:"completed"`
}

// AllianceResponse represents an ally's response to a reinforcement call
type AllianceResponse struct {
	AllyGuildID string    `json:"ally_guild_id"`
	Accepted    bool      `json:"accepted"`
	RespondedAt time.Time `json:"responded_at"`
	SuccessRate float64   `json:"success_rate"` // Based on political relations (0.6-0.8)
}

// DiplomaticConcession represents terms offered for diplomatic victory
type DiplomaticConcession struct {
	Type  ConcessionType `json:"type"`
	Value interface{}    `json:"value"` // Type-specific value
}

// ConcessionType represents types of concessions for negotiation
type ConcessionType string

const (
	ConcessionGold      ConcessionType = "gold"      // Value: int (gold amount)
	ConcessionTerritory ConcessionType = "territory" // Value: string (territory ID)
	ConcessionApology   ConcessionType = "apology"   // Value: string (public apology text)
	ConcessionTribute   ConcessionType = "tribute"   // Value: []string (item IDs)
	ConcessionTrade     ConcessionType = "trade"     // Value: float64 (discount %)
)

// ReputationPenalty represents a reputation impact from aggressive actions
type ReputationPenalty struct {
	GuildID   string    `json:"guild_id"`
	Action    string    `json:"action"`  // "attack", "siege", "embargo", "war_declaration"
	Penalty   float64   `json:"penalty"` // Negative value (-0.1 to -0.5)
	AppliedAt time.Time `json:"applied_at"`
	FactionID string    `json:"faction_id"` // NPC faction affected
}

// String returns string representation of VictoryType
func (vt VictoryType) String() string {
	return string(vt)
}

// String returns string representation of ConcessionType
func (ct ConcessionType) String() string {
	return string(ct)
}

// AppliedConcession records a concession that has been applied
type AppliedConcession struct {
	Type              ConcessionType `json:"type"`
	AttackerGuildID   string         `json:"attacker_guild_id"`
	DefenderGuildID   string         `json:"defender_guild_id"`
	AppliedAt         time.Time      `json:"applied_at"`
	TerritoryID       string         `json:"territory_id,omitempty"`       // For ConcessionTerritory
	ApologyText       string         `json:"apology_text,omitempty"`       // For ConcessionApology
	TributeItemIDs    []string       `json:"tribute_item_ids,omitempty"`   // For ConcessionTribute
	TradeDiscountPct  float64        `json:"trade_discount_pct,omitempty"` // For ConcessionTrade
	TradeDiscountEnds time.Time      `json:"trade_discount_ends,omitempty"`
	GoldAmount        int            `json:"gold_amount,omitempty"` // For ConcessionGold
}

// TradeDiscountDuration is how long trade discount concessions last
const TradeDiscountDuration = 30 * 24 * time.Hour // 30 days

// Concession value calculation constants.
// These are used in calculateConcessionValue() to normalize different
// concession types to a common value scale for diplomatic victory calculations.
const (
	// GoldValueNormalizer normalizes gold amounts to concession value.
	// 10,000 gold = 1.0 concession value.
	GoldValueNormalizer = 10000.0

	// TerritoryValueEquivalent is the concession value of one territory.
	// Equivalent to approximately 20,000 gold.
	TerritoryValueEquivalent = 2.0

	// ApologyValue is the concession value of a public apology.
	// Small symbolic value representing reputation damage.
	ApologyValue = 0.1

	// ItemValueEquivalent is the concession value per tribute item.
	// Each item is worth approximately 5,000 gold equivalent.
	ItemValueEquivalent = 0.5

	// TradeDiscountMultiplier converts discount percentage to concession value.
	// A 10% discount adds 0.05 concession value (10 * 0.5 / 100 = 0.05).
	TradeDiscountMultiplier = 0.5

	// DefaultSeed is the fallback seed when no world seed is provided.
	// Used for deterministic political calculations.
	DefaultSeed = int64(12345)
)
