package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	mockAVSRepo             = "https://github.com/NethermindEth/mock-avs"
	latestMockAVSVersion    = "v4.0.0"
	latestMockAVSCommitHash = "f9a1cbe784c9aa72c1da46aa86c7f910908b1969"
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
			checkInstanceNotInstalled(t, "mock-avs-default")
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
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "install command should succeed")
			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerRunning(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestInstall_FromCommitHash(t *testing.T) {
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
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--commit", latestMockAVSCommitHash, mockAVSRepo)
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "install command should succeed")
			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerRunning(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestInstall_ValidArgumentWithMonitoring(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := runCommand(t, egnPath, "init-monitoring")
			if err != nil {
				return err
			}
			return nil
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "install command should succeed")
			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerRunning(t, "option-returner")

			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")

			waitForMonitoring()
			checkGrafanaHealth(t)
			checkPrometheusTargetsUp(t, "egn_node_exporter:9100", optionReturnerIP+":8080")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestInstall_ValidArgumentNotRun(t *testing.T) {
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
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "install command should succeed")
			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerNotRunning(t, "option-returner")
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
			err := runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--tag", "integration", "--yes", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
			return err
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--tag", "integration", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "install command should fail with duplicated ID")
			checkInstanceInstalled(t, "mock-avs-integration")
			checkContainerRunning(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestInstall_DuplicatedContainerNameWithMonitoring(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := runCommand(t, egnPath, "init-monitoring")
			if err != nil {
				return err
			}
			err = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
			return err
		},
		// Act
		func(t *testing.T, egnPath string) {
			// Uses different tag, but docker compose create will fail because of duplicated container name
			// The install should fail but the monitoring stack should be running and the instance should be cleaned up
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--tag", "integration", "--yes", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "install command should fail with duplicated container name")

			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerRunning(t, "option-returner")
			checkInstanceNotInstalled(t, "mock-avs-integration")
			checkTemporaryPackageNotExisting(t, "mock-avs")

			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")

			waitForMonitoring()
			checkGrafanaHealth(t)
			checkPrometheusTargetsUp(t, "egn_node_exporter:9100", optionReturnerIP+":8080")
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
			runErr[0] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-1", "--option.main-container-name", "main-service-1", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
			runErr[1] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-2", "--option.main-container-name", "main-service-2", "--option.main-port", "8081", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
			runErr[2] = runCommand(t, egnPath, "install", "--profile", "health-checker", "--no-prompt", "--tag", "health-checker", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			for i, err := range runErr {
				assert.NoError(t, err, "install command (%d) should succeed", i)
			}

			checkInstanceInstalled(t, "mock-avs-option-returner-1")
			checkInstanceInstalled(t, "mock-avs-option-returner-2")
			checkContainerRunning(t, "main-service-1", "main-service-2")
			checkContainerNotRunning(t, "health-checker")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestInstall_MultipleAVSWithMonitoring(t *testing.T) {
	// Test context
	var (
		runErr [3]error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := runCommand(t, egnPath, "init-monitoring")
			if err != nil {
				return err
			}
			return nil
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr[0] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-1", "--option.main-container-name", "main-service-1", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
			time.Sleep(5 * time.Second)
			runErr[1] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-2", "--option.main-container-name", "main-service-2", "--option.main-port", "8081", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
			time.Sleep(5 * time.Second)
			runErr[2] = runCommand(t, egnPath, "install", "--profile", "health-checker", "--no-prompt", "--yes", "--tag", "health-checker", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			for i, err := range runErr {
				assert.NoError(t, err, "install command (%d) should succeed", i)
			}

			checkInstanceInstalled(t, "mock-avs-option-returner-1")
			checkInstanceInstalled(t, "mock-avs-option-returner-2")
			checkInstanceInstalled(t, "mock-avs-health-checker")
			checkContainerRunning(t, "main-service-1", "main-service-2", "health-checker")

			mainService1IP, err := getContainerIPByName("main-service-1", "eigenlayer")
			assert.NoError(t, err)
			mainService2IP, err := getContainerIPByName("main-service-2", "eigenlayer")
			assert.NoError(t, err)
			healthCheckerIP, err := getContainerIPByName("health-checker", "eigenlayer")
			assert.NoError(t, err)

			waitForMonitoring()
			checkPrometheusTargetsUp(t, "egn_node_exporter:9100", mainService1IP+":8080", mainService2IP+":8080", healthCheckerIP+":8090")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestInstall_HighRequirements(t *testing.T) {
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
			runErr = runCommand(t, egnPath, "install", "--profile", "high-requirements", "--no-prompt", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "install command should fail")
		},
	)
	// Run test case
	e2eTest.run()
}
