package pull

import (
	"testing"

	"github.com/NethermindEth/egn/internal/package_handler"
	"github.com/stretchr/testify/assert"
)

func TestPull(t *testing.T) {
	ts := []struct {
		name    string
		url     string
		version string
		err     error
	}{
		{
			name:    "valid target",
			url:     "https://github.com/NethermindEth/mock-avs.git",
			version: "v0.1.0",
		},
		{
			name:    "valid target, using latest version",
			url:     "https://github.com/NethermindEth/mock-avs.git",
			version: "",
		},
		{
			name:    "invalid version",
			url:     "https://github.com/NethermindEth/mock-avs.git",
			version: "invalid-tag",
			err:     ErrInvalidVersionTag,
		},
		{
			name:    "not found or private",
			url:     "https://github.com/NethermindEth/mock-avs-invalid.git",
			version: "v0.1.0",
			err: package_handler.RepositoryNotFoundOrPrivateError{
				URL: "https://github.com/NethermindEth/mock-avs-invalid.git",
			},
		},
		// TODO: add test case using GitAuth and a private repository
		// TODO: add test case using GitAuth and an invalid username/password
		// TODO: add test case using GitAuth and an invalid url
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Pull(tc.url, tc.version, t.TempDir())
			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err)
			}
		})
	}
}
