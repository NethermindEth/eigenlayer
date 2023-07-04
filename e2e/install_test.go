package e2e

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	gapi "github.com/grafana/grafana-api-golang-client"
	"github.com/stretchr/testify/assert"
)

func TestInstall_WithoutArguments(t *testing.T) {
	// Prepare E2E test case
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	e2eTest := NewE2ETestCase(t, filepath.Dir(wd))
	defer e2eTest.Cleanup()

	cmd := exec.Command(e2eTest.EgnPath(), "install")
	err = cmd.Run()
	assert.Error(t, err, "install command should fail without arguments")
}

func TestInstall_ValidArgument(t *testing.T) {
	// Prepare E2E test case
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	e2eTest := NewE2ETestCase(t, filepath.Dir(wd))
	defer e2eTest.Cleanup()

	cmd := exec.Command(
		e2eTest.EgnPath(),
		"install",
		"--profile", "option-returner",
		"--run",
		"--no-prompt",
		"https://github.com/NethermindEth/mock-avs",
	)
	err = cmd.Run()
	assert.NoError(t, err)

	checkMonitoringStack(t)

	// Wait for monitoring stack to be ready
	time.Sleep(30 * time.Second)

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
	assert.Len(t, r.Data.ActiveTargets, 1) // TODO: number of targets should be 2 (node exporter + mock-avs)
	// Check success
	assert.Equal(t, "success", r.Status)
	// Check node exporter target
	assert.Contains(t, r.Data.ActiveTargets[0].Labels, "instance")
	assert.Equal(t, "egn_node_exporter:9100", r.Data.ActiveTargets[0].Labels["instance"])
	// TODO: check mock-avs target
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
