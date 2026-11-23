package e2e

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("Pipelines", Label("pipelines"), func() {
	It("publishes a Nix-built image to Zot and runs the pushed revision", func(ctx SpecContext) {
		By("pushing the fixture to ops/example-service and waiting for its exact revision to succeed")
		forgejo := connectPipelineForgejo(ctx)
		revision := pushPipelineFixture(ctx, GinkgoTB().TempDir(), forgejo)
		run := waitForPipeline(ctx, forgejo, revision)
		AddReportEntry("pipeline run", run.HTMLURL)
		image := publishedPipelineImage(ctx, forgejo, run.ID)
		AddReportEntry("published image", image)

		By("pulling the published digest and checking the running service")
		verifyPipelineImage(ctx, "example-service-"+revision[:12], image, revision)
	}, SpecTimeout(14*time.Minute))
})
