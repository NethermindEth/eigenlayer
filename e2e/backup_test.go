package e2e

import (
	"testing"
	"time"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/stretchr/testify/assert"
)

func TestBackupInstance(t *testing.T) {
	// Test context
	var (
		backupErr error
		after     time.Time
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			after = time.Now()
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
			checkBackupExist(t, "mock-avs-default", time.Now(), after)
		},
	)
	// Run test case
	e2eTest.run()
}
