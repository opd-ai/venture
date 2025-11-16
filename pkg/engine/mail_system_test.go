package engine

import (
	"fmt"
	"strings"
	"testing"
)

func getMailComponent(t *testing.T, entity *Entity) *MailComponent {
	t.Helper()
	comp, ok := entity.GetComponent("mail")
	if !ok {
		t.Fatal("Entity has no mail component")
	}
	return comp.(*MailComponent)
}

func TestNewMailSystem(t *testing.T) {
	world := NewWorld()
	ms := NewMailSystem(world)
	if ms == nil {
		t.Fatal("Expected mail system, got nil")
	}
	if ms.world != world {
		t.Error("Expected world to be set")
	}
	if ms.deliveryTime != 300.0 {
		t.Errorf("Expected delivery time 300.0, got %f", ms.deliveryTime)
	}
}

func TestCalculatePostage(t *testing.T) {
	world := NewWorld()
	ms := NewMailSystem(world)

	tests := []struct {
		hops     int
		expected int
	}{
		{0, 10},
		{1, 11},
		{5, 15},
		{10, 20},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.hops)), func(t *testing.T) {
			postage := ms.CalculatePostage(tt.hops)
			if postage != tt.expected {
				t.Errorf("Expected postage %d for %d hops, got %d", tt.expected, tt.hops, postage)
			}
		})
	}
}

func TestSendMail_SubjectTooLong(t *testing.T) {
	world := NewWorld()
	ms := NewMailSystem(world)

	subject := strings.Repeat("a", 51)
	_, err := ms.SendMail("1", "2", subject, "body", nil, "server1", "server2")
	if err == nil {
		t.Error("Expected error for subject > 50 chars")
	}
	if !strings.Contains(err.Error(), "subject exceeds") {
		t.Errorf("Expected subject error, got: %v", err)
	}
}

func TestSendMail_BodyTooLong(t *testing.T) {
	world := NewWorld()
	ms := NewMailSystem(world)

	body := strings.Repeat("a", 501)
	_, err := ms.SendMail("1", "2", "subject", body, nil, "server1", "server2")
	if err == nil {
		t.Error("Expected error for body > 500 chars")
	}
	if !strings.Contains(err.Error(), "body exceeds") {
		t.Errorf("Expected body error, got: %v", err)
	}
}

func TestSendMail_TooManyAttachments(t *testing.T) {
	world := NewWorld()
	ms := NewMailSystem(world)

	attachments := []uint64{1, 2, 3, 4, 5, 6}
	_, err := ms.SendMail("1", "2", "subject", "body", attachments, "server1", "server2")
	if err == nil {
		t.Error("Expected error for > 5 attachments")
	}
	if !strings.Contains(err.Error(), "too many attachments") {
		t.Errorf("Expected attachments error, got: %v", err)
	}
}

func TestSendMail_SenderNotFound(t *testing.T) {
	world := NewWorld()
	ms := NewMailSystem(world)

	_, err := ms.SendMail("999", "2", "subject", "body", nil, "server1", "server2")
	if err == nil {
		t.Error("Expected error for nonexistent sender")
	}
	if !strings.Contains(err.Error(), "sender entity not found") {
		t.Errorf("Expected sender error, got: %v", err)
	}
}

func TestSendMail_Success_SameServer(t *testing.T) {
	world := NewWorld()
	ms := NewMailSystem(world)

	sender := world.CreateEntity()
	sender.AddComponent(NewMailComponent())

	recipient := world.CreateEntity()
	recipient.AddComponent(NewMailComponent())

	world.Update(0.0)

	senderID := fmt.Sprintf("%d", sender.ID)
	recipientID := fmt.Sprintf("%d", recipient.ID)

	msg, err := ms.SendMail(senderID, recipientID, "Test", "Hello", nil, "server1", "server1")
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if msg == nil {
		t.Fatal("Expected message, got nil")
	}
	if msg.Subject != "Test" {
		t.Errorf("Expected subject 'Test', got '%s'", msg.Subject)
	}
	if msg.Body != "Hello" {
		t.Errorf("Expected body 'Hello', got '%s'", msg.Body)
	}
	if msg.Postage != 10 {
		t.Errorf("Expected postage 10 (0 hops), got %d", msg.Postage)
	}

	recipientMail := getMailComponent(t, recipient)
	if len(recipientMail.Inbox) != 1 {
		t.Errorf("Expected 1 message in recipient inbox, got %d", len(recipientMail.Inbox))
	}
	if recipientMail.Inbox[0].DeliveredAt == 0 {
		t.Error("Expected message to be delivered immediately on same server")
	}
}

