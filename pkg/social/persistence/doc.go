// Package persistence provides persistent social data structures for Venture.
//
// This package implements:
//   - Persistent trust scores that survive server restarts
//   - Reputation tracking with automatic decay over time
//   - Trust level tiers (Stranger, Acquaintance, Friend, Trusted)
//   - Trade limits based on trust levels
//   - Cross-server trust synchronization via federation
//
// # Trust Levels
//
// Trust scores range from 0.0 to 1.0 and are categorized into tiers:
//   - Stranger: 0.0-0.3 (can trade common items only)
//   - Acquaintance: 0.3-0.6 (can trade common + uncommon items)
//   - Friend: 0.6-0.8 (can trade up to rare items)
//   - Trusted: 0.8-1.0 (can trade all items including legendary)
//
// # Reputation Decay
//
// Trust scores decay over time at a rate of 0.01 per day of inactivity.
// This encourages active social engagement and prevents stale relationships.
//
// # Usage Example
//
//	manager := persistence.NewTrustManager()
//	
//	// Update trust after successful trade
//	manager.UpdateTrust("player1", "player2", 0.05, time.Now())
//	
//	// Check trust level
//	level := manager.GetTrustLevel("player1", "player2")
//	if level >= persistence.TrustLevelFriend {
//	    // Allow rare item trade
//	}
//	
//	// Apply daily decay
//	manager.ApplyDecay(time.Now())
//	
//	// Save persistent data
//	data, _ := manager.Save()
//	ioutil.WriteFile("trust.json.gz", data, 0644)
package persistence
