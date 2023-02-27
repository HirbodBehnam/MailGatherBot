package util

import (
	"crypto/rand"
	"encoding/base64"
	"net/mail"
)

// ValidateEmail validates an email and returns the valid email.
func ValidateEmail(email string) (string, bool) {
	address, err := mail.ParseAddress(email)
	if err != nil {
		return "", false
	}
	return address.Address, true
}

// RandomID generates a random string ID
func RandomID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}
