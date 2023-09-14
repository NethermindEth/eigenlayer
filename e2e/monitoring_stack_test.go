package e2e

import (
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/stretchr/testify/assert"
)

// TestMonitoringStack_Init tests that the monitoring stack is not initialized if the user does not run the init-monitoring command
func TestMonitoringStack_NotInitialized(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		nil,
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "--help")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr)

			checkMonitoringStackNotInstalled(t)
			checkMonitoringStackContainersNotRunning(t)
		},
	)
	// Run test case
	e2eTest.run()
}

// TestMonitoringStack_Init tests the monitoring stack initialization
func TestMonitoringStack_Init(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		nil,
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "init-monitoring")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr)

			checkMonitoringStackDir(t)
			checkMonitoringStackContainers(t)
			checkGrafanaHealth(t)
		},
	)
	// Run test case
	e2eTest.run()
}

func TestMonitoringStack_NotReinstalled(t *testing.T) {
	// Test context
	var (
		grafanaContainerID      string
		prometheusContainerID   string
		nodeExporterContainerID string
		runErr                  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := runCommand(t, egnPath, "init-monitoring")
			if err != nil {
				return err
			}
			grafanaContainerID, err = getContainerIDByName("egn_grafana")
			if err != nil {
				return err
			}
			prometheusContainerID, err = getContainerIDByName("egn_prometheus")
			if err != nil {
				return err
			}
			nodeExporterContainerID, err = getContainerIDByName("egn_node_exporter")
			return err
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "init-monitoring")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr)

			checkMonitoringStackDir(t)
			checkMonitoringStackContainers(t)
			checkGrafanaHealth(t)
			newGrafanaContainerID, err := getContainerIDByName("egn_grafana")
			assert.NoError(t, err)
			assert.Equal(t, grafanaContainerID, newGrafanaContainerID, "grafana container ID has changed")
			newPrometheusContainerID, err := getContainerIDByName("egn_prometheus")
			assert.NoError(t, err)
			assert.Equal(t, prometheusContainerID, newPrometheusContainerID, "prometheus container ID has changed")
			newNodeExporterContainerID, err := getContainerIDByName("egn_node_exporter")
			assert.NoError(t, err)
			assert.Equal(t, nodeExporterContainerID, newNodeExporterContainerID, "node-exporter container ID has changed")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestMonitoring_Restart(t *testing.T) {
	// TODO: This test is failing, fix it
	t.Skip()
	// Test context
	var (
		mainService1IP string
		runErr         error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			if err := buildMockAvsImages(t); err != nil {
				return err
			}
			err := runCommand(t, egnPath, "init-monitoring")
			if err != nil {
				return err
			}
			err = runCommand(t, egnPath, "install", "--profile", "option-returner", "--yes", "--no-prompt", "--tag", "tag-1", "--option.main-container-name", "main-service-1", common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			mainService1IP, err = getContainerIPByName("main-service-1", "eigenlayer")
			if err != nil {
				return err
			}
			return stopMonitoringStackContainers()
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--yes", "--no-prompt", "--tag", "tag-2", "--option.main-container-name", "main-service-2", "--option.network-name", "eigenlayer-2", "--option.main-port", "8081", common.MockAvsPkg.Repo())
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr)

			mainService2IP, err := getContainerIPByName("main-service-2", "eigenlayer-2")
			assert.NoError(t, err)

			checkGrafanaHealth(t)
			checkPrometheusTargetsUp(t, "egn_node_exporter:9100", mainService1IP+":8080", mainService2IP+":8080")
			checkContainerRunning(t, "main-service-1")
			checkContainerRunning(t, "main-service-2")
		},
	)
	// Run test case
	e2eTest.run()
}
