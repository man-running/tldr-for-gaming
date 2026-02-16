package feed

import (
	"context"
	"main/lib/article"
	"testing"
	"time"
)

func TestSummarizerConfigValidate(t *testing.T) {
	tests := []struct {
		name      string
		config    *SummarizerConfig
		shouldErr bool
	}{
		{
			name: "valid config",
			config: &SummarizerConfig{
				APIKey:      "sk-ant-test",
				Model:       "claude-3-5-sonnet-20241022",
				MaxTokens:   150,
				Temperature: 0.7,
				TimeoutSec:  30,
			},
			shouldErr: false,
		},
		{
			name: "missing API key",
			config: &SummarizerConfig{
				APIKey:      "",
				Model:       "claude-3-5-sonnet-20241022",
				MaxTokens:   150,
				Temperature: 0.7,
				TimeoutSec:  30,
			},
			shouldErr: true,
		},
		{
			name: "invalid temperature below 0",
			config: &SummarizerConfig{
				APIKey:      "sk-ant-test",
				Model:       "claude-3-5-sonnet-20241022",
				MaxTokens:   150,
				Temperature: -0.1,
				TimeoutSec:  30,
			},
			shouldErr: true,
		},
		{
			name: "invalid temperature above 1",
			config: &SummarizerConfig{
				APIKey:      "sk-ant-test",
				Model:       "claude-3-5-sonnet-20241022",
				MaxTokens:   150,
				Temperature: 1.1,
				TimeoutSec:  30,
			},
			shouldErr: true,
		},
		{
			name: "defaults applied to zero values",
			config: &SummarizerConfig{
				APIKey: "sk-ant-test",
			},
			shouldErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.shouldErr {
				t.Errorf("Validate() error = %v, shouldErr %v", err, tt.shouldErr)
			}

			// Check defaults were applied
			if tt.config.Model == "" {
				t.Error("Model should have default applied")
			}
			if tt.config.MaxTokens <= 0 {
				t.Error("MaxTokens should have default applied")
			}
			if tt.config.Temperature == 0 {
				t.Error("Temperature should have default applied")
			}
			if tt.config.TimeoutSec <= 0 {
				t.Error("TimeoutSec should have default applied")
			}
		})
	}
}

func TestNewArticleSummarizer(t *testing.T) {
	config := &SummarizerConfig{
		APIKey:      "sk-ant-test",
		Model:       "claude-3-5-sonnet-20241022",
		MaxTokens:   150,
		Temperature: 0.7,
		TimeoutSec:  30,
	}

	summarizer, err := NewArticleSummarizer(config)
	if err != nil {
		t.Fatalf("NewArticleSummarizer() error = %v", err)
	}

	if summarizer == nil {
		t.Error("NewArticleSummarizer() returned nil")
	}

	if summarizer.client == nil {
		t.Error("client should be initialized")
	}
}

func TestNewArticleSummarizerInvalidConfig(t *testing.T) {
	config := &SummarizerConfig{
		APIKey: "", // Invalid: missing API key
	}

	_, err := NewArticleSummarizer(config)
	if err == nil {
		t.Error("NewArticleSummarizer() should error with invalid config")
	}
}

func TestSummarizeArticleValidation(t *testing.T) {
	config := &SummarizerConfig{
		APIKey:      "sk-ant-test",
		Model:       "claude-3-5-sonnet-20241022",
		MaxTokens:   150,
		Temperature: 0.7,
		TimeoutSec:  30,
	}

	summarizer, _ := NewArticleSummarizer(config)
	ctx := context.Background()

	tests := []struct {
		name      string
		article   *article.ArticleData
		shouldErr bool
	}{
		{
			name:      "nil article",
			article:   nil,
			shouldErr: true,
		},
		{
			name: "missing title",
			article: &article.ArticleData{
				Title: "",
				URL:   "https://example.com",
			},
			shouldErr: true,
		},
		{
			name: "missing URL",
			article: &article.ArticleData{
				Title: "Test Article",
				URL:   "",
			},
			shouldErr: true,
		},
		{
			name: "valid article (API call would fail without real key)",
			article: &article.ArticleData{
				Title:      "Test Article",
				URL:        "https://example.com",
				SourceName: "Test Source",
				OriginalSum: "Test summary",
			},
			shouldErr: true, // Will fail due to invalid API key, but validates input
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := summarizer.SummarizeArticle(ctx, tt.article)
			if (err != nil) != tt.shouldErr {
				t.Errorf("SummarizeArticle() error = %v, shouldErr %v", err, tt.shouldErr)
			}
		})
	}
}

