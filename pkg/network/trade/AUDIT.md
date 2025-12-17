# Trade Package Audit Report

**Audit Date:** 2025-12-17  
**Package:** `github.com/opd-ai/venture/pkg/network/trade`  
**Auditor:** Automated Code Audit  
**Test Coverage:** 69.4% (all tests passing)

## AUDIT SUMMARY

| Category | Count |
|----------|-------|
| CRITICAL BUG | 2 (2 RESOLVED) |
| FUNCTIONAL MISMATCH | 2 (1 RESOLVED) |
| MISSING FEATURE | 2 |
| EDGE CASE BUG | 2 |
| PERFORMANCE ISSUE | 0 |
| **Total Issues** | **8 (3 RESOLVED)** |

---

## DETAILED FINDINGS

### CRITICAL BUG: Rollback Mechanism Does Not Restore Items

**File:** system.go:370-435  
**Severity:** High  
**Status:** RESOLVED (2025-12-17, commit 01a4a10)  
**Description:** The `rollbackTrade` function attempted to restore items to their original owners by calling `resolveItems` on the current inventory. However, by the time rollback was called, items had already been removed from their original inventories.

**Resolution:** Replaced the broken rollback approach with inline rollback tracking in `executeItemTransfer`. The function now tracks all removed and added items during the transfer phases, and if any phase fails, it properly reverses all prior operations using the tracked item references.

---

### CRITICAL BUG: Failed Trade Acceptance Leaves Players Stuck

**File:** system.go:246-256  
**Severity:** High  
**Status:** RESOLVED (2025-12-17, commit 01a4a10)  
**Description:** When `AcceptTrade` failed during `commitTrade`, the trade status was set to "failed" but the trade was never cleared from either participant's TradeComponent.

**Resolution:** Added `clearTrade` calls in the failure path of `AcceptTrade` for both proposer and recipient. Also now stores the failure reason in `proposal.FailureReason` for client debugging.

---

### FUNCTIONAL MISMATCH: Lag Compensation Not Implemented

**File:** doc.go:40, system.go:494-514  
**Severity:** Medium  
**Description:** The documentation explicitly states "Distance validation uses lag compensation for multiplayer fairness", but the implementation performs simple Euclidean distance calculation with no lag compensation whatsoever.

**Expected Behavior:** Distance validation should account for network latency, allowing for some tolerance or using historical position data to compensate for lag in multiplayer scenarios.

**Actual Behavior:** The `validateProximity` function uses current position snapshots with no consideration for network latency, potentially causing unfair trade cancellations in high-latency environments.

**Impact:** In multiplayer scenarios with network latency (especially the 200-5000ms Tor connections the project supports), trades may be incorrectly cancelled due to stale position data.

**Reproduction:**
1. Set up two players at exactly 10 tiles apart (max trade distance)
2. Introduce network latency so position updates are delayed
3. Start a trade - may be incorrectly cancelled due to stale positions

**Code Reference:**
```go
func (s *TradeSystem) validateProximity(id1, id2 uint64, maxDist float64) bool {
    // No lag compensation - just uses current positions
    pos1 := s.getPosition(e1)
    pos2 := s.getPosition(e2)

    dx := pos1.X - pos2.X
    dy := pos1.Y - pos2.Y
    distance := math.Sqrt(dx*dx + dy*dy)

    return distance <= maxDist
}
```

---

### FUNCTIONAL MISMATCH: Cancellation Reason Not Stored

**File:** system.go:469-471  
**Severity:** Medium  
**Status:** RESOLVED (2025-12-17, commit 01a4a10)  
**Description:** The `cancelTrade` function accepted a `reason` parameter but never stored it in the trade proposal's `FailureReason` field.

**Resolution:** Added `proposal.FailureReason = string(reason)` in `cancelTrade` function to store the cancellation reason.

---

### MISSING FEATURE: Counter-Propose Not Implemented

**File:** doc.go:18  
**Severity:** Medium  
**Description:** The documentation states that during the Review phase, "Player B reviews the proposal (can accept, reject, or counter-propose)". However, there is no `CounterPropose` function or any mechanism for counter-proposals in the codebase.

**Expected Behavior:** A `CounterPropose` method should allow the recipient to modify the offered/requested items and send a modified proposal back to the original proposer.

**Actual Behavior:** Recipients can only accept or reject trades. There is no counter-propose functionality.

**Impact:** Reduced trading flexibility - players must completely reject trades and start new negotiations instead of modifying existing proposals.

**Reproduction:**
1. Search codebase for CounterPropose or similar functionality
2. Only find the documentation reference, no implementation

**Code Reference:**
```go
// doc.go:18
// 2. **Review**: Player B reviews the proposal (can accept, reject, or counter-propose)

// No CounterPropose method exists in system.go
```

---

### MISSING FEATURE: ReasonConcurrent and ReasonRarity Unused

**File:** system.go:55-57  
**Severity:** Low  
**Description:** Two TradeFailureReason constants are defined but never used in the codebase:
- `ReasonConcurrent` ("item in another trade") 
- `ReasonRarity` ("rarity exceeds trust level")

There is no validation to check if an item is involved in another concurrent trade, and rarity validation errors use inline error messages rather than the defined constant.

