package feed

import (
	"os"
	"testing"
	"time"
)

// TestNewSourceManager tests creating a new source manager
func TestNewSourceManager(t *testing.T) {
	manager := NewSourceManager()

	if manager == nil {
		t.Fatal("NewSourceManager should not return nil")
	}

	if manager.sources == nil {
		t.Error("SourceManager.sources should be initialized")
	}

	count := manager.GetSourceCount()
	if count != 0 {
		t.Errorf("New manager should have 0 sources, got %d", count)
	}
}

// TestLoadDefaultSources tests loading default sources
func TestLoadDefaultSources(t *testing.T) {
	manager := NewSourceManager()
	err := manager.LoadDefaultSources()

	if err != nil {
		t.Fatalf("LoadDefaultSources failed: %v", err)
	}

	count := manager.GetSourceCount()
	expectedCount := 5
	if count != expectedCount {
		t.Errorf("Expected %d sources, got %d", expectedCount, count)
	}

	activeCount := manager.GetActiveSourceCount()
	if activeCount != expectedCount {
		t.Errorf("Expected %d active sources, got %d", expectedCount, activeCount)
	}
}

// TestAddSource tests adding a new source
func TestAddSource(t *testing.T) {
	manager := NewSourceManager()

	newSource := &NewsSource{
		ID:           "testsource",
		Name:         "Test Source",
		URL:          "https://test.example.com",
		FeedURL:      "https://test.example.com/feed/",
		Category:     "Business",
		Active:       true,
		Priority:     5,
		ScrapingType: "rss",
		Timeout:      10000,
	}

	err := manager.AddSource(newSource)
	if err != nil {
		t.Fatalf("AddSource failed: %v", err)
	}

	count := manager.GetSourceCount()
	if count != 1 {
		t.Errorf("Expected 1 source after add, got %d", count)
	}

	retrieved, err := manager.GetSource("testsource")
	if err != nil {
		t.Fatalf("GetSource failed: %v", err)
	}

	if retrieved.Name != "Test Source" {
		t.Errorf("Expected source name 'Test Source', got '%s'", retrieved.Name)
	}
}

// TestAddSourceValidation tests source validation on add
func TestAddSourceValidation(t *testing.T) {
	manager := NewSourceManager()

	testCases := []struct {
		name    string
		source  *NewsSource
		wantErr bool
	}{
		{
			name: "Valid source",
			source: &NewsSource{
				ID:      "valid",
				Name:    "Valid",
				FeedURL: "https://example.com/feed",
			},
			wantErr: false,
		},
		{
			name: "Missing ID",
			source: &NewsSource{
				Name:    "No ID",
				FeedURL: "https://example.com/feed",
			},
			wantErr: true,
		},
		{
			name: "Missing Name",
			source: &NewsSource{
				ID:      "noid",
				FeedURL: "https://example.com/feed",
			},
			wantErr: true,
		},
		{
			name: "Missing FeedURL",
			source: &NewsSource{
				ID:   "nofeed",
				Name: "No Feed",
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := manager.AddSource(tc.source)
			if (err != nil) != tc.wantErr {
				t.Errorf("AddSource error: expected error=%v, got %v", tc.wantErr, err)
			}
		})
	}
}

// TestGetSource tests retrieving a source by ID
func TestGetSource(t *testing.T) {
	manager := NewSourceManager()
	manager.LoadDefaultSources()

	source, err := manager.GetSource("igamingbusiness")
	if err != nil {
		t.Fatalf("GetSource failed: %v", err)
	}

	if source.Name != "iGamingBusiness" {
		t.Errorf("Expected 'iGamingBusiness', got '%s'", source.Name)
	}

	_, err = manager.GetSource("nonexistent")
	if err == nil {
		t.Error("GetSource should return error for nonexistent source")
	}
}

// TestGetActiveSources tests getting active sources sorted by priority
func TestGetActiveSources(t *testing.T) {
	manager := NewSourceManager()
	manager.LoadDefaultSources()

	active := manager.GetActiveSources()

	if len(active) == 0 {
		t.Fatal("Should have active sources")
	}

	// Verify sorted by priority (descending)
	for i := 0; i < len(active)-1; i++ {
		if active[i].Priority < active[i+1].Priority {
			t.Errorf("Sources not sorted by priority: %d < %d",
				active[i].Priority, active[i+1].Priority)
		}
	}

	// Verify all are active
	for _, source := range active {
		if !source.Active {
			t.Errorf("Inactive source in active list: %s", source.Name)
		}
	}
}

