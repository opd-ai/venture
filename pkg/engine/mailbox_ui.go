package engine

import (
	"fmt"
	"image"
	"image/color"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// MailboxUI represents a mailbox interface for viewing and managing mail.
type MailboxUI struct {
	X, Y          int
	Width, Height int
	GenreID       string
	Visible       bool // Phase 40.3: Toggle visibility

	// Navigation state
	ViewMode              MailboxViewMode
	SelectedInboxIndex    int
	SelectedOutboxIndex   int
	SelectedAttachmentIdx int

	// Content
	InboxMessages  []MailEntry
	OutboxMessages []MailEntry

	// Compose state
	ComposeRecipient   string
	ComposeSubject     string
	ComposeBody        string
	ComposeAttachments []uint64

	// Colors
	BackgroundColor color.Color
	TextColor       color.Color
	HighlightColor  color.Color
	UnreadColor     color.Color
	DeliveredColor  color.Color
	InTransitColor  color.Color
	FailedColor     color.Color
}

// MailboxViewMode represents the current view in the mailbox.
type MailboxViewMode int

const (
	ViewInbox MailboxViewMode = iota
	ViewOutbox
	ViewCompose
	ViewMessageDetail
)

// maxMessagesForHashing limits the number of messages included in state hash
// to avoid excessive string building overhead while still capturing most changes.
const maxMessagesForHashing = 10

// String returns the string representation of the view mode.
func (m MailboxViewMode) String() string {
	switch m {
	case ViewInbox:
		return "Inbox"
	case ViewOutbox:
		return "Outbox"
	case ViewCompose:
		return "Compose"
	case ViewMessageDetail:
		return "Message Detail"
	default:
		return "Unknown"
	}
}

// MailEntry represents a mail message in the UI list.
type MailEntry struct {
	ID              string
	SenderID        string
	RecipientID     string
	Subject         string
	Body            string
	AttachmentCount int
	Postage         int
	Status          MailStatus
	SentAt          int64
	DeliveredAt     int64
	IsUnread        bool
}

// NewMailboxUI creates a new mailbox UI.
func NewMailboxUI(x, y, width, height int, genreID string) *MailboxUI {
	// Determine colors based on genre
	bgColor := color.RGBA{25, 20, 15, 230}           // Dark brown
	textColor := color.RGBA{220, 210, 190, 255}      // Light text
	highlightColor := color.RGBA{255, 200, 100, 255} // Gold highlight
	unreadColor := color.RGBA{255, 255, 100, 255}    // Yellow for unread
	deliveredColor := color.RGBA{100, 255, 100, 255} // Green for delivered
	inTransitColor := color.RGBA{100, 150, 255, 255} // Blue for in transit
	failedColor := color.RGBA{255, 100, 100, 255}    // Red for failed

	if genreID == "scifi" {
		bgColor = color.RGBA{10, 15, 25, 230}
		textColor = color.RGBA{100, 200, 255, 255}
		highlightColor = color.RGBA{0, 255, 255, 255}
		unreadColor = color.RGBA{255, 255, 0, 255}
	} else if genreID == "horror" {
		bgColor = color.RGBA{15, 5, 5, 240}
		textColor = color.RGBA{200, 180, 180, 255}
		highlightColor = color.RGBA{200, 50, 50, 255}
		unreadColor = color.RGBA{255, 200, 200, 255}
	} else if genreID == "cyberpunk" {
		bgColor = color.RGBA{5, 5, 15, 230}
		textColor = color.RGBA{0, 255, 150, 255}
		highlightColor = color.RGBA{255, 0, 255, 255}
		unreadColor = color.RGBA{255, 255, 0, 255}
	} else if genreID == "postapoc" {
		bgColor = color.RGBA{30, 25, 20, 230}
		textColor = color.RGBA{180, 170, 150, 255}
		highlightColor = color.RGBA{200, 150, 100, 255}
		unreadColor = color.RGBA{255, 200, 150, 255}
	}

	return &MailboxUI{
		X:                  x,
		Y:                  y,
		Width:              width,
		Height:             height,
		GenreID:            genreID,
		ViewMode:           ViewInbox,
		BackgroundColor:    bgColor,
		TextColor:          textColor,
		HighlightColor:     highlightColor,
		UnreadColor:        unreadColor,
		DeliveredColor:     deliveredColor,
		InTransitColor:     inTransitColor,
		FailedColor:        failedColor,
		ComposeAttachments: make([]uint64, 0, 5),
	}
}

// LoadFromMailComponent populates the UI from a MailComponent.
func (m *MailboxUI) LoadFromMailComponent(mailComp *MailComponent) {
	// Clear existing data
	m.InboxMessages = make([]MailEntry, 0, len(mailComp.Inbox))
	m.OutboxMessages = make([]MailEntry, 0, len(mailComp.Outbox))

	// Load inbox messages (sort by delivered time, newest first)
	for _, msg := range mailComp.Inbox {
		entry := MailEntry{
			ID:              msg.ID,
			SenderID:        msg.SenderID,
			RecipientID:     msg.RecipientID,
			Subject:         msg.Subject,
			Body:            msg.Body,
			AttachmentCount: len(msg.Attachments),
			Postage:         msg.Postage,
			Status:          msg.GetStatus(),
			SentAt:          msg.SentAt,
			DeliveredAt:     msg.DeliveredAt,
			IsUnread:        m.isMessageUnread(msg),
		}
		m.InboxMessages = append(m.InboxMessages, entry)
	}
	sort.Slice(m.InboxMessages, func(i, j int) bool {
		return m.InboxMessages[i].DeliveredAt > m.InboxMessages[j].DeliveredAt
	})

	// Load outbox messages (sort by sent time, newest first)
	for _, msg := range mailComp.Outbox {
		entry := MailEntry{
			ID:              msg.ID,
			SenderID:        msg.SenderID,
			RecipientID:     msg.RecipientID,
			Subject:         msg.Subject,
			Body:            msg.Body,
			AttachmentCount: len(msg.Attachments),
			Postage:         msg.Postage,
			Status:          msg.GetStatus(),
			SentAt:          msg.SentAt,
			DeliveredAt:     msg.DeliveredAt,
			IsUnread:        false, // Outbox messages are never unread
		}
		m.OutboxMessages = append(m.OutboxMessages, entry)
	}
	sort.Slice(m.OutboxMessages, func(i, j int) bool {
		return m.OutboxMessages[i].SentAt > m.OutboxMessages[j].SentAt
	})
}

// isMessageUnread returns true if message was delivered in the last 24 hours
func (m *MailboxUI) isMessageUnread(msg *MailMessage) bool {
	if msg.DeliveredAt == 0 {
		return false
	}
	now := time.Now().Unix()
	oneDayAgo := now - 86400
	return msg.DeliveredAt > oneDayAgo
}

// Render renders the mailbox UI to an image.
func (m *MailboxUI) Render() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, m.Width, m.Height))

	// Draw background
	m.fillRect(img, 0, 0, m.Width, m.Height, m.BackgroundColor)

	// Draw title bar
	m.drawTitleBar(img)

	// Draw content based on view mode
	switch m.ViewMode {
	case ViewInbox:
		m.drawInboxList(img)
	case ViewOutbox:
		m.drawOutboxList(img)
	case ViewCompose:
		m.drawComposeView(img)
	case ViewMessageDetail:
		m.drawMessageDetail(img)
	}

	return img
}

