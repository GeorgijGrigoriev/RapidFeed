package db

import (
	"regexp"
	"testing"
)

var tokenPattern = regexp.MustCompile(`^[a-zA-Z0-9]+$`)

func TestGenerateToken_ValidLength(t *testing.T) {
	const length = 48

	token, err := generateToken(length)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(token) != length {
		t.Fatalf("expected token length %d, got %d", length, len(token))
	}

	if !tokenPattern.MatchString(token) {
		t.Fatalf("token contains invalid characters: %s", token)
	}
}

func TestGenerateToken_ZeroLength(t *testing.T) {
	if _, err := generateToken(0); err == nil {
		t.Fatalf("expected error for zero length, got nil")
	}
}

func TestGenerateToken_NegativeLength(t *testing.T) {
	if _, err := generateToken(-5); err == nil {
		t.Fatalf("expected error for negative length, got nil")
	}
}

func TestGenerateToken_CharsetCoverage(t *testing.T) {
	const (
		numTokens = 500
		tokenLen  = 16
	)

	seen := make(map[byte]bool)

	for i := 0; i < numTokens; i++ {
		tok, err := generateToken(tokenLen)
		if err != nil {
			t.Fatalf("unexpected error on iteration %d: %v", i, err)
		}

		for j := 0; j < len(tok); j++ {
			seen[tok[j]] = true
		}
	}

	if len(seen) < len(charset)/2 {
		t.Fatalf("charset coverage too low: %d unique chars out of %d", len(seen), len(charset))
	}
}
