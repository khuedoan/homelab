package forgejo

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPushPublishesOnlyCommits(t *testing.T) {
	ctx := context.Background()
	source, remote := t.TempDir(), t.TempDir()
	gitTest(t, source, "init", "--quiet")
	gitTest(t, remote, "init", "--bare", "--quiet")
	writeTest(t, source, "program.txt", "committed version\n")
	gitTest(t, source, "add", ".")
	gitTest(t, source, "-c", "user.name=Test", "-c", "user.email=test@localhost", "commit", "--quiet", "-m", "Initial source")
	first := gitTest(t, source, "rev-parse", "HEAD")
	writeTest(t, source, "program.txt", "uncommitted version\n")
	gitTest(t, source, "add", "program.txt")
	writeTest(t, source, "new.txt", "intent-to-add file\n")
	gitTest(t, source, "add", "--intent-to-add", "new.txt")
	writeTest(t, source, "untracked.txt", "untracked file\n")
	before := gitTest(t, source, "status", "--porcelain")
	for range 2 {
		if err := push(ctx, source, remote, ""); err != nil {
			t.Fatal(err)
		}
		if got := gitTest(t, remote, "rev-parse", "master"); got != first {
			t.Fatal("did not push the original commit")
		}
	}
	if got := gitTest(t, remote, "show", "master:program.txt"); got != "committed version\n" {
		t.Fatalf("published content=%q", got)
	}
	if files := gitTest(t, remote, "ls-tree", "--name-only", "master"); files != "program.txt\n" {
		t.Fatalf("published uncommitted files: %s", files)
	}
	if after := gitTest(t, source, "status", "--porcelain"); after != before {
		t.Fatal("changed source checkout or index")
	}
	gitTest(t, source, "-c", "user.name=Test", "-c", "user.email=test@localhost", "commit", "--quiet", "-m", "Update source")
	if err := push(ctx, source, remote, ""); err != nil {
		t.Fatal(err)
	}
	if got := gitTest(t, remote, "rev-parse", "master"); got != gitTest(t, source, "rev-parse", "HEAD") {
		t.Fatal("did not push the updated commit")
	}
	gitTest(t, source, "reset", "--soft", "HEAD^")
	if err := push(ctx, source, remote, ""); err == nil {
		t.Fatal("accepted a non-fast-forward push")
	}
}

func gitTest(t *testing.T, directory string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = directory
	output, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	return string(output)
}

func writeTest(t *testing.T, directory, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(directory, name), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