// Draw renders the mailbox UI directly to an ebiten.Image.
// Performance Audit V6 fix: eliminates CPU→GPU image conversion by drawing directly.
func (m *MailboxUI) Draw(screen *ebiten.Image) {
	if screen == nil || !m.Visible {
		return
	}

	// Draw background
	m.drawEbitenRect(screen, float32(m.X), float32(m.Y), float32(m.Width), float32(m.Height), m.BackgroundColor)

	// Draw title bar
	m.drawEbitenTitleBar(screen)

	// Draw content based on view mode
	switch m.ViewMode {
	case ViewInbox:
		m.drawEbitenInboxList(screen)
	case ViewOutbox:
		m.drawEbitenOutboxList(screen)
	case ViewCompose:
		m.drawEbitenComposeView(screen)
	case ViewMessageDetail:
		m.drawEbitenMessageDetail(screen)
	}
}

// drawEbitenRect draws a filled rectangle directly to an ebiten.Image.
func (m *MailboxUI) drawEbitenRect(screen *ebiten.Image, x, y, width, height float32, col color.Color) {
	r, g, b, a := col.RGBA()
	vector.DrawFilledRect(screen, x, y, width, height,
		color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}, false)
}

// drawEbitenText draws text directly to an ebiten.Image using ebitenutil.
func (m *MailboxUI) drawEbitenText(screen *ebiten.Image, text string, x, y int) {
	if text == "" {
		return
	}
	ebitenutil.DebugPrintAt(screen, text, m.X+x, m.Y+y)
}

