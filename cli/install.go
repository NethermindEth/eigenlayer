package cli

import (
	"fmt"
	"os"

	"github.com/NethermindEth/eigen-wiz/internal/commands"
	"github.com/NethermindEth/eigen-wiz/internal/data"
	"github.com/NethermindEth/eigen-wiz/internal/package_handler"
	"github.com/NethermindEth/eigen-wiz/internal/prompter"
	"github.com/NethermindEth/eigen-wiz/pkg/daemon"
	"github.com/spf13/cobra"
)

func InstallCmd(d daemon.Daemon) *cobra.Command {
	var (
		url     string
		version string
		tag     string
	)
	cmd := cobra.Command{
		Use:   "install [URL]",
		Short: "Install AVS node software from a git repository",
		Long:  "Installs the AVS node software, downloading it from a git repository. You will need to provide the repository URL as a unique argument, which must be an HTTP or HTTPS URL. Use the --version flag if you need to specify a version.",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			url = args[0]
			return validatePkgURL(url)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			destDir, err := os.MkdirTemp(os.TempDir(), "egn-install")
			if err != nil {
				return err
			}
			fmt.Println("Pulling package...") // TODO: this is for debugging purpose, remove it later when we have a logger
			pullResponse, err := d.Pull(&daemon.PullOptions{
				URL:     url,
				Version: version,
				DestDir: destDir,
			})
			if err != nil {
				return err
			}
			fmt.Println("Package pulled successfully.") // TODO: this is for debugging purpose, remove it later when we have a logger
			var fillResult *fillOptionsResult
			for {
				fillResult, err = fillOptions(destDir, pullResponse.Profiles)
				if err != nil {
					return err
				}
				// TODO: those are for debugging purpose, remove them later when we have a logger
				fmt.Println("Profile: " + fillResult.profile)
				fmt.Println("ENV variables to be set:")
				for k, v := range fillResult.envVariables {
					fmt.Printf("  %s=%s\n", k, v)
				}
				ok, err := prompter.Confirm("Do you want to continue with the selected options?")
				if err != nil {
					return err
				}
				if ok {
					break
				}
			}
			dataDir, err := data.NewDataDirDefault()
			if err != nil {
				return err
			}
			instance, err := dataDir.AddInstance(data.AddInstanceOptions{
				Profile:        fillResult.profile,
				URL:            url,
				Version:        pullResponse.CurrentVersion,
				Tag:            tag,
				PackageHandler: package_handler.NewPackageHandler(destDir),
				Env:            fillResult.envVariables,
			})
			if err != nil {
				return err
			}
			composeRunner := commands.NewDockerComposeRunner()
			return composeRunner.Up(instance.ComposePath(), commands.DockerComposeRunnerOptions{
				Out: os.Stdout,
			})
		},
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "version to install. If not specified, the latest version will be installed.")
	cmd.Flags().StringVarP(&tag, "tag", "t", "default", "tag to use for the new instance name")
	return &cmd
}

type fillOptionsResult struct {
	profile      string
	envVariables map[string]string
}

func fillOptions(pullDir string, profiles map[string][]daemon.Option) (*fillOptionsResult, error) {
	profileNames := make([]string, 0, len(profiles))
	for k := range profiles {
		profileNames = append(profileNames, k)
	}
	selectedProfile, err := prompter.SelectProfile(profileNames)
	if err != nil {
		return nil, err
	}
	profileOptions := profiles[selectedProfile]
	for _, option := range profileOptions {
		_, err := prompter.InputString(option.Name(), option.Default(), option.Help(), func(s string) error {
			return option.Set(s)
		})
		if err != nil {
			return nil, err
		}
	}

	pkgHandler := package_handler.NewPackageHandler(pullDir)
	pkgDotEnv, err := pkgHandler.DotEnv(selectedProfile)
	if err != nil {
		return nil, err
	}

	for _, option := range profileOptions {
		pkgDotEnv[option.Target()] = option.Value()
	}

	return &fillOptionsResult{
		profile:      selectedProfile,
		envVariables: pkgDotEnv,
	}, nil
}
