package cli

import (
	"fmt"
	"io"
	"slices"
	"text/tabwriter"
	"time"

	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/spf13/cobra"
	"kythe.io/kythe/go/util/datasize"
)

func BackupLsCmd(d daemon.Daemon) *cobra.Command {
	cmd := cobra.Command{
		Use:   "ls",
		Short: "List backups",
		Long:  "List backups showing all backups and their details.",
		RunE: func(cmd *cobra.Command, args []string) error {
			backups, err := d.BackupList()
			if err != nil {
				return err
			}
			sortBackupsByTimestamp(backups)
			printBackupTable(backups, cmd.OutOrStdout())
			return nil
		},
	}
	return &cmd
}

func sortBackupsByTimestamp(backups []daemon.BackupInfo) {
	slices.SortFunc(backups, func(a, b daemon.BackupInfo) int {
		if a.Timestamp.After(b.Timestamp) {
			return -1
		} else if a.Timestamp.Before(b.Timestamp) {
			return 1
		} else {
			return 0
		}
	})
}

func printBackupTable(backups []daemon.BackupInfo, out io.Writer) {
	w := tabwriter.NewWriter(out, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "ID\tAVS Instance ID\tVERSION\tCOMMIT\tTIMESTAMP\tSIZE\tURL\t")
	for _, b := range backups {
		fmt.Fprintln(w, backupTableItem{
			id:        b.Id,
			instance:  b.Instance,
			timestamp: b.Timestamp.Format(time.DateTime),
			size:      datasize.Size(b.SizeBytes).String(),
			version:   b.Version,
			commit:    b.Commit,
			url:       b.Url,
		})
	}
	w.Flush()
}

type backupTableItem struct {
	id        string
	instance  string
	timestamp string
	size      string
	version   string
	commit    string
	url       string
}

// func minifiedId(id string) string {
// 	if len(id) > 8 {
// 		return id[:8]
// 	}
// 	return id
// }

func (b backupTableItem) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t", b.id, b.instance, b.version, b.commit, b.timestamp, b.size, b.url)
}
