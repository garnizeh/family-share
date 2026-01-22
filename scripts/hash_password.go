package main

import (
	"fmt"
	"os"

	"familyshare/internal/security"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/hash_password.go <password>")
		fmt.Println("Example: go run scripts/hash_password.go mySecurePassword123")
		os.Exit(1)
	}

	password := os.Args[1]
	hash, err := security.HashPassword(password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error hashing password: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Password hash generated successfully!")
	fmt.Println()
	fmt.Println("Add this to your environment variables:")
	fmt.Printf("export ADMIN_PASSWORD_HASH='%s'\n", hash)
	fmt.Println()
	fmt.Println("Or add it to your .env file:")
	fmt.Printf("ADMIN_PASSWORD_HASH=%s\n", hash)
}
