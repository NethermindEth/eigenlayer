package keys

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

func ImportCmd(p prompter.Prompter) *cobra.Command {
	var (
		keyType  string
		insecure bool
		help     bool
	)

	cmd := cobra.Command{
		Use:   "import --key-type <key-type> [flags] <keyname> <private-key>",
		Short: "Used to import existing keys in local keystore",
		Long: `
Used to import ecdsa and bls key in local keystore

keyname (required) - This will be the name of the imported key file. It will be saved as <keyname>.ecdsa.key.json or <keyname>.bls.key.json

use --key-type ecdsa/bls to import ecdsa/bls key. 
- ecdsa - <private-key> should be plaintext hex encoded private key
- bls - <private-key> should be plaintext bls private key

It will prompt for password to encrypt the key, which is optional but highly recommended.
If you want to import a key with weak/no password, use --insecure flag. Do NOT use those keys in production

This command will import keys in ./operator_keys/ location
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
			if len(args) != 2 {
				return fmt.Errorf("%w: accepts 2 arg, received %d", ErrInvalidNumberOfArgs, len(args))
			}

			keyName := args[0]
			if len(keyName) == 0 {
				return ErrEmptyKeyName
			}

			if match, _ := regexp.MatchString("\\s", keyName); match {
				return ErrKeyContainsWhitespaces
			}

			privateKey := args[1]
			if len(privateKey) == 0 {
				return ErrEmptyPrivateKey
			}

			if match, _ := regexp.MatchString("\\s", privateKey); match {
				return ErrPrivateKeyContainsWhitespaces
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			keyName := args[0]
			privateKey := args[1]

			switch keyType {
			case KeyTypeECDSA:
				privateKey = strings.TrimPrefix(privateKey, "0x")
				privateKeyPair, err := crypto.HexToECDSA(privateKey)
				if err != nil {
					return err
				}
				return saveEcdsaKey(keyName, p, privateKeyPair, insecure)
			case KeyTypeBLS:
				blsKeyPair, err := bls.NewKeyPairFromString(privateKey)
				if err != nil {
					return err
				}
				return saveBlsKey(keyName, p, blsKeyPair, insecure)
			default:
				return ErrInvalidKeyType
			}
		},
	}

	cmd.Flags().StringVar(&keyType, "key-type", "", "Type of key you want to create. Currently supports 'ecdsa' and 'bls'")
	cmd.Flags().BoolVarP(&insecure, "insecure", "i", false, "Use this flag to skip password validation")

	return &cmd
}
