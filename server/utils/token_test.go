package utils

import (
	"crypto/rand"
	"errors"
	"testing"
)

type failingReader struct{}

func (failingReader) Read(p []byte) (int, error) {
	return 0, errors.New("simulated crypto/rand failure")
}

func TestGenTokenLength(t *testing.T) {
	for _, size := range []int{1, 8, 16, 32, 64} {
		tok, err := GenToken(size)
		if err != nil {
			t.Fatalf("GenToken(%d) returned error: %v", size, err)
		}
		if len(tok) != size {
			t.Fatalf("GenToken(%d) returned length %d", size, len(tok))
		}
	}
}

func TestGenTokenUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		tok, err := GenToken(32)
		if err != nil {
			t.Fatalf("GenToken returned error: %v", err)
		}
		if seen[tok] {
			t.Fatalf("duplicate token generated: %s", tok)
		}
		seen[tok] = true
	}
}

func TestGenTokenCharset(t *testing.T) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charSet := make(map[byte]bool)
	for i := 0; i < len(charset); i++ {
		charSet[charset[i]] = true
	}

	tok, err := GenToken(256)
	if err != nil {
		t.Fatalf("GenToken returned error: %v", err)
	}
	for i := 0; i < len(tok); i++ {
		if !charSet[tok[i]] {
			t.Fatalf("GenToken produced invalid character: %c", tok[i])
		}
	}
}

func TestGenTokenReturnsErrorWhenRandReaderFails(t *testing.T) {
	orig := rand.Reader
	rand.Reader = failingReader{}
	t.Cleanup(func() { rand.Reader = orig })

	_, err := GenToken(8)
	if err == nil {
		t.Fatal("expected error when crypto/rand.Reader fails")
	}
	if got := err.Error(); got == "" {
		t.Fatal("expected non-empty error message")
	}
}
