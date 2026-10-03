package process

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Runner struct {
	kubeconfig string
	stdin      io.Reader
	stdout     io.Writer
	stderr     io.Writer
}

func New(kubeconfig string, stdin io.Reader, stdout, stderr io.Writer) Runner {
	return Runner{kubeconfig: kubeconfig, stdin: stdin, stdout: stdout, stderr: stderr}
}

func (r Runner) Command(ctx context.Context, name string, args ...string) *exec.Cmd {
	command := exec.CommandContext(ctx, name, args...)
	if r.kubeconfig != "" {
		command.Env = append(os.Environ(), "KUBECONFIG="+r.kubeconfig)
	}
	return command
}

func (r Runner) Run(ctx context.Context, input io.Reader, name string, args ...string) error {
	command := r.Command(ctx, name, args...)
	if input == nil {
		input = r.stdin
	}
	command.Stdin = input
	command.Stdout = r.stdout
	command.Stderr = r.stderr
	if err := command.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func (r Runner) Output(ctx context.Context, name string, args ...string) ([]byte, error) {
	command := r.Command(ctx, name, args...)
	command.Stderr = r.stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return output, nil
}

func (r Runner) PrivateOutput(ctx context.Context, name string, args ...string) ([]byte, error) {
	command := r.Command(ctx, name, args...)
	command.Stderr = io.Discard
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("%s failed: %w", name, err)
	}
	return output, nil
}
