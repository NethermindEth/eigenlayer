package daemon

type BackupManager interface {
	BackupInstance(instanceId string) (string, error)
}
