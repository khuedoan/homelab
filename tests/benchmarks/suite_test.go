package benchmarks

import (
	"os"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/cluster"
	"github.com/khuedoan/homelab/tests/internal/inventory"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var lab cluster.Cluster

func TestBenchmarks(t *testing.T) {
	if testing.Short() || os.Getenv("E2E") != "1" || os.Getenv("BENCHMARKS") != "1" {
		t.Skip("live tests require E2E=1, an explicit environment, and benchmarks also require BENCHMARKS=1")
	}
	RegisterFailHandler(Fail)
	RunSpecs(t, "Homelab benchmarks")
}

var _ = BeforeSuite(func(ctx SpecContext) {
	t := GinkgoTB()
	By("loading the selected inventory and exporting its kubeconfig")
	lab = cluster.Connect(ctx, t, inventory.ForTest(t))
	By("verifying all expected API members before any workload mutations")
	cluster.VerifyCluster(ctx, t, lab)
	AddReportEntry("environment", lab.Target.Environment)
	AddReportEntry("nodes", lab.Target.Names)
	if len(lab.Target.Names) == 1 {
		AddReportEntry("coverage limitation", "single node: cross-node traffic and storage sharing cannot be exercised")
	}
}, NodeTimeout(5*time.Minute))
