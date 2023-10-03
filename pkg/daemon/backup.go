package daemon

type BackupManager interface {
	// BackupInstance creates a backup of the instance with the given ID.
	BackupInstance(instanceId string) (string, error)
}
