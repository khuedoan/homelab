package test

import (
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/shell"
	"github.com/gruntwork-io/terratest/modules/version-checker"
)

func TestToolsVersions(t *testing.T) {
	t.Parallel()

	var tools = []struct {
		binaryPath        string
		versionArg        string
		versionConstraint string
	}{
		{"ansible", "--version", ">= 2.12.6, < 3.0.0"},
		{"docker", "--version", ">= 25.0.0, < 26.0.0"},
		{"git", "--version", ">= 2.37.1, < 3.0.0"},
		{"go", "version", ">= 1.22.0, < 1.23.0"},
		{"helm", "version", ">= 3.9.4, < 4.0.0"},
		{"kubectl", "version", ">= 1.30.0, < 1.32.0"}, // https://kubernetes.io/releases/version-skew-policy/#kubectl
		{"kustomize", "version", ">= 5.0.3, < 6.0.0"},
		{"pre-commit", "--version", ">= 3.3.2, < 4.0.0"},
		{"terraform", "--version", ">= 1.8.0, < 1.9.0"},
	}

	for _, tool := range tools {
		tool := tool // https://github.com/golang/go/wiki/CommonMistakes#using-goroutines-on-loop-iterator-variables
		t.Run(tool.binaryPath, func(t *testing.T) {
			t.Parallel()
			params := version_checker.CheckVersionParams{
				BinaryPath:        tool.binaryPath,
				VersionConstraint: tool.versionConstraint,
				VersionArg:        tool.versionArg,
				WorkingDir:        ".",
			}

			version_checker.CheckVersion(t, params)
		})
	}
}

func TestToolsNixShell(t *testing.T) {
	t.Parallel()

	projectRoot, err := filepath.Abs("../")
	if err != nil {
		t.FailNow()
	}

	command := shell.Command{
		Command: "nix",
		Args: []string{
			"develop",
			"--experimental-features", "nix-command flakes",
			"--command", "true",
		},
		WorkingDir: projectRoot,
	}

	shell.RunCommand(t, command)
}
