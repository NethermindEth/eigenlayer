package operator

import (
	"context"
	"fmt"
	"os"

	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/NethermindEth/eigenlayer/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	eigenChainio "github.com/Layr-Labs/eigensdk-go/chainio/clients"
	"github.com/Layr-Labs/eigensdk-go/chainio/clients/eth"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/metrics"
	"github.com/Layr-Labs/eigensdk-go/signer"
	eigensdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
)

func RegisterCmd(p prompter.Prompter) *cobra.Command {
	var (
		help           bool
		operatorCfg    types.OperatorConfig
		signerTypeFlag string
		signerType     types.SignerType
		privateKeyHex  string
	)
	cmd := cobra.Command{
		Use:   "register [flags] <configuration-file>",
		Short: "Register the operator and the BLS public key in the Eigenlayer contracts",
		Long: `
		Register command expects a yaml config file as an argument
		to successfully register an operator address to eigenlayer

		This will register operator to DelegationManager and will register
		the BLS public key on eigenlayer
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
			fmt.Printf("validating operator config: %s\n", operatorCfg.Operator.Address)

			err = operatorCfg.Operator.Validate()
			if err != nil {
				return fmt.Errorf("%w: with error %s", ErrInvalidYamlFile, err)
			}

			fmt.Println("Operator file validated successfully")

			signerType, err = validateSignerType(signerTypeFlag, operatorCfg)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			llog, err := eigensdkLogger.NewZapLogger(eigensdkLogger.Development)
			if err != nil {
				return err
			}

			localSigner, err := getSigner(p, signerType, privateKeyHex, operatorCfg)
			if err != nil {
				return err
			}

			blsKeyPassword, err := p.InputHiddenString("Enter password to decrypt the bls private key:", "",
				func(password string) error {
					return nil
				},
			)
			if err != nil {
				fmt.Println("Error while reading bls key password")
				return err
			}

			keyPair, err := bls.ReadPrivateKeyFromFile(operatorCfg.BlsPrivateKeyStorePath, blsKeyPassword)
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

			noopMetrics := metrics.NewNoopMetrics()
			elWriter := elContracts.NewELChainWriter(
				elContractsClient,
				ethClient,
				localSigner,
				llog,
				noopMetrics,
			)
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

			if !status {
				_, err = elWriter.RegisterAsOperator(ctx, operatorCfg.Operator)
				if err != nil {
					return err
				}
			} else {
				llog.Info("Operator is already registered")
			}

			_, err = elWriter.RegisterBLSPublicKey(ctx, keyPair, operatorCfg.Operator)
			if err != nil {
				return err
			}
			llog.Info("Operator is registered and bls key added successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&signerTypeFlag, "signer-type", "", "Signer type (private_key, local_keystore)")
	cmd.Flags().StringVar(&privateKeyHex, "private-key-hex", "", "Private key hex (used only with private_key signer type)")

	return &cmd
}

func validateSignerType(signerType string, operatorCfg types.OperatorConfig) (types.SignerType, error) {
	signerType, err := getSignerType(signerType, operatorCfg)
	if err != nil {
		return "", err
	}

	switch signerType {
	case string(types.PrivateKeySigner):
		return types.PrivateKeySigner, nil
	case string(types.LocalKeystoreSigner):
		return types.LocalKeystoreSigner, nil
	default:
		return "", fmt.Errorf("invalid signer type %s", signerType)
	}
}

func getSignerType(signerType string, operatorCfg types.OperatorConfig) (string, error) {
	// First get from the command line flag
	if signerType != "" {
		return signerType, nil
	}

	// If command line flag is not set, get from the configuration file
	return string(operatorCfg.SignerType), nil
}

func getSigner(p prompter.Prompter, signerType types.SignerType, privateKeyHex string, operatorCfg types.OperatorConfig) (signer.Signer, error) {
	var localSigner signer.Signer
	switch signerType {
	case types.PrivateKeySigner:
		fmt.Println("Using private key signer")
		if privateKeyHex == "" {
			// If not a flag then read from env
			privateKeyHex := os.Getenv("PRIVATE_KEY")
			if privateKeyHex == "" {
				return nil, fmt.Errorf("please set the private key using the flag or the PRIVATE_KEY environment variable")
			}
		}
		privateKey, err := crypto.HexToECDSA(privateKeyHex)
		if err != nil {
			return nil, err
		}
		// TODO: Get chain ID from config
		localSigner, err = signer.NewPrivateKeySigner(privateKey, &operatorCfg.ChainId)
		if err != nil {
			return nil, err
		}
		return localSigner, nil
	case types.LocalKeystoreSigner:
		fmt.Println("Using local keystore signer")
		ecdsaPassword, err := p.InputHiddenString("Enter password to decrypt the ecdsa private key:", "",
			func(password string) error {
				return nil
			},
		)
		if err != nil {
			fmt.Println("Error while reading ecdsa key password")
			return nil, err
		}
		// TODO: Get chain ID from config
		localSigner, err := signer.NewPrivateKeyFromKeystoreSigner(operatorCfg.PrivateKeyStorePath, ecdsaPassword, &operatorCfg.ChainId)
		if err != nil {
			return nil, err
		}
		return localSigner, nil

	default:
		return nil, fmt.Errorf("invalid signer type %s", signerType)
	}
}
