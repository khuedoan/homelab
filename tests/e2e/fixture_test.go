package e2e

import (
	"context"
	"time"

	"github.com/gruntwork-io/terratest/modules/k8s/v2"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const testNamespace = "test"
const otherNamespace = "test-other"
const denyLabel = "homelab.khuedoan.com/deny-egress"

type sourceVolume struct {
	Class, Claim, Path string
}

var sourceVolumes = []sourceVolume{
	{Class: "standard-rwo", Claim: "test-rwo", Path: "/rwo"},
	{Class: "standard-rwx", Claim: "test-rwx", Path: "/rwx"},
}

func fixturePod(ctx context.Context, controller string) *corev1.Pod {
	GinkgoHelper()
	var current corev1.Pod
	Eventually(func(g Gomega) {
		pods, err := lab.Client.CoreV1().Pods(testNamespace).List(ctx, metav1.ListOptions{LabelSelector: "app.kubernetes.io/instance=test,app.kubernetes.io/controller=" + controller})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(pods.Items).To(HaveLen(1))
		current = pods.Items[0]
		g.Expect(current.DeletionTimestamp).To(BeNil())
		g.Expect(k8s.IsPodAvailable(&current)).To(BeTrue(), "test %s must be ready: %+v", controller, current.Status)
	}).WithContext(ctx).WithTimeout(3 * time.Minute).WithPolling(time.Second).Should(Succeed())
	return &current
}

func probePods(ctx context.Context) []corev1.Pod {
	GinkgoHelper()
	var probes []corev1.Pod
	Eventually(func(g Gomega) {
		list, err := lab.Client.CoreV1().Pods(testNamespace).List(ctx, metav1.ListOptions{LabelSelector: "app.kubernetes.io/instance=test,app.kubernetes.io/controller=probe"})
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(list.Items).To(HaveLen(len(lab.Target.Names)))
		var nodes []string
		for _, pod := range list.Items {
			g.Expect(pod.DeletionTimestamp).To(BeNil())
			g.Expect(k8s.IsPodAvailable(&pod)).To(BeTrue(), "probe %s must be ready: %+v", pod.Name, pod.Status)
			nodes = append(nodes, pod.Spec.NodeName)
		}
		g.Expect(nodes).To(ConsistOf(lab.Target.Names), "test probes must cover exactly the inventory nodes")
		probes = list.Items
	}).WithContext(ctx).WithTimeout(3 * time.Minute).WithPolling(time.Second).Should(Succeed())
	return probes
}

func execInPod(ctx context.Context, pod *corev1.Pod, command ...string) string {
	GinkgoHelper()
	output, err := lab.Exec(ctx, pod.Namespace, pod.Name, command...)
	Expect(err).NotTo(HaveOccurred())
	return output
}
