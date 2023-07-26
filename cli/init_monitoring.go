package cli

import (
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/spf13/cobra"
)

func InitMonitoringCmd(d daemon.Daemon) *cobra.Command {
	cmd := cobra.Command{
		Use:   "init-monitoring",
		Short: "Install and run the monitoring stack",
		Long:  "Install and run the monitoring stack. If the monitoring stack is already installed, it will be initialized with its configuration updated.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.InitMonitoring(true, true)
		},
	}
	return &cmd
}
