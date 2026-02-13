package og

// PaperMetadata defines the structure for a paper's metadata.
type PaperMetadata struct {
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	PublishedDate string   `json:"publishedDate,omitempty"`
	ArxivID       string   `json:"arxivId"`
	CachedAt      string   `json:"cachedAt"`
}

// PaperData defines the full structure for a paper, including metadata and other features.
// This is a simplified version of the one in `api/types` for the sole purpose of getting the title.
type PaperData struct {
	Title         string   `json:"title"`
	Abstract      string   `json:"abstract"`
	Authors       []string `json:"authors"`
	PublishedDate string   `json:"publishedDate,omitempty"`
	ArxivID       string   `json:"arxivId"`
	PdfURL        string   `json:"pdfUrl"`
}
