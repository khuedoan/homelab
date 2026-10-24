package benchmarks

import (
	"encoding/json"
	"time"

	"github.com/khuedoan/homelab/tests/internal/cluster"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Storage", Label("storage"), func() {
	It("measures random reads and writes for every Ceph storage class", func(ctx SpecContext) {
		t := GinkgoTB()

		classes, err := cluster.DiscoverStorage(ctx, lab.Client)
		Expect(err).NotTo(HaveOccurred())
		Expect(classes).NotTo(BeEmpty(), "no Ceph storage classes found")
		for _, storage := range classes {
			By(storage.Class)
			namespace := cluster.Namespace(ctx, t, lab.Client)
			claim := cluster.Volume(ctx, t, lab.Client, namespace, storage, "1Gi")
			By("running the pinned fio workload")
			spec := cluster.PodSpec(t, "testdata/fio.yaml", map[string]string{
				"Node": lab.Target.Names[0], "Claim": claim.Name,
			})
			logs := cluster.RunJob(ctx, t, lab, namespace, "fio", spec)
			var report struct {
				Jobs []fioResult `json:"jobs"`
			}
			Expect(json.Unmarshal([]byte(logs), &report)).NotTo(HaveOccurred(), "%s", logs)
			Expect(report.Jobs).To(HaveLen(1))
			result := report.Jobs[0]
			AddReportEntry("fio/"+storage.Class, result)
			Expect(result.Error).To(BeZero(), "fio reported an I/O error")
			Expect(result.Read.Bytes).To(BeNumerically(">", 0), "fio must perform reads")
			Expect(result.Write.Bytes).To(BeNumerically(">", 0), "fio must perform writes")
			Expect(result.Read.Bandwidth).To(BeNumerically(">", 0))
			Expect(result.Write.Bandwidth).To(BeNumerically(">", 0))
			t.Logf("%s read %.1f IOPS, %d KiB/s; write %.1f IOPS, %d KiB/s", storage.Class, result.Read.IOPS, result.Read.Bandwidth, result.Write.IOPS, result.Write.Bandwidth)
		}
	}, SpecTimeout(20*time.Minute))
})

type fioResult struct {
	Error       int `json:"error"`
	Read, Write fioIO
}

type fioIO struct {
	Bytes     int64   `json:"io_bytes"`
	IOPS      float64 `json:"iops"`
	Bandwidth int64   `json:"bw"`
}
