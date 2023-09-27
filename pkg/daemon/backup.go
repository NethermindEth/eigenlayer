package daemon

import "github.com/NethermindEth/eigenlayer/internal/backup"

type BackupManager interface {
	BackupInstance(instanceId string) (string, error)
	BackupList() ([]backup.BackupInfo, error)
}
