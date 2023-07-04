package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type e2eTestCase struct {
	t        *testing.T
	testDir  string
	repoPath string
}

func NewE2ETestCase(t *testing.T, repoPath string) *e2eTestCase {
	t.Helper()
	tc := &e2eTestCase{
		t:        t,
		testDir:  t.TempDir(),
		repoPath: repoPath,
	}
	t.Logf("Creating new E2E test case (%p). TestDir: %s", tc, tc.testDir)
	checkGoInstalled(t)
	tc.installGoModules()
	tc.buildEgn()
	return tc
}

func (e *e2eTestCase) EgnPath() string {
	return filepath.Join(e.testDir, "egn")
}

func (e *e2eTestCase) Cleanup() {
	// Stop and remove monitoring stack
	dataDir := dataDirPath(e.t)
	err := exec.Command("docker", "compose", "-f", filepath.Join(dataDir, "monitoring", "docker-compose.yml"), "down").Run()
	if err != nil {
		e.t.Fatalf("error removing monitoring stack: %v", err)
	}
	// Remove all installed nodes
	nodesDir, err := os.Open(filepath.Join(dataDir, "nodes"))
	if err != nil {
		if !os.IsNotExist(err) {
			e.t.Fatal(err)
		}
	} else {
		dirEntries, err := nodesDir.ReadDir(-1)
		if err != nil {
			e.t.Fatal(err)
		}
		for _, entry := range dirEntries {
			if entry.IsDir() {
				e.t.Logf("Removing node %s", entry.Name())
				err := exec.Command("docker", "compose", "-f", filepath.Join(dataDir, "nodes", entry.Name(), "docker-compose.yml"), "down").Run()
				if err != nil {
					e.t.Fatalf("error removing node %s: %v", entry.Name(), err)
				}
			}
		}
	}
	err = os.RemoveAll(dataDir)
	if err != nil {
		e.t.Fatalf("error removing data dir: %v", err)
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
	outPath := filepath.Join(e.testDir, "egn")
	e.t.Logf("Building egn to %s", outPath)
	err := exec.Command("go", "build", "-o", outPath, filepath.Join(e.repoPath, "cmd", "egn", "main.go")).Run()
	if err != nil {
		e.t.Fatalf("error building egn: %v", err)
	} else {
		e.t.Logf("egn built")
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
