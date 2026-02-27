package federation

import (
	"crypto/ed25519"
	"strings"
	"testing"
	"time"
)

func TestTrustLevelString(t *testing.T) {
	tests := []struct {
		level    TrustLevel
		expected string
	}{
		{TrustUnknown, "Unknown"},
		{TrustVerified, "Verified"},
		{TrustTrusted, "Trusted"},
		{TrustLevel(99), "Invalid"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.expected {
			t.Errorf("TrustLevel(%d).String() = %s, want %s", tt.level, got, tt.expected)
		}
	}
}

func TestNewServerIdentity(t *testing.T) {
	tests := []struct {
		name      string
		wantErr   bool
		errSubstr string
	}{
		{"TestServer", false, ""},
		{"", true, "server name cannot be empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity, err := NewServerIdentity(tt.name)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewServerIdentity() expected error containing %q, got nil", tt.errSubstr)
				} else if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("NewServerIdentity() error = %v, want substring %q", err, tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Fatalf("NewServerIdentity() unexpected error: %v", err)
			}

			if identity.ServerName != tt.name {
				t.Errorf("ServerName = %s, want %s", identity.ServerName, tt.name)
			}

			if len(identity.PublicKey) != ed25519.PublicKeySize {
				t.Errorf("PublicKey size = %d, want %d", len(identity.PublicKey), ed25519.PublicKeySize)
			}

			if len(identity.PrivateKey) != ed25519.PrivateKeySize {
				t.Errorf("PrivateKey size = %d, want %d", len(identity.PrivateKey), ed25519.PrivateKeySize)
			}

			if identity.ServerID == "" {
				t.Error("ServerID is empty")
			}

			if len(identity.ServerID) != 64 { // SHA-256 hex = 64 chars
				t.Errorf("ServerID length = %d, want 64", len(identity.ServerID))
			}
		})
	}
}

func TestServerIdentityGetFingerprint(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("NewServerIdentity() error: %v", err)
	}

	fp := identity.GetFingerprint()
	if fp != identity.ServerID {
		t.Errorf("GetFingerprint() = %s, want %s", fp, identity.ServerID)
	}
}

func TestCreateHandshake(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("NewServerIdentity() error: %v", err)
	}

	tests := []struct {
		name      string
		version   string
		features  []string
		trust     TrustLevel
		wantErr   bool
		errSubstr string
	}{
		{"valid", "6.0.0", []string{"travel", "trade"}, TrustVerified, false, ""},
		{"empty version", "", []string{"travel"}, TrustVerified, true, "version cannot be empty"},
		{"no features", "6.0.0", []string{}, TrustTrusted, false, ""},
		{"nil features", "6.0.0", nil, TrustUnknown, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handshake, err := identity.CreateHandshake(tt.version, tt.features, tt.trust)

			if tt.wantErr {
				if err == nil {
					t.Errorf("CreateHandshake() expected error containing %q, got nil", tt.errSubstr)
				} else if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("CreateHandshake() error = %v, want substring %q", err, tt.errSubstr)
				}
				return
			}

			if err != nil {
				t.Fatalf("CreateHandshake() unexpected error: %v", err)
			}

			if handshake.ServerID != identity.ServerID {
				t.Errorf("ServerID = %s, want %s", handshake.ServerID, identity.ServerID)
			}

			if handshake.ServerName != identity.ServerName {
				t.Errorf("ServerName = %s, want %s", handshake.ServerName, identity.ServerName)
			}

			if handshake.Version != tt.version {
				t.Errorf("Version = %s, want %s", handshake.Version, tt.version)
			}

			if handshake.TrustLevel != tt.trust {
				t.Errorf("TrustLevel = %v, want %v", handshake.TrustLevel, tt.trust)
			}

			if len(handshake.Signature) != ed25519.SignatureSize {
				t.Errorf("Signature size = %d, want %d", len(handshake.Signature), ed25519.SignatureSize)
			}

			if len(handshake.Nonce) != 16 {
				t.Errorf("Nonce size = %d, want 16", len(handshake.Nonce))
			}

			if handshake.Timestamp == 0 {
				t.Error("Timestamp is zero")
			}
		})
	}
}

