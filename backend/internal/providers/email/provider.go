// Package email provides email service abstractions for sending emails.
// It defines the EmailProvider interface that all email implementations must implement,
// allowing for provider-agnostic email sending (SMTP, SendGrid, SES, etc.).
package email

import "context"

// Email represents an email message to be sent.
type Email struct {
	To       string // Recipient email address
	Subject  string // Email subject line
	HTMLBody string // HTML content of the email
	TextBody string // Plain text content of the email
	Language string // Language code (e.g., "en", "fr") for the email
}

// EmailProvider defines the interface that all email providers must implement.
// This abstraction allows swapping email providers without code changes.
type EmailProvider interface {
	// SendEmail sends an email message.
	// Returns an error if the email cannot be sent.
	SendEmail(ctx context.Context, email *Email) error
}