// drawEbitenTitleBar draws the title bar with tab buttons using ebiten drawing.
func (m *MailboxUI) drawEbitenTitleBar(screen *ebiten.Image) {
	titleHeight := 40
	tabWidth := m.Width / 3

	// Draw title background
	titleBg := m.lightenColor(m.BackgroundColor, 0.2)
	m.drawEbitenRect(screen, float32(m.X), float32(m.Y), float32(m.Width), float32(titleHeight), titleBg)

	// Draw tabs
	tabs := []struct {
		name string
		mode MailboxViewMode
	}{
		{"Inbox", ViewInbox},
		{"Outbox", ViewOutbox},
		{"Compose", ViewCompose},
	}

	for i, tab := range tabs {
		x := m.X + i*tabWidth
		// Highlight active tab
		if m.ViewMode == tab.mode {
			m.drawEbitenRect(screen, float32(x), float32(m.Y), float32(tabWidth), float32(titleHeight), m.HighlightColor)
		}
		// Draw tab text centered
		textX := x + tabWidth/2 - len(tab.name)*3
		textY := m.Y + titleHeight/2 - 6
		ebitenutil.DebugPrintAt(screen, tab.name, textX, textY)
	}

	// Draw unread count badge on inbox tab if there are unread messages
	unreadCount := m.GetUnreadCount()
	if unreadCount > 0 {
		badgeText := fmt.Sprintf("(%d)", unreadCount)
		r, g, b, a := m.UnreadColor.RGBA()
		badgeCol := color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
		m.drawEbitenRect(screen, float32(m.X+tabWidth-50), float32(m.Y+8), float32(len(badgeText)*7+4), 14, badgeCol)
		ebitenutil.DebugPrintAt(screen, badgeText, m.X+tabWidth-48, m.Y+10)
	}
}

// drawEbitenInboxList draws the inbox message list using ebiten drawing.
func (m *MailboxUI) drawEbitenInboxList(screen *ebiten.Image) {
	startY := 50
	rowHeight := 60
	padding := 10

	// Draw header
	headerText := fmt.Sprintf("Inbox (%d/%d messages)", len(m.InboxMessages), len(m.InboxMessages))
	ebitenutil.DebugPrintAt(screen, headerText, m.X+padding, m.Y+startY)
	startY += 30

	// Draw messages
	for i, msg := range m.InboxMessages {
		y := startY + (i * rowHeight)
		if y+rowHeight > m.Height {
			break // Don't render beyond window
		}

		// Highlight selected message
		if i == m.SelectedInboxIndex {
			m.drawEbitenRect(screen, float32(m.X+padding), float32(m.Y+y), float32(m.Width-2*padding), float32(rowHeight-5), m.HighlightColor)
		}

		// Draw message row
		m.drawEbitenMessageRow(screen, msg, padding+5, y+5, m.Width-2*padding-10, rowHeight-10)
	}

	// Draw empty state
	if len(m.InboxMessages) == 0 {
		emptyText := "No messages in inbox"
		textX := m.X + m.Width/2 - len(emptyText)*3
		ebitenutil.DebugPrintAt(screen, emptyText, textX, m.Y+m.Height/2-10)
	}
}

// drawEbitenOutboxList draws the outbox message list using ebiten drawing.
func (m *MailboxUI) drawEbitenOutboxList(screen *ebiten.Image) {
	startY := 50
	rowHeight := 60
	padding := 10

	// Draw header
	headerText := fmt.Sprintf("Outbox (%d messages sent)", len(m.OutboxMessages))
	ebitenutil.DebugPrintAt(screen, headerText, m.X+padding, m.Y+startY)
	startY += 30

	// Draw messages
	for i, msg := range m.OutboxMessages {
		y := startY + (i * rowHeight)
		if y+rowHeight > m.Height {
			break
		}

		// Highlight selected message
		if i == m.SelectedOutboxIndex {
			m.drawEbitenRect(screen, float32(m.X+padding), float32(m.Y+y), float32(m.Width-2*padding), float32(rowHeight-5), m.HighlightColor)
		}

		// Draw message row
		m.drawEbitenMessageRow(screen, msg, padding+5, y+5, m.Width-2*padding-10, rowHeight-10)
	}

	// Draw empty state
	if len(m.OutboxMessages) == 0 {
		emptyText := "No messages sent"
		textX := m.X + m.Width/2 - len(emptyText)*3
		ebitenutil.DebugPrintAt(screen, emptyText, textX, m.Y+m.Height/2-10)
	}
}

