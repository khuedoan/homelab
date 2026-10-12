package secrets

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	clienttesting "k8s.io/client-go/testing"
)

func managedSecret(namespace, name string, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name, ResourceVersion: "1", Labels: map[string]string{managedSecretLabel: "true"}},
		Data:       data,
	}
}

func envTemplate(names ...string) corev1.PodTemplateSpec {
	container := corev1.Container{Name: "app", Image: "example:1"}
	for _, name := range names {
		container.EnvFrom = append(container.EnvFrom, corev1.EnvFromSource{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: name}}})
	}
	return corev1.PodTemplateSpec{Spec: corev1.PodSpec{Containers: []corev1.Container{container}}}
}

// The object tracker does not assign resourceVersions. Model the API server so
// assertions cover a real Secret data update followed by a stable second sync.
func versionedClient(objects ...runtime.Object) *fake.Clientset {
	client := fake.NewClientset(objects...)
	version := 1
	client.PrependReactor("update", "secrets", func(action clienttesting.Action) (bool, runtime.Object, error) {
		secret := action.(clienttesting.UpdateAction).GetObject().(*corev1.Secret)
		version++
		secret.ResourceVersion = fmt.Sprint(version)
		return false, nil, nil
	})
	return client
}

func countWrites(client *fake.Clientset, resource string) int {
	count := 0
	for _, action := range client.Actions() {
		if action.GetResource().Resource == resource && (action.GetVerb() == "update" || action.GetVerb() == "patch" || action.GetVerb() == "delete") {
			count++
		}
	}
	return count
}

func TestReconcileRotationAndUnchangedRetry(t *testing.T) {
	ctx := context.Background()
	secret := managedSecret("woodpecker", "woodpecker-secret", map[string][]byte{"WOODPECKER_GITEA_CLIENT": []byte("old-id"), "WOODPECKER_GITEA_SECRET": []byte("old-secret"), "unrelated": []byte("preserve")})
	secret.Annotations = map[string]string{"owner": "keep"}
	agent := managedSecret("woodpecker", "woodpecker-agent-secret", map[string][]byte{"WOODPECKER_AGENT_SECRET": []byte("agent-value")})
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "woodpecker", Name: "server"}, Spec: appsv1.DeploymentSpec{Template: envTemplate(secret.Name, agent.Name)}}
	unrelated := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Namespace: "woodpecker", Name: "unrelated"}, Spec: appsv1.DeploymentSpec{Template: envTemplate("other")}}
	client := versionedClient(secret, agent, deployment, unrelated)
	fail := true
	client.PrependReactor("patch", "deployments", func(clienttesting.Action) (bool, runtime.Object, error) {
		if fail {
			fail = false
			return true, nil, errors.New("admission error containing new-secret-canary")
		}
		return false, nil, nil
	})
	records := map[string]map[string]string{"woodpecker/forgejo": {"client_id": "new-id", "client_secret": "new-secret-canary"}}
	if err := reconcileSecrets(ctx, client.CoreV1(), client.AppsV1(), records); err == nil || strings.Contains(err.Error(), "canary") {
		t.Fatalf("expected a redacted reload error, got %v", err)
	}
	updated, err := client.CoreV1().Secrets("woodpecker").Get(ctx, secret.Name, metav1.GetOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if string(updated.Data["WOODPECKER_GITEA_SECRET"]) != "new-secret-canary" || string(updated.Data["unrelated"]) != "preserve" || updated.Annotations["owner"] != "keep" {
		t.Fatal("rotation failed to preserve unrelated data or metadata")
	}
	// A prior failure must be retried even though both Bao and Secret data match.
	for range 2 {
		if err := reconcileSecrets(ctx, client.CoreV1(), client.AppsV1(), records); err != nil {
			t.Fatal(err)
		}
	}
	if countWrites(client, "secrets") != 1 || countWrites(client, "deployments") != 2 {
		t.Fatal("expected one Secret update, one failed rollout and one retry")
	}
	current, err := client.AppsV1().Deployments("woodpecker").Get(ctx, deployment.Name, metav1.GetOptions{})
	if err != nil || current.Spec.Template.Annotations[revisionAnnotation] == "" {
		t.Fatal("missing rollout marker")
	}
	other, err := client.AppsV1().Deployments("woodpecker").Get(ctx, unrelated.Name, metav1.GetOptions{})
	if err != nil || len(other.Spec.Template.Annotations) != 0 {
		t.Fatal("unrelated deployment was changed")
	}
	checkSeparateSyncs(t, client)
}

