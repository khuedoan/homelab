package cli

import (
	"github.com/spf13/cobra"
)

func newDNSConfigCmd() *cobra.Command {
	return &cobra.Command{Use: "list", Short: "List ingress IP addresses and DNS names", Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return commandRunner(cmd).Run(cmd.Context(), nil, "kubectl", "get", "ingress", "--all-namespaces", "--no-headers", "--output", "custom-columns=ADDRESS:.status.loadBalancer.ingress[0].ip,HOST:.spec.rules[0].host")
		}}
}
