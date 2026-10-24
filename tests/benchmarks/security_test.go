package benchmarks

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/khuedoan/homelab/tests/internal/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Security", Label("security"), func() {
	It("checks master node and etcd CIS controls on every node", func(ctx SpecContext) {
		t := GinkgoTB()

		for _, node := range lab.Target.Names {
			By(node)
			auditNode(ctx, t, node)
		}
	}, SpecTimeout(30*time.Minute))
})

func auditNode(ctx context.Context, t testing.TB, node string) {
	t.Helper()
	namespace := cluster.Namespace(ctx, t, lab.Client)
	By("running CIS checks on " + node)
	spec := cluster.PodSpec(t, "testdata/kube-bench.yaml", map[string]string{"Node": node})
	logs := cluster.RunJob(ctx, t, lab, namespace, "kube-bench", spec)
	var report struct {
		Controls []cisControl
	}
	Expect(json.Unmarshal([]byte(logs), &report)).NotTo(HaveOccurred(), "%s", logs)
	targets := make([]string, 0, len(report.Controls))
	for _, result := range report.Controls {
		targets = append(targets, result.Target)
		AddReportEntry("cis/"+node+"/"+result.Target, result)
		t.Logf("%s %s CIS checks: %d pass, %d fail, %d manual warnings", node, result.Target, result.Pass, result.Fail, result.Warn)
		Expect(result.Pass+result.Fail).To(BeNumerically(">", 0), "CIS checks must execute rather than only warn")
		Expect(result.Fail).To(BeZero(), "CIS failures on %s/%s:\n%s", node, result.Target, logs)
	}
	Expect(targets).To(ConsistOf([]string{"master", "node", "etcd"}))
}

type cisControl struct {
	Target string `json:"node_type"`
	Pass   int    `json:"total_pass"`
	Fail   int    `json:"total_fail"`
	Warn   int    `json:"total_warn"`
}
