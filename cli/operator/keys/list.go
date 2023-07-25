package keys

import (
	"crypto/ecdsa"
	"fmt"
	"os"
	"strings"

	"github.com/Layr-Labs/eigensdk-go/crypto/bls"
	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
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
					privateKey, err := GetECDSAPrivateKey(OperatorKeyFolder+"/"+file.Name(), "")
					publicKey := privateKey.Public()
					publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
					if !ok {
						return err
					}
					publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
					fmt.Println("Public Key: " + hexutil.Encode(publicKeyBytes)[4:])
					address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
					fmt.Println("Address: " + address)
					// privateKeyHex := hex.EncodeToString(privateKey.D.Bytes())
					// fmt.Println("ECDSA Private Key (Hex):", privateKeyHex)
					fmt.Println("====================================================================================")
					fmt.Println()
				case KeyTypeBLS:
					fmt.Println("Key Type: BLS")
					keyPair, err := bls.ReadPrivateKeyFromFile(OperatorKeyFolder + "/" + file.Name())
					if err != nil {
						return err
					}
					fmt.Println("Public Key: " + keyPair.PubKey.String())
					fmt.Println("====================================================================================")
					fmt.Println()
				}

			}

			return nil
		},
	}

	return &cmd
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
