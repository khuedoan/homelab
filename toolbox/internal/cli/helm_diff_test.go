package cli

import (
	"bytes"
	"testing"
)

func TestHelmDiffFlags(t *testing.T) {
	for _, args := range [][]string{
		{},
		{"--repository", "repo", "--source", "source", "--target", "target"},
		{"--repository", "repo", "--source", "source", "--target", "target", "--subpath", "../escape"},
	} {
		cmd := newHelmDiffCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs(args)
		if err := cmd.Execute(); err == nil {
			t.Fatalf("expected invalid arguments to fail: %v", args)
		}
	}
}

func TestHelmDiffRouting(t *testing.T) {
	operationFake(t, "git", `case "$1" in
clone)
  [ "$6" = expected.repo ] || exit 17
  mkdir -p "$7/platform/chart"
  printf 'apiVersion: v2\nname: chart\nversion: 1.0.0\n' > "$7/platform/chart/Chart.yaml"
  case "$7" in */target) printf old;; */source) printf new;; *) exit 18;; esac > "$7/platform/chart/value";;
-C)
  if [ "$3" = fetch ]; then
    case "$2" in */target) [ "$7" = old-ref ];; */source) [ "$7" = new-ref ];; *) exit 19;; esac || exit 20
  fi;;
*) exit 21;; esac`)
	operationFake(t, "helm", `if [ "$1" = template ]; then
  for arg do path="$arg"; done
  cat "$path/value"
fi`)
	operationFake(t, "dyff", `for arg do previous="$last"; last="$arg"; done
cat "$previous"; printf '|'; cat "$last"`)
	out, stderr, err := operationExecute("helm", "diff", "--repository", "expected.repo", "--source", "new-ref", "--target", "old-ref", "--subpath", "platform")
	if err != nil || out != "old|new" || stderr != "" {
		t.Fatalf("CLI chart comparison = %q, %q, %v", out, stderr, err)
	}
}
