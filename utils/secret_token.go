package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
)

func HashString(s string) (string, error) {
	secretKey := os.Getenv("SECRET_KEY")
	if secretKey == "" {
		panic("SECRET_KEY environment variable not set")
	}

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CompareHashAndString compares a HMAC-SHA256 hash with a string
func CompareHashAndString(hash, s string) (bool, error) {
	expectedHash, err := HashString(s)
	if err != nil {
		return false, err
	}

	return hmac.Equal([]byte(hash), []byte(expectedHash)), nil
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
