package cli

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/common"
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
			url:  "http://github.com/NethermindEth/mock-avs-pkg.git",
			err:  nil,
		},
		{
			name: "HTTPS URL",
			url:  common.MockAvsPkg.Repo() + ".git",
			err:  nil,
		},
		{
			name: "non HTTP or HTTPS URL",
			url:  "ftp://github.com/NethermidEth/mock-avs-pkg.git",
			err:  fmt.Errorf("%w: URL must be HTTP or HTTPS", ErrInvalidURL),
		},
		{
			name: "URL with IP instead of domain",
			url:  "https://80.58.61.250/NethermidEth/mock-avs-pkg.git",
			err:  nil,
		},
		{
			name: "non absolute URL",
			url:  "github.com/NethermidEth/mock-avs-pkg.git",
			err:  fmt.Errorf("%w: parse \"github.com/NethermidEth/mock-avs-pkg.git\": invalid URI for request", ErrInvalidURL),
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
