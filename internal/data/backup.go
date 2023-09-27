package data

import (
	"fmt"
	"time"
)

type BackupId struct {
	InstanceId string
	Timestamp  time.Time
}

func (b *BackupId) String() string {
	return fmt.Sprintf("%s-%d", b.InstanceId, b.Timestamp.Unix())
}

type Backup struct {
	Id   BackupId
	path string
}

func (b Backup) Path() string {
	return b.path
}
