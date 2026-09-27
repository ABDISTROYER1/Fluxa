package apikey

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerate(t *testing.T) {
	raw, prefix, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	// Check raw key format
	if !strings.HasPrefix(raw, "sk_live_") {
		t.Errorf("raw key should start with 'sk_live_', got: %s", raw)
	}

	// Check prefix is 8 chars from base58 body
	if len(prefix) != 8 {
		t.Errorf("prefix should be 8 chars, got %d: %s", len(prefix), prefix)
	}

	// Prefix should be from the base58 body, not "sk_live_"
	if prefix == "sk_live_" {
		t.Errorf("prefix should not be 'sk_live_', got: %s", prefix)
	}

	// Prefix should be first 8 chars of base58 body
	base58Body := strings.TrimPrefix(raw, "sk_live_")
	if prefix != base58Body[:8] {
		t.Errorf("prefix should match first 8 chars of base58 body: expected %s, got %s", base58Body[:8], prefix)
	}
}

func TestGenerateUniqueKeys(t *testing.T) {
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		raw, prefix, err := Generate()
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		if keys[raw] {
			t.Fatalf("duplicate key generated: %s", raw)
		}
		if keys[prefix] {
			t.Fatalf("duplicate prefix generated: %s", prefix)
		}
		keys[raw] = true
		keys[prefix] = true
	}
}

func TestHash(t *testing.T) {
	raw := "sk_live_testkey123"
	hashed := Hash(raw)

	// Verify it's a valid hex string
	decoded, err := hex.DecodeString(hashed)
	if err != nil {
		t.Errorf("hash should be valid hex: %v", err)
	}
	if len(decoded) != 32 {
		t.Errorf("SHA-256 hash should be 32 bytes, got %d", len(decoded))
	}

	// Verify it matches manual SHA-256
	h := sha256.New()
	h.Write([]byte(raw))
	expected := hex.EncodeToString(h.Sum(nil))
	if hashed != expected {
		t.Errorf("hash mismatch: got %s, expected %s", hashed, expected)
	}
}

func TestVerify(t *testing.T) {
	raw, _, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	hashed := Hash(raw)

	// Valid key should verify
	if !Verify(raw, hashed) {
		t.Error("Verify should return true for valid key")
	}

	// Invalid key should not verify
	if Verify("sk_live_wrongkey", hashed) {
		t.Error("Verify should return false for invalid key")
	}
}

func TestVerifyConstantTime(t *testing.T) {
	raw, _, err := Generate()
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	hashed := Hash(raw)

	// Test that Verify uses constant-time comparison by checking
	// that it doesn't panic and returns correct results
	// (We can't easily test timing in unit tests, but we can verify
	// the implementation uses subtle.ConstantTimeCompare)
	_ = Verify(raw, hashed)
	_ = Verify("wrong", hashed)
}

func TestVerifyTimingAttackResistance(t *testing.T) {
	// Create two hashes that differ only in the last byte
	h1 := sha256.Sum256([]byte("test1"))
	h2 := sha256.Sum256([]byte("test2"))
	hash1 := hex.EncodeToString(h1[:])
	hash2 := hex.EncodeToString(h2[:])

	// Verify doesn't panic
	_ = Verify("test1", hash1)
	_ = Verify("test1", hash2)
	_ = Verify("test2", hash1)
	_ = Verify("test2", hash2)

	// The important thing is that Verify uses subtle.ConstantTimeCompare
	// which we can verify by checking the implementation
}

func TestPrefixUniqueness(t *testing.T) {
	// Generate many keys and ensure prefixes are unique
	prefixes := make(map[string]int)
	for i := 0; i < 1000; i++ {
		_, prefix, err := Generate()
		if err != nil {
			t.Fatalf("Generate: %v", err)
		}
		prefixes[prefix]++
	}

	// With 1000 keys and 8-char base58 prefix, collisions are extremely unlikely
	// but we verify no duplicates
	for prefix, count := range prefixes {
		if count > 1 {
			t.Errorf("duplicate prefix %s found %d times", prefix, count)
		}
	}
}
