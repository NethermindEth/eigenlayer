package prometheus

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/data"
	"github.com/NethermindEth/eigenlayer/internal/locker/mocks"
	"github.com/NethermindEth/eigenlayer/pkg/monitoring"
	"github.com/NethermindEth/eigenlayer/pkg/monitoring/services/types"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestInit(t *testing.T) {
	// Create an in-memory filesystem
	afs := afero.NewMemMapFs()

	// Create a mock locker
	ctrl := gomock.NewController(t)
	locker := mocks.NewMockLocker(ctrl)

	// Expect the lock to be acquired
	locker.EXPECT().New("/monitoring/.lock").Return(locker)

	// Create a new DataDir with the in-memory filesystem
	dataDir, err := data.NewDataDir("/", afs, locker)
	require.NoError(t, err)
	stack, err := dataDir.MonitoringStack()
	require.NoError(t, err)

	tests := []struct {
		name    string
		options types.ServiceOptions
		wantErr bool
	}{
		{
			name: "ok",
			options: types.ServiceOptions{
				Dotenv: map[string]string{
					"PROM_PORT": "9999",
				},
				Stack: stack,
			},
		},
		{
			name: "missing prometheus port",
			options: types.ServiceOptions{
				Dotenv: map[string]string{},
				Stack:  stack,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prometheus := NewPrometheus()
			err := prometheus.Init(tt.options)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, stack, prometheus.stack)
				assert.Equal(t, tt.options.Dotenv["PROM_PORT"], prometheus.port)
			}
		})
	}
}

func TestInitError(t *testing.T) {
	tests := []struct {
		name   string
		dotenv map[string]string
	}{
		{
			name: "empty port",
			dotenv: map[string]string{
				"PROM_PORT": "",
			},
		},
		{
			name:   "missing port",
			dotenv: map[string]string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an in-memory filesystem
			afs := afero.NewMemMapFs()

			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mocks.NewMockLocker(ctrl)

			// Expect the lock to be acquired
			locker.EXPECT().New("/monitoring/.lock").Return(locker)

			// Create a new DataDir with the in-memory filesystem
			dataDir, err := data.NewDataDir("/", afs, locker)
			require.NoError(t, err)
			stack, err := dataDir.MonitoringStack()
			require.NoError(t, err)

			// Create a new Prometheus service
			prometheus := NewPrometheus()
			err = prometheus.Init(types.ServiceOptions{
				Stack:  stack,
				Dotenv: tt.dotenv,
			})

			assert.Error(t, err)
		})
	}
}

func TestDotEnv(t *testing.T) {
	// Create a new Prometheus service
	prometheus := NewPrometheus()
	// Verify the dotEnv
	assert.EqualValues(t, dotEnv, prometheus.DotEnv())
}

func TestSetup(t *testing.T) {
	okLocker := func(t *testing.T) *mocks.MockLocker {
		// Create a mock locker
		ctrl := gomock.NewController(t)
		locker := mocks.NewMockLocker(ctrl)

		// Expect the lock to be acquired
		gomock.InOrder(
			locker.EXPECT().New("/monitoring/.lock").Return(locker),
			locker.EXPECT().Lock().Return(nil),
			locker.EXPECT().Locked().Return(true),
			locker.EXPECT().Unlock().Return(nil),
		)
		gomock.InOrder(
			locker.EXPECT().Lock().Return(nil),
			locker.EXPECT().Locked().Return(true),
			locker.EXPECT().Unlock().Return(nil),
		)
		return locker
	}
	onlyNewLocker := func(t *testing.T) *mocks.MockLocker {
		// Create a mock locker
		ctrl := gomock.NewController(t)
		locker := mocks.NewMockLocker(ctrl)

		// Expect the lock to be acquired
		locker.EXPECT().New("/monitoring/.lock").Return(locker)
		return locker
	}

	tests := []struct {
		name    string
		mocker  func(t *testing.T) *mocks.MockLocker
		options map[string]string
		targets []string
		wantErr bool
	}{
		{
			name:   "ok",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
			},
		},
		{
			name:   "missing node exporter port",
			mocker: onlyNewLocker,
			options: map[string]string{
				"PROM_PORT": "9999",
			},
			wantErr: true,
		},
		{
			name:   "empty node exporter port",
			mocker: onlyNewLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "",
			},
			wantErr: true,
		},
		{
			name: "lock error",
			mocker: func(t *testing.T) *mocks.MockLocker {
				// Create a mock locker
				ctrl := gomock.NewController(t)
				locker := mocks.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New("/monitoring/.lock").Return(locker),
					locker.EXPECT().Lock().Return(fmt.Errorf("error")),
				)
				return locker
			},
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			wantErr: true,
		},
		{
			name: "unlock error",
			mocker: func(t *testing.T) *mocks.MockLocker {
				// Create a mock locker
				ctrl := gomock.NewController(t)
				locker := mocks.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New("/monitoring/.lock").Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(false),
				)
				return locker
			},
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an in-memory filesystem
			afs := afero.NewMemMapFs()

			// Create a new DataDir with the in-memory filesystem
			dataDir, err := data.NewDataDir("/", afs, tt.mocker(t))
			require.NoError(t, err)
			stack, err := dataDir.MonitoringStack()
			require.NoError(t, err)

			// Create a new Prometheus service
			prometheus := NewPrometheus()
			err = prometheus.Init(types.ServiceOptions{
				Stack:  stack,
				Dotenv: tt.options,
			})
			require.NoError(t, err)

			// Setup the Prometheus service
			err = prometheus.Setup(tt.options)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
				ok, err := afero.Exists(afs, "/monitoring/prometheus/prometheus.yml")
				assert.True(t, ok)
				assert.NoError(t, err)

				// Read the prom.yml file
				var prom Config
				promYml, err := afero.ReadFile(afs, "/monitoring/prometheus/prometheus.yml")
				assert.NoError(t, err)
				err = yaml.Unmarshal(promYml, &prom)
				assert.NoError(t, err)

				// Check the Prometheus initial targets
				for i := 0; i < len(tt.targets); i++ {
					assert.Equal(t, tt.targets[i], prom.ScrapeConfigs[i].JobName)
					assert.Equal(t, tt.targets[i], prom.ScrapeConfigs[i].StaticConfigs[0].Targets[0])
				}
			}
		})
	}
}

