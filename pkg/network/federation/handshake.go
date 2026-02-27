package federation

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/recovery"
	"github.com/opd-ai/venture/pkg/version"
)

// TrustLevel represents the trust relationship with a peer server
type TrustLevel int

const (
	// TrustUnknown indicates no prior interaction
	TrustUnknown TrustLevel = iota
	// TrustVerified indicates certificate exchange complete, limited features
	TrustVerified
	// TrustTrusted indicates known server, full feature access
	TrustTrusted
)

// String returns human-readable trust level name
func (t TrustLevel) String() string {
	switch t {
	case TrustUnknown:
		return "Unknown"
	case TrustVerified:
		return "Verified"
	case TrustTrusted:
		return "Trusted"
	default:
		return "Invalid"
	}
}

// FederationHandshake represents the initial connection setup between servers
type FederationHandshake struct {
	ServerID   string     // Public key fingerprint (hex encoded)
	ServerName string     // Human-readable name
	Version    string     // Protocol version (e.g., "6.0.0")
	Features   []string   // Supported features: ["travel", "trade", "post"]
	TrustLevel TrustLevel // Trust relationship
	PublicKey  []byte     // ed25519 public key (32 bytes)
	Signature  []byte     // Signature of handshake data
	Timestamp  int64      // Unix timestamp (milliseconds)
	Nonce      []byte     // Random nonce for replay prevention (16 bytes)
}

// ServerIdentity holds server keypair and metadata
type ServerIdentity struct {
	ServerID   string // Fingerprint of public key
	ServerName string // Human-readable name
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
	Created    time.Time
	mu         sync.RWMutex
}

// NewServerIdentity generates a new server identity with ed25519 keypair
func NewServerIdentity(serverName string) (*ServerIdentity, error) {
	if serverName == "" {
		return nil, fmt.Errorf("server name cannot be empty")
	}

	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("failed to generate keypair: %w", err)
	}

	fingerprint := generateFingerprint(pub)

	return &ServerIdentity{
		ServerID:   fingerprint,
		ServerName: serverName,
		PublicKey:  pub,
		PrivateKey: priv,
		Created:    time.Now(),
	}, nil
}

// generateFingerprint creates a hex-encoded SHA-256 hash of the public key
func generateFingerprint(pubKey ed25519.PublicKey) string {
	hash := sha256.Sum256(pubKey)
	return hex.EncodeToString(hash[:])
}

// GetFingerprint returns the server's public key fingerprint
func (si *ServerIdentity) GetFingerprint() string {
	si.mu.RLock()
	defer si.mu.RUnlock()
	return si.ServerID
}

// CreateHandshake creates a new handshake message
func (si *ServerIdentity) CreateHandshake(version string, features []string, trustLevel TrustLevel) (*FederationHandshake, error) {
	si.mu.RLock()
	defer si.mu.RUnlock()

	if version == "" {
		return nil, fmt.Errorf("version cannot be empty")
	}

	// Generate random nonce for replay prevention
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	handshake := &FederationHandshake{
		ServerID:   si.ServerID,
		ServerName: si.ServerName,
		Version:    version,
		Features:   features,
		TrustLevel: trustLevel,
		PublicKey:  si.PublicKey,
		Timestamp:  time.Now().UnixMilli(),
		Nonce:      nonce,
	}

	// Sign the handshake
	signature, err := si.signHandshake(handshake)
	if err != nil {
		return nil, fmt.Errorf("failed to sign handshake: %w", err)
	}

	handshake.Signature = signature

	return handshake, nil
}

// signHandshake creates a signature for the handshake data
func (si *ServerIdentity) signHandshake(h *FederationHandshake) ([]byte, error) {
	// Create message to sign (concatenate all fields)
	message := fmt.Sprintf("%s|%s|%s|%s|%d|%x|%x",
		h.ServerID,
		h.ServerName,
		h.Version,
		strings.Join(h.Features, ","),
		h.TrustLevel,
		h.Timestamp,
		h.Nonce,
	)

	signature := ed25519.Sign(si.PrivateKey, []byte(message))
	return signature, nil
}

// VerifyHandshake verifies the signature of a received handshake
func VerifyHandshake(h *FederationHandshake) error {
	if err := validateHandshakeFields(h); err != nil {
		return err
	}

	if err := validateCryptoFields(h); err != nil {
		return err
	}

	if err := verifyFingerprint(h); err != nil {
		return err
	}

	if err := verifySignature(h); err != nil {
		return err
	}

	return verifyTimestamp(h)
}

// validateHandshakeFields validates basic handshake fields.
func validateHandshakeFields(h *FederationHandshake) error {
	if h == nil {
		return fmt.Errorf("handshake is nil")
	}

	if h.ServerID == "" {
		return fmt.Errorf("server ID is empty")
	}

	if h.ServerName == "" {
		return fmt.Errorf("server name is empty")
	}

	if h.Version == "" {
		return fmt.Errorf("version is empty")
	}

	return nil
}

