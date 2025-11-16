package engine

import (
	"testing"
)

func TestCourierComponent_Type(t *testing.T) {
	c := &CourierComponent{}
	if c.Type() != "courier" {
		t.Errorf("Expected type 'courier', got '%s'", c.Type())
	}
}

func TestCourierComponent_IsCarryingMail(t *testing.T) {
	tests := []struct {
		name      string
		messageID string
		want      bool
	}{
		{"empty message ID", "", false},
		{"with message ID", "msg123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CourierComponent{CurrentMessageID: tt.messageID}
			if got := c.IsCarryingMail(); got != tt.want {
				t.Errorf("IsCarryingMail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCourierComponent_GetCurrentServer(t *testing.T) {
	tests := []struct {
		name     string
		route    []string
		progress int
		want     string
	}{
		{"no route", nil, 0, ""},
		{"progress negative", []string{"A", "B"}, -1, ""},
		{"progress beyond route", []string{"A", "B"}, 2, ""},
		{"at start", []string{"A", "B", "C"}, 0, "A"},
		{"mid route", []string{"A", "B", "C"}, 1, "B"},
		{"at end", []string{"A", "B", "C"}, 2, "C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CourierComponent{CurrentRoute: tt.route, RouteProgress: tt.progress}
			if got := c.GetCurrentServer(); got != tt.want {
				t.Errorf("GetCurrentServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCourierComponent_GetNextServer(t *testing.T) {
	tests := []struct {
		name     string
		route    []string
		progress int
		want     string
	}{
		{"at end", []string{"A", "B"}, 1, ""},
		{"beyond end", []string{"A", "B"}, 2, ""},
		{"at start", []string{"A", "B", "C"}, 0, "B"},
		{"mid route", []string{"A", "B", "C"}, 1, "C"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CourierComponent{CurrentRoute: tt.route, RouteProgress: tt.progress}
			if got := c.GetNextServer(); got != tt.want {
				t.Errorf("GetNextServer() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCourierComponent_HasReachedDestination(t *testing.T) {
	tests := []struct {
		name     string
		route    []string
		progress int
		want     bool
	}{
		{"not at destination", []string{"A", "B", "C"}, 0, false},
		{"mid route", []string{"A", "B", "C"}, 1, false},
		{"at destination", []string{"A", "B", "C"}, 2, true},
		{"beyond destination", []string{"A", "B"}, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CourierComponent{CurrentRoute: tt.route, RouteProgress: tt.progress}
			if got := c.HasReachedDestination(); got != tt.want {
				t.Errorf("HasReachedDestination() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCourierComponent_AssignDelivery(t *testing.T) {
	c := &CourierComponent{}
	messageID := "msg-12345"
	route := []string{"Server1", "Server2", "Server3"}

	c.AssignDelivery(messageID, route)

	if c.CurrentMessageID != messageID {
		t.Errorf("Expected CurrentMessageID=%s, got %s", messageID, c.CurrentMessageID)
	}
	if c.RouteProgress != 0 {
		t.Errorf("Expected RouteProgress=0, got %d", c.RouteProgress)
	}
	if len(c.CurrentRoute) != len(route) {
		t.Errorf("Expected route length=%d, got %d", len(route), len(c.CurrentRoute))
	}
	for i, server := range route {
		if c.CurrentRoute[i] != server {
			t.Errorf("Expected CurrentRoute[%d]=%s, got %s", i, server, c.CurrentRoute[i])
		}
	}

	// Verify route is a copy (modifying original shouldn't affect component)
	route[0] = "Modified"
	if c.CurrentRoute[0] == "Modified" {
		t.Error("CurrentRoute should be a copy, not a reference")
	}
}

func TestCourierComponent_AdvanceRoute(t *testing.T) {
	tests := []struct {
		name         string
		route        []string
		progress     int
		wantProgress int
		wantSuccess  bool
	}{
		{"advance from start", []string{"A", "B", "C"}, 0, 1, true},
		{"advance mid route", []string{"A", "B", "C"}, 1, 2, true},
		{"at destination", []string{"A", "B", "C"}, 2, 2, false},
		{"beyond destination", []string{"A", "B"}, 5, 5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CourierComponent{CurrentRoute: tt.route, RouteProgress: tt.progress}
			if got := c.AdvanceRoute(); got != tt.wantSuccess {
				t.Errorf("AdvanceRoute() = %v, want %v", got, tt.wantSuccess)
			}
			if c.RouteProgress != tt.wantProgress {
				t.Errorf("RouteProgress = %d, want %d", c.RouteProgress, tt.wantProgress)
			}
		})
	}
}

func TestCourierComponent_CompleteDelivery(t *testing.T) {
	c := &CourierComponent{
		CurrentMessageID: "msg-123",
		CurrentRoute:     []string{"A", "B", "C"},
		RouteProgress:    2,
	}

	c.CompleteDelivery()

	if c.CurrentMessageID != "" {
		t.Errorf("Expected CurrentMessageID to be empty, got %s", c.CurrentMessageID)
	}
	if c.CurrentRoute != nil {
		t.Errorf("Expected CurrentRoute to be nil, got %v", c.CurrentRoute)
	}
	if c.RouteProgress != 0 {
		t.Errorf("Expected RouteProgress=0, got %d", c.RouteProgress)
	}
}

func TestPostOfficeClerkComponent_Type(t *testing.T) {
	c := &PostOfficeClerkComponent{}
	if c.Type() != "postoffice_clerk" {
		t.Errorf("Expected type 'postoffice_clerk', got '%s'", c.Type())
	}
}

func TestPostOfficeClerkComponent_Fields(t *testing.T) {
	c := &PostOfficeClerkComponent{
		PostOfficeID:     12345,
		GreetingDialogue: "Welcome to the post office!",
		ServiceFee:       5,
	}

	if c.PostOfficeID != 12345 {
		t.Errorf("Expected PostOfficeID=12345, got %d", c.PostOfficeID)
	}
	if c.GreetingDialogue != "Welcome to the post office!" {
		t.Errorf("Expected GreetingDialogue='Welcome to the post office!', got '%s'", c.GreetingDialogue)
	}
	if c.ServiceFee != 5 {
		t.Errorf("Expected ServiceFee=5, got %d", c.ServiceFee)
	}
}

// Benchmark tests
func BenchmarkCourierComponent_AssignDelivery(b *testing.B) {
	c := &CourierComponent{}
	route := []string{"Server1", "Server2", "Server3", "Server4", "Server5"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.AssignDelivery("msg-12345", route)
	}
}

func BenchmarkCourierComponent_AdvanceRoute(b *testing.B) {
	c := &CourierComponent{
		CurrentRoute:  []string{"A", "B", "C", "D", "E"},
		RouteProgress: 0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.RouteProgress = i % (len(c.CurrentRoute) - 1)
		c.AdvanceRoute()
	}
}

func BenchmarkCourierComponent_HasReachedDestination(b *testing.B) {
	c := &CourierComponent{
		CurrentRoute:  []string{"A", "B", "C", "D", "E"},
		RouteProgress: 2,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.HasReachedDestination()
	}
}
