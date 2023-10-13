package keys

import (
	"bufio"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	passwordvalidator "github.com/wagslane/go-password-validator"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const (
	OperatorKeyFolder = "operator_keys"

	KeyTypeECDSA = "ecdsa"
	KeyTypeBLS   = "bls"

	// MinEntropyBits For password validation
	MinEntropyBits = 70
)

func CreateCmd(p prompter.Prompter) *cobra.Command {
	var (
		keyType  string
		insecure bool
		help     bool
	)

	cmd := cobra.Command{
		Use:   "create --key-type <key-type> [flags] <keyname>",
		Short: "Used to create encrypted keys in local keystore",
		Long: `
Used to create ecdsa and bls key in local keystore

keyname (required) - This will be the name of the created key file. It will be saved as <keyname>.ecdsa.key.json or <keyname>.bls.key.json

use --key-type ecdsa/bls to create ecdsa/bls key. 
It will prompt for password to encrypt the key, which is optional but highly recommended.
If you want to create a key with weak/no password, use --insecure flag. Do NOT use those keys in production

This command will create keys in ./operator_keys/ location
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

			keyName := args[0]
			if len(keyName) == 0 {
				return ErrEmptyKeyName
			}

			if match, _ := regexp.MatchString("\\s", keyName); match {
				return ErrKeyContainsWhitespaces
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			keyName := args[0]

			switch keyType {
			case KeyTypeECDSA:
				privateKey, err := crypto.GenerateKey()
				if err != nil {
					return err
				}
				return saveEcdsaKey(keyName, p, privateKey, insecure)
			case KeyTypeBLS:
				blsKeyPair, err := bls.GenRandomBlsKeys()
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

func saveBlsKey(keyName string, p prompter.Prompter, keyPair *bls.KeyPair, insecure bool) error {
	// TODO: Path should be relative to user home dir https://github.com/NethermindEth/eigenlayer/issues/109
	basePath, _ := os.Getwd()
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

	err = keyPair.SaveToFile(OperatorKeyFolder+"/"+keyFileName, password)
	if err != nil {
		return err
	}
	// TODO: display it using `less` of `vi` so that it is not saved in terminal history
	fmt.Println("BLS Private Key: " + keyPair.PrivKey.String())
	fmt.Println("Please backup the above private key in safe place.")
	fmt.Println()
	fmt.Println("BLS Pub key: " + keyPair.PubKey.String())
	fmt.Println("Key location: " + basePath + "/" + OperatorKeyFolder + "/" + keyFileName)
	return nil
}

func saveEcdsaKey(keyName string, p prompter.Prompter, privateKey *ecdsa.PrivateKey, insecure bool) error {
	// TODO: Path should be relative to user home dir https://github.com/NethermindEth/eigenlayer/issues/109
	basePath, _ := os.Getwd()
	keyFileName := keyName + ".ecdsa.key.json"
	if checkIfKeyExists(keyFileName) {
		return errors.New("key name already exists. Please choose a different name")
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

	err = WriteEncryptedECDSAPrivateKeyToPath(keyFileName, privateKey, password)
	if err != nil {
		return err
	}

	privateKeyHex := hex.EncodeToString(privateKey.D.Bytes())
	// TODO: display it using `less` of `vi` so that it is not saved in terminal history
	fmt.Println("ECDSA Private Key (Hex): ", privateKeyHex)
	fmt.Println("Please backup the above private key hex in safe place.")
	fmt.Println()
	fmt.Println("Key location: " + basePath + "/" + OperatorKeyFolder + "/" + keyFileName)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return err
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	fmt.Println("Public Key hex: ", hexutil.Encode(publicKeyBytes)[4:])
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Println("Ethereum Address", address)
	return nil
}

func WriteEncryptedECDSAPrivateKeyToPath(keyName string, privateKey *ecdsa.PrivateKey, password string) error {
	UUID, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	key := &keystore.Key{
		Id:         UUID,
		Address:    crypto.PubkeyToAddress(privateKey.PublicKey),
		PrivateKey: privateKey,
	}

	encryptedBytes, err := keystore.EncryptKey(key, password, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return err
	}

	return writeBytesToFile(keyName, encryptedBytes)
}

func writeBytesToFile(keyName string, data []byte) error {
	err := os.Mkdir(OperatorKeyFolder, 0o755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	file, err := os.Create(filepath.Clean(OperatorKeyFolder + "/" + keyName))
	if err != nil {
		fmt.Println("file create error")
		return err
	}
	// remember to close the file
	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	_, err = file.Write(data)

	return err
}

func checkIfKeyExists(keyName string) bool {
	_, err := os.Stat(OperatorKeyFolder + "/" + keyName)
	return !os.IsNotExist(err)
}

func validatePassword(password string) error {
	err := passwordvalidator.Validate(password, MinEntropyBits)
	if err != nil {
		fmt.Println("if you want to create keys for testing with weak/no password, use --insecure flag. Do NOT use those keys in production")
	}
	return err
}
