package data

import (
	"io"
	"os"
	"testing"

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
			instance, err := NewInstance(tc.path)
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