func TestSendMail_Success_DifferentServer(t *testing.T) {
	world := NewWorld()
	ms := NewMailSystem(world)
	ms.SetDeliveryTime(1.0)

	sender := world.CreateEntity()
	sender.AddComponent(NewMailComponent())

	recipient := world.CreateEntity()
	recipient.AddComponent(NewMailComponent())
	world.Update(0.0)

	senderID := fmt.Sprintf("%d", sender.ID)
	recipientID := fmt.Sprintf("%d", recipient.ID)

	msg, err := ms.SendMail(senderID, recipientID, "Test", "Hello", nil, "server1", "server2")
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if msg.Postage != 11 {
		t.Errorf("Expected postage 11 (1 hop), got %d", msg.Postage)
	}

	courier := ms.GetCourierPosition(msg.ID)
	if courier == nil {
		t.Fatal("Expected courier position, got nil")
	}
	if courier.TotalHops != 1 {
		t.Errorf("Expected 1 hop, got %d", courier.TotalHops)
	}

	recipientMail := getMailComponent(t, recipient)
	if len(recipientMail.Inbox) != 0 {
		t.Error("Expected message not yet delivered")
	}
}

func TestMailSystemUpdate_Delivery(t *testing.T) {
	world := NewWorld()
	ms := NewMailSystem(world)
	ms.SetDeliveryTime(1.0)

	sender := world.CreateEntity()
	sender.AddComponent(NewMailComponent())

	recipient := world.CreateEntity()
	recipient.AddComponent(NewMailComponent())
	world.Update(0.0)

	senderID := fmt.Sprintf("%d", sender.ID)
	recipientID := fmt.Sprintf("%d", recipient.ID)

	msg, _ := ms.SendMail(senderID, recipientID, "Test", "Hello", nil, "server1", "server2")

	ms.Update(0.5)
	recipientMail := getMailComponent(t, recipient)
	if len(recipientMail.Inbox) != 0 {
		t.Error("Expected message not yet delivered after 0.5s")
	}

	ms.Update(0.6)
	if len(recipientMail.Inbox) != 1 {
		t.Errorf("Expected message delivered after 1.1s, got %d messages", len(recipientMail.Inbox))
	}
	if recipientMail.Inbox[0].DeliveredAt == 0 {
		t.Error("Expected DeliveredAt to be set")
	}

	if ms.GetCourierPosition(msg.ID) != nil {
		t.Error("Expected courier position to be cleared after delivery")
	}
}

func TestSetServerHopsFunc(t *testing.T) {
	world := NewWorld()
	ms := NewMailSystem(world)

	customHops := func(from, to string) int {
		if from == "A" && to == "C" {
			return 5
		}
		return 1
	}
	ms.SetServerHopsFunc(customHops)

	sender := world.CreateEntity()
	sender.AddComponent(NewMailComponent())

	recipient := world.CreateEntity()
	recipient.AddComponent(NewMailComponent())
	world.Update(0.0)

	senderID := fmt.Sprintf("%d", sender.ID)
	recipientID := fmt.Sprintf("%d", recipient.ID)

	msg, _ := ms.SendMail(senderID, recipientID, "Test", "Hello", nil, "A", "C")
	if msg.Postage != 15 {
		t.Errorf("Expected postage 15 (5 hops), got %d", msg.Postage)
	}

	courier := ms.GetCourierPosition(msg.ID)
	if courier.TotalHops != 5 {
		t.Errorf("Expected 5 hops, got %d", courier.TotalHops)
	}
}

func TestDeliverMail_InboxFull(t *testing.T) {
	world := NewWorld()
	ms := NewMailSystem(world)

	recipient := world.CreateEntity()
	mailComp := NewMailComponent()
	mailComp.MaxInbox = 1
	mailComp.AddToInbox(&MailMessage{ID: "existing"})
	recipient.AddComponent(mailComp)
	world.Update(0.0)

	msg := &MailMessage{
		ID:          "new",
		RecipientID: fmt.Sprintf("%d", recipient.ID),
	}

	err := ms.DeliverMail(msg, fmt.Sprintf("%d", recipient.ID))
	if err == nil {
		t.Error("Expected error for full inbox")
	}
	if !strings.Contains(err.Error(), "inbox full") {
		t.Errorf("Expected inbox full error, got: %v", err)
	}
}

func TestGenerateMailMessageID(t *testing.T) {
	id1 := generateMailMessageID()
	id2 := generateMailMessageID()

	if len(id1) != 32 {
		t.Errorf("Expected message ID length 32, got %d", len(id1))
	}
	if id1 == id2 {
		t.Error("Expected unique message IDs")
	}
}

func BenchmarkSendMail(b *testing.B) {
	world := NewWorld()
	ms := NewMailSystem(world)

	sender := world.CreateEntity()
	sender.AddComponent(NewMailComponent())

	recipient := world.CreateEntity()
	recipient.AddComponent(NewMailComponent())

	senderID := fmt.Sprintf("%d", sender.ID)
	recipientID := fmt.Sprintf("%d", recipient.ID)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.SendMail(senderID, recipientID, "Test", "Hello", nil, "server1", "server1")
	}
}

func BenchmarkMailSystemUpdate(b *testing.B) {
	world := NewWorld()
	ms := NewMailSystem(world)
	ms.SetDeliveryTime(10.0)

	for i := 0; i < 100; i++ {
		sender := world.CreateEntity()
		sender.AddComponent(NewMailComponent())
		recipient := world.CreateEntity()
		recipient.AddComponent(NewMailComponent())
		ms.SendMail(fmt.Sprintf("%d", sender.ID), fmt.Sprintf("%d", recipient.ID), "Test", "Hello", nil, "server1", "server2")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ms.Update(0.016)
	}
}
