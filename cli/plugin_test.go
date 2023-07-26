package cli

import (
	"errors"
	"testing"

	daemonMock "github.com/NethermindEth/eigenlayer/cli/mocks"
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestPluginCmd(t *testing.T) {
	type testCase struct {
		name       string
		args       []string
		err        error
		daemonMock func(d *daemonMock.MockDaemon)
	}
	ts := []testCase{
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
					d.EXPECT().RunPlugin("instance1", []string{"arg1"}, daemon.RunPluginOptions{
						NoDestroyImage: false,
						HostNetwork:    false,
						Binds:          map[string]string{},
						Volumes:        map[string]string{},
					}).Return(errors.New("run plugin error")),
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
					d.EXPECT().RunPlugin("instance1", []string{"arg1"}, daemon.RunPluginOptions{
						NoDestroyImage: false,
						HostNetwork:    false,
						Binds:          map[string]string{},
						Volumes:        map[string]string{},
					}).Return(nil),
				)
			},
		},
		{
			name: "--host flag",
			args: []string{"--host", "instance1", "arg1"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon) {
				gomock.InOrder(
					d.EXPECT().HasInstance("instance1").Return(true),
					d.EXPECT().RunPlugin("instance1", []string{"arg1"}, daemon.RunPluginOptions{
						NoDestroyImage: false,
						HostNetwork:    true,
						Binds:          map[string]string{},
						Volumes:        map[string]string{},
					}).Return(nil),
				)
			},
		},
		{
			name: "--host flag, but as a plugin argument",
			args: []string{"instance1", "--host", "arg1"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon) {
				gomock.InOrder(
					d.EXPECT().HasInstance("instance1").Return(true),
					d.EXPECT().RunPlugin("instance1", []string{"--host", "arg1"}, daemon.RunPluginOptions{
						NoDestroyImage: false,
						HostNetwork:    false,
						Binds:          map[string]string{},
						Volumes:        map[string]string{},
					}).Return(nil),
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
					d.EXPECT().RunPlugin("instance1", nil, daemon.RunPluginOptions{
						NoDestroyImage: false,
						HostNetwork:    false,
						Binds:          map[string]string{},
						Volumes:        map[string]string{},
					}).Return(nil),
				)
			},
		},
		func(t *testing.T) testCase {
			hostDir := t.TempDir()
			return testCase{
				name: "--volume flag",
				args: []string{
					"--volume", hostDir + ":/container",
					"-v", "docker-volume:/container/path",
					"instance1", "arg1",
				},
				err: nil,
				daemonMock: func(d *daemonMock.MockDaemon) {
					gomock.InOrder(
						d.EXPECT().HasInstance("instance1").Return(true),
						d.EXPECT().RunPlugin("instance1", []string{"arg1"}, daemon.RunPluginOptions{
							NoDestroyImage: false,
							HostNetwork:    false,
							Binds:          map[string]string{hostDir: "/container"},
							Volumes:        map[string]string{"docker-volume": "/container/path"},
						}).Return(nil),
					)
				},
			}
		}(t),
		{
			name: "--volume flag, but as a plugin argument",
			args: []string{"instance1", "--volume", "arg1"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon) {
				gomock.InOrder(
					d.EXPECT().HasInstance("instance1").Return(true),
					d.EXPECT().RunPlugin("instance1", []string{"--volume", "arg1"}, daemon.RunPluginOptions{
						NoDestroyImage: false,
						HostNetwork:    false,
						Binds:          map[string]string{},
						Volumes:        map[string]string{},
					}).Return(nil),
				)
			},
		},
		{
			name: "--no-rm-image flag",
			args: []string{"--no-rm-image", "instance1", "arg1"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon) {
				gomock.InOrder(
					d.EXPECT().HasInstance("instance1").Return(true),
					d.EXPECT().RunPlugin("instance1", []string{"arg1"}, daemon.RunPluginOptions{
						NoDestroyImage: true,
						HostNetwork:    false,
						Binds:          map[string]string{},
						Volumes:        map[string]string{},
					}).Return(nil),
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
