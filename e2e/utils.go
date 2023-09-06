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

	"github.com/NethermindEth/eigenlayer/internal/data"
	"github.com/cenkalti/backoff"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

const (
	mockAvsSrcVersion   = "v0.1.0"
	optionReturnerImage = "mock-avs-option-returner:" + mockAvsSrcVersion
	healthCheckerImage  = "mock-avs-health-checker:" + mockAvsSrcVersion
	pluginImage         = "mock-avs-plugin:" + mockAvsSrcVersion
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

func buildMockAvsImages(t *testing.T) error {
	t.Helper()
	err := runCommand(t, "docker", "build", "-t", optionReturnerImage, fmt.Sprintf("https://github.com/NethermindEth/mock-avs-src.git#%s:option-returner", mockAvsSrcVersion))
	if err != nil {
		return err
	}
	err = runCommand(t, "docker", "build", "-t", pluginImage, fmt.Sprintf("https://github.com/NethermindEth/mock-avs-src.git#%s:plugin", mockAvsSrcVersion))
	if err != nil {
		return err
	}
	return runCommand(t, "docker", "build", "-t", healthCheckerImage, fmt.Sprintf("https://github.com/NethermindEth/mock-avs-src.git#%s:health-checker", mockAvsSrcVersion))
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
