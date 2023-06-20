package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/NethermindEth/egn/pkg/daemon/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestInstall_ValidateArguments(t *testing.T) {
	ts := []struct {
		name       string
		args       []string
		err        error
		daemonMock func(d *mocks.MockDaemon)
	}{
		{
			name: "no arguments",
			args: []string{},
			err:  errors.New("accepts 1 arg(s), received 0"),
		},
		{
			name: "more than one argument",
			args: []string{"arg1", "arg2"},
			err:  errors.New("accepts 1 arg(s), received 2"),
		},
		{
			name: "invalid URL",
			args: []string{"invalid-url"},
			err:  fmt.Errorf("%w: parse \"invalid-url\": invalid URI for request", ErrInvalidURL),
		},
		{
			name: "valid arguments",
			args: []string{"-v", "v2.0.2", "https://github.com/NethermindEth/mock-avs"},
			err:  nil,
			daemonMock: func(d *mocks.MockDaemon) {
				d.EXPECT().Install(gomock.Any()).Return(nil).Times(1)
			},
		},
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			d := mocks.NewMockDaemon(gomock.NewController(t))
			if tc.daemonMock != nil {
				tc.daemonMock(d)
			}

			installCmd := InstallCmd(d)

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
