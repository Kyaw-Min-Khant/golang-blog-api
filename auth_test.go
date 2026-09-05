package main

import "testing"

func TestHashPasswordAndCompare(t *testing.T) {
	hash, err := hashPassword("correct-password")
	if err != nil {
		t.Fatalf("hashPassword returned error: %v", err)
	}
	if hash == "correct-password" {
		t.Fatal("hashPassword did not hash the password")
	}
}

func TestGenerateAndParseToken(t *testing.T) {
	setJWTSecret("test-secret")

	token, err := generateToken(42, "alice")
	if err != nil {
		t.Fatalf("generateToken returned error: %v", err)
	}

	claims, err := parseToken(token)
	if err != nil {
		t.Fatalf("parseToken returned error: %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("UserID = %d, want 42", claims.UserID)
	}
	if claims.Username != "alice" {
		t.Errorf("Username = %q, want %q", claims.Username, "alice")
	}
}

func TestParseTokenRejectsTamperedSecret(t *testing.T) {
	setJWTSecret("original-secret")
	token, err := generateToken(1, "bob")
	if err != nil {
		t.Fatalf("generateToken returned error: %v", err)
	}

	setJWTSecret("different-secret")
	if _, err := parseToken(token); err == nil {
		t.Fatal("parseToken accepted a token signed with a different secret")
	}
}
