package services

import (
	"fmt"
	"net/smtp"
	"os"
	"regexp"
	"strings"
)

type EmailService struct {
	smtpHost     string
	smtpPort     string
	smtpUsername string
	smtpPassword string
	fromEmail    string
	fromName     string
	frontendURL  string
}

func NewEmailService() *EmailService {
	return &EmailService{
		smtpHost:     os.Getenv("SMTP_HOST"),
		smtpPort:     os.Getenv("SMTP_PORT"),
		smtpUsername: os.Getenv("SMTP_USERNAME"),
		smtpPassword: os.Getenv("SMTP_PASSWORD"),
		fromEmail:    os.Getenv("SMTP_FROM_EMAIL"),
		fromName:     os.Getenv("SMTP_FROM_NAME"),
		frontendURL:  getEnv("FRONTEND_URL", "http://localhost:3000"),
	}
}

func (s *EmailService) SendResetPasswordEmail(toEmail, toName, resetToken string) error {
	// ✅ VALIDATION 1: Check required fields
	if err := s.validateInputs(toEmail, toName, resetToken); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	// ✅ VALIDATION 2: Check SMTP configuration
	if err := s.validateSMTPConfig(); err != nil {
		return fmt.Errorf("SMTP configuration error: %w", err)
	}

	// ✅ SANITIZATION: Clean inputs
	toEmail = strings.TrimSpace(strings.ToLower(toEmail))
	toName = strings.TrimSpace(toName)
	resetToken = strings.TrimSpace(resetToken)

	// Build reset link
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", s.frontendURL, resetToken)

	subject := "Reset Password - Shorteny"
	body := s.buildEmailHTML(toName, resetLink)

	return s.sendEmail(toEmail, subject, body)
}

// ✅ NEW: Validate all inputs before processing
func (s *EmailService) validateInputs(toEmail, toName, resetToken string) error {
	// 1. Check email is not empty
	if toEmail == "" {
		return fmt.Errorf("recipient email is required")
	}

	// 2. Validate email format
	if !isValidEmail(toEmail) {
		return fmt.Errorf("invalid email format: %s", toEmail)
	}

	// 3. Check name is not empty
	if toName == "" {
		return fmt.Errorf("recipient name is required")
	}

	// 4. Check reset token is not empty
	if resetToken == "" {
		return fmt.Errorf("reset token is required")
	}

	// 5. Validate token format (UUID)
	if len(resetToken) < 10 {
		return fmt.Errorf("invalid reset token format")
	}

	return nil
}

// ✅ NEW: Validate SMTP configuration
func (s *EmailService) validateSMTPConfig() error {
	if s.smtpHost == "" {
		return fmt.Errorf("SMTP_HOST is not configured")
	}

	if s.smtpPort == "" {
		return fmt.Errorf("SMTP_PORT is not configured")
	}

	if s.smtpUsername == "" {
		return fmt.Errorf("SMTP_USERNAME is not configured")
	}

	if s.smtpPassword == "" {
		return fmt.Errorf("SMTP_PASSWORD is not configured")
	}

	if s.fromEmail == "" {
		return fmt.Errorf("SMTP_FROM_EMAIL is not configured")
	}

	// Validate from email format
	if !isValidEmail(s.fromEmail) {
		return fmt.Errorf("invalid SMTP_FROM_EMAIL format: %s", s.fromEmail)
	}

	return nil
}

// ✅ NEW: Build HTML email template (separated for clarity)
func (s *EmailService) buildEmailHTML(toName, resetLink string) string {
	// Escape HTML special characters in name to prevent XSS
	toName = escapeHTML(toName)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Reset Password</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px; border: 1px solid #ddd; border-radius: 5px;">
        <h2 style="color: #4F46E5;">🔐 Reset Your Password</h2>
        <p>Hi <strong>%s</strong>,</p>
        <p>We received a request to reset your password for your Shorteny account.</p>
        <p>Click the button below to create a new password:</p>
        <div style="text-align: center; margin: 30px 0;">
            <a href="%s" style="background-color: #4F46E5; color: white; padding: 14px 40px; text-decoration: none; border-radius: 5px; display: inline-block; font-weight: bold;">Reset Password</a>
        </div>
        <p>Or copy and paste this link into your browser:</p>
        <p style="word-break: break-all; color: #4F46E5; background: #f5f5f5; padding: 10px; border-radius: 4px;">%s</p>
        <p><strong>⏰ This link will expire in 1 hour.</strong></p>
        <p style="margin-top: 30px; color: #666;">If you didn't request a password reset, please ignore this email or contact support if you have concerns.</p>
        <hr style="margin: 30px 0; border: none; border-top: 1px solid #ddd;">
        <p style="font-size: 12px; color: #999; text-align: center;">
            This is an automated message from Shorteny<br>
            Please do not reply to this email.
        </p>
    </div>
</body>
</html>
	`, toName, resetLink, resetLink)
}

func (s *EmailService) sendEmail(to, subject, body string) error {
	// ✅ SECURITY: Trim whitespace from password (common issue)
	password := strings.TrimSpace(s.smtpPassword)

	// Setup authentication
	auth := smtp.PlainAuth("", s.smtpUsername, password, s.smtpHost)

	// Compose email message
	from := fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail)
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte(fmt.Sprintf("From: %s\nTo: %s\nSubject: %s\n%s\n%s", from, to, subject, mime, body))

	// Send email with proper error handling
	addr := fmt.Sprintf("%s:%s", s.smtpHost, s.smtpPort)

	err := smtp.SendMail(addr, auth, s.fromEmail, []string{to}, msg)
	if err != nil {
		// ✅ Enhanced error message with troubleshooting hints
		return fmt.Errorf("SMTP send failed (check credentials and network): %w", err)
	}

	return nil
}

// ✅ NEW: Email validation using regex
func isValidEmail(email string) bool {
	// RFC 5322 compliant email regex (simplified)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

// ✅ NEW: HTML escape to prevent XSS in emails
func escapeHTML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(s)
}

// ✅ NEW: Helper function for environment variables with default
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
