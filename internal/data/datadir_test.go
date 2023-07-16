package data

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/locker/mocks"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDataDir(t *testing.T) {
	fs := afero.NewOsFs()

	type testCase struct {
		name    string
		path    string
		dataDir *DataDir
		locker  *mocks.MockLocker
		err     error
	}
	ts := []testCase{
		func() testCase {
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

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
					path:   absPath,
					fs:     fs,
					locker: locker,
				},
				locker: locker,
				err:    nil,
			}
		}(),
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			dataDir, err := NewDataDir(tc.path, fs, tc.locker)
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
	fs := afero.NewOsFs()

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
			err := fs.MkdirAll(filepath.Join(path, instancesDir, "mock-avs-default"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			// Create state.json file
			stateFile, err := fs.Create(filepath.Join(path, instancesDir, "mock-avs-default", "state.json"))
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
					path:    filepath.Join(path, instancesDir, "mock-avs-default"),
					fs:      fs,
				},
				err: nil,
			}
		}(),
		func() testCase {
			path := t.TempDir()
			// Create instance dir path
			err := fs.MkdirAll(filepath.Join(path, instancesDir, "mock-avs-default"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			// Create state.json file
			stateFile, err := fs.Create(filepath.Join(path, instancesDir, "mock-avs-default", "state.json"))
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
			err := fs.MkdirAll(filepath.Join(path, instancesDir), 0o755)
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
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			if tc.instance != nil {
				tc.instance.locker = locker
			}

			dataDir, err := NewDataDir(tc.path, fs, locker)
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
	fs := afero.NewOsFs()

	type testCase struct {
		name       string
		path       string
		instance   *Instance
		err        error
		locker     *mocks.MockLocker
		afterCheck func(t *testing.T)
	}
	ts := []testCase{
		func() testCase {
			path := t.TempDir()
			err := fs.MkdirAll(filepath.Join(path, instancesDir, "mock-avs-default"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			return testCase{
				name: "instance already exists",
				path: path,
				instance: &Instance{
					Name:    "mock-avs",
					Tag:     "default",
					URL:     "https://github.com/NethermindEth/mock-avs",
					Version: "v2.0.2",
					Profile: "option-returner",
					fs:      fs,
					locker:  locker,
				},
				err:    ErrInstanceAlreadyExists,
				locker: locker,
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
					fs:      fs,
				},
				err: ErrInvalidInstance,
				afterCheck: func(t *testing.T) {
					assert.NoDirExists(t, filepath.Join(path, instancesDir, "mock-avs-default"))
				},
			}
		}(),
		func() testCase {
			path := t.TempDir()
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)
			locker.EXPECT().New(filepath.Join(path, instancesDir, "mock-avs-default", ".lock")).Return(locker)

			return testCase{
				name: "valid instance",
				path: path,
				instance: &Instance{
					Name:    "mock-avs",
					Tag:     "default",
					URL:     "https://github.com/NethermindEth/mock-avs",
					Version: "v2.0.2",
					Profile: "option-returner",
					fs:      fs,
					locker:  locker,
				},
				err:    nil,
				locker: locker,
				afterCheck: func(t *testing.T) {
					assert.DirExists(t, filepath.Join(path, instancesDir, "mock-avs-default"))
					assert.FileExists(t, filepath.Join(path, instancesDir, "mock-avs-default", "state.json"))
					assert.FileExists(t, filepath.Join(path, instancesDir, "mock-avs-default", ".lock"))
				},
			}
		}(),
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			dataDir, err := NewDataDir(tc.path, fs, tc.locker)
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
			fs := afero.NewMemMapFs()
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			testDir := t.TempDir()
			dataDir, err := NewDataDir(testDir, fs, locker)
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
			fs := afero.NewMemMapFs()
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			testDir := t.TempDir()
			dataDir, err := NewDataDir(testDir, fs, locker)
			if err != nil {
				t.Fatal(err)
			}
			err = fs.MkdirAll(filepath.Join(testDir, "nodes", "mock_avs-latest"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:       "instance dir found",
				dataDir:    dataDir,
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
	fs := afero.NewOsFs()

	type testCase struct {
		name       string
		dataDir    *DataDir
		instanceId string
		err        error
	}
	ts := []testCase{
		func() testCase {
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			testDir := t.TempDir()
			dataDir, err := NewDataDir(testDir, fs, locker)
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
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			testDir := t.TempDir()
			err := os.MkdirAll(filepath.Join(testDir, instancesDir, "mock_avs-latest"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			dataDir, err := NewDataDir(testDir, fs, locker)
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
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			testDir := t.TempDir()
			err := os.MkdirAll(filepath.Join(testDir, instancesDir), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			_, err = os.Create(filepath.Join(testDir, instancesDir, "mock_avs-test"))
			if err != nil {
				t.Fatal(err)
			}
			dataDir, err := NewDataDir(testDir, fs, locker)
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

func TestDataDir_InstancePath(t *testing.T) {
	fs := afero.NewOsFs()

	type testCase struct {
		name       string
		path       string
		instanceId string
		want       string
		wantErr    error
	}
	tests := []testCase{
		func() testCase {
			path := t.TempDir()
			err := fs.MkdirAll(filepath.Join(path, instancesDir, "mock-avs-default"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return testCase{
				name:       "instance dir exists",
				path:       path,
				instanceId: "mock-avs-default",
				want:       filepath.Join(path, instancesDir, "mock-avs-default"),
				wantErr:    nil,
			}
		}(),
		{
			name:       "instance not found",
			path:       t.TempDir(),
			instanceId: "mock-avs-default",
			want:       "",
			wantErr:    ErrInstanceNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			d, err := NewDataDir(tt.path, fs, locker)
			assert.NoError(t, err)
			got, err := d.InstancePath(tt.instanceId)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestDataDir_InitTemp(t *testing.T) {
	fs := afero.NewOsFs()

	type tc struct {
		name    string
		path    string
		id      string
		want    string
		wantErr error
		check   func(t *testing.T)
	}
	tests := []tc{
		func() tc {
			path := t.TempDir()
			return tc{
				name: "empty data dir",
				path: path,
				id:   "temp-dir-id",
				want: filepath.Join(path, tempDir, "temp-dir-id"),
				check: func(t *testing.T) {
					assert.DirExists(t, filepath.Join(path, tempDir, "temp-dir-id"))
				},
			}
		}(),
		func() tc {
			path := t.TempDir()
			err := fs.MkdirAll(filepath.Join(path, tempDir), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return tc{
				name: "empty temp dir",
				path: path,
				id:   "temp-dir-id",
				want: filepath.Join(path, tempDir, "temp-dir-id"),
				check: func(t *testing.T) {
					assert.DirExists(t, filepath.Join(path, tempDir, "temp-dir-id"))
				},
			}
		}(),
		func() tc {
			path := t.TempDir()
			err := fs.MkdirAll(filepath.Join(path, tempDir, "temp-dir-id"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return tc{
				name:    "already exists",
				path:    path,
				id:      "temp-dir-id",
				want:    "",
				wantErr: ErrTempDirAlreadyExists,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			dataDir, err := NewDataDir(tt.path, fs, locker)
			assert.NoError(t, err)
			got, err := dataDir.InitTemp(tt.id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
			if tt.check != nil {
				tt.check(t)
			}
		})
	}
}

func TestDataDir_RemoveTemp(t *testing.T) {
	fs := afero.NewOsFs()

	type tc struct {
		name  string
		path  string
		id    string
		check func(t *testing.T)
	}
	tests := []tc{
		{
			name: "empty data dir",
			path: t.TempDir(),
			id:   "mock-avs-default",
		},
		func() tc {
			path := t.TempDir()
			err := fs.MkdirAll(filepath.Join(path, tempDir), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return tc{
				name: "empty temp dir",
				path: path,
				id:   "mock-avs-default",
			}
		}(),
		func() tc {
			path := t.TempDir()
			err := fs.MkdirAll(filepath.Join(path, tempDir, "temp-dir-id"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return tc{
				name: "temp dir exists",
				path: path,
				id:   "temp-dir-id",
				check: func(t *testing.T) {
					assert.NoDirExists(t, filepath.Join(path, tempDir, "temp-dir-id"))
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			dataDir, err := NewDataDir(tt.path, fs, locker)
			assert.NoError(t, err)
			gotErr := dataDir.RemoveTemp(tt.id)
			assert.NoError(t, gotErr)
			if tt.check != nil {
				tt.check(t)
			}
		})
	}
}

func TestDataDir_TempPath(t *testing.T) {
	fs := afero.NewOsFs()

	type tc struct {
		name    string
		path    string
		id      string
		want    string
		wantErr error
	}
	tests := []tc{
		{
			name:    "empty data dir",
			path:    t.TempDir(),
			id:      "temp-dir-id",
			want:    "",
			wantErr: ErrTempDirDoesNotExist,
		},
		func() tc {
			path := t.TempDir()
			err := fs.MkdirAll(filepath.Join(path, tempDir), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return tc{
				name:    "empty temp dir",
				path:    path,
				id:      "temp-dir-id",
				want:    "",
				wantErr: ErrTempDirDoesNotExist,
			}
		}(),
		func() tc {
			path := t.TempDir()
			err := fs.MkdirAll(filepath.Join(path, tempDir, "temp-dir-id"), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			return tc{
				name:    "temp dir exists",
				path:    path,
				id:      "temp-dir-id",
				want:    filepath.Join(path, tempDir, "temp-dir-id"),
				wantErr: nil,
			}
		}(),
		func() tc {
			path := t.TempDir()
			err := os.MkdirAll(filepath.Join(path, tempDir), 0o755)
			if err != nil {
				t.Fatal(err)
			}
			_, err = os.Create(filepath.Join(path, tempDir, "temp-dir-id"))
			if err != nil {
				t.Fatal(err)
			}
			return tc{
				name:    "not a directory",
				path:    path,
				id:      "temp-dir-id",
				want:    "",
				wantErr: ErrTempIsNotDir,
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			dataDir, err := NewDataDir(tt.path, fs, locker)
			assert.NoError(t, err)
			gotPath, gotErr := dataDir.TempPath(tt.id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, gotErr, tt.wantErr)
			} else {
				assert.NoError(t, gotErr)
				assert.Equal(t, tt.want, gotPath)
			}
		})
	}
}

func TestMonitoringStack(t *testing.T) {
	// Create a memory filesystem
	fs := afero.NewMemMapFs()
	userHome, err := os.UserHomeDir()
	require.NoError(t, err)
	basePath := filepath.Join(userHome, ".local", "share", ".eigen")

	// Create a mock locker
	ctrl := gomock.NewController(t)
	locker := mocks.NewMockLocker(ctrl)
	locker.EXPECT().New(filepath.Join(basePath, "/monitoring", ".lock")).Return(locker).Times(2)

	verify := func(t *testing.T, stack *MonitoringStack) {
		t.Helper()
		assert.Equal(t, filepath.Join(basePath, "/monitoring"), stack.path)
		assert.Equal(t, fs, stack.fs)
		assert.Equal(t, locker, stack.l)

		exists, err := afero.DirExists(fs, filepath.Join(basePath, "/monitoring"))
		assert.NoError(t, err)
		assert.True(t, exists)

		exists, err = afero.Exists(fs, filepath.Join(basePath, "/monitoring", ".lock"))
		assert.NoError(t, err)
		assert.True(t, exists)
	}
	// Create a data dir
	dataDir, err := NewDataDirDefault(fs, locker)
	require.NoError(t, err)

	// Create a monitoring stack
	monitoringStack, err := dataDir.MonitoringStack()
	require.NoError(t, err)
	verify(t, monitoringStack)

	// Try to get a monitoring stack while it does exist
	monitoringStack, err = dataDir.MonitoringStack()
	require.NoError(t, err)
	verify(t, monitoringStack)
}

func TestRemoveMonitoringStack(t *testing.T) {
	// Create monitoring stack
	// Create a memory filesystem
	fs := afero.NewMemMapFs()

	// Create a mock locker
	ctrl := gomock.NewController(t)
	locker := mocks.NewMockLocker(ctrl)
	locker.EXPECT().New(filepath.Join("/monitoring", ".lock")).Return(locker)

	// Create a data dir
	dataDir, err := NewDataDir("/", fs, locker)
	require.NoError(t, err)

	// Create a monitoring stack
	_, err = dataDir.MonitoringStack()
	require.NoError(t, err)

	// Remove monitoring stack
	err = dataDir.RemoveMonitoringStack()
	require.NoError(t, err)

	exists, err := afero.DirExists(fs, filepath.Join("/monitoring"))
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestRemoveMonitoringStackError(t *testing.T) {
	// Create monitoring stack
	// Create a memory filesystem
	fs := afero.NewMemMapFs()

	// Create a mock locker
	ctrl := gomock.NewController(t)
	locker := mocks.NewMockLocker(ctrl)

	// Create a data dir
	dataDir, err := NewDataDir("/", fs, locker)
	require.NoError(t, err)

	// Remove monitoring stack
	err = dataDir.RemoveMonitoringStack()
	require.ErrorIs(t, err, ErrMonitoringStackNotFound)
}
