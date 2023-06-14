package data

import (
	"fmt"
	"os"
	"path/filepath"
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

// NewDataDirDefault creates a new DataDir instance with the default path as root,
// which is the .eigen folder on the user's home directory.
func NewDataDirDefault() (*DataDir, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return NewDataDir(filepath.Join(homeDir, ".eigen"))
}

// Instance returns the instance with the given id.
func (d *DataDir) Instance(instanceId string) (*Instance, error) {
	instancePath := filepath.Join(d.path, "nodes", instanceId)
	return NewInstance(instancePath)
}

// AddInstance adds a new instance to the data directory.
func (d *DataDir) AddInstance(name, url, version, profile, tag string) (*Instance, error) {
	instanceDirName := name + "-" + tag
	ok, err := d.instanceDirExist(instanceDirName)
	if err != nil {
		return nil, err
	}
	if ok {
		return nil, fmt.Errorf("%w: %s", ErrInstanceAlreadyExists, instanceDirName)
	}

	err = os.MkdirAll(filepath.Join(d.path, "nodes", instanceDirName), 0o755)
	if err != nil {
		return nil, err
	}

	instance := Instance{
		Name:    name,
		URL:     url,
		Version: version,
		Profile: profile,
		Tag:     tag,
	}
	return &instance, instance.Init(filepath.Join(d.path, "nodes", instanceDirName))
}

// RemoveInstance removes the instance with the given id.
func (d *DataDir) RemoveInstance(instanceId string) error {
	instancePath := filepath.Join(d.path, "nodes", instanceId)
	instanceDir, err := os.Stat(instancePath)
	if err != nil {
		return err
	}
	if !instanceDir.IsDir() {
		return fmt.Errorf("%s is not a directory", instanceId)
	}
	return os.RemoveAll(instancePath)
}

func (d *DataDir) instanceDirExist(instanceId string) (bool, error) {
	instancePath := filepath.Join(d.path, "nodes", instanceId)
	instanceDir, err := os.Stat(instancePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !instanceDir.IsDir() {
		return false, fmt.Errorf("%s is not a directory", instanceId)
	}
	return true, nil
}
