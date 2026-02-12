# Audit: github.com/opd-ai/venture/pkg/integration/trade_routes
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The trade_routes package implements AI-controlled merchant caravans for cross-server trading with route optimization, bandit encounters, escort missions, and economy integration. Code is well-structured with excellent test coverage (~90% estimated), comprehensive documentation, and proper ECS integration. Critical issue: complete absence of structured logging violates project standards for error visibility and debugging.

## Issues Found
- [ ] **severity:high** Error handling — No structured logging with logrus.WithFields on error paths; all errors return silently without logging context (affects debugging, operations monitoring, error correlation). Add logrus logging to CreateRoute, StartRoute, AddEscort, RemoveEscort, GetRoute, CreateEscortMission, CreateCaravan error paths with fields: routeID, playerID, caravanID, error message (`manager.go:165,169,173,216,224,280,299,306,325,353,394,398,690,695`)
- [ ] **severity:med** Test coverage — GetRoute error case not explicitly tested; only happy-path usage in TestUpdateRoutesProgress. Add table-driven test for GetRoute with valid/invalid routeID cases (`manager_test.go:missing`)
- [ ] **severity:low** Code style — Variable uses snake_case instead of Go camelCase convention (`manager.go:548`)

## Test Coverage
~90% (estimated from code analysis; cannot run tests in headless environment due to Ebiten dependency)

**Evidence:**
- Production code: 1087 lines (manager.go, types.go, constants.go)
- Test code: 1136 lines (manager_test.go, economy_integration_test.go)
- 32 test functions covering all major functionality
- 3 benchmarks (CreateRoute, OptimizeRoute, UpdateRoutes)
- Table-driven tests for CreateRoute, StartRoute, String() methods, encounters
- Lifecycle tests for goroutine management (Start/Stop idempotency, leak prevention)
- Economy integration tests (7 test cases for price handler)
- Determinism test validates seed-based reproducibility

**Tested methods:** NewRouteManager, CreateRoute, StartRoute, AddEscort, RemoveEscort, CreateEscortMission, CreateCaravan, GetActiveRoutes, GetRouteByCaravan, OptimizeRoute, UpdateRoutes, UpdateEncounter (all outcomes), SpawnBanditEncounter, SetPriceUpdateHandler, Start, Stop

**Untested edge cases:** GetRoute with invalid routeID (indirectly tested but no explicit error case test)

## Integration Status
**Fully integrated** across client and server:

- **Client integration** (`cmd/client/handlers.go`): RouteManager instantiated in system initialization
- **Server integration** (`cmd/server/v4_systems.go`): RouteManager part of V4 systems architecture
- **Integration tests** (`cmd/server/trade_routes_integration_test.go`): 5 additional integration tests verify cross-system behavior

**External dependencies:**
- `pkg/procgen/vehicle` — Caravan vehicle generation (CreateCaravan uses VehicleGenerator with TypeCart, RarityCommon)
- `pkg/world/economy` — PriceUpdateHandler interface for marketplace price impacts (completeRoute applies supply/demand effects)

**Federation support:** Cross-server routes supported via serverID in route creation (federation/market.go integration point documented but handler injection optional)

**Persistence:** No save/load implementation detected (routes appear to be runtime-only; may need persistence for long-running routes across server restarts)

## Recommendations
1. **[HIGH PRIORITY]** Add structured logging with logrus.WithFields to all error return paths and critical operations (route creation, completion, encounters). Use standard field names: "routeID", "playerID", "caravanID", "serverID", "error". Estimate: 30-45 minutes to add logging to ~15 error paths.

2. **[MEDIUM PRIORITY]** Add explicit error case test for GetRoute to achieve comprehensive method coverage. Create table-driven test with cases: valid routeID (existing test), invalid routeID (missing), empty routeID (missing). Estimate: 10 minutes.

3. **[LOW PRIORITY]** Rename `strength_ratio` variable to `strengthRatio` on manager.go:548 to follow Go naming conventions (gofmt passes but golangci-lint would flag this). Estimate: 1 minute.

4. **[FUTURE ENHANCEMENT]** Consider adding persistence support for active routes to survive server restarts (integrate with pkg/saveload). Routes with >5 minute travel times could be lost on restart. Document if this is intentional design.

5. **[FUTURE ENHANCEMENT]** Add TestGetRoute explicit function (currently only used indirectly in TestUpdateRoutesProgress line 400). While adequate for coverage, explicit tests improve maintainability and documentation.
