package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	utilexec "k8s.io/client-go/util/exec"
	"k8s.io/utils/ptr"
)

var _ = Describe("NetworkPolicy", Label("networkpolicy"), func() {
	It("denies egress from a selected pod", func(ctx SpecContext) {
		server := fixturePod(ctx, "writer")
		probes := probePods(ctx)
		template := &probes[0]
		for i := range probes {
			if probes[i].Spec.NodeName != server.Spec.NodeName {
				template = &probes[i]
				break
			}
		}
		if len(lab.Target.Names) > 1 {
			Expect(template.Spec.NodeName).NotTo(Equal(server.Spec.NodeName))
		}
		control := fixturePod(ctx, "control")
		Expect(control.UID).NotTo(Equal(server.UID))
		address := "http://" + net.JoinHostPort(server.Status.PodIP, "8080")
		By("checking the backend from an independent unselected workload")
		Expect(execInPod(ctx, control, "wget", "-T", "5", "-qO-", address)).To(Equal("homelab-e2e"))

		By("denying a new connection from a fresh selected probe")
		var selected *corev1.Pod
		// Reusing a source can retain an HBONE connection established before policy propagation.
		Eventually(func(g Gomega, attempt context.Context) {
			candidate, err := lab.Client.CoreV1().Pods(testNamespace).Create(attempt, &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					GenerateName: "test-policy-probe-",
					Labels:       map[string]string{"homelab.khuedoan.com/test-fixture": "true", denyLabel: "true"},
				},
				Spec: corev1.PodSpec{
					NodeName:                     template.Spec.NodeName,
					ServiceAccountName:           template.Spec.ServiceAccountName,
					AutomountServiceAccountToken: ptr.To(false),
					RestartPolicy:                corev1.RestartPolicyNever,
					Containers: []corev1.Container{{
						Name: "main", Image: template.Spec.Containers[0].Image, Command: []string{"sleep", "300"},
						Resources: template.Spec.Containers[0].Resources,
					}},
				},
			}, metav1.CreateOptions{})
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(func(cleanup context.Context) {
				err := lab.Client.CoreV1().Pods(testNamespace).Delete(cleanup, candidate.Name, metav1.DeleteOptions{
					Preconditions: &metav1.Preconditions{UID: &candidate.UID}, GracePeriodSeconds: ptr.To(int64(0)),
				})
				if !apierrors.IsNotFound(err) {
					Expect(err).NotTo(HaveOccurred())
				}
			}, NodeTimeout(time.Minute))
			Expect(k8s.WaitUntilPodAvailableContextE(GinkgoTB(), attempt, lab.Kubectl(testNamespace), candidate.Name, 30, time.Second)).To(Succeed())
			output, err := lab.Exec(attempt, testNamespace, candidate.Name, "wget", "-T", "3", "-qO-", address)
			var exit utilexec.ExitError
			g.Expect(errors.As(err, &exit)).To(BeTrue(), "wget must fail inside the container, rather than the exec transport: %v; output: %q", err, output)
			g.Expect(exit.ExitStatus()).To(Equal(1))
			selected = candidate
		}).WithContext(ctx).WithTimeout(2 * time.Minute).WithPolling(time.Second).Should(Succeed())

		By("checking that the same backend remains healthy from an unselected workload")
		Expect(execInPod(ctx, control, "wget", "-T", "5", "-qO-", address)).To(Equal("homelab-e2e"))
		allowProbe(ctx, selected)
		Eventually(func(g Gomega) {
			output, err := lab.Exec(ctx, testNamespace, selected.Name, "wget", "-T", "5", "-qO-", address)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(output).To(Equal("homelab-e2e"))
		}).WithContext(ctx).WithTimeout(time.Minute).WithPolling(time.Second).Should(Succeed())
	}, SpecTimeout(5*time.Minute))
})

func allowProbe(ctx context.Context, pod *corev1.Pod) {
	GinkgoHelper()
	patch, err := json.Marshal(map[string]any{"metadata": map[string]any{"uid": pod.UID, "labels": map[string]any{denyLabel: nil}}})
	Expect(err).NotTo(HaveOccurred())
	_, err = lab.Client.CoreV1().Pods(testNamespace).Patch(ctx, pod.Name, types.MergePatchType, patch, metav1.PatchOptions{})
	if !apierrors.IsNotFound(err) {
		Expect(err).NotTo(HaveOccurred())
	}
}
