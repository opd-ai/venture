// Package ui provides notification and toast message rendering functionality.
package ui

import (
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// NotificationType represents the type/severity of a notification.
type NotificationType int

const (
	// NotificationInfo - Informational message (blue)
	NotificationInfo NotificationType = iota
	// NotificationSuccess - Success message (green)
	NotificationSuccess
	// NotificationWarning - Warning message (yellow/orange)
	NotificationWarning
	// NotificationError - Error message (red)
	NotificationError
)

// Notification represents a toast notification message.
type Notification struct {
	Type      NotificationType
	Message   string
	CreatedAt time.Time
	Duration  time.Duration
	FadeOut   bool // Whether notification is fading out
}

// IsExpired returns whether the notification has expired.
func (n *Notification) IsExpired() bool {
	return time.Since(n.CreatedAt) >= n.Duration
}

// GetAlpha returns the alpha value (0-255) for fade-in/fade-out effect.
func (n *Notification) GetAlpha() uint8 {
	elapsed := time.Since(n.CreatedAt)

	// Fade in during first 200ms
	fadeInDuration := 200 * time.Millisecond
	if elapsed < fadeInDuration {
		progress := float64(elapsed) / float64(fadeInDuration)
		return uint8(255 * progress)
	}

	// Fade out during last 500ms
	fadeOutDuration := 500 * time.Millisecond
	remaining := n.Duration - elapsed
	if remaining < fadeOutDuration {
		progress := float64(remaining) / float64(fadeOutDuration)
		return uint8(255 * progress)
	}

	// Fully opaque in middle
	return 255
}

// NotificationManager manages and renders toast notifications.
type NotificationManager struct {
	// Position settings
	X, Y      int // Top-left position for notification stack
	Width     int
	MaxHeight int
	Spacing   int
	Padding   int

	// Active notifications
	Notifications    []Notification
	MaxNotifications int

	// Visual settings
	Font font.Face

	// Default duration
	DefaultDuration time.Duration
}

// NewNotificationManager creates a new notification manager.
func NewNotificationManager(x, y, width int) *NotificationManager {
	return &NotificationManager{
		X:                x,
		Y:                y,
		Width:            width,
		MaxHeight:        60,
		Spacing:          10,
		Padding:          10,
		Notifications:    make([]Notification, 0, 10),
		MaxNotifications: 5, // Maximum 5 notifications on screen
		Font:             basicfont.Face7x13,
		DefaultDuration:  3 * time.Second,
	}
}

// AddNotification adds a new notification to the stack.
func (nm *NotificationManager) AddNotification(notifType NotificationType, message string) {
	notification := Notification{
		Type:      notifType,
		Message:   message,
		CreatedAt: time.Now(),
		Duration:  nm.DefaultDuration,
		FadeOut:   false,
	}

	nm.Notifications = append(nm.Notifications, notification)

	// Limit notification count
	if len(nm.Notifications) > nm.MaxNotifications {
		nm.Notifications = nm.Notifications[len(nm.Notifications)-nm.MaxNotifications:]
	}
}

// AddInfo adds an informational notification.
func (nm *NotificationManager) AddInfo(message string) {
	nm.AddNotification(NotificationInfo, message)
}

// AddSuccess adds a success notification.
func (nm *NotificationManager) AddSuccess(message string) {
	nm.AddNotification(NotificationSuccess, message)
}

// AddWarning adds a warning notification.
func (nm *NotificationManager) AddWarning(message string) {
	nm.AddNotification(NotificationWarning, message)
}

// AddError adds an error notification.
func (nm *NotificationManager) AddError(message string) {
	nm.AddNotification(NotificationError, message)
}

// AddCustomDuration adds a notification with custom duration.
func (nm *NotificationManager) AddCustomDuration(notifType NotificationType, message string, duration time.Duration) {
	notification := Notification{
		Type:      notifType,
		Message:   message,
		CreatedAt: time.Now(),
		Duration:  duration,
		FadeOut:   false,
	}

	nm.Notifications = append(nm.Notifications, notification)

	if len(nm.Notifications) > nm.MaxNotifications {
		nm.Notifications = nm.Notifications[len(nm.Notifications)-nm.MaxNotifications:]
	}
}

// Update updates notification state and removes expired notifications.
func (nm *NotificationManager) Update(deltaTime float64) {
	// Remove expired notifications
	activeNotifications := make([]Notification, 0, len(nm.Notifications))
	for _, notif := range nm.Notifications {
		if !notif.IsExpired() {
			activeNotifications = append(activeNotifications, notif)
		}
	}
	nm.Notifications = activeNotifications
}

// Clear removes all notifications.
func (nm *NotificationManager) Clear() {
	nm.Notifications = make([]Notification, 0, 10)
}

// GetCount returns the number of active notifications.
func (nm *NotificationManager) GetCount() int {
	return len(nm.Notifications)
}

// Render draws all active notifications to the screen.
func (nm *NotificationManager) Render(screen *ebiten.Image) {
	// Draw from top to bottom
	drawY := nm.Y

	for i := len(nm.Notifications) - 1; i >= 0; i-- {
		notif := nm.Notifications[i]

		// Get notification color based on type
		bgColor := nm.getBackgroundColor(notif.Type, notif.GetAlpha())
		borderColor := nm.getBorderColor(notif.Type, notif.GetAlpha())
		textColor := nm.getTextColor(notif.GetAlpha())

		// Calculate notification height based on text wrapping
		lines := nm.wrapText(notif.Message, nm.Width-2*nm.Padding)
		height := len(lines)*15 + 2*nm.Padding
		if height > nm.MaxHeight {
			height = nm.MaxHeight
		}

		// Draw notification panel
		nm.drawNotification(screen, nm.X, drawY, nm.Width, height, notif.Message, bgColor, borderColor, textColor)

		drawY += height + nm.Spacing
	}
}

// drawNotification draws a single notification panel.
func (nm *NotificationManager) drawNotification(screen *ebiten.Image, x, y, width, height int, message string, bgColor, borderColor, textColor color.Color) {
	// Background with rounded effect (using filled rect)
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(width), float32(height), bgColor, false)

	// Border
	vector.StrokeRect(screen, float32(x), float32(y), float32(width), float32(height), 2, borderColor, false)

	// Text (wrapped)
	lines := nm.wrapText(message, width-2*nm.Padding)
	textY := y + nm.Padding + 13
	for _, line := range lines {
		text.Draw(screen, line, nm.Font, x+nm.Padding, textY, textColor)
		textY += 15
	}
}

