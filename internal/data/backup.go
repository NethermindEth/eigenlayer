package data

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/spf13/afero"
)

type BackupId struct {
	InstanceId string
	Timestamp  time.Time
}

func (b *BackupId) String() string {
	return fmt.Sprintf("%s-%d", b.InstanceId, b.Timestamp.Unix())
}

type Backup struct {
	BackupId
	path string
	fs   afero.Fs
}

// NewBackup creates a new Backup instance from the given path.
func NewBackup(fs afero.Fs, path string) (*Backup, error) {
	backupFileName := filepath.Base(path)
	instanceId, timestamp, err := parseBackupName(backupFileName)
	if err != nil {
		return nil, err
	}
	return &Backup{
		BackupId: BackupId{
			InstanceId: instanceId,
			Timestamp:  timestamp,
		},
		path: path,
		fs:   fs,
	}, nil
}

// Path returns the path of the backup.
func (b *Backup) Path() string {
	return b.path
}

// InstanceId returns the instance ID of the backup.
func (b *Backup) InstanceId() string {
	return b.BackupId.InstanceId
}

// Timestamp returns the timestamp of the backup.
func (b *Backup) Timestamp() time.Time {
	return b.BackupId.Timestamp
}

// Size returns the size of the backup in bytes.
func (b *Backup) Size() (uint64, error) {
	bStat, err := b.fs.Stat(b.path)
	if err != nil {
		return 0, err
	}
	return uint64(bStat.Size()), nil
}

func parseBackupName(backupName string) (instanceId string, timestamp time.Time, err error) {
	backupFileNameRegex := regexp.MustCompile(`^(?P<instance_id>.*)-(?P<timestamp>[0-9]+)\.tar$`)
	match := backupFileNameRegex.FindStringSubmatch(backupName)
	if len(match) != 3 {
		return "", time.Time{}, fmt.Errorf("%w: %s", ErrInvalidBackupName, backupName)
	}
	instanceId = match[1]
	timestampInt, err := strconv.ParseInt(match[2], 10, 64)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("%w: %s", ErrInvalidBackupName, backupName)
	}
	timestamp = time.Unix(timestampInt, 0)
	return instanceId, timestamp, nil
}
