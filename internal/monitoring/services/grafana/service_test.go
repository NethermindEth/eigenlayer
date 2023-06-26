package grafana

import (
	"fmt"
	"testing"

	"github.com/NethermindEth/egn/internal/data"
	"github.com/NethermindEth/egn/internal/locker/mocks"
	"github.com/NethermindEth/egn/internal/monitoring"
	"github.com/NethermindEth/egn/internal/monitoring/services/types"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type Config struct {
	APIVersion  int          `yaml:"apiVersion"`
	Datasources []Datasource `yaml:"datasources"`
}

type Datasource struct {
	Name     string   `yaml:"name"`
	Type     string   `yaml:"type"`
	Access   string   `yaml:"access"`
	URL      string   `yaml:"url"`
	UID      string   `yaml:"uid"`
	JsonData JsonData `yaml:"jsonData"`
}

type JsonData struct {
	HTTPMethod                    string                       `yaml:"httpMethod"`
	ManageAlerts                  bool                         `yaml:"manageAlerts"`
	PrometheusType                string                       `yaml:"prometheusType"`
	PrometheusVersion             string                       `yaml:"prometheusVersion"`
	IncrementalQuerying           bool                         `yaml:"incrementalQuerying"`
	IncrementalQueryOverlapWindow string                       `yaml:"incrementalQueryOverlapWindow"`
	CacheLevel                    string                       `yaml:"cacheLevel"`
	ExemplarTraceIdDestinations   []ExemplarTraceIdDestination `yaml:"exemplarTraceIdDestinations"`
}

type ExemplarTraceIdDestination struct {
	DatasourceUid string `yaml:"datasourceUid"`
	Name          string `yaml:"name"`
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
		wantErr bool
	}{
		{
			name:   "ok",
			mocker: okLocker,
			options: map[string]string{
				"PROM_PORT": "9090",
			},
		},
		{
			name:    "missing prometheus port",
			mocker:  onlyNewLocker,
			options: map[string]string{},
			wantErr: true,
		},
		{
			name:   "empty prometheus port",
			mocker: onlyNewLocker,
			options: map[string]string{
				"PROM_PORT": "",
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
				"PROM_PORT": "9090",
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
				"PROM_PORT": "9090",
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

			// Create a new Grafana service
			grafana := NewGrafana()
			grafana.Init(types.ServiceOptions{
				Stack: stack,
			})

			// Setup the Grafana service
			err = grafana.Setup(tt.options)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				assert.NoError(t, err)
				ok, err := afero.Exists(afs, "/monitoring/grafana/provisioning/datasources/prom.yml")
				assert.True(t, ok)
				assert.NoError(t, err)

				// Read the prom.yml file
				var prom Config
				promYml, err := afero.ReadFile(afs, "/monitoring/grafana/provisioning/datasources/prom.yml")
				assert.NoError(t, err)
				err = yaml.Unmarshal(promYml, &prom)
				assert.NoError(t, err)

				// Check the Prometheus port
				promEndpoint := fmt.Sprintf("http://%s:%s", monitoring.PrometheusServiceName, tt.options["PROM_PORT"])
				assert.Equal(t, promEndpoint, prom.Datasources[0].URL)
			}
		})
	}
}

func TestDotEnv(t *testing.T) {
	// Create a new Grafana service
	grafana := NewGrafana()
	// Verify the dotEnv
	assert.EqualValues(t, dotEnv, grafana.DotEnv())
}