// drawEbitenComposeView draws the compose mail interface using ebiten drawing.
func (m *MailboxUI) drawEbitenComposeView(screen *ebiten.Image) {
	startY := 50
	padding := 10
	fieldHeight := 30

	// Draw compose form
	ebitenutil.DebugPrintAt(screen, "To:", m.X+padding, m.Y+startY)
	fieldBg := m.lightenColor(m.BackgroundColor, 0.3)
	m.drawEbitenRect(screen, float32(m.X+padding+50), float32(m.Y+startY), float32(m.Width-padding-60), float32(fieldHeight), fieldBg)
	ebitenutil.DebugPrintAt(screen, m.ComposeRecipient, m.X+padding+55, m.Y+startY+8)

	startY += fieldHeight + 10
	ebitenutil.DebugPrintAt(screen, "Subject:", m.X+padding, m.Y+startY)
	m.drawEbitenRect(screen, float32(m.X+padding+70), float32(m.Y+startY), float32(m.Width-padding-80), float32(fieldHeight), fieldBg)
	ebitenutil.DebugPrintAt(screen, m.ComposeSubject, m.X+padding+75, m.Y+startY+8)

	startY += fieldHeight + 10
	ebitenutil.DebugPrintAt(screen, "Body:", m.X+padding, m.Y+startY)
	bodyHeight := 150
	m.drawEbitenRect(screen, float32(m.X+padding), float32(m.Y+startY+25), float32(m.Width-2*padding), float32(bodyHeight), fieldBg)
	// Draw wrapped body text
	wrappedBody := m.wrapText(m.ComposeBody, m.Width-2*padding-10)
	lines := strings.Split(wrappedBody, "\n")
	for i, line := range lines {
		if i*15 < bodyHeight-20 {
			ebitenutil.DebugPrintAt(screen, line, m.X+padding+5, m.Y+startY+30+i*15)
		}
	}

	startY += bodyHeight + 35
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Attachments: %d/5", len(m.ComposeAttachments)), m.X+padding, m.Y+startY)

	startY += 30
	// Draw send button
	buttonWidth := 120
	buttonHeight := 35
	sendX := m.X + m.Width - padding - buttonWidth
	m.drawEbitenRect(screen, float32(sendX), float32(m.Y+startY), float32(buttonWidth), float32(buttonHeight), m.HighlightColor)
	ebitenutil.DebugPrintAt(screen, "Send", sendX+buttonWidth/2-12, m.Y+startY+buttonHeight/2-6)

	// Draw cancel button
	cancelBg := m.lightenColor(m.BackgroundColor, 0.4)
	cancelX := sendX - buttonWidth - 10
	m.drawEbitenRect(screen, float32(cancelX), float32(m.Y+startY), float32(buttonWidth), float32(buttonHeight), cancelBg)
	ebitenutil.DebugPrintAt(screen, "Cancel", cancelX+buttonWidth/2-18, m.Y+startY+buttonHeight/2-6)
}

// drawEbitenMessageDetail draws detailed view of a selected message using ebiten drawing.
func (m *MailboxUI) drawEbitenMessageDetail(screen *ebiten.Image) {
	startY := 50
	padding := 10

	var msg MailEntry
	if m.ViewMode == ViewInbox && m.SelectedInboxIndex < len(m.InboxMessages) {
		msg = m.InboxMessages[m.SelectedInboxIndex]
	} else if m.ViewMode == ViewOutbox && m.SelectedOutboxIndex < len(m.OutboxMessages) {
		msg = m.OutboxMessages[m.SelectedOutboxIndex]
	} else {
		emptyText := "No message selected"
		textX := m.X + m.Width/2 - len(emptyText)*3
		ebitenutil.DebugPrintAt(screen, emptyText, textX, m.Y+m.Height/2-10)
		return
	}

	// Draw message header
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("From: %s", msg.SenderID), m.X+padding, m.Y+startY)
	startY += 20
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("To: %s", msg.RecipientID), m.X+padding, m.Y+startY)
	startY += 20
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Subject: %s", msg.Subject), m.X+padding, m.Y+startY)
	startY += 20

	// Draw status and timestamps
	statusText := fmt.Sprintf("Status: %s", msg.Status.String())
	ebitenutil.DebugPrintAt(screen, statusText, m.X+padding, m.Y+startY)
	startY += 20
	if msg.SentAt > 0 {
		sentTime := time.Unix(msg.SentAt, 0).Format("2006-01-02 15:04:05")
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Sent: %s", sentTime), m.X+padding, m.Y+startY)
		startY += 20
	}
	if msg.DeliveredAt > 0 {
		deliveredTime := time.Unix(msg.DeliveredAt, 0).Format("2006-01-02 15:04:05")
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Delivered: %s", deliveredTime), m.X+padding, m.Y+startY)
		startY += 20
	}

	// Draw separator
	startY += 10
	m.drawEbitenRect(screen, float32(m.X+padding), float32(m.Y+startY), float32(m.Width-2*padding), 2, m.HighlightColor)
	startY += 15

	// Draw body
	bodyText := m.wrapText(msg.Body, m.Width-2*padding)
	lines := strings.Split(bodyText, "\n")
	for i, line := range lines {
		ebitenutil.DebugPrintAt(screen, line, m.X+padding, m.Y+startY+i*15)
	}
	startY += m.textHeight(bodyText) + 20

	// Draw attachments
	if msg.AttachmentCount > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Attachments: %d items", msg.AttachmentCount), m.X+padding, m.Y+startY)
	}
}

