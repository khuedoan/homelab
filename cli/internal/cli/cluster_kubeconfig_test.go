package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestKubeconfigArguments(t *testing.T) {
	for _, args := range [][]string{
		{"cluster", "kubeconfig"},
		{"cluster", "kubeconfig", "--environment", "../production"},
		{"cluster", "kubeconfig", "--environment", "staging", "--timeout", "0s"},
		{"cluster", "kubeconfig", "--environment", "staging", "--output", ""},
		{"cluster", "kubeconfig", "extra", "--environment", "staging"},
	} {
		if out, _, err := operationExecute(args...); err == nil || out != "" {
			t.Fatalf("accepted invalid arguments or emitted credentials: %v, %v", args, err)
		}
	}
}

func TestWriteKubeconfigPermissionsReplacementAndCleanup(t *testing.T) {
	dir := t.TempDir()
	destination := filepath.Join(dir, "config")
	if err := os.WriteFile(destination, []byte("old config"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := writeKubeconfig(destination, []byte("private kubeconfig"), io.Discard); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(destination)
	if err != nil || string(data) != "private kubeconfig" {
		t.Fatalf("wrong output: %q, %v", data, err)
	}
	info, err := os.Stat(destination)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("unsafe output permissions: %v, %v", info, err)
	}
	blocked := filepath.Join(dir, "blocked")
	if err := os.Mkdir(blocked, 0700); err != nil {
		t.Fatal(err)
	}
	if err := writeKubeconfig(blocked, []byte("private kubeconfig"), io.Discard); err == nil {
		t.Fatal("ignored publication failure")
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 2 {
		t.Fatalf("left staged credentials: %v, %v", entries, err)
	}
	data, err = os.ReadFile(destination)
	if err != nil || string(data) != "private kubeconfig" {
		t.Fatal("damaged existing output on failure")
	}
}

func TestWriteKubeconfigStdoutAndSymlink(t *testing.T) {
	var output bytes.Buffer
	if err := writeKubeconfig("-", []byte("private kubeconfig\n"), &output); err != nil || output.String() != "private kubeconfig\n" {
		t.Fatalf("wrong stdout: %q, %v", output.String(), err)
	}
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	link := filepath.Join(dir, "config")
	if err := os.WriteFile(target, []byte("unrelated"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	if err := writeKubeconfig(link, []byte("private kubeconfig"), io.Discard); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "unrelated" {
		t.Fatal("followed output symlink")
	}
	info, err := os.Lstat(link)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm() != 0600 {
		t.Fatalf("unsafe replacement: %v, %v", info, err)
	}
}
