package benchmarks

import (
	"context"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/fixture"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
)

func runJob(t *testing.T, cluster fixture.Cluster, namespace, name string, spec corev1.PodSpec) string {
	t.Helper()
	spec.RestartPolicy = corev1.RestartPolicyNever
	spec.AutomountServiceAccountToken = ptr.To(false)
	job, err := cluster.Client.BatchV1().Jobs(namespace).Create(t.Context(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: batchv1.JobSpec{
			BackoffLimit: ptr.To[int32](0), ActiveDeadlineSeconds: ptr.To[int64](600),
			Template: corev1.PodTemplateSpec{Spec: spec},
		},
	}, metav1.CreateOptions{})
	require.NoError(t, err)
	err = wait.PollUntilContextTimeout(t.Context(), 3*time.Second, 11*time.Minute, true, func(ctx context.Context) (bool, error) {
		current, err := cluster.Client.BatchV1().Jobs(namespace).Get(ctx, job.Name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		job = current
		for _, condition := range job.Status.Conditions {
			if (condition.Type == batchv1.JobComplete || condition.Type == batchv1.JobFailed) && condition.Status == corev1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
	require.NoError(t, err, "Job %s/%s did not finish", namespace, name)
	pods, err := cluster.Client.CoreV1().Pods(namespace).List(t.Context(), metav1.ListOptions{
		LabelSelector: "batch.kubernetes.io/controller-uid=" + string(job.UID),
	})
	require.NoError(t, err)
	require.Len(t, pods.Items, 1, "benchmark must run once without retries")
	logs, err := cluster.Client.CoreV1().Pods(namespace).GetLogs(pods.Items[0].Name, &corev1.PodLogOptions{}).DoRaw(t.Context())
	require.NoError(t, err)
	require.Equal(t, int32(1), job.Status.Succeeded, "Job %s/%s failed:\n%s", namespace, name, logs)
	return string(logs)
}
