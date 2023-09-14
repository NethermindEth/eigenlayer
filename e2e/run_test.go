package e2e

import (
	"testing"
	"time"

	"github.com/NethermindEth/eigenlayer/e2e/docker"
	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/stretchr/testify/require"
)

// Test_run checks that the all the containers of the mock-avs package are
// running after the run command is executed without errors.
func Test_Run(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	e2eTest := newE2ETestCase(t,
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			return runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--version", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo())
		},
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "run", "mock-avs-default")
		},
		func(t *testing.T) {
			require.NoError(t, runErr, "run command should succeed")
			checkContainerRunning(t, "option-returner")
			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")
			checkAVSHealth(t, optionReturnerIP, "8080", 200)
		})
	e2eTest.run()
}

// Test_Run_StoppedInstance checks that the run command starts the instance
// when it is already installed but stopped.
func Test_Run_StoppedInstance(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	e2eTest := newE2ETestCase(t,
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			err = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			return runCommand(t, egnPath, "stop", "mock-avs-default")
		},
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "run", "mock-avs-default")
		},
		func(t *testing.T) {
			require.NoError(t, runErr, "run command should succeed")
			checkContainerRunning(t, "option-returner")
			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")
			checkAVSHealth(t, optionReturnerIP, "8080", 200)
		})
	e2eTest.run()
}

// Test_Run_AlreadyRunningInstance checks that the run command doesn't fail when
// the instance is already running.
func Test_Run_AlreadyRunningInstance(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	e2eTest := newE2ETestCase(t,
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			err = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			// Wait until the container is running
			return docker.WaitUntilRunning("option-returner", 10*time.Second)
		},
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "run", "mock-avs-default")
		},
		func(t *testing.T) {
			require.NoError(t, runErr, "run command should succeed")
			checkContainerRunning(t, "option-returner")
			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")
			checkAVSHealth(t, optionReturnerIP, "8080", 200)
		})
	e2eTest.run()
}

// Test_Run_NonExistingInstance checks that the run command fails when the
// instance does not exist.
func Test_Run_NonExistingInstance(t *testing.T) {
	// Test context
	var (
		runErr error
	)
	e2eTest := newE2ETestCase(t,
		nil,
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "run", "mock-avs-default")
		},
		func(t *testing.T) {
			require.Error(t, runErr, "run command should fail")
		})
	e2eTest.run()
}
