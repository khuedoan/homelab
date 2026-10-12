package secrets

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path"
	"slices"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation"
	appsclient "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclient "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/util/retry"
)

const (
	managedSecretLabel   = "homelab.khuedoan.com/bao-secret"
	referencesAnnotation = "homelab.khuedoan.com/bao-references"
	revisionAnnotation   = "homelab.khuedoan.com/bao-secrets-revision"
)

// The webhook replaces bao references with plaintext. Keep the original bindings
// here for already-deployed Secrets, which have no reference annotation. New
// Secrets can instead retain a JSON key-to-bao-reference map in their metadata.
// Namespace/name is intentional: never infer a binding from a secret's value.
var legacyReferences = map[string]map[string]string{
	"actualbudget/actualbudget-sso":       {"ACTUAL_OPENID_CLIENT_SECRET": "bao:secret/data/actualbudget/sso#client_secret"},
	"argocd/argocd-sso":                   {"client_secret": "bao:secret/data/argocd/sso#client_secret"},
	"cert-manager/cloudflare-api-token":   {"api-token": "bao:secret/data/cert-manager/cloudflare#api_token"},
	"cloudflared/cloudflared-credentials": {"credentials.json": "bao:secret/data/cloudflared/cloudflare#credentials.json"},
	"external-dns/cloudflare-api-token":   {"api-token": "bao:secret/data/external-dns/cloudflare#api_token", "zone-id": "bao:secret/data/external-dns/cloudflare#zone_id"},
	"forgejo/forgejo-admin":               {"password": "bao:secret/data/forgejo/admin#password"},
	"forgejo/forgejo-sso":                 {"secret": "bao:secret/data/forgejo/sso#client_secret"},
	"grafana/grafana-secrets":             {"GRAFANA_SSO_CLIENT_SECRET": "bao:secret/data/grafana/sso#client_secret"},
	"woodpecker/woodpecker-secret":        {"WOODPECKER_GITEA_CLIENT": "bao:secret/data/woodpecker/forgejo#client_id", "WOODPECKER_GITEA_SECRET": "bao:secret/data/woodpecker/forgejo#client_secret"},
	"woodpecker/woodpecker-agent-secret":  {"WOODPECKER_AGENT_SECRET": "bao:secret/data/woodpecker/agent#secret"},
	"zot/registry-admin-secret":           {"password": "bao:secret/data/zot/admin#password"},
}

type secretReference struct {
	path  string
	field string
}

func secretReferences(secret *corev1.Secret) (map[string]secretReference, error) {
	raw := legacyReferences[secret.Namespace+"/"+secret.Name]
	if annotation, ok := secret.Annotations[referencesAnnotation]; ok {
		// Do not overlay legacy bindings: the annotation is the entire mapping.
		raw = nil
		if err := json.Unmarshal([]byte(annotation), &raw); err != nil || len(raw) == 0 {
			return nil, fmt.Errorf("secret %s/%s has invalid Bao reference metadata", secret.Namespace, secret.Name)
		}
	}
	if len(raw) == 0 {
		return nil, fmt.Errorf("managed Secret %s/%s needs %s metadata", secret.Namespace, secret.Name, referencesAnnotation)
	}
	refs := make(map[string]secretReference, len(raw))
	for key, value := range raw {
		record, field, ok := strings.Cut(strings.TrimPrefix(value, "bao:secret/data/"), "#")
		if !strings.HasPrefix(value, "bao:secret/data/") || !ok || field == "" || strings.Contains(field, "#") ||
			len(validation.IsConfigMapKey(key)) != 0 || !strings.HasPrefix(record, secret.Namespace+"/") ||
			path.Clean(record) != record || strings.ContainsAny(record, "\\?#") || strings.HasSuffix(record, "/") {
			return nil, fmt.Errorf("secret %s/%s has unsupported Bao references", secret.Namespace, secret.Name)
		}
		refs[key] = secretReference{path: record, field: field}
	}
	return refs, nil
}

func reconcileSecrets(ctx context.Context, core coreclient.CoreV1Interface, apps appsclient.AppsV1Interface, records map[string]map[string]string) error {
	namespaces := make(map[string]bool)
	for record := range records {
		namespace, _, ok := strings.Cut(record, "/")
		if !ok || len(validation.IsDNS1123Label(namespace)) != 0 {
			return fmt.Errorf("secret record must have a namespace/path destination")
		}
		namespaces[namespace] = true
	}
	for _, namespace := range sortedKeys(namespaces) {
		if err := reconcileNamespace(ctx, core, apps, namespace, records); err != nil {
			return err
		}
	}
	return nil
}

