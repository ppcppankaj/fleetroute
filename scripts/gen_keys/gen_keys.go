// gen_keys.go rotates the JWT RSA-2048 key pair.
// Run once: go run scripts/gen_keys.go
// The secrets/ directory is gitignored — never commit these files.
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

func main() {
	if err := os.MkdirAll("secrets", 0700); err != nil {
		fmt.Fprintf(os.Stderr, "mkdir secrets: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Generating RSA-2048 key pair...")
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Fprintf(os.Stderr, "generate key: %v\n", err)
		os.Exit(1)
	}

	// Write private key
	privFile, err := os.OpenFile("secrets/jwt_private.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open private key file: %v\n", err)
		os.Exit(1)
	}
	defer privFile.Close()
	if err := pem.Encode(privFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privKey),
	}); err != nil {
		fmt.Fprintf(os.Stderr, "encode private key: %v\n", err)
		os.Exit(1)
	}

	// Write public key
	pubBytes, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "marshal public key: %v\n", err)
		os.Exit(1)
	}
	pubFile, err := os.OpenFile("secrets/jwt_public.pem", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "open public key file: %v\n", err)
		os.Exit(1)
	}
	defer pubFile.Close()
	if err := pem.Encode(pubFile, &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubBytes,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "encode public key: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Keys written to secrets/jwt_private.pem and secrets/jwt_public.pem")
	fmt.Println("⚠️  secrets/ is gitignored. Mount these files via Docker secrets or environment in production.")
	fmt.Println("⚠️  All existing sessions are now invalid — users must log in again.")
}
