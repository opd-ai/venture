// Package aitypes defines minimal shared types for the behavior tree AI subsystem.
// It contains only stdlib dependencies so it can be imported by both pkg/engine
// and pkg/engine/ai/behavior without creating circular imports.
package aitypes

// NodeStatus represents the result of a behavior tree node execution.
type NodeStatus int

const (
	// NodeSuccess indicates the node completed successfully.
	NodeSuccess NodeStatus = iota
	// NodeFailure indicates the node failed to complete.
	NodeFailure
	// NodeRunning indicates the node is still executing.
	NodeRunning
)

// String returns the string representation of a node status.
func (s NodeStatus) String() string {
	switch s {
	case NodeSuccess:
		return "Success"
	case NodeFailure:
		return "Failure"
	case NodeRunning:
		return "Running"
	default:
		return "Unknown"
	}
}
