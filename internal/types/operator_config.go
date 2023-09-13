package types

import (
	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
)

type SignerType string

const (
	PrivateKeySigner    SignerType = "private_key"
	LocalKeystoreSigner SignerType = "local_keystore"
)

type OperatorConfig struct {
	Operator                      eigensdkTypes.Operator `yaml:"operator"`
	ELSlasherAddress              string                 `yaml:"el_slasher_address"`
	BlsPublicKeyCompendiumAddress string                 `yaml:"bls_public_key_compendium_address"`
	EthRPCUrl                     string                 `yaml:"eth_rpc_url"`
	PrivateKeyStorePath           string                 `yaml:"private_key_store_path"`
	SignerType                    SignerType             `yaml:"signer_type"`
	BlsPrivateKeyStorePath        string                 `yaml:"bls_private_key_store_path"`
}
