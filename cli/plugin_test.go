package cli

import (
	"errors"
	"testing"

	daemonMock "github.com/NethermindEth/eigenlayer/cli/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPluginCmd(t *testing.T) {
	ts := []struct {
		name       string
		args       []string
		err        error
		daemonMock func(d *daemonMock.MockDaemon)
	}{
		{
			name: "no arguments",
			args: []string{},
			err:  errors.New("requires at least 1 arg(s), only received 0"),
		},
		{
			name: "instance not found",
			args: []string{"instance1"},
			err:  errors.New("instance not found"),
			daemonMock: func(d *daemonMock.MockDaemon) {
				d.EXPECT().HasInstance("instance1").Return(false)
			},
		},
		{
			name: "run plugin error",
			args: []string{"instance1", "arg1"},
			err:  errors.New("run plugin error"),
			daemonMock: func(d *daemonMock.MockDaemon) {
				gomock.InOrder(
					d.EXPECT().HasInstance("instance1").Return(true),
					d.EXPECT().RunPlugin("instance1", []string{"arg1"}, false).Return(errors.New("run plugin error")),
				)
			},
		},
		{
			name: "valid arguments",
			args: []string{"instance1", "arg1"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon) {
				gomock.InOrder(
					d.EXPECT().HasInstance("instance1").Return(true),
					d.EXPECT().RunPlugin("instance1", []string{"arg1"}, false).Return(nil),
				)
			},
		},
		{
			name: "plugin without arguments",
			args: []string{"instance1"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon) {
				gomock.InOrder(
					d.EXPECT().HasInstance("instance1").Return(true),
					d.EXPECT().RunPlugin("instance1", []string{}, false).Return(nil),
				)
			},
		},
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			d := daemonMock.NewMockDaemon(controller)
			if tc.daemonMock != nil {
				tc.daemonMock(d)
			}

			pluginCmd := PluginCmd(d)

			pluginCmd.SetArgs(tc.args)
			err := pluginCmd.Execute()

			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}
