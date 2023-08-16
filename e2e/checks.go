package e2e

import (
	"context"
	"net/url"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/docker/client"
	gapi "github.com/grafana/grafana-api-golang-client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// checkMonitoringStackDir checks that the monitoring stack directory exists and contains the docker-compose file
func checkMonitoringStackDir(t *testing.T) {
	t.Logf("Checking monitoring stack directory")
	// Check monitoring folder exists
	dataDir, err := dataDirPath()
	if err != nil {
		t.Fatal(err)
	}
	monitoringDir := filepath.Join(dataDir, "monitoring")
	assert.DirExists(t, monitoringDir)

	// Check monitoring docker-compose file exists
	assert.FileExists(t, filepath.Join(monitoringDir, "docker-compose.yml"))
}

// checkMonitoringStackNotInstalled checks that the monitoring stack directory exists but is not installed
func checkMonitoringStackNotInstalled(t *testing.T) {
	t.Logf("Checking monitoring stack directory")
	// Check monitoring folder exists
	dataDir, err := dataDirPath()
	if err != nil {
		t.Fatal(err)
	}
	monitoringDir := filepath.Join(dataDir, "monitoring")
	assert.DirExists(t, monitoringDir)

	// Check monitoring docker-compose file does not exists
	assert.NoFileExists(t, filepath.Join(monitoringDir, "docker-compose.yml"))
}

// checkMonitoringStackContainers checks that the monitoring stack containers are running
func checkMonitoringStackContainers(t *testing.T) {
	t.Logf("Checking monitoring stack containers")
	checkContainerRunning(t, "egn_grafana", "egn_prometheus", "egn_node_exporter")
}

// checkContainerRunning checks that the given containers are running
func checkContainerRunning(t *testing.T, containerNames ...string) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}
	defer dockerClient.Close()

	for _, containerName := range containerNames {
		t.Logf("Checking %s container is running", containerName)
		container, err := dockerClient.ContainerInspect(context.Background(), containerName)
		assert.NoError(t, err)
		assert.True(t, container.State.Running, "%s container should be running", containerName)
	}
}

// checkPrometheusTargetsUp checks that the prometheus targets are up
func checkPrometheusTargetsUp(t *testing.T, targets ...string) {
	promTargets, err := prometheusTargets(t)
	if err != nil {
		t.Fatal(err)
	}
	// Check prometheus targets
	// Check number of targets
	assert.Len(t, promTargets.Data.ActiveTargets, len(targets))
	// Check success
	assert.Equal(t, "success", promTargets.Status)

	var labels []string
	for _, target := range promTargets.Data.ActiveTargets {
		assert.Contains(t, promTargets.Data.ActiveTargets[0].Labels, "instance")
		labels = append(labels, target.Labels["instance"])
	}
	for _, target := range targets {
		assert.Contains(t, labels, target)
	}
	// Check all targets are up
	for i := 0; i < len(promTargets.Data.ActiveTargets); i++ {
		// Try 10 times to get the target health different than unknown
		for tries := 0; tries < 10; tries++ {
			if promTargets.Data.ActiveTargets[i].Health == "unknown" {
				t.Logf("Target %s health is unknown. Waiting 1 sec to try again (%d/10)", promTargets.Data.ActiveTargets[i].Labels["instance"], tries+1)
				time.Sleep(time.Second)
				promTargets, err = prometheusTargets(t)
				if err != nil {
					t.Fatal(err)
				}
			} else {
				break
			}
		}
		assert.Equal(t, "up", promTargets.Data.ActiveTargets[i].Health, "target %s is not up", promTargets.Data.ActiveTargets[i].Labels["instance"])
	}
}

// checkPrometheusTargetsDown checks that the prometheus targets are up
func checkPrometheusTargetsDown(t *testing.T, targets ...string) {
	promTargets, err := prometheusTargets(t)
	if err != nil {
		t.Fatal(err)
	}
	// Check prometheus targets
	// Check number of targets
	assert.Len(t, promTargets.Data.ActiveTargets, len(targets))
	// Check success
	assert.Equal(t, "success", promTargets.Status)

	var labels []string
	for _, target := range promTargets.Data.ActiveTargets {
		assert.Contains(t, promTargets.Data.ActiveTargets[0].Labels, "instance")
		labels = append(labels, target.Labels["instance"])
	}
	for _, target := range targets {
		assert.NotContains(t, labels, target)
	}
}

