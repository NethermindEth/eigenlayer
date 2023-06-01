package commands

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Command represents a command to be executed.
type Command struct {
	// Cmd is the command string to be executed.
	Cmd string
	// GetOutput indicates whether the output of the command should be returned.
	GetOutput bool
	// ForceNoSudo forces the command to be run without sudo.
	ForceNoSudo bool
	// IgnoreTerminal indicates whether the command can be executed without using a terminal. This is useful for Windows.
	IgnoreTerminal bool
}

// ScriptFile represents a bash script to be executed.
type ScriptFile struct {
	// Tmp is the script template.
	Tmp *template.Template
	// GetOutput indicates whether the output of the script should be returned.
	GetOutput bool
	// Data is the data object for the template.
	Data interface{}
}

// CMDRunnerOptions provides options for configuring a command runner.
type CMDRunnerOptions struct {
	// RunAsAdmin indicates whether commands should be run as admin.
	RunAsAdmin bool
}

// CMDRunner is a command runner that can run commands with or without sudo.
type CMDRunner struct {
	RunWithSudo bool
}

// NewCMDRunner creates a new command runner with the given options.
func NewCMDRunner(options CMDRunnerOptions) CMDRunner {
	return CMDRunner{
		RunWithSudo: options.RunAsAdmin,
	}
}

// RunCMD runs a command. If the command runner is configured to run with sudo and the command is not forced to run without sudo, the command is run with sudo.
func (cr *CMDRunner) RunCMD(cmd Command) (out string, exitCode int, err error) {
	if cr.RunWithSudo && !cmd.ForceNoSudo {
		log.Debug(`Running command with sudo.`)
		cmd.Cmd = fmt.Sprintf("sudo %s", cmd.Cmd)
	} else {
		log.Debug(`Running command without sudo.`)
	}
	return runCmd(cmd.Cmd, cmd.GetOutput)
}

// RunScript executes a bash script.
func (cr *CMDRunner) RunScript(script ScriptFile) (string, error) {
	return executeBashScript(script, cr.RunWithSudo)
}

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

// executeBashScript executes a bash script defined in a given template. If runWithSudo is true, the script is run with sudo. The function returns the output of the script and any error that occurred during execution.
func executeBashScript(script ScriptFile, runWithSudo bool) (out string, err error) {
	var scriptBuffer, combinedOut bytes.Buffer
	if err = script.Tmp.Execute(&scriptBuffer, script.Data); err != nil {
		return
	}

	var cmd *exec.Cmd
	if runWithSudo {
		cmd = exec.Command("sudo", "bash")
	} else {
		cmd = exec.Command("bash")
	}

	// Prepare pipes for stdin, stdout and stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return
	}

	wait := sync.WaitGroup{}

	// Prepare channel to receive errors from goroutines
	errChans := make([]<-chan error, 0)
	// cmd executes any instructions coming from stdin
	errChans = append(errChans, goCopy(&wait, stdin, &scriptBuffer, true))

	if script.GetOutput {
		// If the script is to get the output, then use an unified buffer to combine stdout and stderr
		cmd.Stdout = &combinedOut
		cmd.Stderr = &combinedOut
	} else {
		// If the script is not to get the output, then pipe the output to stdout and stderr
		errChans = append(errChans, goCopy(&wait, os.Stdout, stdout, false))
		errChans = append(errChans, goCopy(&wait, os.Stderr, stderr, false))
	}

	if err = cmd.Start(); err != nil {
		return
	}

	// Check for errors from goroutines
	for _, errChan := range errChans {
		err = <-errChan
		if err != nil {
			return
		}
	}

	wait.Wait()

	if err = cmd.Wait(); err != nil {
		return
	}

	if script.GetOutput {
		out = combinedOut.String()
	}

	return out, nil
}