func reconcileNamespace(ctx context.Context, core coreclient.CoreV1Interface, apps appsclient.AppsV1Interface, namespace string, records map[string]map[string]string) error {
	secrets, err := core.Secrets(namespace).List(ctx, metav1.ListOptions{LabelSelector: managedSecretLabel + "=true"})
	if err != nil {
		return fmt.Errorf("list managed Secrets in %s failed", namespace)
	}
	versions := make(map[string]string)
	affected := make(map[string]bool)
	for i := range secrets.Items {
		secret, selected, err := reconcileSecret(ctx, core, &secrets.Items[i], records)
		if err != nil {
			return err
		}
		versions[secret.Name] = string(secret.UID) + "/" + secret.ResourceVersion
		affected[secret.Name] = selected
	}
	if !slices.ContainsFunc(sortedKeys(affected), func(name string) bool { return affected[name] }) {
		return nil
	}
	return reconcileConsumers(ctx, apps, namespace, versions, affected)
}

func reconcileSecret(ctx context.Context, core coreclient.CoreV1Interface, secret *corev1.Secret, records map[string]map[string]string) (*corev1.Secret, bool, error) {
	var selected bool
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		// Always re-read on a conflict, preserving unrelated fields and metadata.
		current, err := core.Secrets(secret.Namespace).Get(ctx, secret.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if current.Labels[managedSecretLabel] != "true" {
			return fmt.Errorf("secret is no longer managed")
		}
		var changed bool
		selected, changed, err = applyRecordData(current, records)
		if err != nil {
			return err
		}
		if changed {
			current, err = core.Secrets(current.Namespace).Update(ctx, current, metav1.UpdateOptions{})
			if err != nil {
				return err
			}
		}
		secret = current
		return nil
	})
	if err != nil {
		// Kubernetes/webhook validation errors can contain the submitted values.
		// Never include their response bodies in the error returned to the CLI.
		return nil, false, fmt.Errorf("reconcile managed Secret %s/%s failed; check reference metadata, fields, mutability, and Kubernetes access", secret.Namespace, secret.Name)
	}
	return secret, selected, nil
}

func applyRecordData(current *corev1.Secret, records map[string]map[string]string) (selected, changed bool, err error) {
	refs, err := secretReferences(current)
	if err != nil {
		return false, false, err
	}
	for key, ref := range refs {
		fields, ok := records[ref.path]
		if !ok {
			continue
		}
		selected = true
		value, ok := fields[ref.field]
		if !ok {
			return false, false, fmt.Errorf("required field is absent from synchronized record")
		}
		if existing, ok := current.Data[key]; ok && bytes.Equal(existing, []byte(value)) {
			continue
		}
		if current.Data == nil {
			current.Data = make(map[string][]byte)
		}
		current.Data[key] = []byte(value)
		changed = true
	}
	return selected, changed, nil
}

// Restart only long-running controllers that actually consume a selected Secret.
// Hash non-secret API revisions, not credential values. Including all managed
// dependencies keeps the marker stable when infra, SSO, and Forgejo sync separately.
func consumerRevision(template *corev1.PodTemplateSpec, versions map[string]string, affected map[string]bool) string {
	refs := podSecretNames(&template.Spec)
	selected := false
	inputs := make(map[string]string)
	for name := range refs {
		if affected[name] {
			selected = true
		}
		if version, ok := versions[name]; ok {
			inputs[name] = version
		}
	}
	if !selected {
		return ""
	}
	data, _ := json.Marshal(inputs) // String maps have a deterministic JSON encoding.
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

func podSecretNames(spec *corev1.PodSpec) map[string]bool {
	refs := make(map[string]bool)
	for _, container := range append(slices.Clone(spec.Containers), spec.InitContainers...) {
		containerSecretNames(container, refs)
	}
	for _, volume := range spec.Volumes {
		if volume.Secret != nil {
			refs[volume.Secret.SecretName] = true
		}
		if volume.Projected != nil {
			for _, source := range volume.Projected.Sources {
				if source.Secret != nil {
					refs[source.Secret.Name] = true
				}
			}
		}
	}
	return refs
}

func containerSecretNames(container corev1.Container, refs map[string]bool) {
	for _, env := range container.Env {
		if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
			refs[env.ValueFrom.SecretKeyRef.Name] = true
		}
	}
	for _, env := range container.EnvFrom {
		if env.SecretRef != nil {
			refs[env.SecretRef.Name] = true
		}
	}
}

