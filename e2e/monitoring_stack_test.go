package e2e

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const monitoringWaitTime = 20 * time.Second

func TestMonitoringStack_Init(t *testing.T) {
	// Prepare E2E test case
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	e2eTest := NewE2ETestCase(t, filepath.Dir(wd))
	defer e2eTest.Cleanup()

	err = runCommand(t, e2eTest.EgnPath(), "--help")
	assert.NoError(t, err)

	// Wait for monitoring stack to be ready
	time.Sleep(monitoringWaitTime)

	checkMonitoringStack(t)
	checkMonitoringInit(t)
}

func TestMonitoringStack_NotReinstalled(t *testing.T) {
	// Prepare E2E test case
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	e2eTest := NewE2ETestCase(t, filepath.Dir(wd))
	defer e2eTest.Cleanup()

	err = runCommand(t, e2eTest.EgnPath(), "--help")
	assert.NoError(t, err)

	// Wait for monitoring stack to be ready
	time.Sleep(monitoringWaitTime)

	checkMonitoringStack(t)
	checkMonitoringInit(t)

	var (
		grafanaContainerID      string
		prometheusContainerID   string
		nodeExporterContainerID string
	)
	grafanaContainerID, err = getContainerIDByName("egn_grafana")
	assert.NoError(t, err)
	prometheusContainerID, err = getContainerIDByName("egn_prometheus")
	assert.NoError(t, err)
	nodeExporterContainerID, err = getContainerIDByName("egn_node_exporter")
	assert.NoError(t, err)

	err = runCommand(t, e2eTest.EgnPath(), "--help")
	assert.NoError(t, err, "egn command failed")

	checkMonitoringStack(t)
	checkMonitoringInit(t)

	newGrafanaContainerID, err := getContainerIDByName("egn_grafana")
	assert.NoError(t, err)
	assert.Equal(t, grafanaContainerID, newGrafanaContainerID, "grafana container ID has changed")

	newPrometheusContainerID, err := getContainerIDByName("egn_prometheus")
	assert.NoError(t, err)
	assert.Equal(t, prometheusContainerID, newPrometheusContainerID, "prometheus container ID has changed")

	newNodeExporterContainerID, err := getContainerIDByName("egn_node_exporter")
	assert.NoError(t, err)
	assert.Equal(t, nodeExporterContainerID, newNodeExporterContainerID, "node-exporter container ID has changed")
}

func TestMonitoring_Restart(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	e2eTest := NewE2ETestCase(t, filepath.Dir(wd))
	defer e2eTest.Cleanup()

	err = runCommand(t,
		e2eTest.EgnPath(),
		"install",
		"--profile", "option-returner",
		"--run",
		"--no-prompt",
		"--tag", "tag-1",
		"--option.main-container-name", "main-service-1",
		"https://github.com/NethermindEth/mock-avs",
	)
	assert.NoError(t, err)

	checkMonitoringStack(t)
	checkContainerRunning(t, "main-service-1")

	stopMonitoringStackContainers(t)

	err = runCommand(t,
		e2eTest.EgnPath(),
		"install",
		"--profile", "option-returner",
		"--run",
		"--no-prompt",
		"--tag", "tag-2",
		"--option.main-container-name", "main-service-2",
		"--option.network-name", "eigenlayer-2",
		"--option.main-port", "8081",
		"https://github.com/NethermindEth/mock-avs",
	)
	assert.NoError(t, err)

	checkMonitoringStack(t)

	checkContainerRunning(t, "main-service-1")
	checkContainerRunning(t, "main-service-2")
}