func TestVerifyHandshake(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("NewServerIdentity() error: %v", err)
	}

	validHandshake, err := identity.CreateHandshake("6.0.0", []string{"travel"}, TrustVerified)
	if err != nil {
		t.Fatalf("CreateHandshake() error: %v", err)
	}

	// Create handshakes with future and old timestamps separately
	futureHandshake := &FederationHandshake{
		ServerID:   identity.ServerID,
		ServerName: identity.ServerName,
		Version:    "6.0.0",
		Features:   []string{"travel"},
		TrustLevel: TrustVerified,
		PublicKey:  identity.PublicKey,
		Timestamp:  time.Now().Add(time.Hour).UnixMilli(),
		Nonce:      make([]byte, 16),
	}
	futureHandshake.Signature, _ = identity.signHandshake(futureHandshake)

	oldHandshake := &FederationHandshake{
		ServerID:   identity.ServerID,
		ServerName: identity.ServerName,
		Version:    "6.0.0",
		Features:   []string{"travel"},
		TrustLevel: TrustVerified,
		PublicKey:  identity.PublicKey,
		Timestamp:  time.Now().Add(-2 * time.Minute).UnixMilli(),
		Nonce:      make([]byte, 16),
	}
	oldHandshake.Signature, _ = identity.signHandshake(oldHandshake)

	tests := []struct {
		name      string
		handshake *FederationHandshake
		modify    func(*FederationHandshake)
		wantErr   bool
		errSubstr string
	}{
		{"valid", validHandshake, nil, false, ""},
		{"nil handshake", nil, nil, true, "handshake is nil"},
		{
			"empty server ID",
			validHandshake,
			func(h *FederationHandshake) { h.ServerID = "" },
			true,
			"server ID is empty",
		},
		{
			"empty server name",
			validHandshake,
			func(h *FederationHandshake) { h.ServerName = "" },
			true,
			"server name is empty",
		},
		{
			"empty version",
			validHandshake,
			func(h *FederationHandshake) { h.Version = "" },
			true,
			"version is empty",
		},
		{
			"invalid public key size",
			validHandshake,
			func(h *FederationHandshake) { h.PublicKey = []byte{1, 2, 3} },
			true,
			"invalid public key size",
		},
		{
			"invalid signature size",
			validHandshake,
			func(h *FederationHandshake) { h.Signature = []byte{1, 2, 3} },
			true,
			"invalid signature size",
		},
		{
			"invalid nonce size",
			validHandshake,
			func(h *FederationHandshake) { h.Nonce = []byte{1, 2, 3} },
			true,
			"invalid nonce size",
		},
		{
			"mismatched fingerprint",
			validHandshake,
			func(h *FederationHandshake) {
				h.ServerID = "0000000000000000000000000000000000000000000000000000000000000000"
			},
			true,
			"does not match public key fingerprint",
		},
		{
			"invalid signature",
			validHandshake,
			func(h *FederationHandshake) {
				// Modify signature to make it invalid
				h.Signature[0] ^= 0xFF
			},
			true,
			"signature verification failed",
		},
		{
			"future timestamp",
			futureHandshake,
			nil,
			true,
			"timestamp is in the future",
		},
		{
			"old timestamp",
			oldHandshake,
			nil,
			true,
			"timestamp too old",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a copy of the handshake to avoid modifying original
			var h *FederationHandshake
			if tt.handshake != nil {
				hCopy := *tt.handshake
				// Deep copy slices
				hCopy.Features = append([]string(nil), tt.handshake.Features...)
				hCopy.PublicKey = append([]byte(nil), tt.handshake.PublicKey...)
				hCopy.Signature = append([]byte(nil), tt.handshake.Signature...)
				hCopy.Nonce = append([]byte(nil), tt.handshake.Nonce...)
				h = &hCopy
			}

			if tt.modify != nil {
				tt.modify(h)
			}

			err := VerifyHandshake(h)

			if tt.wantErr {
				if err == nil {
					t.Errorf("VerifyHandshake() expected error containing %q, got nil", tt.errSubstr)
				} else if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Errorf("VerifyHandshake() error = %v, want substring %q", err, tt.errSubstr)
				}
			} else {
				if err != nil {
					t.Errorf("VerifyHandshake() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestHandshakeManagerProcessHandshake(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("NewServerIdentity() error: %v", err)
	}

	manager := NewHandshakeManager(identity)

	handshake, err := identity.CreateHandshake("6.0.0", []string{"travel"}, TrustVerified)
	if err != nil {
		t.Fatalf("CreateHandshake() error: %v", err)
	}

	// First processing should succeed
	err = manager.ProcessHandshake(handshake)
	if err != nil {
		t.Errorf("ProcessHandshake() unexpected error: %v", err)
	}

	// Second processing with same nonce should fail (replay attack)
	err = manager.ProcessHandshake(handshake)
	if err == nil {
		t.Error("ProcessHandshake() expected replay attack error, got nil")
	} else if !strings.Contains(err.Error(), "nonce already seen") {
		t.Errorf("ProcessHandshake() error = %v, want replay attack error", err)
	}

	// New handshake with different nonce should succeed
	handshake2, err := identity.CreateHandshake("6.0.0", []string{"trade"}, TrustTrusted)
	if err != nil {
		t.Fatalf("CreateHandshake() error: %v", err)
	}

	err = manager.ProcessHandshake(handshake2)
	if err != nil {
		t.Errorf("ProcessHandshake() with new nonce unexpected error: %v", err)
	}
}

func TestHandshakeManagerCreateResponse(t *testing.T) {
	serverA, err := NewServerIdentity("ServerA")
	if err != nil {
		t.Fatalf("NewServerIdentity() error: %v", err)
	}

	serverB, err := NewServerIdentity("ServerB")
	if err != nil {
		t.Fatalf("NewServerIdentity() error: %v", err)
	}

	managerA := NewHandshakeManager(serverA)

	// Server B creates initial handshake
	handshakeB, err := serverB.CreateHandshake("6.0.0", []string{"travel", "trade"}, TrustVerified)
	if err != nil {
		t.Fatalf("CreateHandshake() error: %v", err)
	}

	// Server A creates response
	responseA, err := managerA.CreateResponse(handshakeB, TrustTrusted)
	if err != nil {
		t.Fatalf("CreateResponse() error: %v", err)
	}

	// Verify response
	if responseA.ServerID != serverA.ServerID {
		t.Errorf("Response ServerID = %s, want %s", responseA.ServerID, serverA.ServerID)
	}

	if responseA.Version != handshakeB.Version {
		t.Errorf("Response Version = %s, want %s", responseA.Version, handshakeB.Version)
	}

	// Features should match peer's features
	if len(responseA.Features) != len(handshakeB.Features) {
		t.Errorf("Response Features length = %d, want %d", len(responseA.Features), len(handshakeB.Features))
	}

	// Verify signature
	err = VerifyHandshake(responseA)
	if err != nil {
		t.Errorf("VerifyHandshake(response) error: %v", err)
	}
}

func TestIsCompatibleVersion(t *testing.T) {
	tests := []struct {
		ourVersion   string
		theirVersion string
		compatible   bool
		wantErr      bool
	}{
		{"6.0.0", "6.0.0", true, false},
		{"6.0.0", "6.1.0", true, false},
		{"6.0.0", "6.0.1", true, false},
		{"6.0.0", "5.0.0", false, false},
		{"6.0.0", "7.0.0", false, false},
		{"", "6.0.0", false, true},
		{"6.0.0", "", false, true},
		{"", "", false, true},
		{"invalid", "6.0.0", false, true},
	}

	for _, tt := range tests {
		got, err := IsCompatibleVersion(tt.ourVersion, tt.theirVersion)
		if (err != nil) != tt.wantErr {
			t.Errorf("IsCompatibleVersion(%q, %q) error = %v, wantErr %v", tt.ourVersion, tt.theirVersion, err, tt.wantErr)
			continue
		}
		if got != tt.compatible {
			t.Errorf("IsCompatibleVersion(%q, %q) = %v, want %v", tt.ourVersion, tt.theirVersion, got, tt.compatible)
		}
	}
}

func TestNegotiateFeatures(t *testing.T) {
	tests := []struct {
		name          string
		ourFeatures   []string
		theirFeatures []string
		expected      []string
	}{
		{
			"full match",
			[]string{"travel", "trade"},
			[]string{"travel", "trade"},
			[]string{"travel", "trade"},
		},
		{
			"partial match",
			[]string{"travel", "trade", "post"},
			[]string{"travel", "post"},
			[]string{"travel", "post"},
		},
		{
			"no match",
			[]string{"travel"},
			[]string{"trade"},
			nil,
		},
		{
			"empty ours",
			[]string{},
			[]string{"travel"},
			nil,
		},
		{
			"empty theirs",
			[]string{"travel"},
			[]string{},
			nil,
		},
		{
			"both empty",
			[]string{},
			[]string{},
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NegotiateFeatures(tt.ourFeatures, tt.theirFeatures)

			if len(got) != len(tt.expected) {
				t.Errorf("NegotiateFeatures() length = %d, want %d", len(got), len(tt.expected))
				return
			}

			// Check each expected feature is present
			for _, expected := range tt.expected {
				found := false
				for _, feature := range got {
					if feature == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("NegotiateFeatures() missing expected feature %q", expected)
				}
			}
		})
	}
}

func TestNonceCleanup(t *testing.T) {
	identity, err := NewServerIdentity("TestServer")
	if err != nil {
		t.Fatalf("NewServerIdentity() error: %v", err)
	}

	manager := NewHandshakeManager(identity)
	manager.nonceExpiry = 100 * time.Millisecond // Short expiry for testing

	// Create and process a handshake
	handshake, err := identity.CreateHandshake("6.0.0", []string{"travel"}, TrustVerified)
	if err != nil {
		t.Fatalf("CreateHandshake() error: %v", err)
	}

	err = manager.ProcessHandshake(handshake)
	if err != nil {
		t.Fatalf("ProcessHandshake() error: %v", err)
	}

	// Verify nonce is tracked
	manager.mu.RLock()
	initialCount := len(manager.seenNonces)
	manager.mu.RUnlock()

	if initialCount != 1 {
		t.Errorf("Initial nonce count = %d, want 1", initialCount)
	}

	// Wait for expiry
	time.Sleep(150 * time.Millisecond)

	// Trigger cleanup
	manager.cleanupNonces()

	// Verify nonce is cleaned up
	manager.mu.RLock()
	finalCount := len(manager.seenNonces)
	manager.mu.RUnlock()

	if finalCount != 0 {
		t.Errorf("Final nonce count = %d, want 0", finalCount)
	}
}

// Benchmarks
func BenchmarkNewServerIdentity(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := NewServerIdentity("BenchServer")
		if err != nil {
			b.Fatalf("NewServerIdentity() error: %v", err)
		}
	}
}

