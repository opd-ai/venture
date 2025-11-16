package engine

import (
	"testing"
)

func TestNewMailComponent(t *testing.T) {
	mc := NewMailComponent()
	if mc == nil {
		t.Fatal("Expected mail component, got nil")
	}
	if mc.MaxInbox != 50 {
		t.Errorf("Expected MaxInbox=50, got %d", mc.MaxInbox)
	}
	if len(mc.Inbox) != 0 {
		t.Errorf("Expected empty inbox, got %d messages", len(mc.Inbox))
	}
	if len(mc.Outbox) != 0 {
		t.Errorf("Expected empty outbox, got %d messages", len(mc.Outbox))
	}
}

func TestMailComponentType(t *testing.T) {
	mc := NewMailComponent()
	if mc.Type() != "mail" {
		t.Errorf("Expected type 'mail', got '%s'", mc.Type())
	}
}

func TestMailMessageGetStatus(t *testing.T) {
	tests := []struct {
		name        string
		sentAt      int64
		deliveredAt int64
		expected    MailStatus
	}{
		{"Sent", 0, 0, MailStatusSent},
		{"InTransit", 123456, 0, MailStatusInTransit},
		{"Delivered", 123456, 123789, MailStatusDelivered},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := &MailMessage{
				SentAt:      tt.sentAt,
				DeliveredAt: tt.deliveredAt,
			}
			status := msg.GetStatus()
			if status != tt.expected {
				t.Errorf("Expected status %s, got %s", tt.expected, status)
			}
		})
	}
}

func TestMailStatusString(t *testing.T) {
	tests := []struct {
		status   MailStatus
		expected string
	}{
		{MailStatusSent, "Sent"},
		{MailStatusInTransit, "In Transit"},
		{MailStatusDelivered, "Delivered"},
		{MailStatusFailed, "Failed"},
		{MailStatus(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			str := tt.status.String()
			if str != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, str)
			}
		})
	}
}

func TestAddToInbox(t *testing.T) {
	mc := NewMailComponent()
	msg := &MailMessage{
		ID:          "msg1",
		SenderID:    "player1",
		RecipientID: "player2",
		Subject:     "Test",
		Body:        "Hello",
		SentAt:      123456,
	}

	if !mc.AddToInbox(msg) {
		t.Error("Expected AddToInbox to succeed")
	}
	if len(mc.Inbox) != 1 {
		t.Errorf("Expected 1 message in inbox, got %d", len(mc.Inbox))
	}
	if mc.Inbox[0].ID != "msg1" {
		t.Errorf("Expected message ID 'msg1', got '%s'", mc.Inbox[0].ID)
	}
}

func TestAddToInbox_MaxCapacity(t *testing.T) {
	mc := NewMailComponent()
	mc.MaxInbox = 2

	msg1 := &MailMessage{ID: "msg1", SenderID: "p1", RecipientID: "p2"}
	msg2 := &MailMessage{ID: "msg2", SenderID: "p1", RecipientID: "p2"}
	msg3 := &MailMessage{ID: "msg3", SenderID: "p1", RecipientID: "p2"}

	if !mc.AddToInbox(msg1) {
		t.Error("Expected first message to succeed")
	}
	if !mc.AddToInbox(msg2) {
		t.Error("Expected second message to succeed")
	}
	if mc.AddToInbox(msg3) {
		t.Error("Expected third message to fail (inbox full)")
	}
	if len(mc.Inbox) != 2 {
		t.Errorf("Expected 2 messages in inbox, got %d", len(mc.Inbox))
	}
}

func TestAddToOutbox(t *testing.T) {
	mc := NewMailComponent()
	msg := &MailMessage{
		ID:          "msg1",
		SenderID:    "player1",
		RecipientID: "player2",
		Subject:     "Test",
		Body:        "Hello",
		SentAt:      123456,
	}

	mc.AddToOutbox(msg)
	if len(mc.Outbox) != 1 {
		t.Errorf("Expected 1 message in outbox, got %d", len(mc.Outbox))
	}
	if mc.Outbox[0].ID != "msg1" {
		t.Errorf("Expected message ID 'msg1', got '%s'", mc.Outbox[0].ID)
	}
}

