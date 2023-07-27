package cli

import (
	"errors"
	"testing"

	daemonMock "github.com/NethermindEth/eigenlayer/cli/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	ts := []struct {
		name   string
		args   []string
		err    error
		mocker func(d *daemonMock.MockDaemon)
	}{
		{
			name:   "no arguments",
			args:   []string{},
			err:    errors.New("accepts 1 arg(s), received 0"),
			mocker: nil,
		},
		{
			name:   "more than one argument",
			args:   []string{"arg1", "arg2"},
			err:    errors.New("accepts 1 arg(s), received 2"),
			mocker: nil,
		},
		{
			name: "valid arguments, and run success",
			args: []string{"mock-avs-default"},
			err:  nil,
			mocker: func(d *daemonMock.MockDaemon) {
				gomock.InOrder(
					d.EXPECT().InitMonitoring(false, false).Return(nil),
					d.EXPECT().Run("mock-avs-default").Return(nil),
				)
			},
		},
		{
			name: "valid arguments, and run error",
			args: []string{"mock-avs-default"},
			err:  assert.AnError,
			mocker: func(d *daemonMock.MockDaemon) {
				gomock.InOrder(
					d.EXPECT().InitMonitoring(false, false).Return(nil),
					d.EXPECT().Run("mock-avs-default").Return(assert.AnError),
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

			runCmd := RunCmd(d)
			runCmd.SetArgs(tt.args)
			err := runCmd.Execute()

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}
