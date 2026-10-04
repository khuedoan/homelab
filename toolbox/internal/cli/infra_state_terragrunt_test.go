package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type stateBootstrapFixture struct {
	dir, log, tofuLog, terragrunt string
}

func newStateBootstrapFixture(t *testing.T, environment string) stateBootstrapFixture {
	t.Helper()
	terragrunt, err := exec.LookPath("terragrunt")
	if err != nil {
		t.Skip("Terragrunt is required for the state bootstrap integration test")
	}
	t.Setenv("TG_NON_INTERACTIVE", "true")
	t.Setenv("CLOUDFLARE_ACCOUNT_ID", "test-account")
	t.Setenv("CLOUDFLARE_TFSTATE_ACCESS_KEY", "test-access-key")
	t.Setenv("CLOUDFLARE_TFSTATE_SECRET_KEY", "test-secret-key")
	t.Setenv("TFSTATE_BUCKET", "tfstate-override")
	fixture := stateBootstrapFixture{
		dir: filepath.Join(t.TempDir(), environment), log: filepath.Join(t.TempDir(), "calls"),
		tofuLog: filepath.Join(t.TempDir(), "tofu-calls"), terragrunt: terragrunt,
	}
	t.Setenv("BOOTSTRAP_TEST_LOG", fixture.log)
	t.Setenv("TOFU_TEST_LOG", fixture.tofuLog)
	operationFake(t, "toolbox", `printf '%s\n' "$*" >> "$BOOTSTRAP_TEST_LOG"
if [ "${BOOTSTRAP_TEST_FAIL:-}" = yes ]; then exit 17; fi
printf 'State bucket ready.\n'
`)
	operationFake(t, "tofu", `case "$*" in
*-version*) printf 'OpenTofu v1.11.0\n'; exit 0;;
esac
printf '%s\n' "$*" >> "$TOFU_TEST_LOG"
`)
	tofu, err := exec.LookPath("tofu")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("TG_TF_PATH", tofu)
	root, err := os.ReadFile(filepath.Join("..", "..", "..", "infra", environment, "root.hcl"))
	if err != nil {
		t.Fatal(err)
	}
	writeStateFixture(t, filepath.Join(fixture.dir, "root.hcl"), string(root))
	fixture.addUnit(t, "unit-00")
	fixture.resetCalls(t)
	return fixture
}

func writeStateFixture(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0600); err != nil {
		t.Fatal(err)
	}
}

func (f stateBootstrapFixture) addUnit(t *testing.T, name string) {
	t.Helper()
	unit := filepath.Join(f.dir, name)
	writeStateFixture(t, filepath.Join(unit, "terragrunt.hcl"), "include \"root\" {\n  path = find_in_parent_folders(\"root.hcl\")\n}\n")
	writeStateFixture(t, filepath.Join(unit, "main.tf"), `output "checked" { value = "local-only" }`)
}

func (f stateBootstrapFixture) run(t *testing.T, unit string, args ...string) (string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 2*time.Minute)
	defer cancel()
	command := exec.CommandContext(ctx, f.terragrunt, args...)
	command.Dir = filepath.Join(f.dir, unit)
	output, err := command.CombinedOutput()
	return string(output), err
}

func (f stateBootstrapFixture) resetCalls(t *testing.T) {
	t.Helper()
	writeStateFixture(t, f.log, "")
	writeStateFixture(t, f.tofuLog, "")
}

func (f stateBootstrapFixture) assertCalls(t *testing.T, minimum, maximum int) {
	t.Helper()
	data, err := os.ReadFile(f.log)
	if err != nil {
		t.Fatal(err)
	}
	line := "infra state ensure --account-id test-account --bucket tfstate-override\n"
	count := strings.Count(string(data), line)
	if count < minimum || count > maximum || string(data) != strings.Repeat(line, count) {
		t.Fatalf("bootstrap calls = %q, want %d to %d calls", data, minimum, maximum)
	}
	t.Logf("bootstrap command ran %d times", count)
}

