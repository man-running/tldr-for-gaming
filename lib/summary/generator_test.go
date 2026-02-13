package summary

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestToTLDRLink(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HuggingFace URL with arxiv ID",
			input:    "https://huggingface.co/papers/2509.06652",
			expected: "https://tldr.takara.ai/p/2509.06652",
		},
		{
			name:     "HuggingFace URL with different arxiv ID",
			input:    "https://huggingface.co/papers/2509.10441",
			expected: "https://tldr.takara.ai/p/2509.10441",
		},
		{
			name:     "Non-HuggingFace URL",
			input:    "https://example.com/paper/123",
			expected: "https://example.com/paper/123",
		},
		{
			name:     "Empty URL",
			input:    "",
			expected: "",
		},
		{
			name:     "HuggingFace URL without arxiv ID",
			input:    "https://huggingface.co/papers/",
			expected: "https://huggingface.co/papers/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toTLDRLink(tt.input)
			if result != tt.expected {
				t.Errorf("toTLDRLink(%q) = %q; expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGeneratePapersRSS_TLDRLinks(t *testing.T) {
	papers := []Paper{
		{
			Title:    "Test Paper One",
			URL:      "https://huggingface.co/papers/2509.06652",
			Abstract: "This is a test abstract",
			PubDate:  time.Now().UTC(),
		},
		{
			Title:    "Test Paper Two",
			URL:      "https://huggingface.co/papers/2509.10441",
			Abstract: "This is another test abstract",
			PubDate:  time.Now().UTC(),
		},
	}

	requestURL := "https://tldr.takara.ai/api/papers"
	rssData, err := GeneratePapersRSS(papers, requestURL)
	if err != nil {
		t.Fatalf("GeneratePapersRSS failed: %v", err)
	}

	rssString := string(rssData)

	// Check that TLDR links are present
	expectedLinks := []string{
		"https://tldr.takara.ai/p/2509.06652",
		"https://tldr.takara.ai/p/2509.10441",
	}

	for _, expectedLink := range expectedLinks {
		if !strings.Contains(rssString, expectedLink) {
			t.Errorf("Expected TLDR link %s not found in RSS output", expectedLink)
		}
	}

	// Check that original HuggingFace links are NOT present
	unexpectedLinks := []string{
		"https://huggingface.co/papers/2509.06652",
		"https://huggingface.co/papers/2509.10441",
	}

	for _, unexpectedLink := range unexpectedLinks {
		if strings.Contains(rssString, unexpectedLink) {
			t.Errorf("Unexpected HuggingFace link %s found in RSS output", unexpectedLink)
		}
	}

	// Check that titles are present
	if !strings.Contains(rssString, "Test Paper One") {
		t.Error("Expected title 'Test Paper One' not found in RSS output")
	}
	if !strings.Contains(rssString, "Test Paper Two") {
		t.Error("Expected title 'Test Paper Two' not found in RSS output")
	}
}

func TestIsBlobCacheDisabled(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "Environment variable not set",
			envValue: "",
			expected: false,
		},
		{
			name:     "Environment variable set to true",
			envValue: "true",
			expected: true,
		},
		{
			name:     "Environment variable set to TRUE",
			envValue: "TRUE",
			expected: true,
		},
		{
			name:     "Environment variable set to false",
			envValue: "false",
			expected: false,
		},
		{
			name:     "Environment variable set to invalid value",
			envValue: "invalid",
			expected: false,
		},
		{
			name:     "Environment variable set to 1",
			envValue: "1",
			expected: true,
		},
		{
			name:     "Environment variable set to 0",
			envValue: "0",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original value
			originalValue := os.Getenv("DISABLE_BLOB_CACHE")

			// Set test value
			if tt.envValue == "" {
				_ = os.Unsetenv("DISABLE_BLOB_CACHE")
			} else {
				_ = os.Setenv("DISABLE_BLOB_CACHE", tt.envValue)
			}

			// Test function
			result := isBlobCacheDisabled()
			if result != tt.expected {
				t.Errorf("isBlobCacheDisabled() = %v; expected %v", result, tt.expected)
			}

			// Restore original value
			if originalValue == "" {
				_ = os.Unsetenv("DISABLE_BLOB_CACHE")
			} else {
				_ = os.Setenv("DISABLE_BLOB_CACHE", originalValue)
			}
		})
	}
}
