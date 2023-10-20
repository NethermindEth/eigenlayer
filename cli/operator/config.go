package operator

import (
	"github.com/NethermindEth/eigenlayer/cli/operator/config"
	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/spf13/cobra"
)

func ConfigCmd(p prompter.Prompter) *cobra.Command {
	cmd := cobra.Command{
		Use: "config",
	}

	cmd.AddCommand(
		config.CreateCmd(p),
	)

	return &cmd
}
