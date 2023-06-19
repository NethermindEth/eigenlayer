package data

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NethermindEth/egn/internal/package_handler"
)

// DataDir is the directory where all the data is stored.
type DataDir struct {
	path string
}

// NewDataDir creates a new DataDir instance with the given path as root.
func NewDataDir(path string) (*DataDir, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	return &DataDir{path: absPath}, nil
}

// NewDataDirDefault creates a new DataDir instance with the default path as root.
// Default path is $XDG_DATA_HOME/.eigen or $HOME/.local/share/.eigen if $XDG_DATA_HOME is not set
// as defined in the XDG Base Directory Specification
func NewDataDirDefault() (*DataDir, error) {
	userDataHome := os.Getenv("XDG_DATA_HOME")
	if userDataHome == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return nil, err
		}
		userDataHome = filepath.Join(userHome, ".local", "share")
	}
	dataDir := filepath.Join(userDataHome, ".eigen")
	err := os.MkdirAll(dataDir, 0o755)
	if err != nil {
		return nil, err
	}
	return NewDataDir(dataDir)
}

// Instance returns the instance with the given id.
func (d *DataDir) Instance(instanceId string) (*Instance, error) {
	instancePath := filepath.Join(d.path, "nodes", instanceId)
	return NewInstance(instancePath)
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
	instancePath := filepath.Join(d.path, "nodes", instance.Id())
	_, err := os.Stat(instancePath)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(instancePath, 0o755)
		if err != nil {
			return err
		}
		return instance.Init(instancePath)
	}
	if err != nil {
		return err
	}
	return fmt.Errorf("%w: %s", ErrInstanceAlreadyExists, instance.Id())
}

// HasInstance returns true if an instance with the given id already exists in the
// data dir.
func (d *DataDir) HasInstance(instanceId string) bool {
	instancePath := filepath.Join(d.path, "nodes", instanceId)
	_, err := os.Stat(instancePath)
	return err == nil
}

// RemoveInstance removes the instance with the given id.
func (d *DataDir) RemoveInstance(instanceId string) error {
	instancePath := filepath.Join(d.path, "nodes", instanceId)
	instanceDir, err := os.Stat(instancePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrInstanceNotFound, instanceId)
		}
		return err
	}
	if !instanceDir.IsDir() {
		return fmt.Errorf("%s is not a directory", instanceId)
	}
	return os.RemoveAll(instancePath)
}
