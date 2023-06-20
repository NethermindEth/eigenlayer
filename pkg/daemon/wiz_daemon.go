package daemon

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/NethermindEth/egn/internal/data"
	"github.com/NethermindEth/egn/pkg/pull"
	"github.com/NethermindEth/egn/pkg/run"
)

// Checks that WizDaemon implements Daemon.
var _ = Daemon(&WizDaemon{})

// WizDaemon is the main entrypoint for all the functionalities of the daemon.
type WizDaemon struct {
	dataDir *data.DataDir
}

// NewDaemon create a new daemon instance.
func NewWizDaemon() (*WizDaemon, error) {
	dataDir, err := data.NewDataDirDefault()
	if err != nil {
		return nil, err
	}
	return &WizDaemon{
		dataDir: dataDir,
	}, nil
}

// Install installs a node software package using the provided options. If the instance
// already exists, it returns an error.
func (d *WizDaemon) Install(options InstallOptions) error {
	instanceName, err := instanceNameFromURL(options.URL)
	if err != nil {
		return err
	}
	instance := &data.Instance{
		Name: instanceName,
		URL:  options.URL,
		Tag:  options.Tag,
	}

	// Check if instance already exists
	if d.dataDir.HasInstance(instance.Id()) {
		return fmt.Errorf("%w: %s", ErrInstanceAlreadyExists, instance.Id())
	}

	// Pull package to a temporary directory
	destDir, err := os.MkdirTemp(os.TempDir(), "egn-install")
	if err != nil {
		return err
	}
	// Pull package
	pkgHandler, err := pull.Pull(options.URL, options.Version, destDir)
	if err != nil {
		return err
	}
	// Get profiles names and its options
	pkgProfiles, err := pkgHandler.Profiles()
	profiles := make(map[string][]Option, len(pkgProfiles))
	profileNames := make([]string, 0, len(pkgProfiles))
	for _, pkgProfile := range pkgProfiles {
		options := make([]Option, len(pkgProfile.Options))
		for i, o := range pkgProfile.Options {
			switch o.Type {
			case "str":
				options[i] = NewOptionString(o)
			case "int":
				options[i], err = NewOptionInt(o)
			case "float":
				options[i], err = NewOptionFloat(o)
			case "bool":
				options[i], err = NewOptionBool(o)
			case "path_dir":
				options[i] = NewOptionPathDir(o)
			case "path_file":
				options[i] = NewOptionPathFile(o)
			case "uri":
				options[i] = NewOptionURI(o)
			case "select":
				options[i] = NewOptionSelect(o)
			case "port":
				options[i], err = NewOptionPort(o)
			case "id":
				options[i] = NewOptionID(o)
			default:
				return errors.New("unknown option type: " + o.Type)
			}
		}
		if err != nil {
			return err
		}
		profiles[pkgProfile.Name] = options
		profileNames = append(profileNames, pkgProfile.Name)
	}
	// Select profile
	instance.Profile, err = options.ProfileSelector(profileNames)
	if err != nil {
		return err
	}
	// Fill profile options
	filledOptions, err := options.OptionsFiller(profiles[instance.Profile])
	if err != nil {
		return err
	}
	// Install package
	env := make(map[string]string, len(filledOptions))
	for _, o := range filledOptions {
		env[o.Target()] = o.Value()
	}
	instance.Version, err = pkgHandler.CurrentVersion()
	if err != nil {
		return err
	}
	err = d.dataDir.InitInstance(instance)
	if err != nil {
		return err
	}
	err = instance.Setup(env, pkgHandler.ProfileFS(instance.Profile))
	if err != nil {
		return err
	}
	doRun, err := options.RunConfirmation()
	if err != nil {
		return err
	}
	if doRun {
		return run.Run(d.dataDir, instance.Id())
	}
	return nil
}

func instanceNameFromURL(u string) (string, error) {
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return "", err
	}
	return path.Base(parsedURL.Path), nil
}