func checkSeparateSyncs(t *testing.T, client *fake.Clientset) {
	t.Helper()
	ctx := context.Background()
	// Separate commands must not oscillate the same consumer's marker between
	// unrelated single-secret checksums.
	if err := reconcileSecrets(ctx, client.CoreV1(), client.AppsV1(), map[string]map[string]string{"woodpecker/agent": {"secret": "agent-value"}}); err != nil {
		t.Fatal(err)
	}
	if countWrites(client, "secrets") != 1 || countWrites(client, "deployments") != 2 {
		t.Fatal("separate sync with unchanged credentials triggered a rollout")
	}
	if err := reconcileSecrets(ctx, client.CoreV1(), client.AppsV1(), map[string]map[string]string{"woodpecker/agent": {"secret": "rotated-agent"}}); err != nil {
		t.Fatal(err)
	}
	if countWrites(client, "secrets") != 2 || countWrites(client, "deployments") != 3 {
		t.Fatal("second managed dependency did not trigger a rollout")
	}
}

func TestReconcileSecretUpdateFailureAndConflict(t *testing.T) {
	for _, failure := range []string{"validation", "conflict"} {
		t.Run(failure, func(t *testing.T) {
			testSecretUpdateRecovery(t, failure)
		})
	}
}

func testSecretUpdateRecovery(t *testing.T, failure string) {
	t.Helper()
	secret := managedSecret("grafana", "grafana-secrets", map[string][]byte{"GRAFANA_SSO_CLIENT_SECRET": []byte("old")})
	client := versionedClient(secret)
	failSecretUpdateOnce(t, client, secret, failure)
	records := map[string]map[string]string{"grafana/sso": {"client_secret": "new"}}
	err := reconcileSecrets(context.Background(), client.CoreV1(), client.AppsV1(), records)
	if failure == "validation" {
		if err == nil || strings.Contains(err.Error(), "canary") {
			t.Fatalf("expected redacted validation error: %v", err)
		}
		if err := reconcileSecrets(context.Background(), client.CoreV1(), client.AppsV1(), records); err != nil {
			t.Fatal(err)
		}
	} else if err != nil {
		t.Fatal(err)
	}
	current, err := client.CoreV1().Secrets(secret.Namespace).Get(context.Background(), secret.Name, metav1.GetOptions{})
	if err != nil || string(current.Data["GRAFANA_SSO_CLIENT_SECRET"]) != "new" {
		t.Fatal("Secret update did not recover")
	}
	if failure == "conflict" && string(current.Data["concurrent"]) != "preserved" {
		t.Fatal("conflict retry overwrote a concurrent change")
	}
}

func failSecretUpdateOnce(t *testing.T, client *fake.Clientset, secret *corev1.Secret, failure string) {
	t.Helper()
	failed := false
	client.PrependReactor("update", "secrets", func(clienttesting.Action) (bool, runtime.Object, error) {
		if failed {
			return false, nil, nil
		}
		failed = true
		if failure == "conflict" {
			current := secret.DeepCopy()
			current.Data["concurrent"] = []byte("preserved")
			if err := client.Tracker().Update(corev1.SchemeGroupVersion.WithResource("secrets"), current, secret.Namespace); err != nil {
				t.Fatal(err)
			}
			return true, nil, apierrors.NewConflict(schema.GroupResource{Resource: "secrets"}, secret.Name, errors.New("canary"))
		}
		return true, nil, errors.New("validation includes secret-canary")
	})
}

func TestReconcileUsesExplicitBindingsAndManagedLabel(t *testing.T) {
	ctx := context.Background()
	secret := managedSecret("custom", "sso", map[string][]byte{"CLIENT_SECRET": []byte("old"), "static": []byte("keep")})
	secret.Annotations = map[string]string{referencesAnnotation: `{"CLIENT_SECRET":"bao:secret/data/custom/sso#client_secret"}`}
	unmanaged := managedSecret("custom", "unmanaged", map[string][]byte{"CLIENT_SECRET": []byte("untouched")})
	unmanaged.Labels = nil
	client := versionedClient(secret, unmanaged)
	if err := reconcileSecrets(ctx, client.CoreV1(), client.AppsV1(), map[string]map[string]string{"custom/sso": {"client_secret": "new"}}); err != nil {
		t.Fatal(err)
	}
	if countWrites(client, "secrets") != 1 {
		t.Fatal("expected exactly one labeled Secret update")
	}
	current, err := client.CoreV1().Secrets("custom").Get(ctx, secret.Name, metav1.GetOptions{})
	if err != nil || string(current.Data["CLIENT_SECRET"]) != "new" || string(current.Data["static"]) != "keep" {
		t.Fatal("annotation mapping was not applied safely")
	}
	other, err := client.CoreV1().Secrets("custom").Get(ctx, unmanaged.Name, metav1.GetOptions{})
	if err != nil || string(other.Data["CLIENT_SECRET"]) != "untouched" {
		t.Fatal("unlabeled Secret was changed")
	}
}

