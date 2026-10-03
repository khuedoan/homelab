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

func TestTerragruntStateBootstrap(t *testing.T) {
	terragrunt, err := exec.LookPath("terragrunt")
	if err != nil {
		t.Skip("Terragrunt is required for the state bootstrap integration test")
	}
	t.Setenv("TG_NON_INTERACTIVE", "true")
	t.Setenv("CLOUDFLARE_ACCOUNT_ID", "test-account")
	t.Setenv("CLOUDFLARE_TFSTATE_ACCESS_KEY", "test-access-key")
	t.Setenv("CLOUDFLARE_TFSTATE_SECRET_KEY", "test-secret-key")
	t.Setenv("TFSTATE_BUCKET", "tfstate-override")
	log := filepath.Join(t.TempDir(), "calls")
	t.Setenv("BOOTSTRAP_TEST_LOG", log)
	operationFake(t, "toolbox", `printf '%s\n' "$*" >> "$BOOTSTRAP_TEST_LOG"
if [ "${BOOTSTRAP_TEST_FAIL:-}" = yes ]; then exit 17; fi
printf 'State bucket ready.\n'
`)
	tofuLog := filepath.Join(t.TempDir(), "tofu-calls")
	t.Setenv("TOFU_TEST_LOG", tofuLog)
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
	for _, environment := range []string{"production", "staging"} {
		t.Run(environment, func(t *testing.T) {
			write := func(path, text string) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte(text), 0600); err != nil {
					t.Fatal(err)
				}
			}
			run := func(dir string, args ...string) (string, error) {
				t.Helper()
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
				defer cancel()
				command := exec.CommandContext(ctx, terragrunt, args...)
				command.Dir = dir
				output, err := command.CombinedOutput()
				return string(output), err
			}
			calls := func(minimum, maximum int) {
				t.Helper()
				data, err := os.ReadFile(log)
				if os.IsNotExist(err) && minimum == 0 {
					return
				}
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
			root, err := os.ReadFile(filepath.Join("..", "..", "..", "infra", environment, "root.hcl"))
			if err != nil {
				t.Fatal(err)
			}
			dir := filepath.Join(t.TempDir(), environment)
			write(filepath.Join(dir, "root.hcl"), string(root))
			for i := range 10 {
				unit := filepath.Join(dir, fmt.Sprintf("unit-%02d", i))
				write(filepath.Join(unit, "terragrunt.hcl"), `include "root" {
  path = find_in_parent_folders("root.hcl")
}
`)
				write(filepath.Join(unit, "main.tf"), `output "checked" { value = "local-only" }`)
			}
			for _, action := range []string{"init", "plan", "apply"} {
				write(log, "")
				args := []string{"run", "--all", "--parallelism", "10", "--", action, "-input=false"}
				if action == "apply" {
					args = append(args, "-auto-approve")
				}
				if output, err := run(dir, args...); err != nil {
					t.Fatalf("%s: %v\n%s", action, err, output)
				}
				calls(1, 10)
			}
			for i := range 10 {
				unit := fmt.Sprintf("unit-%02d", i)
				files, err := filepath.Glob(filepath.Join(dir, unit, ".terragrunt-cache", "*", "*", "backend.tf.json"))
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
							S3 struct{ Bucket, Key string }
						}
					}
				}
				if err := json.Unmarshal(data, &backend); err != nil {
					t.Fatal(err)
				}
				if backend.Terraform.Backend.S3.Bucket != "tfstate-override" || backend.Terraform.Backend.S3.Key != unit+"/tfstate.json" {
					t.Fatalf("wrong backend for %s: %+v", unit, backend)
				}
			}
			write(log, "")
			if output, err := run(filepath.Join(dir, "unit-00"), "run", "--", "init", "-input=false"); err != nil {
				t.Fatalf("single-unit init: %v\n%s", err, output)
			}
			calls(1, 1)
			write(log, "")
			if output, err := run(dir, "render", "--all", "--json"); err != nil {
				t.Fatalf("render: %v\n%s", err, output)
			}
			calls(0, 0)
			for _, name := range []string{"metal", "cluster"} {
				unit := filepath.Join(dir, name)
				write(filepath.Join(unit, "terragrunt.hcl"), "include \"root\" {\n  path = find_in_parent_folders(\"root.hcl\")\n}\n")
				write(filepath.Join(unit, "main.tf"), `output "checked" { value = "local-only" }`)
				if output, err := run(unit, "run", "--", "init", "-input=false"); err != nil {
					t.Fatalf("%s init: %v\n%s", name, err, output)
				}
				calls(0, 0)
			}
			t.Setenv("BOOTSTRAP_TEST_FAIL", "yes")
			write(tofuLog, "")
			if output, err := run(filepath.Join(dir, "unit-00"), "run", "--", "plan", "-input=false"); err == nil {
				t.Fatalf("bootstrap failure did not block plan: %v\n%s", err, output)
			}
			data, err := os.ReadFile(tofuLog)
			if err != nil || len(data) != 0 {
				t.Fatalf("OpenTofu ran after bootstrap failure: %q, %v", data, err)
			}
		})
	}
}
