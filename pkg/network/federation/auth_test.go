package federation

import (
	"testing"
	"time"
)

func TestNewAuthManager(t *testing.T) {
	am := NewAuthManager()

	if am == nil {
		t.Fatal("NewAuthManager() returned nil")
	}

	if am.tokens == nil {
		t.Error("tokens map not initialized")
	}

	if am.nonces == nil {
		t.Error("nonces map not initialized")
	}

	if am.ttl != 1*time.Hour {
		t.Errorf("ttl = %v, want 1h", am.ttl)
	}

	if am.nonceTTL != 5*time.Minute {
		t.Errorf("nonceTTL = %v, want 5m", am.nonceTTL)
	}
}

func TestSetTTL(t *testing.T) {
	am := NewAuthManager()
	newTTL := 30 * time.Minute

	am.SetTTL(newTTL)

	if am.ttl != newTTL {
		t.Errorf("ttl = %v, want %v", am.ttl, newTTL)
	}
}

func TestCreateSessionToken(t *testing.T) {
	am := NewAuthManager()

	token, err := am.CreateSessionToken(123, "server-1")
	if err != nil {
		t.Fatalf("CreateSessionToken() error = %v", err)
	}

	if token == nil {
		t.Fatal("CreateSessionToken() returned nil")
	}

	if token.PlayerID != 123 {
		t.Errorf("PlayerID = %v, want 123", token.PlayerID)
	}

	if token.ServerID != "server-1" {
		t.Errorf("ServerID = %v, want server-1", token.ServerID)
	}

	if token.Token == "" {
		t.Error("Token is empty")
	}

	if token.Nonce == "" {
		t.Error("Nonce is empty")
	}

	if token.ExpiresAt <= time.Now().Unix() {
		t.Error("ExpiresAt is in the past")
	}
}

func TestCreateSessionToken_UniqueTokens(t *testing.T) {
	am := NewAuthManager()

	token1, _ := am.CreateSessionToken(123, "server-1")
	token2, _ := am.CreateSessionToken(123, "server-1")

	if token1.Token == token2.Token {
		t.Error("Tokens should be unique")
	}

	if token1.Nonce == token2.Nonce {
		t.Error("Nonces should be unique")
	}
}

func TestValidateToken_Valid(t *testing.T) {
	am := NewAuthManager()

	created, _ := am.CreateSessionToken(123, "server-1")

	validated, err := am.ValidateToken(created.Token)
	if err != nil {
		t.Errorf("ValidateToken() error = %v", err)
	}

	if validated.PlayerID != 123 {
		t.Errorf("PlayerID = %v, want 123", validated.PlayerID)
	}
}

