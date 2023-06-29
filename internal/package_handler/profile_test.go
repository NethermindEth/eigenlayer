package package_handler

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"github.com/NethermindEth/egn/internal/package_handler/testdata"
)

func TestOptionValidate(t *testing.T) {
	testDir := t.TempDir()
	testdata.SetupDir(t, "options", testDir)

	tests := []struct {
		name     string
		filePath string
		want     InvalidConfError
	}{
		{
			name:     "Full OK Option",
			filePath: "full-ok/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Invalid Fields Option",
			filePath: "invalid-fields/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Missing Fields Option",
			filePath: "missing-fields/pkg/option.yml",
			want: InvalidConfError{
				missingFields: []string{"options.help"},
			},
		},
		{
			name:     "Missing and Invalid Fields Option",
			filePath: "missing-invalid-fields/pkg/option.yml",
			want: InvalidConfError{
				missingFields: []string{"options.name", "options.target", "options.help"},
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Full Missing Fields Option",
			filePath: "full-missing/pkg/option.yml",
			want: InvalidConfError{
				missingFields: []string{"options.name", "options.target", "options.type", "options.help"},
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Invalid Type in Option",
			filePath: "invalid-type/pkg/option.yml",
			want: InvalidConfError{
				missingFields: []string{"options.target", "options.help"},
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check invalid type float",
			filePath: "check-type-float/pkg/option.yml",
			want: InvalidConfError{
				missingFields: []string{"options.target"},
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check invalid type id",
			filePath: "check-type-id/pkg/option.yml",
			want: InvalidConfError{
				missingFields: []string{"options.target"},
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check invalid type uri",
			filePath: "check-type-uri/pkg/option.yml",
			want: InvalidConfError{
				missingFields: []string{"options.target"},
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check type select",
			filePath: "check-type-select/pkg/option.yml",
			want: InvalidConfError{
				missingFields: []string{"options.target", "options.validate"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			data, err := os.ReadFile(filepath.Join(testDir, "options", tt.filePath))
			if err != nil {
				t.Fatalf("failed reading data from yaml file: %s", err)
			}

			option := Option{}
			if err := yaml.Unmarshal(data, &option); err != nil {
				t.Fatalf("failed unmarshalling yaml: %s", err)
			}

			got := option.validate()
			assert.Equal(t, tt.want, got)
		})
	}
}
