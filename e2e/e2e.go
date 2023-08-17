package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type (
	e2eArranger func(t *testing.T, eigenlayerPath string) error
	e2eAct      func(t *testing.T, eigenlayerPath string)
	e2eAssert   func(t *testing.T)
)

type e2eTestCase struct {
	t        *testing.T
	testDir  string
	repoPath string
	arranger e2eArranger
	act      e2eAct
	assert   e2eAssert
}

func newE2ETestCase(t *testing.T, arranger e2eArranger, act e2eAct, assert e2eAssert) *e2eTestCase {
	t.Helper()
	tc := &e2eTestCase{
		t:        t,
		testDir:  t.TempDir(),
		repoPath: repoPath(t),
		arranger: arranger,
		act:      act,
		assert:   assert,
	}
	t.Logf("Creating new E2E test case (%p). TestDir: %s", tc, tc.testDir)
	checkGoInstalled(t)
	tc.installGoModules()
	tc.buildEgn()
	return tc
}

func (e *e2eTestCase) run() {
	// Cleanup environment before and after test
	e.Cleanup()
	defer e.Cleanup()
	if e.arranger != nil {
		err := e.arranger(e.t, e.EigenlayerPath())
		if err != nil {
			e.t.Fatalf("error in Arrange step: %v", err)
		}
	}
	if e.act != nil {
		e.act(e.t, e.EigenlayerPath())
	}
	if e.assert != nil {
		e.assert(e.t)
	}
}

func (e *e2eTestCase) EigenlayerPath() string {
	return filepath.Join(e.testDir, "eigenlayer")
}

func (e *e2eTestCase) Cleanup() {
	// Stop and remove monitoring stack if installed
	dataDir, err := dataDirPath()
	if err != nil {
		e.t.Log(err)
	}
	err = exec.Command("docker", "compose", "-f", filepath.Join(dataDir, "monitoring", "docker-compose.yml"), "down").Run()
	if err != nil {
		e.t.Logf("error removing monitoring stack. It is possible that the monitoring stack wasn't installed and this is intentional: %v", err)
	}

	// Remove all installed nodes
	nodesDir, err := os.Open(filepath.Join(dataDir, "nodes"))
	if err != nil {
		if !os.IsNotExist(err) {
			e.t.Log(err)
		}
	} else {
		dirEntries, err := nodesDir.ReadDir(-1)
		if err != nil {
			e.t.Log(err)
		}
		for _, entry := range dirEntries {
			if entry.IsDir() {
				e.t.Logf("Removing node %s", entry.Name())
				err := exec.Command("docker", "compose", "-f", filepath.Join(dataDir, "nodes", entry.Name(), "docker-compose.yml"), "down").Run()
				if err != nil {
					e.t.Logf("error removing node %s: %v", entry.Name(), err)
				}
			}
		}
	}
	err = os.RemoveAll(dataDir)
	if err != nil {
		e.t.Logf("error removing data dir: %v", err)
	}
}

func (e *e2eTestCase) installGoModules() {
	e.t.Helper()
	cmd := exec.Command("go", "mod", "download")
	cmd.Dir = e.repoPath
	e.t.Logf("Installing Go modules in %s", e.repoPath)
	if err := cmd.Run(); err != nil {
		e.t.Fatalf("error installing Go modules: %v", err)
	} else {
		e.t.Logf("Go modules installed")
	}
}

func (e *e2eTestCase) buildEgn() {
	e.t.Helper()
	outPath := filepath.Join(e.testDir, "eigenlayer")
	e.t.Logf("Building eigenlayer to %s", outPath)
	err := exec.Command("go", "build", "-o", outPath, filepath.Join(e.repoPath, "cmd", "eigenlayer", "main.go")).Run()
	if err != nil {
		e.t.Fatalf("error building eigenlayer: %v", err)
	} else {
		e.t.Logf("eigenlayer built")
	}
}

func checkGoInstalled(t *testing.T) {
	t.Helper()
	err := exec.Command("go", "version").Run()
	if err != nil {
		t.Fatalf("error checking Go installation: %v", err)
	} else {
		t.Logf("Go installed")
	}
}
