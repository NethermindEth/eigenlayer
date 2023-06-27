package cli

import (
	"github.com/NethermindEth/egn/pkg/daemon"
	"github.com/spf13/cobra"
)

func StopCmd(d daemon.Daemon) *cobra.Command {
	var instanceId string
	cmd := cobra.Command{
		Use:   "stop [INSTANCE_ID]",
		Short: "Stop an AVS node instance",
		Long:  "Stops an AVS node instance. The instance ID is required as the unique argument.",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			instanceId = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.Stop(instanceId)
		},
	}
	return &cmd
}
