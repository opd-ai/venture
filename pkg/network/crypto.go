// Package network provides multiplayer networking functionality.
package network

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"math/big"

	"github.com/sirupsen/logrus"
)

// DiffieHellmanParams holds the DH key exchange parameters.
type DiffieHellmanParams struct {
	P *big.Int // Large prime modulus (2048-bit)
	G *big.Int // Generator (typically 2)
}

// DiffieHellmanKeyPair represents a DH key pair.
type DiffieHellmanKeyPair struct {
	PrivateKey *big.Int
	PublicKey  *big.Int
}

// AESKey represents a 256-bit AES encryption key.
type AESKey [32]byte

// DefaultDHParams returns the default 2048-bit DH parameters.
// Uses a well-known safe prime from RFC 3526 (Group 14).
func DefaultDHParams() *DiffieHellmanParams {
	// RFC 3526 2048-bit MODP Group 14
	p, _ := new(big.Int).SetString("FFFFFFFFFFFFFFFFC90FDAA22168C234C4C6628B80DC1CD129024E088A67CC74020BBEA63B139B22514A08798E3404DDEF9519B3CD3A431B302B0A6DF25F14374FE1356D6D51C245E485B576625E7EC6F44C42E9A637ED6B0BFF5CB6F406B7EDEE386BFB5A899FA5AE9F24117C4B1FE649286651ECE45B3DC2007CB8A163BF0598DA48361C55D39A69163FA8FD24CF5F83655D23DCA3AD961C62F356208552BB9ED529077096966D670C354E4ABC9804F1746C08CA18217C32905E462E36CE3BE39E772C180E86039B2783A2EC07A28FB5C55DF06F4C52C9DE2BCBF6955817183995497CEA956AE515D2261898FA051015728E5A8AACAA68FFFFFFFFFFFFFFFF", 16)
	g := big.NewInt(2)
	return &DiffieHellmanParams{P: p, G: g}
}

// GenerateKeyPair generates a DH key pair using the provided parameters.
func GenerateKeyPair(params *DiffieHellmanParams) (*DiffieHellmanKeyPair, error) {
	// Generate random private key (256 bits)
	privateKey, err := rand.Int(rand.Reader, new(big.Int).Sub(params.P, big.NewInt(1)))
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to generate DH private key")
		return nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	// Compute public key: g^privateKey mod p
	publicKey := new(big.Int).Exp(params.G, privateKey, params.P)

	logrus.Debug("DH key pair generated")
	return &DiffieHellmanKeyPair{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
	}, nil
}

// ComputeSharedSecret computes the DH shared secret from our private key and peer's public key.
func ComputeSharedSecret(privateKey, peerPublicKey *big.Int, params *DiffieHellmanParams) (*big.Int, error) {
	if privateKey == nil || peerPublicKey == nil || params == nil {
		logrus.Warn("invalid DH parameters: nil key or params")
		return nil, fmt.Errorf("invalid parameters: nil key or params")
	}

	// Validate peer public key is in valid range [2, p-2]
	two := big.NewInt(2)
	pMinus2 := new(big.Int).Sub(params.P, two)
	if peerPublicKey.Cmp(two) < 0 || peerPublicKey.Cmp(pMinus2) > 0 {
		logrus.Warn("invalid peer public key: out of range")
		return nil, fmt.Errorf("invalid peer public key: out of range")
	}

	// Compute shared secret: peerPublicKey^privateKey mod p
	sharedSecret := new(big.Int).Exp(peerPublicKey, privateKey, params.P)
	logrus.Debug("DH shared secret computed")
	return sharedSecret, nil
}

// DeriveAESKey derives a 256-bit AES key from the DH shared secret using SHA-256.
func DeriveAESKey(sharedSecret *big.Int) AESKey {
	hash := sha256.Sum256(sharedSecret.Bytes())
	return hash
}

// EncryptMessage encrypts a message using AES-256-GCM with a random IV.
// Returns ciphertext with prepended IV (12 bytes IV + encrypted data + 16 bytes tag).
func EncryptMessage(key AESKey, plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to create AES cipher")
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to create GCM")
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random IV (12 bytes for GCM)
	iv := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to generate IV")
		return nil, fmt.Errorf("failed to generate IV: %w", err)
	}

	// Encrypt: prepend IV to ciphertext
	ciphertext := gcm.Seal(iv, iv, plaintext, nil)
	logrus.WithFields(logrus.Fields{
		"bytes_plaintext":  len(plaintext),
		"bytes_ciphertext": len(ciphertext),
	}).Debug("message encrypted")
	return ciphertext, nil
}

// DecryptMessage decrypts a message encrypted with EncryptMessage.
// Expects ciphertext with prepended IV.
func DecryptMessage(key AESKey, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(key[:])
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to create AES cipher")
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("failed to create GCM")
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum length (IV + tag)
	if len(ciphertext) < gcm.NonceSize()+gcm.Overhead() {
		logrus.WithFields(logrus.Fields{
			"bytes_ciphertext": len(ciphertext),
			"min_size":         gcm.NonceSize() + gcm.Overhead(),
		}).Warn("ciphertext too short")
		return nil, fmt.Errorf("ciphertext too short")
	}

	// Extract IV and ciphertext
	iv := ciphertext[:gcm.NonceSize()]
	actualCiphertext := ciphertext[gcm.NonceSize():]

	// Decrypt
	plaintext, err := gcm.Open(nil, iv, actualCiphertext, nil)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err.Error(),
		}).Error("decryption failed")
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"bytes_ciphertext": len(ciphertext),
		"bytes_plaintext":  len(plaintext),
	}).Debug("message decrypted")

	return plaintext, nil
}
