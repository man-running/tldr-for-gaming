package paper

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

// Regex patterns for parsing ArXiv Atom XML feed.
var (
	axEntryRegex     = regexp.MustCompile(`(?s)<entry>(.*?)</entry>`)
	axTitleRegex     = regexp.MustCompile(`<title>([^<]+)</title>`)
	axSummaryRegex   = regexp.MustCompile(`<summary>([^<]+)</summary>`)
	axAuthorRegex    = regexp.MustCompile(`<name>([^<]+)</name>`)
	axPublishedRegex = regexp.MustCompile(`<published>([^<]+)</published>`)
	axCategoryRegex  = regexp.MustCompile(`<category term="([^"]+)"`)
)

// parseArxivXml extracts paper data from an ArXiv Atom XML string.
func parseArxivXml(xml, arxivId string) (*ArxivApiResponse, error) {
	resp := &ArxivApiResponse{
		PdfURL: fmt.Sprintf("https://arxiv.org/pdf/%s.pdf", arxivId),
	}

	entryMatch := axEntryRegex.FindStringSubmatch(xml)
	if len(entryMatch) < 2 {
		return resp, fmt.Errorf("no <entry> tag found in ArXiv XML")
	}
	entryContent := entryMatch[1]

	titleMatch := axTitleRegex.FindStringSubmatch(entryContent)
	if len(titleMatch) > 1 {
		resp.Title = strings.TrimSpace(titleMatch[1])
	}

	summaryMatch := axSummaryRegex.FindStringSubmatch(entryContent)
	if len(summaryMatch) > 1 {
		resp.Abstract = strings.TrimSpace(summaryMatch[1])
	}

	authorMatches := axAuthorRegex.FindAllStringSubmatch(entryContent, -1)
	authors := []struct{ Name string }{}
	for _, match := range authorMatches {
		if len(match) > 1 {
			authors = append(authors, struct{ Name string }{Name: strings.TrimSpace(match[1])})
		}
	}
	resp.Authors = authors

	publishedMatch := axPublishedRegex.FindStringSubmatch(entryContent)
	if len(publishedMatch) > 1 {
		resp.PublishedDate = strings.TrimSpace(publishedMatch[1])
	}

	categoryMatches := axCategoryRegex.FindAllStringSubmatch(entryContent, -1)
	categories := []string{}
	for _, match := range categoryMatches {
		if len(match) > 1 {
			categories = append(categories, strings.TrimSpace(match[1]))
		}
	}
	resp.Categories = categories

	return resp, nil
}

// FetchArxivData fetches and parses paper data from the ArXiv API.
func FetchArxivData(arxivId string, client *http.Client) (*ArxivApiResponse, error) {
	apiURL := fmt.Sprintf("https://export.arxiv.org/api/query?id_list=%s", arxivId)
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Accept", "application/atom+xml")
	req.Header.Set("User-Agent", "Takara-TLDR/1.0 (Go Port)")

	apiResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("arxiv api request failed: %w", err)
	}
	defer func() { _ = apiResp.Body.Close() }()

	if apiResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("arxiv api returned status %s", apiResp.Status)
	}

	xmlBytes, err := io.ReadAll(apiResp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read arxiv xml body: %w", err)
	}

	return parseArxivXml(string(xmlBytes), arxivId)
}