// drawEbitenMessageRow draws a single message in the list using ebiten drawing.
func (m *MailboxUI) drawEbitenMessageRow(screen *ebiten.Image, msg MailEntry, x, y, width, height int) {
	absX := m.X + x
	absY := m.Y + y

	// Draw unread indicator
	if msg.IsUnread {
		m.drawEbitenRect(screen, float32(absX), float32(absY), 5, float32(height), m.UnreadColor)
		absX += 10
		width -= 10
	}

	// Draw sender/recipient (truncate if too long)
	senderLabel := msg.SenderID
	if m.ViewMode == ViewOutbox {
		senderLabel = "To: " + msg.RecipientID
	}
	if len(senderLabel) > 25 {
		senderLabel = senderLabel[:22] + "..."
	}
	ebitenutil.DebugPrintAt(screen, senderLabel, absX, absY)

	// Draw subject (truncate if too long)
	subject := msg.Subject
	if subject == "" {
		subject = "(no subject)"
	}
	if len(subject) > 35 {
		subject = subject[:32] + "..."
	}
	ebitenutil.DebugPrintAt(screen, subject, absX, absY+15)

	// Draw status indicator
	statusText := msg.Status.String()
	ebitenutil.DebugPrintAt(screen, statusText, absX, absY+30)

	// Draw attachment indicator
	if msg.AttachmentCount > 0 {
		attachText := fmt.Sprintf("[%d]", msg.AttachmentCount)
		ebitenutil.DebugPrintAt(screen, attachText, absX+width-50, absY+30)
	}
}

// drawTitleBar draws the title bar with tab buttons
func (m *MailboxUI) drawTitleBar(img *image.RGBA) {
	titleHeight := 40
	tabWidth := m.Width / 3

	// Draw title background
	titleBg := m.lightenColor(m.BackgroundColor, 0.2)
	m.fillRect(img, 0, 0, m.Width, titleHeight, titleBg)

	// Draw tabs
	tabs := []struct {
		name string
		mode MailboxViewMode
	}{
		{"Inbox", ViewInbox},
		{"Outbox", ViewOutbox},
		{"Compose", ViewCompose},
	}

	for i, tab := range tabs {
		x := i * tabWidth
		// Highlight active tab
		if m.ViewMode == tab.mode {
			m.fillRect(img, x, 0, tabWidth, titleHeight, m.HighlightColor)
		}
		// Draw tab text (simplified - would use actual text rendering in production)
		m.drawCenteredText(img, tab.name, x, 0, tabWidth, titleHeight)
	}

	// Draw unread count badge on inbox tab if there are unread messages
	unreadCount := m.GetUnreadCount()
	if unreadCount > 0 {
		badgeText := fmt.Sprintf("(%d)", unreadCount)
		m.drawText(img, badgeText, tabWidth-50, 10, m.UnreadColor)
	}
}

// drawInboxList draws the inbox message list
func (m *MailboxUI) drawInboxList(img *image.RGBA) {
	startY := 50
	rowHeight := 60
	padding := 10

	// Draw header
	m.drawText(img, fmt.Sprintf("Inbox (%d/%d messages)", len(m.InboxMessages), len(m.InboxMessages)), padding, startY, m.TextColor)
	startY += 30

	// Draw messages
	for i, msg := range m.InboxMessages {
		y := startY + (i * rowHeight)
		if y+rowHeight > m.Height {
			break // Don't render beyond window
		}

		// Highlight selected message
		if i == m.SelectedInboxIndex {
			m.fillRect(img, padding, y, m.Width-2*padding, rowHeight-5, m.HighlightColor)
		}

		// Draw message row
		m.drawMessageRow(img, msg, padding+5, y+5, m.Width-2*padding-10, rowHeight-10)
	}

	// Draw empty state
	if len(m.InboxMessages) == 0 {
		m.drawCenteredText(img, "No messages in inbox", 0, m.Height/2-20, m.Width, 40)
	}
}

// drawOutboxList draws the outbox message list
func (m *MailboxUI) drawOutboxList(img *image.RGBA) {
	startY := 50
	rowHeight := 60
	padding := 10

	// Draw header
	m.drawText(img, fmt.Sprintf("Outbox (%d messages sent)", len(m.OutboxMessages)), padding, startY, m.TextColor)
	startY += 30

	// Draw messages
	for i, msg := range m.OutboxMessages {
		y := startY + (i * rowHeight)
		if y+rowHeight > m.Height {
			break
		}

		// Highlight selected message
		if i == m.SelectedOutboxIndex {
			m.fillRect(img, padding, y, m.Width-2*padding, rowHeight-5, m.HighlightColor)
		}

		// Draw message row
		m.drawMessageRow(img, msg, padding+5, y+5, m.Width-2*padding-10, rowHeight-10)
	}

	// Draw empty state
	if len(m.OutboxMessages) == 0 {
		m.drawCenteredText(img, "No messages sent", 0, m.Height/2-20, m.Width, 40)
	}
}

