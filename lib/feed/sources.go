package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

// NewsSource represents a configured iGaming news source
type NewsSource struct {
	ID           string `json:"id"`           // Unique identifier
	Name         string `json:"name"`         // Display name (e.g., "iGamingBusiness")
	URL          string `json:"url"`          // Publisher website URL
	FeedURL      string `json:"feedUrl"`      // RSS feed URL
	Category     string `json:"category"`     // Primary category
	Active       bool   `json:"active"`       // Whether to include in aggregation
	Priority     int    `json:"priority"`     // Higher = more important in ranking (1-10)
	ScrapingType string `json:"scrapingType"` // "rss", "scrape", "api"
	Timeout      int    `json:"timeout"`      // Request timeout in milliseconds
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// SourceManager manages news sources and their feeds
type SourceManager struct {
	mu      sync.RWMutex
	sources map[string]*NewsSource
}

// NewSourceManager creates a new source manager
func NewSourceManager() *SourceManager {
	return &SourceManager{
		sources: make(map[string]*NewsSource),
	}
}

// LoadDefaultSources loads the default iGaming news sources
func (sm *SourceManager) LoadDefaultSources() error {
	defaultSources := []NewsSource{
		{
			ID:           "igamingbusiness",
			Name:         "iGamingBusiness",
			URL:          "https://www.igamingbusiness.com",
			FeedURL:      "https://www.igamingbusiness.com/feed/",
			Category:     "Business",
			Active:       true,
			Priority:     10,
			ScrapingType: "rss",
			Timeout:      10000,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "gamblinginsider",
			Name:         "Gambling Insider",
			URL:          "https://www.gamblinginsider.com",
			FeedURL:      "https://www.gamblinginsider.com/feed/",
			Category:     "Business",
			Active:       true,
			Priority:     9,
			ScrapingType: "rss",
			Timeout:      10000,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "eganingreview",
			Name:         "eGaming Review",
			URL:          "https://www.egamingreview.com",
			FeedURL:      "https://www.egamingreview.com/feed/",
			Category:     "Regulations",
			Active:       true,
			Priority:     8,
			ScrapingType: "rss",
			Timeout:      10000,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "sporttech",
			Name:         "Sportech",
			URL:          "https://www.sportech.com",
			FeedURL:      "https://www.sportech.com/feed/",
			Category:     "Sports Betting",
			Active:       true,
			Priority:     7,
			ScrapingType: "rss",
			Timeout:      10000,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		{
			ID:           "betindustry",
			Name:         "Betting Industry",
			URL:          "https://www.bettingindustry.com",
			FeedURL:      "https://www.bettingindustry.com/feed/",
			Category:     "Sports Betting",
			Active:       true,
			Priority:     7,
			ScrapingType: "rss",
			Timeout:      10000,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i := range defaultSources {
		sm.sources[defaultSources[i].ID] = &defaultSources[i]
	}

	return nil
}

// LoadSourcesFromFile loads sources from a JSON configuration file
func (sm *SourceManager) LoadSourcesFromFile(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read sources file: %w", err)
	}

	var sources []NewsSource
	if err := json.Unmarshal(data, &sources); err != nil {
		return fmt.Errorf("failed to parse sources JSON: %w", err)
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i := range sources {
		sm.sources[sources[i].ID] = &sources[i]
	}

	return nil
}

// AddSource adds a new news source
func (sm *SourceManager) AddSource(source *NewsSource) error {
	if source.ID == "" {
		return fmt.Errorf("source ID cannot be empty")
	}
	if source.Name == "" {
		return fmt.Errorf("source name cannot be empty")
	}
	if source.FeedURL == "" {
		return fmt.Errorf("source feed URL cannot be empty")
	}

	source.CreatedAt = time.Now()
	source.UpdatedAt = time.Now()

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.sources[source.ID] = source
	return nil
}

// UpdateSource updates an existing news source
func (sm *SourceManager) UpdateSource(id string, updates *NewsSource) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	source, exists := sm.sources[id]
	if !exists {
		return fmt.Errorf("source not found: %s", id)
	}

	if updates.Name != "" {
		source.Name = updates.Name
	}
	if updates.FeedURL != "" {
		source.FeedURL = updates.FeedURL
	}
	if updates.Category != "" {
		source.Category = updates.Category
	}
	if updates.Priority > 0 {
		source.Priority = updates.Priority
	}
	if updates.ScrapingType != "" {
		source.ScrapingType = updates.ScrapingType
	}
	if updates.Timeout > 0 {
		source.Timeout = updates.Timeout
	}

	source.Active = updates.Active
	source.UpdatedAt = time.Now()

	return nil
}

// GetSource retrieves a source by ID
func (sm *SourceManager) GetSource(id string) (*NewsSource, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	source, exists := sm.sources[id]
	if !exists {
		return nil, fmt.Errorf("source not found: %s", id)
	}

	return source, nil
}

// GetActiveSources returns all active sources sorted by priority
func (sm *SourceManager) GetActiveSources() []*NewsSource {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var active []*NewsSource
	for _, source := range sm.sources {
		if source.Active {
			active = append(active, source)
		}
	}

	// Sort by priority (descending)
	for i := 0; i < len(active); i++ {
		for j := i + 1; j < len(active); j++ {
			if active[j].Priority > active[i].Priority {
				active[i], active[j] = active[j], active[i]
			}
		}
	}

	return active
}

// GetSourcesByCategory returns sources filtered by category
func (sm *SourceManager) GetSourcesByCategory(category string) []*NewsSource {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var matching []*NewsSource
	for _, source := range sm.sources {
		if source.Active && strings.EqualFold(source.Category, category) {
			matching = append(matching, source)
		}
	}

	return matching
}

// ListSources returns all sources
func (sm *SourceManager) ListSources() []*NewsSource {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sources []*NewsSource
	for _, source := range sm.sources {
		sources = append(sources, source)
	}

	return sources
}

// DisableSource disables a source by ID
func (sm *SourceManager) DisableSource(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	source, exists := sm.sources[id]
	if !exists {
		return fmt.Errorf("source not found: %s", id)
	}

	source.Active = false
	source.UpdatedAt = time.Now()
	return nil
}

// EnableSource enables a source by ID
func (sm *SourceManager) EnableSource(id string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	source, exists := sm.sources[id]
	if !exists {
		return fmt.Errorf("source not found: %s", id)
	}

	source.Active = true
	source.UpdatedAt = time.Now()
	return nil
}

// ExportSources exports sources to JSON format
func (sm *SourceManager) ExportSources() (string, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var sources []*NewsSource
	for _, source := range sm.sources {
		sources = append(sources, source)
	}

	data, err := json.MarshalIndent(sources, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal sources: %w", err)
	}

	return string(data), nil
}

// GetSourceCount returns the number of sources
func (sm *SourceManager) GetSourceCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return len(sm.sources)
}

// GetActiveSourceCount returns the number of active sources
func (sm *SourceManager) GetActiveSourceCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	count := 0
	for _, source := range sm.sources {
		if source.Active {
			count++
		}
	}
	return count
}

// Validate checks if all required sources are properly configured
func (sm *SourceManager) Validate() error {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if len(sm.sources) == 0 {
		return fmt.Errorf("no sources configured")
	}

	activeSources := 0
	for _, source := range sm.sources {
		if source.Active {
			activeSources++
		}

		if source.ID == "" || source.Name == "" || source.FeedURL == "" {
			return fmt.Errorf("source %s has missing required fields", source.ID)
		}

		if source.Priority < 1 || source.Priority > 10 {
			return fmt.Errorf("source %s has invalid priority (must be 1-10)", source.ID)
		}

		if source.ScrapingType != "rss" && source.ScrapingType != "scrape" && source.ScrapingType != "api" {
			return fmt.Errorf("source %s has invalid scraping type", source.ID)
		}
	}

	if activeSources == 0 {
		return fmt.Errorf("no active sources configured")
	}

	return nil
}
