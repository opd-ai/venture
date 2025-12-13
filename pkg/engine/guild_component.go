package engine

// GuildComponent tracks an entity's guild membership
type GuildComponent struct {
	GuildID  string // Empty string if not in a guild
	Rank     string // "Recruit", "Member", "Officer", "Leader"
	JoinedAt int64  // Unix timestamp when joined
}

// Type returns the component type identifier
func (g *GuildComponent) Type() string {
	return "guild"
}
