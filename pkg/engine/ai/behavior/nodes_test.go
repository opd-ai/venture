package behavior_test

import (
	"fmt"
	"testing"

	"github.com/opd-ai/venture/pkg/engine/ai/behavior"
	"github.com/opd-ai/venture/pkg/engine/aitypes"
)

// ---------------------------------------------------------------------------
// Stub nodes used across tests
// ---------------------------------------------------------------------------

type stubNode struct {
	status    aitypes.NodeStatus
	tickCount int
	resetCount int
	name      string
}

func newStub(name string, status aitypes.NodeStatus) *stubNode {
	return &stubNode{name: name, status: status}
}

func (s *stubNode) Tick(_ aitypes.EntityContext, _ *aitypes.Blackboard, _ float64) aitypes.NodeStatus {
	s.tickCount++
	return s.status
}
func (s *stubNode) Reset() { s.resetCount++ }
func (s *stubNode) String() string { return fmt.Sprintf("stub(%s)", s.name) }

// ---------------------------------------------------------------------------
// SequenceNode
// ---------------------------------------------------------------------------

func TestSequenceNode_AllSuccess(t *testing.T) {
	a := newStub("a", aitypes.NodeSuccess)
	b := newStub("b", aitypes.NodeSuccess)
	seq := behavior.NewSequenceNode("seq", a, b)
	bb := aitypes.NewBlackboard()
	status := seq.Tick(nil, bb, 0.1)
	if status != aitypes.NodeSuccess {
		t.Fatalf("all-success sequence: got %v, want Success", status)
	}
	if a.tickCount != 1 || b.tickCount != 1 {
		t.Fatalf("expected each child ticked once, got a=%d b=%d", a.tickCount, b.tickCount)
	}
}

func TestSequenceNode_FirstFailure_StopsEarly(t *testing.T) {
	a := newStub("a", aitypes.NodeFailure)
	b := newStub("b", aitypes.NodeSuccess)
	seq := behavior.NewSequenceNode("seq", a, b)
	bb := aitypes.NewBlackboard()
	status := seq.Tick(nil, bb, 0.1)
	if status != aitypes.NodeFailure {
		t.Fatalf("first-fail sequence: got %v, want Failure", status)
	}
	// Second child must NOT be ticked.
	if b.tickCount != 0 {
		t.Fatalf("second child ticked despite first failing: b.tickCount=%d", b.tickCount)
	}
}

func TestSequenceNode_Running(t *testing.T) {
	a := newStub("a", aitypes.NodeRunning)
	b := newStub("b", aitypes.NodeSuccess)
	seq := behavior.NewSequenceNode("seq", a, b)
	bb := aitypes.NewBlackboard()
	status := seq.Tick(nil, bb, 0.1)
	if status != aitypes.NodeRunning {
		t.Fatalf("running sequence: got %v, want Running", status)
	}
}

func TestSequenceNode_Empty(t *testing.T) {
	seq := behavior.NewSequenceNode("empty")
	bb := aitypes.NewBlackboard()
	status := seq.Tick(nil, bb, 0.1)
	if status != aitypes.NodeSuccess {
		t.Fatalf("empty sequence: got %v, want Success", status)
	}
}

func TestSequenceNode_Reset(t *testing.T) {
	a := newStub("a", aitypes.NodeSuccess)
	seq := behavior.NewSequenceNode("seq", a)
	bb := aitypes.NewBlackboard()
	seq.Tick(nil, bb, 0.1)
	seq.Reset()
	// a.resetCount should be ≥1 (may also be called during resetOnFinish)
	if a.resetCount == 0 {
		t.Fatal("Reset did not propagate to child")
	}
}

// ---------------------------------------------------------------------------
// SelectorNode
// ---------------------------------------------------------------------------