func reconcileConsumers(ctx context.Context, client appsclient.AppsV1Interface, namespace string, versions map[string]string, affected map[string]bool) error {
	if err := reconcileDeployments(ctx, client, namespace, versions, affected); err != nil {
		return err
	}
	if err := reconcileStatefulSets(ctx, client, namespace, versions, affected); err != nil {
		return err
	}
	return reconcileDaemonSets(ctx, client, namespace, versions, affected)
}

func reconcileDeployments(ctx context.Context, client appsclient.AppsV1Interface, namespace string, versions map[string]string, affected map[string]bool) error {
	deployments, err := client.Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list Deployments in %s failed", namespace)
	}
	for _, deployment := range deployments.Items {
		revision := consumerRevision(&deployment.Spec.Template, versions, affected)
		if revision == "" || deployment.Spec.Template.Annotations[revisionAnnotation] == revision {
			continue
		}
		if deployment.Spec.Paused {
			return fmt.Errorf("deployment %s/%s is paused; resume it before retrying secret sync", namespace, deployment.Name)
		}
		if _, err := client.Deployments(namespace).Patch(ctx, deployment.Name, types.MergePatchType, revisionPatch(deployment.ResourceVersion, revision), metav1.PatchOptions{}); err != nil {
			return fmt.Errorf("reload Deployment %s/%s failed; retry secret sync", namespace, deployment.Name)
		}
	}
	return nil
}

func reconcileStatefulSets(ctx context.Context, client appsclient.AppsV1Interface, namespace string, versions map[string]string, affected map[string]bool) error {
	statefulSets, err := client.StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list StatefulSets in %s failed", namespace)
	}
	for _, statefulSet := range statefulSets.Items {
		revision := consumerRevision(&statefulSet.Spec.Template, versions, affected)
		if revision == "" || statefulSet.Spec.Template.Annotations[revisionAnnotation] == revision {
			continue
		}
		if statefulSet.Spec.UpdateStrategy.Type == appsv1.OnDeleteStatefulSetStrategyType ||
			(statefulSet.Spec.UpdateStrategy.RollingUpdate != nil && statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition != nil && *statefulSet.Spec.UpdateStrategy.RollingUpdate.Partition > 0) {
			return fmt.Errorf("statefulset %s/%s does not automatically roll all replicas; update its rollout strategy before retrying secret sync", namespace, statefulSet.Name)
		}
		if _, err := client.StatefulSets(namespace).Patch(ctx, statefulSet.Name, types.MergePatchType, revisionPatch(statefulSet.ResourceVersion, revision), metav1.PatchOptions{}); err != nil {
			return fmt.Errorf("reload StatefulSet %s/%s failed; retry secret sync", namespace, statefulSet.Name)
		}
	}
	return nil
}

func reconcileDaemonSets(ctx context.Context, client appsclient.AppsV1Interface, namespace string, versions map[string]string, affected map[string]bool) error {
	daemonSets, err := client.DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("list DaemonSets in %s failed", namespace)
	}
	for _, daemonSet := range daemonSets.Items {
		revision := consumerRevision(&daemonSet.Spec.Template, versions, affected)
		if revision == "" || daemonSet.Spec.Template.Annotations[revisionAnnotation] == revision {
			continue
		}
		if daemonSet.Spec.UpdateStrategy.Type == appsv1.OnDeleteDaemonSetStrategyType {
			return fmt.Errorf("daemonset %s/%s uses OnDelete; update its rollout strategy before retrying secret sync", namespace, daemonSet.Name)
		}
		if _, err := client.DaemonSets(namespace).Patch(ctx, daemonSet.Name, types.MergePatchType, revisionPatch(daemonSet.ResourceVersion, revision), metav1.PatchOptions{}); err != nil {
			return fmt.Errorf("reload DaemonSet %s/%s failed; retry secret sync", namespace, daemonSet.Name)
		}
	}
	return nil
}

func revisionPatch(resourceVersion, revision string) []byte {
	patch, _ := json.Marshal(map[string]any{
		"metadata": map[string]string{"resourceVersion": resourceVersion},
		"spec":     map[string]any{"template": map[string]any{"metadata": map[string]any{"annotations": map[string]string{revisionAnnotation: revision}}}},
	})
	return patch
}

func sortedKeys[V any](values map[string]V) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	return keys
}
