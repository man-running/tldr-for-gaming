package paper

// Corresponds to ErrorCode enum
type ErrorCode string

const (
	ErrorCodeInvalidArxivID   ErrorCode = "INVALID_ARXIV_ID"
	ErrorCodePaperNotFound    ErrorCode = "PAPER_NOT_FOUND"
	ErrorCodeExternalAPIError ErrorCode = "EXTERNAL_API_ERROR"
	ErrorCodeRateLimited      ErrorCode = "RATE_LIMITED"
	ErrorCodeInternalError    ErrorCode = "INTERNAL_ERROR"
	ErrorCodeValidationError  ErrorCode = "VALIDATION_ERROR"
)

// ApiError defines the structure for an error response.
type ApiError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// PaperData is the combined paper data structure.
type PaperData struct {
	Title          string   `json:"title"`
	Abstract       string   `json:"abstract"`
	Authors        []string `json:"authors"`
	PublishedDate  string   `json:"publishedDate,omitempty"`
	ArxivID        string   `json:"arxivId"`
	PdfURL         string   `json:"pdfUrl"`
	Upvotes        int      `json:"upvotes,omitempty"`
	GithubURL      string   `json:"githubUrl,omitempty"`
	HuggingfaceURL string   `json:"huggingfaceUrl,omitempty"`
	ArxivURL       string   `json:"arxivUrl,omitempty"`
	Categories     []string `json:"categories,omitempty"`
}

type PaperMetadata struct {
	Title         string   `json:"title"`
	Authors       []string `json:"authors"`
	PublishedDate string   `json:"publishedDate,omitempty"`
	ArxivID       string   `json:"arxivId"`
	CachedAt      string   `json:"cachedAt"`
}

// HuggingFaceApiResponse is the structure of the data from the HF API/scraper.
type HuggingFaceApiResponse struct {
	Title         string                  `json:"title,omitempty"`
	Abstract      string                  `json:"abstract,omitempty"`
	Authors       []struct{ Name string } `json:"authors,omitempty"`
	PublishedDate string                  `json:"publishedDate,omitempty"`
	ArxivID       string                  `json:"arxivId,omitempty"`
	GithubURL     string                  `json:"githubUrl,omitempty"`
	PdfURL        string                  `json:"pdfUrl,omitempty"`
	Upvotes       int                     `json:"upvotes,omitempty"`
}

// ArxivApiResponse is the structure of the data from the ArXiv XML parser.
type ArxivApiResponse struct {
	Title         string                  `json:"title,omitempty"`
	Abstract      string                  `json:"abstract,omitempty"`
	Authors       []struct{ Name string } `json:"authors,omitempty"`
	PublishedDate string                  `json:"publishedDate,omitempty"`
	Categories    []string                `json:"categories,omitempty"`
	PdfURL        string                  `json:"pdfUrl,omitempty"`
}

// DataSourceInfo tracks where the data came from.
type DataSourceInfo struct {
	Source       string
	HasHfData    bool
	HasArxivData bool
}

// FinalApiResponse is the top-level structure for the JSON response.
type FinalApiResponse struct {
	Success bool       `json:"success"`
	Data    *PaperData `json:"data,omitempty"`
	BlobURL *string    `json:"blobURL,omitempty"` // Optional: URL for client to fetch directly
	Error   *ApiError  `json:"error,omitempty"`
}
