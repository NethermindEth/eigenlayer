package e2e

import (
	"regexp"
	"testing"
	"time"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestBackupInstance(t *testing.T) {
	// Test context
	var (
		backupErr error
		start     time.Time
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			start = time.Now()
			err := buildMockAvsImagesLatest(t)
			if err != nil {
				return err
			}
			// Install latest version
			return runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
		},
		// Act
		func(t *testing.T, egnPath string) {
			backupErr = runCommand(t, egnPath, "backup", "mock-avs-default")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, backupErr, "backup command should succeed")
			checkBackupExist(t, "mock-avs-default", start, time.Now())
		},
	)
	// Run test case
	e2eTest.run()
}

func TestBackupList(t *testing.T) {
	// Test context
	var (
		out       []byte
		backupErr error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := buildMockAvsImagesLatest(t)
			if err != nil {
				return err
			}
			// Install latest version
			err = runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			return runCommand(t, egnPath, "backup", "mock-avs-default")
		},
		// Act
		func(t *testing.T, egnPath string) {
			out, backupErr = runCommandOutput(t, egnPath, "backup", "ls")
		},
		// Assert
		func(t *testing.T) {
			t.Log(string(out))
			assert.NoError(t, backupErr, "backup ls command should succeed")
			assert.Regexp(t, regexp.MustCompile(
				`AVS Instance ID     TIMESTAMP              SIZE    
mock-avs-default    .*    9KiB`),
				string(out))
		},
	)
	// Run test case
	e2eTest.run()
}
