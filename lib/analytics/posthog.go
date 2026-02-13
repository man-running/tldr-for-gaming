package analytics

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"time"
)

type PostHogEvent struct {
	APIKey     string                 `json:"api_key"`
	Event      string                 `json:"event"`
	DistinctID string                 `json:"distinct_id"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Timestamp  string                 `json:"timestamp,omitempty"`
}

func Track(event string, distinctID string, properties map[string]interface{}) error {
	apiKey := os.Getenv("POSTHOG_API_KEY")
	if apiKey == "" {
		return nil
	}

	payload := PostHogEvent{
		APIKey:     apiKey,
		Event:      event,
		DistinctID: distinctID,
		Properties: properties,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://eu.i.posthog.com/i/v0/e/", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 2 * time.Second}
	go func() {
		_, _ = client.Do(req)
	}()

	return nil
}

