package storage

import (
	"encoding/hex"
	"testing"
)

func TestGenerateRandomKey(t *testing.T) {
	key, err := generateRandomKey()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(key) != 64 { // 32 bytes * 2 hex chars/byte = 64
		t.Errorf("expected length 64, got %d", len(key))
	}

	// Verify it's hex
	_, err = hex.DecodeString(key)
	if err != nil {
		t.Errorf("key is not valid hex: %v", err)
	}

	// Verify randomness (sanity check)
	key2, _ := generateRandomKey()
	if key == key2 {
		t.Error("generated duplicate keys")
	}
}
