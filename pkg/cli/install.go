package cli

import (
	"os"

	"github.com/NethermindEth/eigen-wiz/pkg/daemon"
	"github.com/spf13/cobra"
)

func InstallCmd(d daemon.Daemon) *cobra.Command {
	var (
		url     string
		version string
	)
	cmd := cobra.Command{
		Use:   "install [URL]",
		Short: "Install AVS node software from a git repository",
		Long:  "Installs the AVS node software, downloading it from a git repository. You will need to provide the repository URL as a unique argument, which must be an HTTP or HTTPS URL. Use the --version flag if you need to specify a version.",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			url = args[0]
			return validatePkgURL(url)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			destDir := os.TempDir() // TODO: change this to the proper directory inside the data directory of the daemon
			_, err := d.Install(&daemon.InstallOptions{
				PullOptions: daemon.PullOptions{
					URL:     url,
					Version: version,
					DestDir: destDir,
				},
			})
			return err
		},
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "version to install. If not specified, the latest version will be installed.")

	return &cmd
}
