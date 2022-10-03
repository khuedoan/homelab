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
		versionConstraint string
	}{
		{"ansible", ">= 2.12.6, < 3.0.0"},
		{"docker", ">= 20.10.17, < 21.0.0"},
		{"git", ">= 2.37.1, < 3.0.0"},
		// TODO add more version checks
		// {"go", ">= 2.37.1, < 3.0.0"},
		// {"helm", ">= 2.37.1, < 3.0.0"},
		// {"kubectl", ">= 2.37.1, < 3.0.0"},
		// {"kustomize", ">= 2.37.1, < 3.0.0"},
		{"pre-commit", ">= 2.20.0, < 3.0.0"},
		{"terraform", ">= 1.3.1, < 1.4.0"},
	}

	for _, tool := range tools {
		params := version_checker.CheckVersionParams{
			BinaryPath: tool.binaryPath,
			VersionConstraint: tool.versionConstraint,
			WorkingDir: ".",
		}

		version_checker.CheckVersion(t, params)
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
