package egothic

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// redirectHTMLTemplate is the pre-optimized HTML template for the redirect page.
const redirectHTMLTemplate = `<html><head><meta http-equiv="refresh" content="0;url=%s"><script type="text/javascript">window.location.href="%s";</script></head><body><p>If you are not redirected automatically, please <a href="%s">click here</a>.</p></body></html>`

// Redirect redirects the user to the given URL.
// This method attempts to avoid browser caching by setting appropriate headers.
// It attempts a server-side redirect first, and if that fails, it sends a page with JavaScript redirect.
func Redirect(e echo.Context, url string, opts ...Options) error {
	config := newConfig(opts...)

	// Set headers to prevent caching
	e.Response().Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	e.Response().Header().Set("Pragma", "no-cache")
	e.Response().Header().Set("Expires", "0")

	if url == "" {
		config.log("Error: Empty URL provided for redirect")
		return e.String(http.StatusBadRequest, "Empty URL provided for redirect")
	}

	// Attempt server-side redirect first
	config.log("Redirecting to '" + url + "'")
	if e.Request().Header.Get("X-Force-HTML-Redirect") != "true" {
		if err := e.Redirect(http.StatusSeeOther, url); err == nil {
			config.log("Server-side redirect to '" + url + "' succeeded")
			return nil
		}
	}
	config.log("Server-side redirect to '" + url + "' failed")

	// If server-side redirect fails, send a page with JavaScript redirect
	config.log("Sending JavaScript redirect to '" + url + "'")
	html := fmt.Sprintf(redirectHTMLTemplate, url, url, url)
	return e.HTML(http.StatusOK, html)
}
