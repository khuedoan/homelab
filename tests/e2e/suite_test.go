package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/cluster"
	"github.com/khuedoan/homelab/tests/internal/inventory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	coordinationv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/utils/ptr"
)

var lab cluster.Cluster
var fixtureLease *coordinationv1.Lease

func TestCluster(t *testing.T) {
	if testing.Short() || os.Getenv("E2E") != "1" {
		t.Skip("live tests require E2E=1 and an explicit environment")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Homelab e2e")
}

var _ = BeforeSuite(func(ctx SpecContext) {
	t := GinkgoTB()
	suite, _ := GinkgoConfiguration()
	Expect(suite.ParallelTotal).To(Equal(1), "shared test fixtures require a single Ginkgo process")
	By("loading the selected inventory and exporting its kubeconfig")
	lab = cluster.Connect(ctx, t, inventory.ForTest(t))
	By("verifying all expected API members before any workload mutations")
	cluster.VerifyCluster(ctx, t, lab)
	By("locking and verifying the permanent test application")
	namespace, err := lab.Client.CoreV1().Namespaces().Get(ctx, testNamespace, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred(), "deploy apps/test.yaml through Argo CD before running e2e")
	Expect(namespace.Labels).To(HaveKeyWithValue("homelab.khuedoan.com/test-fixture", "true"))
	lease, err := lab.Client.CoordinationV1().Leases(testNamespace).Create(ctx, &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{Name: "e2e-run"},
		Spec:       coordinationv1.LeaseSpec{HolderIdentity: ptr.To(string(uuid.NewUUID())), AcquireTime: ptr.To(metav1.NewMicroTime(time.Now()))},
	}, metav1.CreateOptions{})
	Expect(err).NotTo(HaveOccurred(), "another run owns test/e2e-run; a crashed run requires explicit stale-lock removal")
	fixtureLease = lease

	fixturePod(ctx, "control")
	fixturePod(ctx, "writer")
	probes := probePods(ctx)
	for _, pod := range probes {
		allowProbe(ctx, &pod)
	}

	other, err := lab.Client.CoreV1().Namespaces().Get(ctx, otherNamespace, metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(other.Labels).To(HaveKeyWithValue("homelab.khuedoan.com/test-fixture", "true"))
	AddReportEntry("environment", lab.Target.Environment)
	AddReportEntry("nodes", lab.Target.Names)
	if len(lab.Target.Names) == 1 {
		AddReportEntry("coverage limitation", "single node: cross-node traffic and storage sharing cannot be exercised")
	}
}, NodeTimeout(10*time.Minute))

var _ = ReportAfterSuite("release the permanent fixture after a successful run", func(report Report) {
	if fixtureLease == nil {
		return
	}
	if !report.SuiteSucceeded {
		GinkgoWriter.Printf("Retaining test/e2e-run after a failed suite. Inspect its test resources and any running Velero operations before removing the lock.\n")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	Expect(lab.Client.CoordinationV1().Leases(testNamespace).Delete(ctx, fixtureLease.Name, metav1.DeleteOptions{
		Preconditions: &metav1.Preconditions{UID: &fixtureLease.UID},
	})).To(Succeed())
})
