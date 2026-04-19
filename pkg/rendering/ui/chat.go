// Package ui provides chat UI rendering functionality.
package ui

import (
	"fmt"
	"image"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// ChatUI renders the chat interface with message history and input field.
type ChatUI struct {
	// Position and size
	X, Y          int
	Width, Height int

	// Messages to display
	Messages []ChatMessage

	// Input field state
	InputText     string
	InputActive   bool
	CursorVisible bool
	lastBlink     time.Time

	// Channel tabs
	ActiveChannel int
	Channels      []ChatChannel

	// Scrolling
	ScrollOffset int
	MaxMessages  int

	// Visual settings
	BackgroundColor color.Color
	TextColor       color.Color
	InputBGColor    color.Color
	BorderColor     color.Color
	Font            font.Face

	// Layout
	MessageHeight int
	InputHeight   int
	ChannelHeight int
	Padding       int
}

// ChatMessage represents a displayed chat message.
type ChatMessage struct {
	SenderName string
	Content    string
	Channel    int
	Timestamp  time.Time
	IsSystem   bool
}

// ChatChannel represents a chat channel tab.
type ChatChannel struct {
	ID          int
	Name        string
	UnreadCount int
}

// NewChatUI creates a new chat UI instance.
func NewChatUI(x, y, width, height int) *ChatUI {
	return &ChatUI{
		X:               x,
		Y:               y,
		Width:           width,
		Height:          height,
		Messages:        make([]ChatMessage, 0, 100),
		Channels:        defaultChannels(),
		ActiveChannel:   0,
		MaxMessages:     100,
		BackgroundColor: color.RGBA{20, 20, 30, 200},
		TextColor:       color.RGBA{220, 220, 220, 255},
		InputBGColor:    color.RGBA{30, 30, 40, 255},
		BorderColor:     color.RGBA{100, 100, 120, 255},
		Font:            basicfont.Face7x13,
		MessageHeight:   15,
		InputHeight:     30,
		ChannelHeight:   25,
		Padding:         5,
		lastBlink:       time.Now(),
	}
}

// defaultChannels returns the default chat channels.
func defaultChannels() []ChatChannel {
	return []ChatChannel{
		{ID: 0, Name: "Global", UnreadCount: 0},
		{ID: 1, Name: "Local", UnreadCount: 0},
		{ID: 2, Name: "Party", UnreadCount: 0},
		{ID: 3, Name: "Whisper", UnreadCount: 0},
	}
}

// Update updates the chat UI state (cursor blinking, etc.)
func (ui *ChatUI) Update(deltaTime float64) {
	// Cursor blink (500ms interval)
	if time.Since(ui.lastBlink) > 500*time.Millisecond {
		ui.CursorVisible = !ui.CursorVisible
		ui.lastBlink = time.Now()
	}

	// Limit message history
	if len(ui.Messages) > ui.MaxMessages {
		ui.Messages = ui.Messages[len(ui.Messages)-ui.MaxMessages:]
	}
}

// Render draws the chat UI to the screen.
func (ui *ChatUI) Render(screen *ebiten.Image) {
	// Draw background panel
	ui.drawPanel(screen, ui.X, ui.Y, ui.Width, ui.Height, ui.BackgroundColor)

	// Draw channel tabs
	ui.drawChannelTabs(screen)

	// Draw message area
	messageY := ui.Y + ui.ChannelHeight + ui.Padding
	messageHeight := ui.Height - ui.ChannelHeight - ui.InputHeight - 3*ui.Padding
	ui.drawMessages(screen, messageY, messageHeight)

	// Draw input field
	inputY := ui.Y + ui.Height - ui.InputHeight - ui.Padding
	ui.drawInputField(screen, inputY)
}

// drawPanel draws a filled rectangle with border.
func (ui *ChatUI) drawPanel(screen *ebiten.Image, x, y, width, height int, bgColor color.Color) {
	// Background
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), bgColor, false)

	// Border
	vector.StrokeRect(screen, float32(x), float32(y), float32(width), float32(height), 1, ui.BorderColor, false)
}

