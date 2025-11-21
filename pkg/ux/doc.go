// Package ux provides user experience flow validation and journey testing.
//
// Phase 65.3: User Experience Flow
//
// This package implements automated testing for 20 critical user journeys,
// measuring task completion rates, time to complete, user satisfaction,
// and error rates. Each journey simulates a real player workflow from
// start to finish, ensuring all features are accessible and usable.
//
// # User Journeys
//
// The package validates 20 distinct user experiences:
//   - New player onboarding (character creation → tutorial → first quest)
//   - Crafting workflow (gather → find recipe → craft → equip)
//   - Social interaction (join guild → participate → earn reward)
//   - Exploration (discover dungeon → complete → collect loot)
//   - Trading (list item → sell → receive gold)
//   - Housing (purchase → place furniture → invite friends)
//   - Raiding (join group → defeat boss → distribute loot)
//   - PvP (challenge → duel → earn reputation)
//   - Questing (accept → complete → turn in)
//   - Companion management (tame → train → use in combat)
//   - Vehicle usage (acquire → upgrade → use in travel)
//   - Storytelling (discover lore → complete arc → unlock epilogue)
//   - Prestige progression (max level → unlock prestige → earn points)
//   - Guild leadership (create → recruit → declare war)
//   - Modding (install → configure → observe effects)
//   - Cross-server travel (enter portal → transfer → explore)
//   - Legendary quests (start → complete all steps → claim reward)
//   - Housing decoration (buy → place → showcase)
//   - Siege participation (join → attack/defend → claim territory)
//   - Economy trading (buy low → sell high → profit)
//
// # Metrics
//
// Each journey tracks four key metrics:
//   - Task completion rate: ≥90% target
//   - Time to complete: within expected duration ±20%
//   - User satisfaction: ≥80% positive feedback (simulated)
//   - Error rate: <5% (users getting stuck/confused)
//
// # Usage
//
// Validate all journeys:
//
//	validator := ux.NewJourneyValidator()
//	results := validator.ValidateAll()
//	for _, result := range results {
//	    if !result.Passed {
//	        log.Printf("Journey %s failed: %s", result.Name, result.Error)
//	    }
//	}
//
// Validate specific journey:
//
//	result := validator.ValidateJourney(ux.JourneyNewPlayer)
//	fmt.Printf("Completion rate: %.1f%%\n", result.CompletionRate*100)
//	fmt.Printf("Average duration: %v\n", result.AverageDuration)
//
// # Implementation
//
// Journey validation uses simulation rather than requiring human testers:
//   - Deterministic AI "players" follow the journey steps
//   - Success/failure determined by reachability and functionality
//   - Time estimates based on expected workflow duration
//   - Satisfaction simulated based on feature availability and polish
//
// This allows automated regression testing of UX flows across versions.
package ux
