package test

import (
	"testing"

	"github.com/gruntwork-io/terratest/modules/docker"
)

func TestToolsContainer(t *testing.T) {
	tag := "homelab-tools"
	buildOptions := &docker.BuildOptions{
		Tags: []string{tag},
	}

	docker.Build(t, "../tools", buildOptions)
}
