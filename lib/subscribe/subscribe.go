package subscribe

import (
	"fmt"
	"main/lib/analytics"
	"main/lib/logger"
	"os"
	"regexp"

	"github.com/resend/resend-go/v2"
)

// emailRegex is a simple regex to validate email format.
var emailRegex = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// SubscribeEmail adds a user to the Resend audience and sends them a welcome email.
func SubscribeEmail(email string) error {
	// 1. Validate email format
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	// 2. Get Resend configuration from environment
	apiKey := os.Getenv("RESEND_API_KEY")
	audienceID := os.Getenv("RESEND_AUDIENCE_ID")
	fromEmail := os.Getenv("RESEND_FROM_EMAIL")

	if apiKey == "" || audienceID == "" || fromEmail == "" {
		return fmt.Errorf("missing Resend configuration in environment variables")
	}

	client := resend.NewClient(apiKey)

	// 3. Add contact to Resend audience (critical step)
	createContactParams := &resend.CreateContactRequest{
		Email:      email,
		AudienceId: audienceID,
	}
	_, err := client.Contacts.Create(createContactParams)
	if err != nil {
		// This is a critical failure. If we can't add them, the subscription failed.
		return fmt.Errorf("failed to add contact to Resend audience: %w", err)
	}

	_ = analytics.Track("email_subscribed", email, map[string]interface{}{
		"source": "subscribe",
	})

	// 4. Fetch current feed for welcome email (non-critical)
	feed, err := ParseRssFeed()
	if err != nil {
		logger.Error("Failed to fetch current feed for welcome email", err, nil)
		// Do not return; continue to send a welcome email without the feed.
	}

	// 5. Generate and send welcome email (non-critical)
	emailHTML, err := GenerateWelcomeEmailHTML(feed)
	if err != nil {
		logger.Error("Failed to generate welcome email HTML", err, nil)
		// Do not return; the main subscription was successful.
		return nil
	}

	sendEmailParams := &resend.SendEmailRequest{
		From:    fromEmail,
		To:      []string{email},
		Subject: "Welcome to Takara TLDR",
		Html:    emailHTML,
	}
	_, err = client.Emails.Send(sendEmailParams)
	if err != nil {
		logger.Error("Failed to send welcome email", err, nil)
		// Do not return; the main subscription was successful.
	}

	return nil
}