func (f stateBootstrapFixture) assertBackend(t *testing.T, unit string) {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(f.dir, unit, ".terragrunt-cache", "*", "*", "backend.tf.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("generated backends for %s: %v, %v", unit, files, err)
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	var backend struct {
		Terraform struct {
			Backend struct {
				S3    struct{ Bucket, Key string }
				Local struct{ Path string }
			}
		}
	}
	if err := json.Unmarshal(data, &backend); err != nil {
		t.Fatal(err)
	}
	got := backend.Terraform.Backend
	if got.S3.Bucket != "tfstate-override" || got.S3.Key != unit+"/tfstate.json" || got.Local.Path != "" {
		t.Fatalf("wrong remote backend for %s: %+v", unit, got)
	}
}

func TestTerragruntStateBootstrapRunAll(t *testing.T) {
	for _, environment := range []string{"production", "staging"} {
		t.Run(environment, func(t *testing.T) {
			fixture := newStateBootstrapFixture(t, environment)
			for i := 1; i < 10; i++ {
				fixture.addUnit(t, fmt.Sprintf("unit-%02d", i))
			}
			for _, action := range []string{"init", "plan", "apply"} {
				fixture.resetCalls(t)
				args := []string{"run", "--all", "--parallelism", "10", "--", action, "-input=false"}
				if action == "apply" {
					args = append(args, "-auto-approve")
				}
				if output, err := fixture.run(t, "", args...); err != nil {
					t.Fatalf("%s: %v\n%s", action, err, output)
				}
				fixture.assertCalls(t, 1, 10)
			}
			for i := range 10 {
				fixture.assertBackend(t, fmt.Sprintf("unit-%02d", i))
			}
		})
	}
}

func TestTerragruntStateBootstrapSingleUnit(t *testing.T) {
	for _, tc := range []struct {
		environment, unit string
		calls             int
	}{
		{"production", "unit-00", 1}, {"staging", "unit-00", 1},
		{"production", "metal", 0}, {"production", "cluster", 0},
		{"staging", "metal", 0}, {"staging", "cluster", 0},
	} {
		t.Run(tc.environment+"/"+tc.unit, func(t *testing.T) {
			fixture := newStateBootstrapFixture(t, tc.environment)
			fixture.addUnit(t, tc.unit)
			if output, err := fixture.run(t, tc.unit, "run", "--", "init", "-input=false"); err != nil {
				t.Fatalf("init: %v\n%s", err, output)
			}
			fixture.assertCalls(t, tc.calls, tc.calls)
		})
	}
}

func TestTerragruntStateRenderIsOffline(t *testing.T) {
	for _, environment := range []string{"production", "staging"} {
		t.Run(environment, func(t *testing.T) {
			fixture := newStateBootstrapFixture(t, environment)
			if output, err := fixture.run(t, "", "render", "--all", "--json"); err != nil {
				t.Fatalf("render: %v\n%s", err, output)
			}
			fixture.assertCalls(t, 0, 0)
		})
	}
}

func TestTerragruntStateBootstrapFailure(t *testing.T) {
	for _, environment := range []string{"production", "staging"} {
		t.Run(environment, func(t *testing.T) {
			fixture := newStateBootstrapFixture(t, environment)
			t.Setenv("BOOTSTRAP_TEST_FAIL", "yes")
			output, planErr := fixture.run(t, "unit-00", "run", "--", "plan", "-input=false")
			if planErr == nil {
				t.Fatalf("bootstrap failure did not block plan\n%s", output)
			}
			data, err := os.ReadFile(fixture.tofuLog)
			if err != nil {
				t.Fatal(err)
			}
			if len(data) != 0 {
				t.Fatalf("OpenTofu ran after bootstrap failure: %q", data)
			}
		})
	}
}
