package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newWireguardConfigCmd() *cobra.Command {
	return &cobra.Command{Use: "config PEER", Short: "Show a WireGuard peer QR code and configuration", Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			for _, step := range [][]string{{"/app/show-peer", args[0]}, {"cat", fmt.Sprintf("/config/peer_%s/peer_%s.conf", args[0], args[0])}} {
				command := append([]string{"-n", "wireguard", "exec", "-it", "deployment/wireguard", "--"}, step...)
				if err := commandRunner(cmd).Run(cmd.Context(), nil, "kubectl", command...); err != nil {
					return err
				}
			}
			return nil
		}}
}
