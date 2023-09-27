package cli

import (
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/spf13/cobra"
)

func StopCmd(d daemon.Daemon) *cobra.Command {
	var instanceId string
	cmd := cobra.Command{
		Use:   "stop <instance_id>",
		Short: "Stop an AVS node instance",
		Long:  "Stops an AVS node instance. The instance ID is required as the unique argument. instance_id is required as the unique argument, and it is the combination of the instance repository name and the instance tag computed during the installation, like this: <repository-name>-<tag>.",
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
