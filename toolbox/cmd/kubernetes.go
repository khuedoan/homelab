package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func ingressHost(cmd *cobra.Command, namespace, name string) (string, error) {
	output, err := commandOutput(cmd, "kubectl", "get", "ingress", "--namespace", namespace, name, "--output=jsonpath={.spec.rules[0].host}")
	if err != nil {
		return "", err
	}
	host := strings.TrimSpace(string(output))
	if host == "" {
		return "", fmt.Errorf("ingress %s/%s has no host", namespace, name)
	}
	return host, nil
}

func applyObject(cmd *cobra.Command, object any) error {
	data, err := json.Marshal(object)
	if err != nil {
		return err
	}
	return runCommand(cmd, bytes.NewReader(data), "kubectl", "apply", "--filename=-")
}
