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

	"github.com/NethermindEth/eigenlayer/internal/data"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func runCommand(t *testing.T, path string, args ...string) error {
	t.Helper()
	t.Logf("Running command: %s %s", path, strings.Join(args, " "))
	out, err := exec.Command(path, args...).CombinedOutput()
	t.Logf("===== OUTPUT =====\n%s\n==================", out)
	return err
}

func buildMockAvsImages(t *testing.T) error {
	t.Helper()
	err := runCommand(t, "docker", "build", "-t", "mock-avs-option-returner:latest", "https://github.com/NethermindEth/mock-avs-src.git#main:option-returner")
	if err != nil {
		return err
	}
	err = runCommand(t, "docker", "build", "-t", "mock-avs-plugin:latest", "https://github.com/NethermindEth/mock-avs-src.git#main:plugin")
	if err != nil {
		return err
	}
	return runCommand(t, "docker", "build", "-t", "mock-avs-health-checker:latest", "https://github.com/NethermindEth/mock-avs-src.git#main:health-checker")
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
