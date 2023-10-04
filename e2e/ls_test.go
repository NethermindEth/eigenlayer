package e2e

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/NethermindEth/eigenlayer/internal/data"
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
			err := buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			return runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--no-prompt", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
		},
		func(t *testing.T, eigenlayerPath string) {
			out, lsErr = runCommandOutput(t, eigenlayerPath, "ls")
		},
		func(t *testing.T) {
			assert.NoError(t, lsErr, "ls command should not return an error")
			assert.Equal(t, out, []byte(
				"AVS Instance ID     RUNNING    HEALTH     VERSION    COMMIT          COMMENT    \n"+
					"mock-avs-default    false      unknown    "+common.MockAvsPkg.Version()+"     "+common.MockAvsPkg.CommitHash()[:12]+"               \n",
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
			err := buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--yes", "--no-prompt", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
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
					"mock-avs-default    true       healthy    "+common.MockAvsPkg.Version()+"     "+common.MockAvsPkg.CommitHash()[:12]+"               \n",
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
			err := buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--yes", "--no-prompt", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
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
					"mock-avs-default    true       partially healthy    "+common.MockAvsPkg.Version()+"     "+common.MockAvsPkg.CommitHash()[:12]+"               \n",
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
			err := buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--yes", "--no-prompt", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
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
					"mock-avs-default    true       unhealthy    "+common.MockAvsPkg.Version()+"     "+common.MockAvsPkg.CommitHash()[:12]+"               \n",
			))
		})
	e2eTest.run()
}

func TestLs_Comment(t *testing.T) {
	// Test context
	var (
		containerIP string
		out         []byte
		lsErr       error
	)
	e2eTest := newE2ETestCase(t,
		func(t *testing.T, eigenlayerPath string) error {
			err := buildOptionReturnerImageLatest(t)
			if err != nil {
				return err
			}
			err = runCommand(t, eigenlayerPath, "install", "--profile", "option-returner", "--yes", "--no-prompt", "--version", common.MockAvsPkg.Version(), "--option.test-option-hidden", "12345678", "--option.test-option-enum-hidden", "option3", common.MockAvsPkg.Repo())
			if err != nil {
				return err
			}
			err = waitHealthy(t, "option-returner", 8080, "eigenlayer", 3*time.Second)
			if err != nil {
				return err
			}
			containerIP, err = getContainerIPByName("option-returner", "eigenlayer")
			if err != nil {
				return err
			}
			err = changeHealthStatus(t, "option-returner", 8080, "eigenlayer", 503)
			if err != nil {
				return err
			}
			return changeStateAPIPort("mock-avs-default", "8081")
		},
		func(t *testing.T, eigenlayerPath string) {
			out, lsErr = runCommandOutput(t, eigenlayerPath, "ls")
		},
		func(t *testing.T) {
			assert.NoError(t, lsErr, "ls command should not return an error")
			var ds string
			for i := 0; i < len(containerIP)*2; i++ {
				ds += " "
			}
			assert.Equal(t, []byte(
				"AVS Instance ID     RUNNING    HEALTH     VERSION    COMMIT          COMMENT"+ds+"                                                                                                                                \n"+
					"mock-avs-default    true       unknown    "+common.MockAvsPkg.Version()+"     "+common.MockAvsPkg.CommitHash()[:12]+"    API container is running but health check failed: Get \"http://"+containerIP+":8081/eigen/node/health\": dial tcp "+containerIP+":8081: connect: connection refused    \n",
			), out)
		})
	e2eTest.run()
}

func changeStateAPIPort(instanceID string, port string) error {
	dirPath, err := dataDirPath()
	if err != nil {
		return err
	}
	stateFile, err := os.OpenFile(filepath.Join(dirPath, "nodes", instanceID, "state.json"), os.O_RDONLY, 0o644)
	if err != nil {
		return err
	}
	stateData, err := io.ReadAll(stateFile)
	if err != nil {
		return err
	}
	err = stateFile.Close()
	if err != nil {
		return err
	}
	var instance data.Instance
	err = json.Unmarshal(stateData, &instance)
	if err != nil {
		return err
	}
	instance.APITarget.Port = port
	newStateJson, err := json.MarshalIndent(&instance, "", "  ")
	if err != nil {
		return err
	}
	stateFile, err = os.OpenFile(filepath.Join(dirPath, "nodes", instanceID, "state.json"), os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	err = stateFile.Truncate(0)
	if err != nil {
		return err
	}
	_, err = stateFile.Write(newStateJson)
	if err != nil {
		return err
	}
	return stateFile.Close()
}
