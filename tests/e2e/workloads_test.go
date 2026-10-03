package e2e

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func readyPod(t *testing.T, cluster fixture.Cluster, namespace, name string) *corev1.Pod {
	t.Helper()
	var pod *corev1.Pod
	err := wait.PollUntilContextTimeout(t.Context(), 2*time.Second, 3*time.Minute, true, func(ctx context.Context) (bool, error) {
		var err error
		pod, err = cluster.Client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if pod.Status.Phase == corev1.PodFailed {
			return false, fmt.Errorf("pod %s/%s failed: %s", namespace, name, pod.Status.Message)
		}
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
	require.NoError(t, err, "pod %s/%s did not become Ready; last observed pod: %+v", namespace, name, pod)
	return pod
}

func runPod(t *testing.T, cluster fixture.Cluster, namespace string, pod *corev1.Pod) string {
	t.Helper()
	pod.Spec.RestartPolicy = corev1.RestartPolicyNever
	created, err := cluster.Client.CoreV1().Pods(namespace).Create(t.Context(), pod, metav1.CreateOptions{})
	require.NoError(t, err)
	err = wait.PollUntilContextTimeout(t.Context(), 2*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		current, err := cluster.Client.CoreV1().Pods(namespace).Get(ctx, created.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		return current.Status.Phase == corev1.PodSucceeded || current.Status.Phase == corev1.PodFailed, nil
	})
	require.NoError(t, err, "pod %s/%s did not finish", namespace, created.Name)
	current, err := cluster.Client.CoreV1().Pods(namespace).Get(t.Context(), created.Name, metav1.GetOptions{})
	require.NoError(t, err)
	logs, err := cluster.Client.CoreV1().Pods(namespace).GetLogs(created.Name, &corev1.PodLogOptions{}).DoRaw(t.Context())
	require.NoError(t, err)
	require.Equal(t, corev1.PodSucceeded, current.Status.Phase, "pod %s/%s failed:\n%s", namespace, created.Name, logs)
	return string(logs)
}
