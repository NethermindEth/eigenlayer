package commands

import (
	"fmt"
	"html/template"
	"testing"
)

func TestRunCmd(t *testing.T) {
	inputs := []struct {
		cmd       string
		getOutput bool
		output    string
		isErr     bool
	}{
		{
			cmd:       "echo hello world",
			getOutput: true,
			output:    "hello world\n",
			isErr:     false,
		},
		{
			cmd:       "wr0n6",
			getOutput: true,
			isErr:     true,
		},
	}

	runner := NewCMDRunner(CMDRunnerOptions{
		RunAsAdmin: false,
	})

	for _, input := range inputs {
		descr := fmt.Sprintf("RunCmd(%s,%t)", input.cmd, input.getOutput)

		got, _, err := runner.RunCMD(Command{
			Cmd:       input.cmd,
			GetOutput: input.getOutput,
		})
		if input.isErr && err == nil {
			t.Errorf("%s expected to fail", descr)
		} else if !input.isErr {
			if err != nil {
				t.Errorf("%s failed: %v", descr, err)
			} else if input.getOutput && input.output != got {
				t.Errorf("%s expected %s but got %s", descr, input.output, got)
			}
		}
	}
}

func ExampleCMDRunner_RunCMD() {
	cmdRunner := NewCMDRunner(CMDRunnerOptions{RunAsAdmin: false})
	out, exitCode, err := cmdRunner.RunCMD(Command{Cmd: "echo hello", GetOutput: true})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Exit code: %d\nOutput: %s", exitCode, out)
	// Output: Exit code: 0
	// Output: hello
}

func TestRunBashScript(t *testing.T) {
	inputs := []struct {
		cmd       string
		getOutput bool
		output    string
		isErr     bool
	}{
		{
			cmd:       "echo hello world",
			getOutput: true,
			output:    "hello world\n",
			isErr:     false,
		},
		{
			cmd:       "wr0n6",
			getOutput: true,
			isErr:     true,
		},
	}

	runner := NewCMDRunner(CMDRunnerOptions{
		RunAsAdmin: false,
	})

	for _, input := range inputs {
		descr := fmt.Sprintf("RunBashCmd(%s,%t)", input.cmd, input.getOutput)

		tmp, err := template.New("script").Parse(string(input.cmd))
		if err != nil {
			t.Fatalf("Unexpected error at case %q: %v", input.cmd, err)
		}

		got, err := runner.RunScript(ScriptFile{
			Tmp:       tmp,
			GetOutput: input.getOutput,
		})
		if input.isErr && err == nil {
			t.Errorf("%s expected to fail", descr)
		} else if !input.isErr {
			if err != nil {
				t.Errorf("%s failed: %v", descr, err)
			} else if input.getOutput && input.output != got {
				t.Errorf("%s expected %s but got %s", descr, input.output, got)
			}
		}
	}
}

func ExampleCMDRunner_RunScript() {
	cmdRunner := NewCMDRunner(CMDRunnerOptions{RunAsAdmin: false})
	tmp := template.Must(template.New("test").Parse("echo hello"))
	out, err := cmdRunner.RunScript(ScriptFile{Tmp: tmp, GetOutput: true})
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
	// Output: hello
}
