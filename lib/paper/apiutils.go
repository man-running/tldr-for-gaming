package paper

import (
	"fmt"
	"regexp"
	"strings"
)

var arxivPattern = regexp.MustCompile(`^\d{4}\.\d{4,5}$`)

// ValidateArxivId checks if the string matches the expected ArXiv ID format.
func ValidateArxivId(arxivId string) bool {
	return arxivPattern.MatchString(arxivId)
}

// TransformHfResponse converts the raw HuggingFace API response to the common PaperData format.
func TransformHfResponse(hfData *HuggingFaceApiResponse, arxivId string) *PaperData {
	if hfData == nil {
		return &PaperData{}
	}
	authors := make([]string, len(hfData.Authors))
	for i, author := range hfData.Authors {
		authors[i] = author.Name
	}
	return &PaperData{
		Title:          hfData.Title,
		Abstract:       hfData.Abstract,
		Authors:        authors,
		PublishedDate:  hfData.PublishedDate,
		ArxivID:        arxivId,
		PdfURL:         fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", arxivId),
		Upvotes:        hfData.Upvotes,
		GithubURL:      hfData.GithubURL,
		HuggingfaceURL: fmt.Sprintf("https://huggingface.co/papers/%s", arxivId),
	}
}

// TransformArxivResponse converts the raw ArXiv API response to the common PaperData format.
func TransformArxivResponse(arxivData *ArxivApiResponse, arxivId string) *PaperData {
	if arxivData == nil {
		return &PaperData{}
	}
	authors := make([]string, len(arxivData.Authors))
	for i, author := range arxivData.Authors {
		authors[i] = author.Name
	}
	return &PaperData{
		Title:         arxivData.Title,
		Abstract:      arxivData.Abstract,
		Authors:       authors,
		PublishedDate: arxivData.PublishedDate,
		ArxivID:       arxivId,
		PdfURL:        fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", arxivId),
		Categories:    arxivData.Categories,
		ArxivURL:      fmt.Sprintf("https://arxiv.org/abs/%s", arxivId),
	}
}

// MergePaperData combines data from HuggingFace and ArXiv sources with preference.
func MergePaperData(hfData, arxivData *PaperData) *PaperData {
	merged := &PaperData{}

	// Prefer HF title, fallback to ArXiv
	merged.Title = hfData.Title
	if merged.Title == "" {
		merged.Title = arxivData.Title
	}

	// Prefer HF abstract, fallback to ArXiv
	merged.Abstract = hfData.Abstract
	if merged.Abstract == "" {
		merged.Abstract = arxivData.Abstract
	}

	// Combine authors and remove duplicates
	authorSet := make(map[string]struct{})
	for _, author := range hfData.Authors {
		authorSet[author] = struct{}{}
	}
	for _, author := range arxivData.Authors {
		authorSet[author] = struct{}{}
	}
	for author := range authorSet {
		merged.Authors = append(merged.Authors, author)
	}

	// ArXiv ID should be consistent
	merged.ArxivID = hfData.ArxivID
	if merged.ArxivID == "" {
		merged.ArxivID = arxivData.ArxivID
	}

	// Prefer ArXiv PDF URL, fallback to HF
	merged.PdfURL = arxivData.PdfURL
	if merged.PdfURL == "" {
		merged.PdfURL = hfData.PdfURL
	}

	// Prefer HF publish date, fallback to ArXiv
	merged.PublishedDate = hfData.PublishedDate
	if merged.PublishedDate == "" {
		merged.PublishedDate = arxivData.PublishedDate
	}

	// Add optional fields from their respective sources
	merged.Upvotes = hfData.Upvotes
	merged.GithubURL = hfData.GithubURL
	merged.HuggingfaceURL = hfData.HuggingfaceURL
	merged.ArxivURL = arxivData.ArxivURL
	merged.Categories = arxivData.Categories

	return merged
}

// SanitizePaperData trims whitespace and cleans up the final merged data.
func SanitizePaperData(data *PaperData) *PaperData {
	sanitized := &PaperData{
		Title:    strings.TrimSpace(data.Title),
		Abstract: strings.TrimSpace(data.Abstract),
		ArxivID:  strings.TrimSpace(data.ArxivID),
		PdfURL:   strings.TrimSpace(data.PdfURL),
	}

	seenAuthors := make(map[string]bool)
	for _, author := range data.Authors {
		trimmed := strings.TrimSpace(author)
		if trimmed != "" && !seenAuthors[trimmed] {
			sanitized.Authors = append(sanitized.Authors, trimmed)
			seenAuthors[trimmed] = true
		}
	}

	sanitized.PublishedDate = data.PublishedDate
	sanitized.Upvotes = data.Upvotes
	sanitized.GithubURL = data.GithubURL
	sanitized.HuggingfaceURL = data.HuggingfaceURL
	sanitized.ArxivURL = data.ArxivURL
	sanitized.Categories = data.Categories

	return sanitized
}

// GetDataSourceInfo determines the source of the data based on which fields are present.
func GetDataSourceInfo(hfData, arxivData *PaperData) *DataSourceInfo {
	hasHfData := hfData.Title != "" || hfData.Abstract != "" || len(hfData.Authors) > 0
	hasArxivData := arxivData.Title != "" || arxivData.Abstract != "" || len(arxivData.Authors) > 0

	source := "combined"
	if hasHfData && !hasArxivData {
		source = "huggingface"
	}
	if !hasHfData && hasArxivData {
		source = "arxiv"
	}

	return &DataSourceInfo{
		Source:       source,
		HasHfData:    hasHfData,
		HasArxivData: hasArxivData,
	}
}
