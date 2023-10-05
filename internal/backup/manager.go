package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NethermindEth/docker-volumes-snapshotter/pkg/backuptar"
	"github.com/NethermindEth/eigenlayer/internal/compose"
	"github.com/NethermindEth/eigenlayer/internal/data"
	"github.com/NethermindEth/eigenlayer/internal/docker"
	"github.com/compose-spec/compose-go/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	SnapshotterVersion       = "v0.2.0"
	SnapshotterRepo          = "github.com/NethermindEth/docker-volumes-snapshotter"
	SnapshotterRemoteContext = SnapshotterRepo + ".git#" + SnapshotterVersion
	SnapshotterImage         = "eigenlayer-snapshotter:" + SnapshotterVersion
)

type BackupInfo struct {
	Instance  string
	Timestamp time.Time
	SizeBytes uint64
}

type BackupManager struct {
	dataDir    *data.DataDir
	dockerMgr  *docker.DockerManager
	composeMgr *compose.ComposeManager
	fs         afero.Fs
}

func NewBackupManager(fs afero.Fs, dataDir *data.DataDir, dockerMgr *docker.DockerManager, composeMgr *compose.ComposeManager) *BackupManager {
	return &BackupManager{
		dataDir:    dataDir,
		dockerMgr:  dockerMgr,
		composeMgr: composeMgr,
		fs:         fs,
	}
}

// BackupInstance creates a backup of the instance with the given ID.
func (b *BackupManager) BackupInstance(instanceId string) (string, error) {
	if !b.dataDir.HasInstance(instanceId) {
		return "", fmt.Errorf("%w: instance %s", data.ErrInstanceNotFound, instanceId)
	}
	instance, err := b.dataDir.Instance(instanceId)
	if err != nil {
		return "", err
	}
	log.Info("Backing up instance ", instanceId)
	if err := b.buildSnapshotterImage(); err != nil {
		return "", err
	}
	instanceProject, err := instance.ComposeProject()
	if err != nil {
		return "", err
	}

	backup := &data.Backup{
		InstanceId: instanceId,
		Timestamp:  time.Now(),
		Version:    instance.Version,
		Commit:     instance.Commit,
		Url:        instance.URL,
	}

	err = b.dataDir.InitBackup(backup)
	if err != nil {
		return "", err
	}

	// Add volumes of each service
	for _, service := range instanceProject.Services {
		err := b.backupInstanceServiceVolumes(service, backup)
		if err != nil {
			return "", err
		}
	}

	// Add instance data
	err = b.backupInstanceData(instanceId, backup)
	if err != nil {
		return "", err
	}

	// Add timestamp
	err = b.addTimestamp(backup)
	if err != nil {
		return "", err
	}

	return backup.Id(), nil
}

