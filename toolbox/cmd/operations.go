package cmd

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{Use: "status", Short: "Show Argo CD applications and cluster ingresses", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			for _, args := range [][]string{{"get", "applicationsets", "--namespace", "argocd"}, {"get", "applications", "--namespace", "argocd"}, {"get", "ingress", "--all-namespaces"}} {
				if err := runCommand(cmd, nil, "kubectl", args...); err != nil {
					return err
				}
			}
			return nil
		}}
}

func newDNSConfigCmd() *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List ingress IP addresses and DNS names", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runCommand(cmd, nil, "kubectl", "get", "ingress", "--all-namespaces", "--no-headers", "--output", "custom-columns=ADDRESS:.status.loadBalancer.ingress[0].ip,HOST:.spec.rules[0].host")
		}}
}

func newArgoCDPasswordCmd() *cobra.Command {
	return &cobra.Command{Use: "admin-password", Short: "Show the initial Argo CD admin password", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: ArgoCD admin can do anything in the cluster, only use it for just enough initial setup or in emergencies.")
			out, err := commandOutput(cmd, "kubectl", "-n", "argocd", "get", "secret", "argocd-initial-admin-secret", "-o", "jsonpath={.data.password}")
			if err != nil {
				return err
			}
			password, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(out)))
			if err != nil {
				return fmt.Errorf("decode Argo CD admin password: %w", err)
			}
			if len(password) == 0 {
				return fmt.Errorf("Argo CD admin secret has no password")
			}
			_, err = cmd.OutOrStdout().Write(password)
			return err
		}}
}

func newKanidmPasswordCmd() *cobra.Command {
	return &cobra.Command{Use: "reset-password ACCOUNT", Short: "Recover a Kanidm account password", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: Kanidm admin can do anything in the cluster, only use it for just enough initial setup or in emergencies.")
			return runCommand(cmd, nil, "kubectl", "exec", "-it", "-n", "kanidm", "statefulset/kanidm", "--", "kanidmd", "recover-account", args[0])
		}}
}

func newOnboardUserCmd() *cobra.Command {
	return &cobra.Command{Use: "create USERNAME FULL_NAME EMAIL", Short: "Create a Kanidm user and issue a credential reset token", Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			host, err := ingressHost(cmd, "kanidm", "kanidm")
			if err != nil {
				return err
			}
			options := []string{"--url", "https://" + host, "--name", "idm_admin"}
			for _, step := range [][]string{{"person", "create", args[0], args[1]}, {"person", "update", args[0]}, {"group", "add-members", "editor", args[0]}, {"person", "credential", "create-reset-token", args[0]}} {
				command := append(step, options...)
				if step[1] == "update" {
					command = append(command, "--mail", args[2])
				}
				if err := runCommand(cmd, nil, "kanidm", command...); err != nil {
					return err
				}
			}
			return nil
		}}
}

func newWireguardConfigCmd() *cobra.Command {
	return &cobra.Command{Use: "config PEER", Short: "Show a WireGuard peer QR code and configuration", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, step := range [][]string{{"/app/show-peer", args[0]}, {"cat", fmt.Sprintf("/config/peer_%s/peer_%s.conf", args[0], args[0])}} {
				command := append([]string{"-n", "wireguard", "exec", "-it", "deployment/wireguard", "--"}, step...)
				if err := runCommand(cmd, nil, "kubectl", command...); err != nil {
					return err
				}
			}
			return nil
		}}
}

func newServiceCmd() *cobra.Command {
	return &cobra.Command{Use: "create NAME", Short: "Scaffold an application Helm chart", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`).MatchString(args[0]) {
				return fmt.Errorf("name must contain only lowercase letters, digits and internal hyphens")
			}
			if err := os.MkdirAll("apps", 0755); err != nil {
				return err
			}
			path := filepath.Join("apps", args[0])
			if err := os.Mkdir(path, 0755); err != nil {
				return fmt.Errorf("create application directory: %w", err)
			}
			complete := false
			defer func() {
				if !complete {
					_ = os.RemoveAll(path)
				}
			}()
			chart := "apiVersion: v2\nname: CHANGEME\nversion: 0.0.0\ndependencies:\n- name: CHANGEME\n  version: CHANGEME\n  repository: CHANGEME\n"
			if err := os.WriteFile(filepath.Join(path, "Chart.yaml"), []byte(chart), 0644); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(path, "values.yaml"), nil, 0644); err != nil {
				return err
			}
			complete = true
			return nil
		}}
}
