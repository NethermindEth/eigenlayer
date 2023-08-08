package cli

import (
	"bytes"
	"testing"

	daemonMock "github.com/NethermindEth/eigenlayer/cli/mocks"
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestList(t *testing.T) {
	tests := []struct {
		name   string
		mocker func(d *daemonMock.MockDaemon)
		err    error
		stdOut []byte
		errOut []byte
	}{
		{
			name: "success",
			mocker: func(d *daemonMock.MockDaemon) {
				d.EXPECT().ListInstances().Return([]daemon.ListInstanceItem{
					{
						ID:      "id1",
						Running: true,
						Health:  daemon.NodeHealthy,
						Comment: "comment1",
						Version: "v3.1.1",
						Commit:  "d1d4bb7009549c431d7b3317f004a56e2c3b2031",
					}, {
						ID:      "id2",
						Running: false,
						Health:  daemon.NodeHealthUnknown,
						Comment: "comment2",
						Version: "v3.1.1",
						Commit:  "d1d4bb7009549c431d7b3317f004a56e2c3b2031",
					},
				}, nil)
			},
			stdOut: []byte(
				"AVS Instance ID    RUNNING    HEALTH     VERSION    COMMIT          COMMENT     \n" +
					"id1                true       healthy    v3.1.1     d1d4bb700954    comment1    \n" +
					"id2                false      unknown    v3.1.1     d1d4bb700954    comment2    \n",
			),
		},
		{
			name: "success, short commit hash",
			mocker: func(d *daemonMock.MockDaemon) {
				d.EXPECT().ListInstances().Return([]daemon.ListInstanceItem{
					{
						ID:      "id1",
						Running: true,
						Health:  daemon.NodeHealthy,
						Comment: "comment1",
						Version: "v3.1.1",
						Commit:  "d1d4bb7009549c431d7b3317f004a56e2c3b2031",
					}, {
						ID:      "id2",
						Running: false,
						Health:  daemon.NodeHealthUnknown,
						Comment: "comment2",
						Version: "v3.1.1",
						Commit:  "d1d4bb7",
					},
				}, nil)
			},
			stdOut: []byte(
				"AVS Instance ID    RUNNING    HEALTH     VERSION    COMMIT          COMMENT     \n" +
					"id1                true       healthy    v3.1.1     d1d4bb700954    comment1    \n" +
					"id2                false      unknown    v3.1.1     d1d4bb7         comment2    \n",
			),
		},
		{
			name: "success, empty list",
			mocker: func(d *daemonMock.MockDaemon) {
				d.EXPECT().ListInstances().Return([]daemon.ListInstanceItem{}, nil)
			},
			stdOut: []byte(
				"AVS Instance ID    RUNNING    HEALTH    VERSION    COMMIT    COMMENT    \n",
			),
		},
		{
			name: "daemon list error",
			mocker: func(d *daemonMock.MockDaemon) {
				d.EXPECT().ListInstances().Return(nil, assert.AnError)
			},
			err:    assert.AnError,
			errOut: []byte("Error: " + assert.AnError.Error() + "\n"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := daemonMock.NewMockDaemon(gomock.NewController(t))
			if tt.mocker != nil {
				tt.mocker(d)
			}

			var (
				stdOut bytes.Buffer
				errOut bytes.Buffer
			)

			cmd := ListCmd(d)
			cmd.SetOut(&stdOut)
			cmd.SetErr(&errOut)
			err := cmd.Execute()

			if tt.err != nil {
				assert.ErrorIs(t, tt.err, err)
				assert.Equal(t, tt.errOut, errOut.Bytes())
			} else {
				assert.NoError(t, err)
				assert.Empty(t, errOut.Bytes())
				assert.Equal(t, tt.stdOut, stdOut.Bytes())
			}
		})
	}
}
