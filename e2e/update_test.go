package e2e

import (
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	// Test context
	var (
		initialVersion = "v5.4.0"
		updateVersion  = common.MockAvsPkg.Version()
		updateError    error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			return runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", initialVersion, common.MockAvsPkg.Repo())
		},
		// Act
		func(t *testing.T, egnPath string) {
			updateError = runCommand(t, egnPath, "update", "--no-prompt", "--option.test-option-hidden", "12345678", "mock-avs-default", updateVersion)
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, updateError, "update command should succeed")
			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerNotRunning(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestUpdate_Run(t *testing.T) {
	// Test context
	var (
		initialVersion = "v5.4.0"
		updateCommit   = common.MockAvsPkg.CommitHash()
		updateError    error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			return runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", initialVersion, common.MockAvsPkg.Repo())
		},
		// Act
		func(t *testing.T, egnPath string) {
			updateError = runCommand(t, egnPath, "update", "--yes", "--no-prompt", "--option.test-option-hidden", "12345678", "mock-avs-default", updateCommit)
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, updateError, "update command should succeed")
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

func TestUpdate_SameVersion(t *testing.T) {
	// Test context
	var (
		version     = common.MockAvsPkg.Version()
		updateError error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			return runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--option.test-option-hidden", "12345678", "--yes", "--version", version, common.MockAvsPkg.Repo())
		},
		// Act
		func(t *testing.T, egnPath string) {
			updateError = runCommand(t, egnPath, "update", "--yes", "--no-prompt", "--option.test-option-hidden", "12345678", "mock-avs-default", version)
		},
		// Assert
		func(t *testing.T) {
			require.NoError(t, updateError, "update command should not fail")
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

func TestUpdate_OldVersion(t *testing.T) {
	// Test context
	var (
		installVersion = common.MockAvsPkg.Version()
		updateVersion  = "v5.4.0"
		updateError    error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			return runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--option.test-option-hidden", "12345678", "--yes", "--version", installVersion, common.MockAvsPkg.Repo())
		},
		// Act
		func(t *testing.T, egnPath string) {
			updateError = runCommand(t, egnPath, "update", "--yes", "--no-prompt", "--option.test-option-hidden", "12345678", "mock-avs-default", updateVersion)
		},
		// Assert
		func(t *testing.T) {
			require.Error(t, updateError, "update command should fail")
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

func TestUpdate_SameCommit(t *testing.T) {
	// Test context
	var (
		installVersion = common.MockAvsPkg.Version()
		updateCommit   = "a3406616b848164358fdd24465b8eecda5f5ae34"
		updateError    error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			return runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--option.test-option-hidden", "12345678", "--yes", "--version", installVersion, common.MockAvsPkg.Repo())
		},
		// Act
		func(t *testing.T, egnPath string) {
			updateError = runCommand(t, egnPath, "update", "--yes", "--no-prompt", "--option.test-option-hidden", "12345678", "mock-avs-default", updateCommit)
		},
		// Assert
		func(t *testing.T) {
			require.NoError(t, updateError, "update command should not fail")
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

func TestUpdate_OldCommit(t *testing.T) {
	// Test context
	var (
		installVersion = common.MockAvsPkg.Version()
		updateCommit   = "b64c50c15e53ae7afebbdbe210b834d1ee471043"
		updateError    error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImages(t)
			if err != nil {
				return err
			}
			return runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--option.test-option-hidden", "12345678", "--yes", "--version", installVersion, common.MockAvsPkg.Repo())
		},
		// Act
		func(t *testing.T, egnPath string) {
			updateError = runCommand(t, egnPath, "update", "--yes", "--no-prompt", "--option.test-option-hidden", "12345678", "mock-avs-default", updateCommit)
		},
		// Assert
		func(t *testing.T) {
			require.Error(t, updateError, "update command should fail")
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
