package cli

import (
	"github.com/NethermindEth/eigen-wiz/pkg/daemon"
	"github.com/spf13/cobra"
)

func RootCmd(d daemon.Daemon) *cobra.Command {
	cmd := cobra.Command{
		Use: "eigen-wiz",
	}
	cmd.AddCommand(
		PullCmd(d),
		InstallCmd(d),
	)
	return &cmd
}
