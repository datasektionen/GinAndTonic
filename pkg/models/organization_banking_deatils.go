package models

import (
	"errors"
	"fmt"
	"os"
	"regexp"

	"github.com/DowLucas/gin-ticket-release/pkg/encryption"
	"gorm.io/gorm"
)

type BankingDetail struct {
	gorm.Model
	TeamID         uint   `json:"team_id" gorm:"index"`
	BankName       string `gorm:"size:512" json:"bank_name"`       // Encrypted, typically 4-10 characters, provided as size:255 for potential extra characters
	AccountHolder  string `gorm:"size:512" json:"account_holder"`  // Encrypted, typically 4-10 characters, provided as size:255 for potential extra characters
	AccountNumber  string `gorm:"size:512" json:"account_number"`  // Encrypted, typically 4-10 digits, provided as size:255 for potential extra characters
	ClearingNumber string `gorm:"size:512" json:"clearing_number"` // Encrypted, typically 4-5 digits, provided as size:6 for potential extra characters
	IBAN           string `gorm:"size:512" json:"-"`               // Encrypted, up to 34 characters for IBAN
	BIC            string `gorm:"size:512" json:"-"`               // Encrypted, up to 11 characters for BIC/SWIFT
}

func isValidSwedishBankAccount(clearing, account string) bool {
	var re *regexp.Regexp
	switch {
	case regexp.MustCompile(`^(3300|3782)$`).MatchString(clearing):
		re = regexp.MustCompile(`^\d{10}$`)
	case regexp.MustCompile(`^[7-8][0-9]{3}$`).MatchString(clearing):
		if clearing[0] == '7' {
			re = regexp.MustCompile(`^\d{11}$`)
		} else {
			re = regexp.MustCompile(`^\d{15}$`)
		}
	default:
		re = regexp.MustCompile(`^\d{6,10}$`)
	}
	return re.MatchString(account)
}

func isValidClearingNumber(number string) bool {
	match, _ := regexp.MatchString(`^\d{4,5}$`, number)
	return match
}

func (b *BankingDetail) Validate() error {
	if b.TeamID == 0 {
		return errors.New("team id is required")
	}
	if b.BankName == "" {
		return errors.New("bank name is required")
	}

	if b.AccountHolder == "" {
		return errors.New("account holder name is required")
	}
	if !isValidSwedishBankAccount(b.ClearingNumber, b.AccountNumber) {
		return errors.New("invalid account number format")
	}
	if !isValidClearingNumber(b.ClearingNumber) {
		return errors.New("invalid clearing number format")
	}

	// IBAN and BIC are not used yet since we dont operate outside of Sweden
	return nil
}

func (b *BankingDetail) EncryptFields() error {
	encryptionKey := encryption.DeriveKey(os.Getenv("ENCRYPTION_KEY"), []byte("salt"))
	// Example: Implement the encryption here using your chosen encryption method
	// This is where you would call your encryption function for each field
	var errBN, errAH, errAN, errCN error
	b.BankName, errBN = encryption.Encrypt([]byte(b.BankName), encryptionKey)
	b.AccountHolder, errAH = encryption.Encrypt([]byte(b.AccountHolder), encryptionKey)
	b.AccountNumber, errAN = encryption.Encrypt([]byte(b.AccountNumber), encryptionKey)
	b.ClearingNumber, errCN = encryption.Encrypt([]byte(b.ClearingNumber), encryptionKey)

	if errBN != nil || errAH != nil || errAN != nil || errCN != nil {
		fmt.Errorf("Failed to encrypt fields %v %v %v %v", errBN, errAH, errAN, errCN)
		return errors.New("failed to encrypt fields")
	}

	return nil
}

func (b *BankingDetail) DecryptFields() error {
	encryptionKey := encryption.DeriveKey(os.Getenv("ENCRYPTION_KEY"), []byte("salt"))
	// Example: Implement the decryption here
	var errBN, errAH, errAN, errCN error
	b.BankName, errBN = encryption.Decrypt(b.BankName, encryptionKey)
	b.AccountHolder, errAH = encryption.Decrypt(b.AccountHolder, encryptionKey)
	b.AccountNumber, errAN = encryption.Decrypt(b.AccountNumber, encryptionKey)
	b.ClearingNumber, errCN = encryption.Decrypt(b.ClearingNumber, encryptionKey)

	if errBN != nil || errAH != nil || errAN != nil || errCN != nil {
		fmt.Errorf("Failed to decrypt fields %v %v %v %v", errBN, errAH, errAN, errCN)
		return errors.New("failed to decrypt fields")
	}

	return nil
}
