package daemon

import (
	"errors"
	"testing"

	"github.com/NethermindEth/eigen-wiz/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestInstall(t *testing.T) {
	ts := []struct {
		name    string
		options *InstallOptions
		err     error
	}{
		{
			name: "success",
			options: &InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs.git",
				Version: "v0.1.0",
				DestDir: "/tmp",
			},
			err: nil,
		},
		{
			name: "error",
			options: &InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs.git",
				Version: "v0.1.0",
				DestDir: "/tmp",
			},
			err: errors.New("installer error"),
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			installer := mocks.NewMockInstaller(gomock.NewController(t))
			installer.EXPECT().
				Install(tc.options.URL, tc.options.Version, tc.options.DestDir).
				Return(tc.err)

			d := NewDaemon(installer)
			response, err := d.Install(tc.options)
			if tc.err != nil {
				assert.EqualError(t, err, tc.err.Error())
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, &InstallResponse{}, response)
			}
		})
	}
}
