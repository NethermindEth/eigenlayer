package package_handler

import (
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"

	"github.com/NethermindEth/egn/internal/package_handler/testdata"
)

func TestOptionValidate(t *testing.T) {
	afs := afero.NewMemMapFs()
	testDir, err := afero.TempDir(afs, "", "test")
	require.NoError(t, err)
	testdata.SetupDir(t, "options", testDir, afs)

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
			name:     "Check invalid type int with min-max value",
			filePath: "check-invalid-int-validate/pkg/option.yml",
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
			name:     "Check valid type int with validate",
			filePath: "check-valid-int-validate/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check valid type int without max value",
			filePath: "check-valid-int-without-max/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check valid type int without min value",
			filePath: "check-valid-int-without-min/pkg/option.yml",
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
			name:     "Check invalid type port with negative value",
			filePath: "check-negative-port/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check invalid type port with huge value",
			filePath: "check-huge-port/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check invalid type port with zero value",
			filePath: "check-zero-port/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check invalid type port with decimal value",
			filePath: "check-decimal-port/pkg/option.yml",
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
			name:     "Check invalid type float with min-max value",
			filePath: "check-invalid-float-validate/pkg/option.yml",
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
			name:     "Check valid type float with validate",
			filePath: "check-valid-float-validate/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check valid type float without max value",
			filePath: "check-valid-float-without-max/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check valid type float without min value",
			filePath: "check-valid-float-without-min/pkg/option.yml",
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
			name:     "Check valid type uri with invalid scheme",
			filePath: "check-valid-uri-invalid-scheme/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
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
		{
			name:     "Check valid type str with validate",
			filePath: "check-valid-str-validate/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check invalid type str with validate",
			filePath: "check-invalid-str-validate/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check valid type path_file",
			filePath: "check-valid-path-file/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check valid type path_file with validate",
			filePath: "check-valid-path-file-validate/pkg/option.yml",
			want:     InvalidConfError{},
		},
		{
			name:     "Check invalid type path_file",
			filePath: "check-invalid-path-file/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
		{
			name:     "Check invalid type path_file with validate",
			filePath: "check-invalid-path-file-validate/pkg/option.yml",
			want: InvalidConfError{
				invalidFields: []string{"options.default"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			data, err := afero.ReadFile(afs, filepath.Join(testDir, "options", tt.filePath))
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
