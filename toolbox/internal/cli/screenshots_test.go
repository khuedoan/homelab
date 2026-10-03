package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestScreenshotsFirefox(t *testing.T) {
	log := filepath.Join(t.TempDir(), "firefox.log")
	operationFake(t, "firefox", `
printf '%s\n' "$@" >> "$FIREFOX_LOG"
if [ "$FIREFOX_FAIL" = yes ]; then exit 41; fi
while [ "$#" -gt 0 ]; do
  if [ "$1" = --screenshot ]; then shift; printf 'PNG' > "$1"; break; fi
  shift
done
`)
	t.Setenv("FIREFOX_LOG", log)
	for _, profile := range []string{"", filepath.Join(t.TempDir(), "profile with spaces")} {
		t.Run("profile="+profile, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "screenshots with spaces")
			if err := os.WriteFile(log, nil, 0644); err != nil {
				t.Fatal(err)
			}
			cmd := newScreenshotsCmd()
			cmd.SetOut(&bytes.Buffer{})
			cmd.SetErr(&bytes.Buffer{})
			args := []string{"--output", output}
			if profile != "" {
				args = append(args, "--profile", profile)
			}
			cmd.SetArgs(args)
			if err := cmd.Execute(); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(log)
			if err != nil {
				t.Fatal(err)
			}
			var expected strings.Builder
			for _, app := range []struct{ name, url string }{
				{"home", "https://home.khuedoan.com"},
				{"gitea", "https://git.khuedoan.com/ops/homelab"},
				{"argocd", "https://argocd.khuedoan.com/applications/root"},
				{"matrix", "https://chat.khuedoan.com/#/room/#random:matrix.khuedoan.com"},
				{"grafana", "https://grafana.khuedoan.com/d/efa86fd1d0c121a26444b636a3f509a8/kubernetes-compute-resources-cluster"},
			} {
				path := filepath.Join(output, app.name+".png")
				expected.WriteString("--headless\n--window-size\n1920,1080\n")
				if profile != "" {
					expected.WriteString("--profile\n" + profile + "\n")
				}
				expected.WriteString("--screenshot\n" + path + "\n" + app.url + "\n")
				image, err := os.ReadFile(path)
				if err != nil || string(image) != "PNG" {
					t.Fatalf("missing screenshot %s: %q, %v", path, image, err)
				}
			}
			if string(data) != expected.String() {
				t.Fatalf("unexpected Firefox arguments:\n%s", data)
			}
		})
	}
	t.Run("failure", func(t *testing.T) {
		t.Setenv("FIREFOX_FAIL", "yes")
		cmd := newScreenshotsCmd()
		cmd.SetOut(&bytes.Buffer{})
		cmd.SetErr(&bytes.Buffer{})
		cmd.SetArgs([]string{"--output", t.TempDir()})
		if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "firefox:") {
			t.Fatalf("expected Firefox failure, got %v", err)
		}
	})
}

func TestScreenshotsDefaultOutput(t *testing.T) {
	cmd := newScreenshotsCmd()
	if value, err := cmd.Flags().GetString("output"); err != nil || value != "." {
		t.Fatalf("output default = %q, %v", value, err)
	}
}
