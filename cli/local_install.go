package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func LocalInstallCmd(d daemon.Daemon) *cobra.Command {
	var (
		path     string
		profile  string
		name     string
		tag      string
		help     bool
		options  = make(map[string]string)
		logDebug bool
	)
	cmd := cobra.Command{
		Use:   "local-install [path]",
		Short: "Install AVS node software from a local directory",
		Long: `
!!! THIS INSTALLATION METHOD IS INSECURE !!!
!!! USE ONLY FOR DEVELOPMENT PURPOSES !!!

Installs the AVS node software from a local directory. Make sure to select
the correct profile and set its options properly.	

To ensure each instance of the node software is uniquely identified, use the
--tag flag to create an unique id which is in the format of 
<name>-<tag>. If the tag is not specified, the "default" tag will be 
used.

Profile options can be specified using the --option.<option-name> flag.
Flags are the only way to specify options for local installations, and its
responsibility of the user to know which options are available for each
profile.`,
		DisableFlagParsing: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			log.Warn("This command is insecure and should only be used for development purposes")
			if logDebug {
				log.SetLevel(log.DebugLevel)
			}
			// Get option values
			for i := 0; i < len(args); {
				if strings.HasPrefix(args[i], "--option.") {
					if len(args) < i+2 {
						return fmt.Errorf("%w: option %s requires a value", ErrInvalidNumberOfArgs, args[i])
					}
					options[strings.TrimPrefix(args[i], "--option.")] = args[i+1]
					i += 2
				} else {
					i++
				}
			}
			log.Info("Options: ", options)
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
			path, err = filepath.Abs(args[0])
			return err
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Run help if help flag is set
			if help {
				return cmd.Help()
			}

			// Create temporary tar file
			tarFile, err := os.CreateTemp(os.TempDir(), "eigenlayer-local-install-*.tar.gz")
			if err != nil {
				return err
			}

			// Build tar file
			err = buildTar(path, tarFile)
			if err != nil {
				return err
			}
			if err := tarFile.Close(); err != nil {
				return err
			}

			tarFile, err = os.Open(tarFile.Name())
			if err != nil {
				return err
			}
			instanceId, err := d.LocalInstall(tarFile, daemon.LocalInstallOptions{
				Name:    name,
				Tag:     tag,
				Profile: profile,
				Options: options,
			})
			if err != nil {
				return err
			}
			log.Info("Installed successfully with instance id: ", instanceId)
			return nil
		},
	}
	cmd.Flags().BoolVar(&logDebug, "log-debug", false, "enable debug logs")
	cmd.Flags().StringVar(&name, "name", "", "name to use for the new instance name.")
	cmd.Flags().StringVarP(&profile, "profile", "p", "", "profile to use for the new instance name. If not specified, the installation will fail.")
	cmd.Flags().StringVarP(&tag, "tag", "t", "default", "tag to use for the new instance name.")
	return &cmd
}
