package test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
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
		{"docker", "--version", ">= 20.10.17, < 21.0.0"},
		{"git", "--version", ">= 2.37.1, < 3.0.0"},
		{"go", "version", ">= 1.19.0, < 1.20.0"},
		{"helm", "version", ">= 3.9.4, < 4.0.0"},
		{"kubectl", "version", ">= 1.25.0, < 1.27.0"}, // https://kubernetes.io/releases/version-skew-policy/#kubectl
		{"kustomize", "version", ">= 4.5.4, < 5.0.0"},
		{"pre-commit", "--version", ">= 2.20.0, < 3.0.0"},
		{"terraform", "--version", ">= 1.3.1, < 1.4.0"},
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

func TestToolsContainer(t *testing.T) {
	t.Parallel()

	image := "nixos/nix"
	projectRoot, err := filepath.Abs("../")
	if err != nil {
		t.FailNow()
	}

	options := &docker.RunOptions{
		Remove: true,
		Volumes: []string{
			fmt.Sprintf("%s:%s", projectRoot, projectRoot),
			"homelab-tools-cache:/root/.cache",
			"homelab-tools-nix:/nix",
		},
		OtherOptions: []string{
			"--workdir", projectRoot,
		},
		Command: []string{
			"nix-shell",
			"--pure",
			"--command", "exit",
		},
	}

	docker.Run(t, image, options)
}

func TestToolsNixShell(t *testing.T) {
	t.Parallel()

	projectRoot, err := filepath.Abs("../")
	if err != nil {
		t.FailNow()
	}

	command := shell.Command{
		Command: "nix-shell",
		Args: []string{
			"--pure",
			"--command", "exit",
		},
		WorkingDir: projectRoot,
	}

	shell.RunCommand(t, command)
}
