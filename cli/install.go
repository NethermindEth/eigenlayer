package cli

import (
	"fmt"

	"github.com/NethermindEth/egn/cli/prompter"
	"github.com/NethermindEth/egn/pkg/daemon"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

func InstallCmd(d daemon.Daemon, p prompter.Prompter) *cobra.Command {
	var (
		url      string
		version  string
		profile  string
		tag      string
		noPrompt bool
		help     bool
	)
	cmd := cobra.Command{
		Use:   "install [URL] [flags]",
		Short: "Install AVS node software from a git repository",
		Long: `
Installs the AVS node software by downloading it from a git repository. The 
repository URL is required as the unique argument, which must be an HTTP or 
HTTPS URL. Use the --version flag if you need to specify a version.

To preselect a profile, use the --profile flag and the CLI will not prompt you
to select a profile, meaning that the correct profile selection is the user's
responsibility in this case.

To ensure each instance of the node software is uniquely identified, use the
--tag flag to create an unique id which is in the format of 
<repository-name>-<tag>. If the tag is not specified, the "default" tag will be 
used.

Profile options can be specified using the --option.<option-name> flag. The
options are dynamic and depend on the profile selected. If the profile is not
specified, the CLI will prompt you to select a profile. It is responsibility of
the user to know which options are available for each profile.
`,
		DisableFlagParsing: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Parse static flags
			cmd.DisableFlagParsing = false
			cmd.FParseErrWhitelist.UnknownFlags = true // Don't show error for unknown flags to allow dynamic flags
			err := cmd.ParseFlags(args)
			if err != nil {
				return err
			}

			// Skip execution if help flag is set
			help, err = cmd.Flags().GetBool("help")
			if err != nil {
				return err
			}
			if help {
				return nil
			}

			// Validate args
			args = cmd.Flags().Args()
			if len(args) != 1 {
				return fmt.Errorf("%w: accepts 1 arg, received %d", ErrInvalidNumberOfArgs, len(args))
			}
			url = args[0]
			return validatePkgURL(url)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Run help if help flag is set
			if help {
				return cmd.Help()
			}

			// Pull the package
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
			if !noPrompt && profile == "" {
				profileNames := make([]string, 0, len(pullResult.Options))
				for profileName := range pullResult.Options {
					profileNames = append(profileNames, profileName)
				}
				profile, err = p.Select("Select a profile", profileNames)
				if err != nil {
					return err
				}
			}

			profileOptions, ok := pullResult.Options[profile]
			if !ok {
				return fmt.Errorf("profile %s not found", profile)
			}

			// Build dynamic flags with the profile options
			for _, o := range profileOptions {
				cmd.Flags().String("option."+o.Name(), o.Default(), o.Help())
			}

			// Parse dynamic flags
			cmd.FParseErrWhitelist.UnknownFlags = false
			if err = cmd.ParseFlags(args); err != nil {
				return err
			}

			// Fill options
			for _, o := range profileOptions {
				flagValue, err := cmd.Flags().GetString("option." + o.Name())
				if err != nil {
					return err
				}
				if noPrompt {
					if flagValue == "" {
						return fmt.Errorf("%w: %s", ErrOptionWithoutDefault, o.Name())
					}
					if err = o.Set(flagValue); err != nil {
						return err
					}
				} else {
					_, err := p.InputString(o.Name(), o.Default(), o.Help(), func(s string) error {
						return o.Set(s)
					})
					if err != nil {
						return err
					}
				}
			}

			instanceId, err := d.Install(daemon.InstallOptions{
				URL:            url,
				Version:        version,
				Tag:            tag,
				Profile:        profile,
				Options:        profileOptions,
				PackageHandler: pullResult.PackageHandler,
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
	cmd.Flags().BoolVar(&noPrompt, "no-prompt", false, "disable command prompts, and all options should be passed using command flags.")
	return &cmd
}
