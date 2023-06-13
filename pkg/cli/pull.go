package cli

import (
	"fmt"
	"os"

	"github.com/NethermindEth/eigen-wiz/pkg/daemon"
	"github.com/spf13/cobra"
)

func PullCmd(d daemon.Daemon) *cobra.Command {
	var (
		url     string
		version string
	)
	cmd := &cobra.Command{
		Use:   "pull [url]",
		Short: "Pull a package from a remote repository",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			url = args[0]
			return validatePkgURL(url)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			dest, err := os.MkdirTemp(os.TempDir(), "eigen-wiz") // TODO: change this to the proper directory inside the data directory of the daemon
			if err != nil {
				return err
			}
			pullResponse, err := d.Pull(&daemon.PullOptions{
				URL:     url,
				Version: version,
				DestDir: dest,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Pulled version: %s\n", pullResponse.CurrentVersion)
			fmt.Printf("Latest version: %s\n", pullResponse.LatestVersion)
			fmt.Println("Available profiles for pulled version:")
			for _, p := range pullResponse.Profiles {
				fmt.Printf("  - %s\n", p.Name)
			}
			return nil
		},
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "version to pull. ")

	return cmd
}
