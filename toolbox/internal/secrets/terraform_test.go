package secrets

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestOutputRecords(t *testing.T) {
	input := `{
        "ordinary": {"value": "not-published"},
        "secrets": {"sensitive":true,"value":{
            "cert-manager/cloudflare":{"api_token":"token-canary"},
            "external-dns/cloudflare":{"api_token":"dns-canary","zone_id":"zone-id"},
            "example/nested/secret":{"empty":"","json":"{\"enabled\":true}"}
        }}
    }`
	got, err := outputRecords([]byte(input), "cloudflare")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]map[string]string{
		"cert-manager/cloudflare": {"api_token": "token-canary"},
		"external-dns/cloudflare": {"api_token": "dns-canary", "zone_id": "zone-id"},
		"example/nested/secret":   {"empty": "", "json": `{"enabled":true}`},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected secret exports: %v", got)
	}
	for _, input := range []string{`{}`, `{"ordinary":{"value":123}}`, `{"secrets":{"value":{}}}`} {
		got, err := outputRecords([]byte(input), "module")
		if err != nil || len(got) != 0 {
			t.Fatalf("module without exports: %v, %v", got, err)
		}
	}
	for _, input := range []string{
		`null`, `[]`, `{} {"value":"secret-canary"}`,
		`{"secrets":{"value":null}}`,
		`{"secrets":{"value":{"namespace/secret":{"token":42}}}}`,
		`{"secrets":{"value":{"namespace/secret":{"token":null}}}}`,
		`{"secrets":{"value":{"namespace/secret":null}}}`,
		`{"secrets":{"value":{"namespace/secret":{}}}}`,
		`{"secrets":{"value":{"namespace/secret":{"":"secret-canary"}}}}`,
		`{"secrets":{"value":{"../secret":{"token":"secret-canary"}}}}`,
		`{"secrets":{"value":{"namespace/../other":{"token":"secret-canary"}}}}`,
		`{"secrets":{"value":{"namespace//secret":{"token":"secret-canary"}}}}`,
		`{"secrets":{"value":{"namespace/":{"token":"secret-canary"}}}}`,
		`{"secrets":{"value":{"namespace":{"token":"secret-canary"}}}}`,
		`{"secrets":{"value":{"Not-A-Namespace/secret":{"token":"secret-canary"}}}}`,
	} {
		_, err := outputRecords([]byte(input), "module")
		if err == nil || strings.Contains(err.Error(), "secret-canary") {
			t.Fatalf("unsafe malformed-output error: %v", err)
		}
	}
}

func TestEnvironmentRecords(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	for _, file := range []string{"infra/test/foo/bar/terragrunt.hcl", "infra/test/empty/terragrunt.hcl", "infra/test/.terragrunt-cache/copied/terragrunt.hcl", "infra/other/unit/terragrunt.hcl"} {
		if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, nil, 0600); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(root, "bin")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	script := `#!/bin/sh
case "$PWD" in
  */foo/bar) printf '%s\n' '{"secrets":{"value":{"example/token":{"value":"secret-canary"}}}}' ;;
  */empty) printf '%s\n' '{}' ;;
  *) exit 91 ;;
esac
`
	if err := os.WriteFile(filepath.Join(bin, "terragrunt"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	got, err := EnvironmentRecords(context.Background(), "test")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, map[string]map[string]string{"example/token": {"value": "secret-canary"}}) {
		t.Fatalf("unexpected environment output mapping: %v", got)
	}
	duplicate := filepath.Join(root, "infra/test/duplicate")
	if err := os.MkdirAll(duplicate, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(duplicate, "terragrunt.hcl"), nil, 0600); err != nil {
		t.Fatal(err)
	}
	duplicateScript := strings.Replace(script, "*/foo/bar)", "*/foo/bar|*/duplicate)", 1)
	if err := os.WriteFile(filepath.Join(bin, "terragrunt"), []byte(duplicateScript), 0700); err != nil {
		t.Fatal(err)
	}
	_, err = EnvironmentRecords(context.Background(), "test")
	if err == nil || !strings.Contains(err.Error(), "duplicate secret destination example/token") {
		t.Fatalf("duplicate destination accepted: %v", err)
	}
	if err := os.WriteFile(filepath.Join(bin, "terragrunt"), []byte("#!/bin/sh\nprintf 'secret-canary' >&2\nexit 1\n"), 0700); err != nil {
		t.Fatal(err)
	}
	_, err = EnvironmentRecords(context.Background(), "test")
	if err == nil || strings.Contains(err.Error(), "secret-canary") {
		t.Fatalf("unsafe command failure: %v", err)
	}
	for _, env := range []string{"../test", "/test", "test/foo", "missing"} {
		if _, err := EnvironmentRecords(context.Background(), env); err == nil {
			t.Fatalf("accepted invalid environment %q", env)
		}
	}
}
