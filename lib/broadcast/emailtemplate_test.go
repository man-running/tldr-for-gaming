package broadcast

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

// TestRenderDailyEmailTemplate fetches the RSS feed using ParseRssFeed,
// renders the daily email HTML with the existing template, and writes
// the output to tests/email/daily-email.html for manual viewing.
func TestRenderDailyEmailTemplate(t *testing.T) {
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		t.Skip("BASE_URL is not set; skipping email template render test")
	}

	feed, err := ParseRssFeed()
	if err != nil {
		t.Fatalf("failed to parse RSS feed: %v", err)
	}
	if feed == nil {
		t.Fatalf("ParseRssFeed returned nil feed")
	}

	html, err := generateEmailHTML(*feed)
	if err != nil {
		t.Fatalf("failed to generate email HTML: %v", err)
	}

	outDir := filepath.Join("tests", "email")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}

	outPath := filepath.Join(outDir, "daily-email.html")
	if err := os.WriteFile(outPath, []byte(html), 0o644); err != nil {
		t.Fatalf("failed to write HTML output: %v", err)
	}

	t.Logf("Wrote rendered email HTML to %s", outPath)
}

// TestRenderDailyEmailTemplateFromAPI fetches the TLDR JSON feed from a local API
// (default http://localhost:3000/api/feed or TLDR_API_FEED_URL), renders the email
// HTML with the existing template, and writes the output to tests/email/daily-email.html.
func TestRenderDailyEmailTemplateFromAPI(t *testing.T) {
	apiURL := os.Getenv("TLDR_API_FEED_URL")
	if apiURL == "" {
		base := os.Getenv("BASE_URL")
		if base != "" {
			// trim any trailing slash
			if base[len(base)-1] == '/' {
				base = base[:len(base)-1]
			}
			apiURL = base + "/api/feed"
		} else {
			apiURL = "http://localhost:3000/api/feed"
		}
	}

	resp, err := http.Get(apiURL)
	if err != nil {
		t.Fatalf("failed to fetch TLDR API feed (%s): %v", apiURL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status from TLDR API feed: %s; body=%s", resp.Status, string(b))
	}

	var feed RssFeed
	if err := json.NewDecoder(resp.Body).Decode(&feed); err != nil {
		t.Fatalf("failed to decode TLDR API feed JSON: %v", err)
	}

	html, err := generateEmailHTML(feed)
	if err != nil {
		t.Fatalf("failed to generate email HTML: %v", err)
	}

	outDir := filepath.Join("tests", "email")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatalf("failed to create output directory: %v", err)
	}

	outPath := filepath.Join(outDir, "daily-email.html")
	if err := os.WriteFile(outPath, []byte(html), 0o644); err != nil {
		t.Fatalf("failed to write HTML output: %v", err)
	}

	t.Logf("Wrote rendered email HTML to %s (source: %s)", outPath, apiURL)
}
