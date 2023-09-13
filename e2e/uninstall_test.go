package e2e

import (
	"testing"
	"time"

	"github.com/NethermindEth/eigenlayer/e2e/docker"
	"github.com/stretchr/testify/require"
)

// Test_Uninstall_After_Stop checks that the uninstall command removes all the
// container of the mock-avs option-returner profile without error
func Test_Uninstall(t *testing.T) {
	// Test context
	var (
		uninstallErr error
	)
	e2eTest := newE2ETestCase(t,
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			// Install the mock-avs option-returner profile
			err = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", latestMockAVSPkgVersion, mockAVSPkgRepo)
			if err != nil {
				return err
			}
			// Wait until the container is running
			return docker.WaitUntilRunning("option-returner", 10*time.Second)
		},
		func(t *testing.T, egnPath string) {
			uninstallErr = runCommand(t, egnPath, "uninstall", "mock-avs-default")
		},
		func(t *testing.T) {
			require.NoError(t, uninstallErr, "uninstall command should not return an error")
			checkInstanceNotInstalled(t, "mock-avs-default")
			checkContainerNotExisting(t, "option-returner")
		})
	e2eTest.run()
}

// Test_Uninstall_After_Stop checks that the uninstall command removes all the
// container of the mock-avs option-returner profile without error when the
// AVS instance is stopped.
func Test_Uninstall_After_Stop(t *testing.T) {
	// Test context
	var (
		uninstallErr error
	)
	e2eTest := newE2ETestCase(t,
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			// Install the mock-avs option-returner profile
			err = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", latestMockAVSPkgVersion, mockAVSPkgRepo)
			if err != nil {
				return err
			}
			// Wait until the container is running
			err = docker.WaitUntilRunning("option-returner", 10*time.Second)
			if err != nil {
				return err
			}
			// Stop the AVS instance
			return runCommand(t, egnPath, "stop", "mock-avs-default")
		},
		func(t *testing.T, egnPath string) {
			uninstallErr = runCommand(t, egnPath, "uninstall", "mock-avs-default")
		},
		func(t *testing.T) {
			require.NoError(t, uninstallErr, "uninstall command should not return an error")
			checkInstanceNotInstalled(t, "mock-avs-default")
			checkContainerNotExisting(t, "option-returner")
		})
	e2eTest.run()
}

// Test_Uninstall_NonExistingInstance checks that the uninstall command succeeds.
// if the instance does not exist.
func Test_Uninstall_NonExistingInstance(t *testing.T) {
	// Test context
	var (
		uninstallErr error
	)
	e2eTest := newE2ETestCase(t,
		nil,
		func(t *testing.T, egnPath string) {
			uninstallErr = runCommand(t, egnPath, "uninstall", "mock-avs-default")
		},
		func(t *testing.T) {
			require.NoError(t, uninstallErr, "uninstall command should success")
		})
	e2eTest.run()
}
