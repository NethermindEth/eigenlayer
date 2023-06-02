package commands

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Command represents a command to be executed.
type Command struct {
	// Cmd is the command string to be executed.
	Cmd string
	// GetOutput indicates whether the output of the command should be returned.
	GetOutput bool
}

// CMDRunner is a command runner that can run commands with or without sudo.
type CMDRunner struct {
	runWithSudo bool
}

// NewCMDRunner creates a new command runner with the given options.
func NewCMDRunner() CMDRunner {
	return CMDRunner{}
}

// NewCMDRunnerWithSudo creates a new command runner that runs commands with sudo.
func NewCMDRunnerWithSudo() CMDRunner {
	return CMDRunner{runWithSudo: true}
}

// RunCMD runs a command. If the command runner is configured to run with sudo and the command is not forced to run without sudo, the command is run with sudo.
func (cr *CMDRunner) RunCMD(cmd Command) (out string, exitCode int, err error) {
	if cr.runWithSudo {
		log.Debug(`Running command with sudo.`)
		cmd.Cmd = fmt.Sprintf("sudo %s", cmd.Cmd)
	} else {
		log.Debug(`Running command without sudo.`)
	}
	return runCmd(cmd.Cmd, cmd.GetOutput)
}

// TODO: Refactor to be able to opt for show output to stdout/stderr, and by default show output to stdout/stderr and return output
// runCmd executes a command and returns the output, exit code, and any error that occurred during execution. If getOutput is true, the output of the command is returned.
func runCmd(cmd string, getOutput bool) (out string, exitCode int, err error) {
	r := strings.ReplaceAll(cmd, "\n", "")
	spl := strings.Split(r, " ")
	c, args := spl[0], spl[1:]

	exc := exec.Command(c, args...)

	var combinedOut bytes.Buffer
	if getOutput {
		// If the cmd is to get the output, then use an unified buffer to combine stdout and stderr
		exc.Stdout = &combinedOut
		exc.Stderr = &combinedOut
	} else {
		// Pipe output to stdout and stderr
		exc.Stdout = os.Stdout
		exc.Stderr = os.Stderr
	}

	// Start and wait for the command to finish
	if err = exc.Start(); err != nil {
		return
	}
	// Return this error at the end as we need to check if the output from stderr is to be returned
	err = exc.Wait()
	exitCode = exc.ProcessState.ExitCode()

	if getOutput {
		out = combinedOut.String()
	}

	return
}