// drawComposeView draws the compose mail interface
func (m *MailboxUI) drawComposeView(img *image.RGBA) {
	startY := 50
	padding := 10
	fieldHeight := 30

	// Draw compose form
	m.drawText(img, "To:", padding, startY, m.TextColor)
	m.fillRect(img, padding+50, startY, m.Width-padding-60, fieldHeight, m.lightenColor(m.BackgroundColor, 0.3))
	m.drawText(img, m.ComposeRecipient, padding+55, startY+5, m.TextColor)

	startY += fieldHeight + 10
	m.drawText(img, "Subject:", padding, startY, m.TextColor)
	m.fillRect(img, padding+70, startY, m.Width-padding-80, fieldHeight, m.lightenColor(m.BackgroundColor, 0.3))
	m.drawText(img, m.ComposeSubject, padding+75, startY+5, m.TextColor)

	startY += fieldHeight + 10
	m.drawText(img, "Body:", padding, startY, m.TextColor)
	bodyHeight := 150
	m.fillRect(img, padding, startY+25, m.Width-2*padding, bodyHeight, m.lightenColor(m.BackgroundColor, 0.3))
	m.drawText(img, m.wrapText(m.ComposeBody, m.Width-2*padding-10), padding+5, startY+30, m.TextColor)

	startY += bodyHeight + 35
	m.drawText(img, fmt.Sprintf("Attachments: %d/5", len(m.ComposeAttachments)), padding, startY, m.TextColor)

	startY += 30
	// Draw send button
	buttonWidth := 120
	buttonHeight := 35
	sendX := m.Width - padding - buttonWidth
	m.fillRect(img, sendX, startY, buttonWidth, buttonHeight, m.HighlightColor)
	m.drawCenteredText(img, "Send", sendX, startY, buttonWidth, buttonHeight)

	// Draw cancel button
	cancelX := sendX - buttonWidth - 10
	m.fillRect(img, cancelX, startY, buttonWidth, buttonHeight, m.lightenColor(m.BackgroundColor, 0.4))
	m.drawCenteredText(img, "Cancel", cancelX, startY, buttonWidth, buttonHeight)
}

// drawMessageDetail draws detailed view of a selected message
func (m *MailboxUI) drawMessageDetail(img *image.RGBA) {
	startY := 50
	padding := 10

	var msg MailEntry
	if m.ViewMode == ViewInbox && m.SelectedInboxIndex < len(m.InboxMessages) {
		msg = m.InboxMessages[m.SelectedInboxIndex]
	} else if m.ViewMode == ViewOutbox && m.SelectedOutboxIndex < len(m.OutboxMessages) {
		msg = m.OutboxMessages[m.SelectedOutboxIndex]
	} else {
		m.drawCenteredText(img, "No message selected", 0, m.Height/2-20, m.Width, 40)
		return
	}

	// Draw message header
	m.drawText(img, fmt.Sprintf("From: %s", msg.SenderID), padding, startY, m.TextColor)
	startY += 20
	m.drawText(img, fmt.Sprintf("To: %s", msg.RecipientID), padding, startY, m.TextColor)
	startY += 20
	m.drawText(img, fmt.Sprintf("Subject: %s", msg.Subject), padding, startY, m.HighlightColor)
	startY += 20

	// Draw status and timestamps
	statusColor := m.getStatusColor(msg.Status)
	m.drawText(img, fmt.Sprintf("Status: %s", msg.Status.String()), padding, startY, statusColor)
	startY += 20
	if msg.SentAt > 0 {
		sentTime := time.Unix(msg.SentAt, 0).Format("2006-01-02 15:04:05")
		m.drawText(img, fmt.Sprintf("Sent: %s", sentTime), padding, startY, m.TextColor)
		startY += 20
	}
	if msg.DeliveredAt > 0 {
		deliveredTime := time.Unix(msg.DeliveredAt, 0).Format("2006-01-02 15:04:05")
		m.drawText(img, fmt.Sprintf("Delivered: %s", deliveredTime), padding, startY, m.TextColor)
		startY += 20
	}

	// Draw separator
	startY += 10
	m.fillRect(img, padding, startY, m.Width-2*padding, 2, m.HighlightColor)
	startY += 15

	// Draw body
	bodyText := m.wrapText(msg.Body, m.Width-2*padding)
	m.drawText(img, bodyText, padding, startY, m.TextColor)
	startY += m.textHeight(bodyText) + 20

	// Draw attachments
	if msg.AttachmentCount > 0 {
		m.drawText(img, fmt.Sprintf("Attachments: %d items", msg.AttachmentCount), padding, startY, m.HighlightColor)
	}
}

