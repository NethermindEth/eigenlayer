package e2e

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mockAVSRepo = "https://github.com/NethermindEth/mock-avs"

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
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "https://github.com/NethermindEth/mock-avs")
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
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "install command should succeed")
			checkContainerRunning(t, "option-returner")

			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")

			waitForMonitoring()
			checkGrafanaHealth(t)
			checkPrometheusTargets(t, "egn_node_exporter:9100", optionReturnerIP+":8080")
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
			runErr = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "install command should succeed")
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
			runErr[0] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-1", "--option.main-container-name", "main-service-1", "https://github.com/NethermindEth/mock-avs")
			runErr[1] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-2", "--option.main-container-name", "main-service-2", "--option.main-port", "8081", "https://github.com/NethermindEth/mock-avs")
			runErr[2] = runCommand(t, egnPath, "install", "--profile", "health-checker", "--no-prompt", "--tag", "health-checker", "https://github.com/NethermindEth/mock-avs")
		},
		// Assert
		func(t *testing.T) {
			for i, err := range runErr {
				assert.NoError(t, err, "install command (%d) should succeed", i)
			}

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
			runErr[0] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-1", "--option.main-container-name", "main-service-1", "https://github.com/NethermindEth/mock-avs")
			time.Sleep(5 * time.Second)
			runErr[1] = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--tag", "option-returner-2", "--option.main-container-name", "main-service-2", "--option.main-port", "8081", "https://github.com/NethermindEth/mock-avs")
			time.Sleep(5 * time.Second)
			runErr[2] = runCommand(t, egnPath, "install", "--profile", "health-checker", "--no-prompt", "--yes", "--tag", "health-checker", "https://github.com/NethermindEth/mock-avs")
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

func TestLocalInstall(t *testing.T) {
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", "v3.0.3", mockAVSRepo, pkgDir)
			if err != nil {
				return err
			}
			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "local-install command should succeed")
			checkContainerRunning(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestLocalInstallNotRunning(t *testing.T) {
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", "v3.0.3", mockAVSRepo, pkgDir)
			if err != nil {
				return err
			}
			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--log-debug")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "local-install command should succeed")
			checkContainerNotRunning(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestLocalInstallWithMonitoring(t *testing.T) {
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
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
			err = os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", "v3.0.3", mockAVSRepo, pkgDir)
			if err != nil {
				return err
			}
			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "local-install command should succeed")

			checkContainerRunning(t, "option-returner")
			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")

			waitForMonitoring()
			checkGrafanaHealth(t)
			checkPrometheusTargets(t, "egn_node_exporter:9100", optionReturnerIP+":8080")
		},
	)
	// Run test case
	e2eTest.run()
}
