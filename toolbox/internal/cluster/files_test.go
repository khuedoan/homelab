package cluster

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pkg/sftp"
)

func testSFTP(t *testing.T) *sftp.Client {
	t.Helper()
	server := os.Getenv("SFTP_SERVER")
	if server == "" {
		sshd, err := exec.LookPath("sshd")
		if err != nil {
			t.Skip("set SFTP_SERVER to OpenSSH's sftp-server; Nix package checks supply it")
		}
		sshd, err = filepath.EvalSymlinks(sshd)
		if err != nil {
			t.Fatal(err)
		}
		server = filepath.Join(filepath.Dir(sshd), "..", "libexec", "sftp-server")
	}
	process := exec.Command(server)
	input, err := process.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	output, err := process.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := process.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// The SFTP client normally closes its input pipe first.
		if err := input.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
			t.Error(err)
		}
		if err := process.Wait(); err != nil {
			t.Errorf("sftp-server failed: %v", err)
		}
	})
	client, err := sftp.NewClientPipe(output, input)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := client.Close(); err != nil {
			t.Error(err)
		}
	})
	return client
}

func TestSFTPPublicationDoesNotReplaceExistingToken(t *testing.T) {
	files := testSFTP(t)
	destination := filepath.Join(t.TempDir(), "token")
	if err := publishToken(files, destination, "first-secret"); err != nil {
		t.Fatal(err)
	}
	if err := publishToken(files, destination, "replacement-secret"); err == nil {
		t.Fatal("replaced existing credentials")
	}
	data, err := os.ReadFile(destination)
	if err != nil || string(data) != "first-secret" {
		t.Fatalf("wrong token: %q, %v", data, err)
	}
	info, err := os.Stat(destination)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("unsafe mode: %v, %v", info, err)
	}
	entries, err := os.ReadDir(filepath.Dir(destination))
	if err != nil || len(entries) != 1 {
		t.Fatalf("left staging credentials: %v, %v", entries, err)
	}
}

func TestSFTPPublicationRejectsSymlinkAndConcurrentWriters(t *testing.T) {
	files := testSFTP(t)
	dir := t.TempDir()
	link := filepath.Join(dir, "link")
	if err := os.Symlink("missing", link); err != nil {
		t.Fatal(err)
	}
	if err := publishToken(files, link, "secret"); err == nil {
		t.Fatal("replaced symlink")
	}
	destination := filepath.Join(dir, "token")
	type result struct {
		token string
		err   error
	}
	results := make(chan result, 2)
	for _, token := range []string{"first", "second"} {
		go func() { results <- result{token, publishToken(files, destination, token)} }()
	}
	winner := ""
	for range 2 {
		result := <-results
		if result.err == nil {
			if winner != "" {
				t.Fatal("two conflicting publishers succeeded")
			}
			winner = result.token
		}
	}
	data, err := os.ReadFile(destination)
	if err != nil || winner == "" || string(data) != winner {
		t.Fatalf("no single durable winner: %q, %q, %v", winner, data, err)
	}
}

func TestSecureTokenFormat(t *testing.T) {
	secure := "K10" + strings.Repeat("a", 64) + "::server:secret"
	if !validToken(secure) {
		t.Fatal("rejected secure server token")
	}
	for _, token := range []string{"secret", "", strings.Replace(secure, "server:", "agent:", 1), secure + "\nextra", "K10short::server:secret"} {
		if validToken(token) {
			t.Fatal("accepted malformed or non-server token")
		}
	}
}
