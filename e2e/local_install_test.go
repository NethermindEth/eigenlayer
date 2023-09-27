package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLocalInstall(t *testing.T) {
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			if err := buildMockAvsImagesLatest(t); err != nil {
				return err
			}
			err := os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug", "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "local-install command should succeed")
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

func TestLocalInstallNotRunning(t *testing.T) {
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			if err := buildMockAvsImagesLatest(t); err != nil {
				return err
			}
			err := os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--log-debug", "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "local-install command should succeed")
			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerNotRunning(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestLocalInstallWithMonitoring(t *testing.T) {
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			if err := buildMockAvsImagesLatest(t); err != nil {
				return err
			}
			err := runCommand(t, egnPath, "init-monitoring")
			if err != nil {
				return err
			}
			err = os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug", "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "local-install command should succeed")

			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerRunning(t, "option-returner")
			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")
			checkAVSHealth(t, optionReturnerIP, "8080", 200)

			checkGrafanaHealth(t)
			checkPrometheusTargetsUp(t, "egn_node_exporter:9100", optionReturnerIP+":8080")
			checkPrometheusLabels(t, optionReturnerIP+":8080")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestLocalInstallInvalidManifest(t *testing.T) {
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			if err := buildMockAvsImagesLatest(t); err != nil {
				return err
			}
			err := os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// Modify manifest file to make it invalid
			err = os.WriteFile(filepath.Join(pkgDir, "pkg", "manifest.yml"), []byte("invalid: invalid"), 0o644)
			if err != nil {
				return err
			}
			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug", "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3")
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "local-install command should fail")
			checkInstanceNotInstalled(t, "mock-avs-default")
			checkContainerNotExisting(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestLocalInstallInvalidManifestCleanup(t *testing.T) {
	// Try to install with an invalid manifest and check that the temp directory is cleaned up
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			if err := buildMockAvsImagesLatest(t); err != nil {
				return err
			}
			err := os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// Modify manifest file to make it invalid
			err = os.WriteFile(filepath.Join(pkgDir, "pkg", "manifest.yml"), []byte("invalid: invalid"), 0o644)
			if err != nil {
				return err
			}
			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug", "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3")
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "local-install command should fail")
			checkInstanceNotInstalled(t, "mock-avs-default")
			checkContainerNotExisting(t, "option-returner")
			checkTemporaryPackageNotExisting(t, "mock-avs")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestLocalInstallInvalidManifestCleanupWithMonitoring(t *testing.T) {
	// Try to install with an invalid manifest and check that the temp directory is cleaned up
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			if err := buildMockAvsImagesLatest(t); err != nil {
				return err
			}
			err := runCommand(t, egnPath, "init-monitoring")
			if err != nil {
				return err
			}
			err = os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// Modify manifest file to make it invalid
			err = os.WriteFile(filepath.Join(pkgDir, "pkg", "manifest.yml"), []byte("invalid: invalid"), 0o644)
			if err != nil {
				return err
			}
			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug", "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3")
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "local-install command should fail")
			checkInstanceNotInstalled(t, "mock-avs-default")
			checkContainerNotExisting(t, "option-returner")
			checkTemporaryPackageNotExisting(t, "mock-avs")

			checkPrometheusTargetsDown(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestLocalInstallInvalidComposeCleanup(t *testing.T) {
	// Try to install with an invalid compose (docker compose create fails) and check that the temp directory is cleaned up
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			if err := buildMockAvsImagesLatest(t); err != nil {
				return err
			}
			err := os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// Modify manifest file to make it invalid
			err = os.WriteFile(filepath.Join(pkgDir, "pkg", "option-returner", "docker-compose.yml"), []byte("invalid: invalid"), 0o644)
			if err != nil {
				return err
			}
			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug", "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3")
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "local-install command should fail")
			checkInstanceNotInstalled(t, "mock-avs-default")
			checkContainerNotExisting(t, "option-returner")
			checkTemporaryPackageNotExisting(t, "mock-avs")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestLocalInstall_DuplicatedContainerNameWithMonitoring(t *testing.T) {
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			if err := buildMockAvsImagesLatest(t); err != nil {
				return err
			}
			err := runCommand(t, egnPath, "init-monitoring")
			if err != nil {
				return err
			}
			err = os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// remove .git folder
			err = os.RemoveAll(filepath.Join(pkgDir, ".git"))
			if err != nil {
				return err
			}
			return runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug", "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3")
		},
		// Act
		func(t *testing.T, egnPath string) {
			// Uses different tag, but docker compose create will fail because of duplicated container name
			// The install should fail but the monitoring stack should be running and the instance should be cleaned up
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug", "--tag", "integration")
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "install command should fail with duplicated container name")

			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerRunning(t, "option-returner")
			checkInstanceNotInstalled(t, "mock-avs-integration")
			checkTemporaryPackageNotExisting(t, "mock-avs")

			optionReturnerIP, err := getContainerIPByName("option-returner", "eigenlayer")
			require.NoError(t, err, "failed to get option-returner container IP")
			checkAVSHealth(t, optionReturnerIP, "8080", 200)

			checkGrafanaHealth(t)
			checkPrometheusTargetsUp(t, "egn_node_exporter:9100", optionReturnerIP+":8080")
			checkPrometheusLabels(t, optionReturnerIP+":8080")
		},
	)
	// Run test case
	e2eTest.run()
}

