package data

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/NethermindEth/egn/internal/data/testdata"
	"github.com/stretchr/testify/assert"
)

func TestNewInstance(t *testing.T) {
	type testCase struct {
		name     string
		path     string
		instance *Instance
		err      error
	}
	ts := []testCase{
		func() testCase {
			testDir := t.TempDir()
			return testCase{
				name:     "empty directory",
				path:     testDir,
				instance: nil,
				err:      ErrInvalidInstanceDir,
			}
		}(),
		func() testCase {
			testDir := t.TempDir()
			_, err := os.Create(testDir + "/state.json")
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:     "empty state file",
				path:     testDir,
				instance: &Instance{path: testDir},
				err:      ErrInvalidInstance,
			}
		}(),
		func() testCase {
			testDir := t.TempDir()
			stateFile, err := os.Create(testDir + "/state.json")
			if err != nil {
				t.Fatal(err)
			}
			defer stateFile.Close()
			_, err = io.WriteString(stateFile, "{}")
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:     "valid state file (empty state)",
				path:     testDir,
				instance: &Instance{path: testDir},
				err:      ErrInvalidInstance,
			}
		}(),
		func() testCase {
			testDir := t.TempDir()
			stateFile, err := os.Create(testDir + "/state.json")
			if err != nil {
				t.Fatal(err)
			}
			defer stateFile.Close()
			_, err = io.WriteString(stateFile, `{"name":"test_name","url":"https://github.com/NethermindEth/mock-avs","version":"v0.1.0","profile":"mainnet","tag":"test_tag"}`)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name: "valid state file",
				path: testDir,
				instance: &Instance{
					Name:    "test_name",
					Tag:     "test_tag",
					URL:     "https://github.com/NethermindEth/mock-avs",
					Version: "v0.1.0",
					Profile: "mainnet",
					path:    testDir,
				},
				err: nil,
			}
		}(),
		func() testCase {
			testDir := t.TempDir()
			stateFile, err := os.Create(testDir + "/state.json")
			if err != nil {
				t.Fatal(err)
			}
			defer stateFile.Close()
			_, err = io.WriteString(stateFile, `{"name":"test_name","url":"https://github.com/NethermindEth/mock-avs","version":"v0.1.0"}`)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:     "invalid state file, missing fields",
				path:     testDir,
				instance: nil,
				err:      ErrInvalidInstance,
			}
		}(),
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			instance, err := newInstance(tc.path)
			if tc.err != nil {
				assert.Nil(t, instance)
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.Equal(t, *tc.instance, *instance)
				assert.NoError(t, err)
			}
		})
	}
}

func TestInstance_Init(t *testing.T) {
	ts := []struct {
		name      string
		instance  *Instance
		path      string
		stateJSON []byte
		err       error
	}{
		{
			name:      "invalid instance",
			instance:  &Instance{},
			path:      t.TempDir(),
			stateJSON: nil,
			err:       ErrInvalidInstance,
		},
		{
			name: "valid instance",
			instance: &Instance{
				Name:    "test_name",
				Tag:     "test_tag",
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: "v2.0.1",
				Profile: "option-returner",
			},
			path:      t.TempDir(),
			stateJSON: []byte(`{"name":"test_name","url":"https://github.com/NethermindEth/mock-avs","version":"v2.0.1","profile":"option-returner","tag":"test_tag"}`),
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.instance.init(tc.path)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				stateFile, err := os.Open(filepath.Join(tc.path, "state.json"))
				assert.NoError(t, err)
				stateData, err := io.ReadAll(stateFile)
				assert.NoError(t, err)
				assert.Equal(t, tc.stateJSON, stateData)
			}
		})
	}
}

func TestInstance_Setup(t *testing.T) {
	i := Instance{
		Name:    "mock-avs",
		URL:     "https://github.com/NethermindEth/mock-avs",
		Version: "v2.0.2",
		Profile: "option-returner",
		Tag:     "test-tag",
	}
	instancePath := t.TempDir()
	err := i.init(instancePath)
	if err != nil {
		t.Fatal(err)
	}
	env := map[string]string{
		"VAR_1": "value-1",
	}
	profileFs := testdata.SetupProfileFS(t, "option-returner")

	err = i.Setup(env, profileFs)
	assert.NoError(t, err)

	assert.NoFileExists(t, filepath.Join(instancePath, "profile.yml"))
	assert.FileExists(t, filepath.Join(instancePath, ".env"))
	assert.FileExists(t, filepath.Join(instancePath, "docker-compose.yml"))
	assert.DirExists(t, filepath.Join(instancePath, "src"))

	envFile, err := os.Open(filepath.Join(instancePath, ".env"))
	assert.NoError(t, err)
	envData, err := io.ReadAll(envFile)
	assert.NoError(t, err)
	assert.Equal(t, []byte("VAR_1=value-1\n"), envData)
}
