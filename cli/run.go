package cli

import (
	"log"

	"github.com/NethermindEth/eigen-wiz/pkg/daemon"
	"github.com/spf13/cobra"
)

func RunCmd(d daemon.Daemon) *cobra.Command {
	var (
		pkgName    string
		pkgVersion string
	)
	cmd := cobra.Command{
		Use:   "run [name]",
		Short: "Runs AVS node software from a package that has been installed",
		Long:  "Runs the AVS node software from a package that has been installed. You will need to provide the package name as a unique argument. Use the --version flag if you need to specify a version.",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			pkgName = args[0]
			// TODO: check if the package is installed
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			log.Println("Running package", pkgName, "version", pkgVersion)
			// TODO: call daemon Run
			return nil
		},
	}

	cmd.Flags().StringVarP(&pkgVersion, "version", "v", "latest", "version to run")
	return &cmd
}
