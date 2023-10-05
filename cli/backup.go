package cli

import (
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func BackupCmd(d daemon.Daemon) *cobra.Command {
	var instanceId string
	cmd := cobra.Command{
		Use:   "backup <instance-id>",
		Short: "Backup an instance",
		Long:  "Backup an instance saving the data into a tarball file. To list backups, use 'eigenlayer backup ls'",
		Args:  cobra.MinimumNArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			instanceId = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			backupId, err := d.Backup(instanceId)
			if err != nil {
				return err
			}
			log.Info("Backup created with id: ", backupId)
			return nil
		},
	}

	// Add ls subcommand
	lsCmd := BackupLsCmd(d)
	cmd.AddCommand(lsCmd)

	return &cmd
}