func TestSelectorNode_FirstSuccess(t *testing.T) {
	a := newStub("a", aitypes.NodeSuccess)
	b := newStub("b", aitypes.NodeSuccess)
	sel := behavior.NewSelectorNode("sel", a, b)
	bb := aitypes.NewBlackboard()
	status := sel.Tick(nil, bb, 0.1)
	if status != aitypes.NodeSuccess {
		t.Fatalf("selector first-success: got %v, want Success", status)
	}
	// Second child must NOT be ticked.
	if b.tickCount != 0 {
		t.Fatalf("selector ticked second child despite first succeeding")
	}
}

func TestSelectorNode_AllFailure(t *testing.T) {
	a := newStub("a", aitypes.NodeFailure)
	b := newStub("b", aitypes.NodeFailure)
	sel := behavior.NewSelectorNode("sel", a, b)
	bb := aitypes.NewBlackboard()
	status := sel.Tick(nil, bb, 0.1)
	if status != aitypes.NodeFailure {
		t.Fatalf("all-fail selector: got %v, want Failure", status)
	}
}

func TestSelectorNode_Empty(t *testing.T) {
	sel := behavior.NewSelectorNode("empty")
	bb := aitypes.NewBlackboard()
	status := sel.Tick(nil, bb, 0.1)
	if status != aitypes.NodeFailure {
		t.Fatalf("empty selector: got %v, want Failure", status)
	}
}

// ---------------------------------------------------------------------------
// ParallelNode
// ---------------------------------------------------------------------------

func TestParallelNode_AllSuccess(t *testing.T) {
	a := newStub("a", aitypes.NodeSuccess)
	b := newStub("b", aitypes.NodeSuccess)
	par := behavior.NewParallelNode("par", a, b)
	bb := aitypes.NewBlackboard()
	status := par.Tick(nil, bb, 0.1)
	if status != aitypes.NodeSuccess {
		t.Fatalf("parallel all-success: got %v, want Success", status)
	}
	if a.tickCount != 1 || b.tickCount != 1 {
		t.Fatalf("parallel must tick all children; got a=%d b=%d", a.tickCount, b.tickCount)
	}
}

func TestParallelNode_AnyFailure_ReturnsFailure(t *testing.T) {
	a := newStub("a", aitypes.NodeSuccess)
	b := newStub("b", aitypes.NodeFailure)
	par := behavior.NewParallelNode("par", a, b)
	bb := aitypes.NewBlackboard()
	status := par.Tick(nil, bb, 0.1)
	if status != aitypes.NodeFailure {
		t.Fatalf("parallel any-failure: got %v, want Failure", status)
	}
	// Both children must still be ticked.
	if a.tickCount != 1 || b.tickCount != 1 {
		t.Fatalf("parallel must tick all children even on failure; got a=%d b=%d", a.tickCount, b.tickCount)
	}
}

func TestParallelNode_AnyRunning_ReturnsRunning(t *testing.T) {
	a := newStub("a", aitypes.NodeSuccess)
	b := newStub("b", aitypes.NodeRunning)
	par := behavior.NewParallelNode("par", a, b)
	bb := aitypes.NewBlackboard()
	status := par.Tick(nil, bb, 0.1)
	if status != aitypes.NodeRunning {
		t.Fatalf("parallel any-running: got %v, want Running", status)
	}
}

func TestParallelNode_Empty(t *testing.T) {
	par := behavior.NewParallelNode("empty")
	bb := aitypes.NewBlackboard()
	status := par.Tick(nil, bb, 0.1)
	if status != aitypes.NodeSuccess {
		t.Fatalf("empty parallel: got %v, want Success", status)
	}
}

// ---------------------------------------------------------------------------
// InverterNode
// ---------------------------------------------------------------------------

func TestInverterNode_InvertsSuccess(t *testing.T) {
	child := newStub("c", aitypes.NodeSuccess)
	inv := behavior.NewInverterNode("inv", child)
	bb := aitypes.NewBlackboard()
	status := inv.Tick(nil, bb, 0.1)
	if status != aitypes.NodeFailure {
		t.Fatalf("inverter success→failure: got %v", status)
	}
}

