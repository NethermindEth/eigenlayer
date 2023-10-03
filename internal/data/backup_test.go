package data

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseBackupName(t *testing.T) {
	tc := []struct {
		name       string
		backupName string
		instanceId string
		timestamp  time.Time
		err        error
	}{
		{
			name:       "valid backup name",
			backupName: "mock-avs-default-1696317683.tar",
			instanceId: "mock-avs-default",
			timestamp:  time.Unix(1696317683, 0),
			err:        nil,
		},
		{
			name:       "no .tar file",
			backupName: "mock-avs-default-1696317683",
			instanceId: "",
			timestamp:  time.Time{},
			err:        ErrInvalidBackupName,
		},
		{
			name:       "without dash separator between instance ID and timestamp",
			backupName: "mock-avs-default1696317683.tar",
			instanceId: "",
			timestamp:  time.Time{},
			err:        ErrInvalidBackupName,
		},
		{
			name:       "invalid timestamp",
			backupName: "mock-avs-default-1696317683a.tar",
			instanceId: "",
			timestamp:  time.Time{},
			err:        ErrInvalidBackupName,
		},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			instanceId, timestamp, err := ParseBackupName(tt.backupName)
			if tt.err != nil {
				assert.ErrorIs(t, err, tt.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.instanceId, instanceId)
				assert.Equal(t, tt.timestamp.Unix(), timestamp.Unix())
			}
		})
	}
}
