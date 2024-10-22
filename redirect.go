package egothic

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"text/template"

	"github.com/labstack/echo/v4"
)

// redirectHTMLTemplate is the HTML template for the redirect page.
const redirectHTMLTemplate = `
<html>
	<head>
		<meta http-equiv="refresh" content="0;url=%s">
		<script type="text/javascript">
			window.location.href = "%s";
		</script>
	</head>
	<body>
		<p>If you are not redirected automatically, please <a href="%s">click here</a>.</p>
	</body>
</html>
`

// optimizedRedirectHTMLTemplate is the optimized HTML template for the redirect page.
var optimizedRedirectHTMLTemplate = optimizeTemplate(redirectHTMLTemplate)

// optimizeTemplate optimizes the HTML template.
func optimizeTemplate(template string) string {
	// Remove all newlines and tabs
	optimized := strings.ReplaceAll(template, "\n", "")
	optimized = strings.ReplaceAll(optimized, "\t", "")
	// Remove extra spaces
	optimized = regexp.MustCompile(`\s+`).ReplaceAllString(optimized, " ")
	// Remove spaces between tags
	optimized = regexp.MustCompile(`>\s+<`).ReplaceAllString(optimized, "><")
	return strings.TrimSpace(optimized)
}

func Redirect(e echo.Context, url string, opts ...Options) error {
	config := newConfig(opts...)

	// Set headers to prevent caching
	e.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	e.Response().Header().Set("Pragma", "no-cache")
	e.Response().Header().Set("Expires", "0")

	// Attempt server-side redirect first
	config.log("Redirecting to '" + url + "'")
	if err := e.Redirect(http.StatusSeeOther, url); err == nil {
		config.log("Server-side redirect to '" + url + "' succeeded")
		return nil
	}
	config.log("Server-side redirect to '" + url + "' failed")

	// If server-side redirect fails, send a page with JavaScript redirect
	e.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	escapedURL := template.HTMLEscapeString(url)
	html := fmt.Sprintf(optimizedRedirectHTMLTemplate, escapedURL, escapedURL, escapedURL)
	config.log("Sending JavaScript redirect to '" + url + "'")
	return e.HTML(http.StatusOK, html)
}
