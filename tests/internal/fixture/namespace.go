package fixture

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
)

// Namespace creates an isolated namespace and waits for its default ServiceAccount.
// Cleanup waits for namespace deletion and reclamation of all PVs bound to its claims.
func Namespace(t *testing.T, client *kubernetes.Clientset) string {
	t.Helper()
	namespace, err := client.CoreV1().Namespaces().Create(t.Context(), &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{GenerateName: "homelab-test-"},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		uid := namespace.UID
		err := client.CoreV1().Namespaces().Delete(ctx, namespace.Name, metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{UID: &uid},
		})
		if err != nil && !apierrors.IsNotFound(err) {
			t.Errorf("delete namespace %s: %v", namespace.Name, err)
			return
		}
		err = wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
			_, err := client.CoreV1().Namespaces().Get(ctx, namespace.Name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		})
		if err != nil {
			t.Errorf("wait for namespace %s deletion: %v", namespace.Name, err)
			return
		}
		if err := waitForVolumeReclamation(ctx, client, namespace.Name); err != nil {
			t.Errorf("volumes for namespace %s were not reclaimed: %v", namespace.Name, err)
		}
	})
	err = wait.PollUntilContextTimeout(t.Context(), 200*time.Millisecond, 30*time.Second, true, func(ctx context.Context) (bool, error) {
		_, err := client.CoreV1().ServiceAccounts(namespace.Name).Get(ctx, "default", metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return err == nil, err
	})
	require.NoError(t, err, "namespace must have its default ServiceAccount before creating workloads")
	return namespace.Name
}

func waitForVolumeReclamation(ctx context.Context, client *kubernetes.Clientset, namespace string) error {
	return wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
		volumes, err := client.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
		if err != nil {
			return false, err
		}
		for _, volume := range volumes.Items {
			if volume.Spec.ClaimRef != nil && volume.Spec.ClaimRef.Namespace == namespace {
				return false, nil
			}
		}
		return true, nil
	})
}
