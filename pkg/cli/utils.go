package cli

import (
	"fmt"
	"net/url"
)

func validatePkgURL(urlStr string) error {
	parsedURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidURL, err.Error())
	}
	if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
		return fmt.Errorf("%w: %s", ErrInvalidURL, "URL must be HTTP or HTTPS")
	}
	return nil
}