// drawMessageRow draws a single message in the list
func (m *MailboxUI) drawMessageRow(img *image.RGBA, msg MailEntry, x, y, width, height int) {
	// Draw unread indicator
	if msg.IsUnread {
		m.fillRect(img, x, y, 5, height, m.UnreadColor)
		x += 10
		width -= 10
	}

	// Draw sender/recipient (truncate if too long)
	senderLabel := msg.SenderID
	if m.ViewMode == ViewOutbox {
		senderLabel = "To: " + msg.RecipientID
	}
	if len(senderLabel) > 25 {
		senderLabel = senderLabel[:22] + "..."
	}
	m.drawText(img, senderLabel, x, y, m.TextColor)

	// Draw subject (truncate if too long)
	subject := msg.Subject
	if subject == "" {
		subject = "(no subject)"
	}
	if len(subject) > 35 {
		subject = subject[:32] + "..."
	}
	m.drawText(img, subject, x, y+15, m.HighlightColor)

	// Draw status indicator
	statusColor := m.getStatusColor(msg.Status)
	statusText := msg.Status.String()
	m.drawText(img, statusText, x, y+30, statusColor)

	// Draw attachment indicator
	if msg.AttachmentCount > 0 {
		attachText := fmt.Sprintf("[%d]", msg.AttachmentCount)
		m.drawText(img, attachText, x+width-50, y+30, m.TextColor)
	}
}

// Helper functions

// GetUnreadCount returns the number of unread messages in the inbox
func (m *MailboxUI) GetUnreadCount() int {
	count := 0
	for _, msg := range m.InboxMessages {
		if msg.IsUnread {
			count++
		}
	}
	return count
}

func (m *MailboxUI) getStatusColor(status MailStatus) color.Color {
	switch status {
	case MailStatusDelivered:
		return m.DeliveredColor
	case MailStatusInTransit:
		return m.InTransitColor
	case MailStatusFailed:
		return m.FailedColor
	default:
		return m.TextColor
	}
}

// Navigation methods

// SelectNext selects the next message in the current view
func (m *MailboxUI) SelectNext() {
	switch m.ViewMode {
	case ViewInbox:
		if m.SelectedInboxIndex < len(m.InboxMessages)-1 {
			m.SelectedInboxIndex++
		}
	case ViewOutbox:
		if m.SelectedOutboxIndex < len(m.OutboxMessages)-1 {
			m.SelectedOutboxIndex++
		}
	}
}

// SelectPrevious selects the previous message in the current view
func (m *MailboxUI) SelectPrevious() {
	switch m.ViewMode {
	case ViewInbox:
		if m.SelectedInboxIndex > 0 {
			m.SelectedInboxIndex--
		}
	case ViewOutbox:
		if m.SelectedOutboxIndex > 0 {
			m.SelectedOutboxIndex--
		}
	}
}

// SwitchView changes the current view mode
func (m *MailboxUI) SwitchView(mode MailboxViewMode) {
	m.ViewMode = mode
}

// OpenSelectedMessage opens the selected message in detail view
func (m *MailboxUI) OpenSelectedMessage() {
	var hasMessage bool
	switch m.ViewMode {
	case ViewInbox:
		hasMessage = m.SelectedInboxIndex < len(m.InboxMessages)
	case ViewOutbox:
		hasMessage = m.SelectedOutboxIndex < len(m.OutboxMessages)
	}
	if hasMessage {
		m.ViewMode = ViewMessageDetail
	}
}

// CloseMessageDetail closes the message detail view
func (m *MailboxUI) CloseMessageDetail() {
	if m.ViewMode == ViewMessageDetail {
		m.ViewMode = ViewInbox
	}
}

// AddAttachment adds an item to the compose attachments list
func (m *MailboxUI) AddAttachment(itemID uint64) bool {
	if len(m.ComposeAttachments) >= 5 {
		return false
	}
	m.ComposeAttachments = append(m.ComposeAttachments, itemID)
	return true
}

// RemoveAttachment removes an item from the compose attachments list
func (m *MailboxUI) RemoveAttachment(index int) bool {
	if index < 0 || index >= len(m.ComposeAttachments) {
		return false
	}
	m.ComposeAttachments = append(m.ComposeAttachments[:index], m.ComposeAttachments[index+1:]...)
	return true
}

// ClearCompose resets the compose form
func (m *MailboxUI) ClearCompose() {
	m.ComposeRecipient = ""
	m.ComposeSubject = ""
	m.ComposeBody = ""
	m.ComposeAttachments = make([]uint64, 0, 5)
}

// GetComposeMessage returns the composed message data for sending
func (m *MailboxUI) GetComposeMessage() (recipient, subject, body string, attachments []uint64) {
	return m.ComposeRecipient, m.ComposeSubject, m.ComposeBody, m.ComposeAttachments
}

// Rendering helper methods (simplified - would use actual font rendering in production)

func (m *MailboxUI) fillRect(img *image.RGBA, x, y, width, height int, col color.Color) {
	for py := y; py < y+height && py < m.Height; py++ {
		for px := x; px < x+width && px < m.Width; px++ {
			if px >= 0 && py >= 0 {
				img.Set(px, py, col)
			}
		}
	}
}

