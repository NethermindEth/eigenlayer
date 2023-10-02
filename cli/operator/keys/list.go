package keys

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/spf13/cobra"
)

func ListCmd(p prompter.Prompter) *cobra.Command {
	cmd := cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			files, err := os.ReadDir(OperatorKeyFolder + "/")
			if err != nil {
				return err
			}

			for _, file := range files {
				keySplits := strings.Split(file.Name(), ".")
				fileName := keySplits[0]
				keyType := keySplits[1]
				fmt.Println("Key Name: " + fileName)
				switch keyType {
				case KeyTypeECDSA:
					fmt.Println("Key Type: ECDSA")
					address, err := GetAddress(OperatorKeyFolder + "/" + file.Name())
					if err != nil {
						return err
					}
					fmt.Println("Address: 0x" + address)
					fmt.Println("====================================================================================")
					fmt.Println()
				case KeyTypeBLS:
					fmt.Println("Key Type: BLS")
					pubKey, err := GetPubKey(OperatorKeyFolder + "/" + file.Name())
					if err != nil {
						return err
					}
					fmt.Println("Public Key: " + pubKey)
					fmt.Println("====================================================================================")
					fmt.Println()
				}

			}

			return nil
		},
	}

	return &cmd
}

func GetPubKey(keyStoreFile string) (string, error) {
	keyJson, err := os.ReadFile(keyStoreFile)
	if err != nil {
		return "", err
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(keyJson, &m); err != nil {
		return "", err
	}

	if pubKey, ok := m["pubKey"].(string); !ok {
		return "", fmt.Errorf("pubKey not found in key file")
	} else {
		return pubKey, nil
	}
}

func GetAddress(keyStoreFile string) (string, error) {
	keyJson, err := os.ReadFile(keyStoreFile)
	if err != nil {
		return "", err
	}

	m := make(map[string]interface{})
	if err := json.Unmarshal(keyJson, &m); err != nil {
		return "", err
	}

	if address, ok := m["address"].(string); !ok {
		return "", fmt.Errorf("address not found in key file")
	} else {
		return address, nil
	}
}

func GetECDSAPrivateKey(keyStoreFile string, password string) (*ecdsa.PrivateKey, error) {
	keyStoreContents, err := os.ReadFile(keyStoreFile)
	if err != nil {
		return nil, err
	}

	sk, err := keystore.DecryptKey(keyStoreContents, password)
	if err != nil {
		return nil, err
	}

	return sk.PrivateKey, nil
}
