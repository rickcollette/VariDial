package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"strings"
)

/*
This code demonstrates how to use the `generateKey` and `readKey` functions to encrypt and decrypt a string.
The `generateKey` function takes a username, date format, and number as input and returns an encrypted ciphertext.
The `readKey` function takes the ciphertext as input, decrypts it, and returns the original username,
date format, and number values.
*/

func generateKey(username string, dateFormat string, number int) ([]byte, error) {
	// AES key (must be 32 bytes)
	key := []byte("your_secret_key_must_be_32_bytes")

	// Format the plaintext string
	plaintext := []byte(fmt.Sprintf("%s###%s###%03d", username, dateFormat, number))

	// Generate a new cipher block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// Generate a random initialisation vector
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return nil, err
	}

	// Create a new stream cipher using the block and initialisation vector
	stream := cipher.NewCFBEncrypter(block, iv)

	// Encrypt the plaintext and write it to the ciphertext buffer
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func readKey(ciphertext []byte) (username string, dateFormat string, number int, err error) {
	// AES key (must be 32 bytes)
	key := []byte("your_secret_key_must_be_32_bytes")

	// Generate a new cipher block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", 0, err
	}

	// Get the initialisation vector from the ciphertext
	iv := ciphertext[:aes.BlockSize]

	// Create a new stream cipher using the block and initialisation vector
	stream := cipher.NewCFBDecrypter(block, iv)

	// Decrypt the ciphertext
	plaintext := make([]byte, len(ciphertext)-aes.BlockSize)
	stream.XORKeyStream(plaintext, ciphertext[aes.BlockSize:])

	// Split the decrypted plaintext string into its original values
	values := strings.Split(string(plaintext), "###")
	username = values[0]
	dateFormat = values[1]
	fmt.Sscanf(values[2], "%03d", &number)

	return username, dateFormat, number, nil
}

func main() {
	ciphertext, err := generateKey("your_username", "020923", 123)
	if err != nil {
		panic(err)
	}

	username, dateFormat, number, err := readKey(ciphertext)
	if err != nil {
		panic(err)
	}

	fmt.Println("Username:", username)
	fmt.Println("Date format:", dateFormat)
	fmt.Println("Number:", number)
}
