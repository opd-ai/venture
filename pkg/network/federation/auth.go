package federation

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// SessionToken represents a player session for authentication
type SessionToken struct {
	PlayerID  uint64
	Token     string
	ServerID  string
	ExpiresAt int64 // Unix timestamp
	Nonce     string
}

// AuthManager manages player authentication for transfers
type AuthManager struct {
	mu       sync.RWMutex
	tokens   map[string]*SessionToken // token -> SessionToken
	nonces   map[string]int64         // nonce -> timestamp
	ttl      time.Duration
	nonceTTL time.Duration
}

// NewAuthManager creates a new authentication manager
func NewAuthManager() *AuthManager {
	return &AuthManager{
		tokens:   make(map[string]*SessionToken),
		nonces:   make(map[string]int64),
		ttl:      1 * time.Hour,
		nonceTTL: 5 * time.Minute,
	}
}

// SetTTL sets the token time-to-live
func (am *AuthManager) SetTTL(ttl time.Duration) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.ttl = ttl
}

// CreateSessionToken creates a new session token for a player
func (am *AuthManager) CreateSessionToken(playerID uint64, serverID string) (*SessionToken, error) {
	token, err := generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	nonce, err := generateNonce()
	if err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	sessionToken := &SessionToken{
		PlayerID:  playerID,
		Token:     token,
		ServerID:  serverID,
		ExpiresAt: time.Now().Add(am.ttl).Unix(),
		Nonce:     nonce,
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	am.tokens[token] = sessionToken
	am.nonces[nonce] = time.Now().Unix()

	return sessionToken, nil
}

// ValidateToken validates a session token
func (am *AuthManager) ValidateToken(token string) (*SessionToken, error) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	sessionToken, exists := am.tokens[token]
	if !exists {
		return nil, fmt.Errorf("token not found")
	}

	if time.Now().Unix() > sessionToken.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}

	return sessionToken, nil
}

// ValidateTokenWithServer validates token and verifies server match
func (am *AuthManager) ValidateTokenWithServer(token, expectedServerID string) (*SessionToken, error) {
	sessionToken, err := am.ValidateToken(token)
	if err != nil {
		return nil, err
	}

	if sessionToken.ServerID != expectedServerID {
		return nil, fmt.Errorf("server ID mismatch: expected %s, got %s", expectedServerID, sessionToken.ServerID)
	}

	return sessionToken, nil
}

// RevokeToken revokes a session token
func (am *AuthManager) RevokeToken(token string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.tokens[token]; !exists {
		return fmt.Errorf("token not found")
	}

	delete(am.tokens, token)
	return nil
}

// ValidateNonce validates a nonce for replay attack prevention
func (am *AuthManager) ValidateNonce(nonce string) error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	timestamp, exists := am.nonces[nonce]
	if !exists {
		return fmt.Errorf("nonce not found")
	}

	// Check if nonce is expired
	if time.Now().Unix()-timestamp > int64(am.nonceTTL.Seconds()) {
		return fmt.Errorf("nonce expired")
	}

	return nil
}

// MarkNonceUsed marks a nonce as used (for replay prevention)
func (am *AuthManager) MarkNonceUsed(nonce string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.nonces[nonce]; !exists {
		return fmt.Errorf("nonce not found")
	}

	// Remove nonce to prevent reuse
	delete(am.nonces, nonce)
	return nil
}

// CleanupExpired removes expired tokens and nonces
func (am *AuthManager) CleanupExpired() int {
	am.mu.Lock()
	defer am.mu.Unlock()

	now := time.Now().Unix()
	count := 0

	// Clean up expired tokens
	for token, sessionToken := range am.tokens {
		if sessionToken.ExpiresAt < now {
			delete(am.tokens, token)
			count++
		}
	}

	// Clean up expired nonces
	nonceTTLSeconds := int64(am.nonceTTL.Seconds())
	for nonce, timestamp := range am.nonces {
		if now-timestamp > nonceTTLSeconds {
			delete(am.nonces, nonce)
			count++
		}
	}

	return count
}

// GetActiveTokenCount returns the number of active tokens
func (am *AuthManager) GetActiveTokenCount() int {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return len(am.tokens)
}

// GetActiveNonceCount returns the number of active nonces
func (am *AuthManager) GetActiveNonceCount() int {
	am.mu.RLock()
	defer am.mu.RUnlock()
	return len(am.nonces)
}

// generateToken generates a random UUID v4 token
func generateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Set version (4) and variant bits
	bytes[6] = (bytes[6] & 0x0f) | 0x40 // Version 4
	bytes[8] = (bytes[8] & 0x3f) | 0x80 // Variant is 10

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		bytes[0:4],
		bytes[4:6],
		bytes[6:8],
		bytes[8:10],
		bytes[10:16],
	), nil
}

// generateNonce generates a random 16-byte nonce
func generateNonce() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
