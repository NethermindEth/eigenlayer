package data

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/NethermindEth/egn/internal/data/testdata"
	"github.com/NethermindEth/egn/internal/locker/mocks"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestNewInstance(t *testing.T) {
	fs := afero.NewOsFs()

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
			_, err := fs.Create(testDir + "/state.json")
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
			stateFile, err := fs.Create(testDir + "/state.json")
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
			stateFile, err := fs.Create(testDir + "/state.json")
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
			stateFile, err := fs.Create(testDir + "/state.json")
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
		func() testCase {
			testDir := t.TempDir()
			stateFile, err := fs.Create(testDir + "/state.json")
			if err != nil {
				t.Fatal(err)
			}
			defer stateFile.Close()
			_, err = io.WriteString(stateFile, `{"name":"test_name","url":"https://github.com/NethermindEth/mock-avs","version":"v0.1.0","profile":"mainnet","tag":"test_tag","plugin":{"image":"nethermind/egn-plugin:latest"}}`)
			if err != nil {
				t.Fatal(err)
			}

			return testCase{
				name: "with plugin image",
				path: testDir,
				instance: &Instance{
					Name:    "test_name",
					Tag:     "test_tag",
					URL:     "https://github.com/NethermindEth/mock-avs",
					Version: "v0.1.0",
					Profile: "mainnet",
					Plugin: &Plugin{
						Image: "nethermind/egn-plugin:latest",
					},
					path: testDir,
				},
				err: nil,
			}
		}(),
		func() testCase {
			testDir := t.TempDir()
			stateFile, err := fs.Create(testDir + "/state.json")
			if err != nil {
				t.Fatal(err)
			}
			defer stateFile.Close()
			_, err = io.WriteString(stateFile, `{"name":"test_name","url":"https://github.com/NethermindEth/mock-avs","version":"v0.1.0","profile":"mainnet","tag":"test_tag","plugin":{"build_from":"https://github.com/NethermindEth/mock-avs.git#main:plugin"}}`)
			if err != nil {
				t.Fatal(err)
			}

			return testCase{
				name: "with plugin git url",
				path: testDir,
				instance: &Instance{
					Name:    "test_name",
					Tag:     "test_tag",
					URL:     "https://github.com/NethermindEth/mock-avs",
					Version: "v0.1.0",
					Profile: "mainnet",
					Plugin: &Plugin{
						BuildFrom: "https://github.com/NethermindEth/mock-avs.git#main:plugin",
					},
					path: testDir,
				},
				err: nil,
			}
		}(),
		func() testCase {
			testDir := t.TempDir()
			stateFile, err := fs.Create(testDir + "/state.json")
			if err != nil {
				t.Fatal(err)
			}
			defer stateFile.Close()
			_, err = io.WriteString(stateFile, `{"name":"test_name","url":"https://github.com/NethermindEth/mock-avs","version":"v0.1.0","profile":"mainnet","tag":"test_tag","plugin":{}}`)
			if err != nil {
				t.Fatal(err)
			}

			return testCase{
				name: "error, empty plugin",
				path: testDir,
				instance: &Instance{
					Name:    "test_name",
					Tag:     "test_tag",
					URL:     "https://github.com/NethermindEth/mock-avs",
					Version: "v0.1.0",
					Profile: "mainnet",
					Plugin:  &Plugin{},
					path:    testDir,
				},
				err: ErrInvalidInstance,
			}
		}(),
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)
			if tc.instance != nil {
				tc.instance.fs = fs
				tc.instance.locker = locker
			}

			instance, err := newInstance(tc.path, fs, locker)
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
	// TODO: Use always the latest version of mock-avs
	ts := []struct {
		name      string
		instance  *Instance
		stateJSON []byte
		err       error
		mocker    func(path string, locker *mocks.MockLocker)
	}{
		{
			name:      "invalid instance",
			instance:  &Instance{},
			stateJSON: nil,
			err:       ErrInvalidInstance,
		},
		{
			name: "valid instance",
			instance: &Instance{
				Name:    "test_name",
				Tag:     "test_tag",
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: "v2.1.0",
				Profile: "option-returner",
				MonitoringTargets: MonitoringTargets{
					Targets: []MonitoringTarget{
						{
							Service: "main-service",
							Port:    "8080",
							Path:    "/metrics",
						},
					},
				},
			},
			stateJSON: []byte(`{"name":"test_name","url":"https://github.com/NethermindEth/mock-avs","version":"v2.1.0","profile":"option-returner","tag":"test_tag","monitoring":{"targets":[{"service":"main-service","port":"8080","path":"/metrics"}]}}`),
			mocker: func(path string, locker *mocks.MockLocker) {
				locker.EXPECT().New(filepath.Join(path, ".lock")).Return(locker)
			},
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()

			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			path := t.TempDir()

			if tc.mocker != nil {
				tc.mocker(path, locker)
			}

			err := tc.instance.init(path, fs, locker)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				stateFile, err := fs.Open(filepath.Join(path, "state.json"))
				assert.NoError(t, err)
				stateData, err := io.ReadAll(stateFile)
				assert.NoError(t, err)
				assert.Equal(t, tc.stateJSON, stateData)
			}
		})
	}
}

func TestInstance_Setup(t *testing.T) {
	instancePath := t.TempDir()
	fs := afero.NewOsFs()

	// Create a mock locker
	ctrl := gomock.NewController(t)
	locker := mocks.NewMockLocker(ctrl)
	gomock.InOrder(
		locker.EXPECT().New(filepath.Join(instancePath, ".lock")).Return(locker),
		locker.EXPECT().Lock().Return(nil),
		locker.EXPECT().Locked().Return(true),
		locker.EXPECT().Unlock().Return(nil),
	)

	i := Instance{
		Name:    "mock-avs",
		URL:     "https://github.com/NethermindEth/mock-avs",
		Version: "v2.0.2",
		Profile: "option-returner",
		Tag:     "test-tag",
	}
	err := i.init(instancePath, fs, locker)
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
