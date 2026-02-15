# Audit: github.com/opd-ai/venture/pkg/procgen/quest
**Date**: 2026-02-15
**Status**: Needs Work

## Summary
The quest package provides procedural quest generation with deterministic seeded algorithms, good test coverage (90.7%), and clean structure. However, Quest/Objective types contain business logic methods that should be in systems, lack serialize/deserialize support for persistence, and only support 2 of 4 advertised genres (fantasy, sci-fi; missing horror, cyberpunk).

## Issues Found
- [ ] **high** ECS compliance — Quest and Objective types contain logic methods (IsComplete, Progress, GetRewardValue) that should be in systems, not data types (`types.go:137-233`)
- [ ] **high** Integration — Quest types lack Serialize/Deserialize methods needed for save/load persistence (all quest types in `types.go`)
- [ ] **med** Genre support — Package advertises 4 genres (fantasy, sci-fi, horror, cyberpunk) but only implements 2; selectTemplates() has no cases for horror/cyberpunk (`generator.go:88-103`)
- [ ] **low** Documentation — GetFantasy/SciFi template functions lack godoc comments explaining template structure and usage (`types.go:250-404`)
- [ ] **low** Validation — validateQuestRewards only checks XP > 0; should also validate Gold >= 0, Items is valid array (`generator.go:438-443`)

## Test Coverage
90.7% (target: 65%) ✅

## Integration Status
**Engine Integration**: ✅ Fully integrated
- QuestTrackerComponent in pkg/engine/quest_tracker.go wraps quest.Quest for runtime tracking
- Quest UI system in pkg/engine/quest_ui.go for rendering quest lists
- Generator properly used via procgen.Generator interface

**Persistence Integration**: ⚠️ Partial
- Quest types stored in QuestTrackerComponent but lack Serialize/Deserialize methods
- Prevents quest state from being saved/loaded with player data
- Would cause loss of active quests on save/restore

**Missing Integrations**:
- No horror/cyberpunk genre template implementations
- No quest event hooks for narrative system integration (moral choice, faction conflict quest types exist but not used)

## Recommendations
1. **Move business logic to systems** — Create QuestProgressSystem to handle IsComplete/Progress calculations; make Quest/Objective pure data
2. **Add serialization support** — Implement Serialize/Deserialize on Quest, Objective, Reward types for persistence
3. **Complete genre support** — Add GetHorror*/GetCyberpunk* template functions to match advertised feature set
4. **Enhance validation** — Validate all reward fields, not just XP
5. **Add godoc to templates** — Document template functions with structure/usage examples
