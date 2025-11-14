package network

import (
	"bytes"
	"math/big"
	"testing"
)

func TestDefaultDHParams(t *testing.T) {
	params := DefaultDHParams()
	if params == nil {
		t.Fatal("DefaultDHParams returned nil")
	}
	if params.P == nil || params.G == nil {
		t.Fatal("DH params P or G is nil")
	}
	if params.G.Cmp(big.NewInt(2)) != 0 {
		t.Errorf("Expected G = 2, got %v", params.G)
	}
	// Verify P is 2048-bit (256 bytes)
	if len(params.P.Bytes()) != 256 {
		t.Errorf("Expected 2048-bit P (256 bytes), got %d bytes", len(params.P.Bytes()))
	}
}

func TestGenerateKeyPair(t *testing.T) {
	params := DefaultDHParams()
	keyPair, err := GenerateKeyPair(params)
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}
	if keyPair == nil {
		t.Fatal("Generated key pair is nil")
	}
	if keyPair.PrivateKey == nil || keyPair.PublicKey == nil {
		t.Fatal("Private or public key is nil")
	}

	// Verify private key is in range [1, p-1]
	if keyPair.PrivateKey.Cmp(big.NewInt(1)) < 0 || keyPair.PrivateKey.Cmp(params.P) >= 0 {
		t.Error("Private key out of valid range")
	}

	// Verify public key is in range [2, p-2]
	two := big.NewInt(2)
	pMinus2 := new(big.Int).Sub(params.P, two)
	if keyPair.PublicKey.Cmp(two) < 0 || keyPair.PublicKey.Cmp(pMinus2) > 0 {
		t.Error("Public key out of valid range")
	}
}

func TestComputeSharedSecret(t *testing.T) {
	params := DefaultDHParams()

	// Generate two key pairs
	alice, err := GenerateKeyPair(params)
	if err != nil {
		t.Fatalf("Failed to generate Alice's key pair: %v", err)
	}

	bob, err := GenerateKeyPair(params)
	if err != nil {
		t.Fatalf("Failed to generate Bob's key pair: %v", err)
	}

	// Compute shared secrets
	aliceSecret, err := ComputeSharedSecret(alice.PrivateKey, bob.PublicKey, params)
	if err != nil {
		t.Fatalf("Failed to compute Alice's shared secret: %v", err)
	}

	bobSecret, err := ComputeSharedSecret(bob.PrivateKey, alice.PublicKey, params)
	if err != nil {
		t.Fatalf("Failed to compute Bob's shared secret: %v", err)
	}

	// Verify shared secrets match
	if aliceSecret.Cmp(bobSecret) != 0 {
		t.Error("Shared secrets do not match")
	}
}

func TestComputeSharedSecretInvalidPeerKey(t *testing.T) {
	params := DefaultDHParams()
	keyPair, _ := GenerateKeyPair(params)

	tests := []struct {
		name      string
		peerKey   *big.Int
		expectErr bool
	}{
		{"Valid key", keyPair.PublicKey, false},
		{"Key too small (1)", big.NewInt(1), true},
		{"Key too large (p-1)", new(big.Int).Sub(params.P, big.NewInt(1)), true},
		{"Key equals p", params.P, true},
		{"Nil peer key", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ComputeSharedSecret(keyPair.PrivateKey, tt.peerKey, params)
			if tt.expectErr && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.expectErr && err != nil {
				t.Errorf("Expected no error, got %v", err)
			}
		})
	}
}

func TestComputeSharedSecretNilParams(t *testing.T) {
	keyPair, _ := GenerateKeyPair(DefaultDHParams())
	_, err := ComputeSharedSecret(keyPair.PrivateKey, keyPair.PublicKey, nil)
	if err == nil {
		t.Error("Expected error for nil params, got nil")
	}
}