func (b *BackupManager) RestoreInstance(backupId string) error {
	backup, err := b.dataDir.Backup(backupId)
	if err != nil {
		return err
	}

	log.Infof("Restoring backup INSTANCE_ID: %s, VERSION: %s, COMMIT: %s", backup.InstanceId, backup.Version, backup.Commit)

	backupPath := b.dataDir.BackupPath(backup.Id())
	if err != nil {
		return err
	}

	// Restore instance data
	err = b.restoreInstanceData(backup.InstanceId, backupPath)
	if err != nil {
		return err
	}

	if err := b.buildSnapshotterImage(); err != nil {
		return err
	}

	instance, err := b.dataDir.Instance(backup.InstanceId)
	if err != nil {
		return err
	}

	// Create compose project
	err = b.composeMgr.Create(compose.DockerComposeCreateOptions{
		Path: instance.ComposePath(),
	})
	if err != nil {
		return err
	}

	instanceProject, err := instance.ComposeProject()
	if err != nil {
		return err
	}

	// Restore volumes of each service
	for _, service := range instanceProject.Services {
		err := b.restoreInstanceServiceVolumes(service, backupPath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *BackupManager) backupInstanceData(instanceId string, backup *data.Backup) error {
	log.Info("Backing up instance data...")
	backupPath := b.dataDir.BackupPath(backup.Id())

	instancePath, err := b.dataDir.InstancePath(instanceId)
	if err != nil {
		return err
	}
	backupWriter, err := backuptar.NewBackupWriter(backupPath)
	if err != nil {
		return err
	}
	defer backupWriter.Close()
	return backupWriter.AddDir(instancePath, filepath.Join("data"))
}

func (b *BackupManager) backupInstanceServiceVolumes(service types.ServiceConfig, backup *data.Backup) (err error) {
	if len(service.Volumes) == 0 {
		return nil
	}
	log.Infof("Backing up %d volumes from service \"%s\"...", len(service.Volumes), service.Name)
	backupPath := b.dataDir.BackupPath(backup.Id())

	volumes := make([]string, 0, len(service.Volumes))
	for _, v := range service.Volumes {
		volumes = append(volumes, v.Target)
	}
	config := backupConfig{
		Prefix:  snapshotterConfigPrefix(service.Name),
		Volumes: volumes,
	}
	f, err := afero.TempFile(b.fs, os.TempDir(), "eigenlayer-snapshotter-config-*.yml")
	if err != nil {
		return err
	}
	defer f.Close()
	err = config.Save(f)
	if err != nil {
		return err
	}
	err = b.dockerMgr.Run(SnapshotterImage, docker.RunOptions{
		Args:       []string{"backup"},
		AutoRemove: true,
		Mounts: []docker.Mount{
			{
				Type:   docker.VolumeTypeBind,
				Source: f.Name(),
				Target: "/snapshotter.yml",
			},
			{
				Type:   docker.VolumeTypeBind,
				Source: backupPath,
				Target: "/backup.tar",
			},
		},
		VolumesFrom: []string{service.ContainerName},
	})
	if err != nil {
		return fmt.Errorf("snapshotter failed with error: %w", err)
	}
	return nil
}

func (b *BackupManager) addTimestamp(backup *data.Backup) error {
	log.Infof("Adding timestamp %s...", backup.Timestamp.Format(time.DateTime))
	backupPath := b.dataDir.BackupPath(backup.Id())

	timestampTmp, err := afero.TempFile(b.fs, afero.GetTempDir(b.fs, ""), "backup-timestamp-*")
	if err != nil {
		return err
	}
	defer timestampTmp.Close()
	defer b.fs.Remove(timestampTmp.Name())

	_, err = timestampTmp.WriteString(fmt.Sprintf("%d", backup.Timestamp.Unix()))
	if err != nil {
		return err
	}

	backupWriter, err := backuptar.NewBackupWriter(backupPath)
	if err != nil {
		return err
	}
	defer backupWriter.Close()

	return backupWriter.AddFile(timestampTmp.Name(), "timestamp")
}

func (b *BackupManager) restoreInstanceData(instanceId string, backupPath string) error {
	return b.dataDir.ReplaceInstanceDirFromTar(instanceId, backupPath, "data")
}

func (b *BackupManager) restoreInstanceServiceVolumes(service types.ServiceConfig, backupPath string) error {
	if len(service.Volumes) == 0 {
		return nil
	}
	log.Infof("Restoring %d volumes from service \"%s\"...", len(service.Volumes), service.Name)

	volumes := make([]string, 0, len(service.Volumes))
	for _, v := range service.Volumes {
		volumes = append(volumes, v.Target)
	}
	config := backupConfig{
		Prefix:  filepath.Join("volumes", service.Name),
		Volumes: volumes,
	}
	f, err := afero.TempFile(b.fs, os.TempDir(), "eigenlayer-snapshotter-config-*.yml")
	if err != nil {
		return err
	}
	defer f.Close()
	err = config.Save(f)
	if err != nil {
		return err
	}
	err = b.dockerMgr.Run(SnapshotterImage, docker.RunOptions{
		Args:       []string{"restore"},
		AutoRemove: true,
		Mounts: []docker.Mount{
			{
				Type:   docker.VolumeTypeBind,
				Source: f.Name(),
				Target: "/snapshotter.yml",
			},
			{
				Type:   docker.VolumeTypeBind,
				Source: backupPath,
				Target: "/backup.tar",
			},
		},
		VolumesFrom: []string{service.ContainerName},
	})
	if err != nil {
		return fmt.Errorf("snapshotter failed with error: %w", err)
	}
	return nil
}

func (b *BackupManager) buildSnapshotterImage() error {
	ok, err := b.dockerMgr.ImageExist(SnapshotterImage)
	if err != nil {
		return err
	}
	if !ok {
		log.Infof("Building snapshotter image \"%s\" from \"%s\" ...", SnapshotterImage, SnapshotterRemoteContext)
		log.Infof("To learn more about the snapshotter, visit https://%s/tree/%s", SnapshotterRepo, SnapshotterVersion)
		err = b.dockerMgr.BuildImageFromURI(SnapshotterRemoteContext, SnapshotterImage, nil)
		if err != nil {
			return fmt.Errorf("%w: %s", data.ErrCreatingBackup, err.Error())
		}
	}
	return nil
}

func snapshotterConfigPrefix(service string) string {
	return filepath.Join("volumes", service)
}
