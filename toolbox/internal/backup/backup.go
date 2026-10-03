package backup

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/khuedoan/homelab/toolbox/internal/process"
)

type PVC struct {
	Namespace string
	Name      string
}

func Setup(ctx context.Context, run process.Runner, pvc PVC) error {
	spec := map[string]any{
		"sourcePVC": pvc.Name,
		"trigger":   map[string]string{"schedule": "*/30 * * * *"},
		"restic": map[string]any{
			"pruneIntervalDays": 14,
			"repository":        pvc.Name + "-backup-repository",
			"retain":            map[string]int{"hourly": 6, "daily": 5, "weekly": 4, "monthly": 2, "yearly": 1},
			"copyMethod":        "Snapshot",
		},
	}
	return apply(ctx, run, pvc, "ReplicationSource", spec)
}

func Restore(ctx context.Context, run process.Runner, pvc PVC) error {
	spec := map[string]any{
		"trigger": map[string]string{"manual": "restore-once"},
		"restic": map[string]string{
			"repository":     pvc.Name + "-backup-repository",
			"destinationPVC": pvc.Name,
			"copyMethod":     "Direct",
		},
	}
	return apply(ctx, run, pvc, "ReplicationDestination", spec)
}

func apply(ctx context.Context, run process.Runner, pvc PVC, kind string, spec map[string]any) error {
	repository := pvc.Name + "-backup-repository"
	metadata := func(name string) map[string]any {
		return map[string]any{
			"name":        name,
			"namespace":   pvc.Namespace,
			"annotations": map[string]string{"app.kubernetes.io/managed-by": "toolbox"},
		}
	}
	fields := []struct{ property, key string }{
		{"restic-s3-bucket", "restic_s3_bucket"},
		{"restic-s3-access-key", "restic_s3_access_key"},
		{"restic-s3-secret-key", "restic_s3_secret_key"},
		{"restic-password", "restic_password"},
	}
	data := make([]map[string]any, 0, len(fields))
	for _, field := range fields {
		data = append(data, map[string]any{
			"remoteRef": map[string]string{"key": "external", "property": field.property},
			"secretKey": field.key,
		})
	}
	credentials := map[string]string{
		"RESTIC_REPOSITORY":     "s3:{{ .restic_s3_bucket }}/" + pvc.Namespace + "/" + pvc.Name,
		"RESTIC_PASSWORD":       "{{ .restic_password }}",
		"AWS_ACCESS_KEY_ID":     "{{ .restic_s3_access_key }}",
		"AWS_SECRET_ACCESS_KEY": "{{ .restic_s3_secret_key }}",
	}
	secret := map[string]any{
		"apiVersion": "external-secrets.io/v1beta1",
		"kind":       "ExternalSecret",
		"metadata":   metadata(repository),
		"spec": map[string]any{
			"secretStoreRef": map[string]string{"kind": "ClusterSecretStore", "name": "global-secrets"},
			"data":           data,
			"target": map[string]any{
				"template": map[string]any{"data": credentials},
			},
		},
	}
	if err := applyObject(ctx, run, secret); err != nil {
		return err
	}
	return applyObject(ctx, run, map[string]any{
		"apiVersion": "volsync.backube/v1alpha1",
		"kind":       kind,
		"metadata":   metadata(pvc.Name),
		"spec":       spec,
	})
}

func applyObject(ctx context.Context, run process.Runner, object any) error {
	data, err := json.Marshal(object)
	if err != nil {
		return err
	}
	return run.Run(ctx, bytes.NewReader(data), "kubectl", "apply", "--filename=-")
}
