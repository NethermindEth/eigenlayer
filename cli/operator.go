package cli

import (
	"github.com/NethermindEth/eigenlayer/cli/operator"
	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/spf13/cobra"
)

func OperatorCmd(p prompter.Prompter) *cobra.Command {
	cmd := cobra.Command{
		Use:   "operator",
		Short: "Execute onchain operations for the operator",
	}

	cmd.AddCommand(
		operator.RegisterCmd(p),
		operator.UpdateCmd(p),
		operator.StatusCmd(),
		operator.ConfigCmd(),
		operator.KeysCmd(p),
	)
	return &cmd
}
