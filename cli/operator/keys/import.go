package keys

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

Command:
	eigenlayer operator keys import --key-type <key-type> <keyname> <private-key>

keyname and private-key is required

use --key-type ecdsa/bls to create ecdsa/bls key. 
- ecdsa - private-key should be hex encoded private key
- bls - private-key should be bls private key

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
			basePath, _ := os.Getwd()

			switch keyType {
			case KeyTypeECDSA:
				keyFileName := keyName + ".ecdsa.key.json"
				if checkIfKeyExists(keyFileName) {
					return errors.New("key name already exists. Please choose a different name")
				}

				if strings.HasPrefix(privateKey, "0x") {
					privateKey = privateKey[2:]
				}

				privateKeyPair, err := crypto.HexToECDSA(privateKey)
				if err != nil {
					return err
				}

				password, err := p.InputHiddenString("Enter password to encrypt the ecdsa private key:", "",
					func(s string) error {
						if insecure {
							return nil
						}
						return validatePassword(s)
					},
				)
				if err != nil {
					return err
				}

				err = WriteEncryptedECDSAPrivateKeyToPath(keyFileName, privateKeyPair, password)
				if err != nil {
					return err
				}

				privateKeyHex := hex.EncodeToString(privateKeyPair.D.Bytes())
				// TODO: display it using `less` of `vi` so that it is not saved in terminal history
				fmt.Println("ECDSA Private Key (Hex): ", privateKeyHex)
				fmt.Println("Please backup the above private key hex in safe place.")
				fmt.Println()
				fmt.Println("Key location: " + basePath + "/" + OperatorKeyFolder + "/" + keyFileName)
				publicKey := privateKeyPair.Public()
				publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
				if !ok {
					return err
				}
				publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
				fmt.Println(hexutil.Encode(publicKeyBytes)[4:])
				address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
				fmt.Println(address)

				return nil
			case KeyTypeBLS:
				keyFileName := keyName + ".bls.key.json"
				if checkIfKeyExists(keyFileName) {
					return errors.New("key name already exists. Please choose a different name")
				}
				password, err := p.InputHiddenString("Enter password to encrypt the bls private key:", "",
					func(s string) error {
						if insecure {
							return nil
						}
						return validatePassword(s)
					},
				)
				if err != nil {
					return err
				}
				blsKeyPair, err := bls.NewKeyPairFromString(privateKey)
				if err != nil {
					return err
				}
				err = blsKeyPair.SaveToFile(OperatorKeyFolder+"/"+keyFileName, password)
				if err != nil {
					return err
				}
				// TODO: display it using `less` of `vi` so that it is not saved in terminal history
				fmt.Println("BLS Private Key: " + blsKeyPair.PrivKey.String())
				fmt.Println("Please backup the above private key in safe place.")
				fmt.Println()
				fmt.Println("BLS Pub key: " + blsKeyPair.PubKey.String())
				fmt.Println("Key location: " + basePath + "/" + OperatorKeyFolder + "/" + keyFileName)
			default:
				return ErrInvalidKeyType
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&keyType, "key-type", "", "Type of key you want to create. Currently supports 'ecdsa' and 'bls'")
	cmd.Flags().BoolVarP(&insecure, "insecure", "i", false, "Use this flag to skip password validation")

	return &cmd
}