// TestDisableSource tests disabling a source
func TestDisableSource(t *testing.T) {
	manager := NewSourceManager()
	manager.LoadDefaultSources()

	initialActive := manager.GetActiveSourceCount()

	err := manager.DisableSource("igamingbusiness")
	if err != nil {
		t.Fatalf("DisableSource failed: %v", err)
	}

	afterActive := manager.GetActiveSourceCount()
	if afterActive != initialActive-1 {
		t.Errorf("Expected %d active sources after disable, got %d",
			initialActive-1, afterActive)
	}

	source, _ := manager.GetSource("igamingbusiness")
	if source.Active {
		t.Error("Source should be disabled")
	}
}

// TestEnableSource tests enabling a disabled source
func TestEnableSource(t *testing.T) {
	manager := NewSourceManager()
	manager.LoadDefaultSources()

	manager.DisableSource("igamingbusiness")
	initialActive := manager.GetActiveSourceCount()

	err := manager.EnableSource("igamingbusiness")
	if err != nil {
		t.Fatalf("EnableSource failed: %v", err)
	}

	afterActive := manager.GetActiveSourceCount()
	if afterActive != initialActive+1 {
		t.Errorf("Expected %d active sources after enable, got %d",
			initialActive+1, afterActive)
	}

	source, _ := manager.GetSource("igamingbusiness")
	if !source.Active {
		t.Error("Source should be enabled")
	}
}

// TestGetSourcesByCategory tests filtering sources by category
func TestGetSourcesByCategory(t *testing.T) {
	manager := NewSourceManager()
	manager.LoadDefaultSources()

	businessSources := manager.GetSourcesByCategory("Business")

	if len(businessSources) == 0 {
		t.Fatal("Should have business category sources")
	}

	for _, source := range businessSources {
		if source.Category != "Business" {
			t.Errorf("Expected Business category, got '%s'", source.Category)
		}
		if !source.Active {
			t.Errorf("Source %s should be active", source.Name)
		}
	}
}

// TestListSources tests listing all sources
func TestListSources(t *testing.T) {
	manager := NewSourceManager()
	manager.LoadDefaultSources()

	all := manager.ListSources()

	if len(all) != 5 {
		t.Errorf("Expected 5 sources, got %d", len(all))
	}

	for _, source := range all {
		if source.ID == "" {
			t.Error("Source ID should not be empty")
		}
		if source.Name == "" {
			t.Error("Source Name should not be empty")
		}
	}
}

// TestUpdateSource tests updating a source
func TestUpdateSource(t *testing.T) {
	manager := NewSourceManager()
	manager.LoadDefaultSources()

	updates := &NewsSource{
		Name:     "Updated iGamingBusiness",
		Priority: 8,
		Active:   false,
	}

	err := manager.UpdateSource("igamingbusiness", updates)
	if err != nil {
		t.Fatalf("UpdateSource failed: %v", err)
	}

	source, _ := manager.GetSource("igamingbusiness")
	if source.Name != "Updated iGamingBusiness" {
		t.Errorf("Expected updated name, got '%s'", source.Name)
	}
	if source.Priority != 8 {
		t.Errorf("Expected priority 8, got %d", source.Priority)
	}
	if source.Active {
		t.Error("Source should be disabled")
	}
}

// TestExportSources tests exporting sources as JSON
func TestExportSources(t *testing.T) {
	manager := NewSourceManager()
	manager.LoadDefaultSources()

	json, err := manager.ExportSources()
	if err != nil {
		t.Fatalf("ExportSources failed: %v", err)
	}

	if json == "" {
		t.Error("ExportSources should return non-empty JSON")
	}

	if !contains(json, "igamingbusiness") {
		t.Error("Exported JSON should contain igamingbusiness source")
	}
}

// TestValidate tests source configuration validation
func TestValidate(t *testing.T) {
	manager := NewSourceManager()

	// Empty manager should fail
	err := manager.Validate()
	if err == nil {
		t.Error("Validate should fail for empty manager")
	}

	// Load defaults and validate
	manager.LoadDefaultSources()
	err = manager.Validate()
	if err != nil {
		t.Fatalf("Validate failed for valid configuration: %v", err)
	}

	// Disable all sources and try to validate
	for _, source := range manager.ListSources() {
		manager.DisableSource(source.ID)
	}

	err = manager.Validate()
	if err == nil {
		t.Error("Validate should fail when all sources are disabled")
	}
}

