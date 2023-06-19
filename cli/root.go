package cli

import (
	"github.com/NethermindEth/egn/pkg/daemon"
	"github.com/spf13/cobra"
)

func RootCmd(d daemon.Daemon) *cobra.Command {
	cmd := cobra.Command{
		Use: "egn",
	}
	cmd.AddCommand(
		InstallCmd(d),
	)
	cmd.CompletionOptions.DisableDefaultCmd = true
	return &cmd
}
