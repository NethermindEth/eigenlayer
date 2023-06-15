package data

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NethermindEth/eigen-wiz/internal/package_handler"
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

// AddInstance adds a new instance to the data directory.
func (d *DataDir) AddInstance(opts AddInstanceOptions) (*Instance, error) {
	splits := strings.Split(opts.URL, "/")
	name := splits[len(splits)-1]
	instanceDirName := name + "-" + opts.Tag
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
		URL:     opts.URL,
		Version: opts.Version,
		Profile: opts.Profile,
		Tag:     opts.Tag,
	}
	err = instance.Init(filepath.Join(d.path, "nodes", instanceDirName), opts.Env, opts.PackageHandler.ProfileFS(instance.Profile))
	if err != nil {
		return nil, err
	}

	return &instance, nil
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
