package operator

import (
	"context"
	"fmt"

	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/NethermindEth/eigenlayer/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	eigenChainio "github.com/Layr-Labs/eigensdk-go/chainio/clients"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/elcontracts"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	eigensdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
)

func UpdateCmd(p prompter.Prompter) *cobra.Command {
	var (
		configurationFilePath string
		help                  bool
		operatorCfg           types.OperatorConfig
		signerTypeFlag        string
		signerType            types.SignerType
		privateKeyHex         string
	)
	cmd := cobra.Command{
		Use: "update [flags]",
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

			signerType, err = validateSignerType(signerTypeFlag, operatorCfg)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			llog, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			localSigner, err := getSigner(p, signerType, privateKeyHex, operatorCfg)
			if err != nil {
				return err
			}

			ethClient, err := eigenChainio.NewEthClient(operatorCfg.EthRPCUrl)
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
			elWriter, err := elContracts.NewELChainWriter(
				elContractsClient,
				ethClient,
				localSigner,
				llog,
			)

			if err != nil {
				return err
			}
			_, err = elWriter.UpdateOperatorDetails(context.Background(), operatorCfg.Operator)
			if err != nil {
				return err
			}
			fmt.Println("Operator updated successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&configurationFilePath, "configuration-file", "", "Path to the configuration file")
	cmd.Flags().StringVar(&signerTypeFlag, "signer-type", "", "Signer type (private_key, local_keystore)")
	cmd.Flags().StringVar(&privateKeyHex, "private-key-hex", "", "Private key hex (used only with private_key signer type)")

	return &cmd
}