// wrapText wraps text to fit within maxWidth (in pixels).
// Returns a slice of lines.
func (nm *NotificationManager) wrapText(message string, maxWidth int) []string {
	// Simple word wrapping based on character count
	// Approximate: 7 pixels per character with Face7x13
	maxChars := maxWidth / 7
	if maxChars < 10 {
		maxChars = 10
	}

	words := splitWords(message)
	lines := make([]string, 0)
	currentLine := ""

	for _, word := range words {
		testLine := currentLine
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) <= maxChars {
			currentLine = testLine
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

	if len(lines) == 0 {
		lines = append(lines, message) // Fallback: just use the original message
	}

	return lines
}

// splitWords splits a string into words (simple space-based splitting).
func splitWords(s string) []string {
	if s == "" {
		return []string{}
	}

	words := make([]string, 0)
	currentWord := ""

	for _, char := range s {
		if char == ' ' || char == '\t' || char == '\n' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(char)
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

// getBackgroundColor returns the background color for a notification type with alpha.
func (nm *NotificationManager) getBackgroundColor(notifType NotificationType, alpha uint8) color.Color {
	switch notifType {
	case NotificationInfo:
		return color.RGBA{40, 80, 140, alpha}
	case NotificationSuccess:
		return color.RGBA{40, 140, 80, alpha}
	case NotificationWarning:
		return color.RGBA{180, 120, 40, alpha}
	case NotificationError:
		return color.RGBA{160, 40, 40, alpha}
	default:
		return color.RGBA{60, 60, 80, alpha}
	}
}

// getBorderColor returns the border color for a notification type with alpha.
func (nm *NotificationManager) getBorderColor(notifType NotificationType, alpha uint8) color.Color {
	switch notifType {
	case NotificationInfo:
		return color.RGBA{80, 140, 220, alpha}
	case NotificationSuccess:
		return color.RGBA{80, 220, 140, alpha}
	case NotificationWarning:
		return color.RGBA{240, 180, 80, alpha}
	case NotificationError:
		return color.RGBA{240, 80, 80, alpha}
	default:
		return color.RGBA{120, 120, 160, alpha}
	}
}

// getTextColor returns the text color with alpha.
func (nm *NotificationManager) getTextColor(alpha uint8) color.Color {
	return color.RGBA{255, 255, 255, alpha}
}
