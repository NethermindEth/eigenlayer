package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/NethermindEth/eigen-wiz/mocks"
	"github.com/NethermindEth/eigen-wiz/pkg/daemon"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestInstall_ValidateArguments(t *testing.T) {
	type testCase struct {
		name string
		d    daemon.Daemon
		args []string
		err  error
	}
	ts := []testCase{
		{
			name: "no arguments",
			d:    nil,
			args: []string{},
			err:  errors.New("accepts 1 arg(s), received 0"),
		},
		{
			name: "more than one argument",
			d:    daemon.NewWizDaemon(nil),
			args: []string{"arg1", "arg2"},
			err:  errors.New("accepts 1 arg(s), received 2"),
		},
		func() testCase {
			url := "http://github.com/NethermidEth/mock-avs.git"

			mockInstaller := mocks.NewMockInstaller(gomock.NewController(t))
			mockInstaller.EXPECT().Install(url, "latest", gomock.Any()).Return(nil)

			return testCase{
				name: "HTTP URL as argument",
				d:    daemon.NewWizDaemon(mockInstaller),
				args: []string{url},
				err:  nil,
			}
		}(),
		func() testCase {
			url := "https://github.com/NethermidEth/mock-avs.git"

			mockInstaller := mocks.NewMockInstaller(gomock.NewController(t))
			mockInstaller.EXPECT().Install(url, "latest", gomock.Any()).Return(nil)

			return testCase{
				name: "HTTPS URL as argument",
				d:    daemon.NewWizDaemon(mockInstaller),
				args: []string{url},
				err:  nil,
			}
		}(),
		{
			name: "non HTTP or HTTPS URL as argument",
			d:    daemon.NewWizDaemon(nil),
			args: []string{"ftp://github.com/NethermidEth/mock-avs.git"},
			err:  fmt.Errorf("%w: %s", ErrInvalidURL, "URL must be HTTP or HTTPS"),
		},
		{
			name: "non absolute URL as argument",
			d:    daemon.NewWizDaemon(nil),
			args: []string{"github.com/NethermidEth/mock-avs.git"},
			err:  fmt.Errorf("%w: %s", ErrInvalidURL, "parse \"github.com/NethermidEth/mock-avs.git\": invalid URI for request"),
		},
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			installCmd := InstallCmd(tc.d)

			installCmd.SetArgs(tc.args)
			err := installCmd.Execute()

			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}

func TestInstall_ExecutesInstall(t *testing.T) {
	ts := []struct {
		name string
		d    daemon.Daemon
		args []string
		err  error
	}{
		{
			name: "only URL",
			d: func() daemon.Daemon {
				installer := mocks.NewMockInstaller(gomock.NewController(t))
				installer.EXPECT().Install("https://github.com/NethermindEth/mock-avs.git", "latest", "").Return(nil)
				return daemon.NewWizDaemon(installer)
			}(),
			args: []string{"https://github.com/NethermindEth/mock-avs.git"},
			err:  nil,
		},
		{
			name: "URL and version flag",
			d: func() daemon.Daemon {
				installer := mocks.NewMockInstaller(gomock.NewController(t))
				installer.EXPECT().Install("https://github.com/NethermindEth/mock-avs.git", "v0.1.0", "").Return(nil)
				return daemon.NewWizDaemon(installer)
			}(),
			args: []string{"--version", "v0.1.0", "https://github.com/NethermindEth/mock-avs.git"},
			err:  nil,
		},
		{
			name: "URL and version flag shorthand",
			d: func() daemon.Daemon {
				installer := mocks.NewMockInstaller(gomock.NewController(t))
				installer.EXPECT().Install("https://github.com/NethermindEth/mock-avs.git", "v0.1.0", "").Return(nil)
				return daemon.NewWizDaemon(installer)
			}(),
			args: []string{"-v", "v0.1.0", "https://github.com/NethermindEth/mock-avs.git"},
			err:  nil,
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			installCmd := InstallCmd(tc.d)
			installCmd.SetArgs(tc.args)
			err := installCmd.Execute()
			assert.NoError(t, err)
		})
	}
}
