package broadcast

import (
	"encoding/json"
	"fmt"
)

// GUIDString can unmarshal from a JSON string or an object (various shapes).
type GUIDString string

func (g *GUIDString) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		*g = ""
		return nil
	}
	switch data[0] {
	case '"':
		var s string
		if err := json.Unmarshal(data, &s); err != nil {
			return err
		}
		*g = GUIDString(s)
		return nil
	case '{':
		var m map[string]interface{}
		if err := json.Unmarshal(data, &m); err != nil {
			return err
		}
		for _, k := range []string{"guid", "value", "_", "#text", "text", "content", "id"} {
			if v, ok := m[k]; ok {
				if sv, ok := v.(string); ok {
					*g = GUIDString(sv)
					return nil
				}
			}
		}
		*g = GUIDString(string(data))
		return nil
	default:
		var any interface{}
		if err := json.Unmarshal(data, &any); err == nil {
			*g = GUIDString(fmt.Sprint(any))
			return nil
		}
		*g = ""
		return nil
	}
}

// FeedItem corresponds to the TypeScript FeedItemType.
type FeedItem struct {
	Title       string     `json:"title"`
	Link        string     `json:"link"`
	Description string     `json:"description"`
	PubDate     string     `json:"pubDate"`
	GUID        GUIDString `json:"guid"`
}

// RssFeed corresponds to the TypeScript RssFeed type.
type RssFeed struct {
	Title         string     `json:"title"`
	Description   string     `json:"description"`
	Link          string     `json:"link"`
	LastBuildDate string     `json:"lastBuildDate,omitempty"`
	Items         []FeedItem `json:"items"`
}