func BenchmarkCreateHandshake(b *testing.B) {
	identity, err := NewServerIdentity("BenchServer")
	if err != nil {
		b.Fatalf("NewServerIdentity() error: %v", err)
	}

	features := []string{"travel", "trade", "post"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := identity.CreateHandshake("6.0.0", features, TrustVerified)
		if err != nil {
			b.Fatalf("CreateHandshake() error: %v", err)
		}
	}
}

func BenchmarkVerifyHandshake(b *testing.B) {
	identity, err := NewServerIdentity("BenchServer")
	if err != nil {
		b.Fatalf("NewServerIdentity() error: %v", err)
	}

	handshake, err := identity.CreateHandshake("6.0.0", []string{"travel"}, TrustVerified)
	if err != nil {
		b.Fatalf("CreateHandshake() error: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := VerifyHandshake(handshake)
		if err != nil {
			b.Fatalf("VerifyHandshake() error: %v", err)
		}
	}
}

func BenchmarkProcessHandshake(b *testing.B) {
	identity, err := NewServerIdentity("BenchServer")
	if err != nil {
		b.Fatalf("NewServerIdentity() error: %v", err)
	}

	manager := NewHandshakeManager(identity)

	// Pre-generate handshakes
	handshakes := make([]*FederationHandshake, b.N)
	for i := 0; i < b.N; i++ {
		h, err := identity.CreateHandshake("6.0.0", []string{"travel"}, TrustVerified)
		if err != nil {
			b.Fatalf("CreateHandshake() error: %v", err)
		}
		handshakes[i] = h
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := manager.ProcessHandshake(handshakes[i])
		if err != nil {
			b.Fatalf("ProcessHandshake() error: %v", err)
		}
	}
}

func BenchmarkNegotiateFeatures(b *testing.B) {
	ourFeatures := []string{"travel", "trade", "post", "bounty", "territory"}
	theirFeatures := []string{"travel", "post", "territory"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NegotiateFeatures(ourFeatures, theirFeatures)
	}
}
