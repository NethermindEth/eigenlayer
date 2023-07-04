package e2e

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func dataDirPath(t *testing.T) string {
	t.Helper()
	userDataHome := os.Getenv("XDG_DATA_HOME")
	if userDataHome == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			t.Fatal(err)
		}
		userDataHome = filepath.Join(userHome, ".local", "share")
	}
	dataDir := filepath.Join(userDataHome, ".eigen")
	return dataDir
}

func runCommand(t *testing.T, path string, args ...string) error {
	t.Helper()
	t.Logf("Running command: %s %s", path, strings.Join(args, " "))
	out, err := exec.Command(path, args...).CombinedOutput()
	t.Logf("===== OUTPUT =====\n%s\n==================", out)
	return err
}

func checkMonitoringStack(t *testing.T) {
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

type Target struct {
	Labels Labels `json:"labels"`
	Health string `json:"health"`
}

type Labels map[string]string

type Data struct {
	ActiveTargets []Target `json:"activeTargets"`
}

type PrometheusTargetsResponse struct {
	Status string `json:"status"`
	Data   Data   `json:"data"`
}
