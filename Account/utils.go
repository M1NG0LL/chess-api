package account

import (
	"fmt"
	"math/rand"
	"net/smtp"
	"os"
	"time"
)

// Function to send the activation email
func SendEmail(to string, subject string, message string) error {
	from := os.Getenv("EMAIL")
	password := os.Getenv("PASSWORD")

	// SMTP server configuration
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Email message body, formatted to include subject and message content
	emailMessage := []byte(fmt.Sprintf("Subject: %s\r\n\r\n%s", subject, message))

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Send the email
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, emailMessage)
	if err != nil {
		return fmt.Errorf("failed to send email to %s: %w", to, err) // Added context to the error
	}

	return nil
}

// Creating Code of 6 chars
func GenerateCode(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, length)
	for i := range code {
		code[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(code)
}