package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/spf13/cobra"
)

type tableItem struct {
	avs     string
	version string
	commit  string
	running bool
	health  string
	comment string
}

func (i tableItem) String() string {
	return fmt.Sprintf("%s\t%t\t%s\t%s\t%s\t%s\t", i.avs, i.running, i.health, i.version, commitPrefix(i.commit), i.comment)
}

func commitPrefix(commit string) string {
	if len(commit) > 12 {
		return commit[:12]
	}
	return commit
}

func ListCmd(d daemon.Daemon) *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "List all installed AVS nodes and their health status.",
		Long: `List all installed AVS nodes and their health status. If the AVS node is not running the health check will not be
performed. An AVS node is considered running if it is installed and has at least one running service. The health check
is performed by calling the health endpoint of the AVS node, to know more about this endpoint please refer to this
Eigenlayer AVS Specification link https://eigen.nethermind.io/docs/metrics/metrics-api#get-eigennodehealth.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			instances, err := d.ListInstances()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 4, ' ', 0)
			fmt.Fprintln(w, "AVS Instance ID\tRUNNING\tHEALTH\tVERSION\tCOMMIT\tCOMMENT\t")
			for _, instance := range instances {
				fmt.Fprintln(w, tableItem{
					avs:     instance.ID,
					running: instance.Running,
					health:  instance.Health.String(),
					comment: instance.Comment,
					version: instance.Version,
					commit:  instance.Commit,
				})
			}
			w.Flush()

			return nil
		},
	}
}
