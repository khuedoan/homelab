package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newBackupCmd() *cobra.Command {
	backup := &cobra.Command{Use: "backup", Short: "Set up or restore PVC backups"}
	for _, action := range []string{"setup", "restore"} {
		var namespace, pvc string
		cmd := &cobra.Command{Use: action, Short: fmt.Sprintf("%s a PVC backup", action), Args: cobra.NoArgs,
			RunE: func(cmd *cobra.Command, _ []string) error {
				if namespace == "" || pvc == "" {
					return fmt.Errorf("namespace and pvc must not be empty")
				}
				secret := pvc + "-backup-repository"
				metadata := func(name string) map[string]any {
					return map[string]any{"name": name, "namespace": namespace, "annotations": map[string]string{"app.kubernetes.io/managed-by": "toolbox"}}
				}
				data := []map[string]any{}
				for _, field := range []struct{ property, key string }{{"restic-s3-bucket", "restic_s3_bucket"}, {"restic-s3-access-key", "restic_s3_access_key"}, {"restic-s3-secret-key", "restic_s3_secret_key"}, {"restic-password", "restic_password"}} {
					data = append(data, map[string]any{"remoteRef": map[string]string{"key": "external", "property": field.property}, "secretKey": field.key})
				}
				if err := applyObject(cmd, map[string]any{"apiVersion": "external-secrets.io/v1beta1", "kind": "ExternalSecret", "metadata": metadata(secret), "spec": map[string]any{
					"secretStoreRef": map[string]string{"kind": "ClusterSecretStore", "name": "global-secrets"}, "data": data,
					"target": map[string]any{"template": map[string]any{"data": map[string]string{"RESTIC_REPOSITORY": "s3:{{ .restic_s3_bucket }}/" + namespace + "/" + pvc, "RESTIC_PASSWORD": "{{ .restic_password }}", "AWS_ACCESS_KEY_ID": "{{ .restic_s3_access_key }}", "AWS_SECRET_ACCESS_KEY": "{{ .restic_s3_secret_key }}"}}},
				}}); err != nil {
					return err
				}
				kind := "ReplicationSource"
				spec := map[string]any{"sourcePVC": pvc, "trigger": map[string]string{"schedule": "*/30 * * * *"}, "restic": map[string]any{"pruneIntervalDays": 14, "repository": secret, "retain": map[string]int{"hourly": 6, "daily": 5, "weekly": 4, "monthly": 2, "yearly": 1}, "copyMethod": "Snapshot"}}
				if action == "restore" {
					kind = "ReplicationDestination"
					spec = map[string]any{"trigger": map[string]string{"manual": "restore-once"}, "restic": map[string]string{"repository": secret, "destinationPVC": pvc, "copyMethod": "Direct"}}
				}
				return applyObject(cmd, map[string]any{"apiVersion": "volsync.backube/v1alpha1", "kind": kind, "metadata": metadata(pvc), "spec": spec})
			}}
		cmd.Flags().StringVarP(&namespace, "namespace", "n", "", "PVC namespace")
		cmd.Flags().StringVar(&pvc, "pvc", "", "PVC name")
		_ = cmd.MarkFlagRequired("namespace")
		_ = cmd.MarkFlagRequired("pvc")
		backup.AddCommand(cmd)
	}
	return backup
}