**Expected Behavior:** These constants should be used when their respective conditions are detected.

**Actual Behavior:** `ReasonConcurrent` is never used (no concurrent item check exists). `ReasonRarity` is defined but trust validation uses custom error messages instead.

**Impact:** Dead code and inconsistent error handling; potential for items to be offered in multiple simultaneous trades.

**Reproduction:**
1. grep for ReasonConcurrent usage - not found except definition
2. grep for ReasonRarity usage - not found except definition

**Code Reference:**
```go
const (
    // ...
    ReasonConcurrent TradeFailureReason = "item in another trade"  // Never used
    // ...
    ReasonRarity     TradeFailureReason = "rarity exceeds trust level"  // Never used
)
```

---

### EDGE CASE BUG: Inventory Space Validation Ignores Outgoing Items

**File:** system.go:610-629  
**Severity:** Medium  
**Description:** The `validateInventorySpace` function checks if the inventory can accommodate incoming items, but does not account for items that will be removed during the same trade. This causes valid trades to be incorrectly rejected.

**Expected Behavior:** When validating if proposer can receive requestedItems, the validation should account for offeredItems being removed (freeing slots and weight capacity).

**Actual Behavior:** Validation assumes items are only being added, not that outgoing items free up space. A player with 99/100 slots trading 10 items for 10 items would fail validation despite the trade being valid.

**Impact:** Valid trades are incorrectly rejected when inventory is near capacity, even though the trade would net-neutral or net-positive in terms of space.

**Reproduction:**
1. Create player with inventory: 95/100 slots, 900/1000 weight
2. Player offers 10 items (100 weight), requests 5 items (50 weight)
3. Trade is rejected due to "insufficient inventory slots" despite net gain of 5 slots

**Code Reference:**
```go
func (s *TradeSystem) validateInventorySpace(inventory *engine.InventoryComponent, items []*item.Item) error {
    // BUG: Does not subtract items being traded away
    if len(inventory.Items)+len(items) > inventory.MaxItems {
        return fmt.Errorf("insufficient inventory slots...")
    }
    // Same issue for weight validation
}
```

---

### EDGE CASE BUG: Duplicate Item IDs Allow Double-Trading Same Item

**File:** system.go:556-572  
**Severity:** Medium  
**Description:** The `resolveItems` function does not check for duplicate item IDs in the input list. If a malicious client sends the same item ID multiple times (e.g., `["sword1", "sword1"]`), the function returns the same item pointer twice. When `removeItemsFromInventory` then processes this list, the first removal succeeds but the second fails (item already removed), triggering an unnecessary rollback.

**Expected Behavior:** The function should detect and reject duplicate item IDs, or deduplicate the input.

**Actual Behavior:** Duplicate IDs result in the same `*item.Item` pointer appearing multiple times in the returned slice, causing the second removal attempt to fail.

**Impact:** Potential for failed trades and confusing error messages; could be exploited to trigger rollback scenarios.

**Reproduction:**
1. Call ProposeTrade with offeredItemIDs containing duplicates: `["item1", "item1"]`
2. Trade proposal succeeds (validation passes)
3. AcceptTrade attempts to remove "item1" twice
4. Second removal fails, triggering (broken) rollback

**Code Reference:**
```go
func (s *TradeSystem) resolveItems(inventory *engine.InventoryComponent, itemIDs []string) ([]*item.Item, error) {
    var items []*item.Item
    for _, id := range itemIDs {
        // No duplicate check - same ID can be added multiple times
        for _, itm := range inventory.Items {
            if itm.ID == id {
                items = append(items, itm)  // Same pointer added twice for duplicate IDs
                break
            }
        }
    }
    return items, nil
}
```

---

## RECOMMENDATIONS

1. **Priority 1 (Critical):** Fix the rollback mechanism to track removed items in a separate slice that persists through the transaction, or implement proper two-phase commit with prepare/commit phases.

2. **Priority 2 (Critical):** Add `clearTrade` calls in the failure path of `AcceptTrade` to prevent players from being stuck.

3. **Priority 3 (Medium):** Store the failure reason in `proposal.FailureReason` within `cancelTrade`.

4. **Priority 4 (Medium):** Either implement lag compensation as documented, or update documentation to reflect current behavior.

5. **Priority 5 (Medium):** Fix inventory space validation to account for outgoing items when calculating available capacity.

6. **Priority 6 (Medium):** Add duplicate item ID detection in `resolveItems` or `ProposeTrade`.

7. **Priority 7 (Low):** Either implement counter-propose functionality or update documentation to remove the reference.

8. **Priority 8 (Low):** Remove or implement the unused `ReasonConcurrent` and `ReasonRarity` constants.

---

## FILES ANALYZED

| File | Lines | Purpose |
|------|-------|---------|
| doc.go | 95 | Package documentation |
| system.go | 691 | Core trade system implementation |
| system_test.go | 645 | Unit and benchmark tests |

## DEPENDENCY ANALYSIS

**Level 0 (No internal imports):** None  
**Level 1 (External only):**
- system.go imports: `fmt`, `math`, `time`, `engine`, `procgen/item`
- doc.go: No imports

The package has minimal internal dependencies and clear external interfaces.