// validateCryptoFields validates cryptographic field sizes.
func validateCryptoFields(h *FederationHandshake) error {
	if len(h.PublicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("invalid public key size: %d (expected %d)", len(h.PublicKey), ed25519.PublicKeySize)
	}

	if len(h.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("invalid signature size: %d (expected %d)", len(h.Signature), ed25519.SignatureSize)
	}

	if len(h.Nonce) != 16 {
		return fmt.Errorf("invalid nonce size: %d (expected 16)", len(h.Nonce))
	}

	return nil
}

// verifyFingerprint verifies the fingerprint matches the public key.
func verifyFingerprint(h *FederationHandshake) error {
	expectedFingerprint := generateFingerprint(h.PublicKey)
	if h.ServerID != expectedFingerprint {
		return fmt.Errorf("server ID does not match public key fingerprint")
	}
	return nil
}

// verifySignature verifies the handshake signature.
func verifySignature(h *FederationHandshake) error {
	message := fmt.Sprintf("%s|%s|%s|%s|%d|%x|%x",
		h.ServerID,
		h.ServerName,
		h.Version,
		strings.Join(h.Features, ","),
		h.TrustLevel,
		h.Timestamp,
		h.Nonce,
	)

	if !ed25519.Verify(h.PublicKey, []byte(message), h.Signature) {
		return fmt.Errorf("signature verification failed")
	}

	return nil
}

// verifyTimestamp checks timestamp for replay attack protection.
func verifyTimestamp(h *FederationHandshake) error {
	now := time.Now().UnixMilli()
	age := now - h.Timestamp
	maxAge := int64(60000)

	if age < 0 {
		return fmt.Errorf("handshake timestamp is in the future")
	}

	if age > maxAge {
		return fmt.Errorf("handshake timestamp too old: %dms (max %dms)", age, maxAge)
	}

	return nil
}

// HandshakeManager manages handshake state and nonce tracking
type HandshakeManager struct {
	identity    *ServerIdentity
	seenNonces  map[string]time.Time // Track nonces to prevent replay
	mu          sync.RWMutex
	nonceExpiry time.Duration // How long to remember nonces (default: 5 minutes)
}

// NewHandshakeManager creates a new handshake manager
func NewHandshakeManager(identity *ServerIdentity) *HandshakeManager {
	return &HandshakeManager{
		identity:    identity,
		seenNonces:  make(map[string]time.Time),
		nonceExpiry: 5 * time.Minute,
	}
}

// ProcessHandshake validates and processes an incoming handshake
func (hm *HandshakeManager) ProcessHandshake(h *FederationHandshake) error {
	// Verify handshake signature and fields
	if err := VerifyHandshake(h); err != nil {
		return fmt.Errorf("handshake verification failed: %w", err)
	}

	// Check for replay attacks (nonce reuse)
	nonceKey := hex.EncodeToString(h.Nonce)
	hm.mu.Lock()
	if lastSeen, exists := hm.seenNonces[nonceKey]; exists {
		hm.mu.Unlock()
		return fmt.Errorf("nonce already seen at %s (replay attack)", lastSeen.Format(time.RFC3339))
	}
	hm.seenNonces[nonceKey] = time.Now()
	hm.mu.Unlock()

	// Cleanup old nonces (async to avoid blocking)
	go hm.cleanupNonces()

	return nil
}

// cleanupNonces removes expired nonces from memory
func (hm *HandshakeManager) cleanupNonces() {
	defer recovery.RecoverPanicWithLogger("federation_handshake", "cleanup nonces", nil)()

	hm.mu.Lock()
	defer hm.mu.Unlock()

	now := time.Now()
	for nonce, seenAt := range hm.seenNonces {
		if now.Sub(seenAt) > hm.nonceExpiry {
			delete(hm.seenNonces, nonce)
		}
	}
}

// CreateResponse creates a handshake response to a received handshake
func (hm *HandshakeManager) CreateResponse(peerHandshake *FederationHandshake, trustLevel TrustLevel) (*FederationHandshake, error) {
	// Validate peer handshake first
	if err := hm.ProcessHandshake(peerHandshake); err != nil {
		return nil, fmt.Errorf("cannot respond to invalid handshake: %w", err)
	}

	// Create response with same version and features
	return hm.identity.CreateHandshake(peerHandshake.Version, peerHandshake.Features, trustLevel)
}

// IsCompatibleVersion checks if protocol versions are compatible.
// Compatible versions share the same major version number.
// This is a convenience wrapper around version.IsCompatible for federation protocol use.
// Returns an error if either version string is invalid.
func IsCompatibleVersion(ourVersion, theirVersion string) (bool, error) {
	return version.IsCompatible(ourVersion, theirVersion)
}

// NegotiateFeatures returns the intersection of supported features
func NegotiateFeatures(ourFeatures, theirFeatures []string) []string {
	featureSet := make(map[string]bool)
	for _, f := range theirFeatures {
		featureSet[f] = true
	}

	var common []string
	for _, f := range ourFeatures {
		if featureSet[f] {
			common = append(common, f)
		}
	}

	return common
}
