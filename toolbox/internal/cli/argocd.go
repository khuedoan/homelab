package cli

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newArgoCDPasswordCmd() *cobra.Command {
	return &cobra.Command{Use: "admin-password", Short: "Show the initial Argo CD admin password", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintln(cmd.ErrOrStderr(), "WARNING: ArgoCD admin can do anything in the cluster, only use it for just enough initial setup or in emergencies.")
			out, err := commandRunner(cmd).Output(cmd.Context(), "kubectl", "-n", "argocd", "get", "secret", "argocd-initial-admin-secret", "-o", "jsonpath={.data.password}")
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
		},
	}
}
