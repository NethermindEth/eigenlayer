package cli

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/spf13/cobra"
)

func BackupLsCmd(d daemon.Daemon) *cobra.Command {
	cmd := cobra.Command{
		Use:   "ls",
		Short: "List backups",
		RunE: func(cmd *cobra.Command, args []string) error {
			backus, err := d.BackupList()
			if err != nil {
				return err
			}
			printBackupTable(backus, cmd.OutOrStdout())
			return nil
		},
	}
	return &cmd
}

func printBackupTable(backups []daemon.BackupInfo, out io.Writer) {
	w := tabwriter.NewWriter(out, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "AVS Instance ID\tTIMESTAMP\tSIZE (Bytes)\t")
	for _, b := range backups {
		fmt.Fprintln(w, backupTableItem{
			instance:  b.Instance,
			timestamp: b.Timestamp.Format(time.DateTime),
			size:      b.SizeBytes,
		})
	}
	w.Flush()
}

type backupTableItem struct {
	instance  string
	timestamp string
	size      uint64
}

func (b backupTableItem) String() string {
	return fmt.Sprintf("%s\t%s\t%d\t", b.instance, b.timestamp, b.size)
}
