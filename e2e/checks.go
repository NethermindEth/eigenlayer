package e2e

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/NethermindEth/eigenlayer/pkg/monitoring"
	"github.com/cenkalti/backoff"
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
		require.NoError(t, err)
		assert.True(t, container.State.Running, "%s container should be running", containerName)
	}
}

// checkPrometheusTargetsUp checks that the prometheus targets are up
func checkPrometheusTargetsUp(t *testing.T, targets ...string) {
	var (
		tries       int           = 0
		timeOut     time.Duration = 30 * time.Second
		promTargets *PrometheusTargetsResponse
		err         error
	)
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	b := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	err = backoff.Retry(func() error {
		tries++
		logPrefix := fmt.Sprintf("checkPrometheusTargetsUp (%d)", tries)
		promTargets, err = prometheusTargets(t)
		if err != nil {
			return logAndPipeError(t, logPrefix, err)
		}
		if promTargets.Status != "success" {
			return logAndPipeError(t, logPrefix, fmt.Errorf("expected status success, got %s", promTargets.Status))
		}
		if len(promTargets.Data.ActiveTargets) != len(targets) {
			return logAndPipeError(t, logPrefix, fmt.Errorf("expected %d targets, got %d", len(targets), len(promTargets.Data.ActiveTargets)))
		}
		for i, target := range promTargets.Data.ActiveTargets {
			var labels []string
			for label := range target.Labels {
				labels = append(labels, label)
			}
			if !slices.Contains(labels, "instance") {
				return logAndPipeError(t, logPrefix, fmt.Errorf("target %d does not have instance label", i))
			}
			instanceLabel := target.Labels["instance"]
			if !slices.Contains(targets, instanceLabel) {
				return logAndPipeError(t, logPrefix, fmt.Errorf("target %d instance label is not expected", i))
			}
			if target.Health == "unknown" {
				return logAndPipeError(t, logPrefix, fmt.Errorf("target %d health is unknown", i))
			}
		}
		return nil
	}, b)
	assert.NoError(t, err, `targets "%s" should be up, but after %d tries they are not`, targets, tries)
}

// checkPrometheusTargetsDown checks that the prometheus targets are up
func checkPrometheusTargetsDown(t *testing.T, targets ...string) {
	promTargets, err := prometheusTargets(t)
	if err != nil {
		t.Fatal(err)
	}
	// Check success
	assert.Equal(t, "success", promTargets.Status)

	var labels []string
	for _, target := range promTargets.Data.ActiveTargets {
		assert.Contains(t, target.Labels, "instance")
		labels = append(labels, target.Labels["instance"])
	}
	for _, target := range targets {
		assert.NotContains(t, labels, target)
	}
}

// checkPrometheusHealth checks that the prometheus health is ok
func checkGrafanaHealth(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tries := 0
	b := backoff.WithContext(backoff.NewConstantBackOff(time.Second), ctx)
	err := backoff.Retry(func() error {
		logPrefix := fmt.Sprintf("checkGrafanaHealth (%d)", tries+1)
		tries++
		// Check Grafana health
		gClient, err := gapi.New("http://localhost:3000", gapi.Config{
			BasicAuth: url.UserPassword("admin", "admin"),
		})
		if err != nil {
			return logAndPipeError(t, logPrefix, err)
		}
		healthResponse, err := gClient.Health()
		if err != nil {
			return logAndPipeError(t, logPrefix, err)
		}
		if healthResponse.Database != "ok" {
			return logAndPipeError(t, logPrefix, fmt.Errorf("expected database ok, got %s", healthResponse.Database))
		}
		return nil
	}, b)
	assert.NoError(t, err, "Grafana should be ok, but it is not")
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

// checkPrometheusLabels checks that the prometheus metrics from the given targets contain the injected labels.
// Should be called after checkPrometheusTargetsUp.
func checkPrometheusLabels(t *testing.T, targets ...string) {
	promTargets, err := prometheusTargets(t)
	if err != nil {
		t.Fatal(err)
	}
	// Check prometheus targets
	// Check number of targets
	assert.Len(t, promTargets.Data.ActiveTargets, len(targets)+1)
	// Check success
	assert.Equal(t, "success", promTargets.Status)

	labels := [...]string{
		monitoring.InstanceIDLabel,
		monitoring.CommitHashLabel,
		monitoring.AVSNameLabel,
		monitoring.AVSVersionLabel,
		monitoring.SpecVersionLabel,
	}

	var instanceLabels []string
	for _, target := range promTargets.Data.ActiveTargets {
		require.Contains(t, target.Labels, "instance")
		instanceLabels = append(instanceLabels, target.Labels["instance"])

		// Skip node exporter target
		if target.Labels["instance"] == monitoring.NodeExporterContainerName {
			continue
		}
		// Skip if not in targets
		if !slices.Contains(targets, target.Labels["instance"]) {
			continue
		}
		for _, label := range labels {
			assert.Contains(t, target.Labels, label, "target %s does not contain label %s", target.Labels[monitoring.InstanceIDLabel], label)
			assert.NotEmpty(t, target.Labels[label], "target %s label %s is empty", target.Labels[monitoring.InstanceIDLabel], label)
		}
	}
	// Check all targets are in the prometheus targets
	for _, target := range targets {
		assert.Contains(t, instanceLabels, target)
	}
}

func checkAVSHealth(t *testing.T, ip string, port string, wantCode int) {
	t.Logf("Checking AVS health")
	var (
		timeOut      time.Duration = 30 * time.Second
		responseCode               = -1
		err          error
	)
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	b := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	err = backoff.Retry(func() error {
		responseCode, err = getAVSHealth(t, fmt.Sprintf("http://%s:%s%s", ip, port, "/eigen/node/health"))
		if err != nil {
			return err
		}
		if responseCode != wantCode {
			return fmt.Errorf("expected response code %d, got %d", wantCode, responseCode)
		}
		return nil
	}, b)
	assert.NoError(t, err, "AVS health should be ok, but it is not")
	assert.Equal(t, wantCode, responseCode, "AVS health should be %d, but it is %d", wantCode, responseCode)
}
