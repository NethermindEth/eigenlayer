package prometheus

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
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
				assert.Equal(t, tt.options.Dotenv["PROM_PORT"], strconv.Itoa(int(prometheus.port)))
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

	type target struct {
		instanceID  string
		commitHash  string
		avsName     string
		avsVersion  string
		specVersion string
		network     string
		target      types.MonitoringTarget
	}

	tests := []struct {
		name        string
		mocker      func(t *testing.T, times int) *mocks.MockLocker
		options     map[string]string
		toAdd       []target
		targets     []ScrapeConfig
		badEndpoint bool
		wantErr     bool
	}{
		{
			name:   "ok, 1 target, no path",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []target{
				{
					instanceID:  "test-avs",
					commitHash:  "76973ce6755edb6cce37efd62266e98c838f6968",
					avsName:     "crazy-avs",
					avsVersion:  "v0.0.1",
					specVersion: "v1.0.0",
					network:     "testnet",
					target: types.MonitoringTarget{
						Host: "localhost",
						Port: 8000,
					},
				},
			},
			targets: []ScrapeConfig{
				{
					JobName: fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
					StaticConfigs: []StaticConfig{
						{
							Targets: []string{
								fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
							},
						},
					},
				},
				{
					JobName: "test-avs--0++testnet",
					StaticConfigs: []StaticConfig{
						{
							Targets: []string{
								"localhost:8000",
							},
							Labels: map[string]string{
								monitoring.InstanceIDLabel:  "test-avs",
								monitoring.CommitHashLabel:  "76973ce6755edb6cce37efd62266e98c838f6968",
								monitoring.AVSNameLabel:     "crazy-avs",
								monitoring.AVSVersionLabel:  "v0.0.1",
								monitoring.SpecVersionLabel: "v1.0.0",
							},
						},
					},
					MetricsPath: "/metrics",
				},
			},
		},
		{
			name:   "ok, 1 target, custom path",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []target{
				{
					instanceID:  "test-avs",
					commitHash:  "a0c93c0ce7af88bd6387d2a2522b6d7390e50d09",
					avsName:     "mad-avs",
					avsVersion:  "v0.1.1",
					specVersion: "v1.1.0",
					network:     "testnet",
					target: types.MonitoringTarget{
						Host: "localhost",
						Port: 8000,
						Path: "/custom-path",
					},
				},
			},
			targets: []ScrapeConfig{
				{
					JobName: fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
					StaticConfigs: []StaticConfig{
						{
							Targets: []string{
								fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
							},
						},
					},
				},
				{
					JobName: "test-avs--0++testnet",
					StaticConfigs: []StaticConfig{
						{
							Targets: []string{
								"localhost:8000",
							},
							Labels: map[string]string{
								monitoring.InstanceIDLabel:  "test-avs",
								monitoring.CommitHashLabel:  "a0c93c0ce7af88bd6387d2a2522b6d7390e50d09",
								monitoring.AVSNameLabel:     "mad-avs",
								monitoring.AVSVersionLabel:  "v0.1.1",
								monitoring.SpecVersionLabel: "v1.1.0",
							},
						},
					},
					MetricsPath: "/custom-path",
				},
			},
		},
		{
			name:   "ok, 2 targets",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []target{
				{
					instanceID:  "test-avs1",
					commitHash:  "95d30fb815626090240f993aec21752f4a6866d5",
					avsName:     "evil-avs",
					avsVersion:  "v6.6.6",
					specVersion: "v0.0.3",
					network:     "testnet1",
					target: types.MonitoringTarget{
						Host: "localhost",
						Port: 8000,
					},
				},
				{
					instanceID:  "test-avs2",
					commitHash:  "76973ce6755edb6cce37efd62266e98c838f6968",
					avsName:     "insane-avs",
					avsVersion:  "v19.0.1",
					specVersion: "v13.0.0",
					network:     "testnet2",
					target: types.MonitoringTarget{
						Host: "168.0.0.66",
						Port: 8001,
						Path: "/custom-path2",
					},
				},
			},
			targets: []ScrapeConfig{
				{
					JobName: fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
					StaticConfigs: []StaticConfig{
						{
							Targets: []string{
								fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
							},
						},
					},
				},
				{
					JobName: "test-avs1--0++testnet1",
					StaticConfigs: []StaticConfig{
						{
							Targets: []string{
								"localhost:8000",
							},
							Labels: map[string]string{
								monitoring.InstanceIDLabel:  "test-avs1",
								monitoring.CommitHashLabel:  "95d30fb815626090240f993aec21752f4a6866d5",
								monitoring.AVSNameLabel:     "evil-avs",
								monitoring.AVSVersionLabel:  "v6.6.6",
								monitoring.SpecVersionLabel: "v0.0.3",
							},
						},
					},
					MetricsPath: "/metrics",
				},
				{
					JobName: "test-avs2--1++testnet2",
					StaticConfigs: []StaticConfig{
						{
							Targets: []string{
								"168.0.0.66:8001",
							},
							Labels: map[string]string{
								monitoring.InstanceIDLabel:  "test-avs2",
								monitoring.CommitHashLabel:  "76973ce6755edb6cce37efd62266e98c838f6968",
								monitoring.AVSNameLabel:     "insane-avs",
								monitoring.AVSVersionLabel:  "v19.0.1",
								monitoring.SpecVersionLabel: "v13.0.0",
							},
						},
					},
					MetricsPath: "/custom-path2",
				},
			},
		},
		{
			name:   "bad endpoint",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []target{
				{
					instanceID:  "test-avs",
					commitHash:  "ddc7b0e122129a79bd74d922cdb9dcdaaa24c3ee",
					avsName:     "lunatic-avs",
					avsVersion:  "v1.0.0",
					specVersion: "v1.0.0",
					network:     "testnet",
					target: types.MonitoringTarget{
						Host: "localhost",
						Port: 8000,
					},
				},
			},
			targets: []ScrapeConfig{
				{
					JobName: fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
					StaticConfigs: []StaticConfig{
						{
							Targets: []string{
								fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
							},
						},
					},
				},
				{
					JobName: "test-avs--0++testnet",
					StaticConfigs: []StaticConfig{
						{
							Targets: []string{
								"localhost:8000",
							},
							Labels: map[string]string{
								monitoring.InstanceIDLabel:  "test-avs",
								monitoring.CommitHashLabel:  "ddc7b0e122129a79bd74d922cdb9dcdaaa24c3ee",
								monitoring.AVSNameLabel:     "lunatic-avs",
								monitoring.AVSVersionLabel:  "v1.0.0",
								monitoring.SpecVersionLabel: "v1.0.0",
							},
						},
					},
					MetricsPath: "/metrics",
				},
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
			toAdd: []target{
				{
					instanceID:  "test-avs",
					commitHash:  "ddc7b0e122129a79bd74d922cdb9dcdaaa24c3ee",
					avsName:     "wild-avs",
					avsVersion:  "v1.0.0",
					specVersion: "v1.0.0",
					network:     "testnet",
					target: types.MonitoringTarget{
						Host: "localhost",
						Port: 8000,
					},
				},
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
			toAdd: []target{
				{
					instanceID:  "test-avs",
					commitHash:  "da5c2d3f4d9494ac9f7f03e9b11bd7610ec537dc",
					avsName:     "nutty-avs",
					avsVersion:  "v1.0.0",
					specVersion: "v1.0.0",
					network:     "testnet",
					target: types.MonitoringTarget{
						Host: "localhost",
						Port: 8000,
					},
				},
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
				prometheus.containerIP = net.ParseIP(host)
				p, err := strconv.Atoi(port)
				require.NoError(t, err)
				prometheus.port = uint16(p)
			}

			// Add the targets
			for i, target := range tt.toAdd {
				labels := map[string]string{
					monitoring.InstanceIDLabel:  target.instanceID,
					monitoring.CommitHashLabel:  target.commitHash,
					monitoring.AVSNameLabel:     target.avsName,
					monitoring.AVSVersionLabel:  target.avsVersion,
					monitoring.SpecVersionLabel: target.specVersion,
				}
				err = prometheus.AddTarget(target.target, labels, fmt.Sprintf("%s--%d++%s", target.instanceID, i, target.network))
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
				assert.EqualValues(t, target, prom.ScrapeConfigs[i], target)
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

	type target struct {
		instanceID string
		network    string
		endpoint   string
	}

	tests := []struct {
		name        string
		mocker      func(t *testing.T, times int) *mocks.MockLocker
		options     map[string]string
		toAdd       []target
		toRem       []target
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
			toAdd: []target{
				{
					instanceID: "test-avs",
					network:    "testnet",
					endpoint:   "localhost:8000",
				},
			},
			toRem: []target{
				{
					instanceID: "test-avs",
					network:    "testnet",
				},
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
			},
		},
		{
			name:   "ok, 2 targets, different instances",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []target{
				{
					instanceID: "test-avs1",
					network:    "testnet1",
					endpoint:   "localhost:8000",
				},
				{
					instanceID: "test-avs2",
					network:    "testnet2",
					endpoint:   "168.0.0.66:8001",
				},
			},
			toRem: []target{
				{
					instanceID: "test-avs1",
					network:    "testnet1",
				},
				{
					instanceID: "test-avs2",
					network:    "testnet2",
				},
			},
			targets: []string{
				fmt.Sprintf("%s:9100", monitoring.NodeExporterContainerName),
			},
		},
		{
			name:   "ok, 2 targets, same instance",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT":          "9999",
				"NODE_EXPORTER_PORT": "9100",
			},
			toAdd: []target{
				{
					instanceID: "test-avs",
					network:    "testnet",
					endpoint:   "localhost:8000",
				},
				{
					instanceID: "test-avs",
					network:    "testnet",
					endpoint:   "168.0.0.66:8001",
				},
			},
			toRem: []target{
				{
					instanceID: "test-avs",
					network:    "testnet",
				},
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
			toRem: []target{
				{
					instanceID: "test-avs",
					network:    "testnet",
				},
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
			toAdd: []target{
				{
					instanceID: "test-avs",
					network:    "testnet",
					endpoint:   "localhost:8000",
				},
			},
			toRem: []target{
				{
					instanceID: "test-avs",
					network:    "testnet",
					endpoint:   "localhost:8000",
				},
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
			toRem: []target{
				{
					instanceID: "test-avs",
					network:    "testnet",
				},
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
			toRem: []target{
				{
					instanceID: "test-avs",
					network:    "testnet",
					endpoint:   "localhost:8000",
				},
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
				prometheus.containerIP = net.ParseIP(host)
				p, err := strconv.Atoi(port)
				require.NoError(t, err)
				prometheus.port = uint16(p)
			}

			// Read the prom.yml file
			var prom Config
			promYml, err := afero.ReadFile(afs, "/monitoring/prometheus/prometheus.yml")
			assert.NoError(t, err)
			err = yaml.Unmarshal(promYml, &prom)
			assert.NoError(t, err)
			// Add the targets
			for i, target := range tt.toAdd {
				job := ScrapeConfig{
					JobName: fmt.Sprintf("%s--%d++%s", target.instanceID, i, target.network),
					StaticConfigs: []StaticConfig{
						{
							Targets: []string{target.endpoint},
							Labels:  map[string]string{"instanceID": target.instanceID},
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
				network, err := prometheus.RemoveTarget(target.instanceID)
				if tt.wantErr || tt.badEndpoint {
					require.Error(t, err)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, target.network, network)
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

func TestSetContainerIP(t *testing.T) {
	tests := []struct {
		name string
		ip   net.IP
	}{
		{
			name: "ok",
			ip:   net.ParseIP("127.0.0.1"),
		},
		{
			name: "empty",
			ip:   nil,
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
	want := monitoring.PrometheusContainerName

	prometheus := NewPrometheus()
	assert.Equal(t, want, prometheus.ContainerName())
}

func TestEndpoint(t *testing.T) {
	dotenv := map[string]string{
		"PROM_PORT": "9999",
	}
	want := "http://168.44.55.66:9999"

	// Create a new Prometheus service
	prometheus := NewPrometheus()
	err := prometheus.Init(types.ServiceOptions{
		Dotenv: dotenv,
	})
	require.NoError(t, err)
	prometheus.SetContainerIP(net.ParseIP("168.44.55.66"))

	endpoint := prometheus.Endpoint()
	assert.Equal(t, want, endpoint)
}
