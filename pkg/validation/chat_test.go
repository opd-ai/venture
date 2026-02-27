package validation

import (
	"strings"
	"testing"
)

func TestChatValidator_ValidateMessage(t *testing.T) {
	validator := NewChatValidator()

	tests := []struct {
		name    string
		message string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid message",
			message: "Hello world!",
			wantErr: false,
		},
		{
			name:    "empty message",
			message: "",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "whitespace only",
			message: "   ",
			wantErr: true,
			errMsg:  "cannot be empty",
		},
		{
			name:    "message too long",
			message: strings.Repeat("a", MaxChatMessageLength+1),
			wantErr: true,
			errMsg:  "too long",
		},
		{
			name:    "max length message",
			message: strings.Repeat("a", MaxChatMessageLength),
			wantErr: false,
		},
		{
			name:    "profanity detected",
			message: "This contains badword1 which is filtered",
			wantErr: true,
			errMsg:  "inappropriate content",
		},
		{
			name:    "profanity case insensitive",
			message: "This contains BADWORD1 which is filtered",
			wantErr: true,
			errMsg:  "inappropriate content",
		},
		{
			name:    "unicode message",
			message: "Hello 世界! 🎮",
			wantErr: false,
		},
		{
			name:    "long unicode message",
			message: strings.Repeat("世", MaxChatMessageLength),
			wantErr: false,
		},
		{
			name:    "too long unicode message",
			message: strings.Repeat("世", MaxChatMessageLength+1),
			wantErr: true,
			errMsg:  "too long",
		},
		{
			name:    "oversized UTF-8 payload - 4-byte emoji under rune limit",
			message: strings.Repeat("🎮", 500), // 500 runes × 4 bytes = 2000 bytes
			wantErr: false,                    // Exactly at byte limit
		},
		{
			name:    "oversized UTF-8 payload - exceeds byte limit",
			message: strings.Repeat("🎮", 501), // 501 runes × 4 bytes = 2004 bytes
			wantErr: true,
			errMsg:  "too large",
		},
		{
			name:    "max byte length with 1-byte chars",
			message: strings.Repeat("a", MaxChatMessageBytes),
			wantErr: true, // Exceeds rune limit (2000 > 500)
			errMsg:  "too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateMessage(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateMessage() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestChatValidator_SanitizeMessage(t *testing.T) {
	validator := NewChatValidator()

	tests := []struct {
		name     string
		message  string
		expected string
	}{
		{
			name:     "no sanitization needed",
			message:  "Hello world!",
			expected: "Hello world!",
		},
		{
			name:     "remove HTML tags",
			message:  "Hello <script>alert('xss')</script> world",
			expected: "Hello alert('xss') world",
		},
		{
			name:     "remove control characters",
			message:  "Hello\x00\x01world",
			expected: "Helloworld",
		},
		{
			name:     "normalize whitespace",
			message:  "Hello    world   !",
			expected: "Hello world !",
		},
		{
			name:     "trim whitespace",
			message:  "  Hello world  ",
			expected: "Hello world",
		},
		{
			name:     "multiple HTML tags",
			message:  "<b>Bold</b> <i>Italic</i> <u>Underline</u>",
			expected: "Bold Italic Underline",
		},
		{
			name:     "nested HTML tags",
			message:  "<div><span>Test</span></div>",
			expected: "Test",
		},
		{
			name:     "control chars and HTML",
			message:  "<script>\x00\x1F</script>Clean text",
			expected: "Clean text",
		},
		{
			name:     "unicode preserved",
			message:  "Hello 世界! 🎮",
			expected: "Hello 世界! 🎮",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeMessage(tt.message)
			if result != tt.expected {
				t.Errorf("SanitizeMessage() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestChatValidator_ValidateAndSanitize(t *testing.T) {
	validator := NewChatValidator()

	tests := []struct {
		name     string
		message  string
		wantErr  bool
		expected string
	}{
		{
			name:     "valid message sanitized",
			message:  "  Hello world!  ",
			wantErr:  false,
			expected: "Hello world!",
		},
		{
			name:     "HTML removed and validated",
			message:  "<b>Hello</b> world",
			wantErr:  false,
			expected: "Hello world",
		},
		{
			name:    "empty after sanitization",
			message: "<div></div>",
			wantErr: true,
		},
		{
			name:    "too long after sanitization",
			message: strings.Repeat("a", MaxChatMessageLength+10),
			wantErr: true,
		},
		{
			name:    "profanity after sanitization",
			message: "  badword1  ",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateAndSanitize(tt.message)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAndSanitize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("ValidateAndSanitize() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestChatValidator_containsProfanity(t *testing.T) {
	validator := NewChatValidator()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "no profanity",
			message: "Hello world",
			want:    false,
		},
		{
			name:    "profanity exact match",
			message: "badword1",
			want:    true,
		},
		{
			name:    "profanity in sentence",
			message: "This is badword1 in text",
			want:    true,
		},
		{
			name:    "profanity case insensitive",
			message: "BADWORD1",
			want:    true,
		},
		{
			name:    "profanity with punctuation",
			message: "badword1!",
			want:    true,
		},
		{
			name:    "profanity embedded in word",
			message: "mybadword1here",
			want:    true,
		},
		{
			name:    "multiple profanity",
			message: "badword1 and badword2",
			want:    true,
		},
		{
			name:    "clean text similar to profanity",
			message: "goodword1",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.containsProfanity(tt.message)
			if result != tt.want {
				t.Errorf("containsProfanity() = %v, want %v", result, tt.want)
			}
		})
	}
}

func BenchmarkChatValidator_ValidateMessage(b *testing.B) {
	validator := NewChatValidator()
	message := "Hello world! This is a typical chat message."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.ValidateMessage(message)
	}
}

func BenchmarkChatValidator_SanitizeMessage(b *testing.B) {
	validator := NewChatValidator()
	message := "<b>Hello</b> <i>world</i>! This is a message with HTML."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validator.SanitizeMessage(message)
	}
}

func BenchmarkChatValidator_ValidateAndSanitize(b *testing.B) {
	validator := NewChatValidator()
	message := "  <b>Hello</b> world!  "

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = validator.ValidateAndSanitize(message)
	}
}

func TestChatValidator_SanitizeMessageWithURLFilter(t *testing.T) {
	validator := NewChatValidator()

	tests := []struct {
		name       string
		message    string
		filterURLs bool
		expected   string
	}{
		{
			name:       "no URL, filter disabled",
			message:    "Hello world",
			filterURLs: false,
			expected:   "Hello world",
		},
		{
			name:       "no URL, filter enabled",
			message:    "Hello world",
			filterURLs: true,
			expected:   "Hello world",
		},
		{
			name:       "URL present, filter disabled",
			message:    "Check out http://example.com",
			filterURLs: false,
			expected:   "Check out http://example.com",
		},
		{
			name:       "URL present, filter enabled",
			message:    "Check out http://example.com",
			filterURLs: true,
			expected:   "Check out [link removed]",
		},
		{
			name:       "HTTPS URL, filter enabled",
			message:    "Visit https://secure.example.com/path",
			filterURLs: true,
			expected:   "Visit [link removed]",
		},
		{
			name:       "multiple URLs, filter enabled",
			message:    "See http://a.com and https://b.com",
			filterURLs: true,
			expected:   "See [link removed] and [link removed]",
		},
		{
			name:       "URL with query string",
			message:    "Link: http://example.com/page?id=123&name=test",
			filterURLs: true,
			expected:   "Link: [link removed]",
		},
		{
			name:       "HTML and URL combined",
			message:    "<b>Check</b> http://example.com",
			filterURLs: true,
			expected:   "Check [link removed]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.SanitizeMessageWithURLFilter(tt.message, tt.filterURLs)
			if result != tt.expected {
				t.Errorf("SanitizeMessageWithURLFilter() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestChatValidator_ContainsURL(t *testing.T) {
	validator := NewChatValidator()

	tests := []struct {
		name    string
		message string
		want    bool
	}{
		{
			name:    "no URL",
			message: "Hello world",
			want:    false,
		},
		{
			name:    "HTTP URL",
			message: "Visit http://example.com",
			want:    true,
		},
		{
			name:    "HTTPS URL",
			message: "Go to https://secure.example.com",
			want:    true,
		},
		{
			name:    "URL only",
			message: "https://example.com/path",
			want:    true,
		},
		{
			name:    "partial URL-like (no scheme)",
			message: "Visit example.com",
			want:    false,
		},
		{
			name:    "ftp URL (not matched)",
			message: "ftp://files.example.com",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.ContainsURL(tt.message)
			if result != tt.want {
				t.Errorf("ContainsURL() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestNewChatValidatorWithConfig(t *testing.T) {
	tests := []struct {
		name         string
		config       ChatValidatorConfig
		testMessage  string
		wantFiltered bool
		description  string
	}{
		{
			name: "custom profanity list",
			config: ChatValidatorConfig{
				CustomProfanityList: map[string]bool{
					"customword": true,
					"testbad":    true,
				},
			},
			testMessage:  "This contains customword which is filtered",
			wantFiltered: true,
			description:  "should filter custom profanity",
		},
		{
			name: "empty custom list falls back to default",
			config: ChatValidatorConfig{
				CustomProfanityList: map[string]bool{},
			},
			testMessage:  "This contains badword1 from default list",
			wantFiltered: true,
			description:  "should fall back to default profanity list",
		},
		{
			name: "nil custom list falls back to default",
			config: ChatValidatorConfig{
				CustomProfanityList: nil,
			},
			testMessage:  "This contains badword1 from default list",
			wantFiltered: true,
			description:  "should fall back to default profanity list",
		},
		{
			name: "custom list does not filter default words",
			config: ChatValidatorConfig{
				CustomProfanityList: map[string]bool{
					"onlythisword": true,
				},
			},
			testMessage:  "This contains badword1 but not in custom list",
			wantFiltered: false,
			description:  "should only filter custom words, not default",
		},
		{
			name: "case insensitive custom filtering",
			config: ChatValidatorConfig{
				CustomProfanityList: map[string]bool{
					"testword": true,
				},
			},
			testMessage:  "This contains TESTWORD in uppercase",
			wantFiltered: true,
			description:  "should be case insensitive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewChatValidatorWithConfig(tt.config)

			err := validator.ValidateMessage(tt.testMessage)
			hasError := err != nil

			if tt.wantFiltered && !hasError {
				t.Errorf("%s: expected profanity to be filtered but validation passed", tt.description)
			}
			if !tt.wantFiltered && hasError {
				t.Errorf("%s: expected validation to pass but got error: %v", tt.description, err)
			}
		})
	}
}

func TestChatValidatorConfig_Integration(t *testing.T) {
	// Test production-like usage pattern
	productionProfanity := map[string]bool{
		"spam":      true,
		"scam":      true,
		"phishing":  true,
		"malicious": true,
	}

	config := ChatValidatorConfig{
		CustomProfanityList: productionProfanity,
	}

	validator := NewChatValidatorWithConfig(config)

	// Should filter production words
	tests := []struct {
		message string
		blocked bool
	}{
		{"Click this spam link", true},
		{"This is a scam offer", true},
		{"Phishing attempt here", true},
		{"Normal message", false},
		{"Clean conversation", false},
	}

	for _, tt := range tests {
		t.Run(tt.message, func(t *testing.T) {
			err := validator.ValidateMessage(tt.message)
			hasError := err != nil
			if tt.blocked && !hasError {
				t.Errorf("expected message to be blocked: %s", tt.message)
			}
			if !tt.blocked && hasError {
				t.Errorf("expected message to pass but got error: %v", err)
			}
		})
	}
}
