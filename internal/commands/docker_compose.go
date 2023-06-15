package commands

import (
	"bytes"
	"io"
)

type DockerComposeRunnerOptions struct {
	Out io.Writer
}

type DockerComposeRunner struct {
	cmdRunner CMDRunner
}

func NewDockerComposeRunner() DockerComposeRunner {
	return DockerComposeRunner{cmdRunner: NewCMDRunner()}
}

func (d *DockerComposeRunner) Up(composePath string, runnerOpts DockerComposeRunnerOptions) error {
	cmd := "docker compose"
	if composePath != "" {
		cmd = cmd + " -f " + composePath
	}
	out, _, err := d.cmdRunner.RunCMD(Command{
		Cmd:       cmd + " up -d",
		GetOutput: false,
	})
	if err == nil && runnerOpts.Out != nil {
		_, err = bytes.NewBufferString(out).WriteTo(runnerOpts.Out)
	}
	return err
}
