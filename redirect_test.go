//go:build !integration
// +build !integration

package egothic

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRedirect(t *testing.T) {
	tests := []struct {
		name              string
		url               string
		forceHTMLRedirect bool
		expectedStatus    int
		expectedBody      string
		expectError       bool
	}{
		{
			name:              "Valid URL - Server-side redirect",
			url:               "https://example.com",
			forceHTMLRedirect: false,
			expectedStatus:    http.StatusSeeOther,
			expectedBody:      "",
			expectError:       false,
		},
		{
			name:              "Empty URL",
			url:               "",
			forceHTMLRedirect: false,
			expectedStatus:    http.StatusBadRequest,
			expectedBody:      "Empty URL provided for redirect",
			expectError:       false,
		},
		{
			name:              "Client-side redirect",
			url:               "https://example.com",
			forceHTMLRedirect: true,
			expectedStatus:    http.StatusOK,
			expectedBody:      "<html><head><meta http-equiv=\"refresh\" content=\"0;url=https://example.com\">",
			expectError:       false,
		},
		{
			name:              "Relative URL",
			url:               "/relative/path",
			forceHTMLRedirect: false,
			expectedStatus:    http.StatusSeeOther,
			expectedBody:      "",
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.forceHTMLRedirect {
				req.Header.Set("X-Force-HTML-Redirect", "true")
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := Redirect(c, tt.url)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedStatus, rec.Code)

			if tt.expectedBody != "" {
				assert.True(t, strings.Contains(rec.Body.String(), tt.expectedBody))
			}

			// Check headers
			assert.Equal(t, "no-store, no-cache, must-revalidate, max-age=0", rec.Header().Get("Cache-Control"))
			assert.Equal(t, "no-cache", rec.Header().Get("Pragma"))
			assert.Equal(t, "0", rec.Header().Get("Expires"))

			// Add new checks for redirect URL
			if tt.expectedStatus == http.StatusSeeOther {
				assert.Equal(t, tt.url, rec.Header().Get("Location"), "Location header should match the input URL")
			} else if tt.forceHTMLRedirect {
				expectedContent := fmt.Sprintf("0;url=%s", tt.url)
				assert.Contains(t, rec.Body.String(), expectedContent, "HTML redirect content should contain the input URL")
			}
		})
	}
}
