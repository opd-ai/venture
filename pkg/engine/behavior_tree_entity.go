// Package engine provides helpers for behavior tree entity context handling.
package engine

import "github.com/opd-ai/venture/pkg/engine/aitypes"

// entityFromContext converts an aitypes.EntityContext to the concrete *Entity type.
// All leaf-node Tick implementations use this instead of repeating the two-line
// assertion pattern inline. Returns (entity, true) on success, (nil, false) when
// ctx is nil or not a *Entity, in which case the caller should return NodeFailure.
func entityFromContext(ctx aitypes.EntityContext) (*Entity, bool) {
	e, ok := ctx.(*Entity)
	return e, ok
}
