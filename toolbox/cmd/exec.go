package cmd

import (
	"fmt"
	"io"
	"os/exec"

	"github.com/spf13/cobra"
)

func runCommand(cmd *cobra.Command, input io.Reader, name string, args ...string) error {
	process := exec.CommandContext(cmd.Context(), name, args...)
	if input == nil {
		input = cmd.InOrStdin()
	}
	process.Stdin = input
	process.Stdout = cmd.OutOrStdout()
	process.Stderr = cmd.ErrOrStderr()
	if err := process.Run(); err != nil {
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func commandOutput(cmd *cobra.Command, name string, args ...string) ([]byte, error) {
	process := exec.CommandContext(cmd.Context(), name, args...)
	process.Stderr = cmd.ErrOrStderr()
	output, err := process.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return output, nil
}
