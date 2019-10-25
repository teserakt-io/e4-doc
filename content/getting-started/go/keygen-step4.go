package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	e4crypto "github.com/teserakt-io/e4go/crypto"
	"golang.org/x/crypto/ed25519"
)

func main() {
	var name string
	var password string

	flag.StringVar(&name, "name", "", "the key filename")
	flag.StringVar(&password, "password", "", "the password to generate the keys from")
	flag.Parse()

	if len(name) == 0 {
		fmt.Println("-name is required")
		os.Exit(1)
	}

	if len(password) == 0 {
		fmt.Println("-password is required")
		os.Exit(1)
	}

	// Generate private key from password
	privateKey, err := e4crypto.Ed25519PrivateKeyFromPassword(password)
	if err != nil {
		fmt.Printf("failed to generate private key: %v\n", err)
		os.Exit(1)
	}

	// Write private key file
	if err := ioutil.WriteFile(name, privateKey, 0600); err != nil {
		fmt.Printf("failed to create private key file './%s': %v\n", name, err)
		os.Exit(1)
	}
	fmt.Printf("Generated private key: ./%s\n", name)

	// Write public key file
	pubName := fmt.Sprintf("%s.pub", name)
	pubBytes, ok := privateKey.Public().(ed25519.PublicKey)
	if !ok {
		panic(fmt.Sprintf("%T is invalid for public key, wanted ed25519.PublicKey", privateKey.Public()))
	}
	if err := ioutil.WriteFile(pubName, pubBytes, 0644); err != nil {
		fmt.Printf("failed to create public key file './%s': %v\n", pubName, err)
		os.Exit(1)
	}
	fmt.Printf("Generated public key: ./%s\n", pubName)
}
