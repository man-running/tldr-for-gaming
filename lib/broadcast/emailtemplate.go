package broadcast

import (
	"bytes"
	"fmt"
	"html/template"
	"time"
)

// The main template string. Note the `range .Items` which will iterate over a slice of structs
// where the `Description` field is of type template.HTML to prevent escaping.
const dailyEmailTemplateStr = `
<!DOCTYPE html>
<html>
<head>
    <style>
        a { color: rgb(217, 16, 9) !important; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body style="font-family: 'Lato', 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <table width="100%" border="0" cellspacing="0" cellpadding="0">
        <tr>
            <td align="center" style="padding-top: 40px; padding-bottom: 20px;">
                <a href="https://tldr.takara.ai" style="text-decoration: none;">
                    <span style="font-family: 'Lato', sans-serif; font-weight: 900; font-size: 40px; color: rgb(74, 77, 78); display: inline;">tldr.</span><span style="font-family: 'Lato', sans-serif; font-weight: 900; font-size: 40px; color: rgb(217, 16, 9); display: inline;">takara.ai</span>
                </a>
            </td>
        </tr>
    </table>

    <table width="100%" border="0" cellspacing="0" cellpadding="0" style="padding: 20px;">
        <tr>
            <td>
                <h1 style="font-family: 'Noto Sans', Helvetica, Arial, sans-serif; font-weight: bold; font-size: 80px; color: rgb(74, 77, 78); margin-top: 10px; margin-bottom: 10px;">
                    TLDR: {{ .FormattedDate }}
                </h1>
                <p style="font-family: 'Lato', sans-serif; font-weight: normal; font-size: 23px; color: rgb(74, 77, 78);">
                    {{ .Feed.Description }}
                </p>
                <hr style="margin-top: 16px; margin-bottom: 16px; border: 0; border-top: 1px solid rgba(74, 77, 78, 0.4);" />
                
                {{range .Items}}
                <div style="margin-bottom: 24px; font-family: 'Lato', sans-serif; font-size: 23px; font-weight: normal; color: rgb(74, 77, 78);">
                    {{.Description}}
                </div>
                {{end}}
            </td>
        </tr>
    </table>

    <table width="100%" border="0" cellspacing="0" cellpadding="0" style="padding: 12px 20px; max-width: 100%;">
        <tr>
            <td>
                <hr style="margin-top: 16px; margin-bottom: 16px; border: 0; border-top: 1px solid rgba(74, 77, 78, 0.4);" />
                <table width="100%" border="0" cellspacing="0" cellpadding="0">
                    <tr>
                        <td valign="top">
                            <p style="font-family: 'Lato', sans-serif; line-height: 100%; font-weight: bolder; font-size: 50px; color: rgb(74, 77, 78); margin: 0;">Transforming Humanity</p>
                            <p style="font-family: 'Noto Sans', sans-serif; line-height: 200%; font-weight: bold; font-size: 25px; color: rgb(217, 16, 9); margin: 0;">類を変革する</p>
                        </td>
                        <td align="right" valign="top" style="width: 50%;">
                            <a href="https://takara.ai">
                                <img src="https://tldr.takara.ai/icon.svg" alt="Origami Crane Logo" style="width: 250px; height: 194px; max-width: 100%;" />
                            </a>
                        </td>
                    </tr>
                </table>

                <table width="100%" border="0" cellspacing="0" cellpadding="0" style="text-align: center; font-family: 'Lato', sans-serif; font-size: 12px; color: rgba(74, 77, 78, 0.8);">
                    <tr><td style="padding-top: 10px;">Want to stop receiving emails? {{ "<a href=\"{{{RESEND_UNSUBSCRIBE_URL}}}\">Unsubscribe</a>" | safeHTML }}</td></tr>
                    <tr><td>Disclaimer: takara.ai Ltd is not responsible for the content of these articles as all content is from third-party sources.</td></tr>
                    <tr><td>© 2025 takara.ai Ltd. All rights reserved.</td></tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`

// TemplateData holds the data for the email template.
// The Items slice contains structs with a `template.HTML` field to ensure
// the content is not escaped by the template engine.
type TemplateData struct {
	Feed          RssFeed
	FormattedDate string
	Items         []struct{ Description template.HTML }
}

// generateEmailHTML executes the Go template to produce the final email body.
func generateEmailHTML(feed RssFeed) (string, error) {
	// Prepare the data for the template.
	templateData := TemplateData{
		Feed:          feed,
		FormattedDate: formatDate(feed.LastBuildDate),
		Items:         make([]struct{ Description template.HTML }, len(feed.Items)),
	}

	// Convert each item's description from string to template.HTML
	for i, item := range feed.Items {
		templateData.Items[i] = struct{ Description template.HTML }{Description: template.HTML(item.Description)}
	}

	// Parse the template string.
	tpl, err := template.New("dailyEmail").Funcs(template.FuncMap{
		"safeHTML": func(s string) template.HTML { return template.HTML(s) },
	}).Parse(dailyEmailTemplateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse email template: %w", err)
	}

	// Execute the template into a buffer.
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, templateData); err != nil {
		return "", fmt.Errorf("failed to execute email template: %w", err)
	}

	return buf.String(), nil
}

// formatDate converts a date string from the RSS feed into a more readable format.
func formatDate(dateStr string) string {
	if dateStr == "" {
		return "No date available"
	}
	// Attempt to parse the date using common RSS feed time formats.
	layouts := []string{time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822, time.RubyDate}
	var t time.Time
	var err error
	for _, layout := range layouts {
		t, err = time.Parse(layout, dateStr)
		if err == nil {
			break
		}
	}

	if err != nil {
		// If parsing fails with all layouts, return the original string.
		return dateStr
	}

	// Format to "Month Day, Year" e.g., "January 2, 2006"
	return t.Format("January 2, 2006")
}
