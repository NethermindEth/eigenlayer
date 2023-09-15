package e2e

import (
	"testing"
	"time"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		func(t *testing.T, egnPath string) error {
			return buildMockAvsImages(t)
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "install command should succeed")
			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerRunning(t, "option-returner")
			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")
			checkAVSHealth(t, optionReturnerIP, "8080", 200)
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
		func(t *testing.T, egnPath string) error {
			return buildMockAvsImages(t)
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--commit", common.MockAvsPkg.CommitHash(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "install command should succeed")
			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerRunning(t, "option-returner")
			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")
			checkAVSHealth(t, optionReturnerIP, "8080", 200)
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
			if err := buildMockAvsImages(t); err != nil {
				return err
			}
			err := runCommand(t, egnPath, "init-monitoring")
			if err != nil {
				return err
			}
			return nil
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "install command should succeed")
			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerRunning(t, "option-returner")

			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")
			checkAVSHealth(t, optionReturnerIP, "8080", 200)

			checkGrafanaHealth(t)
			checkPrometheusTargetsUp(t, "egn_node_exporter:9100", optionReturnerIP+":8080")
			checkPrometheusLabels(t, optionReturnerIP+":8080")
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
		func(t *testing.T, egnPath string) error {
			return buildMockAvsImages(t)
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
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
			if err := buildMockAvsImages(t); err != nil {
				return err
			}
			return runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--tag", "integration", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--tag", "integration", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "install command should fail with duplicated ID")
			checkInstanceInstalled(t, "mock-avs-integration")
			checkContainerRunning(t, "option-returner")
			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")
			checkAVSHealth(t, optionReturnerIP, "8080", 200)
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
			if err := buildMockAvsImages(t); err != nil {
				return err
			}
			err := runCommand(t, egnPath, "init-monitoring")
			if err != nil {
				return err
			}
			return runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
		},
		// Act
		func(t *testing.T, egnPath string) {
			// Uses different tag, but docker compose create will fail because of duplicated container name
			// The install should fail but the monitoring stack should be running and the instance should be cleaned up
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--tag", "integration", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
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
			checkAVSHealth(t, optionReturnerIP, "8080", 200)

			checkGrafanaHealth(t)
			checkPrometheusTargetsUp(t, "egn_node_exporter:9100", optionReturnerIP+":8080")
			checkPrometheusLabels(t, optionReturnerIP+":8080")
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
		func(t *testing.T, egnPath string) error {
			return buildMockAvsImages(t)
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr[0] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-1", "--option.main-container-name", "main-service-1", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
			runErr[1] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-2", "--option.main-container-name", "main-service-2", "--option.main-port", "8081", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345679", "--option.test-option-enum-hidden", "option2", common.MockAvsPkg.Repo())
			runErr[2] = runCommand(t, egnPath, "install", "--profile", "health-checker", "--no-prompt", "--tag", "health-checker", "--version", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo())
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

			mainService1IP, err := getContainerIPByName("main-service-1", "eigenlayer")
			require.NoError(t, err, "failed to get main-service-1 container IP")
			mainService2IP, err := getContainerIPByName("main-service-2", "eigenlayer")
			require.NoError(t, err, "failed to get main-service-2 container IP")
			checkAVSHealth(t, mainService1IP, "8080", 200)
			checkAVSHealth(t, mainService2IP, "8080", 200)
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
			if err := buildMockAvsImages(t); err != nil {
				return err
			}
			err := runCommand(t, egnPath, "init-monitoring")
			if err != nil {
				return err
			}
			return nil
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr[0] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-1", "--option.main-container-name", "main-service-1", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option1", common.MockAvsPkg.Repo())
			time.Sleep(5 * time.Second)
			runErr[1] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-2", "--option.main-container-name", "main-service-2", "--option.main-port", "8081", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "123456789", "--option.test-option-enum-hidden", "option2", common.MockAvsPkg.Repo())
			time.Sleep(5 * time.Second)
			runErr[2] = runCommand(t, egnPath, "install", "--profile", "health-checker", "--no-prompt", "--yes", "--tag", "health-checker", "--version", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo())
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
			checkAVSHealth(t, mainService1IP, "8080", 200)
			checkAVSHealth(t, mainService2IP, "8080", 200)
			checkAVSHealth(t, healthCheckerIP, "8090", 200)

			checkPrometheusTargetsUp(t, "egn_node_exporter:9100", mainService1IP+":8080", mainService2IP+":8080", healthCheckerIP+":8090")
			checkPrometheusLabels(t, mainService1IP+":8080", mainService2IP+":8080", healthCheckerIP+":8090")
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
		func(t *testing.T, egnPath string) error {
			return buildMockAvsImages(t)
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "high-requirements", "--no-prompt", "--version", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo())
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "install command should fail")
			checkInstanceNotInstalled(t, "mock-avs-default")
			checkContainerNotExisting(t, "high-requirements")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestInstall_ProfileWithHiddenOptionsNotSet(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			return buildMockAvsImages(t)
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo())
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "install command should fail")
			checkInstanceNotInstalled(t, "mock-avs-default")
			checkContainerNotExisting(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}
