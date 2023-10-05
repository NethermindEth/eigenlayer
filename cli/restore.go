package cli

import (
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/spf13/cobra"
)

func RestoreCmd(d daemon.Daemon) *cobra.Command {
	var (
		backupId string
		run      bool
	)
	cmd := cobra.Command{
		Use:   "restore <backup-id>",
		Short: "Restore an instance from a backup",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			backupId = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.Restore(backupId, run)
		},
	}

	cmd.Flags().BoolVarP(&run, "run", "r", false, "Run the instance after restoring it")
	return &cmd
}