func (m *MailboxUI) drawText(img *image.RGBA, text string, x, y int, col color.Color) {
	// Simplified text rendering - in production would use actual font library
	// This is a placeholder that draws a colored rectangle to represent text
	if text == "" {
		return
	}
	textWidth := len(text) * 7
	textHeight := 12
	for py := y; py < y+textHeight && py < m.Height; py++ {
		for px := x; px < x+textWidth && px < m.Width; px++ {
			if px >= 0 && py >= 0 {
				// Only draw on some pixels to simulate text appearance
				if (px-x)%7 == 0 || (py-y)%3 == 0 {
					img.Set(px, py, col)
				}
			}
		}
	}
}

func (m *MailboxUI) drawCenteredText(img *image.RGBA, text string, x, y, width, height int) {
	textWidth := len(text) * 7
	textX := x + (width-textWidth)/2
	textY := y + (height-12)/2
	m.drawText(img, text, textX, textY, m.TextColor)
}

func (m *MailboxUI) lightenColor(col color.Color, amount float64) color.Color {
	r, g, b, a := col.RGBA()
	factor := 1.0 + amount
	return color.RGBA{
		R: uint8(min(255, int(float64(r>>8)*factor))),
		G: uint8(min(255, int(float64(g>>8)*factor))),
		B: uint8(min(255, int(float64(b>>8)*factor))),
		A: uint8(a >> 8),
	}
}

func (m *MailboxUI) wrapText(text string, maxWidth int) string {
	if text == "" {
		return text
	}
	// Simple word wrapping
	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	currentLine := ""
	maxLineWidth := maxWidth / 7 // Approximate characters per line

	for _, word := range words {
		if len(currentLine)+len(word)+1 <= maxLineWidth {
			if currentLine != "" {
				currentLine += " "
			}
			currentLine += word
		} else {
			if currentLine != "" {
				lines = append(lines, currentLine)
			}
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	return strings.Join(lines, "\n")
}

func (m *MailboxUI) textHeight(text string) int {
	lines := strings.Count(text, "\n") + 1
	return lines * 15 // Approximate line height
}

// Toggle toggles the mailbox UI visibility (Phase 40.3).
func (m *MailboxUI) Toggle() {
	m.Visible = !m.Visible
	if m.Visible {
		m.SwitchView(ViewInbox) // Reset to inbox when opening
	}
}

// Open shows the mailbox UI (Phase 40.3).
func (m *MailboxUI) Open() {
	m.Visible = true
	m.SwitchView(ViewInbox)
}

// Close hides the mailbox UI (Phase 40.3).
func (m *MailboxUI) Close() {
	m.Visible = false
}

// IsOpen returns true if the mailbox UI is currently visible (Phase 40.3).
func (m *MailboxUI) IsOpen() bool {
	return m.Visible
}

// GetStateHash returns a hash of the mailbox UI state for caching purposes.
// Used to detect when the UI content has changed and needs to be re-rendered.
// Uses strings.Builder for efficient string building without per-call allocations.
func (m *MailboxUI) GetStateHash() string {
	var sb strings.Builder
	// Estimate capacity based on compose fields plus overhead for integers and separators
	estimated := len(m.ComposeRecipient) + len(m.ComposeSubject) + len(m.ComposeBody) + 64
	if estimated < 256 {
		estimated = 256
	}
	sb.Grow(estimated)

	// Write state values separated by colons
	sb.WriteString(strconv.Itoa(int(m.ViewMode)))
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(m.SelectedInboxIndex))
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(m.SelectedOutboxIndex))
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(m.SelectedAttachmentIdx))
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(len(m.InboxMessages)))
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(len(m.OutboxMessages)))
	sb.WriteByte(':')
	sb.WriteString(m.ComposeRecipient)
	sb.WriteByte(':')
	sb.WriteString(m.ComposeSubject)
	sb.WriteByte(':')
	sb.WriteString(m.ComposeBody)
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(len(m.ComposeAttachments)))

	// Include message content hashes for inbox messages (status, read state, timestamps)
	for i, msg := range m.InboxMessages {
		if i >= maxMessagesForHashing {
			break
		}
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(int(msg.Status)))
		sb.WriteByte(',')
		if msg.IsUnread {
			sb.WriteByte('1')
		} else {
			sb.WriteByte('0')
		}
		sb.WriteByte(',')
		sb.WriteString(strconv.FormatInt(msg.DeliveredAt, 10))
	}

	// Include message content hashes for outbox messages
	for i, msg := range m.OutboxMessages {
		if i >= maxMessagesForHashing {
			break
		}
		sb.WriteByte(':')
		sb.WriteString(strconv.Itoa(int(msg.Status)))
		sb.WriteByte(',')
		sb.WriteString(strconv.FormatInt(msg.SentAt, 10))
	}

	return sb.String()
}
