package data

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInstance(t *testing.T) {
	type testCase struct {
		name     string
		path     string
		instance *Instance
		errStr   string
	}
	ts := []testCase{
		func() testCase {
			testDir := t.TempDir()
			return testCase{
				name:     "empty directory",
				path:     testDir,
				instance: nil,
				errStr:   fmt.Sprintf("open %s: no such file or directory", testDir+"/state.json"),
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
				errStr:   "unexpected end of JSON input",
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
				errStr:   "",
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
				errStr: "",
			}
		}(),
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			instance, err := NewInstance(tc.path)
			if tc.errStr != "" {
				assert.EqualError(t, err, tc.errStr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tc.instance, instance)
		})
	}
}

func TestInstance_Init(t *testing.T) {
	ts := []struct {
		name     string
		instance *Instance
		check    func(t *testing.T, initErr error, instanceDir fs.FS)
	}{
		{
			name: ".lock is created as regular file",
			instance: &Instance{
				Name: "test_name",
				Tag:  "test_tag",
			},
			check: func(t *testing.T, initErr error, instanceDir fs.FS) {
				assert.NoError(t, initErr)
				lockFile, err := instanceDir.Open(".lock")
				assert.NoError(t, err)
				lockFileInfo, err := lockFile.Stat()
				assert.NoError(t, err)
				assert.True(t, !lockFileInfo.IsDir())
				assert.True(t, lockFileInfo.Mode().IsRegular())
			},
		},
		{
			name: "state.json is created as regular file",
			instance: &Instance{
				Name: "test_name",
				Tag:  "test_tag",
			},
			check: func(t *testing.T, initErr error, instanceDir fs.FS) {
				assert.NoError(t, initErr)
				stateFile, err := instanceDir.Open("state.json")
				assert.NoError(t, err)
				stateFileInfo, err := stateFile.Stat()
				assert.NoError(t, err)
				assert.True(t, !stateFileInfo.IsDir())
				assert.True(t, stateFileInfo.Mode().IsRegular())
			},
		},
		{
			name: "state.json has correct data",
			instance: &Instance{
				Name:    "test_name",
				Tag:     "test_tag",
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: "v0.1.0",
				Profile: "mainnet",
			},
			check: func(t *testing.T, initErr error, instanceDir fs.FS) {
				assert.NoError(t, initErr)
				stateFile, err := instanceDir.Open("state.json")
				assert.NoError(t, err)
				stateData, err := io.ReadAll(stateFile)
				assert.NoError(t, err)
				assert.Equal(t, `{"name":"test_name","url":"https://github.com/NethermindEth/mock-avs","version":"v0.1.0","profile":"mainnet","tag":"test_tag"}`, string(stateData))
			},
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			instanceDirPath := t.TempDir()
			err := tc.instance.Init(instanceDirPath)
			tc.check(t, err, os.DirFS(instanceDirPath))
		})
	}
}
