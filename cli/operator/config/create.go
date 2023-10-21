package config

import (
	"encoding/json"
	"errors"
	"math/big"
	"os"
	"regexp"
	"strings"

	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/NethermindEth/eigenlayer/internal/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const zeroAddress = "0x0000000000000000000000000000000000000000"

func CreateCmd(p prompter.Prompter) *cobra.Command {
	cmd := cobra.Command{
		Use:   "create",
		Short: "Used to create operator config and metadata json sample file",
		Long: `
		This command is used to create a sample empty operator config file 
		and also an empty metadata json file which you need to upload for
		operator metadata

		Both of these are needed for operator registration
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			op := types.OperatorConfig{}

			// Prompt user to generate empty or non-empty files
			populate, err := p.Confirm("Would you like to populate the operator config file?")
			if err != nil {
				return err
			}

			if populate {
				op, err = promptOperatorInfo(&op, p)
				if err != nil {
					return err
				}
			}

			yamlData, err := yaml.Marshal(&op)
			if err != nil {
				return err
			}
			operatorFile := "operator.yaml"
			err = os.WriteFile(operatorFile, yamlData, 0o644)
			if err != nil {
				return err
			}

			metadata := eigensdkTypes.OperatorMetadata{}
			jsonData, err := json.MarshalIndent(metadata, "", "  ")
			if err != nil {
				return err
			}

			metadataFile := "metadata.json"
			err = os.WriteFile(metadataFile, jsonData, 0o644)
			if err != nil {
				return err
			}
			return nil
		},
	}

	return &cmd
}

func promptOperatorInfo(config *types.OperatorConfig, p prompter.Prompter) (types.OperatorConfig, error) {
	// Prompt and set operator address
	operatorAddress, err := p.InputString("Enter your operator address:", "", "",
		func(s string) error {
			return validateAddressIsNonZeroAndValid(s)
		},
	)
	if err != nil {
		return types.OperatorConfig{}, err
	}
	config.Operator.Address = operatorAddress

	// Prompt to gate stakers approval
	gateApproval, err := p.Confirm("Do you want to gate stakers approval?")
	if err != nil {
		return types.OperatorConfig{}, err
	}

	// Prompt for address if operator wants to gate approvals
	if gateApproval {
		delegationApprover, err := p.InputString("Enter your staker approver address:", "", "",
			func(s string) error {
				return validateAddress(s)
			},
		)
		if err != nil {
			return types.OperatorConfig{}, err
		}
		config.Operator.DelegationApproverAddress = delegationApprover
	} else {
		config.Operator.DelegationApproverAddress = zeroAddress
	}

	// Prompt and set earnings address
	earningsAddress, err := p.InputString("Enter your earnings address:", config.Operator.Address, "",
		func(s string) error {
			return validateAddressIsNonZeroAndValid(s)
		},
	)
	if err != nil {
		return types.OperatorConfig{}, err
	}
	config.Operator.EarningsReceiverAddress = earningsAddress

	// Prompt for eth node
	rpcUrl, err := p.InputString("Enter your rpc url:", "", "",
		func(s string) error { return nil },
	)
	if err != nil {
		return types.OperatorConfig{}, err
	}
	config.EthRPCUrl = rpcUrl

	// Prompt for ecdsa key path
	ecdsaKeyPath, err := p.InputString("Enter your ecdsa key path:", "", "",
		func(s string) error { return nil },
	)
	if err != nil {
		return types.OperatorConfig{}, err
	}
	config.PrivateKeyStorePath = ecdsaKeyPath

	// Prompt for bls key path
	blsKeyPath, err := p.InputString("Enter your bls key path:", "", "",
		func(s string) error { return nil },
	)
	if err != nil {
		return types.OperatorConfig{}, err
	}
	config.BlsPrivateKeyStorePath = blsKeyPath

	// Prompt for network & set chainId
	chainId, err := p.Select("Select your network:", []string{"mainnet", "goerli", "local"})
	if err != nil {
		return types.OperatorConfig{}, err
	}

	switch chainId {
	case "mainnet":
		config.ChainId = *big.NewInt(1)
	case "goerli":
		config.ChainId = *big.NewInt(5)
	case "local":
		config.ChainId = *big.NewInt(31337)
	}

	return *config, nil
}

func validateAddressIsNonZeroAndValid(address string) error {
	if address == zeroAddress {
		return errors.New("address is 0")
	}

	return validateAddress(address)
}

func validateAddress(address string) error {
	// Remove 0x
	address = strings.TrimPrefix(address, "0x")

	// Check if address has 40 hexadecimal characters
	isValid, _ := regexp.MatchString("^[0-9a-fA-F]{40}$", address)

	if !isValid {
		return errors.New("invalid address")
	}

	return nil
}
