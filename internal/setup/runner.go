package setup

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
)

type CommandSpec struct {
	Name        string
	Args        []string
	Stdin       string
	Interactive bool
}

type Result struct {
	Stdout string
	Stderr string
}

type Runner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, spec CommandSpec) (Result, error)
}

type execRunner struct {
	stdout io.Writer
	stderr io.Writer
}

func NewExecRunner(stdout, stderr io.Writer) Runner {
	return &execRunner{
		stdout: stdout,
		stderr: stderr,
	}
}

func (r *execRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (r *execRunner) Run(ctx context.Context, spec CommandSpec) (Result, error) {
	cmd := exec.CommandContext(ctx, spec.Name, spec.Args...)

	if spec.Interactive {
		cmd.Stdout = r.stdout
		cmd.Stderr = r.stderr
		if spec.Stdin != "" {
			cmd.Stdin = bytes.NewBufferString(spec.Stdin)
		} else {
			cmd.Stdin = os.Stdin
		}

		return Result{}, cmd.Run()
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if spec.Stdin != "" {
		cmd.Stdin = bytes.NewBufferString(spec.Stdin)
	}

	err := cmd.Run()

	return Result{
		Stdout: stdout.String(),
		Stderr: stderr.String(),
	}, err
}
