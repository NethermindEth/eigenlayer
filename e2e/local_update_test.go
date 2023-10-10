package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalUpdate(t *testing.T) {
	// Test context
	var (
		testDir        = t.TempDir()
		pkgDir         = filepath.Join(testDir, "mock-avs")
		initialVersion = "v5.4.0"
		updateVersion  = common.MockAvsPkg.Version()
		updateError    error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := buildOptionReturnerImage(t, "v0.1.0")
			if err != nil {
				return err
			}
			err = buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			// Clone initial version
			err = runCommand(t, "git", "clone", "--single-branch", "-b", initialVersion, common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// Local install initial version
			err = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug", "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3")
			if err != nil {
				return err
			}
			err = os.RemoveAll(pkgDir)
			if err != nil {
				return err
			}
			// Clone update version
			return runCommand(t, "git", "clone", "--single-branch", "-b", updateVersion, common.MockAvsPkg.Repo(), pkgDir)
		},
		// Act
		func(t *testing.T, egnPath string) {
			updateError = runCommand(t, egnPath, "local-update", "--no-prompt", "--option.test-option-hidden", "12345678", "mock-avs-default", pkgDir)
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

func TestLocalUpdate_Run(t *testing.T) {
	// Test context
	var (
		testDir        = t.TempDir()
		pkgDir         = filepath.Join(testDir, "mock-avs")
		initialVersion = "v5.4.0"
		updateVersion  = common.MockAvsPkg.Version()
		updateError    error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := buildOptionReturnerImage(t, "v0.1.0")
			if err != nil {
				return err
			}
			err = buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			// Clone initial version
			err = runCommand(t, "git", "clone", "--single-branch", "-b", initialVersion, common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// Local install initial version
			err = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug", "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3")
			if err != nil {
				return err
			}
			// Clone update version
			err = os.RemoveAll(pkgDir)
			if err != nil {
				return err
			}
			return runCommand(t, "git", "clone", "--single-branch", "-b", updateVersion, common.MockAvsPkg.Repo(), pkgDir)
		},
		// Act
		func(t *testing.T, egnPath string) {
			updateError = runCommand(t, egnPath, "local-update", "--yes", "--no-prompt", "--option.test-option-hidden", "12345678", "mock-avs-default", pkgDir)
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

func TestLocalUpdate_SamePackage(t *testing.T) {
	// Test context
	var (
		testDir     = t.TempDir()
		pkgDir      = filepath.Join(testDir, "mock-avs")
		updateError error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := buildOptionReturnerImage(t, "v0.1.0")
			if err != nil {
				return err
			}
			err = buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			// Clone initial version
			err = runCommand(t, "git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// Local install initial version
			return runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug", "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3")
		},
		// Act
		func(t *testing.T, egnPath string) {
			updateError = runCommand(t, egnPath, "local-update", "--yes", "--no-prompt", "--option.test-option-hidden", "12345678", "mock-avs-default", pkgDir)
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
