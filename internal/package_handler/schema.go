package package_handler

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

func validateYAMLSchema(schemaFile, documentFile string) error {
	// Read the YAML schema
	schemaData, err := os.ReadFile(schemaFile)
	if err != nil {
		return fmt.Errorf("error reading the schema file: %v", err)
	}

	// Deserialize from YAML to generic interface{}
	var schemaYaml interface{}
	if err := yaml.Unmarshal(schemaData, &schemaYaml); err != nil {
		return fmt.Errorf("error deserializing the yaml schema: %v", err)
	}

	// Serialize from generic interface{} to JSON
	schemaJson, err := json.Marshal(schemaYaml)
	if err != nil {
		return fmt.Errorf("error serializing the schema to json: %v", err)
	}

	// Compile the JSON schema
	schemaLoader := gojsonschema.NewStringLoader(string(schemaJson))
	schema, err := gojsonschema.NewSchema(schemaLoader)
	if err != nil {
		return fmt.Errorf("error compiling the schema: %v", err)
	}

	// Read the document
	documentData, err := os.ReadFile(documentFile)
	if err != nil {
		return fmt.Errorf("error reading the document file: %v", err)
	}

	var document interface{}
	if err := yaml.Unmarshal(documentData, &document); err != nil {
		return fmt.Errorf("error deserializing the document: %v", err)
	}

	// Validate the document with the schema
	result, err := schema.Validate(gojsonschema.NewGoLoader(document))
	if err != nil {
		return fmt.Errorf("validation error: %v", err)
	}

	if !result.Valid() {
		return fmt.Errorf("many errors found: %v", result.Errors())
	}
	return nil
}
