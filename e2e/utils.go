package e2e

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/NethermindEth/eigenlayer/internal/data"
	"github.com/NethermindEth/eigenlayer/internal/env"
	"github.com/cenkalti/backoff"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func runCommand(t *testing.T, path string, args ...string) error {
	_, err := runCommandOutput(t, path, args...)
	return err
}

func runCommandOutput(t *testing.T, path string, args ...string) ([]byte, error) {
	t.Helper()
	t.Logf("Running command: %s %s", path, strings.Join(args, " "))
	out, err := exec.Command(path, args...).CombinedOutput()
	t.Logf("===== OUTPUT =====\n%s\n==================", out)
	return out, err
}

func buildOptionReturnerImageLatest(t *testing.T) error {
	t.Helper()
	ok, err := imageExists(t, common.OptionReturnerImage.FullImage())
	if err != nil {
		return err
	}
	if !ok {
		return runCommand(t, "docker", "build", "-t", common.OptionReturnerImage.FullImage(), "https://github.com/NethermindEth/mock-avs.git#main:option-returner")
	}
	return nil
}

func buildOptionReturnerImage(t *testing.T, version string) error {
	t.Helper()
	ok, err := imageExists(t, common.OptionReturnerImage.Image()+":"+version)
	if err != nil {
		return err
	}
	if !ok {
		return runCommand(t, "docker", "build", "-t", common.OptionReturnerImage.Image()+":"+version, "https://github.com/NethermindEth/mock-avs.git#main:option-returner")
	}
	return nil
}

func buildHealthCheckerImageLatest(t *testing.T) error {
	t.Helper()
	ok, err := imageExists(t, common.HealthCheckerImage.FullImage())
	if err != nil {
		return err
	}
	if !ok {
		return runCommand(t, "docker", "build", "-t", common.HealthCheckerImage.FullImage(), "https://github.com/NethermindEth/mock-avs.git#main:health-checker")
	}
	return nil
}

func buildPluginImageLatest(t *testing.T) error {
	t.Helper()
	ok, err := imageExists(t, common.PluginImage.FullImage())
	if err != nil {
		return err
	}
	if !ok {
		return runCommand(t, "docker", "build", "-t", common.PluginImage.FullImage(), "https://github.com/NethermindEth/mock-avs.git#main:plugin")
	}
	return nil
}

func imageExists(t *testing.T, image string) (bool, error) {
	t.Helper()
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Fatal(err)
	}
	defer dockerClient.Close()

	_, _, err = dockerClient.ImageInspectWithRaw(context.Background(), image)
	if err != nil {
		if client.IsErrNotFound(err) {
			return false, nil
		}
		return false, err
	}
	t.Log("Image exists " + image)
	return true, nil
}

func repoPath(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Dir(wd)
}

func stopMonitoringStackContainers() error {
	// Stop monitoring stack
	dataDir, err := dataDirPath()
	if err != nil {
		return err
	}
	return exec.Command("docker", "compose", "-f", filepath.Join(dataDir, "monitoring", "docker-compose.yml"), "stop").Run()
}

func dataDirPath() (string, error) {
	userDataHome := os.Getenv("XDG_DATA_HOME")
	if userDataHome == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		userDataHome = filepath.Join(userHome, ".local", "share")
	}
	return filepath.Join(userDataHome, ".eigen"), nil
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

func getNetworkIDByName(networkName string) (string, error) {
	// Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return "", err
	}
	defer dockerClient.Close()

	network, err := dockerClient.NetworkInspect(context.Background(), networkName, types.NetworkInspectOptions{})
	if err != nil {
		return "", err
	}

	return network.ID, nil
}

