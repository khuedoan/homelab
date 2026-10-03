package cli

import (
	"fmt"

	"github.com/khuedoan/homelab/toolbox/internal/backup"
	"github.com/spf13/cobra"
)

func newBackupCmd() *cobra.Command {
	group := &cobra.Command{Use: "backup", Short: "Set up or restore PVC backups"}
	for _, action := range []string{"setup", "restore"} {
		var namespace, pvc string
		cmd := &cobra.Command{Use: action, Short: fmt.Sprintf("%s a PVC backup", action), Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				if namespace == "" || pvc == "" {
					return fmt.Errorf("namespace and pvc must not be empty")
				}
				volume := backup.PVC{Namespace: namespace, Name: pvc}
				if action == "restore" {
					return backup.Restore(cmd.Context(), commandRunner(cmd), volume)
				}
				return backup.Setup(cmd.Context(), commandRunner(cmd), volume)
			},
		}
		cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "PVC namespace")
		cmd.Flags().StringVar(&pvc, "pvc", "", "PVC name")
		_ = cmd.MarkFlagRequired("namespace")
		_ = cmd.MarkFlagRequired("pvc")
		group.AddCommand(cmd)
	}
	return group
}
