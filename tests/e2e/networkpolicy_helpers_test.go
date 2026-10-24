package e2e

import (
	"context"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func newPolicyProbe(ctx context.Context, template *corev1.Pod) *corev1.Pod {
	GinkgoHelper()
	candidate, err := lab.Client.CoreV1().Pods(testNamespace).Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-policy-probe-",
			Labels: map[string]string{
				"homelab.khuedoan.com/test-fixture": "true",
				denyLabel:                           "true",
			},
		},
		Spec: corev1.PodSpec{
			NodeName:                     template.Spec.NodeName,
			ServiceAccountName:           template.Spec.ServiceAccountName,
			AutomountServiceAccountToken: ptr.To(false),
			RestartPolicy:                corev1.RestartPolicyNever,
			Containers: []corev1.Container{
				{
					Name:      "main",
					Image:     template.Spec.Containers[0].Image,
					Command:   []string{"sleep", "300"},
					Resources: template.Spec.Containers[0].Resources,
				},
			},
		},
	}, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func(cleanup context.Context) error {
		err := lab.Client.CoreV1().Pods(testNamespace).Delete(cleanup, candidate.Name, metav1.DeleteOptions{
			Preconditions:      &metav1.Preconditions{UID: &candidate.UID},
			GracePeriodSeconds: ptr.To(int64(0)),
		})
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}, NodeTimeout(time.Minute))
	Expect(k8s.WaitUntilPodAvailableContextE(
		GinkgoTB(),
		ctx,
		lab.Kubectl(testNamespace),
		candidate.Name,
		30,
		time.Second,
	)).To(Succeed())

	return candidate
}
