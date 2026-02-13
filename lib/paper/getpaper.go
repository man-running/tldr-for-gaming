package paper

import (
	"context"
	"main/lib/analytics"
	"main/lib/logger"
	"net/http"
	"sync"
	"time"
)

// Custom error types to propagate specific failure modes to the handler.
type PaperNotFoundError struct{ msg string }

func (e *PaperNotFoundError) Error() string { return e.msg }

type InvalidIdError struct{ msg string }

func (e *InvalidIdError) Error() string { return e.msg }

// GetPaperRawResult holds the final data and its source.
type GetPaperRawResult struct {
	Data    *PaperData
	Source  string
	BlobURL *string // Optional: URL if available from blob cache
}

// GetPaperRaw orchestrates the fetching of paper data, including caching and external fallbacks.
func GetPaperRaw(arxivId string) (*GetPaperRawResult, error) {
	if !ValidateArxivId(arxivId) {
		return nil, &InvalidIdError{msg: "Invalid ArXiv ID format"}
	}

	// 1. Check blob cache first - get URL without fetching content
	blobURL, err := GetPaperURL(arxivId)
	if err != nil {
		logCtx := map[string]interface{}{"arxiv_id": arxivId}
		logger.Error("Failed to check blob cache", err, logCtx)
	}
	if blobURL != "" {
		// Blob exists, return URL for client to fetch directly
		_ = analytics.Track("paper_viewed", arxivId, map[string]interface{}{
			"arxiv_id": arxivId,
			"source":   "blob",
		})
		return &GetPaperRawResult{
			Data:    nil, // Client will fetch from blob URL
			Source:   "blob",
			BlobURL:  &blobURL,
		}, nil
	}

	logCtx := map[string]interface{}{"arxiv_id": arxivId}
	logger.Debug("Blob cache miss", logCtx)

	// 2. Fetch from external sources concurrently with a timeout.
	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	httpClient := &http.Client{}

	var hfData, arxivData *PaperData
	var hfErr, arxivErr error
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		rawHfData, err := FetchHuggingFaceData(arxivId, httpClient)
		if err != nil {
			// Check if the error is due to the context deadline being exceeded.
			if timeoutCtx.Err() == context.DeadlineExceeded {
				hfErr = timeoutCtx.Err()
			} else {
				hfErr = err
			}
			return
		}
		hfData = TransformHfResponse(rawHfData, arxivId)
	}()

	go func() {
		defer wg.Done()
		rawArxivData, err := FetchArxivData(arxivId, httpClient)
		if err != nil {
			if timeoutCtx.Err() == context.DeadlineExceeded {
				arxivErr = timeoutCtx.Err()
			} else {
				arxivErr = err
			}
			return
		}
		arxivData = TransformArxivResponse(rawArxivData, arxivId)
	}()

	wg.Wait()

	if hfErr != nil {
		logCtx := map[string]interface{}{"arxiv_id": arxivId}
		logger.Error("HuggingFace fetch failed", hfErr, logCtx)
		hfData = &PaperData{} // Ensure it's not nil for merging
	}
	if arxivErr != nil {
		logCtx := map[string]interface{}{"arxiv_id": arxivId}
		logger.Error("ArXiv fetch failed", arxivErr, logCtx)
		arxivData = &PaperData{} // Ensure it's not nil for merging
	}

	// 3. Merge and sanitize the data
	merged := MergePaperData(hfData, arxivData)
	if merged.Title == "" || merged.Abstract == "" || len(merged.Authors) == 0 {
		return nil, &PaperNotFoundError{msg: "Paper not found from any source"}
	}

	sanitized := SanitizePaperData(merged)

	// 4. Asynchronously store the result in the blob cache (fire-and-forget)
	go func() {
		err := StorePaper(arxivId, sanitized)
		if err != nil {
			logCtx := map[string]interface{}{"arxiv_id": arxivId}
			logger.Error("Failed to store paper in blob cache", err, logCtx)
		}
	}()

	sourceInfo := GetDataSourceInfo(hfData, arxivData)

	_ = analytics.Track("paper_viewed", arxivId, map[string]interface{}{
		"arxiv_id": arxivId,
		"source":   sourceInfo.Source,
	})

	return &GetPaperRawResult{Data: sanitized, Source: sourceInfo.Source}, nil
}