// drawChannelTabs draws the channel selection tabs.
func (ui *ChatUI) drawChannelTabs(screen *ebiten.Image) {
	tabWidth := ui.Width / len(ui.Channels)
	tabY := ui.Y + ui.Padding

	for i, channel := range ui.Channels {
		tabX := ui.X + i*tabWidth

		// Tab background (highlight active channel)
		tabColor := ui.InputBGColor
		if i == ui.ActiveChannel {
			tabColor = color.RGBA{50, 50, 70, 255}
		}
		ui.drawPanel(screen, tabX, tabY, tabWidth-2, ui.ChannelHeight-ui.Padding, tabColor)

		// Tab text
		text.Draw(screen, channel.Name, ui.Font, tabX+ui.Padding, tabY+16, ui.TextColor)

		// Unread count badge
		if channel.UnreadCount > 0 {
			badgeText := fmt.Sprintf("(%d)", channel.UnreadCount)
			badgeX := tabX + tabWidth - ui.Padding - len(badgeText)*7
			text.Draw(screen, badgeText, ui.Font, badgeX, tabY+16, color.RGBA{255, 200, 100, 255})
		}
	}
}

// drawMessages draws the message history.
func (ui *ChatUI) drawMessages(screen *ebiten.Image, y, height int) {
	// Message area background
	ui.drawPanel(screen, ui.X+ui.Padding, y, ui.Width-2*ui.Padding, height, ui.InputBGColor)

	// Calculate visible messages
	maxVisible := height / ui.MessageHeight
	startIdx := len(ui.Messages) - maxVisible - ui.ScrollOffset
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := len(ui.Messages) - ui.ScrollOffset
	if endIdx > len(ui.Messages) {
		endIdx = len(ui.Messages)
	}

	// Draw messages
	drawY := y + ui.Padding
	for i := startIdx; i < endIdx; i++ {
		msg := ui.Messages[i]

		// Skip messages from other channels (unless system message)
		if !msg.IsSystem && msg.Channel != ui.ActiveChannel {
			continue
		}

		// Format message
		msgText := ui.formatMessage(msg)

		// Message color (system messages in yellow)
		msgColor := ui.TextColor
		if msg.IsSystem {
			msgColor = color.RGBA{255, 255, 100, 255}
		}

		// Draw message text
		text.Draw(screen, msgText, ui.Font, ui.X+ui.Padding*2, drawY+13, msgColor)
		drawY += ui.MessageHeight
	}
}

// formatMessage formats a message for display.
func (ui *ChatUI) formatMessage(msg ChatMessage) string {
	if msg.IsSystem {
		return fmt.Sprintf("[SYSTEM] %s", msg.Content)
	}

	// Format: [HH:MM] SenderName: Content
	timestamp := msg.Timestamp.Format("15:04")
	return fmt.Sprintf("[%s] %s: %s", timestamp, msg.SenderName, msg.Content)
}

// drawInputField draws the chat input field.
func (ui *ChatUI) drawInputField(screen *ebiten.Image, y int) {
	// Input field background
	ui.drawPanel(screen, ui.X+ui.Padding, y, ui.Width-2*ui.Padding, ui.InputHeight, ui.InputBGColor)

	// Input text
	inputText := ui.InputText
	if ui.InputActive && ui.CursorVisible {
		inputText += "|" // Cursor
	}

	text.Draw(screen, inputText, ui.Font, ui.X+ui.Padding*2, y+20, ui.TextColor)

	// Placeholder text when empty
	if ui.InputText == "" && !ui.InputActive {
		placeholderText := "Press Enter to chat..."
		text.Draw(screen, placeholderText, ui.Font, ui.X+ui.Padding*2, y+20, color.RGBA{150, 150, 150, 255})
	}
}

// AddMessage adds a message to the chat history.
func (ui *ChatUI) AddMessage(msg ChatMessage) {
	ui.Messages = append(ui.Messages, msg)

	// Increment unread count for non-active channels
	if msg.Channel != ui.ActiveChannel && !msg.IsSystem {
		for i := range ui.Channels {
			if ui.Channels[i].ID == msg.Channel {
				ui.Channels[i].UnreadCount++
				break
			}
		}
	}

	// Auto-scroll to bottom when new message arrives
	ui.ScrollOffset = 0
}

