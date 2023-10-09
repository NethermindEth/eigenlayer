package e2e

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/maps"
)

func TestRestore(t *testing.T) {
	// Test context
	var (
		instanceState json.RawMessage
		instanceEnv   map[string]string
		backupId      string
		restoreErr    error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, eigenlayerPath string) error {
			// Build option returner image
			err := buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			// Install option returner AVS
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			// Save instance env
			instanceEnv = loadEnv(t, "mock-avs-default")
			// Save instance state
			instanceState = loadStateJSON(t, "mock-avs-default")
			// Backup AVS
			backupOut, err := runCommandOutput(t, eigenlayerPath, "backup", "mock-avs-default")
			if err != nil {
				return err
			}
			// Parse backup id
			r := regexp.MustCompile(`.*Backup created with id: (?P<backup_id>[a-f0-9]+).*`)
			matches := r.FindSubmatch(backupOut)
			require.Len(t, matches, 2)
			backupId = string(matches[1])
			// Uninstall AVS
			return runCommand(t, eigenlayerPath, "uninstall", "mock-avs-default")
		},
		// Act
		func(t *testing.T, egnPath string) {
			restoreErr = runCommand(t, egnPath, "restore", backupId)
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, restoreErr, "restore command should not fail")
			checkInstanceExists(t, "mock-avs-default")
			checkEnvValues(t, "mock-avs-default", instanceEnv)
			checkStateJson(t, "mock-avs-default", instanceState)
		},
	)
	// Run test case
	e2eTest.run()
}

func TestRestore_InstanceExists(t *testing.T) {
	// Test context
	var (
		instanceState json.RawMessage
		instanceEnv   map[string]string
		backupId      string
		restoreErr    error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, eigenlayerPath string) error {
			// Build option returner image
			err := buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			// Install option returner AVS
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			// Save instance env
			instanceEnv = loadEnv(t, "mock-avs-default")
			// Save instance state
			instanceState = loadStateJSON(t, "mock-avs-default")
			// Backup AVS
			backupOut, err := runCommandOutput(t, eigenlayerPath, "backup", "mock-avs-default")
			if err != nil {
				return err
			}
			// Parse backup id
			r := regexp.MustCompile(`.*Backup created with id: (?P<backup_id>[a-f0-9]+).*`)
			matches := r.FindSubmatch(backupOut)
			require.Len(t, matches, 2)
			backupId = string(matches[1])
			// Uninstall AVS instance
			err = runCommand(t, eigenlayerPath, "uninstall", "mock-avs-default")
			if err != nil {
				return err
			}
			// Install new AVS instance, with the same id, but with different
			// options values
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--no-prompt", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "87654321", "--option.test-option-enum-hidden", "option1", common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			// Check that the new instance env is different from the old one
			newInstanceEnv := loadEnv(t, "mock-avs-default")
			require.False(t, maps.Equal(instanceEnv, newInstanceEnv), "instance env should be different")
			return nil
		},
		// Act
		func(t *testing.T, egnPath string) {
			restoreErr = runCommand(t, egnPath, "restore", backupId)
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, restoreErr, "restore command should not fail")
			checkInstanceExists(t, "mock-avs-default")
			checkEnvValues(t, "mock-avs-default", instanceEnv)
			checkStateJson(t, "mock-avs-default", instanceState)
		},
	)
	// Run test case
	e2eTest.run()
}

func TestRestore_InstanceExistsRunning(t *testing.T) {
	// Test context
	var (
		instanceState json.RawMessage
		instanceEnv   map[string]string
		backupId      string
		restoreErr    error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, eigenlayerPath string) error {
			// Build option returner image
			err := buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			// Install option returner AVS
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			// Save instance env
			instanceEnv = loadEnv(t, "mock-avs-default")
			// Save instance state
			instanceState = loadStateJSON(t, "mock-avs-default")
			// Backup AVS
			backupOut, err := runCommandOutput(t, eigenlayerPath, "backup", "mock-avs-default")
			if err != nil {
				return err
			}
			// Parse backup id
			r := regexp.MustCompile(`.*Backup created with id: (?P<backup_id>[a-f0-9]+).*`)
			matches := r.FindSubmatch(backupOut)
			require.Len(t, matches, 2)
			backupId = string(matches[1])
			// Uninstall AVS instance
			err = runCommand(t, eigenlayerPath, "uninstall", "mock-avs-default")
			if err != nil {
				return err
			}
			// Install new AVS instance, with the same id, but with different
			// options values
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--yes", "--no-prompt", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "87654321", "--option.test-option-enum-hidden", "option1", common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			// Check that the new instance env is different from the old one
			newInstanceEnv := loadEnv(t, "mock-avs-default")
			require.False(t, maps.Equal(instanceEnv, newInstanceEnv), "instance env should be different")
			return nil
		},
		// Act
		func(t *testing.T, egnPath string) {
			restoreErr = runCommand(t, egnPath, "restore", backupId)
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, restoreErr, "restore command should not fail")
			checkInstanceExists(t, "mock-avs-default")
			checkEnvValues(t, "mock-avs-default", instanceEnv)
			checkStateJson(t, "mock-avs-default", instanceState)
		},
	)
	// Run test case
	e2eTest.run()
}

func TestRestore_Run(t *testing.T) {
	// Test context
	var (
		instanceState json.RawMessage
		instanceEnv   map[string]string
		backupId      string
		restoreErr    error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, eigenlayerPath string) error {
			// Build option returner image
			err := buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			// Install option returner AVS
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			// Save instance env
			instanceEnv = loadEnv(t, "mock-avs-default")
			// Save instance state
			instanceState = loadStateJSON(t, "mock-avs-default")
			// Backup AVS
			backupOut, err := runCommandOutput(t, eigenlayerPath, "backup", "mock-avs-default")
			if err != nil {
				return err
			}
			// Parse backup id
			r := regexp.MustCompile(`.*Backup created with id: (?P<backup_id>[a-f0-9]+).*`)
			matches := r.FindSubmatch(backupOut)
			require.Len(t, matches, 2)
			backupId = string(matches[1])
			// Uninstall AVS
			return runCommand(t, eigenlayerPath, "uninstall", "mock-avs-default")
		},
		// Act
		func(t *testing.T, egnPath string) {
			restoreErr = runCommand(t, egnPath, "restore", "--run", backupId)
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, restoreErr, "restore command should not fail")
			checkInstanceExists(t, "mock-avs-default")
			checkEnvValues(t, "mock-avs-default", instanceEnv)
			checkStateJson(t, "mock-avs-default", instanceState)
			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")
			checkAVSHealth(t, optionReturnerIP, "8080", 200)
		},
	)
	// Run test case
	e2eTest.run()
}