func TestDeriveAESKey(t *testing.T) {
	// Create a known shared secret
	secret, _ := new(big.Int).SetString("12345678901234567890", 10)
	key1 := DeriveAESKey(secret)
	key2 := DeriveAESKey(secret)

	// Verify deterministic: same secret → same key
	if key1 != key2 {
		t.Error("DeriveAESKey not deterministic")
	}

	// Verify key is 32 bytes (256 bits)
	if len(key1) != 32 {
		t.Errorf("Expected 32-byte key, got %d bytes", len(key1))
	}

	// Verify different secrets → different keys
	differentSecret, _ := new(big.Int).SetString("98765432109876543210", 10)
	key3 := DeriveAESKey(differentSecret)
	if key1 == key3 {
		t.Error("Different secrets produced same key")
	}
}

func TestEncryptDecryptMessage(t *testing.T) {
	// Generate encryption key
	params := DefaultDHParams()
	alice, _ := GenerateKeyPair(params)
	bob, _ := GenerateKeyPair(params)
	secret, _ := ComputeSharedSecret(alice.PrivateKey, bob.PublicKey, params)
	key := DeriveAESKey(secret)

	tests := []struct {
		name      string
		plaintext string
	}{
		{"Short message", "Hello, World!"},
		{"Empty message", ""},
		{"Long message", string(make([]byte, 1024))},
		{"Unicode message", "こんにちは世界 🌍"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plaintext := []byte(tt.plaintext)

			// Encrypt
			ciphertext, err := EncryptMessage(key, plaintext)
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			// Verify ciphertext is different from plaintext (except empty)
			if len(plaintext) > 0 && bytes.Equal(ciphertext, plaintext) {
				t.Error("Ciphertext equals plaintext")
			}

			// Decrypt
			decrypted, err := DecryptMessage(key, ciphertext)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			// Verify plaintext matches
			if !bytes.Equal(decrypted, plaintext) {
				t.Errorf("Decrypted text doesn't match. Expected %q, got %q", plaintext, decrypted)
			}
		})
	}
}

func TestEncryptMessageRandomIV(t *testing.T) {
	key := DeriveAESKey(big.NewInt(123456))
	plaintext := []byte("Test message")

	// Encrypt same plaintext twice
	ciphertext1, err1 := EncryptMessage(key, plaintext)
	ciphertext2, err2 := EncryptMessage(key, plaintext)

	if err1 != nil || err2 != nil {
		t.Fatal("Encryption failed")
	}

	// Verify ciphertexts differ (random IV)
	if bytes.Equal(ciphertext1, ciphertext2) {
		t.Error("Same plaintext produced identical ciphertext (IV not random)")
	}

	// Verify both decrypt correctly
	decrypted1, _ := DecryptMessage(key, ciphertext1)
	decrypted2, _ := DecryptMessage(key, ciphertext2)

	if !bytes.Equal(decrypted1, plaintext) || !bytes.Equal(decrypted2, plaintext) {
		t.Error("Decryption failed for random IV test")
	}
}

func TestDecryptMessageInvalidKey(t *testing.T) {
	// Encrypt with one key
	key1 := DeriveAESKey(big.NewInt(123))
	plaintext := []byte("Secret message")
	ciphertext, _ := EncryptMessage(key1, plaintext)

	// Try to decrypt with different key
	key2 := DeriveAESKey(big.NewInt(456))
	_, err := DecryptMessage(key2, ciphertext)

	if err == nil {
		t.Error("Expected decryption to fail with wrong key")
	}
}

func TestDecryptMessageTooShort(t *testing.T) {
	key := DeriveAESKey(big.NewInt(123))

	tests := []struct {
		name       string
		ciphertext []byte
	}{
		{"Empty", []byte{}},
		{"Too short (1 byte)", []byte{0x00}},
		{"Too short (10 bytes)", make([]byte, 10)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := DecryptMessage(key, tt.ciphertext)
			if err == nil {
				t.Error("Expected error for short ciphertext")
			}
		})
	}
}

func TestDecryptMessageCorrupted(t *testing.T) {
	key := DeriveAESKey(big.NewInt(123))
	plaintext := []byte("Test message")
	ciphertext, _ := EncryptMessage(key, plaintext)

	// Corrupt the ciphertext
	ciphertext[len(ciphertext)/2] ^= 0xFF

	_, err := DecryptMessage(key, ciphertext)
	if err == nil {
		t.Error("Expected error for corrupted ciphertext")
	}
}

