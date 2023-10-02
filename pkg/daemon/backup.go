package daemon

import "github.com/NethermindEth/eigenlayer/internal/backup"

type BackupManager interface {
	// BackupInstance creates a backup of the instance with the given ID.
	BackupInstance(instanceId string) (string, error)
	// BackupList returns a list of all backups.
	BackupList() ([]backup.BackupInfo, error)
}