func TestSecretReferencesValidation(t *testing.T) {
	for _, annotation := range []string{
		`not-json-canary`, `null`, `{}`, `{"x":null}`, `{"x":"vault:secret/data/custom/a#key"}`,
		`{"x":"bao:secret/data/other/a#key"}`, `{"x":"bao:secret/data/custom/../other#key"}`,
		`{"x":"bao:secret/data/custom/a#"}`, `{"x":"bao:secret/data/custom/a#key#1"}`,
		`{"bad key":"bao:secret/data/custom/a#key"}`,
	} {
		secret := managedSecret("custom", "test", nil)
		secret.Annotations = map[string]string{referencesAnnotation: annotation}
		if _, err := secretReferences(secret); err == nil || strings.Contains(err.Error(), "canary") {
			t.Errorf("accepted invalid references or leaked contents: %v", err)
		}
	}
	for identity, expected := range legacyReferences {
		namespace, name, _ := strings.Cut(identity, "/")
		refs, err := secretReferences(managedSecret(namespace, name, nil))
		if err != nil || len(refs) != len(expected) {
			t.Errorf("legacy mapping %s is invalid", identity)
		}
	}
}

func TestReconcileRejectsMissingFieldsAndUnknownBindings(t *testing.T) {
	for _, name := range []string{"woodpecker-secret", "unknown"} {
		client := versionedClient(managedSecret("woodpecker", name, nil))
		err := reconcileSecrets(context.Background(), client.CoreV1(), client.AppsV1(), map[string]map[string]string{"woodpecker/forgejo": {"client_id": "new-id"}})
		if err == nil || countWrites(client, "secrets") != 0 {
			t.Errorf("missing mapping or field must fail without writing: %s", name)
		}
	}
}

func TestReconcileAbsentSecretsIsBootstrapSafe(t *testing.T) {
	client := versionedClient()
	if err := reconcileSecrets(context.Background(), client.CoreV1(), client.AppsV1(), map[string]map[string]string{"woodpecker/forgejo": {"client_id": "id", "client_secret": "secret"}}); err != nil {
		t.Fatal(err)
	}
	for _, action := range client.Actions() {
		if action.GetVerb() != "list" || action.GetResource().Resource != "secrets" {
			t.Fatal("bootstrap must not create or delete missing Secrets")
		}
	}
}

func TestConsumerSecretReferences(t *testing.T) {
	spec := corev1.PodSpec{
		Containers:     []corev1.Container{{Name: "app", Env: []corev1.EnvVar{{Name: "KEY", ValueFrom: &corev1.EnvVarSource{SecretKeyRef: &corev1.SecretKeySelector{LocalObjectReference: corev1.LocalObjectReference{Name: "env"}}}}}}},
		InitContainers: []corev1.Container{{Name: "init", EnvFrom: []corev1.EnvFromSource{{SecretRef: &corev1.SecretEnvSource{LocalObjectReference: corev1.LocalObjectReference{Name: "init"}}}}}},
		Volumes: []corev1.Volume{
			{Name: "secret", VolumeSource: corev1.VolumeSource{Secret: &corev1.SecretVolumeSource{SecretName: "volume"}}},
			{Name: "projected", VolumeSource: corev1.VolumeSource{Projected: &corev1.ProjectedVolumeSource{Sources: []corev1.VolumeProjection{{Secret: &corev1.SecretProjection{LocalObjectReference: corev1.LocalObjectReference{Name: "projected"}}}}}}},
		},
	}
	expected := map[string]bool{"env": true, "init": true, "volume": true, "projected": true}
	if !reflect.DeepEqual(podSecretNames(&spec), expected) {
		t.Fatal("missed a Secret consumer")
	}
}

func TestReconcileAllControllerTypes(t *testing.T) {
	ctx := context.Background()
	secret := managedSecret("woodpecker", "woodpecker-secret", map[string][]byte{"WOODPECKER_GITEA_CLIENT": []byte("id"), "WOODPECKER_GITEA_SECRET": []byte("secret")})
	meta := metav1.ObjectMeta{Namespace: "woodpecker", Name: "consumer"}
	template := envTemplate(secret.Name)
	client := versionedClient(secret,
		&appsv1.Deployment{ObjectMeta: meta, Spec: appsv1.DeploymentSpec{Template: template}},
		&appsv1.StatefulSet{ObjectMeta: meta, Spec: appsv1.StatefulSetSpec{Template: template}},
		&appsv1.DaemonSet{ObjectMeta: meta, Spec: appsv1.DaemonSetSpec{Template: template}},
	)
	records := map[string]map[string]string{"woodpecker/forgejo": {"client_id": "id", "client_secret": "secret"}}
	for range 2 {
		if err := reconcileSecrets(ctx, client.CoreV1(), client.AppsV1(), records); err != nil {
			t.Fatal(err)
		}
	}
	if countWrites(client, "secrets") != 0 {
		t.Fatal("unchanged Secret was updated")
	}
	for _, resource := range []string{"deployments", "statefulsets", "daemonsets"} {
		if countWrites(client, resource) != 1 {
			t.Errorf("expected one baseline rollout for %s", resource)
		}
	}
}

