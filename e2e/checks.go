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

// checkPrometheusTargets checks that the prometheus targets are up
func checkPrometheusTargets(t *testing.T, targets ...string) {
	promTargets, err := prometheusTargets(t)
	if err != nil {
		t.Fatal(err)
	}
	// Check prometheus targets
	// Check number of targets
	assert.Len(t, promTargets.Data.ActiveTargets, len(targets))
	// Check success
	assert.Equal(t, "success", promTargets.Status)
	// Check node exporter target
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
