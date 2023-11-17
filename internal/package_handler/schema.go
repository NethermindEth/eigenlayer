package package_handler

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

const (
	manifestSchema = "schema/manifest_schema.yml"
	profileSchema  = "schema/profile_schema.yml"
)

//go:embed schema/*
var fs embed.FS

func validateYAMLSchema(schemaFile, documentFile string) error {
	// Read the YAML schema

	schemaData, err := fs.ReadFile(schemaFile)
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

func ValidateFromRepository(repoURL, repoPath string) (err error) {
	err = cloneRepository(repoURL, repoPath)
	if err != nil {
		log.Fatal("Failed to clone repository:", err)
	}
	// Walk through the directories and validate YAML files
	err = walkDirectories(repoPath)
	if err != nil {
		log.Fatal("Failed to walk directories:", err)
	}

	// Remove repoPath
	err = os.RemoveAll(repoPath)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func cloneRepository(url string, path string) error {
	_, err := git.PlainClone(path, false, &git.CloneOptions{
		URL: url,
	})
	if err != nil {
		log.Println("Failed to clone repository:", err)
		return err
	}
	return nil
}

func walkDirectories(root string) error {
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Failed to walk directory:", err)
			return err
		}

		// Check if the file is a 'profile.yml' or 'manifest.yml'
		if info.IsDir() {
			return nil
		}

		if info.Name() == "profile.yml" {
			// Validate the YAML file against the schema
			err := validateYAMLSchema(profileSchema, path)
			if err != nil {
				log.Printf("Failed to validate %s: %s\n", path, err)
			} else {
				log.Printf("%s is valid\n", path)
			}
		}
		if info.Name() == "manifest.yml" {
			// Validate the YAML file against the schema
			err := validateYAMLSchema(manifestSchema, path)
			if err != nil {
				log.Printf("Failed to validate %s: %s\n", path, err)
			} else {
				log.Printf("%s is valid\n", path)
			}
		}

		return nil
	})
	if err != nil {
		log.Println("Failed to walk directories:", err)
		return err
	}

	return nil
}