func TestInverterNode_InvertsFailure(t *testing.T) {
	child := newStub("c", aitypes.NodeFailure)
	inv := behavior.NewInverterNode("inv", child)
	bb := aitypes.NewBlackboard()
	status := inv.Tick(nil, bb, 0.1)
	if status != aitypes.NodeSuccess {
		t.Fatalf("inverter failure→success: got %v", status)
	}
}

func TestInverterNode_PassthroughRunning(t *testing.T) {
	child := newStub("c", aitypes.NodeRunning)
	inv := behavior.NewInverterNode("inv", child)
	bb := aitypes.NewBlackboard()
	status := inv.Tick(nil, bb, 0.1)
	if status != aitypes.NodeRunning {
		t.Fatalf("inverter running passthrough: got %v", status)
	}
}

func TestInverterNode_Reset(t *testing.T) {
	child := newStub("c", aitypes.NodeSuccess)
	inv := behavior.NewInverterNode("inv", child)
	inv.Reset()
	if child.resetCount != 1 {
		t.Fatalf("inverter Reset did not propagate to child: got %d", child.resetCount)
	}
}

// ---------------------------------------------------------------------------
// RepeatNode
// ---------------------------------------------------------------------------

func TestRepeatNode_RepeatsUntilLimit(t *testing.T) {
	const max = 3
	child := newStub("c", aitypes.NodeSuccess)
	rep := behavior.NewRepeatNode("rep", max, child)
	bb := aitypes.NewBlackboard()

	// Should return Running until max is reached.
	for i := 0; i < max-1; i++ {
		status := rep.Tick(nil, bb, 0.1)
		if status != aitypes.NodeRunning {
			t.Fatalf("repeat tick %d: got %v, want Running", i+1, status)
		}
	}
	// Final tick returns Success and resets.
	status := rep.Tick(nil, bb, 0.1)
	if status != aitypes.NodeSuccess {
		t.Fatalf("repeat final tick: got %v, want Success", status)
	}
}

func TestRepeatNode_ChildFailureStops(t *testing.T) {
	child := newStub("c", aitypes.NodeFailure)
	rep := behavior.NewRepeatNode("rep", 5, child)
	bb := aitypes.NewBlackboard()
	status := rep.Tick(nil, bb, 0.1)
	if status != aitypes.NodeFailure {
		t.Fatalf("repeat child-failure: got %v, want Failure", status)
	}
	if child.tickCount != 1 {
		t.Fatalf("repeat must stop on child failure; child.tickCount=%d", child.tickCount)
	}
}

func TestRepeatNode_InfiniteNeverCompletes(t *testing.T) {
	child := newStub("c", aitypes.NodeSuccess)
	rep := behavior.NewRepeatNode("rep", -1, child)
	bb := aitypes.NewBlackboard()
	for i := 0; i < 100; i++ {
		status := rep.Tick(nil, bb, 0.1)
		if status == aitypes.NodeSuccess || status == aitypes.NodeFailure {
			t.Fatalf("infinite repeat returned terminal status %v at tick %d", status, i+1)
		}
	}
}

func TestRepeatNode_Reset(t *testing.T) {
	child := newStub("c", aitypes.NodeSuccess)
	rep := behavior.NewRepeatNode("rep", 3, child)
	bb := aitypes.NewBlackboard()
	rep.Tick(nil, bb, 0.1)
	rep.Reset()
	if child.resetCount == 0 {
		t.Fatal("repeat Reset did not propagate to child")
	}
}

// ---------------------------------------------------------------------------
// String helpers
// ---------------------------------------------------------------------------

func TestNodeString(t *testing.T) {
	bb := aitypes.NewBlackboard()
	_ = bb // suppress unused warning for any future use

	nodes := []behavior.BehaviorNode{
		behavior.NewSequenceNode("s"),
		behavior.NewSelectorNode("sel"),
		behavior.NewParallelNode("p"),
		behavior.NewInverterNode("i", newStub("c", aitypes.NodeSuccess)),
		behavior.NewRepeatNode("r", 2, newStub("c", aitypes.NodeSuccess)),
	}
	for _, n := range nodes {
		if n.String() == "" {
			t.Errorf("node %T returned empty String()", n)
		}
	}
}
