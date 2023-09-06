package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLs_NoAVS(t *testing.T) {
	// Test context
	var (
		out   []byte
		lsErr error
	)
	e2eTest := newE2ETestCase(t,
		nil,
		func(t *testing.T, eigenlayerPath string) {
			out, lsErr = runCommandOutput(t, eigenlayerPath, "ls")
		},
		func(t *testing.T) {
			assert.NoError(t, lsErr, "ls command should not return an error")
			assert.Equal(t, out, []byte(
				"AVS Instance ID    RUNNING    HEALTH    VERSION    COMMIT    COMMENT    \n",
			))
		})
	e2eTest.run()
}

func TestLs_NotRunning(t *testing.T) {
	// Test context
	var (
		out   []byte
		lsErr error
	)
	e2eTest := newE2ETestCase(t,
		func(t *testing.T, eigenlayerPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			return runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--no-prompt", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
		},
		func(t *testing.T, eigenlayerPath string) {
			out, lsErr = runCommandOutput(t, eigenlayerPath, "ls")
		},
		func(t *testing.T) {
			assert.NoError(t, lsErr, "ls command should not return an error")
			assert.Equal(t, out, []byte(
				"AVS Instance ID     RUNNING    HEALTH     VERSION    COMMIT          COMMENT    \n"+
					"mock-avs-default    false      unknown    "+latestMockAVSVersion+"     a7ca2dca2cc9               \n",
			))
		})
	e2eTest.run()
}

func TestLs_RunningHealthy(t *testing.T) {
	// Test context
	var (
		out   []byte
		lsErr error
	)
	e2eTest := newE2ETestCase(t,
		func(t *testing.T, eigenlayerPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--yes", "--no-prompt", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
			if err != nil {
				return err
			}
			return waitHealthy(t, "option-returner", 8080, "eigenlayer", 3*time.Second)
		},
		func(t *testing.T, eigenlayerPath string) {
			out, lsErr = runCommandOutput(t, eigenlayerPath, "ls")
		},
		func(t *testing.T) {
			assert.NoError(t, lsErr, "ls command should not return an error")
			assert.Equal(t, out, []byte(
				"AVS Instance ID     RUNNING    HEALTH     VERSION    COMMIT          COMMENT    \n"+
					"mock-avs-default    true       healthy    "+latestMockAVSVersion+"     a7ca2dca2cc9               \n",
			))
		})
	e2eTest.run()
}

func TestLs_RunningPartiallyHealthy(t *testing.T) {
	// Test context
	var (
		out   []byte
		lsErr error
	)
	e2eTest := newE2ETestCase(t,
		func(t *testing.T, eigenlayerPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--yes", "--no-prompt", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
			if err != nil {
				return err
			}
			err = waitHealthy(t, "option-returner", 8080, "eigenlayer", 3*time.Second)
			if err != nil {
				return err
			}
			return changeHealthStatus(t, "option-returner", 8080, "eigenlayer", 206)
		},
		func(t *testing.T, eigenlayerPath string) {
			out, lsErr = runCommandOutput(t, eigenlayerPath, "ls")
		},
		func(t *testing.T) {
			assert.NoError(t, lsErr, "ls command should not return an error")
			assert.Equal(t, out, []byte(
				"AVS Instance ID     RUNNING    HEALTH               VERSION    COMMIT          COMMENT    \n"+
					"mock-avs-default    true       partially healthy    "+latestMockAVSVersion+"     a7ca2dca2cc9               \n",
			))
		})
	e2eTest.run()
}

func TestLs_RunningUnhealthy(t *testing.T) {
	// Test context
	var (
		out   []byte
		lsErr error
	)
	e2eTest := newE2ETestCase(t,
		func(t *testing.T, eigenlayerPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--yes", "--no-prompt", "--version", latestMockAVSVersion, "https://github.com/NethermindEth/mock-avs")
			if err != nil {
				return err
			}
			err = waitHealthy(t, "option-returner", 8080, "eigenlayer", 3*time.Second)
			if err != nil {
				return err
			}
			return changeHealthStatus(t, "option-returner", 8080, "eigenlayer", 503)
		},
		func(t *testing.T, eigenlayerPath string) {
			out, lsErr = runCommandOutput(t, eigenlayerPath, "ls")
		},
		func(t *testing.T) {
			assert.NoError(t, lsErr, "ls command should not return an error")
			assert.Equal(t, out, []byte(
				"AVS Instance ID     RUNNING    HEALTH       VERSION    COMMIT          COMMENT    \n"+
					"mock-avs-default    true       unhealthy    "+latestMockAVSVersion+"     a7ca2dca2cc9               \n",
			))
		})
	e2eTest.run()
}
