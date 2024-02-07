package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	mrand "math/rand" // aliased because it's used less frequently or it's not the standard rand in your app
	"os"
	"strings"
	"time"
)

func init() {
	// Initialize the random number generator with a time-based seed.
	mrand.Seed(time.Now().UnixNano())
}

// Encrypts text with the given key
func EncryptString(text string) (string, error) {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		panic("SECRET_KEY environment variable not set")
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	plaintext := []byte(text)
	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

// Decrypts text with the given key
func DecryptString(cryptoText string) (string, error) {
	key := os.Getenv("SECRET_KEY")
	if key == "" {
		panic("SECRET_KEY environment variable not set")
	}

	ciphertext, err := base64.URLEncoding.DecodeString(cryptoText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", err
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

// CompareHashAndString compares a HMAC-SHA256 hash with a string
func CompareHashAndString(hash, s string) (bool, error) {
	decryptedString, err := DecryptString(hash)
	if err != nil {
		return false, err
	}

	return hmac.Equal([]byte(decryptedString), []byte(s)), nil
}

func GenerateSecretToken() (string, error) {
	token := make([]byte, 32) // Generate a 32 characters long token
	_, err := rand.Read(token)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(token), nil
}

func EncryptSecretToken(secretToken, secretKey string) (string, error) {
	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", err
	}

	ciphertext := make([]byte, aes.BlockSize+len(secretToken))
	iv := ciphertext[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], []byte(secretToken))

	return hex.EncodeToString(ciphertext), nil
}

func DecryptSecretToken(encryptedToken, secretKey string) (string, error) {
	ciphertext, err := hex.DecodeString(encryptedToken)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(secretKey))
	if err != nil {
		return "", err
	}

	if len(ciphertext) < aes.BlockSize {
		return "", errors.New("ciphertext too short")
	}

	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)

	// XORKeyStream can work in-place if the two arguments are the same.
	stream.XORKeyStream(ciphertext, ciphertext)

	return string(ciphertext), nil
}

func GenerateRandomString(n int) string {
	const characters = "DATSEKIONabcdefghijklmnopqrstuvwxyz0123456789"
	var result strings.Builder
	length := len(characters)

	for i := 0; i < n; i++ {
		randomIndex := mrand.Intn(length)
		result.WriteByte(characters[randomIndex])
	}

	return result.String()
}
