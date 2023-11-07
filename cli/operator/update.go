package operator

import (
	"context"
	"fmt"

	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/NethermindEth/eigenlayer/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/spf13/cobra"

	eigenChainio "github.com/Layr-Labs/eigensdk-go/chainio/clients"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/elcontracts"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/metrics"
	eigensdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
)

func UpdateCmd(p prompter.Prompter) *cobra.Command {
	var (
		help           bool
		operatorCfg    types.OperatorConfig
		signerTypeFlag string
		signerType     types.SignerType
		privateKeyHex  string
	)
	cmd := cobra.Command{
		Use:   "update [flags] <configuration-file>",
		Short: "Updates the operator metadata",
		Long: `
		Updates the operator metadata onchain which includes 
			- metadata url
			- delegation approver address
			- earnings receiver address
			- staker opt out window blocks

		Requires the same file used for registration as argument
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
				return fmt.Errorf("%w: with error %s", ErrInvalidYamlFile, err)
			}

			fmt.Printf("Operator configuration file read successfully %s\n", operatorCfg.Operator.Address)

			signerType, err = validateSignerType(signerTypeFlag, operatorCfg)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			localSigner, err := getSigner(p, signerType, privateKeyHex, operatorCfg)
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
				logger,
			)
			if err != nil {
				return err
			}

			noopMetrics := metrics.NewNoopMetrics()
			elWriter := elContracts.NewELChainWriter(
				elContractsClient,
				ethClient,
				localSigner,
				logger,
				noopMetrics,
			)

			if err != nil {
				return err
			}
			receipt, err := elWriter.UpdateOperatorDetails(context.Background(), operatorCfg.Operator)
			if err != nil {
				return err
			}
			logger.Infof("Operator details updated at: %s", getTransactionLink(receipt.TxHash.String(), &operatorCfg.ChainId))

			logger.Info("Operator updated successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&signerTypeFlag, "signer-type", "", "Signer type (private_key, local_keystore)")
	cmd.Flags().StringVar(&privateKeyHex, "private-key-hex", "", "Private key hex (used only with private_key signer type)")

	return &cmd
}
