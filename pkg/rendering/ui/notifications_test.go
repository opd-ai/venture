package ui

import (
	"testing"
	"time"
)

func TestNewNotificationManager(t *testing.T) {
	nm := NewNotificationManager(10, 10, 300)

	if nm.X != 10 || nm.Y != 10 {
		t.Errorf("Expected position (10,10), got (%d,%d)", nm.X, nm.Y)
	}
	if nm.Width != 300 {
		t.Errorf("Expected width 300, got %d", nm.Width)
	}
	if nm.MaxNotifications != 5 {
		t.Errorf("Expected max 5 notifications, got %d", nm.MaxNotifications)
	}
	if nm.DefaultDuration != 3*time.Second {
		t.Errorf("Expected default duration 3s, got %v", nm.DefaultDuration)
	}
	if len(nm.Notifications) != 0 {
		t.Errorf("Expected 0 notifications initially, got %d", len(nm.Notifications))
	}
}

func TestAddNotification(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)

	nm.AddNotification(NotificationInfo, "Test message")

	if len(nm.Notifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(nm.Notifications))
	}

	notif := nm.Notifications[0]
	if notif.Type != NotificationInfo {
		t.Errorf("Expected type Info, got %v", notif.Type)
	}
	if notif.Message != "Test message" {
		t.Errorf("Expected message 'Test message', got '%s'", notif.Message)
	}
	if notif.Duration != 3*time.Second {
		t.Errorf("Expected duration 3s, got %v", notif.Duration)
	}
}

func TestAddConvenienceMethods(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)

	nm.AddInfo("Info message")
	nm.AddSuccess("Success message")
	nm.AddWarning("Warning message")
	nm.AddError("Error message")

	if len(nm.Notifications) != 4 {
		t.Fatalf("Expected 4 notifications, got %d", len(nm.Notifications))
	}

	types := []NotificationType{NotificationInfo, NotificationSuccess, NotificationWarning, NotificationError}
	for i, expectedType := range types {
		if nm.Notifications[i].Type != expectedType {
			t.Errorf("Notification %d: expected type %v, got %v", i, expectedType, nm.Notifications[i].Type)
		}
	}
}

func TestAddCustomDuration(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)

	customDuration := 10 * time.Second
	nm.AddCustomDuration(NotificationInfo, "Custom duration message", customDuration)

	if len(nm.Notifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(nm.Notifications))
	}

	notif := nm.Notifications[0]
	if notif.Duration != customDuration {
		t.Errorf("Expected duration %v, got %v", customDuration, notif.Duration)
	}
}

func TestMaxNotificationsLimit(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)
	nm.MaxNotifications = 3

	// Add more notifications than the limit
	for i := 0; i < 5; i++ {
		nm.AddInfo("Message")
	}

	if len(nm.Notifications) != 3 {
		t.Errorf("Expected 3 notifications (limit), got %d", len(nm.Notifications))
	}
}

func TestUpdateRemovesExpired(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)

	// Add notification with short duration
	nm.AddCustomDuration(NotificationInfo, "Short message", 50*time.Millisecond)

	if len(nm.Notifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(nm.Notifications))
	}

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)
	nm.Update(0.1)

	if len(nm.Notifications) != 0 {
		t.Errorf("Expected 0 notifications after expiry, got %d", len(nm.Notifications))
	}
}

func TestUpdateKeepsActive(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)

	// Add notification with long duration
	nm.AddCustomDuration(NotificationInfo, "Long message", 10*time.Second)

	if len(nm.Notifications) != 1 {
		t.Fatalf("Expected 1 notification, got %d", len(nm.Notifications))
	}

	// Update immediately (should still be active)
	nm.Update(0.016)

	if len(nm.Notifications) != 1 {
		t.Errorf("Expected 1 notification still active, got %d", len(nm.Notifications))
	}
}

func TestClear(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)

	nm.AddInfo("Message 1")
	nm.AddInfo("Message 2")
	nm.AddInfo("Message 3")

	if len(nm.Notifications) != 3 {
		t.Fatalf("Expected 3 notifications, got %d", len(nm.Notifications))
	}

	nm.Clear()

	if len(nm.Notifications) != 0 {
		t.Errorf("Expected 0 notifications after Clear, got %d", len(nm.Notifications))
	}
}

func TestGetCount(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)

	if nm.GetCount() != 0 {
		t.Errorf("Expected count 0, got %d", nm.GetCount())
	}

	nm.AddInfo("Message 1")
	if nm.GetCount() != 1 {
		t.Errorf("Expected count 1, got %d", nm.GetCount())
	}

	nm.AddInfo("Message 2")
	if nm.GetCount() != 2 {
		t.Errorf("Expected count 2, got %d", nm.GetCount())
	}

	nm.Clear()
	if nm.GetCount() != 0 {
		t.Errorf("Expected count 0 after Clear, got %d", nm.GetCount())
	}
}

func TestNotificationIsExpired(t *testing.T) {
	// Not expired
	notif := Notification{
		Type:      NotificationInfo,
		Message:   "Test",
		CreatedAt: time.Now(),
		Duration:  1 * time.Second,
	}

	if notif.IsExpired() {
		t.Error("Expected notification not to be expired immediately")
	}

	// Expired
	expiredNotif := Notification{
		Type:      NotificationInfo,
		Message:   "Test",
		CreatedAt: time.Now().Add(-2 * time.Second),
		Duration:  1 * time.Second,
	}

	if !expiredNotif.IsExpired() {
		t.Error("Expected notification to be expired")
	}
}

