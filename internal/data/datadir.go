package data

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NethermindEth/eigenlayer/internal/locker"
	"github.com/NethermindEth/eigenlayer/internal/package_handler"
	"github.com/spf13/afero"
)

const (
	instancesDir = "nodes"
	tempDir      = "temp"
)

const monitoringStackDirName = "monitoring"

// DataDir is the directory where all the data is stored.
type DataDir struct {
	path   string
	fs     afero.Fs
	locker locker.Locker
}

// NewDataDir creates a new DataDir instance with the given path as root.
func NewDataDir(path string, fs afero.Fs, locker locker.Locker) (*DataDir, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return &DataDir{path: absPath, fs: fs, locker: locker}, nil
}

// NewDataDirDefault creates a new DataDir instance with the default path as root.
// Default path is $XDG_DATA_HOME/.eigen or $HOME/.local/share/.eigen if $XDG_DATA_HOME is not set
// as defined in the XDG Base Directory Specification
func NewDataDirDefault(fs afero.Fs, locker locker.Locker) (*DataDir, error) {
	userDataHome := os.Getenv("XDG_DATA_HOME")
	if userDataHome == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		userDataHome = filepath.Join(userHome, ".local", "share")
	}
	dataDir := filepath.Join(userDataHome, ".eigen")
	err := fs.MkdirAll(dataDir, 0o755)
	if err != nil {
		return nil, err
	}

	return NewDataDir(dataDir, fs, locker)
}

// Instance returns the instance with the given id.
func (d *DataDir) Instance(instanceId string) (*Instance, error) {
	instancePath := filepath.Join(d.path, instancesDir, instanceId)
	return newInstance(instancePath, d.fs, d.locker)
}

type AddInstanceOptions struct {
	URL            string
	Version        string
	Profile        string
	Tag            string
	PackageHandler *package_handler.PackageHandler
	Env            map[string]string
}

// InitInstance initializes a new instance. If an instance with the same id already
// exists, an error is returned.
func (d *DataDir) InitInstance(instance *Instance) error {
	instancePath := filepath.Join(d.path, instancesDir, InstanceId(instance.Name, instance.Tag))
	_, err := d.fs.Stat(instancePath)
	if err != nil && os.IsNotExist(err) {
		return instance.init(instancePath, d.fs, d.locker)
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("%w: %s", ErrInstanceAlreadyExists, InstanceId(instance.Name, instance.Tag))
}

// HasInstance returns true if an instance with the given id already exists in the
// data dir.
func (d *DataDir) HasInstance(instanceId string) bool {
	instancePath := filepath.Join(d.path, instancesDir, instanceId)
	_, err := d.fs.Stat(instancePath)
	return err == nil
}

// InstancePath return the path to the directory of the instance with the given id.
func (d *DataDir) InstancePath(instanceId string) (string, error) {
	instancePath := filepath.Join(d.path, instancesDir, instanceId)
	_, err := d.fs.Stat(instancePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrInstanceNotFound
		}
		return "", err
	}
	return instancePath, nil
}

// RemoveInstance removes the instance with the given id.
func (d *DataDir) RemoveInstance(instanceId string) error {
	instancePath := filepath.Join(d.path, instancesDir, instanceId)
	instanceDir, err := d.fs.Stat(instancePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrInstanceNotFound, instanceId)
		}
		return err
	}
	if !instanceDir.IsDir() {
		return fmt.Errorf("%s is not a directory", instanceId)
	}
	return d.fs.RemoveAll(instancePath)
}

// InitTemp creates a new temporary directory for the given id. If already exists,
// an error is returned.
func (d *DataDir) InitTemp(id string) (string, error) {
	tempPath := filepath.Join(d.path, tempDir, id)
	_, err := d.fs.Stat(tempPath)
	if err != nil {
		if os.IsNotExist(err) {
			return tempPath, d.fs.MkdirAll(tempPath, 0o755)
		}
		return "", err
	}
	return "", ErrTempDirAlreadyExists
}

// RemoveTemp removes the temporary directory with the given id.
func (d *DataDir) RemoveTemp(id string) error {
	return d.fs.RemoveAll(filepath.Join(d.path, tempDir, id))
}

// TempPath returns the path to the temporary directory with the given id.
func (d *DataDir) TempPath(id string) (string, error) {
	tempPath := filepath.Join(d.path, tempDir, id)
	tempStat, err := d.fs.Stat(tempPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrTempDirDoesNotExist
		}
		return "", err
	}
	if !tempStat.IsDir() {
		return "", ErrTempIsNotDir
	}
	return tempPath, nil
}

// MonitoringStack checks if a monitoring stack directory exists in the data directory.
// If the directory does not exist, it creates it and initializes a new MonitoringStack instance.
// If the directory exists, it simply returns a new MonitoringStack instance.
// It returns an error if there is any issue accessing or creating the directory, or initializing the MonitoringStack.
func (d *DataDir) MonitoringStack() (*MonitoringStack, error) {
	monitoringStackPath := filepath.Join(d.path, monitoringStackDirName)
	_, err := d.fs.Stat(monitoringStackPath)
	if os.IsNotExist(err) {
		if err = d.fs.MkdirAll(monitoringStackPath, 0o755); err != nil {
			return nil, err
		}

		monitoringStack := &MonitoringStack{path: monitoringStackPath, fs: d.fs, l: d.locker}
		if err = monitoringStack.Init(); err != nil {
			return nil, err
		}
		return monitoringStack, nil
	} else if err != nil {
		return nil, err
	}

	return newMonitoringStack(monitoringStackPath, d.fs, d.locker), nil
}

// RemoveMonitoringStack removes the monitoring stack directory from the data directory.
// It returns an error if there is any issue accessing or removing the directory.
func (d *DataDir) RemoveMonitoringStack() error {
	monitoringStackPath := filepath.Join(d.path, monitoringStackDirName)
	_, err := d.fs.Stat(monitoringStackPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("%w: %s", ErrMonitoringStackNotFound, monitoringStackPath)
	} else if err != nil {
		return err
	}

	return d.fs.RemoveAll(monitoringStackPath)
}
