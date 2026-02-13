package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"main/lib/logger"
	"main/lib/summary"
)

func main() {
	// Initialize environment (load .env if available)
	err := godotenv.Load()
	if err != nil {
		logger.Warn("Error loading .env file", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Create summary service
	service := summary.NewService()

	// Base URL for the deployment
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://tldr.takara.ai"
	}

	// Construct URLs
	papersURL := baseURL + "/api/papers"
	tldrURL := baseURL + "/api/tldr" // Now defaults to summary

	logger.Info("Starting summary generation", map[string]interface{}{
		"papersURL": papersURL,
		"tldrURL":   tldrURL,
		"baseURL":   baseURL,
	})

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second) // 5 minutes
	defer cancel()

	// Step 1: Get papers data from papers API
	logger.Info("Step 1: Fetching papers data from API", map[string]interface{}{
		"papersURL": papersURL,
	})
	papersResult, err := service.GetPapersRaw(ctx, papersURL)
	if err != nil {
		logger.Error("Failed to get papers data", err, map[string]interface{}{
			"papersURL": papersURL,
		})
		log.Fatalf("Failed to get papers data: %v", err)
	}
	logger.Info("Papers data fetched successfully", map[string]interface{}{
		"source": papersResult.Source,
		"size":   len(papersResult.Data),
	})

	// Step 2: Generate summary from papers data
	logger.Info("Step 2: Generating summary with OpenAI", map[string]interface{}{
		"papersSize": len(papersResult.Data),
		"tldrURL":    tldrURL,
	})
	summaryData, err := service.GenerateSummaryFromRSS(ctx, papersResult.Data, tldrURL)
	if err != nil {
		logger.Error("Failed to generate summary", err, map[string]interface{}{
			"papersSize": len(papersResult.Data),
			"tldrURL":    tldrURL,
		})
		log.Fatalf("Failed to generate summary: %v", err)
	}
	logger.Info("Summary generated successfully", map[string]interface{}{
		"summarySize": len(summaryData),
	})

	// Step 3: Store summary in blob cache
	logger.Info("Step 3: Storing summary in blob cache", map[string]interface{}{
		"summarySize": len(summaryData),
	})
	err = summary.StoreSummary(summaryData)
	if err != nil {
		logger.Error("Failed to store summary", err, map[string]interface{}{
			"summarySize": len(summaryData),
		})
		log.Fatalf("Failed to store summary: %v", err)
	}
	logger.Info("Summary stored in blob cache successfully", map[string]interface{}{})

	logger.Info("Summary generation completed successfully", map[string]interface{}{
		"summarySize": len(summaryData),
		"tldrURL":     tldrURL,
	})
}
