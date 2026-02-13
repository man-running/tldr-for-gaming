package subscribe

import (
	"bytes"
	"fmt"
	"html/template"
	"time"
)

// The main template string. It uses Go template's `if .Feed` to conditionally render the feed section.
const welcomeEmailTemplateStr = `
<!DOCTYPE html>
<html>
<head>
    <style>
        a { color: rgb(217, 16, 9) !important; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body style="font-family: 'Lato', 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <table width="100%" border="0" cellspacing="0" cellpadding="0" style="padding: 40px 0 20px 0;">
        <tr>
            <td align="center">
                <a href="https://tldr.takara.ai" style="text-decoration: none;">
                    <span style="font-family: 'Lato', sans-serif; font-weight: 900; font-size: 40px; color: rgb(74, 77, 78);">tldr.</span><span style="font-family: 'Lato', sans-serif; font-weight: 900; font-size: 40px; color: rgb(217, 16, 9);">takara.ai</span>
                </a>
            </td>
        </tr>
    </table>

    <table width="100%" border="0" cellspacing="0" cellpadding="0" style="padding: 20px;">
        <tr>
            <td>
                <h1 style="font-family: 'Noto Sans', Helvetica, Arial, sans-serif; font-weight: bold; font-size: 60px; color: rgb(74, 77, 78); margin: 10px 0 20px 0;">Welcome</h1>
                <p style="font-family: 'Lato', sans-serif; font-weight: normal; font-size: 23px; color: rgb(74, 77, 78); line-height: 140%;">You're now subscribed to daily AI research summaries from Takara's Frontier Research Team.</p>
                <p style="font-family: 'Lato', sans-serif; font-weight: normal; font-size: 20px; color: rgb(74, 77, 78); line-height: 140%; margin-top: 20px;">Your first summary will arrive tomorrow at 7am UTC.</p>
            </td>
        </tr>
    </table>

    {{if .Feed}}
    <table width="100%" border="0" cellspacing="0" cellpadding="0" style="padding: 20px;">
        <tr>
            <td>
                <h2 style="font-family: 'Noto Sans', Helvetica, Arial, sans-serif; font-weight: bold; font-size: 40px; color: rgb(74, 77, 78); margin: 10px 0;">Today's Summary: {{ .FormattedDate }}</h2>
                <p style="font-family: 'Lato', sans-serif; font-weight: normal; font-size: 23px; color: rgb(74, 77, 78);">{{ .Feed.Description }}</p>
                <hr style="margin: 16px 0; border: 0; border-top: 1px solid rgba(74, 77, 78, 0.4);" />
                {{range .Items}}
                <div style="margin-bottom: 24px; font-family: 'Lato', sans-serif; font-size: 23px; font-weight: normal; color: rgb(74, 77, 78);">{{.Description}}</div>
                {{end}}
            </td>
        </tr>
    </table>
    {{end}}

    <table width="100%" border="0" cellspacing="0" cellpadding="0" style="padding: 12px 20px; max-width: 100%;">
        <tr>
            <td>
                <table width="100%" border="0" cellspacing="0" cellpadding="0">
                    <tr>
                        <td valign="top">
                            <p style="font-family: 'Lato', sans-serif; line-height: 100%; font-weight: bolder; font-size: 50px; color: rgb(74, 77, 78); margin: 0;">Transforming Humanity</p>
                            <p style="font-family: 'Noto Sans', sans-serif; line-height: 200%; font-weight: bold; font-size: 25px; color: rgb(217, 16, 9); margin: 0;">類を変革する</p>
                        </td>
                        <td align="right" valign="top" style="width: 50%;">
                            <a href="https://takara.ai"><img src="https://tldr.takara.ai/icon.svg" alt="Origami Crane Logo" style="width: 250px; height: 194px; max-width: 100%;" /></a>
                        </td>
                    </tr>
                </table>
                <table width="100%" border="0" cellspacing="0" cellpadding="0" style="text-align: center; font-family: 'Lato', sans-serif; font-size: 12px; color: rgba(74, 77, 78, 0.8);">
                    <tr><td style="padding-top: 20px;">© {{ .CurrentYear }} takara.ai Ltd. All rights reserved.</td></tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>
`

// TemplateData holds all the necessary data for rendering the welcome email.
type TemplateData struct {
	Feed          *RssFeed
	FormattedDate string
	CurrentYear   int
	Items         []struct{ Description template.HTML }
}

// formatDate converts a date string from the RSS feed into a more readable format.
func formatDate(dateStr string) string {
	if dateStr == "" {
		return "Latest Research"
	}
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
		return "Latest Research"
	}
	return t.Format("January 2, 2006")
}

// GenerateWelcomeEmailHTML executes the Go template to produce the welcome email body.
func GenerateWelcomeEmailHTML(feed *RssFeed) (string, error) {
	data := TemplateData{
		Feed:        feed,
		CurrentYear: time.Now().Year(),
	}

	if feed != nil {
		data.FormattedDate = formatDate(feed.LastBuildDate)
		data.Items = make([]struct{ Description template.HTML }, len(feed.Items))
		for i, item := range feed.Items {
			data.Items[i] = struct{ Description template.HTML }{Description: template.HTML(item.Description)}
		}
	}

	tpl, err := template.New("welcomeEmail").Parse(welcomeEmailTemplateStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse welcome email template: %w", err)
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute welcome email template: %w", err)
	}

	return buf.String(), nil
}
