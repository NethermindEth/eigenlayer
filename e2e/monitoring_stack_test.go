package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

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

	// Check monitoring folder exists
	monitoringDir := filepath.Join(dataDirPath(t), "monitoring")
	assert.DirExists(t, monitoringDir)

	// Check monitoring docker-compose file exists
	assert.FileExists(t, filepath.Join(monitoringDir, "docker-compose.yml"))

	// Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}
	defer dockerClient.Close()

	// Check grafana container is running
	grafanaContainer, err := dockerClient.ContainerInspect(context.Background(), "egn_grafana")
	assert.NoError(t, err)
	assert.True(t, grafanaContainer.State.Running, "grafana container is not running")

	// Check prometheus container is running
	prometheusContainer, err := dockerClient.ContainerInspect(context.Background(), "egn_prometheus")
	assert.NoError(t, err)
	assert.True(t, prometheusContainer.State.Running, "prometheus container is not running")

	// Check node-exporter container is running
	nodeExporterContainer, err := dockerClient.ContainerInspect(context.Background(), "egn_node_exporter")
	assert.NoError(t, err)
	assert.True(t, nodeExporterContainer.State.Running, "node-exporter container is not running")
}