func TestAddTarget(t *testing.T) {
	okLocker := func(t *testing.T, times int) *mocks.MockLocker {
		// Create a mock locker
		ctrl := gomock.NewController(t)
		locker := mocks.NewMockLocker(ctrl)

		// Expect the lock to be acquired
		gomock.InOrder(
			locker.EXPECT().New("/monitoring/.lock").Return(locker),
			locker.EXPECT().Lock().Return(nil),
			locker.EXPECT().Locked().Return(true),
			locker.EXPECT().Unlock().Return(nil),
		)
		for i := 0; i < times*2+1; i++ {
			gomock.InOrder(
				locker.EXPECT().Lock().Return(nil),
				locker.EXPECT().Locked().Return(true),
				locker.EXPECT().Unlock().Return(nil),
			)
		}

		return locker
	}

	tests := []struct {
		name        string
		mocker      func(t *testing.T, times int) *mocks.MockLocker
		options     map[string]string
		toAdd       []string
		targets     []string
		badEndpoint bool
		wantErr     bool
	}{
		{
			name:   "ok, 1 target",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []string{
				"http://localhost:8000",
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
				"localhost:8000",
			},
		},
		{
			name:   "ok, 2 targets",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []string{
				"http://localhost:8000",
				"http://168.0.0.66:8001",
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
				"localhost:8000",
				"168.0.0.66:8001",
			},
		},
		{
			name: "ok, already existing target",
			mocker: func(t *testing.T, times int) *mocks.MockLocker {
				// Create a mock locker
				ctrl := gomock.NewController(t)
				locker := mocks.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New("/monitoring/.lock").Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				for i := 0; i < times+1; i++ {
					gomock.InOrder(
						locker.EXPECT().Lock().Return(nil),
						locker.EXPECT().Locked().Return(true),
						locker.EXPECT().Unlock().Return(nil),
					)
				}

				return locker
			},
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
			},
		},
		{
			name:   "bad endpoint",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []string{
				"http://localhost:8000",
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
				"localhost:8000",
			},
			badEndpoint: true,
		},
		{
			name: "lock error",
			mocker: func(t *testing.T, times int) *mocks.MockLocker {
				// Create a mock locker
				ctrl := gomock.NewController(t)
				locker := mocks.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New("/monitoring/.lock").Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				locker.EXPECT().Lock().Return(fmt.Errorf("error"))
				return locker
			},
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []string{
				"http://localhost:8000",
			},
			wantErr: true,
		},
		{
			name: "unlock error",
			mocker: func(t *testing.T, times int) *mocks.MockLocker {
				// Create a mock locker
				ctrl := gomock.NewController(t)
				locker := mocks.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New("/monitoring/.lock").Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				gomock.InOrder(
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(false),
				)
				return locker
			},
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []string{
				"http://localhost:8000",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an in-memory filesystem
			afs := afero.NewMemMapFs()

			// Create a new DataDir with the in-memory filesystem
			dataDir, err := data.NewDataDir("/", afs, tt.mocker(t, len(tt.toAdd)))
			require.NoError(t, err)
			stack, err := dataDir.MonitoringStack()
			require.NoError(t, err)

			// Create a new Prometheus service
			prometheus := NewPrometheus()
			err = prometheus.Init(types.ServiceOptions{
				Stack:  stack,
				Dotenv: tt.options,
			})
			require.NoError(t, err)

			// Setup the Prometheus service
			err = prometheus.Setup(tt.options)
			require.NoError(t, err)

			if !tt.badEndpoint {
				// Setup mock http server
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Listen for the POST request to /-/reload
					if r.URL.Path == "/-/reload" && r.Method == http.MethodPost {
						// All good
						w.WriteHeader(http.StatusOK)
					} else if r.Method != http.MethodGet {
						// Unexpected method
						w.WriteHeader(http.StatusMethodNotAllowed)
					} else {
						// Unexpected path
						w.WriteHeader(http.StatusNotFound)
					}
				}))
				defer server.Close()
				split := strings.Split(server.URL, ":")
				host, port := split[1][2:], split[2]
				prometheus.containerIP = host
				prometheus.port = port
			}

			// Add the targets
			for _, target := range tt.toAdd {
				id := strings.TrimPrefix(target, "http://")
				err = prometheus.AddTarget(target, id)
				if tt.wantErr || tt.badEndpoint {
					require.Error(t, err)
					return
				}
				assert.NoError(t, err)
			}
			// Read the prom.yml file
			var prom Config
			promYml, err := afero.ReadFile(afs, "/monitoring/prometheus/prometheus.yml")
			assert.NoError(t, err)
			err = yaml.Unmarshal(promYml, &prom)
			assert.NoError(t, err)

			// Check the Prometheus targets
			for i, target := range tt.targets {
				if i == 0 {
					// Skip node exporter
					continue
				}
				assert.Equal(t, target, prom.ScrapeConfigs[i].StaticConfigs[0].Targets[0])
				assert.Equal(t, target, prom.ScrapeConfigs[i].JobName)
				assert.Equal(t, target, prom.ScrapeConfigs[i].StaticConfigs[0].Labels["instanceID"])
			}
		})
	}
}

