package e2e

import (
	"context"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func verifyPipelineImage(ctx context.Context, name, image, revision string) {
	GinkgoHelper()

	pods := lab.Client.CoreV1().Pods("zot")
	pod, err := pods.Create(ctx, &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Spec: corev1.PodSpec{
			RestartPolicy:                corev1.RestartPolicyNever,
			AutomountServiceAccountToken: ptr.To(false),
			Containers: []corev1.Container{{
				Name:            "service",
				Image:           image,
				ImagePullPolicy: corev1.PullAlways,
				SecurityContext: pipelineContainerSecurity(),
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("10m"),
						corev1.ResourceMemory: resource.MustParse("16Mi"),
					},
					Limits: corev1.ResourceList{
						corev1.ResourceCPU:    resource.MustParse("200m"),
						corev1.ResourceMemory: resource.MustParse("64Mi"),
					},
				},
			}},
		},
	}, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred())
	DeferCleanup(func(ctx context.Context) {
		Expect(pods.Delete(ctx, name, metav1.DeleteOptions{
			Preconditions: &metav1.Preconditions{UID: &pod.UID},
		})).To(Succeed())
	}, NodeTimeout(time.Minute))

	_, digest, _ := strings.Cut(image, "@")
	Eventually(func(g Gomega) {
		current, err := pods.Get(ctx, name, metav1.GetOptions{})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(current.Spec.ImagePullSecrets).To(BeEmpty())
		g.Expect(current.Status.ContainerStatuses).To(HaveLen(1))

		service := current.Status.ContainerStatuses[0]
		g.Expect(service.Ready).To(BeTrue(), "fixture Pod must be running")
		g.Expect(service.ImageID).To(HaveSuffix(digest))

		body, err := pods.ProxyGet("http", name, "8080", "info", nil).DoRaw(ctx)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(strings.TrimSpace(string(body))).To(Equal(revision))
	}).WithContext(ctx).WithTimeout(2 * time.Minute).Should(Succeed())
}

func pipelineContainerSecurity() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: ptr.To(false),
		Capabilities:             &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}},
		SeccompProfile:           &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
	}
}
