package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// GenerateNodeID generates a random node ID
func GenerateNodeID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// GenerateKeyPair generates a key pair for signing transactions and blocks
// Note: This is a simplified version. In a real implementation, use proper crypto libraries.
func GenerateKeyPair() (string, string, error) {
	// Generate private key
	privKeyBytes := make([]byte, 32)
	_, err := rand.Read(privKeyBytes)
	if err != nil {
		return "", "", err
	}
	privateKey := hex.EncodeToString(privKeyBytes)
	
	// Generate public key (in a real system this would derive from private key)
	pubKeyBytes := make([]byte, 32)
	_, err = rand.Read(pubKeyBytes)
	if err != nil {
		return "", "", err
	}
	publicKey := hex.EncodeToString(pubKeyBytes)
	
	return privateKey, publicKey, nil
}

// Hash computes the SHA-256 hash of the input data
func Hash(data []byte) string {
	hasher := sha256.New()
	hasher.Write(data)
	return hex.EncodeToString(hasher.Sum(nil))
}

// Sign signs data with a private key
// Note: This is a simplified version. In a real implementation, use proper crypto libraries.
func Sign(data []byte, privateKey string) (string, error) {
	// Placeholder for actual signature logic
	hash := Hash(data)
	signature := fmt.Sprintf("signed:%s:%s", privateKey[:8], hash[:16])
	return signature, nil
}

// VerifySignature verifies a signature against data and a public key
// Note: This is a simplified version. In a real implementation, use proper crypto libraries.
func VerifySignature(data []byte, signature string, publicKey string) bool {
	// Placeholder for actual verification logic
	return len(signature) > 0
}

// GenerateRandomHex generates a random hex string of the specified length
func GenerateRandomHex(length int) (string, error) {
	bytes := make([]byte, length/2)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