func TestRemoveFromInbox(t *testing.T) {
	mc := NewMailComponent()
	msg1 := &MailMessage{ID: "msg1", SenderID: "p1", RecipientID: "p2"}
	msg2 := &MailMessage{ID: "msg2", SenderID: "p1", RecipientID: "p2"}

	mc.AddToInbox(msg1)
	mc.AddToInbox(msg2)

	if !mc.RemoveFromInbox("msg1") {
		t.Error("Expected RemoveFromInbox to succeed")
	}
	if len(mc.Inbox) != 1 {
		t.Errorf("Expected 1 message in inbox, got %d", len(mc.Inbox))
	}
	if mc.Inbox[0].ID != "msg2" {
		t.Errorf("Expected remaining message ID 'msg2', got '%s'", mc.Inbox[0].ID)
	}
}

func TestRemoveFromInbox_NotFound(t *testing.T) {
	mc := NewMailComponent()
	if mc.RemoveFromInbox("nonexistent") {
		t.Error("Expected RemoveFromInbox to fail for nonexistent message")
	}
}

func TestRemoveFromOutbox(t *testing.T) {
	mc := NewMailComponent()
	msg1 := &MailMessage{ID: "msg1", SenderID: "p1", RecipientID: "p2"}
	msg2 := &MailMessage{ID: "msg2", SenderID: "p1", RecipientID: "p2"}

	mc.AddToOutbox(msg1)
	mc.AddToOutbox(msg2)

	if !mc.RemoveFromOutbox("msg1") {
		t.Error("Expected RemoveFromOutbox to succeed")
	}
	if len(mc.Outbox) != 1 {
		t.Errorf("Expected 1 message in outbox, got %d", len(mc.Outbox))
	}
	if mc.Outbox[0].ID != "msg2" {
		t.Errorf("Expected remaining message ID 'msg2', got '%s'", mc.Outbox[0].ID)
	}
}

func TestGetUnreadCount(t *testing.T) {
	mc := NewMailComponent()

	now := int64(90000)
	msg1 := &MailMessage{ID: "msg1", DeliveredAt: now}
	msg2 := &MailMessage{ID: "msg2", DeliveredAt: now - 48*3600}
	msg3 := &MailMessage{ID: "msg3", DeliveredAt: 0}

	mc.AddToInbox(msg1)
	mc.AddToInbox(msg2)
	mc.AddToInbox(msg3)

	unread := mc.GetUnreadCount()
	if unread != 1 {
		t.Errorf("Expected 1 unread message (delivered in last 24h from first message), got %d", unread)
	}
}

func TestNewPostOfficeComponent(t *testing.T) {
	po := NewPostOfficeComponent("Bob")
	if po == nil {
		t.Fatal("Expected post office component, got nil")
	}
	if po.ClerkName != "Bob" {
		t.Errorf("Expected clerk name 'Bob', got '%s'", po.ClerkName)
	}
	if po.ServiceFee != 10 {
		t.Errorf("Expected service fee 10, got %d", po.ServiceFee)
	}
	if po.MaxDistance != 100 {
		t.Errorf("Expected max distance 100, got %d", po.MaxDistance)
	}
}

func TestPostOfficeComponentType(t *testing.T) {
	po := NewPostOfficeComponent("Alice")
	if po.Type() != "postoffice" {
		t.Errorf("Expected type 'postoffice', got '%s'", po.Type())
	}
}

func BenchmarkMailComponentAddToInbox(b *testing.B) {
	mc := NewMailComponent()
	mc.MaxInbox = 1000
	msg := &MailMessage{ID: "msg1", SenderID: "p1", RecipientID: "p2"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.Inbox = mc.Inbox[:0]
		mc.AddToInbox(msg)
	}
}

func BenchmarkMailComponentRemoveFromInbox(b *testing.B) {
	mc := NewMailComponent()
	for i := 0; i < 50; i++ {
		mc.AddToInbox(&MailMessage{ID: "msg", SenderID: "p1", RecipientID: "p2"})
	}
	mc.AddToInbox(&MailMessage{ID: "target", SenderID: "p1", RecipientID: "p2"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mc.RemoveFromInbox("target")
		mc.AddToInbox(&MailMessage{ID: "target", SenderID: "p1", RecipientID: "p2"})
	}
}

func BenchmarkGetUnreadCount(b *testing.B) {
	mc := NewMailComponent()
	now := int64(100000)
	for i := 0; i < 50; i++ {
		mc.AddToInbox(&MailMessage{ID: "msg", DeliveredAt: now})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mc.GetUnreadCount()
	}
}