func TestNotificationGetAlpha(t *testing.T) {
	// Test fade-in
	fadeInNotif := Notification{
		Type:      NotificationInfo,
		Message:   "Test",
		CreatedAt: time.Now().Add(-100 * time.Millisecond), // 100ms elapsed
		Duration:  3 * time.Second,
	}
	alpha := fadeInNotif.GetAlpha()
	// Should be fading in (100ms / 200ms = 0.5, so alpha ~= 127)
	if alpha < 100 || alpha > 150 {
		t.Errorf("Expected alpha around 127 during fade-in, got %d", alpha)
	}

	// Test fully opaque (middle of duration)
	opaqueNotif := Notification{
		Type:      NotificationInfo,
		Message:   "Test",
		CreatedAt: time.Now().Add(-1 * time.Second), // 1s elapsed
		Duration:  3 * time.Second,
	}
	alpha = opaqueNotif.GetAlpha()
	if alpha != 255 {
		t.Errorf("Expected alpha 255 (fully opaque), got %d", alpha)
	}

	// Test fade-out
	fadeOutNotif := Notification{
		Type:      NotificationInfo,
		Message:   "Test",
		CreatedAt: time.Now().Add(-2750 * time.Millisecond), // 250ms remaining (within 500ms fade-out)
		Duration:  3 * time.Second,
	}
	alpha = fadeOutNotif.GetAlpha()
	// Should be fading out (250ms / 500ms = 0.5, so alpha ~= 127)
	if alpha < 100 || alpha > 150 {
		t.Errorf("Expected alpha around 127 during fade-out, got %d", alpha)
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", []string{}},
		{"word", []string{"word"}},
		{"hello world", []string{"hello", "world"}},
		{"one two three", []string{"one", "two", "three"}},
		{"  extra   spaces  ", []string{"extra", "spaces"}},
		{"tab\tseparated\twords", []string{"tab", "separated", "words"}},
		{"newline\nseparated", []string{"newline", "separated"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := splitWords(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d words, got %d", len(tt.expected), len(result))
				return
			}
			for i := range result {
				if result[i] != tt.expected[i] {
					t.Errorf("Word %d: expected '%s', got '%s'", i, tt.expected[i], result[i])
				}
			}
		})
	}
}

func TestNotificationWrapText(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)

	tests := []struct {
		message  string
		maxWidth int
		minLines int
		maxLines int
	}{
		{"Short", 300, 1, 1},
		{"This is a longer message that should wrap", 150, 2, 3},
		{"Word", 50, 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			lines := nm.wrapText(tt.message, tt.maxWidth)
			if len(lines) < tt.minLines || len(lines) > tt.maxLines {
				t.Errorf("Expected %d-%d lines, got %d", tt.minLines, tt.maxLines, len(lines))
			}
		})
	}
}

func TestGetBackgroundColor(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)

	types := []NotificationType{NotificationInfo, NotificationSuccess, NotificationWarning, NotificationError}
	for _, notifType := range types {
		// Just verify it doesn't panic and returns a color
		color := nm.getBackgroundColor(notifType, 255)
		if color == nil {
			t.Errorf("Expected non-nil color for type %v", notifType)
		}
	}
}

func TestGetBorderColor(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)

	types := []NotificationType{NotificationInfo, NotificationSuccess, NotificationWarning, NotificationError}
	for _, notifType := range types {
		color := nm.getBorderColor(notifType, 255)
		if color == nil {
			t.Errorf("Expected non-nil color for type %v", notifType)
		}
	}
}

func TestGetTextColor(t *testing.T) {
	nm := NewNotificationManager(0, 0, 300)

	// Test a few alpha values without loops
	testAlphas := []uint8{0, 50, 100, 150, 200, 255}
	for _, alpha := range testAlphas {
		color := nm.getTextColor(alpha)
		if color == nil {
			t.Errorf("Expected non-nil text color for alpha %d", alpha)
		}
	}
}

// Benchmarks

func BenchmarkAddNotification(b *testing.B) {
	nm := NewNotificationManager(0, 0, 300)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nm.AddNotification(NotificationInfo, "Test message")
	}
}

func BenchmarkNotificationUpdate(b *testing.B) {
	nm := NewNotificationManager(0, 0, 300)
	for i := 0; i < 5; i++ {
		nm.AddInfo("Test message")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		nm.Update(0.016)
	}
}

func BenchmarkGetAlpha(b *testing.B) {
	notif := Notification{
		Type:      NotificationInfo,
		Message:   "Test",
		CreatedAt: time.Now(),
		Duration:  3 * time.Second,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = notif.GetAlpha()
	}
}

func BenchmarkWrapText(b *testing.B) {
	nm := NewNotificationManager(0, 0, 300)
	message := "This is a longer message that should wrap to multiple lines based on the available width"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = nm.wrapText(message, 300)
	}
}

func BenchmarkSplitWords(b *testing.B) {
	message := "This is a test message with several words to split"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = splitWords(message)
	}
}