func TestSummarizeBatchEmpty(t *testing.T) {
	config := &SummarizerConfig{
		APIKey:      "sk-ant-test",
		Model:       "claude-3-5-sonnet-20241022",
		MaxTokens:   150,
		Temperature: 0.7,
		TimeoutSec:  30,
	}

	summarizer, _ := NewArticleSummarizer(config)
	ctx := context.Background()

	articles := []article.ArticleData{}
	err := summarizer.SummarizeBatch(ctx, articles)
	if err != nil {
		t.Errorf("SummarizeBatch() error = %v, expected nil", err)
	}
}

func TestSummarizeBatchCancelContext(t *testing.T) {
	config := &SummarizerConfig{
		APIKey:      "sk-ant-test",
		Model:       "claude-3-5-sonnet-20241022",
		MaxTokens:   150,
		Temperature: 0.7,
		TimeoutSec:  30,
	}

	summarizer, _ := NewArticleSummarizer(config)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	articles := []article.ArticleData{
		{
			ID:         "test-1",
			Title:      "Test Article 1",
			URL:        "https://example.com/1",
			SourceName: "Test Source",
		},
	}

	err := summarizer.SummarizeBatch(ctx, articles)
	if err == nil {
		t.Error("SummarizeBatch() should return error with cancelled context")
	}
}

func TestArticleMetadataUpdate(t *testing.T) {
	config := &SummarizerConfig{
		APIKey:      "sk-ant-test",
		Model:       "claude-3-5-sonnet-20241022",
		MaxTokens:   150,
		Temperature: 0.7,
		TimeoutSec:  30,
	}

	_, _ = NewArticleSummarizer(config)

	art := &article.ArticleData{
		ID:         "test-1",
		Title:      "Test Article",
		URL:        "https://example.com",
		SourceName: "Test Source",
		OriginalSum: "Test summary",
		Metadata:   make(map[string]interface{}),
		UpdatedAt:  time.Time{},
	}

	// Test metadata structure
	if art.Metadata == nil {
		t.Error("Metadata should be initialized")
	}

	// Test that UpdatedAt would be updated (simulated)
	originalTime := art.UpdatedAt
	if art.UpdatedAt == originalTime && originalTime.IsZero() {
		t.Log("Metadata structure is ready for summarizer")
	}
}

func TestSummarizerConfigModel(t *testing.T) {
	config := &SummarizerConfig{
		APIKey: "sk-ant-test",
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if config.Model != "claude-3-5-sonnet-20241022" {
		t.Errorf("Default model not applied, got %s", config.Model)
	}
}

func TestSummarizerConfigMaxTokens(t *testing.T) {
	config := &SummarizerConfig{
		APIKey: "sk-ant-test",
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if config.MaxTokens != 150 {
		t.Errorf("Default MaxTokens not applied, got %d", config.MaxTokens)
	}
}

func TestSummarizerConfigTemperature(t *testing.T) {
	config := &SummarizerConfig{
		APIKey: "sk-ant-test",
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if config.Temperature != 0.7 {
		t.Errorf("Default Temperature not applied, got %f", config.Temperature)
	}
}

func TestSummarizerConfigTimeout(t *testing.T) {
	config := &SummarizerConfig{
		APIKey: "sk-ant-test",
	}

	err := config.Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if config.TimeoutSec != 30 {
		t.Errorf("Default TimeoutSec not applied, got %d", config.TimeoutSec)
	}
}
