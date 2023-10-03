package backup

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NethermindEth/eigenlayer/internal/data"
	"github.com/NethermindEth/eigenlayer/internal/docker"
	"github.com/NethermindEth/eigenlayer/internal/utils"
	"github.com/compose-spec/compose-go/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	SnapshotterVersion       = "v0.1.0"
	SnapshotterRemoteContext = "github.com/NethermindEth/docker-volumes-snapshotter.git#" + SnapshotterVersion
	SnapshotterImage         = "eigenlayer-snapshotter:" + SnapshotterVersion
)

type BackupInfo struct {
	Instance  string
	Timestamp time.Time
	SizeBytes uint64
}

type BackupManager struct {
	dataDir   *data.DataDir
	dockerMgr *docker.DockerManager
	fs        afero.Fs
}

func NewBackupManager(fs afero.Fs, dataDir *data.DataDir, dockerMgr *docker.DockerManager) *BackupManager {
	return &BackupManager{
		dataDir:   dataDir,
		dockerMgr: dockerMgr,
		fs:        fs,
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

	backupId := data.BackupId{
		InstanceId: instanceId,
		Timestamp:  time.Now(),
	}
	backup, err := b.dataDir.InitBackup(backupId)
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

	return backup.BackupId.String(), nil
}

// BackupList returns a list of all backups.
func (b *BackupManager) BackupList() ([]BackupInfo, error) {
	// Get the list of backup paths from the data dir
	backups, err := b.dataDir.BackupList()
	if err != nil {
		return nil, err
	}
	// Build backup info for each backup
	var backupsInfo []BackupInfo
	for _, b := range backups {
		size, err := b.Size()
		if err != nil {
			return nil, err
		}
		backupsInfo = append(backupsInfo, BackupInfo{
			Instance:  b.InstanceId(),
			Timestamp: b.Timestamp(),
			SizeBytes: size,
		})
	}
	return backupsInfo, nil
}

func (b *BackupManager) backupInstanceData(instanceId string, backup *data.Backup) (err error) {
	log.Info("Backing up instance data...")
	backupPath := backup.Path()
	tarFile, err := b.fs.OpenFile(backupPath, os.O_RDWR, 0o644)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	err = utils.TarPrepareToAppend(tarFile)
	if err != nil {
		return err
	}
	instancePath, err := b.dataDir.InstancePath(instanceId)
	if err != nil {
		return err
	}
	return utils.TarAddDir(instancePath, filepath.Join("data"), tarFile)
}

func (b *BackupManager) backupInstanceServiceVolumes(service types.ServiceConfig, backup *data.Backup) (err error) {
	if len(service.Volumes) == 0 {
		return nil
	}
	log.Infof("Backing up %d volumes from service \"%s\"...", len(service.Volumes), service.Name)
	volumes := make([]string, 0, len(service.Volumes))
	for _, v := range service.Volumes {
		volumes = append(volumes, v.Target)
	}
	config := backupConfig{
		Prefix:  snapshotterConfigPrefix(service.Name),
		Volumes: volumes,
	}
	f, err := afero.TempFile(b.fs, os.TempDir(), "eigenlayer-backup-config-*.yaml")
	if err != nil {
		return err
	}
	defer f.Close()
	err = config.Save(f)
	if err != nil {
		return err
	}
	err = b.dockerMgr.Run("eigenlayer-snapshotter", docker.RunOptions{
		AutoRemove: true,
		Mounts: []docker.Mount{
			{
				Type:   docker.VolumeTypeBind,
				Source: f.Name(),
				Target: "/snapshotter.yml",
			},
			{
				Type:   docker.VolumeTypeBind,
				Source: backup.Path(),
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
