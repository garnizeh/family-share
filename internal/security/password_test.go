package security

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "mySecurePassword123"
	
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	
	if hash == "" {
		t.Error("Hash should not be empty")
	}
	
	if hash == password {
		t.Error("Hash should not be the same as the password")
	}
}

func TestVerifyPassword_Correct(t *testing.T) {
	password := "mySecurePassword123"
	hash, _ := HashPassword(password)
	
	if !VerifyPassword(hash, password) {
		t.Error("Should verify correct password")
	}
}

func TestVerifyPassword_Incorrect(t *testing.T) {
	password := "mySecurePassword123"
	wrongPassword := "wrongPassword"
	hash, _ := HashPassword(password)
	
	if VerifyPassword(hash, wrongPassword) {
		t.Error("Should not verify incorrect password")
	}
}

func TestVerifyPassword_EmptyPassword(t *testing.T) {
	password := "mySecurePassword123"
	hash, _ := HashPassword(password)
	
	if VerifyPassword(hash, "") {
		t.Error("Should not verify empty password")
	}
}

func TestHashPassword_ConsistentHashing(t *testing.T) {
	password := "mySecurePassword123"
	
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)
	
	// Bcrypt should produce different hashes for the same password (salt)
	if hash1 == hash2 {
		t.Error("Two hashes of the same password should be different (due to salt)")
	}
	
	// But both should verify correctly
	if !VerifyPassword(hash1, password) {
		t.Error("First hash should verify")
	}
	if !VerifyPassword(hash2, password) {
		t.Error("Second hash should verify")
	}
}
