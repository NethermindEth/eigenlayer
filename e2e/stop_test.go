package e2e

import (
	"testing"
	"time"

	"github.com/NethermindEth/eigenlayer/e2e/docker"
	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/stretchr/testify/require"
)

// Test_Stop checks that the stop command stops all the container of the mock-avs
// option-returner profile without error.
func Test_Stop(t *testing.T) {
	// Test context
	var (
		stopErr error
	)
	e2eTest := newE2ETestCase(t,
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			err = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			err = runCommand(t, egnPath, "run", "mock-avs-default")
			if err != nil {
				return err
			}
			return docker.WaitUntilRunning("option-returner", 10*time.Second)
		},
		func(t *testing.T, egnPath string) {
			stopErr = runCommand(t, egnPath, "stop", "mock-avs-default")
		},
		func(t *testing.T) {
			require.NoError(t, stopErr, "stop command should succeed")
			checkContainerNotRunning(t, "option-returner")
		})
	e2eTest.run()
}

// Test_Stop_NonExistingInstance checks that the stop command fails when the
// instance does not exist.
func Test_Stop_NonExistingInstance(t *testing.T) {
	// Test context
	var (
		stopErr error
	)
	e2eTest := newE2ETestCase(t,
		nil,
		func(t *testing.T, egnPath string) {
			stopErr = runCommand(t, egnPath, "stop", "mock-avs-default")
		},
		func(t *testing.T) {
			require.Error(t, stopErr, "stop command should fail")
		})
	e2eTest.run()
}
