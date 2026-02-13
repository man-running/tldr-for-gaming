package paper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

// Regex patterns translated from the JavaScript source.
var (
	hfTitleRegex    = regexp.MustCompile(`<h1[^>]*class="[^"]*"[^>]*>([^<]+)</h1>`)
	hfAbstractRegex = regexp.MustCompile(`<h2[^>]*>Abstract</h2>\s*<[^>]*>\s*([^<]+)`)
	hfAuthorRegex   = regexp.MustCompile(`href="/(author|user)/[^"]+"[^>]*>([^<]+)</a>`)
	hfUpvoteRegex   = regexp.MustCompile(`Upvote\s*(\d+)`)
	hfGithubRegex   = regexp.MustCompile(`href="(https://github\.com/[^"]+)"`)
)

// extractDataFromHtml scrapes paper data from a Hugging Face HTML page.
func extractDataFromHtml(html, arxivId string) (*HuggingFaceApiResponse, error) {
	resp := &HuggingFaceApiResponse{
		ArxivID: arxivId,
		PdfURL:  fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", arxivId),
	}

	titleMatch := hfTitleRegex.FindStringSubmatch(html)
	if len(titleMatch) > 1 {
		resp.Title = strings.TrimSpace(titleMatch[1])
	}

	abstractMatch := hfAbstractRegex.FindStringSubmatch(html)
	if len(abstractMatch) > 1 {
		resp.Abstract = strings.TrimSpace(abstractMatch[1])
	}

	authorMatches := hfAuthorRegex.FindAllStringSubmatch(html, -1)
	authors := []struct{ Name string }{}
	for _, match := range authorMatches {
		if len(match) > 2 {
			authors = append(authors, struct{ Name string }{Name: strings.TrimSpace(match[2])})
		}
	}
	resp.Authors = authors

	upvoteMatch := hfUpvoteRegex.FindStringSubmatch(html)
	if len(upvoteMatch) > 1 {
		if upvotes, err := strconv.Atoi(upvoteMatch[1]); err == nil {
			resp.Upvotes = upvotes
		}
	}

	githubMatch := hfGithubRegex.FindStringSubmatch(html)
	if len(githubMatch) > 1 {
		resp.GithubURL = githubMatch[1]
	}

	return resp, nil
}

// FetchHuggingFaceData tries to get paper data from the HF API, with a fallback to HTML scraping.
func FetchHuggingFaceData(arxivId string, client *http.Client) (*HuggingFaceApiResponse, error) {
	// 1. Try the API first
	apiURL := fmt.Sprintf("https://huggingface.co/api/papers/%s", arxivId)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Takara-TLDR/1.0 (Go Port)")

	apiResp, err := client.Do(req)
	if err == nil && apiResp.StatusCode == http.StatusOK {
		defer func() { _ = apiResp.Body.Close() }()
		var hfData HuggingFaceApiResponse
		if err := json.NewDecoder(apiResp.Body).Decode(&hfData); err == nil {
			return &hfData, nil
		}
		// If JSON decoding fails, we'll proceed to the scraper.
	}
	if apiResp != nil {
		_ = apiResp.Body.Close()
	}

	// 2. Fallback to HTML scraping
	scrapeURL := fmt.Sprintf("https://huggingface.co/papers/%s", arxivId)
	req, _ = http.NewRequest("GET", scrapeURL, nil)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Takara-TLDR/1.0; Go Port)")

	scrapeResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("huggingface scraper fetch failed: %w", err)
	}
	defer func() { _ = scrapeResp.Body.Close() }()

	if scrapeResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("huggingface scraper returned status %s", scrapeResp.Status)
	}

	htmlBytes, err := io.ReadAll(scrapeResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read huggingface html body: %w", err)
	}

	return extractDataFromHtml(string(htmlBytes), arxivId)
}
