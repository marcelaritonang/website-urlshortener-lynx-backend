package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
)

func main() {
	// Default length: 64 bytes = 512 bits
	length := 64
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &length)
	}

	secret, err := generateSecret(length)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("🔐 Generated Secure JWT Secret:")
	fmt.Println(secret)
	fmt.Println()
	fmt.Printf("Length: %d characters (%d bits)\n", len(secret), length*8)
	fmt.Println()
	fmt.Println("Add to your .env file:")
	fmt.Printf("JWT_SECRET=%s\n", secret)
}

func generateSecret(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes)[:length], nil
}