func TestValidateToken_NotFound(t *testing.T) {
	am := NewAuthManager()

	_, err := am.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	am := NewAuthManager()
	am.SetTTL(1 * time.Second)

	token, _ := am.CreateSessionToken(123, "server-1")

	time.Sleep(2 * time.Second)

	_, err := am.ValidateToken(token.Token)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestValidateTokenWithServer_Valid(t *testing.T) {
	am := NewAuthManager()

	created, _ := am.CreateSessionToken(123, "server-1")

	validated, err := am.ValidateTokenWithServer(created.Token, "server-1")
	if err != nil {
		t.Errorf("ValidateTokenWithServer() error = %v", err)
	}

	if validated.PlayerID != 123 {
		t.Errorf("PlayerID = %v, want 123", validated.PlayerID)
	}
}

func TestValidateTokenWithServer_ServerMismatch(t *testing.T) {
	am := NewAuthManager()

	created, _ := am.CreateSessionToken(123, "server-1")

	_, err := am.ValidateTokenWithServer(created.Token, "server-2")
	if err == nil {
		t.Error("Expected error for server mismatch, got nil")
	}
}

func TestRevokeToken(t *testing.T) {
	am := NewAuthManager()

	token, _ := am.CreateSessionToken(123, "server-1")

	err := am.RevokeToken(token.Token)
	if err != nil {
		t.Errorf("RevokeToken() error = %v", err)
	}

	// Should not be able to validate revoked token
	_, err = am.ValidateToken(token.Token)
	if err == nil {
		t.Error("Expected error for revoked token, got nil")
	}
}

func TestRevokeToken_NotFound(t *testing.T) {
	am := NewAuthManager()

	err := am.RevokeToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestValidateNonce(t *testing.T) {
	am := NewAuthManager()

	token, _ := am.CreateSessionToken(123, "server-1")

	err := am.ValidateNonce(token.Nonce)
	if err != nil {
		t.Errorf("ValidateNonce() error = %v", err)
	}
}

func TestValidateNonce_NotFound(t *testing.T) {
	am := NewAuthManager()

	err := am.ValidateNonce("invalid-nonce")
	if err == nil {
		t.Error("Expected error for invalid nonce, got nil")
	}
}

func TestValidateNonce_Expired(t *testing.T) {
	am := NewAuthManager()
	am.nonceTTL = 1 * time.Second

	token, _ := am.CreateSessionToken(123, "server-1")

	time.Sleep(2 * time.Second)

	err := am.ValidateNonce(token.Nonce)
	if err == nil {
		t.Error("Expected error for expired nonce, got nil")
	}
}

func TestMarkNonceUsed(t *testing.T) {
	am := NewAuthManager()

	token, _ := am.CreateSessionToken(123, "server-1")

	err := am.MarkNonceUsed(token.Nonce)
	if err != nil {
		t.Errorf("MarkNonceUsed() error = %v", err)
	}

	// Should not be able to validate used nonce
	err = am.ValidateNonce(token.Nonce)
	if err == nil {
		t.Error("Expected error for used nonce, got nil")
	}
}

func TestMarkNonceUsed_NotFound(t *testing.T) {
	am := NewAuthManager()

	err := am.MarkNonceUsed("invalid-nonce")
	if err == nil {
		t.Error("Expected error for invalid nonce, got nil")
	}
}

func TestCleanupExpired(t *testing.T) {
	am := NewAuthManager()
	am.SetTTL(1 * time.Second)
	am.nonceTTL = 1 * time.Second

	// Create some tokens
	am.CreateSessionToken(123, "server-1")
	am.CreateSessionToken(456, "server-2")

	time.Sleep(2 * time.Second)

	// Create one more token that won't be expired
	am.CreateSessionToken(789, "server-3")

	count := am.CleanupExpired()
	if count < 2 {
		t.Errorf("CleanupExpired() = %v, want at least 2", count)
	}

	activeTokens := am.GetActiveTokenCount()
	if activeTokens != 1 {
		t.Errorf("GetActiveTokenCount() = %v, want 1", activeTokens)
	}

	activeNonces := am.GetActiveNonceCount()
	if activeNonces != 1 {
		t.Errorf("GetActiveNonceCount() = %v, want 1", activeNonces)
	}
}

func TestGetActiveTokenCount(t *testing.T) {
	am := NewAuthManager()

	if am.GetActiveTokenCount() != 0 {
		t.Errorf("Initial token count = %v, want 0", am.GetActiveTokenCount())
	}

	am.CreateSessionToken(123, "server-1")
	am.CreateSessionToken(456, "server-2")

	if am.GetActiveTokenCount() != 2 {
		t.Errorf("Token count = %v, want 2", am.GetActiveTokenCount())
	}
}

func TestGetActiveNonceCount(t *testing.T) {
	am := NewAuthManager()

	if am.GetActiveNonceCount() != 0 {
		t.Errorf("Initial nonce count = %v, want 0", am.GetActiveNonceCount())
	}

	am.CreateSessionToken(123, "server-1")
	am.CreateSessionToken(456, "server-2")

	if am.GetActiveNonceCount() != 2 {
		t.Errorf("Nonce count = %v, want 2", am.GetActiveNonceCount())
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	tokens := make(map[string]bool)

	for i := 0; i < 100; i++ {
		token, err := generateToken()
		if err != nil {
			t.Fatalf("generateToken() error = %v", err)
		}

		if tokens[token] {
			t.Errorf("Duplicate token generated: %s", token)
		}

		tokens[token] = true
	}
}

func TestGenerateNonce_Unique(t *testing.T) {
	nonces := make(map[string]bool)

	for i := 0; i < 100; i++ {
		nonce, err := generateNonce()
		if err != nil {
			t.Fatalf("generateNonce() error = %v", err)
		}

		if nonces[nonce] {
			t.Errorf("Duplicate nonce generated: %s", nonce)
		}

		nonces[nonce] = true

		// Verify nonce is 32 hex characters (16 bytes)
		if len(nonce) != 32 {
			t.Errorf("Nonce length = %d, want 32", len(nonce))
		}
	}
}

// Benchmarks

func BenchmarkCreateSessionToken(b *testing.B) {
	am := NewAuthManager()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am.CreateSessionToken(uint64(i), "server-1")
	}
}

func BenchmarkValidateToken(b *testing.B) {
	am := NewAuthManager()
	token, _ := am.CreateSessionToken(123, "server-1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am.ValidateToken(token.Token)
	}
}

func BenchmarkValidateNonce(b *testing.B) {
	am := NewAuthManager()
	token, _ := am.CreateSessionToken(123, "server-1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		am.ValidateNonce(token.Nonce)
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateToken()
	}
}

func BenchmarkGenerateNonce(b *testing.B) {
	for i := 0; i < b.N; i++ {
		generateNonce()
	}
}
