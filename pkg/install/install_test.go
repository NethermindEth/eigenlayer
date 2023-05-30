package install_test

import (
	"testing"

	"github.com/NethermindEth/eigen-wiz/pkg/install"
	"github.com/stretchr/testify/assert"
)

func TestInstall(t *testing.T) {
	ts := []struct {
		name      string
		installer *install.Installer
		url       string
		version   string
		err       error
	}{
		{
			name:      "valid target",
			installer: install.NewInstaller(),
			url:       "https://github.com/NethermindEth/sedge.git", // TODO: use a mock tap repository
			version:   "v1.1.0",
		},
		{
			name:      "invalid version",
			installer: install.NewInstaller(),
			url:       "https://github.com/NethermindEth/sedge.git", // TODO: use a mock tap repository
			version:   "invalid-tag",
			err: install.TagNotFoundError{
				Tag: "invalid-tag",
			},
		},
		{
			name:      "not found or private",
			installer: install.NewInstaller(),
			url:       "https://github.com/NethermindEth/sedge-invalid.git", // TODO: use a mock tap repository
			version:   "v1.1.0",
			err: install.RepositoryNotFoundOrPrivateError{
				URL: "https://github.com/NethermindEth/sedge-invalid.git",
			},
		},
		// TODO: add testcase using GitAuth and a private repository
		// TODO: add testcase using GitAuth and an invalid username/password
		// TODO: add testcase using GitAuth and an invalid url
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.installer.Install(tc.url, tc.version, t.TempDir())
			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tc.err, err)
			}
		})
	}
}
