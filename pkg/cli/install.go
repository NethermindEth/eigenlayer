package cli

import (
	"fmt"
	"net/url"

	"github.com/NethermindEth/eigen-wiz/pkg/daemon"
	"github.com/spf13/cobra"
)

func InstallCmd(d daemon.Daemon) *cobra.Command {
	var (
		pkgURL     string
		pkgVersion string
	)
	cmd := cobra.Command{
		Use:   "install [URL]",
		Short: "Install AVS node software from a git repository",
		Long:  "Installs the AVS node software, downloading it from a git repository. You will need to provide the repository URL as a unique argument, which must be an HTTP or HTTPS URL. Use the --version flag if you need to specify a version.",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			parsedURL, err := url.ParseRequestURI(args[0])
			if err != nil {
				return fmt.Errorf("%w: %s", ErrInvalidURL, err.Error())
			}
			if parsedURL.Scheme != "https" && parsedURL.Scheme != "http" {
				return fmt.Errorf("%w: %s", ErrInvalidURL, "URL must be HTTP or HTTPS")
			}
			pkgURL = args[0]
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := d.Install(&daemon.InstallOptions{
				URL:     pkgURL,
				Version: pkgVersion,
			})
			return err
		},
	}

	cmd.Flags().StringVarP(&pkgVersion, "version", "v", "latest", "version to install")
	return &cmd
}
