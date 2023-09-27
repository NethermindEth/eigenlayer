package cli

import (
	"context"
	"os"

	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/spf13/cobra"
)

func LogsCmd(d daemon.Daemon) *cobra.Command {
	var (
		instanceID string
		follow     bool
		since      string
		until      string
		timestamps bool
		tail       string
	)

	cmd := cobra.Command{
		Use:   "logs <instance_id>",
		Short: "Show AVS node logs",
		Long:  "Show AVS node logs, which are the logs of all the services running in the node.",
		Args:  cobra.ExactArgs(1),
		PreRun: func(cmd *cobra.Command, args []string) {
			instanceID = args[0]
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.NodeLogs(context.Background(), os.Stdout, instanceID, daemon.NodeLogsOptions{
				Follow:     follow,
				Since:      since,
				Until:      until,
				Timestamps: timestamps,
				Tail:       tail,
			})
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "Follow log output")
	cmd.Flags().StringVar(&since, "since", "", "Show logs since timestamp (e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)")
	cmd.Flags().StringVar(&until, "until", "", "Show logs before a timestamp (e.g. 2013-01-02T13:23:37Z) or relative (e.g. 42m for 42 minutes)")
	cmd.Flags().BoolVarP(&timestamps, "timestamps", "t", false, "Show timestamps")
	cmd.Flags().StringVarP(&tail, "tail", "n", "all", "Number of lines to show from the end of the logs")
	return &cmd
}
