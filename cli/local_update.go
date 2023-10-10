package cli

import (
	"fmt"
	"os"

	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/NethermindEth/eigenlayer/internal/utils"
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func LocalUpdateCmd(d daemon.Daemon, p prompter.Prompter) *cobra.Command {
	var (
		instanceId string
		dirPath    string
		noPrompt   bool
		backup     bool
		help       bool
		yes        bool
	)
	cmd := cobra.Command{
		Use:   "local-update [flags] <instance_id> <dir-path>",
		Short: "Update an instance to a new version from a local directory.",
		Long: `
!!! THIS UPDATE METHOD IS INSECURE !!!
!!! USE ONLY FOR DEVELOPMENT PURPOSES !!!

Updates an AVS node software from a new package in a local directory.

Options of the new version can be specified using the --option.<option-name> flag.

To avoid any data loss during the update process, the user can specify the --backup
flag. In this case, the current instance will be backed up before uninstalling it,
and if the update process fails, the instance will be restored.`,
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
			if len(args) != 2 {
				return fmt.Errorf("%w: accepts 2 args, received %d", ErrInvalidNumberOfArgs, len(args))
			}
			instanceId = args[0]
			dirPath = args[1]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if help {
				return cmd.Help()
			}
			// Create temporary tar file
			tarFile, err := os.CreateTemp(os.TempDir(), "eigenlayer-local-update-*.tar.gz")
			if err != nil {
				return err
			}
			defer tarFile.Close()
			defer os.Remove(tarFile.Name())

			// Build tar file
			err = utils.CompressToTarGz(dirPath, tarFile)
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

			// Pull update from tar.gz file
			pullResult, err := d.LocalPullUpdate(instanceId, tarFile)
			if err != nil {
				return err
			}

			// Log useful information about the update
			logVersionChange(pullResult.OldVersion, pullResult.NewVersion)
			logCommitChange(pullResult.OldCommit, pullResult.NewCommit)
			printOptionsTable(pullResult.OldOptions, pullResult.MergedOptions)

			// Build dynamic flags with the profile options
			for _, o := range pullResult.MergedOptions {
				cmd.Flags().String("option."+o.Name(), o.Default(), o.Help())
			}

			// Parse dynamic flags
			cmd.FParseErrWhitelist.UnknownFlags = false
			if err = cmd.ParseFlags(args); err != nil {
				return err
			}

			// Fill necessary options
			for _, o := range pullResult.MergedOptions {
				if noPrompt {
					flagValue, err := cmd.Flags().GetString("option." + o.Name())
					if err != nil {
						return err
					}
					if flagValue == "" {
						return fmt.Errorf("%w: %s", ErrOptionWithoutDefault, o.Name())
					}
					if err = o.Set(flagValue); err != nil {
						return err
					}
				}
				if !o.IsSet() {
					_, err := p.InputString(o.Name(), o.Default(), o.Help(), func(s string) error {
						return o.Set(s)
					})
					if err != nil {
						return err
					}
				}
			}

			// Backup instance
			var backupId string
			if backup {
				backupId, err = d.Backup(instanceId)
				if err != nil {
					return err
				}
				log.Info("Backup created with id: ", backupId)
			}

			// Uninstall current instance
			err = uninstallPackage(d, instanceId)
			if err != nil {
				if backup {
					return abortWithRestore(d, backupId, err)
				}
				return err
			}

			// Build options map
			options := make(map[string]string)
			for _, o := range pullResult.MergedOptions {
				v, err := o.Value()
				if err != nil {
					return err
				}
				options[o.Name()] = v
			}

			// Reset tarFile reader
			_, err = tarFile.Seek(0, 0)
			if err != nil {
				return err
			}

			// Install new instance's version
			newInstanceId, err := d.LocalInstall(tarFile, daemon.LocalInstallOptions{
				Name:    pullResult.Name,
				Tag:     pullResult.Tag,
				Profile: pullResult.Profile,
				Options: options,
			})
			if err != nil {
				if backup {
					return abortWithRestore(d, backupId, err)
				}
				return err
			}
			if newInstanceId != instanceId {
				// NOTE: I think this never happens but it could be useful to check
				// that the instance ID is the same as the one we started with. Also
				// we can manage this case as an error, but I think it's better to
				// just log it for now.
				log.Infof("Instance ID changed: %s -> %s", instanceId, newInstanceId)
			}

			if pullResult.HasPlugin {
				log.Info("The installed node software has a plugin.")
			}

			return runInstance(d, newInstanceId, p, yes, noPrompt)
		},
	}

	cmd.Flags().BoolVar(&noPrompt, "no-prompt", false, "disable command prompts, and all options should be passed using command flags.")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompts.")
	cmd.Flags().BoolVar(&backup, "backup", false, "backup current instance before updating.")
	return &cmd
}
