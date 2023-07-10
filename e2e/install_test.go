package e2e

import (
	"os"
	"path/filepath"
	"testing"
	"time"

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

	err = runCommand(t, e2eTest.EgnPath(), "install")

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

	// Install the mock-avs
	err = runCommand(t,
		e2eTest.EgnPath(),
		"install",
		"--profile", "option-returner",
		"--run",
		"--no-prompt",
		"https://github.com/NethermindEth/mock-avs",
	)
	assert.NoError(t, err)

	// Wait for the monitoring stack to be ready
	time.Sleep(monitoringWaitTime)

	checkMonitoringStack(t)

	checkContainerRunning(t, "option-returner")
}

func TestInstall_DuplicatedID(t *testing.T) {
	// Prepare E2E test case
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	e2eTest := NewE2ETestCase(t, filepath.Dir(wd))
	defer e2eTest.Cleanup()

	err = runCommand(t,
		e2eTest.EgnPath(),
		"install",
		"--profile", "option-returner",
		"--run",
		"--no-prompt",
		"--tag", "integration",
		"https://github.com/NethermindEth/mock-avs",
	)
	assert.NoError(t, err)

	time.Sleep(monitoringWaitTime)

	checkMonitoringStack(t)

	optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
	assert.NoError(t, err)

	checkPrometheusTargets(t, "egn_node_exporter:9100", optionReturnerIP+":8080") // Expecting 2 targets (node exporter + option-returner)

	checkGrafanaHealth(t)

	checkContainerRunning(t, "option-returner")

	err = runCommand(t,
		e2eTest.EgnPath(),
		"install",
		"--profile", "option-returner",
		"--run",
		"--no-prompt",
		"--tag", "integration",
		"https://github.com/NethermindEth/mock-avs",
	)
	assert.Error(t, err)
}

func TestInstall_MultipleAVS(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	e2eTest := NewE2ETestCase(t, filepath.Dir(wd))
	defer e2eTest.Cleanup()

	err = runCommand(t,
		e2eTest.EgnPath(),
		"install",
		"--profile", "option-returner",
		"--run",
		"--no-prompt",
		"--tag", "tag-1",
		"--option.main-container-name", "main-service-1",
		"https://github.com/NethermindEth/mock-avs",
	)
	assert.NoError(t, err)

	checkMonitoringStack(t)
	checkContainerRunning(t, "main-service-1")

	err = runCommand(t,
		e2eTest.EgnPath(),
		"install",
		"--profile", "option-returner",
		"--run",
		"--no-prompt",
		"--tag", "tag-2",
		"--option.main-container-name", "main-service-2",
		"--option.main-port", "8081",
		"https://github.com/NethermindEth/mock-avs",
	)
	assert.NoError(t, err)

	checkMonitoringStack(t)

	checkContainerRunning(t, "main-service-1")
	checkContainerRunning(t, "main-service-2")
}
