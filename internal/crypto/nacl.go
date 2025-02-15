package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/crypto/nacl/box"
)

const (
	naclBackupDir  = ".jot/backup"
	naclPubKeyFile = "jot.pub"
	naclSecKeyFile = "jot.sec"
)

// KeyPair represents a NaCl public/private key pair
type KeyPair struct {
	PublicKey  *[32]byte
	PrivateKey *[32]byte
}

// GenerateNaclKey generates a new NaCl key pair for the journal
func GenerateNaclKey() (string, error) {
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate key pair: %w", err)
	}

	// Convert keys to storable format
	pubKeyStr := base64.StdEncoding.EncodeToString(publicKey[:])
	privKeyStr := base64.StdEncoding.EncodeToString(privateKey[:])

	// Store keys
	if err := backupNaclKey(pubKeyStr, privKeyStr); err != nil {
		return "", err
	}

	return pubKeyStr, nil
}

// backupNaclKey exports and saves both public and private keys to the backup directory
func backupNaclKey(pubKeyStr, privKeyStr string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	backupPath := filepath.Join(homeDir, naclBackupDir)
	if err := os.MkdirAll(backupPath, 0700); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Save public key
	pubKeyPath := filepath.Join(backupPath, naclPubKeyFile)
	if err := os.WriteFile(pubKeyPath, []byte(pubKeyStr), 0644); err != nil {
		return fmt.Errorf("failed to save public key backup: %w", err)
	}

	// Save private key with restricted permissions
	secKeyPath := filepath.Join(backupPath, naclSecKeyFile)
	if err := os.WriteFile(secKeyPath, []byte(privKeyStr), 0600); err != nil {
		return fmt.Errorf("failed to save private key backup: %w", err)
	}

	return nil
}

// RestoreNaclFromBackup attempts to restore the NaCl key pair from backup
func RestoreNaclFromBackup() (*KeyPair, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	backupPath := filepath.Join(homeDir, naclBackupDir)
	pubKeyPath := filepath.Join(backupPath, naclPubKeyFile)
	secKeyPath := filepath.Join(backupPath, naclSecKeyFile)

	// Check if backup files exist
	if _, err := os.Stat(pubKeyPath); err != nil {
		return nil, fmt.Errorf("public key backup not found: %w", err)
	}
	if _, err := os.Stat(secKeyPath); err != nil {
		return nil, fmt.Errorf("private key backup not found: %w", err)
	}

	// Read public key
	pubKeyData, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key: %w", err)
	}

	// Read private key
	privKeyData, err := os.ReadFile(secKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	// Decode keys from Base64
	pubKeyBytes, err := base64.StdEncoding.DecodeString(string(pubKeyData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode public key: %w", err)
	}

	privKeyBytes, err := base64.StdEncoding.DecodeString(string(privKeyData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	// Convert to key pair
	var publicKey, privateKey [32]byte
	copy(publicKey[:], pubKeyBytes)
	copy(privateKey[:], privKeyBytes)

	return &KeyPair{
		PublicKey:  &publicKey,
		PrivateKey: &privateKey,
	}, nil
}

// EncryptNacl encrypts the given text using NaCl box
func EncryptNacl(text string, keyPair *KeyPair) ([]byte, error) {
	// Generate random nonce
	var nonce [24]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return nil, fmt.Errorf("nonce generation failed: %w", err)
	}

	// Encrypt message
	encrypted := box.Seal(nonce[:], []byte(text), &nonce, keyPair.PublicKey, keyPair.PrivateKey)
	return encrypted, nil
}

// DecryptNacl decrypts the given data using NaCl box
func DecryptNacl(data []byte, keyPair *KeyPair) (string, error) {
	if len(data) < 24 {
		return "", fmt.Errorf("invalid encrypted data: too short")
	}

	// Extract nonce
	var nonce [24]byte
	copy(nonce[:], data[:24])

	// Decrypt message
	decrypted, ok := box.Open(nil, data[24:], &nonce, keyPair.PublicKey, keyPair.PrivateKey)
	if !ok {
		return "", fmt.Errorf("decryption failed")
	}

	return string(decrypted), nil
}

// Clear securely zeros sensitive data
func (k *KeyPair) Clear() {
	if k.PrivateKey != nil {
		for i := range k.PrivateKey {
			k.PrivateKey[i] = 0
		}
	}
	if k.PublicKey != nil {
		for i := range k.PublicKey {
			k.PublicKey[i] = 0
		}
	}
}
