package daemon

import (
	"errors"
	"net/url"
	"os"
	"path"

	"github.com/NethermindEth/eigen-wiz/internal/data"
	"github.com/NethermindEth/eigen-wiz/pkg/pull"
	"github.com/NethermindEth/eigen-wiz/pkg/run"
)

// Checks that WizDaemon implements Daemon.
var _ = Daemon(&WizDaemon{})

// WizDaemon is the main entrypoint for all the functionalities of the daemon.
type WizDaemon struct{}

// NewDaemon create a new daemon instance.
func NewWizDaemon() *WizDaemon {
	return &WizDaemon{}
}

// InstallOptions is a set of options for installing a node software package.
type InstallOptions struct {
	URL             string
	Version         string
	Tag             string
	ProfileSelector func(profiles []string) (string, error)
	OptionsFiller   func(opts []Option) ([]Option, error)
	RunConfirmation func() (bool, error)
}

// Install installs a node software package using the provided options.
func (d *WizDaemon) Install(options InstallOptions) error {
	// Check if instance already exists
	dataDir, err := data.NewDataDirDefault()
	if err != nil {
		return err
	}

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
	selectedProfile, err := options.ProfileSelector(profileNames)
	if err != nil {
		return err
	}
	// Fill profile options
	filledOptions, err := options.OptionsFiller(profiles[selectedProfile])
	if err != nil {
		return err
	}
	// Install package
	env := make(map[string]string, len(filledOptions))
	for _, o := range filledOptions {
		env[o.Target()] = o.Value()
	}
	version, err := pkgHandler.CurrentVersion()
	if err != nil {
		return err
	}
	instanceName, err := instanceNameFromURL(options.URL)
	if err != nil {
		return err
	}
	instance := &data.Instance{
		Name:    instanceName,
		Profile: selectedProfile,
		URL:     options.URL,
		Version: version,
		Tag:     options.Tag,
	}
	err = dataDir.InitInstance(instance)
	if err != nil {
		return err
	}
	err = instance.Setup(env, pkgHandler.ProfileFS(selectedProfile))
	if err != nil {
		return err
	}
	doRun, err := options.RunConfirmation()
	if err != nil {
		return err
	}
	if doRun {
		return run.Run(dataDir, instance.Id())
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
