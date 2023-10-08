package cli

import (
	"bytes"
	"testing"
	"time"

	"github.com/NethermindEth/eigenlayer/cli/mocks"
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupLs(t *testing.T) {
	tc := []struct {
		name   string
		err    error
		stdErr []byte
		stdOut []byte
		mocker func(d *mocks.MockDaemon)
	}{
		{
			name:   "no backups",
			err:    nil,
			stdErr: nil,
			stdOut: []byte("ID    AVS Instance ID    VERSION    COMMIT    TIMESTAMP    SIZE    URL    \n"),
			mocker: func(d *mocks.MockDaemon) {
				d.EXPECT().BackupList().Return([]daemon.BackupInfo{}, nil)
			},
		},
		{
			name:   "with backups",
			err:    nil,
			stdErr: nil,
			stdOut: []byte(
				"ID          AVS Instance ID     VERSION    COMMIT                                      TIMESTAMP              SIZE     URL                                              \n" +
					"7ba32f63    mock-avs-second     v5.5.1     d5af645fffb93e8263b099082a4f512e1917d0af    2023-10-04 07:12:19    10KiB    https://github.com/NethermindEth/mock-avs-pkg    \n" +
					"33de69fe    mock-avs-default    v5.5.0     a3406616b848164358fdd24465b8eecda5f5ae34    2023-10-03 21:18:36    10KiB    https://github.com/NethermindEth/mock-avs-pkg    \n",
			),
			mocker: func(d *mocks.MockDaemon) {
				d.EXPECT().BackupList().Return([]daemon.BackupInfo{
					{
						Id:        "33de69fe9225b95c8fb909cb418e5102970c8d73",
						Instance:  "mock-avs-default",
						Version:   "v5.5.0",
						Commit:    "a3406616b848164358fdd24465b8eecda5f5ae34",
						Timestamp: time.Date(2023, 10, 3, 21, 18, 36, 0, time.UTC),
						SizeBytes: 10240,
						Url:       "https://github.com/NethermindEth/mock-avs-pkg",
					},
					{
						Id:        "7ba32f630af2cede1388b5712d6ef3ac63175bae",
						Instance:  "mock-avs-second",
						Version:   "v5.5.1",
						Commit:    "d5af645fffb93e8263b099082a4f512e1917d0af",
						Timestamp: time.Date(2023, 10, 4, 7, 12, 19, 0, time.UTC),
						SizeBytes: 10240,
						Url:       "https://github.com/NethermindEth/mock-avs-pkg",
					},
				}, nil)
			},
		},
		{
			name:   "error",
			err:    assert.AnError,
			stdErr: []byte("Error: " + assert.AnError.Error() + "\n"),
			stdOut: []byte{},
			mocker: func(d *mocks.MockDaemon) {
				d.EXPECT().BackupList().Return(nil, assert.AnError)
			},
		},
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			d := mocks.NewMockDaemon(ctrl)

			tt.mocker(d)

			var (
				stdOut bytes.Buffer
				stdErr bytes.Buffer
			)

			backupLsCmd := BackupLsCmd(d)
			backupLsCmd.SetOut(&stdOut)
			backupLsCmd.SetErr(&stdErr)
			err := backupLsCmd.Execute()

			if tt.err != nil {
				require.Error(t, err)
				assert.EqualError(t, err, tt.err.Error())
				assert.Equal(t, tt.stdErr, stdErr.Bytes())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.stdOut, stdOut.Bytes())
			}
		})
	}
}
