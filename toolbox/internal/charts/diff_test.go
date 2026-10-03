package charts

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/khuedoan/homelab/toolbox/internal/process"
)

func helmDiffWrite(t *testing.T, path, data string, mode os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(data), mode); err != nil {
		t.Fatal(err)
	}
}

func helmDiffGit(t *testing.T, repo string, args ...string) string {
	t.Helper()
	command := exec.Command("git", append([]string{"-C", repo}, args...)...)
	command.Env = append(os.Environ(), "GIT_AUTHOR_NAME=Test", "GIT_AUTHOR_EMAIL=test@example.com", "GIT_COMMITTER_NAME=Test", "GIT_COMMITTER_EMAIL=test@example.com")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, output)
	}
	return strings.TrimSpace(string(output))
}

func TestHelmDiffLocalRevisions(t *testing.T) {
	repo := t.TempDir()
	helmDiffGit(t, repo, "init", "-b", "target")
	chart := func(name, value string) {
		helmDiffWrite(t, filepath.Join(repo, "system", name, "Chart.yaml"), "apiVersion: v2\nname: test\nversion: 1.0.0\n", 0644)
		helmDiffWrite(t, filepath.Join(repo, "system", name, "templates", "deep", "config.yaml"), value, 0644)
	}
	chart("changed", "old")
	chart("deleted", "deleted")
	chart("unchanged", "same")
	helmDiffGit(t, repo, "add", ".")
	helmDiffGit(t, repo, "commit", "-m", "target")
	helmDiffGit(t, repo, "checkout", "-b", "source")
	chart("changed", "new")
	chart("added", "added")
	if err := os.RemoveAll(filepath.Join(repo, "system", "deleted")); err != nil {
		t.Fatal(err)
	}
	helmDiffWrite(t, filepath.Join(repo, "system", "not-a-chart", "README"), "ignore me", 0644)
	helmDiffGit(t, repo, "add", ".")
	helmDiffGit(t, repo, "commit", "-m", "source")
	sha := helmDiffGit(t, repo, "rev-parse", "HEAD")
	bin := t.TempDir()
	helmDiffWrite(t, filepath.Join(bin, "helm"), `#!/bin/sh
if [ "$HELM_FAIL" = "$1" ]; then exit 42; fi
if [ "$1" = template ]; then
  for arg do path="$arg"; done
  cat "$path/templates/deep/config.yaml"
fi
`, 0755)
	helmDiffWrite(t, filepath.Join(bin, "dyff"), `#!/bin/sh
if [ "$DYFF_FAIL" = yes ]; then exit 43; fi
for arg do previous="$last"; last="$arg"; done
cat "$previous"
printf '|'
cat "$last"
printf '\n'
`, 0755)
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	tmp := t.TempDir()
	t.Setenv("TMPDIR", tmp)
	for _, tc := range []struct{ name, source, target, fail, message string }{
		{"branches", "source", "target", "", ""},
		{"commit", sha, "target", "", ""},
		{"dependency failure", "source", "target", "dependency", "helm:"},
		{"template failure", "source", "target", "template", "helm:"},
		{"diff failure", "source", "target", "dyff", "dyff:"},
		{"git failure", "missing-ref", "target", "", "git:"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("HELM_FAIL", tc.fail)
			t.Setenv("DYFF_FAIL", "no")
			if tc.fail == "dyff" {
				t.Setenv("DYFF_FAIL", "yes")
			}
			var output, stderr bytes.Buffer
			run := process.New("", nil, &output, &stderr)
			err := Diff(t.Context(), run, DiffOptions{Repository: repo, Source: tc.source, Target: tc.target, Subpath: "system"})
			if tc.message != "" {
				if err == nil || !strings.Contains(err.Error(), tc.message) {
					t.Fatalf("want %q failure, got %v", tc.message, err)
				}
			} else if err != nil {
				t.Fatalf("%v\n%s", err, stderr.String())
			} else if output.String() != "|added\nold|new\ndeleted|\n" {
				t.Fatalf("unexpected diff output %q", output.String())
			}
			entries, err := os.ReadDir(tmp)
			if err != nil || len(entries) != 0 {
				t.Fatalf("temporary files remain: %v, %v", entries, err)
			}
		})
	}
}
