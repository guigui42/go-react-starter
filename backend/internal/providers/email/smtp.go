// Package email provides email service implementations.
package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/wneessen/go-mail"
)

// SMTPConfig holds SMTP server configuration.
type SMTPConfig struct {
	Host        string        // SMTP server hostname
	Port        int           // SMTP server port
	Username    string        // SMTP authentication username
	Password    string        // SMTP authentication password
	UseTLS      bool          // Whether to use TLS
	FromAddress string        // Default sender email address
	FromName    string        // Default sender display name
	Timeout     time.Duration // Connection timeout
}

// SMTPProvider implements EmailProvider using SMTP.
type SMTPProvider struct {
	config SMTPConfig
}

// NewSMTPProvider creates a new SMTP email provider.
func NewSMTPProvider(config SMTPConfig) *SMTPProvider {
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}
	return &SMTPProvider{
		config: config,
	}
}

// SendEmail sends an email using SMTP.
func (p *SMTPProvider) SendEmail(ctx context.Context, email *Email) error {
	// Create new message
	msg := mail.NewMsg()

	// Set sender
	if err := msg.From(p.config.FromAddress); err != nil {
		return fmt.Errorf("failed to set from address: %w", err)
	}

	// Set recipient
	if err := msg.To(email.To); err != nil {
		return fmt.Errorf("failed to set to address: %w", err)
	}

	// Set subject
	msg.Subject(email.Subject)

	// Set body - in multipart/alternative, parts go from simplest to richest.
	// Email clients display the LAST part they can render, so HTML must come last.
	if email.TextBody != "" && email.HTMLBody != "" {
		msg.SetBodyString(mail.TypeTextPlain, email.TextBody)
		msg.AddAlternativeString(mail.TypeTextHTML, email.HTMLBody)
	} else if email.HTMLBody != "" {
		msg.SetBodyString(mail.TypeTextHTML, email.HTMLBody)
	} else if email.TextBody != "" {
		msg.SetBodyString(mail.TypeTextPlain, email.TextBody)
	}

	// Configure client options
	opts := []mail.Option{
		mail.WithPort(p.config.Port),
		mail.WithTimeout(p.config.Timeout),
	}

	// Configure TLS
	if p.config.UseTLS {
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSMandatory))
		opts = append(opts, mail.WithTLSConfig(&tls.Config{
			ServerName: p.config.Host,
			MinVersion: tls.VersionTLS12,
		}))
	} else {
		opts = append(opts, mail.WithTLSPortPolicy(mail.TLSOpportunistic))
	}

	// Add authentication if credentials provided
	// Use AutoDiscover to automatically select the best auth mechanism
	// (supports PLAIN, LOGIN, CRAM-MD5, SCRAM-SHA-*, etc.)
	if p.config.Username != "" && p.config.Password != "" {
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthAutoDiscover))
		opts = append(opts, mail.WithUsername(p.config.Username))
		opts = append(opts, mail.WithPassword(p.config.Password))
	}

	// Create SMTP client
	client, err := mail.NewClient(p.config.Host, opts...)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}

	// Send the email
	if err := client.DialAndSendWithContext(ctx, msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// NoOpProvider is a no-operation email provider for testing or when email is disabled.
type NoOpProvider struct{}

// NewNoOpProvider creates a new no-op email provider.
func NewNoOpProvider() *NoOpProvider {
	return &NoOpProvider{}
}

// SendEmail does nothing and returns nil.
func (p *NoOpProvider) SendEmail(_ context.Context, _ *Email) error {
	return nil
}