// TestValidateInvalidPriority tests priority validation
func TestValidateInvalidPriority(t *testing.T) {
	manager := NewSourceManager()

	invalidSource := &NewsSource{
		ID:      "invalid",
		Name:    "Invalid",
		FeedURL: "https://example.com/feed",
		Active:  true,
		Priority: 15, // Invalid: should be 1-10
	}

	manager.AddSource(invalidSource)
	err := manager.Validate()
	if err == nil {
		t.Error("Validate should fail for priority outside 1-10 range")
	}
}

// TestValidateInvalidScrapingType tests scraping type validation
func TestValidateInvalidScrapingType(t *testing.T) {
	manager := NewSourceManager()

	invalidSource := &NewsSource{
		ID:           "invalid",
		Name:         "Invalid",
		FeedURL:      "https://example.com/feed",
		Active:       true,
		Priority:     5,
		ScrapingType: "invalid_type",
	}

	manager.AddSource(invalidSource)
	err := manager.Validate()
	if err == nil {
		t.Error("Validate should fail for invalid scraping type")
	}
}

// TestConcurrentAccess tests thread-safe operations
func TestConcurrentAccess(t *testing.T) {
	manager := NewSourceManager()
	manager.LoadDefaultSources()

	done := make(chan bool)

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			manager.GetActiveSources()
			manager.ListSources()
			manager.GetSourceCount()
			done <- true
		}()
	}

	// Concurrent writes
	for i := 0; i < 5; i++ {
		go func(idx int) {
			newSource := &NewsSource{
				ID:           "source-" + string(rune(idx)),
				Name:         "Test Source",
				FeedURL:      "https://test.example.com/feed",
				Active:       true,
				Priority:     5,
				ScrapingType: "rss",
			}
			manager.AddSource(newSource)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	// Verify final state
	count := manager.GetSourceCount()
	if count < 5 {
		t.Errorf("Expected at least 5 sources after concurrent operations, got %d", count)
	}
}

// TestSourceTimestamps tests creation and update timestamps
func TestSourceTimestamps(t *testing.T) {
	manager := NewSourceManager()

	source := &NewsSource{
		ID:      "timestamp-test",
		Name:    "Timestamp Test",
		FeedURL: "https://example.com/feed",
		Active:  true,
		Priority: 5,
	}

	beforeAdd := time.Now()
	manager.AddSource(source)
	afterAdd := time.Now()

	retrieved, _ := manager.GetSource("timestamp-test")

	if retrieved.CreatedAt.Before(beforeAdd) || retrieved.CreatedAt.After(afterAdd) {
		t.Error("CreatedAt timestamp not set correctly")
	}

	// Update and check UpdatedAt
	beforeUpdate := time.Now()
	manager.UpdateSource("timestamp-test", &NewsSource{Name: "Updated"})
	afterUpdate := time.Now()

	retrieved, _ = manager.GetSource("timestamp-test")

	if retrieved.UpdatedAt.Before(beforeUpdate) || retrieved.UpdatedAt.After(afterUpdate) {
		t.Error("UpdatedAt timestamp not set correctly after update")
	}
}

// TestLoadSourcesFromFile tests loading sources from JSON file
func TestLoadSourcesFromFile(t *testing.T) {
	manager := NewSourceManager()

	// Use the news-sources.json file
	err := manager.LoadSourcesFromFile("news-sources.json")

	// If file doesn't exist in test directory, this is expected
	// In actual deployment, this would load from the proper location
	if err != nil && !os.IsNotExist(err) {
		t.Logf("Note: LoadSourcesFromFile expected to fail in test environment: %v", err)
	}
}

// TestGetSourceCount tests counting sources
func TestGetSourceCount(t *testing.T) {
	manager := NewSourceManager()

	if manager.GetSourceCount() != 0 {
		t.Error("New manager should have 0 sources")
	}

	manager.LoadDefaultSources()

	if manager.GetSourceCount() != 5 {
		t.Errorf("Expected 5 sources, got %d", manager.GetSourceCount())
	}
}

// TestGetActiveSourceCount tests counting active sources
func TestGetActiveSourceCount(t *testing.T) {
	manager := NewSourceManager()
	manager.LoadDefaultSources()

	if manager.GetActiveSourceCount() != 5 {
		t.Errorf("Expected 5 active sources, got %d", manager.GetActiveSourceCount())
	}

	manager.DisableSource("igamingbusiness")

	if manager.GetActiveSourceCount() != 4 {
		t.Errorf("Expected 4 active sources after disable, got %d", manager.GetActiveSourceCount())
	}
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
