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
		"text":{"sensitive":true,"value":"plain-secret"},
		"object":{"value":{"password":"secret","user":"example"}},
		"number":{"value":9007199254740993},
		"list":{"value":["a",2]},
		"boolean":{"value":true},
		"empty":{"value":""}
	}`
	got, err := outputRecords([]byte(input), "foo/bar")
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]map[string]string{
		"infra/foo/bar/text":    {"value": "plain-secret"},
		"infra/foo/bar/object":  {"value": `{"password":"secret","user":"example"}`},
		"infra/foo/bar/number":  {"value": "9007199254740993"},
		"infra/foo/bar/list":    {"value": `["a",2]`},
		"infra/foo/bar/boolean": {"value": "true"},
		"infra/foo/bar/empty":   {"value": ""},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected output mapping: %v", got)
	}
	for _, input := range []string{`null`, `[]`, `{"bad":{"sensitive":"secret-canary"}}`, `{"../bad":{"value":"secret-canary"}}`, `{} {"value":"secret-canary"}`} {
		_, err := outputRecords([]byte(input), "foo/bar")
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
  */foo/bar) printf '%s\n' '{"text":{"value":"secret-canary"}}' ;;
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
	if !reflect.DeepEqual(got, map[string]map[string]string{"infra/foo/bar/text": {"value": "secret-canary"}}) {
		t.Fatalf("unexpected environment output mapping: %v", got)
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
