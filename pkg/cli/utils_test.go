package cli

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateURL(t *testing.T) {
	ts := []struct {
		name string
		url  string
		err  error
	}{
		{
			name: "empty URL",
			url:  "",
			err:  fmt.Errorf("%w: parse \"\": empty url", ErrInvalidURL),
		},
		{
			name: "HTTP URL",
			url:  "http://github.com/NethermidEth/mock-avs.git",
			err:  nil,
		},
		{
			name: "HTTPS URL",
			url:  "https://github.com/NethermidEth/mock-avs.git",
			err:  nil,
		},
		{
			name: "non HTTP or HTTPS URL",
			url:  "ftp://github.com/NethermidEth/mock-avs.git",
			err:  fmt.Errorf("%w: URL must be HTTP or HTTPS", ErrInvalidURL),
		},
		{
			name: "non absolute URL",
			url:  "github.com/NethermidEth/mock-avs.git",
			err:  fmt.Errorf("%w: parse \"github.com/NethermidEth/mock-avs.git\": invalid URI for request", ErrInvalidURL),
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			err := validatePkgURL(tc.url)
			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}
