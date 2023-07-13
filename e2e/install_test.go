package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstall_WithoutArguments(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		nil,
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install")
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "install command should fail without arguments")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestInstall_ValidArgument(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		nil,
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "install command should succeed")
			checkContainerRunning(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestInstall_DuplicatedID(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--tag", "integration", "https://github.com/NethermindEth/mock-avs")
			if err != nil {
				return err
			}
			return nil
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--tag", "integration", "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "install command should fail with duplicated ID")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestInstall_MultipleAVS(t *testing.T) {
	// Test context
	var (
		runErr [3]error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		nil,
		// Act
		func(t *testing.T, egnPath string) {
			runErr[0] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--tag", "option-returner-1", "--option.main-container-name", "main-service-1", "https://github.com/NethermindEth/mock-avs")
			runErr[1] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--tag", "option-returner-2", "--option.main-container-name", "main-service-2", "--option.main-port", "8081", "https://github.com/NethermindEth/mock-avs")
			runErr[2] = runCommand(t, egnPath, "install", "--profile", "health-checker", "--no-prompt", "--tag", "health-checker", "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			for i, err := range runErr {
				assert.NoError(t, err, "install command (%d) should succeed", i)
			}

			checkContainerRunning(t, "main-service-1", "main-service-2", "health-checker")

			mainService1IP, err := getContainerIPByName("main-service-1", "eigenlayer")
			assert.NoError(t, err)
			mainService2IP, err := getContainerIPByName("main-service-2", "eigenlayer")
			assert.NoError(t, err)
			healthCheckerIP, err := getContainerIPByName("health-checker", "eigenlayer")
			assert.NoError(t, err)

			waitForMonitoring()
			checkPrometheusTargets(t, "egn_node_exporter:9100", mainService1IP+":8080", mainService2IP+":8080", healthCheckerIP+":8090")
		},
	)
	// Run test case
	e2eTest.run()
}
