package data

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/gofrs/flock"
)

// Instance represents the data stored about a node software instance
type Instance struct {
	Name      string `json:"name"`
	URL       string `json:"url"`
	Version   string `json:"version"`
	Profile   string `json:"profile"`
	Tag       string `json:"tag"`
	path      string
	lock      *flock.Flock
	lockMutex sync.Mutex
}

func NewInstance(path string) (*Instance, error) {
	i := Instance{
		path: path,
	}
	stateFile, err := os.Open(filepath.Join(i.path, "state.json"))
	if err != nil {
		return nil, err
	}
	defer func() {
		closeErr := stateFile.Close()
		if err == nil {
			err = closeErr
		}
	}()

	stateData, err := io.ReadAll(stateFile)
	if err != nil {
		return nil, err
	}
	return &i, json.Unmarshal(stateData, &i)
}

func (i *Instance) Init(instancePath string) (err error) {
	i.path = instancePath
	// TODO: Validate instance properties, like name and tag that are required
	// Create the lock file
	_, err = os.Create(filepath.Join(i.path, ".lock"))
	if err != nil {
		return err
	}
	// Create state file
	stateFile, err := os.Create(filepath.Join(i.path, "state.json"))
	if err != nil {
		return err
	}
	defer func() {
		closeErr := stateFile.Close()
		if err == nil {
			err = closeErr
		}
	}()

	stateData, err := json.Marshal(i)
	if err != nil {
		return err
	}
	_, err = stateFile.Write(stateData)
	return err
}

func (i *Instance) Lock() error {
	i.lockMutex.Lock()
	defer i.lockMutex.Unlock()
	if i.lock == nil {
		i.lock = flock.New(filepath.Join(i.path, ".lock"))
	}
	return i.lock.Lock()
}

func (i *Instance) Unlock() error {
	i.lockMutex.Lock()
	defer i.lockMutex.Unlock()
	if i.lock == nil || !i.lock.Locked() {
		return errors.New("instance is not locked")
	}
	return i.lock.Unlock()
}
