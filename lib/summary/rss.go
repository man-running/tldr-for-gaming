package summary

import (
	"encoding/xml"
	"fmt"
	"time"
)

// RSS structures for generating RSS feeds
type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Version string   `xml:"version,attr"`
	Channel Channel  `xml:"channel"`
	XMLNS   string   `xml:"xmlns:atom,attr"`
}

type Channel struct {
	Title         string   `xml:"title"`
	Link          string   `xml:"link"`
	Description   string   `xml:"description"`
	LastBuildDate string   `xml:"lastBuildDate"`
	AtomLink      AtomLink `xml:"atom:link"`
	Items         []Item   `xml:"item"`
}

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description CDATA  `xml:"description"`
	PubDate     string `xml:"pubDate"`
	GUID        GUID   `xml:"guid"`
}

type GUID struct {
	IsPermaLink bool   `xml:"isPermaLink,attr"`
	Text        string `xml:",chardata"`
}

// CDATA represents CDATA-wrapped content in XML
type CDATA struct {
	Text string `xml:",cdata"`
}

// GeneratePapersRSS generates an RSS feed from a list of papers
func GeneratePapersRSS(papers []Paper, requestURL string) ([]byte, error) {
	items := make([]Item, len(papers))
	for i, paper := range papers {
		tldrLink := toTLDRLink(paper.URL)
		items[i] = Item{
			Title:       paper.Title,
			Link:        tldrLink,
			Description: CDATA{Text: paper.Abstract},
			PubDate:     paper.PubDate.Format(time.RFC1123Z),
			GUID: GUID{
				IsPermaLink: true,
				Text:        tldrLink,
			},
		}
	}

	rss := RSS{
		Version: "2.0",
		XMLNS:   "http://www.w3.org/2005/Atom",
		Channel: Channel{
			Title:         "Takara TLDR - Daily AI Papers",
			Link:          "https://tldr.takara.ai",
			Description:   "Daily AI research papers from Takara.ai",
			LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
			AtomLink: AtomLink{
				Href: requestURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: items,
		},
	}

	// Add XML header and proper encoding
	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return nil, err
	}

	// Prepend the XML header
	return append([]byte(xml.Header), output...), nil
}

// GenerateSummaryRSS generates an RSS feed containing a summary
func GenerateSummaryRSS(summaryHTML, requestURL string, date time.Time) ([]byte, error) {
	wrappedHtmlSummary := fmt.Sprintf("<div>%s</div>", summaryHTML)

	item := Item{
		Title:       "AI Research Papers Summary for " + date.Format("January 2, 2006"),
		Link:        "https://tldr.takara.ai",
		Description: CDATA{Text: wrappedHtmlSummary},
		PubDate:     date.Format(time.RFC1123Z),
		GUID: GUID{
			IsPermaLink: false,
			Text:        fmt.Sprintf("summary-%s", date.Format("2006-01-02")),
		},
	}

	rss := RSS{
		Version: "2.0",
		XMLNS:   "http://www.w3.org/2005/Atom",
		Channel: Channel{
			Title:         "Takara TLDR",
			Link:          "https://tldr.takara.ai",
			Description:   "Daily summaries of AI research papers from takara.ai",
			LastBuildDate: date.Format(time.RFC1123Z),
			AtomLink: AtomLink{
				Href: requestURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: []Item{item},
		},
	}

	// Add XML header and proper encoding
	output, err := xml.MarshalIndent(rss, "", "  ")
	if err != nil {
		return nil, err
	}

	// Prepend the XML header
	return append([]byte(xml.Header), output...), nil
}
