package test

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
)

func TestToolsContainer(t *testing.T) {
	image := "nixos/nix"
	projectRoot, _ := filepath.Abs("../")

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
			"--command", "exit",
		},
	}

	docker.Run(t, image, options)
}
