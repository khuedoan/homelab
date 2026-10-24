package e2e

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

var _ = Describe("GitOps", Label("gitops"), func() {
	It("repairs a deliberate change to the deployed image", func(ctx SpecContext) {
		client, err := dynamic.NewForConfig(lab.RESTConfig())
		Expect(err).NotTo(HaveOccurred())
		applications := client.Resource(schema.GroupVersionResource{Group: "argoproj.io", Version: "v1alpha1", Resource: "applications"}).Namespace("argocd")
		Eventually(func(g Gomega) {
			application, err := applications.Get(ctx, "test", metav1.GetOptions{})
			g.Expect(err).NotTo(HaveOccurred())
			sync, _, err := unstructured.NestedString(application.Object, "status", "sync", "status")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(sync).To(Equal("Synced"), "application argocd/test")
			health, _, err := unstructured.NestedString(application.Object, "status", "health", "status")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(health).To(Equal("Healthy"), "application argocd/test")
		}).WithContext(ctx).WithTimeout(5 * time.Minute).WithPolling(3 * time.Second).Should(Succeed())

		deployments := lab.Client.AppsV1().Deployments(testNamespace)
		original, err := deployments.Get(ctx, "test-writer", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(original.Spec.Template.Spec.Containers).To(HaveLen(1))
		Expect(original.Spec.Template.Spec.Containers[0].Image).To(Equal("docker.io/library/busybox:1.37.0"))
		DeferCleanup(func(ctx SpecContext) {
			restore := []byte(`{"metadata":{"uid":"` + string(original.UID) + `"},"spec":{"template":{"spec":{"containers":[{"name":"main","image":"docker.io/library/busybox:1.37.0"}]}}}}`)
			_, err := deployments.Patch(ctx, original.Name, types.StrategicMergePatchType, restore, metav1.PatchOptions{})
			Expect(err).NotTo(HaveOccurred())
		}, NodeTimeout(time.Minute))
		patch := []byte(`{"spec":{"template":{"spec":{"containers":[{"name":"main","image":"docker.io/library/busybox:1.36.1"}]}}}}`)
		changed, err := deployments.Patch(ctx, original.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(changed.Spec.Template.Spec.Containers).To(HaveLen(1))
		Expect(changed.Spec.Template.Spec.Containers[0].Image).To(Equal("docker.io/library/busybox:1.36.1"))

		Eventually(func(g Gomega) {
			current, err := deployments.Get(ctx, original.Name, metav1.GetOptions{})
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(current.UID).To(Equal(original.UID))
			g.Expect(current.Spec.Template.Spec.Containers).To(HaveLen(1))
			g.Expect(current.Spec.Template.Spec.Containers[0].Image).To(Equal("docker.io/library/busybox:1.37.0"))
			g.Expect(current.Status.ObservedGeneration).To(Equal(current.Generation))
			g.Expect(current.Status.UpdatedReplicas).To(Equal(int32(1)))
			g.Expect(current.Status.Replicas).To(Equal(int32(1)))
			g.Expect(current.Status.AvailableReplicas).To(Equal(int32(1)))
		}).WithContext(ctx).WithTimeout(3 * time.Minute).WithPolling(3 * time.Second).Should(Succeed())
		Expect(execInPod(ctx, fixturePod(ctx, "writer"), "wget", "-T", "5", "-qO-", "http://127.0.0.1:8080")).To(Equal("homelab-e2e"))
	}, SpecTimeout(10*time.Minute))
})
