package http

import (
	"fmt"
	"net/url"
	"strings"
)

func getBaseUrl(s string) (string, error) {
	// Reject URLs that start with "://" (missing scheme) or contain spaces
	if strings.HasPrefix(s, SchemeSeparator) || strings.Contains(s, " ") {
		return "", &url.Error{Op: "parse", URL: s, Err: url.InvalidHostError(fmt.Sprintf("invalid URL format: %s", s))}
	}

	// If no scheme separator is present, prepend default HTTP scheme
	if !strings.Contains(s, SchemeSeparator) {
		s = DefaultScheme + SchemeSeparator + s
	}

	// Parse the URL
	u, err := url.Parse(s)
	if err != nil {
		return "", &url.Error{Op: "parse", URL: s, Err: err}
	}

	return u.String(), nil
}

func prepBaseRequest(reqHeaders map[string]string) (map[string]string, error) {
	if reqHeaders == nil {
		reqHeaders = make(map[string]string)
	}
	return reqHeaders, nil
}
