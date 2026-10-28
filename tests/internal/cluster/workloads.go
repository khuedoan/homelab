package cluster

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gruntwork-io/terratest/modules/core/v2/retry"
	"github.com/gruntwork-io/terratest/modules/k8s/v2"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
)

// Kubectl selects the exported kubeconfig and an explicit namespace.
func (cluster Cluster) Kubectl(namespace string) *k8s.KubectlOptions {
	options := k8s.NewKubectlOptions("", cluster.Kubeconfig, namespace)
	options.RequestTimeout = 10 * time.Second
	return options
}

// RunPod runs one transient pod, returns its exact logs, and deletes it even on failure.
func RunPod(ctx context.Context, t testing.TB, cluster Cluster, namespace string, pod *corev1.Pod) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	pod.Spec.RestartPolicy = corev1.RestartPolicyNever
	created, err := cluster.Client.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	require.NoError(t, err)
	defer func() {
		cleanup, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		uid := created.UID
		err := cluster.Client.CoreV1().Pods(namespace).Delete(cleanup, created.Name, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid}})
		if err != nil && !apierrors.IsNotFound(err) {
			t.Errorf("delete transient pod %s/%s: %v", namespace, created.Name, err)
		}
	}()
	err = wait.PollUntilContextCancel(ctx, 2*time.Second, true, func(ctx context.Context) (bool, error) {
		current, err := cluster.Client.CoreV1().Pods(namespace).Get(ctx, created.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		switch current.Status.Phase {
		case corev1.PodSucceeded:
			return true, nil
		case corev1.PodFailed:
			logs, logErr := podLogs(ctx, cluster, namespace, created.Name)
			return false, fmt.Errorf("pod %s/%s failed: %s; logs: %s; log error: %v", namespace, created.Name, current.Status.Message, logs, logErr)
		default:
			return false, nil
		}
	})
	require.NoError(t, err)
	logs, err := podLogs(ctx, cluster, namespace, created.Name)
	require.NoError(t, err)
	return logs
}

// RunJob runs exactly one bounded Job attempt and returns its pod logs.
func RunJob(ctx context.Context, t testing.TB, cluster Cluster, namespace, name string, spec corev1.PodSpec) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(ctx, 11*time.Minute)
	defer cancel()
	spec.RestartPolicy = corev1.RestartPolicyNever
	spec.AutomountServiceAccountToken = ptr.To(false)
	job, err := cluster.Client.BatchV1().Jobs(namespace).Create(ctx, &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec:       batchv1.JobSpec{BackoffLimit: ptr.To[int32](0), ActiveDeadlineSeconds: ptr.To[int64](600), Template: corev1.PodTemplateSpec{Spec: spec}},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	options := cluster.Kubectl(namespace)
	retry.DoWithRetryContext(t, ctx, "wait for job "+namespace+"/"+name, 220, 3*time.Second, func() (string, error) {
		current, err := k8s.GetJobContextE(t, ctx, options, name)
		if err != nil {
			return "", retry.FatalError{Underlying: err}
		}
		job = current
		for _, condition := range job.Status.Conditions {
			if condition.Status != corev1.ConditionTrue {
				continue
			}
			switch condition.Type {
			case batchv1.JobComplete:
				return "job completed", nil
			case batchv1.JobFailed:
				return "job failed", nil
			}
		}
		return "", fmt.Errorf("job %s/%s has not finished", namespace, name)
	})
	pods := k8s.ListPodsContext(t, ctx, options, metav1.ListOptions{LabelSelector: "batch.kubernetes.io/controller-uid=" + string(job.UID)})
	require.Len(t, pods, 1, "job must run once without retries")
	logs, err := podLogs(ctx, cluster, namespace, pods[0].Name)
	require.NoError(t, err)
	require.Equal(t, int32(1), job.Status.Succeeded, "job %s/%s did not succeed:\n%s", namespace, name, logs)
	return logs
}

func podLogs(ctx context.Context, cluster Cluster, namespace, name string) (string, error) {
	data, err := cluster.Client.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{}).DoRaw(ctx)
	return string(data), err
}