func TestEndToEndKeyExchange(t *testing.T) {
	params := DefaultDHParams()

	// Alice generates key pair
	aliceKeyPair, err := GenerateKeyPair(params)
	if err != nil {
		t.Fatalf("Failed to generate Alice's key pair: %v", err)
	}

	// Bob generates key pair
	bobKeyPair, err := GenerateKeyPair(params)
	if err != nil {
		t.Fatalf("Failed to generate Bob's key pair: %v", err)
	}

	// Alice computes shared secret and derives key
	aliceSharedSecret, err := ComputeSharedSecret(aliceKeyPair.PrivateKey, bobKeyPair.PublicKey, params)
	if err != nil {
		t.Fatalf("Alice failed to compute shared secret: %v", err)
	}
	aliceKey := DeriveAESKey(aliceSharedSecret)

	// Bob computes shared secret and derives key
	bobSharedSecret, err := ComputeSharedSecret(bobKeyPair.PrivateKey, aliceKeyPair.PublicKey, params)
	if err != nil {
		t.Fatalf("Bob failed to compute shared secret: %v", err)
	}
	bobKey := DeriveAESKey(bobSharedSecret)

	// Verify keys match
	if aliceKey != bobKey {
		t.Fatal("Alice and Bob derived different keys")
	}

	// Alice encrypts message
	message := []byte("Hello Bob, this is Alice!")
	ciphertext, err := EncryptMessage(aliceKey, message)
	if err != nil {
		t.Fatalf("Alice failed to encrypt message: %v", err)
	}

	// Bob decrypts message
	decrypted, err := DecryptMessage(bobKey, ciphertext)
	if err != nil {
		t.Fatalf("Bob failed to decrypt message: %v", err)
	}

	// Verify message matches
	if !bytes.Equal(decrypted, message) {
		t.Errorf("Message mismatch. Expected %q, got %q", message, decrypted)
	}
}

// Benchmarks

func BenchmarkGenerateKeyPair(b *testing.B) {
	params := DefaultDHParams()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateKeyPair(params)
	}
}

func BenchmarkComputeSharedSecret(b *testing.B) {
	params := DefaultDHParams()
	alice, _ := GenerateKeyPair(params)
	bob, _ := GenerateKeyPair(params)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ComputeSharedSecret(alice.PrivateKey, bob.PublicKey, params)
	}
}

func BenchmarkDeriveAESKey(b *testing.B) {
	secret, _ := new(big.Int).SetString("12345678901234567890", 10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DeriveAESKey(secret)
	}
}

func BenchmarkEncryptMessage(b *testing.B) {
	key := DeriveAESKey(big.NewInt(123))
	plaintext := []byte("This is a benchmark message for encryption testing.")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = EncryptMessage(key, plaintext)
	}
}

func BenchmarkDecryptMessage(b *testing.B) {
	key := DeriveAESKey(big.NewInt(123))
	plaintext := []byte("This is a benchmark message for decryption testing.")
	ciphertext, _ := EncryptMessage(key, plaintext)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = DecryptMessage(key, ciphertext)
	}
}

func BenchmarkEndToEndEncryption(b *testing.B) {
	// Setup
	params := DefaultDHParams()
	alice, _ := GenerateKeyPair(params)
	bob, _ := GenerateKeyPair(params)
	aliceSecret, _ := ComputeSharedSecret(alice.PrivateKey, bob.PublicKey, params)
	bobSecret, _ := ComputeSharedSecret(bob.PrivateKey, alice.PublicKey, params)
	aliceKey := DeriveAESKey(aliceSecret)
	bobKey := DeriveAESKey(bobSecret)
	message := []byte("Hello Bob!")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ciphertext, _ := EncryptMessage(aliceKey, message)
		_, _ = DecryptMessage(bobKey, ciphertext)
	}
}
