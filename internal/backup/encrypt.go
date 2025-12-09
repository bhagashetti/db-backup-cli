package backup

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"os"
)

// EncryptFile encrypts src into dst using AES-256-GCM with the given key bytes.
func EncryptFile(src, dst string, key []byte) error {
	if len(key) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes for AES-256")
	}

	plaintext, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read src for encrypt: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("new cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("new GCM: %w", err)
	}

	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("read nonce: %w", err)
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)

	// write nonce + ciphertext to dst
	f, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst for encrypt: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(nonce); err != nil {
		return fmt.Errorf("write nonce: %w", err)
	}
	if _, err := f.Write(ciphertext); err != nil {
		return fmt.Errorf("write ciphertext: %w", err)
	}

	return nil
}
