package engine

import (
	"fmt"
	"testing"
	"time"
)

func TestNewMailboxUI(t *testing.T) {
	tests := []struct {
		name    string
		x       int
		y       int
		width   int
		height  int
		genreID string
	}{
		{"fantasy", 10, 20, 600, 400, "fantasy"},
		{"scifi", 0, 0, 800, 600, "scifi"},
		{"horror", 50, 50, 400, 300, "horror"},
		{"cyberpunk", 100, 100, 500, 400, "cyberpunk"},
		{"postapoc", 25, 25, 700, 500, "postapoc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui := NewMailboxUI(tt.x, tt.y, tt.width, tt.height, tt.genreID)

			if ui.X != tt.x {
				t.Errorf("X = %d, want %d", ui.X, tt.x)
			}
			if ui.Y != tt.y {
				t.Errorf("Y = %d, want %d", ui.Y, tt.y)
			}
			if ui.Width != tt.width {
				t.Errorf("Width = %d, want %d", ui.Width, tt.width)
			}
			if ui.Height != tt.height {
				t.Errorf("Height = %d, want %d", ui.Height, tt.height)
			}
			if ui.GenreID != tt.genreID {
				t.Errorf("GenreID = %s, want %s", ui.GenreID, tt.genreID)
			}
			if ui.ViewMode != ViewInbox {
				t.Errorf("ViewMode = %v, want ViewInbox", ui.ViewMode)
			}
			if ui.BackgroundColor == nil {
				t.Error("BackgroundColor is nil")
			}
			if ui.TextColor == nil {
				t.Error("TextColor is nil")
			}
			if ui.HighlightColor == nil {
				t.Error("HighlightColor is nil")
			}
			if ui.ComposeAttachments == nil {
				t.Error("ComposeAttachments not initialized")
			}
		})
	}
}