func TestRemoveTarget(t *testing.T) {
	okLocker := func(t *testing.T, times int) *mocks.MockLocker {
		// Create a mock locker
		ctrl := gomock.NewController(t)
		locker := mocks.NewMockLocker(ctrl)

		// Expect the lock to be acquired
		gomock.InOrder(
			locker.EXPECT().New("/monitoring/.lock").Return(locker),
			locker.EXPECT().Lock().Return(nil),
			locker.EXPECT().Locked().Return(true),
			locker.EXPECT().Unlock().Return(nil),
		)
		for i := 0; i < times*2+1; i++ {
			gomock.InOrder(
				locker.EXPECT().Lock().Return(nil),
				locker.EXPECT().Locked().Return(true),
				locker.EXPECT().Unlock().Return(nil),
			)
		}

		return locker
	}

	tests := []struct {
		name        string
		mocker      func(t *testing.T, times int) *mocks.MockLocker
		options     map[string]string
		toAdd       []string
		toRem       []string
		targets     []string
		badEndpoint bool
		wantErr     bool
	}{
		{
			name:   "ok, 1 target",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []string{
				"localhost:8000",
			},
			toRem: []string{
				"http://localhost:8000",
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
			},
		},
		{
			name:   "ok, 2 targets",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []string{
				"localhost:8000",
				"168.0.0.66:8001",
			},
			toRem: []string{
				"http://localhost:8000",
				"http://168.0.0.66:8001",
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
			},
		},
		{
			name:   "ok, already existing target",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []string{
				"localhost:8000",
			},
			toRem: []string{
				"http://localhost:8000",
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
			},
		},
		{
			name: "error, nonexisting target",
			mocker: func(t *testing.T, times int) *mocks.MockLocker {
				// Create a mock locker
				ctrl := gomock.NewController(t)
				locker := mocks.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New("/monitoring/.lock").Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				for i := 0; i < times+1; i++ {
					gomock.InOrder(
						locker.EXPECT().Lock().Return(nil),
						locker.EXPECT().Locked().Return(true),
						locker.EXPECT().Unlock().Return(nil),
					)
				}

				return locker
			},
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toRem: []string{
				"http://localhost:8000",
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
			},
			wantErr: true,
		},
		{
			name:   "bad endpoint",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []string{
				"localhost:8000",
			},
			toRem: []string{
				"http://localhost:8000",
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
			},
			badEndpoint: true,
		},
		{
			name: "lock error",
			mocker: func(t *testing.T, times int) *mocks.MockLocker {
				// Create a mock locker
				ctrl := gomock.NewController(t)
				locker := mocks.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New("/monitoring/.lock").Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				locker.EXPECT().Lock().Return(fmt.Errorf("error"))
				return locker
			},
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toRem: []string{
				"localhost:8000",
			},
			wantErr: true,
		},
		{
			name: "unlock error",
			mocker: func(t *testing.T, times int) *mocks.MockLocker {
				// Create a mock locker
				ctrl := gomock.NewController(t)
				locker := mocks.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New("/monitoring/.lock").Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				gomock.InOrder(
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(false),
				)
				return locker
			},
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toRem: []string{
				"localhost:8000",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an in-memory filesystem
			afs := afero.NewMemMapFs()

			// Create a new DataDir with the in-memory filesystem
			dataDir, err := data.NewDataDir("/", afs, tt.mocker(t, len(tt.toRem)))
			require.NoError(t, err)
			stack, err := dataDir.MonitoringStack()
			require.NoError(t, err)

			// Create a new Prometheus service
			prometheus := NewPrometheus()
			err = prometheus.Init(types.ServiceOptions{
				Stack:  stack,
				Dotenv: tt.options,
			})
			require.NoError(t, err)

			// Setup the Prometheus service
			err = prometheus.Setup(tt.options)
			require.NoError(t, err)

			if !tt.badEndpoint {
				// Setup mock http server
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Listen for the POST request to /-/reload
					if r.URL.Path == "/-/reload" && r.Method == http.MethodPost {
						// All good
						w.WriteHeader(http.StatusOK)
					} else if r.Method != http.MethodGet {
						// Unexpected method
						w.WriteHeader(http.StatusMethodNotAllowed)
					} else {
						// Unexpected path
						w.WriteHeader(http.StatusNotFound)
					}
				}))
				defer server.Close()
				split := strings.Split(server.URL, ":")
				host, port := split[1][2:], split[2]
				prometheus.containerIP = host
				prometheus.port = port
			}

			// Read the prom.yml file
			var prom Config
			promYml, err := afero.ReadFile(afs, "/monitoring/prometheus/prometheus.yml")
			assert.NoError(t, err)
			err = yaml.Unmarshal(promYml, &prom)
			assert.NoError(t, err)
			// Add the targets
			for _, target := range tt.toAdd {
				job := ScrapeConfig{
					JobName: target,
					StaticConfigs: []StaticConfig{
						{
							Targets: []string{target},
							Labels:  map[string]string{"instanceID": target},
						},
					},
				}
				prom.ScrapeConfigs = append(prom.ScrapeConfigs, job)
			}
			// Save the prom.yml file
			promYml, err = yaml.Marshal(prom)
			assert.NoError(t, err)
			err = afero.WriteFile(afs, "/monitoring/prometheus/prometheus.yml", promYml, 0o644)
			assert.NoError(t, err)

			// Remove the targets
			for _, target := range tt.toRem {
				err = prometheus.RemoveTarget(target)
				if tt.wantErr || tt.badEndpoint {
					require.Error(t, err)
					return
				}
				assert.NoError(t, err)
			}

			// Read the prom.yml file
			promYml, err = afero.ReadFile(afs, "/monitoring/prometheus/prometheus.yml")
			assert.NoError(t, err)
			err = yaml.Unmarshal(promYml, &prom)
			assert.NoError(t, err)

			// Check the Prometheus targets
			assert.Equal(t, tt.targets, prom.ScrapeConfigs[0].StaticConfigs[0].Targets)
		})
	}
}

func TestContainerIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		want string
	}{
		{
			name: "Prometheus container name",
			ip:   "127.0.0.1",
			want: "127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prometheus := NewPrometheus()
			prometheus.SetContainerIP(tt.ip)
			assert.Equal(t, tt.want, prometheus.containerIP)
		})
	}
}

func TestSetContainerIP(t *testing.T) {
	tests := []struct {
		name string
		ip   string
	}{
		{
			name: "ok",
			ip:   "127.0.0.1",
		},
		{
			name: "empty",
			ip:   "",
		},
		{
			name: "domain name",
			ip:   "node-exporter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new Prometheus service
			prometheus := NewPrometheus()
			prometheus.SetContainerIP(tt.ip)
			assert.Equal(t, tt.ip, prometheus.containerIP)
		})
	}
}

func TestContainerName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "Prometheus container name",
			want: monitoring.PrometheusContainerName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prometheus := NewPrometheus()
			assert.Equal(t, tt.want, prometheus.ContainerName())
		})
	}
}

func TestEndpoint(t *testing.T) {
	dotenv := map[string]string{
		"PROM_PORT": "9999",
	}
	want := "http://prometheus:9999"

	// Create a new Prometheus service
	prometheus := NewPrometheus()
	err := prometheus.Init(types.ServiceOptions{
		Dotenv: dotenv,
	})
	require.NoError(t, err)
	prometheus.SetContainerIP("prometheus")

	endpoint := prometheus.Endpoint()
	assert.Equal(t, want, endpoint)
}
