package utils

import "testing"

func TestGenTokenLength(t *testing.T) {
	for _, size := range []int{1, 8, 16, 32, 64} {
		tok := GenToken(size)
		if len(tok) != size {
			t.Fatalf("GenToken(%d) returned length %d", size, len(tok))
		}
	}
}

func TestGenTokenUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		tok := GenToken(32)
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

	tok := GenToken(256)
	for i := 0; i < len(tok); i++ {
		if !charSet[tok[i]] {
			t.Fatalf("GenToken produced invalid character: %c", tok[i])
		}
	}
}