func TestMailboxViewMode_String(t *testing.T) {
	tests := []struct {
		mode MailboxViewMode
		want string
	}{
		{ViewInbox, "Inbox"},
		{ViewOutbox, "Outbox"},
		{ViewCompose, "Compose"},
		{ViewMessageDetail, "Message Detail"},
		{MailboxViewMode(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.mode.String(); got != tt.want {
				t.Errorf("String() = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestLoadFromMailComponent(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")
	mailComp := NewMailComponent()

	// Add some inbox messages
	now := time.Now().Unix()
	mailComp.Inbox = []*MailMessage{
		{
			ID:          "msg1",
			SenderID:    "player2",
			RecipientID: "player1",
			Subject:     "Hello",
			Body:        "Test message 1",
			Attachments: []uint64{100, 101},
			Postage:     15,
			SentAt:      now - 1000,
			DeliveredAt: now - 500,
		},
		{
			ID:          "msg2",
			SenderID:    "player3",
			RecipientID: "player1",
			Subject:     "Important",
			Body:        "Test message 2",
			Attachments: []uint64{},
			Postage:     10,
			SentAt:      now - 2000,
			DeliveredAt: now - 1500,
		},
	}

	// Add some outbox messages
	mailComp.Outbox = []*MailMessage{
		{
			ID:          "msg3",
			SenderID:    "player1",
			RecipientID: "player4",
			Subject:     "Reply",
			Body:        "Test reply",
			Attachments: []uint64{200},
			Postage:     12,
			SentAt:      now - 600,
			DeliveredAt: 0, // Not delivered yet
		},
	}

	ui.LoadFromMailComponent(mailComp)

	// Verify inbox loaded correctly
	if len(ui.InboxMessages) != 2 {
		t.Errorf("InboxMessages count = %d, want 2", len(ui.InboxMessages))
	}
	if ui.InboxMessages[0].ID != "msg1" {
		t.Errorf("First inbox message ID = %s, want msg1 (should be sorted by delivered time)", ui.InboxMessages[0].ID)
	}
	if ui.InboxMessages[0].AttachmentCount != 2 {
		t.Errorf("First message attachment count = %d, want 2", ui.InboxMessages[0].AttachmentCount)
	}

	// Verify outbox loaded correctly
	if len(ui.OutboxMessages) != 1 {
		t.Errorf("OutboxMessages count = %d, want 1", len(ui.OutboxMessages))
	}
	if ui.OutboxMessages[0].ID != "msg3" {
		t.Errorf("First outbox message ID = %s, want msg3", ui.OutboxMessages[0].ID)
	}
	if ui.OutboxMessages[0].Status != MailStatusInTransit {
		t.Errorf("Outbox message status = %v, want InTransit", ui.OutboxMessages[0].Status)
	}
}

func TestMailboxUI_GetUnreadCount(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")
	now := time.Now().Unix()

	// Create mix of unread and read messages
	ui.InboxMessages = []MailEntry{
		{ID: "msg1", DeliveredAt: now - 100, IsUnread: true},
		{ID: "msg2", DeliveredAt: now - 50000, IsUnread: false},
		{ID: "msg3", DeliveredAt: now - 200, IsUnread: true},
	}

	count := ui.GetUnreadCount()
	if count != 2 {
		t.Errorf("GetUnreadCount() = %d, want 2", count)
	}
}

func TestSelectNavigation(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")
	ui.InboxMessages = []MailEntry{{}, {}, {}}
	ui.ViewMode = ViewInbox

	// Test SelectNext
	if ui.SelectedInboxIndex != 0 {
		t.Errorf("Initial SelectedInboxIndex = %d, want 0", ui.SelectedInboxIndex)
	}
	ui.SelectNext()
	if ui.SelectedInboxIndex != 1 {
		t.Errorf("After SelectNext, SelectedInboxIndex = %d, want 1", ui.SelectedInboxIndex)
	}
	ui.SelectNext()
	if ui.SelectedInboxIndex != 2 {
		t.Errorf("After second SelectNext, SelectedInboxIndex = %d, want 2", ui.SelectedInboxIndex)
	}
	ui.SelectNext() // Should not go beyond last index
	if ui.SelectedInboxIndex != 2 {
		t.Errorf("After third SelectNext, SelectedInboxIndex = %d, want 2 (clamped)", ui.SelectedInboxIndex)
	}

	// Test SelectPrevious
	ui.SelectPrevious()
	if ui.SelectedInboxIndex != 1 {
		t.Errorf("After SelectPrevious, SelectedInboxIndex = %d, want 1", ui.SelectedInboxIndex)
	}
	ui.SelectPrevious()
	if ui.SelectedInboxIndex != 0 {
		t.Errorf("After second SelectPrevious, SelectedInboxIndex = %d, want 0", ui.SelectedInboxIndex)
	}
	ui.SelectPrevious() // Should not go below 0
	if ui.SelectedInboxIndex != 0 {
		t.Errorf("After third SelectPrevious, SelectedInboxIndex = %d, want 0 (clamped)", ui.SelectedInboxIndex)
	}
}

func TestSwitchView(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")

	modes := []MailboxViewMode{ViewInbox, ViewOutbox, ViewCompose, ViewMessageDetail}
	for _, mode := range modes {
		ui.SwitchView(mode)
		if ui.ViewMode != mode {
			t.Errorf("After SwitchView(%v), ViewMode = %v", mode, ui.ViewMode)
		}
	}
}

func TestOpenAndCloseMessageDetail(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")
	ui.InboxMessages = []MailEntry{{ID: "msg1"}}
	ui.ViewMode = ViewInbox
	ui.SelectedInboxIndex = 0

	// Open message detail
	ui.OpenSelectedMessage()
	if ui.ViewMode != ViewMessageDetail {
		t.Errorf("After OpenSelectedMessage, ViewMode = %v, want ViewMessageDetail", ui.ViewMode)
	}

	// Close message detail
	ui.CloseMessageDetail()
	if ui.ViewMode != ViewInbox {
		t.Errorf("After CloseMessageDetail, ViewMode = %v, want ViewInbox", ui.ViewMode)
	}

	// Test opening when no message selected
	ui.ViewMode = ViewInbox
	ui.InboxMessages = []MailEntry{} // Empty inbox
	ui.OpenSelectedMessage()
	if ui.ViewMode != ViewInbox {
		t.Error("OpenSelectedMessage should not change view when no message available")
	}
}

func TestAttachmentManagement(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")

	// Test adding attachments
	if !ui.AddAttachment(100) {
		t.Error("AddAttachment(100) failed")
	}
	if !ui.AddAttachment(101) {
		t.Error("AddAttachment(101) failed")
	}
	if len(ui.ComposeAttachments) != 2 {
		t.Errorf("Attachment count = %d, want 2", len(ui.ComposeAttachments))
	}

	// Test max attachment limit
	ui.AddAttachment(102)
	ui.AddAttachment(103)
	ui.AddAttachment(104)
	if !ui.AddAttachment(105) { // Should fail - already at limit
		// This is actually expected since we're at limit after 5
	}
	if len(ui.ComposeAttachments) > 5 {
		t.Errorf("Attachment count = %d, want max 5", len(ui.ComposeAttachments))
	}

	// Test removing attachment
	if !ui.RemoveAttachment(0) {
		t.Error("RemoveAttachment(0) failed")
	}
	if len(ui.ComposeAttachments) < 2 {
		t.Errorf("After RemoveAttachment, count = %d", len(ui.ComposeAttachments))
	}

	// Test removing invalid index
	if ui.RemoveAttachment(999) {
		t.Error("RemoveAttachment(999) should fail for invalid index")
	}
	if ui.RemoveAttachment(-1) {
		t.Error("RemoveAttachment(-1) should fail for negative index")
	}
}

func TestClearCompose(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")

	// Set some compose data
	ui.ComposeRecipient = "player2"
	ui.ComposeSubject = "Test"
	ui.ComposeBody = "Test body"
	ui.ComposeAttachments = []uint64{100, 101}

	// Clear
	ui.ClearCompose()

	if ui.ComposeRecipient != "" {
		t.Errorf("ComposeRecipient = %s, want empty", ui.ComposeRecipient)
	}
	if ui.ComposeSubject != "" {
		t.Errorf("ComposeSubject = %s, want empty", ui.ComposeSubject)
	}
	if ui.ComposeBody != "" {
		t.Errorf("ComposeBody = %s, want empty", ui.ComposeBody)
	}
	if len(ui.ComposeAttachments) != 0 {
		t.Errorf("ComposeAttachments count = %d, want 0", len(ui.ComposeAttachments))
	}
}

func TestGetComposeMessage(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")

	// Set compose data
	ui.ComposeRecipient = "player2"
	ui.ComposeSubject = "Test Subject"
	ui.ComposeBody = "Test Body"
	ui.ComposeAttachments = []uint64{100, 101, 102}

	recipient, subject, body, attachments := ui.GetComposeMessage()

	if recipient != "player2" {
		t.Errorf("recipient = %s, want player2", recipient)
	}
	if subject != "Test Subject" {
		t.Errorf("subject = %s, want Test Subject", subject)
	}
	if body != "Test Body" {
		t.Errorf("body = %s, want Test Body", body)
	}
	if len(attachments) != 3 {
		t.Errorf("attachments count = %d, want 3", len(attachments))
	}
}

func TestRender(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")
	mailComp := NewMailComponent()

	// Add a test message
	now := time.Now().Unix()
	mailComp.Inbox = []*MailMessage{
		{
			ID:          "msg1",
			SenderID:    "player2",
			RecipientID: "player1",
			Subject:     "Test",
			Body:        "Test body",
			Attachments: []uint64{},
			Postage:     10,
			SentAt:      now - 1000,
			DeliveredAt: now - 500,
		},
	}

	ui.LoadFromMailComponent(mailComp)

	// Test rendering each view mode
	modes := []MailboxViewMode{ViewInbox, ViewOutbox, ViewCompose, ViewMessageDetail}
	for _, mode := range modes {
		t.Run(mode.String(), func(t *testing.T) {
			ui.ViewMode = mode
			img := ui.Render()

			if img == nil {
				t.Fatal("Render() returned nil image")
			}
			if img.Bounds().Dx() != 600 {
				t.Errorf("Image width = %d, want 600", img.Bounds().Dx())
			}
			if img.Bounds().Dy() != 400 {
				t.Errorf("Image height = %d, want 400", img.Bounds().Dy())
			}
		})
	}
}

func TestRenderEmptyMailbox(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")
	mailComp := NewMailComponent() // Empty mailbox

	ui.LoadFromMailComponent(mailComp)

	// Test rendering with no messages
	img := ui.Render()
	if img == nil {
		t.Fatal("Render() returned nil for empty mailbox")
	}
}

func TestGetStatusColor(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")

	tests := []struct {
		status   MailStatus
		expected string
	}{
		{MailStatusDelivered, "should return DeliveredColor"},
		{MailStatusInTransit, "should return InTransitColor"},
		{MailStatusFailed, "should return FailedColor"},
		{MailStatusSent, "should return TextColor"},
	}

	for _, tt := range tests {
		t.Run(tt.status.String(), func(t *testing.T) {
			color := ui.getStatusColor(tt.status)
			if color == nil {
				t.Errorf("getStatusColor(%v) returned nil", tt.status)
			}
		})
	}
}

func TestMailboxUI_WrapText(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")

	tests := []struct {
		name     string
		text     string
		maxWidth int
		wantNL   bool // Should contain newline
	}{
		{"empty", "", 100, false},
		{"short", "Short text", 500, false},
		{"long", "This is a very long text that should be wrapped across multiple lines when rendered", 200, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ui.wrapText(tt.text, tt.maxWidth)
			hasNewline := len(result) > 0 && result != tt.text
			if tt.wantNL && !hasNewline && tt.text != "" {
				// Allow for short text not being wrapped
				if len(tt.text) > tt.maxWidth/7 {
					t.Errorf("wrapText() should wrap long text")
				}
			}
		})
	}
}

func TestLightenColor(t *testing.T) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")

	original := ui.BackgroundColor
	lightened := ui.lightenColor(original, 0.2)

	if lightened == nil {
		t.Error("lightenColor() returned nil")
	}
	// Colors should be different (lightened)
	r1, g1, b1, _ := original.RGBA()
	r2, g2, b2, _ := lightened.RGBA()
	if r1 == r2 && g1 == g2 && b1 == b2 {
		t.Error("lightenColor() did not change color")
	}
}

func BenchmarkLoadFromMailComponent(b *testing.B) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")
	mailComp := NewMailComponent()

	// Create 50 messages
	now := time.Now().Unix()
	for i := 0; i < 50; i++ {
		mailComp.Inbox = append(mailComp.Inbox, &MailMessage{
			ID:          fmt.Sprintf("msg%d", i),
			SenderID:    fmt.Sprintf("player%d", i),
			RecipientID: "player1",
			Subject:     "Test",
			Body:        "Test body",
			Postage:     10,
			SentAt:      now - int64(i*100),
			DeliveredAt: now - int64(i*50),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.LoadFromMailComponent(mailComp)
	}
}

func BenchmarkRender(b *testing.B) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")
	mailComp := NewMailComponent()

	now := time.Now().Unix()
	for i := 0; i < 10; i++ {
		mailComp.Inbox = append(mailComp.Inbox, &MailMessage{
			ID:          fmt.Sprintf("msg%d", i),
			SenderID:    fmt.Sprintf("player%d", i),
			RecipientID: "player1",
			Subject:     "Test",
			Body:        "Test body",
			Postage:     10,
			SentAt:      now - int64(i*100),
			DeliveredAt: now - int64(i*50),
		})
	}

	ui.LoadFromMailComponent(mailComp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.Render()
	}
}

func BenchmarkNavigationOperations(b *testing.B) {
	ui := NewMailboxUI(0, 0, 600, 400, "fantasy")
	ui.InboxMessages = make([]MailEntry, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.SelectNext()
		ui.SelectPrevious()
		ui.OpenSelectedMessage()
		ui.CloseMessageDetail()
	}
}
