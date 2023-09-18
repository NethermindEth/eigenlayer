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
)

func CreateCmd(p prompter.Prompter) *cobra.Command {
	var keyType string

	cmd := cobra.Command{
		Use: "create [keyname] [flags]",
		RunE: func(cmd *cobra.Command, args []string) error {
			keyName := args[0]

			if len(keyName) == 0 {
				return errors.New("key name cannot be empty")
			}

			if match, _ := regexp.MatchString("\\s", keyName); match {
				return errors.New("key name contains whitespace")
			}

			basePath, _ := os.Getwd()

			switch keyType {
			case KeyTypeECDSA:
				keyFileName := keyName + ".ecdsa.key.json"
				if checkIfKeyExists(keyFileName) {
					return errors.New("key name already exists. Please choose a different name")
				}

				privateKey, err := crypto.GenerateKey()
				if err != nil {
					return err
				}

				password, err := p.InputHiddenString("Enter password to encrypt the ecdsa private key:", "",
					func(password string) error {
						return nil
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
					func(password string) error {
						return nil
					},
				)
				if err != nil {
					return err
				}
				blsKeyPair, err := bls.GenRandomBlsKeys()
				if err != nil {
					return err
				}
				err = blsKeyPair.SaveToFile(OperatorKeyFolder+"/"+keyFileName, password)
				if err != nil {
					return err
				}
				fmt.Println("BLS Private Key: " + blsKeyPair.PrivKey.String())
				fmt.Println("Please backup the above private key in safe place.")
				fmt.Println()
				fmt.Println("BLS Pub key: " + blsKeyPair.PubKey.String())
				fmt.Println("Key location: " + basePath + "/" + OperatorKeyFolder + "/" + keyFileName)
			default:
				return errors.New("key type must be either 'ecdsa' or 'bls'")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&keyType, "key-type", "", "Type of key you want to create. Currently supports 'ecdsa' and 'bls'")

	return &cmd
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
	err := os.Mkdir(OperatorKeyFolder, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	file, err := os.Create(filepath.Clean(OperatorKeyFolder + "/" + keyName))
	if err != nil {
		fmt.Println("file create error")
		return err
	}
	// remember to close the file
	defer file.Close()

	fileScanner := bufio.NewScanner(file)
	fileScanner.Split(bufio.ScanLines)
	_, err = file.Write(data)

	return err
}

func checkIfKeyExists(keyName string) bool {
	_, err := os.Stat(OperatorKeyFolder + "/" + keyName)
	return !os.IsNotExist(err)
}
