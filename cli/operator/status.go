package operator

import (
	"context"
	"fmt"

	"github.com/NethermindEth/eigenlayer/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	eigenChainio "github.com/Layr-Labs/eigensdk-go/chainio/clients"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/elcontracts"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	eigensdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
)

func StatusCmd() *cobra.Command {
	var (
		help        bool
		operatorCfg types.OperatorConfig
	)
	cmd := cobra.Command{
		Use:   "status <configuration-file>",
		Short: "Check if the operator is registered and get the operator details",
		Long: `
		Check the registration status of operator to Eigenlayer.

		It expects the same configuration yaml file as arugment to register command
		`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Parse static flags
			cmd.DisableFlagParsing = false
			cmd.FParseErrWhitelist.UnknownFlags = true // Don't show error for unknown flags to allow dynamic flags
			err := cmd.ParseFlags(args)
			if err != nil {
				return err
			}

			// Skip execution if help flag is set
			help, err = cmd.Flags().GetBool("help")
			if err != nil {
				return err
			}
			if help {
				return nil
			}

			// Validate args
			args = cmd.Flags().Args()
			if len(args) != 1 {
				return fmt.Errorf("%w: accepts 1 arg, received %d", ErrInvalidNumberOfArgs, len(args))
			}

			configurationFilePath := args[0]

			err = eigensdkUtils.ReadYamlConfig(configurationFilePath, &operatorCfg)
			if err != nil {
				return err
			}

			fmt.Printf("Operator configuration file read successfully %s\n", operatorCfg.Operator.Address)
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			llog, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			ethClient, err := eth.NewClient(operatorCfg.EthRPCUrl)
			if err != nil {
				return err
			}

			elContractsClient, err := eigenChainio.NewELContractsChainClient(
				common.HexToAddress(operatorCfg.ELSlasherAddress),
				common.HexToAddress(operatorCfg.BlsPublicKeyCompendiumAddress),
				ethClient,
				ethClient,
				llog)
			if err != nil {
				return err
			}

			reader, err := elContracts.NewELChainReader(
				elContractsClient,
				llog,
				ethClient,
			)
			if err != nil {
				return err
			}

			status, err := reader.IsOperatorRegistered(context.Background(), operatorCfg.Operator)
			if err != nil {
				return err
			}

			if status {
				fmt.Println("Operator is registered")
				operatorDetails, err := reader.GetOperatorDetails(context.Background(), operatorCfg.Operator)
				if err != nil {
					return err
				}
				fmt.Printf("Operator details: %+v\n", operatorDetails)
			} else {
				fmt.Println("Operator is not registered")
			}
			return nil
		},
	}

	return &cmd
}