func TestReconcileReportsControllersThatCannotRoll(t *testing.T) {
	meta := metav1.ObjectMeta{Namespace: "woodpecker", Name: "consumer"}
	template := envTemplate("woodpecker-secret")
	partition := int32(1)
	for _, controller := range []runtime.Object{
		&appsv1.Deployment{ObjectMeta: meta, Spec: appsv1.DeploymentSpec{Template: template, Paused: true}},
		&appsv1.StatefulSet{ObjectMeta: meta, Spec: appsv1.StatefulSetSpec{Template: template, UpdateStrategy: appsv1.StatefulSetUpdateStrategy{Type: appsv1.OnDeleteStatefulSetStrategyType}}},
		&appsv1.StatefulSet{ObjectMeta: meta, Spec: appsv1.StatefulSetSpec{Template: template, UpdateStrategy: appsv1.StatefulSetUpdateStrategy{RollingUpdate: &appsv1.RollingUpdateStatefulSetStrategy{Partition: &partition}}}},
		&appsv1.DaemonSet{ObjectMeta: meta, Spec: appsv1.DaemonSetSpec{Template: template, UpdateStrategy: appsv1.DaemonSetUpdateStrategy{Type: appsv1.OnDeleteDaemonSetStrategyType}}},
	} {
		client := versionedClient(managedSecret("woodpecker", "woodpecker-secret", nil), controller)
		err := reconcileSecrets(context.Background(), client.CoreV1(), client.AppsV1(), map[string]map[string]string{"woodpecker/forgejo": {"client_id": "id", "client_secret": "secret"}})
		if err == nil {
			t.Errorf("accepted controller that cannot roll: %T", controller)
		}
	}
}

func TestRevisionPatchUsesOptimisticConcurrency(t *testing.T) {
	data := revisionPatch("version-123", "revision")
	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["metadata"].(map[string]any)["resourceVersion"] != "version-123" {
		t.Fatal("missing optimistic concurrency guard")
	}
}

func TestReconcileRechecksManagedLabelBeforeUpdating(t *testing.T) {
	secret := managedSecret("woodpecker", "woodpecker-secret", nil)
	client := versionedClient(secret)
	client.PrependReactor("get", "secrets", func(clienttesting.Action) (bool, runtime.Object, error) {
		current := secret.DeepCopy()
		current.Labels = nil
		return true, current, nil
	})
	err := reconcileSecrets(context.Background(), client.CoreV1(), client.AppsV1(), map[string]map[string]string{"woodpecker/forgejo": {"client_id": "id", "client_secret": "secret"}})
	if err == nil || countWrites(client, "secrets") != 0 {
		t.Fatal("changed a Secret that opted out after the initial list")
	}
}

func TestReferenceAnnotationReplacesLegacyBindings(t *testing.T) {
	secret := managedSecret("woodpecker", "woodpecker-secret", nil)
	secret.Annotations = map[string]string{referencesAnnotation: `{"CUSTOM_KEY":"bao:secret/data/woodpecker/custom#key"}`}
	refs, err := secretReferences(secret)
	if err != nil || len(refs) != 1 || refs["CUSTOM_KEY"].path != "woodpecker/custom" {
		t.Fatal("explicit annotation did not replace the legacy bindings")
	}
}

func TestRevisionUsesMetadataAndAllDependencies(t *testing.T) {
	template := envTemplate("oauth", "agent")
	versions := map[string]string{"oauth": "uid/1", "agent": "uid/2"}
	first := consumerRevision(&template, versions, map[string]bool{"oauth": true})
	second := consumerRevision(&template, versions, map[string]bool{"agent": true})
	if first == "" || first != second {
		t.Fatal("partial syncs produced different markers for identical dependencies")
	}
	versions["agent"] = "uid/3"
	if consumerRevision(&template, versions, map[string]bool{"oauth": true}) == first {
		t.Fatal("marker ignored a different dependency's version")
	}
	if consumerRevision(&template, versions, map[string]bool{"unrelated": true}) != "" {
		t.Fatal("unrelated Secret selected the workload")
	}
}
