package broadcast

import (
	"fmt"
	"main/lib/analytics"
	"main/lib/logger"
	"os"
	"time"

	"github.com/resend/resend-go/v2"
	feedpkg "main/lib/feed"
)

// SendDailyBroadcast orchestrates fetching, parsing, and sending the broadcast email.
func SendDailyBroadcast() error {
	logger.Info("Starting daily broadcast process", nil)

	// 1. Get Feed Data
	feed, err := ParseRssFeed()
	if err != nil {
		return fmt.Errorf("failed during feed parsing: %w", err)
	}

	if feed == nil || feed.LastBuildDate == "" {
		logger.Error("No valid feed data available to send broadcast", nil, nil)
		return fmt.Errorf("no valid feed data")
	}

	// Guard: after 07:05 UTC, ensure the feed date is today; warn if not
	if t, err := time.Parse(time.RFC1123Z, feed.LastBuildDate); err == nil {
		now := time.Now().UTC()
		afterSevenOhFive := now.Hour() > 7 || (now.Hour() == 7 && now.Minute() >= 5)
		sameYMD := t.UTC().Year() == now.Year() && t.UTC().Month() == now.Month() && t.UTC().Day() == now.Day()
		if afterSevenOhFive && !sameYMD {
			logger.Warn("Daily broadcast feed date appears stale after 07:05 UTC", map[string]interface{}{"feedLastBuildDate": feed.LastBuildDate})
		}
	}

	// 2. Generate Email HTML
	emailHTML, err := generateEmailHTML(*feed)
	if err != nil {
		return fmt.Errorf("failed to generate email HTML: %w", err)
	}

	// 3. Set up Resend Client
	apiKey := os.Getenv("RESEND_API_KEY")
	audienceID := os.Getenv("RESEND_AUDIENCE_ID")
	fromEmail := os.Getenv("RESEND_FROM_EMAIL")

	if apiKey == "" || audienceID == "" || fromEmail == "" {
		return fmt.Errorf("missing RESEND_API_KEY, RESEND_AUDIENCE_ID, or RESEND_FROM_EMAIL environment variables")
	}

	client := resend.NewClient(apiKey)

	// 4. Create Broadcast
	dateStr := formatDateForSubject(feed.LastBuildDate)
	subject := fmt.Sprintf("Takara TLDR: %s", dateStr)

	logger.Info("Creating Resend broadcast", map[string]interface{}{"subject": subject, "audienceId": audienceID})

	createParams := &resend.CreateBroadcastRequest{
		From:       fromEmail,
		Subject:    subject,
		Html:       emailHTML,
		AudienceId: audienceID,
	}

	createdBroadcast, err := client.Broadcasts.Create(createParams)
	if err != nil {
		logger.Error("Failed to create Resend broadcast", err, nil)
		return fmt.Errorf("resend broadcast creation failed: %w", err)
	}

	if createdBroadcast.Id == "" {
		logger.Error("Resend broadcast creation returned no data", nil, nil)
		return fmt.Errorf("resend broadcast creation returned no data")
	}

	logger.Info("Successfully created Resend broadcast", map[string]interface{}{"broadcastId": createdBroadcast.Id})

	// Best-effort: store the feed in blob storage using the feed's own date
	go func(f RssFeed) {
		items := make([]feedpkg.FeedItem, 0, len(f.Items))
		for _, it := range f.Items {
			items = append(items, feedpkg.FeedItem{
				Title:       it.Title,
				Link:        it.Link,
				Description: it.Description,
				PubDate:     it.PubDate,
				GUID:        feedpkg.GUIDString(it.GUID),
			})
		}
		converted := &feedpkg.RssFeed{
			Title:         f.Title,
			Description:   f.Description,
			Link:          f.Link,
			LastBuildDate: f.LastBuildDate,
			Items:         items,
		}
		if err := feedpkg.StoreTldrFeed(converted); err != nil {
			logger.Warn("Failed to store TLDR feed from broadcast", map[string]interface{}{"error": err.Error()})
		}
	}(*feed)

	// 5. Send Broadcast
	logger.Info("Sending Resend broadcast", map[string]interface{}{"broadcastId": createdBroadcast.Id})
	sendParams := &resend.SendBroadcastRequest{
		BroadcastId: createdBroadcast.Id,
	}
	_, sendErr := client.Broadcasts.Send(sendParams)
	if sendErr != nil {
		logger.Error("Failed to send Resend broadcast", sendErr, map[string]interface{}{"broadcastId": createdBroadcast.Id})
		return fmt.Errorf("resend broadcast send failed: %w", sendErr)
	}

	logger.Info("Successfully sent daily broadcast", map[string]interface{}{"broadcastId": createdBroadcast.Id})
	_ = analytics.Track("broadcast_sent", createdBroadcast.Id, map[string]interface{}{"subject": subject})
	return nil
}

// formatDateForSubject formats the date specifically for the email subject line.
func formatDateForSubject(dateStr string) string {
	if dateStr == "" {
		return time.Now().UTC().Format("January 2, 2006")
	}
	// Use the same robust parsing as the template formatter.
	layouts := []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822, time.RubyDate}
	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, dateStr)
		if err == nil {
			break
		}
	}
	if err != nil {
		return time.Now().UTC().Format("January 2, 2006")
	}
	return t.Format("January 2, 2006")
}