func readState(stateFilePath string) (*data.Instance, error) {
	stateFile, err := os.Open(stateFilePath)
	if err != nil {
		return nil, err
	}
	defer stateFile.Close()
	var state data.Instance
	err = json.NewDecoder(stateFile).Decode(&state)
	if err != nil {
		return nil, err
	}
	return &state, nil
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

func prometheusTargets(t *testing.T) (*PrometheusTargetsResponse, error) {
	response, err := http.Get("http://localhost:9090/api/v1/targets")
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("prometheus targets status code should be 200")
	}
	var r PrometheusTargetsResponse
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func tempID(url string) string {
	tempHash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(tempHash[:])
}

func getInstance(t *testing.T, instanceID string) (*data.Instance, error) {
	dataDir, err := dataDirPath()
	if err != nil {
		return nil, err
	}
	stateFilePath := filepath.Join(dataDir, "nodes", instanceID, "state.json")

	// Get instance state
	instance, err := readState(stateFilePath)
	return instance, err
}

func logAndPipeError(t *testing.T, prefix string, err error) error {
	t.Helper()
	if err != nil {
		t.Log(prefix, err)
	}
	return err
}

func getAVSHealth(t *testing.T, url string) (int, error) {
	t.Helper()
	response, err := http.Get(url)
	if err != nil {
		return -1, err
	}
	return response.StatusCode, nil
}

func getAVSVersion(t *testing.T, ip, port string) (string, error) {
	t.Helper()
	response, err := http.Get(fmt.Sprintf("http://%s:%s/eigen/node/version", ip, port))
	if err != nil {
		return "", err
	}
	if response.StatusCode != 200 {
		return "", fmt.Errorf("expected response code %d, got %d", 200, response.StatusCode)
	}
	var r struct {
		Version string `json:"version"`
	}
	bodyData, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return r.Version, json.Unmarshal(bodyData, &r)
}

func waitHealthy(t *testing.T, containerID string, port int, networkName string, timeout time.Duration) error {
	containerIP, err := getContainerIPByName(containerID, networkName)
	if err != nil {
		return err
	}
	var (
		timeOut      time.Duration = 30 * time.Second
		responseCode               = -1
	)
	ctx, cancel := context.WithTimeout(context.Background(), timeOut)
	defer cancel()
	b := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)
	return backoff.Retry(func() error {
		responseCode, err = getAVSHealth(t, fmt.Sprintf("http://%s:%d%s", containerIP, port, "/eigen/node/health"))
		if err != nil {
			return err
		}
		if responseCode != 200 {
			return fmt.Errorf("expected response code %d, got %d", 200, responseCode)
		}
		return nil
	}, b)
}

func changeHealthStatus(t *testing.T, containerID string, port int, networkName string, healthStatus int) error {
	containerIP, err := getContainerIPByName(containerID, networkName)
	if err != nil {
		return err
	}
	response, err := http.Post(fmt.Sprintf("http://%s:%d/health/%d", containerIP, port, healthStatus), "", nil)
	if err != nil {
		return err
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("expected response code %d, got %d", 200, response.StatusCode)
	}
	return nil
}

func loadEnv(t *testing.T, instanceId string) map[string]string {
	dataDir, err := dataDirPath()
	require.NoError(t, err)
	nodesDir := filepath.Join(dataDir, "nodes")

	// Check instance directory does exist and is not empty
	instancePath := filepath.Join(nodesDir, instanceId)
	require.DirExists(t, instancePath)
	require.FileExists(t, filepath.Join(instancePath, ".env"))

	// Load .env file
	fs := afero.NewOsFs()
	envData, err := env.LoadEnv(fs, filepath.Join(instancePath, ".env"))
	require.NoError(t, err)
	return envData
}

func loadStateJSON(t *testing.T, instanceId string) json.RawMessage {
	dataDir, err := dataDirPath()
	require.NoError(t, err)
	nodesDir := filepath.Join(dataDir, "nodes")

	// Check instance directory does exist and is not empty
	instancePath := filepath.Join(nodesDir, instanceId)
	require.DirExists(t, instancePath)
	statePath := filepath.Join(instancePath, "state.json")
	require.FileExists(t, statePath)

	// Load state.json file
	fs := afero.NewOsFs()
	stateData, err := afero.ReadFile(fs, statePath)
	require.NoError(t, err)

	var stateRaw json.RawMessage
	err = json.Unmarshal(stateData, &stateRaw)
	require.NoError(t, err)
	return stateRaw
}
