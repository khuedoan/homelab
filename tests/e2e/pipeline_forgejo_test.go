package e2e

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayclient "sigs.k8s.io/gateway-api/pkg/client/clientset/versioned"
)

const pipelineRepository = "ops/example-service"
const pipelineRepositoryAPI = "/repos/" + pipelineRepository

type pipelineForgejo struct {
	url      string
	registry string
	client   *http.Client
	password string
}

type pipelineRun struct {
	ID        int    `json:"id"`
	HTMLURL   string `json:"html_url"`
	CommitSHA string `json:"commit_sha"`
	Status    string `json:"status"`
}

func connectPipelineForgejo(ctx context.Context) pipelineForgejo {
	GinkgoHelper()

	gateway, err := gatewayclient.NewForConfig(lab.RESTConfig())
	Expect(err).NotTo(HaveOccurred())

	hostname := func(name string) string {
		route, err := gateway.GatewayV1().HTTPRoutes(name).Get(ctx, name, metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		Expect(route.Spec.Hostnames).To(HaveLen(1), "%s must have one public hostname", name)

		return string(route.Spec.Hostnames[0])
	}

	root, err := lab.Client.CoreV1().Secrets("openbao").Get(ctx, "openbao-unseal", metav1.GetOptions{})
	Expect(err).NotTo(HaveOccurred())
	Expect(root.Data["vault-root"]).NotTo(BeEmpty(), "OpenBao bootstrap credential must exist")
	client, err := lab.OpenBao()
	Expect(err).NotTo(HaveOccurred())
	client.SetToken(string(root.Data["vault-root"]))
	secret, err := client.KVv2("secret").Get(ctx, "forgejo/admin")
	Expect(err).NotTo(HaveOccurred())
	password, ok := secret.Data["password"].(string)
	Expect(ok && password != "").To(BeTrue(), "Forgejo bootstrap credential must exist")

	return pipelineForgejo{
		url:      "https://" + hostname("forgejo"),
		registry: hostname("zot"),
		password: password,
		client: &http.Client{
			Timeout:       15 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		},
	}
}

func (forgejo pipelineForgejo) get(ctx context.Context, path string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, forgejo.url+"/api/v1"+path, nil)
	if err != nil {
		return nil, err
	}
	request.SetBasicAuth("forgejo_admin", forgejo.password)
	response, err := forgejo.client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		_ = response.Body.Close()
		return response, fmt.Errorf("Forgejo GET %s returned HTTP %d", path, response.StatusCode)
	}

	return response, nil
}

func (forgejo pipelineForgejo) json(ctx context.Context, path string, output any) error {
	response, err := forgejo.get(ctx, path)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	return json.NewDecoder(response.Body).Decode(output)
}

func pushPipelineFixture(ctx context.Context, checkout string, forgejo pipelineForgejo) string {
	GinkgoHelper()

	remote := forgejo.url + "/" + pipelineRepository + ".git"
	basic := base64.StdEncoding.EncodeToString([]byte("forgejo_admin:" + forgejo.password))
	git := func(args ...string) string {
		command := exec.CommandContext(ctx, "git", args...)
		command.Dir = checkout
		command.Env = append(os.Environ(),
			"GIT_TERMINAL_PROMPT=0", "GIT_CONFIG_COUNT=1",
			"GIT_CONFIG_KEY_0=http."+forgejo.url+"/.extraHeader",
			"GIT_CONFIG_VALUE_0=Authorization: Basic "+basic,
		)
		output, err := command.Output()
		Expect(err == nil).To(BeTrue(), "git %s failed", args[0])

		return strings.TrimSpace(string(output))
	}

	git("init", "--initial-branch=main")
	response, err := forgejo.get(ctx, pipelineRepositoryAPI)
	if err != nil {
		Expect(response).NotTo(BeNil(), "%v", err)
		Expect(response.StatusCode).To(Equal(http.StatusNotFound), "%v", err)
	} else {
		var repository struct {
			Empty   bool `json:"empty"`
			Private bool `json:"private"`
		}
		decodeErr := json.NewDecoder(response.Body).Decode(&repository)
		_ = response.Body.Close()
		Expect(decodeErr).NotTo(HaveOccurred())
		Expect(repository.Private).To(BeFalse(), "example-service must be public")

		if !repository.Empty {
			git("fetch", "--depth=1", remote, "refs/heads/main")
			// Use the fetched commit as the parent while keeping the index empty.
			// The new commit replaces the tree with the checked-in fixture.
			git("reset", "--soft", "FETCH_HEAD")
		}
	}

	Expect(os.CopyFS(checkout, os.DirFS("testdata/example-service"))).To(Succeed())
	git("add", ".")
	git("-c", "user.name=E2E fixture", "-c", "user.email=ci@example.invalid",
		"commit", "--allow-empty", "-m", "Run Nix pipeline fixture")
	revision := git("rev-parse", "HEAD")
	git("push", "-o", "repo.private=false", remote, "HEAD:refs/heads/main")
	AddReportEntry("pushed revision", revision)

	return revision
}

func waitForPipeline(ctx context.Context, forgejo pipelineForgejo, revision string) pipelineRun {
	GinkgoHelper()

	var run pipelineRun
	Eventually(func(g Gomega) {
		var payload struct {
			WorkflowRuns []pipelineRun `json:"workflow_runs"`
		}
		g.Expect(forgejo.json(ctx, pipelineRepositoryAPI+"/actions/runs?limit=12", &payload)).To(Succeed())
		for _, candidate := range payload.WorkflowRuns {
			if candidate.CommitSHA == revision {
				run = candidate
				break
			}
		}
		switch run.Status {
		case "failure", "cancelled", "skipped", "timed_out":
			StopTrying(fmt.Sprintf("workflow %s finished with status %s", run.HTMLURL, run.Status)).Now()
		default:
			g.Expect(run.Status).To(Equal("success"), "workflow %s", run.HTMLURL)
		}
	}).WithContext(ctx).WithTimeout(8 * time.Minute).WithPolling(2 * time.Second).Should(Succeed())

	return run
}

func publishedPipelineImage(ctx context.Context, forgejo pipelineForgejo, runID int) string {
	GinkgoHelper()

	var jobs []struct {
		ID int `json:"id"`
	}
	Expect(forgejo.json(ctx, pipelineRepositoryAPI+"/actions/runs/"+strconv.Itoa(runID)+"/jobs", &jobs)).To(Succeed())
	Expect(jobs).To(HaveLen(1))
	response, err := forgejo.get(ctx, pipelineRepositoryAPI+"/actions/jobs/"+strconv.Itoa(jobs[0].ID)+"/logs")
	Expect(err).NotTo(HaveOccurred())
	data, err := io.ReadAll(io.LimitReader(response.Body, 4<<20))
	closeErr := response.Body.Close()
	Expect(err).NotTo(HaveOccurred())
	Expect(closeErr).NotTo(HaveOccurred())
	Expect(bytes.Contains(data, []byte("Cross-repository upload denied (HTTP 403)"))).To(BeTrue())
	image := regexp.MustCompile("Published (" + regexp.QuoteMeta(forgejo.registry+"/"+pipelineRepository) + "@sha256:[a-f0-9]{64})").FindSubmatch(data)
	Expect(image).To(HaveLen(2), "workflow must report the published Zot digest")

	return string(image[1])
}
