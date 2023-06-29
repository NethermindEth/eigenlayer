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
			name:     "Check invalid type int",
			filePath: "check-invalid-int/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check valid type int",
			filePath: "check-valid-int/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check invalid type port",
			filePath: "check-invalid-port/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check invalid type bool",
			filePath: "check-invalid-bool/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check invalid type float",
			filePath: "check-invalid-float/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check valid type float",
			filePath: "check-valid-float/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check invalid type id",
			filePath: "check-invalid-id/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check valid type id",
			filePath: "check-valid-id/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check invalid type uri",
			filePath: "check-invalid-uri/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check invalid type uri with scheme",
			filePath: "check-invalid-uri-scheme/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check valid type uri",
			filePath: "check-valid-uri/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check type select",
			filePath: "check-type-select/pkg/option.yml",
			want: InvalidConfError{
				missingFields: []string{"options.validate"},
			},
		},
		{
			name:     "Check type select with validate",
			filePath: "check-select-validate/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check type str",
			filePath: "check-type-str/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check valid type path_dir",
			filePath: "check-valid-path_dir/pkg/option.yml",
			want:     InvalidConfError{},
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
