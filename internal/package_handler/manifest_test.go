package package_handler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NethermindEth/eigen-wiz/internal/package_handler/testdata"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestManifestValidate(t *testing.T) {
	testDir := t.TempDir()
	testdata.SetupDir(t, "manifests", testDir)

	tests := []struct {
		name      string
		filePath  string
		wantError string
	}{
		{
			name:      "Full OK Manifest",
			filePath:  "full-ok/pkg/manifest.yml",
			wantError: "",
		},
		{
			name:      "Invalid Fields Manifest",
			filePath:  "invalid-fields/pkg/manifest.yml",
			wantError: "Invalid hardware requirements -> invalid fields: hardware_requirements.min_cpu_cores -> (negative value), hardware_requirements.min_ram -> (negative value), hardware_requirements.min_free_space -> (negative value).  Invalid plugin -> invalid fields: plugin.git -> (invalid git url), plugin.image -> (invalid docker image).  ",
		},
		{
			name:      "Minimal Manifest",
			filePath:  "minimal/pkg/manifest.yml",
			wantError: "",
		},
		{
			name:      "Missing Fields Manifest",
			filePath:  "missing-fields/pkg/manifest.yml",
			wantError: "Invalid manifest file -> missing fields: version, node_version, name, upgrade, profiles. ",
		},
		{
			name:      "Missing Fields Manifest in profile",
			filePath:  "missing-fields-profile/pkg/manifest.yml",
			wantError: "Invalid manifest file -> missing fields: version, node_version, name, upgrade.    Invalid profile 1 -> missing fields: name. ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			data, err := os.ReadFile(filepath.Join(testDir, "manifests", tt.filePath))
			if err != nil {
				t.Fatalf("failed reading data from yaml file: %s", err)
			}

			manifest := Manifest{}
			if err := yaml.Unmarshal(data, &manifest); err != nil {
				t.Fatalf("failed unmarshalling yaml: %s", err)
			}

			err = manifest.validate()
			if tt.wantError == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.wantError)
			}
		})
	}
}
