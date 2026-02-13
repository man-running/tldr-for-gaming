package subscribe

// RequestBody defines the structure for the incoming subscription request.
type RequestBody struct {
	Email          string `json:"email"`
	TurnstileToken string `json:"turnstileToken"`
}

// TurnstileResponse defines the structure of the JSON response from Cloudflare's siteverify endpoint.
type TurnstileResponse struct {
	Success     bool     `json:"success"`
	ErrorCodes  []string `json:"error-codes"`
	ChallengeTS string   `json:"challenge_ts"`
	Hostname    string   `json:"hostname"`
}

// FeedItem corresponds to a single item in an RSS feed.
type FeedItem struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Description string `json:"description"`
	PubDate     string `json:"pubDate"`
	GUID        string `json:"guid"`
}

// RssFeed corresponds to the overall RSS feed structure.
type RssFeed struct {
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Link          string     `json:"link"`
	LastBuildDate string     `json:"lastBuildDate,omitempty"`
	Items         []FeedItem `json:"items"`
}

// ApiResponse defines a generic success/error response structure.
type ApiResponse struct {
	Success bool   `json:"success,omitempty"`
	Error   string `json:"error,omitempty"`
}
