package cli

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/NethermindEth/egn/cli/prompter"
	"github.com/NethermindEth/egn/pkg/daemon"
	"github.com/spf13/cobra"
)

func InstallCmd(d daemon.Daemon, p prompter.Prompter) *cobra.Command {
	var (
		url     string
		version string
		profile string
		tag     string
	)
	cmd := cobra.Command{
		Use:   "install [URL]",
		Short: "Install AVS node software from a git repository",
		Long: `Installs the AVS node software by downloading it from a git repository. The repository URL is required as the unique argument, which must be an HTTP or HTTPS URL. Use the --version flag if you need to specify a version.
To preselect a profile, use the --profile flag and the CLI will not prompt you to select a profile, meaning that the correct profile selection is the user's responsibility in this case.
To ensure each instance of the node software is uniquely identified, use the --tag flag to create an unique id which is in the format of <repository-name>-<tag>. If the tag is not specified, the "default" tag will be used.`,
		Args: cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			url = args[0]
			return validatePkgURL(url)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			pullResult, err := d.Pull(url, version, true)
			if err != nil {
				return err
			}
			// Select version if not specified
			if version == "" {
				log.Println("Version not specified, using latest.")
			}
			version = pullResult.Version
			log.Printf("Using version %s\n", version)
			// Select profile if not specified
			if profile == "" {
				profileNames := make([]string, 0, len(pullResult.Options))
				for profileName := range pullResult.Options {
					profileNames = append(profileNames, profileName)
				}
				profile, err = p.Select("Select a profile", profileNames)
				if err != nil {
					return err
				}
			}
			// Fill options
			profileOptions, ok := pullResult.Options[profile]
			if !ok {
				return fmt.Errorf("profile %s not found", profile)
			}
			for _, o := range profileOptions {
				_, err := p.InputString(o.Name(), o.Default(), o.Help(), func(s string) error {
					return o.Set(s)
				})
				if err != nil {
					return err
				}
			}
			instanceId, err := d.Install(daemon.InstallOptions{
				URL:     url,
				Version: version,
				Tag:     tag,
				Profile: profile,
				Options: profileOptions,
			})
			if err != nil {
				return err
			}

			log.Info("Installed successfully with instance id: ", instanceId)

			if pullResult.HasPlugin {
				// TODO: improve this message with the command to run the plugin
				log.Info("The installed node software has a plugin.")
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "version to install. If not specified the latest version will be installed.")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "profile to use for the new instance name. If not specified a list of available profiles will be shown to select from.")
	cmd.Flags().StringVarP(&tag, "tag", "t", "default", "tag to use for the new instance name.")
	return &cmd
}
