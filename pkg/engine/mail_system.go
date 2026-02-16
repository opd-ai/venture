package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// MailSystem manages sending, receiving, and delivery of mail messages.
type MailSystem struct {
	world            *World
	serverHops       func(from, to string) int
	deliveryTime     float64
	courierPositions map[string]*CourierPosition
}

// CourierPosition tracks the position of a courier NPC delivering mail.
type CourierPosition struct {
	MessageID        string
	CurrentServer    string
	TargetServer     string
	Progress         float64
	TotalHops        int
	EstimatedArrival int64
}

// NewMailSystem creates a new mail system.
func NewMailSystem(world *World) *MailSystem {
	return &MailSystem{
		world:            world,
		serverHops:       defaultServerHops,
		deliveryTime:     300.0,
		courierPositions: make(map[string]*CourierPosition),
	}
}

// defaultServerHops returns a default hop count between servers.
func defaultServerHops(from, to string) int {
	if from == to {
		return 0
	}
	return 1
}

// SetServerHopsFunc sets the function to calculate hops between servers.
func (s *MailSystem) SetServerHopsFunc(fn func(from, to string) int) {
	s.serverHops = fn
}

// SetDeliveryTime sets the base delivery time per hop in seconds.
func (s *MailSystem) SetDeliveryTime(seconds float64) {
	s.deliveryTime = seconds
}

// SendMail sends a mail message from sender to recipient.
func (s *MailSystem) SendMail(senderID, recipientID, subject, body string, attachments []uint64, serverFrom, serverTo string) (*MailMessage, error) {
	if len(subject) > 50 {
		return nil, fmt.Errorf("subject exceeds 50 characters")
	}
	if len(body) > 500 {
		return nil, fmt.Errorf("body exceeds 500 characters")
	}
	if len(attachments) > 5 {
		return nil, fmt.Errorf("too many attachments (max 5)")
	}

	hops := s.serverHops(serverFrom, serverTo)
	postage := s.CalculatePostage(hops)

	msg := &MailMessage{
		ID:          generateMailMessageID(),
		SenderID:    senderID,
		RecipientID: recipientID,
		Subject:     subject,
		Body:        body,
		Attachments: attachments,
		Postage:     postage,
		SentAt:      time.Now().Unix(),
		DeliveredAt: 0,
	}

	sender, exists := s.world.GetEntity(parseEntityID(senderID))
	if !exists {
		return nil, fmt.Errorf("sender entity not found")
	}

	mailComp, ok := sender.GetComponent("mail")
	if !ok {
		return nil, fmt.Errorf("sender has no mail component")
	}
	mailCompTyped, ok := mailComp.(*MailComponent)
	if !ok {
		return nil, fmt.Errorf("invalid mail component type")
	}

	mailCompTyped.AddToOutbox(msg)

	if hops > 0 {
		s.courierPositions[msg.ID] = &CourierPosition{
			MessageID:        msg.ID,
			CurrentServer:    serverFrom,
			TargetServer:     serverTo,
			Progress:         0.0,
			TotalHops:        hops,
			EstimatedArrival: time.Now().Unix() + int64(s.deliveryTime*float64(hops)),
		}
	} else {
		s.DeliverMail(msg, recipientID)
	}

	return msg, nil
}

// CalculatePostage calculates the postage cost based on distance.
func (s *MailSystem) CalculatePostage(hops int) int {
	return 10 + hops
}

// DeliverMail delivers a message to the recipient's inbox.
func (s *MailSystem) DeliverMail(msg *MailMessage, recipientID string) error {
	recipient, exists := s.world.GetEntity(parseEntityID(recipientID))
	if !exists {
		return fmt.Errorf("recipient entity not found")
	}

	mailComp, ok := recipient.GetComponent("mail")
	if !ok {
		return fmt.Errorf("recipient has no mail component")
	}
	mailCompTyped, ok := mailComp.(*MailComponent)
	if !ok {
		return fmt.Errorf("invalid mail component type")
	}

	msg.DeliveredAt = time.Now().Unix()
	if !mailCompTyped.AddToInbox(msg) {
		return fmt.Errorf("recipient inbox full")
	}

	delete(s.courierPositions, msg.ID)
	return nil
}

// Update processes mail delivery for in-transit messages.
func (s *MailSystem) Update(deltaTime float64) {
	for msgID, courier := range s.courierPositions {
		courier.Progress += deltaTime / s.deliveryTime

		if courier.Progress >= float64(courier.TotalHops) {
			msg := s.findMessageByID(msgID)
			if msg != nil {
				s.DeliverMail(msg, msg.RecipientID)
			} else {
				delete(s.courierPositions, msgID)
			}
		}
	}
}

// GetCourierPosition returns the position of a courier delivering a message.
func (s *MailSystem) GetCourierPosition(messageID string) *CourierPosition {
	return s.courierPositions[messageID]
}

// findMessageByID searches for a message in all player outboxes.
func (s *MailSystem) findMessageByID(messageID string) *MailMessage {
	entities := s.world.GetEntitiesWith("mail")
	for _, entity := range entities {
		mailComp, ok := entity.GetComponent("mail")
		if !ok {
			continue
		}
		mailCompTyped, ok := mailComp.(*MailComponent)
		if !ok {
			continue
		}
		for _, msg := range mailCompTyped.Outbox {
			if msg.ID == messageID {
				return msg
			}
		}
	}
	return nil
}

// generateMailMessageID generates a unique message ID using random bytes.
func generateMailMessageID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based ID on error
		return hex.EncodeToString([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	}
	return hex.EncodeToString(bytes)
}

// parseEntityID parses an entity ID string to uint64.
func parseEntityID(id string) uint64 {
	var entityID uint64
	fmt.Sscanf(id, "%d", &entityID)
	return entityID
}
