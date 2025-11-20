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
	CallingGuildID  string             `json:"calling_guild_id"`
	TargetGuildID   string             `json:"target_guild_id"` // Guild being attacked
	CalledAt        time.Time          `json:"called_at"`
	ResponingAllies []AllianceResponse `json:"responding_allies"`
	Completed       bool               `json:"completed"`
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
