package cli

import (
	"github.com/NethermindEth/egn/cli/prompter"
	"github.com/NethermindEth/egn/pkg/daemon"
	"github.com/spf13/cobra"
)

func RootCmd(d daemon.Daemon, p prompter.Prompter) *cobra.Command {
	cmd := cobra.Command{
		Use: "egn",
	}
	cmd.AddCommand(
		InstallCmd(d, p),
		StopCmd(d),
		UninstallCmd(d),
	)
	cmd.CompletionOptions.DisableDefaultCmd = true
	return &cmd
}
