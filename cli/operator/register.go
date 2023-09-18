package operator

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/NethermindEth/eigenlayer/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"

	eigenChainio "github.com/Layr-Labs/eigensdk-go/chainio/clients"
	elContracts "github.com/Layr-Labs/eigensdk-go/chainio/elcontracts"
	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	eigensdkLogger "github.com/Layr-Labs/eigensdk-go/logging"
	"github.com/Layr-Labs/eigensdk-go/signer"
	eigensdkUtils "github.com/Layr-Labs/eigensdk-go/utils"
)

func RegisterCmd(p prompter.Prompter) *cobra.Command {
	var (
		configurationFilePath string
		help                  bool
		operatorCfg           types.OperatorConfig
		signerTypeFlag        string
		signerType            types.SignerType
		privateKeyHex         string
	)
	cmd := cobra.Command{
		Use:   "register [flags]",
		Short: "Register the operator and the BLS public key in the Eigenlayer contracts",
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
			fmt.Printf("validating operator config: %s\n", operatorCfg.Operator.Address)

			err = operatorCfg.Operator.Validate()
			if err != nil {
				return err
			}

			fmt.Printf("Operator file correct")

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

			_, err = elWriter.RegisterAsOperator(ctx, operatorCfg.Operator)
			if err != nil {
				return err
			}

			keyPair, err := bls.ReadPrivateKeyFromFile(operatorCfg.BlsPrivateKeyStorePath, "")
			if err != nil {
				return err
			}

			_, err = elWriter.RegisterBLSPublicKey(ctx, keyPair, operatorCfg.Operator)
			if err != nil {
				return err
			}
			fmt.Println("Operator registered successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&configurationFilePath, "configuration-file", "", "Path to the configuration file")
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
	chainId := big.NewInt(31337)
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
		localSigner, err = signer.NewPrivateKeySigner(privateKey, chainId)
		if err != nil {
			return nil, err
		}
		return localSigner, nil
	case types.LocalKeystoreSigner:
		fmt.Println("Using local keystore signer")
		// TODO: Get chain ID from config
		localSigner, err := signer.NewPrivateKeyFromKeystoreSigner(operatorCfg.PrivateKeyStorePath, "", chainId)
		if err != nil {
			return nil, err
		}
		return localSigner, nil

	default:
		return nil, fmt.Errorf("invalid signer type %s", signerType)
	}
}