func TestLocalInstall_ProfileWithHiddenOptionsNotSet(t *testing.T) {
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			if err := buildMockAvsImagesLatest(t); err != nil {
				return err
			}
			err := os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "option-returner", "--run", "--log-debug")
		},
		// Assert
		func(t *testing.T) {
			assert.Error(t, runErr, "local-install command should fail")
			checkInstanceNotInstalled(t, "mock-avs-default")
			checkContainerNotExisting(t, "option-returner")
		},
	)
	// Run test case
	e2eTest.run()
}

const newOptions = `
  - name: "hidden-option-test"
    target: HIDDEN_WITH_DEFAULT_TEST
    type: str
    default: test
    hidden: true
    help: "Hidden option test"
  - name: "option-without-default-test"
    target: OPTION_WITHOUT_DEFAULT_TEST
    type: str
    validate:
      re2_regex: "^eigen.*" # Words that start with eigen
    help: "Option without default"
`

func TestLocalInstall_OptionsWithoutDefault(t *testing.T) {
	// Test context
	var (
		testDir = t.TempDir()
		pkgDir  = filepath.Join(testDir, "mock-avs")
		runErr  error
	)
	// Build test case
	e2eTest := newE2ETestCase(
		t,
		// Arrange
		func(t *testing.T, egnPath string) error {
			if err := buildMockAvsImages(t); err != nil {
				return err
			}
			err := os.MkdirAll(pkgDir, 0o755)
			if err != nil {
				return err
			}
			err = runCommand(t, "git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir)
			if err != nil {
				return err
			}
			// Modify health-checker to add a hidden option will default and a non-hidden option without default.
			// Both TARGETS should be the empty string in the .env.
			f, err := os.OpenFile(filepath.Join(pkgDir, "pkg", "health-checker", "profile.yml"), os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				t.Fatalf("failed to open profile.yml: %v", err)
			}
			defer f.Close()

			if _, err = f.WriteString(newOptions); err != nil {
				t.Fatalf("failed to write to profile.yml: %v", err)
			}

			// remove .git folder
			return os.RemoveAll(filepath.Join(pkgDir, ".git"))
		},
		// Act
		func(t *testing.T, egnPath string) {
			runErr = runCommand(t, egnPath, "local-install", pkgDir, "--profile", "health-checker", "--run", "--log-debug")
		},
		// Assert
		func(t *testing.T) {
			assert.NoError(t, runErr, "local-install command should succeed")
			checkInstanceInstalled(t, "mock-avs-default")
			checkContainerRunning(t, "health-checker")
			healthCheckerIP, err := getContainerIPByName("health-checker", "eigenlayer")
			require.NoError(t, err, "failed to get health-checker container IP")
			checkAVSHealth(t, healthCheckerIP, "8090", 200)
			checkEnvTargets(t, "mock-avs-default", "HIDDEN_WITH_DEFAULT_TEST", "OPTION_WITHOUT_DEFAULT_TEST")
		},
	)
	// Run test case
	e2eTest.run()
}
