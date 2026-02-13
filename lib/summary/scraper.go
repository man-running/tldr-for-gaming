package summary

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"main/lib/logger"
	"net/http"
	"strings"
	"time"
)

const (
	huggingFaceAPIBase = "https://huggingface.co/api"
	dailyPapersURL     = huggingFaceAPIBase + "/daily_papers"
	scrapeTimeout      = 30 * time.Second
	maxPapers          = 50
)

// fetchDailyPapers fetches the daily papers from HuggingFace API
func fetchDailyPapers(ctx context.Context) ([]DailyPaperItem, error) {
	logger.Info("Starting fetchDailyPapers", map[string]interface{}{"url": dailyPapersURL})

	client := &http.Client{
		Timeout: scrapeTimeout,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", dailyPapersURL, nil)
	if err != nil {
		logger.Error("Failed to create request", err, map[string]interface{}{"url": dailyPapersURL})
		return nil, fmt.Errorf("failed to create request for %s: %w", dailyPapersURL, err)
	}

	logger.Info("Making HTTP request to HuggingFace API", map[string]interface{}{"url": dailyPapersURL})
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("HTTP request failed", err, map[string]interface{}{"url": dailyPapersURL})
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout fetching daily papers: %w", err)
		}
		return nil, fmt.Errorf("failed to fetch daily papers: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	logger.Info("Received HTTP response", map[string]interface{}{
		"status_code":  resp.StatusCode,
		"content_type": resp.Header.Get("Content-Type"),
	})

	if resp.StatusCode != http.StatusOK {
		// Read response body for debugging
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.Error("Non-200 status code from HuggingFace API", nil, map[string]interface{}{
			"status_code": resp.StatusCode,
			"body":        string(bodyBytes),
		})
		return nil, fmt.Errorf("failed to fetch daily papers: status code %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Read the full response body for logging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("Failed to read response body", err, nil)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	logger.Info("Raw API response received", map[string]interface{}{
		"body_length":     len(bodyBytes),
		"first_200_chars": string(bodyBytes[:min(200, len(bodyBytes))]),
	})

	var response DailyPapersResponse
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		logger.Error("Failed to decode JSON response", err, map[string]interface{}{
			"body": string(bodyBytes),
		})
		return nil, fmt.Errorf("failed to decode daily papers response: %w", err)
	}

	logger.Info("Successfully parsed API response", map[string]interface{}{"papers_count": len(response)})

	if len(response) > 0 {
		logger.Info("First paper sample", map[string]interface{}{
			"title":    response[0].Title,
			"paper_id": response[0].Paper.ID,
		})
	}

	return response, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


// scrapePapers fetches paper information using the HuggingFace API
func scrapePapers(ctx context.Context) ([]Paper, error) {
	logger.Info("Starting scrapePapers function", nil)

	// Fetch daily papers from HuggingFace API
	dailyPaperItems, err := fetchDailyPapers(ctx)
	if err != nil {
		logger.Error("fetchDailyPapers failed", err, nil)
		return nil, fmt.Errorf("failed to fetch daily papers: %w", err)
	}

	logger.Info("Successfully fetched daily papers", map[string]interface{}{"count": len(dailyPaperItems)})

	var papers []Paper
	for i, dailyPaperItem := range dailyPaperItems {
		// Use the main-level title and summary, fall back to paper-level if needed
		title := dailyPaperItem.Title
		summary := dailyPaperItem.Summary
		publishedAt := dailyPaperItem.PublishedAt
		paperID := dailyPaperItem.Paper.ID

		if title == "" {
			title = dailyPaperItem.Paper.Title
		}
		if summary == "" {
			summary = dailyPaperItem.Paper.Summary
		}

		logger.Info("Processing paper", map[string]interface{}{
			"index":    i,
			"paper_id": paperID,
			"title":    title,
		})

		// Construct URL from paper ID
		url := fmt.Sprintf("https://huggingface.co/papers/%s", paperID)
		logger.Info("Constructed URL from ID", map[string]interface{}{
			"paper_id": paperID,
			"url":      url,
		})

		paper := Paper{
			Title:    strings.TrimSpace(title),
			URL:      url,
			Abstract: summary,
			PubDate:  publishedAt,
		}

		papers = append(papers, paper)
		logger.Info("Added paper to collection", map[string]interface{}{
			"title":           paper.Title,
			"url":             paper.URL,
			"abstract_length": len(paper.Abstract),
		})

		// Limit to maxPapers
		if len(papers) >= maxPapers {
			logger.Info("Reached maxPapers limit", map[string]interface{}{"limit": maxPapers})
			break
		}
	}

	logger.Info("scrapePapers completed successfully", map[string]interface{}{"total_papers": len(papers)})
	return papers, nil
}
