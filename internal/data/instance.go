package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

// Instance represents the data stored about a node software instance
type Instance struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	Version string `json:"version"`
	Profile string `json:"profile"`
	Tag     string `json:"tag"`
	path    string
	lock    *flock.Flock
}

// NewInstance creates a new instance with the given path as root. It loads the
// state.json file and validates it.
func NewInstance(path string) (*Instance, error) {
	i := Instance{
		path: path,
	}
	stateFile, err := os.Open(filepath.Join(i.path, "state.json"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w %s: state.json not found", ErrInvalidInstanceDir, path)
		}
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
	err = json.Unmarshal(stateData, &i)
	if err != nil {
		return nil, fmt.Errorf("%w %s: invalid state.json file: %s", ErrInvalidInstance, path, err)
	}
	err = i.validate()
	if err != nil {
		return nil, err
	}
	return &i, nil
}

// Init initializes a new instance with the given path as root. It creates the
// .lock and state.json files. If the instance is invalid, an error is returned.
func (i *Instance) Init(instancePath string) error {
	err := i.validate()
	if err != nil {
		return err
	}
	i.path = instancePath
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

// Setup creates the instance directory and copies the profile files into it from
// the given fs.FS. It also creates the .env file with the given environment variables
// on the env map.
func (i *Instance) Setup(env map[string]string, profileFs fs.FS) (err error) {
	err = i.Lock()
	if err != nil {
		return err
	}
	defer func() {
		unlockErr := i.Unlock()
		if err == nil {
			err = unlockErr
		}
	}()
	// Create .env file
	envFile, err := os.Create(filepath.Join(i.path, ".env"))
	if err != nil {
		return err
	}
	for k, v := range env {
		envFile.WriteString(fmt.Sprintf("%s=%s\n", k, v))
	}
	defer envFile.Close()

	// Copy docker-compose.yml
	pkgComposeFile, err := profileFs.Open("docker-compose.yml")
	if err != nil {
		return err
	}
	defer pkgComposeFile.Close()
	composeFile, err := os.Create(filepath.Join(i.path, "docker-compose.yml"))
	if err != nil {
		return err
	}
	defer composeFile.Close()
	if _, err := io.Copy(composeFile, pkgComposeFile); err != nil {
		return err
	}

	// Copy src directory
	return fs.WalkDir(profileFs, "src", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		targetPath := filepath.Join(i.path, path)
		if d.IsDir() {
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
		} else {
			pkgFile, err := profileFs.Open(path)
			if err != nil {
				return err
			}
			defer pkgFile.Close()
			targetFile, err := os.Create(targetPath)
			if err != nil {
				return err
			}
			if _, err := io.Copy(targetFile, pkgFile); err != nil {
				return err
			}
		}
		return nil
	})
}

// Id returns the instance id, which is the name and tag concatenated with a dash
// <name>-<tag>
func (i *Instance) Id() string {
	return fmt.Sprintf("%s-%s", i.Name, i.Tag)
}

// ComposePath returns the path to the docker-compose.yml file of the instance.
func (i *Instance) ComposePath() string {
	return filepath.Join(i.path, "docker-compose.yml")
}

// Lock locks the .lock file of the instance.
func (i *Instance) Lock() error {
	if i.lock == nil {
		i.lock = flock.New(filepath.Join(i.path, ".lock"))
	}
	return i.lock.Lock()
}

// Unlock unlocks the .lock file of the instance.
func (i *Instance) Unlock() error {
	if i.lock == nil || !i.lock.Locked() {
		return errors.New("instance is not locked")
	}
	return i.lock.Unlock()
}

func (i *Instance) validate() error {
	if i.Name == "" {
		return fmt.Errorf("%w: name is empty", ErrInvalidInstance)
	}
	if i.URL == "" {
		return fmt.Errorf("%w: url is empty", ErrInvalidInstance)
	}
	if i.Version == "" {
		return fmt.Errorf("%w: version is empty", ErrInvalidInstance)
	}
	if i.Profile == "" {
		return fmt.Errorf("%w: profile is empty", ErrInvalidInstance)
	}
	if i.Tag == "" {
		return fmt.Errorf("%w: tag is empty", ErrInvalidInstance)
	}
	return nil
}
