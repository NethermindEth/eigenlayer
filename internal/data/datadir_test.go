package data

import (
	"errors"
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

func TestDataDir_AddInstance(t *testing.T) {
	type testCase struct {
		tName    string
		dataDir  *DataDir
		name     string
		url      string
		version  string
		profile  string
		tag      string
		instance *Instance
		err      error
	}
	ts := []testCase{
		func() testCase {
			testDir := t.TempDir()
			dataDir, err := NewDataDir(testDir)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				tName:   "success",
				dataDir: dataDir,
				name:    "mock_avs",
				url:     "https://github.com/NethermindEth/mock-avs",
				version: "v0.1.0",
				profile: "mainnet",
				tag:     "latest",
				instance: &Instance{
					Name:    "mock_avs",
					URL:     "https://github.com/NethermindEth/mock-avs",
					Version: "v0.1.0",
					Profile: "mainnet",
					Tag:     "latest",
					path:    filepath.Join(testDir, "nodes", "mock_avs-latest"),
				},
				err: nil,
			}
		}(),
		func() testCase {
			testDir := t.TempDir()
			dataDir, err := NewDataDir(testDir)
			if err != nil {
				t.Fatal(err)
			}
			err = os.MkdirAll(filepath.Join(testDir, "nodes", "mock_avs-latest"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				tName:    "folder already exists",
				dataDir:  dataDir,
				name:     "mock_avs",
				url:      "https://github.com/NethermindEth/mock-avs",
				version:  "v0.1.0",
				profile:  "mainnet",
				tag:      "latest",
				instance: nil,
				err:      ErrInstanceAlreadyExists,
			}
		}(),
	}
	for _, tc := range ts {
		t.Run(tc.tName, func(t *testing.T) {
			instance, err := tc.dataDir.AddInstance(tc.name, tc.url, tc.version, tc.profile, tc.tag)
			if tc.err == nil {
				assert.NoError(t, err)
				assert.Equal(t, tc.instance, instance)
			} else {
				assert.Nil(t, instance)
				assert.ErrorIs(t, err, tc.err)
			}
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