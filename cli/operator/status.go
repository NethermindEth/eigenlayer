package operator

import (
	"context"
	"fmt"

	"github.com/NethermindEth/eigenlayer/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	eigensdkReader "github.com/Layr-Labs/eigensdk-go/chainio/readers"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	eigensdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
)

func StatusCmd() *cobra.Command {
	var (
		configurationFilePath string
		help                  bool
		operatorCfg           types.OperatorConfig
	)
	cmd := cobra.Command{
		Use: "status",
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

			reader, err := eigensdkReader.NewEigenlayerReader(
				common.HexToAddress(operatorCfg.ELSlasherAddress),
				common.HexToAddress(operatorCfg.BlsPublicKeyCompendiumAddress),
				operatorCfg.EthRPCUrl,
				llog,
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

	cmd.Flags().StringVar(&configurationFilePath, "configuration-file", "", "Path to the configuration file")

	return &cmd
}
