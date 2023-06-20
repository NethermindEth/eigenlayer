package data

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDataDir(t *testing.T) {
	type testCase struct {
		name    string
		path    string
		dataDir *DataDir
		err     error
	}
	ts := []testCase{
		func() testCase {
			testDir := t.TempDir()
			wd, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}
			relativePth, err := filepath.Rel(wd, testDir)
			if err != nil {
				t.Fatal(err)
			}
			absPath, err := filepath.Abs(relativePth)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name: "path to absolute",
				path: testDir,
				dataDir: &DataDir{
					path: absPath,
				},
				err: nil,
			}
		}(),
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			dataDir, err := NewDataDir(tc.path)
			if tc.err != nil {
				assert.Nil(t, dataDir)
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.dataDir, dataDir)
			}
		})
	}
}

func TestDataDir_Instance(t *testing.T) {
	type testCase struct {
		name       string
		instanceId string
		path       string
		instance   *Instance
		err        error
	}
	ts := []testCase{
		func() testCase {
			path := t.TempDir()
			// Create instance dir path
			err := os.MkdirAll(filepath.Join(path, "nodes", "mock-avs-default"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			// Create state.json file
			stateFile, err := os.Create(filepath.Join(path, "nodes", "mock-avs-default", "state.json"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = stateFile.WriteString(`{"name":"mock-avs","url":"https://github.com/NethermindEth/mock-avs","version":"v2.0.2","profile":"option-returner","tag":"default"}`)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:       "valid instance",
				instanceId: "mock-avs-default",
				path:       path,
				instance: &Instance{
					Name:    "mock-avs",
					URL:     "https://github.com/NethermindEth/mock-avs",
					Version: "v2.0.2",
					Tag:     "default",
					Profile: "option-returner",
					path:    filepath.Join(path, "nodes", "mock-avs-default"),
				},
				err: nil,
			}
		}(),
		func() testCase {
			path := t.TempDir()
			// Create instance dir path
			err := os.MkdirAll(filepath.Join(path, "nodes", "mock-avs-default"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			// Create state.json file
			stateFile, err := os.Create(filepath.Join(path, "nodes", "mock-avs-default", "state.json"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = stateFile.WriteString(`{"url":"https://github.com/NethermindEth/mock-avs","version":"v2.0.2","profile":"option-returner","tag":"default"}`)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:       "invalid instance, state without name field",
				instanceId: "mock-avs-default",
				path:       path,
				instance:   nil,
				err:        ErrInvalidInstance,
			}
		}(),
		func() testCase {
			path := t.TempDir()
			// Create nodes dir
			err := os.MkdirAll(filepath.Join(path, "nodes"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:       "instance not found",
				instanceId: "mock-avs-default",
				path:       path,
				instance:   nil,
				err:        ErrInvalidInstanceDir,
			}
		}(),
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			dataDir, err := NewDataDir(tc.path)
			assert.NoError(t, err)
			instance, err := dataDir.Instance(tc.instanceId)
			if tc.err != nil {
				assert.Nil(t, instance)
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.instance, instance)
			}
		})
	}
}

func TestDataDir_InitInstance(t *testing.T) {
	type testCase struct {
		name       string
		path       string
		instance   *Instance
		err        error
		afterCheck func(t *testing.T)
	}
	ts := []testCase{
		func() testCase {
			path := t.TempDir()
			err := os.MkdirAll(filepath.Join(path, "nodes", "mock-avs-default"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name: "instance already exists",
				path: path,
				instance: &Instance{
					Name:    "mock-avs",
					Tag:     "default",
					URL:     "https://github.com/NethermindEth/mock-avs",
					Version: "v2.0.2",
					Profile: "option-returner",
				},
				err: ErrInstanceAlreadyExists,
			}
		}(),
		func() testCase {
			path := t.TempDir()
			return testCase{
				name: "invalid instance, state without url field",
				path: path,
				instance: &Instance{
					Name:    "mock-avs",
					Tag:     "default",
					Version: "v2.0.2",
					Profile: "option-returner",
				},
				err: ErrInvalidInstance,
				afterCheck: func(t *testing.T) {
					assert.NoDirExists(t, filepath.Join(path, "nodes", "mock-avs-default"))
				},
			}
		}(),
		func() testCase {
			path := t.TempDir()
			return testCase{
				name: "valid instance",
				path: path,
				instance: &Instance{
					Name:    "mock-avs",
					Tag:     "default",
					URL:     "https://github.com/NethermindEth/mock-avs",
					Version: "v2.0.2",
					Profile: "option-returner",
				},
				err: nil,
				afterCheck: func(t *testing.T) {
					assert.DirExists(t, filepath.Join(path, "nodes", "mock-avs-default"))
					assert.FileExists(t, filepath.Join(path, "nodes", "mock-avs-default", "state.json"))
					assert.FileExists(t, filepath.Join(path, "nodes", "mock-avs-default", ".lock"))
				},
			}
		}(),
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			dataDir, err := NewDataDir(tc.path)
			assert.NoError(t, err)
			err = dataDir.InitInstance(tc.instance)
			if tc.err != nil {
				assert.ErrorIs(t, err, tc.err)
			} else {
				assert.NoError(t, err)
			}
			if tc.afterCheck != nil {
				tc.afterCheck(t)
			}
		})
	}
}

func TestDataDir_HasInstance(t *testing.T) {
	type testCase struct {
		name       string
		dataDir    *DataDir
		instanceId string
		has        bool
	}
	ts := []testCase{
		func() testCase {
			testDir := t.TempDir()
			dataDir, err := NewDataDir(testDir)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:       "instance dir not found",
				dataDir:    dataDir,
				instanceId: "mock_avs-latest",
				has:        false,
			}
		}(),
		func() testCase {
			testDir := t.TempDir()
			err := os.MkdirAll(filepath.Join(testDir, "nodes", "mock_avs-latest"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:       "instance dir found",
				dataDir:    &DataDir{path: testDir},
				instanceId: "mock_avs-latest",
				has:        true,
			}
		}(),
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			has := tc.dataDir.HasInstance(tc.instanceId)
			assert.Equal(t, tc.has, has)
		})
	}
}

func TestDataDir_RemoveInstance(t *testing.T) {
	type testCase struct {
		name       string
		dataDir    *DataDir
		instanceId string
		err        error
	}
	ts := []testCase{
		func() testCase {
			testDir := t.TempDir()
			dataDir, err := NewDataDir(testDir)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:       "instance dir not found",
				dataDir:    dataDir,
				instanceId: "mock_avs-latest",
				err:        fmt.Errorf("%w: mock_avs-latest", ErrInstanceNotFound),
			}
		}(),
		func() testCase {
			testDir := t.TempDir()
			err := os.MkdirAll(filepath.Join(testDir, "nodes", "mock_avs-latest"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			dataDir, err := NewDataDir(testDir)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:       "success",
				dataDir:    dataDir,
				instanceId: "mock_avs-latest",
				err:        nil,
			}
		}(),
		func() testCase {
			testDir := t.TempDir()
			err := os.MkdirAll(filepath.Join(testDir, "nodes"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			_, err = os.Create(filepath.Join(testDir, "nodes", "mock_avs-test"))
			if err != nil {
				t.Fatal(err)
			}
			dataDir, err := NewDataDir(testDir)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:       "file instead of directory",
				dataDir:    dataDir,
				instanceId: "mock_avs-test",
				err:        errors.New("mock_avs-test is not a directory"),
			}
		}(),
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.dataDir.RemoveInstance(tc.instanceId)
			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}
