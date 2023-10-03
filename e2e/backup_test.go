package e2e

import (
	"regexp"
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackupInstance(t *testing.T) {
	// Test context
	var (
		output    []byte
		backupErr error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			err := buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			// Install latest version
			return runCommand(t, egnPath, "install", "--profile", "option-returner", "--no-prompt", "--yes", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
		},
		// Act
		func(t *testing.T, egnPath string) {
			output, backupErr = runCommandOutput(t, egnPath, "backup", "mock-avs-default")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, backupErr, "backup command should succeed")

			r := regexp.MustCompile(`.*"Backup created with id: (?P<backupId>[a-f0-9]+)".*`)
			match := r.FindSubmatch(output)
			require.Len(t, match, 2, "backup command output should match regex")
			instanceId := string(match[1])

			checkBackupExist(t, instanceId)
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
			err := buildOptionReturnerImageLatest(t)
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
			assert.Regexp(t, regexp.MustCompile(`ID\s+AVS Instance ID\s+VERSION\s+COMMIT\s+TIMESTAMP\s+SIZE\s+URL\s+`+
				`[a-f0-9]{8}\s+mock-avs-default\s+v5\.5\.1\s+d5af645fffb93e8263b099082a4f512e1917d0af\s+.*\s+10KiB\s+https://github.com/NethermindEth/mock-avs-pkg\s+`),
				string(out))
		},
	)
	// Run test case
	e2eTest.run()
}