// AddSystemMessage adds a system message (visible in all channels).
func (ui *ChatUI) AddSystemMessage(content string) {
	ui.AddMessage(ChatMessage{
		SenderName: "System",
		Content:    content,
		Channel:    ui.ActiveChannel,
		Timestamp:  time.Now(),
		IsSystem:   true,
	})
}

// SetActiveChannel switches to a different channel.
func (ui *ChatUI) SetActiveChannel(channelID int) {
	if channelID < 0 || channelID >= len(ui.Channels) {
		return
	}

	ui.ActiveChannel = channelID

	// Clear unread count for activated channel
	for i := range ui.Channels {
		if ui.Channels[i].ID == channelID {
			ui.Channels[i].UnreadCount = 0
			break
		}
	}
}

// AppendInputChar adds a character to the input field.
func (ui *ChatUI) AppendInputChar(char rune) {
	if len(ui.InputText) < 200 { // Max message length
		ui.InputText += string(char)
	}
}

// BackspaceInput removes the last character from input.
func (ui *ChatUI) BackspaceInput() {
	if len(ui.InputText) > 0 {
		ui.InputText = ui.InputText[:len(ui.InputText)-1]
	}
}

// ClearInput clears the input field.
func (ui *ChatUI) ClearInput() {
	ui.InputText = ""
}

// GetInputText returns the current input text.
func (ui *ChatUI) GetInputText() string {
	return ui.InputText
}

// SetInputActive sets whether the input field is active.
func (ui *ChatUI) SetInputActive(active bool) {
	ui.InputActive = active
	if active {
		ui.CursorVisible = true
		ui.lastBlink = time.Now()
	}
}

// IsInputActive returns whether the input field is active.
func (ui *ChatUI) IsInputActive() bool {
	return ui.InputActive
}

// ScrollUp scrolls the message history up.
func (ui *ChatUI) ScrollUp() {
	maxScroll := len(ui.Messages) - (ui.Height-ui.ChannelHeight-ui.InputHeight-3*ui.Padding)/ui.MessageHeight
	if ui.ScrollOffset < maxScroll {
		ui.ScrollOffset++
	}
}

// ScrollDown scrolls the message history down.
func (ui *ChatUI) ScrollDown() {
	if ui.ScrollOffset > 0 {
		ui.ScrollOffset--
	}
}

// SetPosition sets the UI position.
func (ui *ChatUI) SetPosition(x, y int) {
	ui.X = x
	ui.Y = y
}

// SetSize sets the UI size.
func (ui *ChatUI) SetSize(width, height int) {
	ui.Width = width
	ui.Height = height
}

// GetBounds returns the UI bounding box.
func (ui *ChatUI) GetBounds() image.Rectangle {
	return image.Rect(ui.X, ui.Y, ui.X+ui.Width, ui.Y+ui.Height)
}

// ContainsPoint checks if a point is within the UI bounds.
func (ui *ChatUI) ContainsPoint(x, y int) bool {
	bounds := ui.GetBounds()
	return x >= bounds.Min.X && x < bounds.Max.X && y >= bounds.Min.Y && y < bounds.Max.Y
}

// HandleClick handles a mouse click event.
func (ui *ChatUI) HandleClick(x, y int) {
	// Check channel tab clicks
	if y >= ui.Y+ui.Padding && y <= ui.Y+ui.ChannelHeight+ui.Padding {
		tabWidth := ui.Width / len(ui.Channels)
		for i := range ui.Channels {
			tabX := ui.X + i*tabWidth
			if x >= tabX && x < tabX+tabWidth {
				ui.SetActiveChannel(i)
				break
			}
		}
	} else {
		// Check input field click — only if the tab row was not hit, so that
		// overlapping regions (small window) do not activate both at once.
		inputY := ui.Y + ui.Height - ui.InputHeight - ui.Padding
		if y >= inputY && y <= inputY+ui.InputHeight {
			ui.SetInputActive(true)
		}
	}
}
