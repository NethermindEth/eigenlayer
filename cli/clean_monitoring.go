package cli

import (
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/spf13/cobra"
)

func CleanMonitoringCmd(d daemon.Daemon) *cobra.Command {
	cmd := cobra.Command{
		Use:   "clean-monitoring",
		Short: "Stop and uninstall the monitoring stack",
		Long:  "Stop and uninstall the monitoring stack. If the monitoring stack is already uninstalled, the command won't do anything.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.CleanMonitoring()
		},
	}
	return &cmd
}