// checkPrometheusHealth checks that the prometheus health is ok
func checkGrafanaHealth(t *testing.T) {
	t.Logf("Checking Grafana health")
	// Check Grafana health
	gClient, err := gapi.New("http://localhost:3000", gapi.Config{
		BasicAuth: url.UserPassword("admin", "admin"),
	})
	assert.NoError(t, err)
	healthResponse, err := gClient.Health()
	assert.NoError(t, err)
	assert.Equal(t, "ok", healthResponse.Database)
}

// checkMonitoringStackContainersNotRunning checks that the monitoring stack containers are not running
func checkMonitoringStackContainersNotRunning(t *testing.T) {
	t.Logf("Checking monitoring stack containers are not running")
	checkContainerNotExisting(t, "egn_grafana", "egn_prometheus", "egn_node_exporter")
}

// checkContainerNotRunning checks that the given containers are not running
func checkContainerNotRunning(t *testing.T, containerNames ...string) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}
	defer dockerClient.Close()

	for _, containerName := range containerNames {
		t.Logf("Checking %s container is not running", containerName)
		container, err := dockerClient.ContainerInspect(context.Background(), containerName)
		assert.NoError(t, err)
		assert.False(t, container.State.Running, "%s container should not be running", containerName)
	}
}

// checkContainerNotExisting checks that the given containers are not existing
func checkContainerNotExisting(t *testing.T, containerNames ...string) {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}
	defer dockerClient.Close()

	for _, containerName := range containerNames {
		t.Logf("Checking %s container is not running", containerName)
		_, err := dockerClient.ContainerInspect(context.Background(), containerName)
		assert.Error(t, err)
	}
}

// checkInstanceInstalled checks that the instance directory does exist and is not empty
func checkInstanceInstalled(t *testing.T, instanceId string) {
	t.Logf("Checking instance directory")
	// Check nodes folder exists
	dataDir, err := dataDirPath()
	if err != nil {
		t.Fatal(err)
	}
	nodesDir := filepath.Join(dataDir, "nodes")

	// Check instance directory does exist and is not empty
	instancePath := filepath.Join(nodesDir, instanceId)
	assert.DirExists(t, instancePath)
	assert.FileExists(t, filepath.Join(instancePath, "docker-compose.yml"))
	assert.FileExists(t, filepath.Join(instancePath, ".env"))
	assert.FileExists(t, filepath.Join(instancePath, "profile.yml"))
	assert.FileExists(t, filepath.Join(instancePath, "state.json"))
}

// checkInstanceNotInstalled checks that the instance directory does not exist
func checkInstanceNotInstalled(t *testing.T, instanceId string) {
	t.Logf("Checking instance directory")
	// Check nodes folder exists
	dataDir, err := dataDirPath()
	if err != nil {
		t.Fatal(err)
	}
	nodesDir := filepath.Join(dataDir, "nodes")

	// Check instance directory does not exist
	assert.NoDirExists(t, filepath.Join(nodesDir, instanceId))
}

// checkTemporaryPackageNotExisting checks that the temporary package directory does not exist
func checkTemporaryPackageNotExisting(t *testing.T, instance string) {
	t.Logf("Checking temporary package directory")
	dataDir, err := dataDirPath()
	if err != nil {
		t.Fatal(err)
	}
	tempDir := filepath.Join(dataDir, "temp")
	tID := tempID(instance)

	// Check package directory does not exist
	assert.NoDirExists(t, filepath.Join(tempDir, tID))
}

func checkInstanceExists(t *testing.T, instanceID string) {
	t.Logf("Checking instance %s exists", instanceID)

	// Check nodes folder exists
	dataDir, err := dataDirPath()
	if err != nil {
		t.Fatal(err)
	}
	nodesDir := filepath.Join(dataDir, "nodes")

	// Check instance directory does exist and is not empty
	instancePath := filepath.Join(nodesDir, instanceID)
	require.DirExists(t, instancePath)
	stateFilePath := filepath.Join(instancePath, "state.json")
	require.FileExists(t, filepath.Join(stateFilePath))
}
