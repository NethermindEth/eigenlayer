package e2e

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/client"
	gapi "github.com/grafana/grafana-api-golang-client"
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

	cmd := exec.Command(e2eTest.EgnPath(), "--help")
	err = cmd.Run()
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

	cmd := exec.Command(e2eTest.EgnPath(), "--help")
	err = cmd.Run()
	assert.NoError(t, err)

	// Wait for monitoring stack to be ready
	time.Sleep(monitoringWaitTime)

	checkMonitoringStack(t)
	checkMonitoringInit(t)

	// Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}
	defer dockerClient.Close()
	var (
		grafanaContainerID      string
		prometheusContainerID   string
		nodeExporterContainerID string
	)
	// Check grafana container is running
	grafanaContainer, err := dockerClient.ContainerInspect(context.Background(), "egn_grafana")
	assert.NoError(t, err)
	assert.True(t, grafanaContainer.State.Running, "grafana container is not running")
	grafanaContainerID = grafanaContainer.ID

	// Check prometheus container is running
	prometheusContainer, err := dockerClient.ContainerInspect(context.Background(), "egn_prometheus")
	assert.NoError(t, err)
	assert.True(t, prometheusContainer.State.Running, "prometheus container is not running")
	prometheusContainerID = prometheusContainer.ID

	// Check node-exporter container is running
	nodeExporterContainer, err := dockerClient.ContainerInspect(context.Background(), "egn_node_exporter")
	assert.NoError(t, err)
	assert.True(t, nodeExporterContainer.State.Running, "node-exporter container is not running")
	nodeExporterContainerID = nodeExporterContainer.ID

	// Run egn command again
	cmd = exec.Command(e2eTest.EgnPath(), "--help")
	err = cmd.Run()
	assert.NoError(t, err)

	// Check grafana container is running
	grafanaContainer, err = dockerClient.ContainerInspect(context.Background(), "egn_grafana")
	assert.NoError(t, err)
	assert.True(t, grafanaContainer.State.Running, "grafana container is not running")
	assert.Equal(t, grafanaContainerID, grafanaContainer.ID, "grafana container ID has changed")

	// Check prometheus container is running
	prometheusContainer, err = dockerClient.ContainerInspect(context.Background(), "egn_prometheus")
	assert.NoError(t, err)
	assert.True(t, prometheusContainer.State.Running, "prometheus container is not running")
	assert.Equal(t, prometheusContainerID, prometheusContainer.ID, "prometheus container ID has changed")

	// Check node-exporter container is running
	nodeExporterContainer, err = dockerClient.ContainerInspect(context.Background(), "egn_node_exporter")
	assert.NoError(t, err)
	assert.True(t, nodeExporterContainer.State.Running, "node-exporter container is not running")
	assert.Equal(t, nodeExporterContainerID, nodeExporterContainer.ID, "node-exporter container ID has changed")

	checkMonitoringStack(t)
	checkMonitoringInit(t)
}

func checkMonitoringInit(t *testing.T) {
	// Check prometheus
	response, err := http.Get("http://localhost:9090/api/v1/targets")
	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)
	var r PrometheusTargetsResponse
	body, err := io.ReadAll(response.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(body, &r)
	assert.NoError(t, err)
	// Check number of targets
	assert.Len(t, r.Data.ActiveTargets, 1)
	// Check success
	assert.Equal(t, "success", r.Status)
	// Check node exporter target
	assert.Contains(t, r.Data.ActiveTargets[0].Labels, "instance")
	assert.Equal(t, "egn_node_exporter:9100", r.Data.ActiveTargets[0].Labels["instance"])
	// Check all targets are up
	for i := 0; i < len(r.Data.ActiveTargets); i++ {
		assert.Equal(t, "up", r.Data.ActiveTargets[i].Health)
	}

	// Check grafana
	gClient, err := gapi.New("http://localhost:3000", gapi.Config{
		BasicAuth: url.UserPassword("admin", "admin"),
	})
	assert.NoError(t, err)
	healthResponse, err := gClient.Health()
	assert.NoError(t, err)
	assert.Equal(t, "ok", healthResponse.Database)
}
