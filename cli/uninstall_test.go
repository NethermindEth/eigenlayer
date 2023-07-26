package cli

import (
	"errors"
	"testing"

	daemonMock "github.com/NethermindEth/eigenlayer/cli/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUninstall(t *testing.T) {
	ts := []struct {
		name   string
		args   []string
		err    error
		mocker func(d *daemonMock.MockDaemon)
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
			name: "success",
			args: []string{"instance1"},
			err:  nil,
			mocker: func(d *daemonMock.MockDaemon) {
				gomock.InOrder(
					d.EXPECT().InitMonitoring(false, false).Return(nil),
					d.EXPECT().Uninstall("instance1").Return(nil),
				)
			},
		},
		{
			name: "init monitoring error",
			args: []string{"instance1"},
			err:  assert.AnError,
			mocker: func(d *daemonMock.MockDaemon) {
				d.EXPECT().InitMonitoring(false, false).Return(assert.AnError)
			},
		},
		{
			name: "uninstall error",
			args: []string{"instance1"},
			err:  errors.New("uninstall error"),
			mocker: func(d *daemonMock.MockDaemon) {
				gomock.InOrder(
					d.EXPECT().InitMonitoring(false, false).Return(nil),
					d.EXPECT().Uninstall("instance1").Return(errors.New("uninstall error")),
				)
			},
		},
	}
	for _, tt := range ts {
		t.Run(tt.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			d := daemonMock.NewMockDaemon(controller)

			if tt.mocker != nil {
				tt.mocker(d)
			}

			uninstallCmd := UninstallCmd(d)

			uninstallCmd.SetArgs(tt.args)
			err := uninstallCmd.Execute()

			if tt.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}
