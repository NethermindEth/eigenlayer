package e2e

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/docker/docker/client"
	gapi "github.com/grafana/grafana-api-golang-client"
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

	// Check Grafana health using API
	gClient, err := gapi.New("http://localhost:3000", gapi.Config{
		BasicAuth: url.UserPassword("admin", "admin"),
	})
	assert.NoError(t, err)
	healthResponse, err := gClient.Health()
	assert.NoError(t, err)
	assert.Equal(t, "ok", healthResponse.Database)
}

func getContainerIPByName(containerName string, networkName string) (string, error) {
	// Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	defer dockerClient.Close()

	container, err := dockerClient.ContainerInspect(context.Background(), containerName)
	if err != nil {
		return "", err
	}
	network, ok := container.NetworkSettings.Networks[networkName]
	if !ok {
		return "", fmt.Errorf("network %s not found", networkName)
	}
	return network.IPAddress, nil
}

func getContainerIDByName(containerName string) (string, error) {
	// Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	defer dockerClient.Close()

	container, err := dockerClient.ContainerInspect(context.Background(), containerName)
	if err != nil {
		return "", err
	}

	return container.ID, nil
}

func checkPrometheusTargets(t *testing.T, targets ...string) {
	// Check prometheus targets
	response, err := http.Get("http://localhost:9090/api/v1/targets")
	assert.NoError(t, err)
	assert.Equal(t, 200, response.StatusCode)
	var r PrometheusTargetsResponse
	body, err := io.ReadAll(response.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(body, &r)
	assert.NoError(t, err)
	// Check number of targets
	assert.Len(t, r.Data.ActiveTargets, len(targets))
	// Check success
	assert.Equal(t, "success", r.Status)
	// Check node exporter target
	var labels []string
	for _, target := range r.Data.ActiveTargets {
		assert.Contains(t, r.Data.ActiveTargets[0].Labels, "instance")
		labels = append(labels, target.Labels["instance"])
	}
	for _, target := range targets {
		assert.Contains(t, labels, target)
	}
	// TODO: check mock-avs target
	// Check all targets are up
	for i := 0; i < len(r.Data.ActiveTargets); i++ {
		assert.Equal(t, "up", r.Data.ActiveTargets[i].Health)
	}
}

func checkGrafanaHealth(t *testing.T) {
	// Check Grafana health
	gClient, err := gapi.New("http://localhost:3000", gapi.Config{
		BasicAuth: url.UserPassword("admin", "admin"),
	})
	assert.NoError(t, err)
	healthResponse, err := gClient.Health()
	assert.NoError(t, err)
	assert.Equal(t, "ok", healthResponse.Database)
}

func checkContainerRunning(t *testing.T, containerName string) {
	// Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}
	defer dockerClient.Close()

	// Check if the container is running
	container, err := dockerClient.ContainerInspect(context.Background(), containerName)
	assert.NoError(t, err)
	assert.True(t, container.State.Running, "%s container should be running", containerName)
}

func stopMonitoringStackContainers(t *testing.T) {
	// Stop monitoring stack
	dataDir := dataDirPath(t)
	err := exec.Command("docker", "compose", "-f", filepath.Join(dataDir, "monitoring", "docker-compose.yml"), "stop").Run()
	if err != nil {
		t.Fatalf("error stopping monitoring stack: %v", err)
	}
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
