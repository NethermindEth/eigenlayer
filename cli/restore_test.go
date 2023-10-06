package cli

import (
	"errors"
	"testing"

	"github.com/NethermindEth/eigenlayer/cli/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRestore(t *testing.T) {
	tc := []struct {
		name   string
		args   []string
		err    error
		mocker func(d *mocks.MockDaemon)
	}{
		{
			name: "no args",
			args: []string{},
			err:  errors.New("accepts 1 arg(s), received 0"),
		},
		{
			name: "daemon restore error",
			args: []string{"backup-id"},
			err:  assert.AnError,
			mocker: func(d *mocks.MockDaemon) {
				d.EXPECT().Restore("backup-id", false).Return(assert.AnError)
			},
		},
		{
			name: "daemon restore success",
			args: []string{"backup-id"},
			mocker: func(d *mocks.MockDaemon) {
				d.EXPECT().Restore("backup-id", false).Return(nil)
			},
		},
		{
			name: "restore with run flag",
			args: []string{"backup-id", "--run"},
			mocker: func(d *mocks.MockDaemon) {
				d.EXPECT().Restore("backup-id", true).Return(nil)
			},
		},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			d := mocks.NewMockDaemon(ctrl)

			if tt.mocker != nil {
				tt.mocker(d)
			}

			cmd := RestoreCmd(d)

			cmd.SetArgs(tt.args)
			err := cmd.Execute()

			if tt.err != nil {
				assert.EqualError(t, err, tt.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
