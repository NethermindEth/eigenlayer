package config

import (
	"encoding/json"
	"os"

	eigensdkTypes "github.com/Layr-Labs/eigensdk-go/types"
	"github.com/NethermindEth/eigenlayer/internal/types"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func CreateCmd() *cobra.Command {
	cmd := cobra.Command{
		Use:   "create",
		Short: "Used to creata operator config and metadata json sample file",
		Long: `
		This command is used to create a sample empty operator config file 
		and also an empty metadata json file which you need to upload for
		operator metadata

		Both of these are needed for operator registration
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			op := types.OperatorConfig{}
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
