package cluster

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
// Cleanup is registered with t and waits for PV reclamation.
func Namespace(ctx context.Context, t testing.TB, client *kubernetes.Clientset) string {
	t.Helper()
	namespace, err := client.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
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
		if !apierrors.IsNotFound(err) {
			require.NoError(t, err, "delete namespace %s", namespace.Name)
		}
		err = wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
			_, err := client.CoreV1().Namespaces().Get(ctx, namespace.Name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		})
		require.NoError(t, err, "wait for namespace %s deletion", namespace.Name)
		require.NoError(t, waitForVolumeReclamation(ctx, client, namespace.Name), "volumes for namespace %s were not reclaimed", namespace.Name)
	})
	err = wait.PollUntilContextTimeout(ctx, 200*time.Millisecond, 30*time.Second, true, func(ctx context.Context) (bool, error) {
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
