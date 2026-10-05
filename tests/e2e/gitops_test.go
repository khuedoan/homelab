package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
)

func checkGitOps(t *testing.T, cluster fixture.Cluster) {
	t.Helper()
	if cluster.Target.Config.GitOpsNamespace == "" {
		t.Skip("no GitOps controller configured")
	}
	namespace := fixture.Namespace(t, cluster.Client)
	applications := cluster.Dynamic.Resource(schema.GroupVersionResource{
		Group: "argoproj.io", Version: "v1alpha1", Resource: "applications",
	}).Namespace(cluster.Target.Config.GitOpsNamespace)
	const revision = "8088f4c0d970abb09e250248cc97e35623447cb5"
	app, err := applications.Create(t.Context(), &unstructured.Unstructured{Object: map[string]any{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Application",
		"metadata":   map[string]any{"name": namespace},
		"spec": map[string]any{
			"project": "default",
			"source": map[string]any{
				"repoURL":        "https://github.com/argoproj/argocd-example-apps.git",
				"targetRevision": revision,
				"path":           "helm-guestbook",
				"helm": map[string]any{"valuesObject": map[string]any{
					"fullnameOverride": "web",
					"image":            map[string]any{"repository": "docker.io/library/nginx", "tag": "1.28.0"},
				}},
			},
			"destination": map[string]any{"server": "https://kubernetes.default.svc", "namespace": namespace},
			"syncPolicy":  map[string]any{"automated": map[string]any{"prune": true, "selfHeal": true}},
		},
	}}, metav1.CreateOptions{})
	require.NoError(t, err)
	name, uid := app.GetName(), app.GetUID()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		err := applications.Delete(ctx, name, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid}})
		if err != nil && !apierrors.IsNotFound(err) {
			t.Errorf("delete GitOps application %s: %v", name, err)
			return
		}
		err = wait.PollUntilContextCancel(ctx, time.Second, true, func(ctx context.Context) (bool, error) {
			_, err := applications.Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		})
		if err != nil {
			t.Errorf("wait for GitOps application deletion: %v", err)
		}
	})
	err = wait.PollUntilContextTimeout(t.Context(), 3*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		current, err := applications.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		app = current
		sync, _, err := unstructured.NestedString(app.Object, "status", "sync", "status")
		if err != nil {
			return false, err
		}
		health, _, err := unstructured.NestedString(app.Object, "status", "health", "status")
		return sync == "Synced" && health == "Healthy", err
	})
	require.NoError(t, err, "GitOps application did not converge; last status: %v", app.Object["status"])
	actualRevision, _, err := unstructured.NestedString(app.Object, "status", "sync", "revision")
	require.NoError(t, err)
	require.Equal(t, revision, actualRevision)
	response, err := cluster.Client.CoreV1().RESTClient().Get().
		AbsPath("/api/v1/namespaces", namespace, "services", "web:80", "proxy").DoRaw(t.Context())
	require.NoError(t, err)
	require.Contains(t, string(response), "<title>Welcome to nginx!</title>")

	deployment, err := cluster.Client.AppsV1().Deployments(namespace).Get(t.Context(), "web", metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, "docker.io/library/nginx:1.28.0", deployment.Spec.Template.Spec.Containers[0].Image)
	patch := []byte(`{"spec":{"template":{"spec":{"containers":[{"name":"helm-guestbook","image":"docker.io/library/nginx:1.27.5"}]}}}}`)
	patched, err := cluster.Client.AppsV1().Deployments(namespace).Patch(t.Context(), "web", types.StrategicMergePatchType, patch, metav1.PatchOptions{})
	require.NoError(t, err)
	require.Equal(t, "docker.io/library/nginx:1.27.5", patched.Spec.Template.Spec.Containers[0].Image)
	err = wait.PollUntilContextTimeout(t.Context(), 3*time.Second, 3*time.Minute, true, func(ctx context.Context) (bool, error) {
		current, err := cluster.Client.AppsV1().Deployments(namespace).Get(ctx, deployment.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return current.Spec.Template.Spec.Containers[0].Image == "docker.io/library/nginx:1.28.0" &&
			current.Status.ObservedGeneration == current.Generation && current.Status.UpdatedReplicas == 1 &&
			current.Status.Replicas == 1 && current.Status.AvailableReplicas == 1, nil
	})
	require.NoError(t, err, "GitOps controller did not repair workload drift")
}
